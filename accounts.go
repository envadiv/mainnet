package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	passage "github.com/envadiv/Passage3D/app"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/types"
)

var genesisTime = time.Now().UTC().AddDate(0, 0, 5) // TODO: update genesis time

const errorsAsWarnings = true

const removeAccount = "pasg197h5mwfpj3znhrcngjy36x4esaq8y0pmg7zp9q"

var deductDelegation = map[string]sdk.Dec{
	"pasg1qf755atr9rxy24t5ccnsctln04u8qzplt7x3qx": sdk.NewDec(5802170000000),
	"pasg12ktnvjqvv39x8pta82f55fc4n7k2rnn4r7sy8f": sdk.NewDec(1450541000000),
	"pasg1l3rh6794pnch3xz5sp7h4dcu0lees4puywjs5f": sdk.NewDec(5802170000000),
}

func MigrateAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "migrate-state [accounts-csv] [genesesis.json] [exported-genesis.json]",
		Args: cobra.ExactArgs(3),
		Example: `
		go run . migrate-state accounts.csv exported-genesis.json exportted-genesis.json
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := MigrateAccount(args); err != nil {
				panic(err)
			}

			return nil
		},
	}

	return cmd
}

func MigrateAccount(args []string) error {
	accountsCsv, err := os.Open(args[0])
	if err != nil {
		return err
	}

	doc, err := types.GenesisDocFromFile(args[1])
	if err != nil {
		return err
	}

	auditTsv, err := os.OpenFile(filepath.Join("./", "account_dump.tsv"), os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	accounts, balances, err := buildAccounts(accountsCsv, genesisTime, auditTsv, false)
	if err != nil {
		return err
	}

	var genState map[string]json.RawMessage
	err = json.Unmarshal(doc.AppState, &genState)
	if err != nil {
		return err
	}

	cdc := passage.MakeEncodingConfig()

	var authState authtypes.GenesisState
	cdc.Marshaler.MustUnmarshalJSON(genState[authtypes.ModuleName], &authState)

	addressToAccount := make(map[string]authtypes.AccountI) // account address to account map: old state
	addressToIndex := make(map[string]int)                  // account address to account index: old state
	fmt.Println(len(authState.Accounts))
	for j := 0; j < len(authState.Accounts); j++ {
		anyAccount := authState.Accounts[j]
		account, ok := anyAccount.GetCachedValue().(authtypes.AccountI)
		if !ok {
			return errors.New(fmt.Sprintf("failed to decode account: %v", anyAccount))
		}
		addressToAccount[account.GetAddress().String()] = account
		addressToIndex[account.GetAddress().String()] = j
	}

	var newAccountsToAdd []authtypes.AccountI
	for i := 0; i < len(accounts); i++ {
		account := accounts[i]
		address := account.GetAddress().String()
		if oldAccount, ok := addressToAccount[address]; ok {
			oldIndex, ok := addressToIndex[address]
			if !ok {
				panic(fmt.Sprintf("account: account not found,%s", address))
			}

			if nVestingAcc, ok := account.(*vesting.PeriodicVestingAccount); ok {
				if pAccount, ok := oldAccount.(*vesting.PeriodicVestingAccount); ok {
					if pAccount.DelegatedVesting.IsAllGT(nVestingAcc.OriginalVesting) {
						fmt.Println("More = ", address)
						fmt.Println("More = ", pAccount.DelegatedVesting.String())
						// remainingBalance := pAccount.OriginalVesting.Sub(pAccount.DelegatedVesting)
						pAccount.OriginalVesting = nVestingAcc.OriginalVesting
						pAccount.StartTime = nVestingAcc.StartTime
						pAccount.EndTime = nVestingAcc.EndTime
						pAccount.VestingPeriods = nVestingAcc.VestingPeriods
						pAccount.DelegatedVesting = nVestingAcc.DelegatedVesting // nVestingAcc.OriginalVesting.Sub(remainingBalance)
						pAccount.DelegatedFree = nVestingAcc.DelegatedFree

						// addressToVestingAmount[address] = pAccount.DelegatedVesting
						any, err := codectypes.NewAnyWithValue(pAccount)
						if err != nil {
							return err
						}

						authState.Accounts[oldIndex] = any
					} else {
						pAccount.OriginalVesting = nVestingAcc.OriginalVesting
						pAccount.StartTime = nVestingAcc.StartTime
						pAccount.EndTime = nVestingAcc.EndTime
						pAccount.VestingPeriods = nVestingAcc.VestingPeriods

						any, err := codectypes.NewAnyWithValue(pAccount)
						if err != nil {
							return err
						}

						authState.Accounts[oldIndex] = any
					}
				} else {
					any, err := codectypes.NewAnyWithValue(nVestingAcc)
					if err != nil {
						return err
					}

					authState.Accounts[oldIndex] = any
				}
			} else {
				any, err := codectypes.NewAnyWithValue(account)
				if err != nil {
					return err
				}

				authState.Accounts[oldIndex] = any
			}

		} else {
			newAccountsToAdd = append(newAccountsToAdd, account)
		}
	}

	// pasg197h5mwfpj3znhrcngjy36x4esaq8y0pmg7zp9q account balance is set to zero, so changing account type to base account
	for i, account := range authState.Accounts {
		a, ok := account.GetCachedValue().(authtypes.AccountI)
		if !ok {
			panic("failed to get account")
		}

		addr := a.GetAddress().String()
		if addr == removeAccount {
			x := authtypes.NewBaseAccount(a.GetAddress(), a.GetPubKey(), a.GetAccountNumber(), a.GetSequence())
			any, err := codectypes.NewAnyWithValue(x)
			if err != nil {
				return err
			}

			authState.Accounts[i] = any
			break
		}
	}

	// add new accounts to auth state
	for _, account := range newAccountsToAdd {
		any, err := codectypes.NewAnyWithValue(account)
		if err != nil {
			panic("failed to convert to any account")
		}
		authState.Accounts = append(authState.Accounts, any)
	}

	var bankState banktypes.GenesisState
	cdc.Marshaler.MustUnmarshalJSON(genState[banktypes.ModuleName], &bankState)

	addressToBalance := make(map[string]banktypes.Balance) // account address to balance map: old state
	addressToBalanceIndex := make(map[string]int)          // account address to balance index: old state
	for j := 0; j < len(bankState.Balances); j++ {
		balance := bankState.Balances[j]
		addressToBalance[balance.Address] = balance
		addressToBalanceIndex[balance.Address] = j
	}

	var balancesToAdd []banktypes.Balance
	for _, b := range balances {
		index, ok := addressToBalanceIndex[b.Address]
		if !ok {
			balancesToAdd = append(balancesToAdd, b)
		} else {
			bankState.Balances[index] = b
		}
	}

	// staking
	var oldStakeGenesis stakingtypes.GenesisState
	err = cdc.Marshaler.UnmarshalJSON(genState[stakingtypes.ModuleName], &oldStakeGenesis)
	if err != nil {
		return err
	}

	// validators index

	validatorToStatusMap := make(map[string]stakingtypes.BondStatus) // validator to status map
	validatorToIndexMap := make(map[string]int)                      // validator to index map
	for index, validator := range oldStakeGenesis.Validators {
		validatorToStatusMap[validator.OperatorAddress] = validator.Status
		validatorToIndexMap[validator.OperatorAddress] = index
	}

	// update delegations
	var delegations stakingtypes.Delegations
	validatorToLastPowerMap := make(map[string]int64)
	bondedTokensToRemove := sdk.NewDec(0)
	notBondedTokensToRemove := sdk.NewDec(0)
	for _, delegation := range oldStakeGenesis.Delegations {
		if delegation.DelegatorAddress == removeAccount {
			if index, ok := validatorToIndexMap[delegation.ValidatorAddress]; ok {
				oldStakeGenesis.Validators[index].DelegatorShares = oldStakeGenesis.Validators[index].DelegatorShares.Sub(delegation.Shares)
				oldStakeGenesis.Validators[index].Tokens = oldStakeGenesis.Validators[index].Tokens.Sub(delegation.Shares.TruncateInt())
				validatorToLastPowerMap[delegation.ValidatorAddress] = oldStakeGenesis.Validators[index].DelegatorShares.TruncateInt64()
				if oldStakeGenesis.Validators[index].Status == stakingtypes.Bonded {
					bondedTokensToRemove = bondedTokensToRemove.Add(delegation.Shares)
				} else {
					notBondedTokensToRemove = notBondedTokensToRemove.Add(delegation.Shares)
				}
			}
		} else {
			if _, ok := deductDelegation[delegation.DelegatorAddress]; ok {
				if index, ok := validatorToIndexMap[delegation.ValidatorAddress]; ok {
					oldStakeGenesis.Validators[index].DelegatorShares = oldStakeGenesis.Validators[index].DelegatorShares.Sub(delegation.Shares)
					oldStakeGenesis.Validators[index].Tokens = oldStakeGenesis.Validators[index].Tokens.Sub(delegation.Shares.TruncateInt())
					validatorToLastPowerMap[delegation.ValidatorAddress] = oldStakeGenesis.Validators[index].DelegatorShares.TruncateInt64()
					if oldStakeGenesis.Validators[index].Status == stakingtypes.Bonded {
						bondedTokensToRemove = bondedTokensToRemove.Add(delegation.Shares)
					} else {
						notBondedTokensToRemove = notBondedTokensToRemove.Add(delegation.Shares)
					}
				}
			} else {
				delegations = append(delegations, delegation)
			}
		}
	}

	// update validator's last power
	for i, lastPower := range oldStakeGenesis.LastValidatorPowers {
		if power, ok := validatorToLastPowerMap[lastPower.Address]; ok {
			oldStakeGenesis.LastValidatorPowers[i].Power = power
		}
	}
	oldStakeGenesis.Delegations = delegations

	// update distribution VP
	var distrGenesis distributiontypes.GenesisState
	var startingInfoRecords []distributiontypes.DelegatorStartingInfoRecord
	err = cdc.Marshaler.UnmarshalJSON(genState[distributiontypes.ModuleName], &distrGenesis)
	if err != nil {
		return err
	}

	for i := 0; i < len(distrGenesis.DelegatorStartingInfos); i++ {
		distr := distrGenesis.DelegatorStartingInfos[i]
		if distr.DelegatorAddress == removeAccount {
			continue
		} else if _, ok := deductDelegation[distr.DelegatorAddress]; ok {
			continue
		} else {
			startingInfoRecords = append(startingInfoRecords, distr)
		}
	}
	distrGenesis.DelegatorStartingInfos = startingInfoRecords

	// TODO: update bonded and not-bonded pool balance and bank total supply
	var bondedPool, notBondedPool string
	vestingAccountToRemaining := make(map[string]sdk.Coins)
	for _, acc := range authState.Accounts {
		switch acc.TypeUrl {
		case "/cosmos.auth.v1beta1.BaseAccount":

		case "/cosmos.vesting.v1beta1.PeriodicVestingAccount":
			if a1, ok := acc.GetCachedValue().(authtypes.AccountI); ok {
				if vestingAcc, ok := a1.(*vesting.PeriodicVestingAccount); ok {
					delegated := vestingAcc.DelegatedVesting
					original := vestingAcc.OriginalVesting
					available := original.Sub(delegated)
					vestingAccountToRemaining[vestingAcc.Address] = available
				}
			} else {
				panic("failed to decode")
			}
		case "/cosmos.auth.v1beta1.ModuleAccount":
			if ma, ok := acc.GetCachedValue().(authtypes.ModuleAccountI); ok {
				if ma.GetName() == "bonded_tokens_pool" {
					bondedPool = ma.GetAddress().String()
				} else if ma.GetName() == "not_bonded_tokens_pool" {
					notBondedPool = ma.GetAddress().String()
				}
			}
		}
	}

	for _, balance := range balancesToAdd {
		bankState.Balances = append(bankState.Balances, balance)
	}

	for index, balance := range bankState.Balances {
		if balance.Address == removeAccount {
			bankState.Balances[index] = banktypes.Balance{
				Address: balance.Address,
				Coins:   sdk.NewCoins(sdk.NewCoin(UPassageDenom, sdk.NewInt(0))),
			}
		}
		if remaining, exist := vestingAccountToRemaining[balance.Address]; exist {
			bankState.Balances[index] = banktypes.Balance{
				Address: balance.Address,
				Coins:   remaining,
			}
		}

		if balance.Address == bondedPool {
			bankState.Balances[index] = banktypes.Balance{
				Address: balance.Address,
				Coins:   balance.Coins.Sub(sdk.NewCoins(sdk.NewCoin(UPassageDenom, bondedTokensToRemove.TruncateInt()))),
			}
		}

		if balance.Address == notBondedPool {
			bankState.Balances[index] = banktypes.Balance{
				Address: balance.Address,
				Coins:   balance.Coins.Sub(sdk.NewCoins(sdk.NewCoin(UPassageDenom, notBondedTokensToRemove.TruncateInt()))),
			}
		}
	}

	genState[authtypes.ModuleName] = cdc.Marshaler.MustMarshalJSON(&authState)
	genState[banktypes.ModuleName] = cdc.Marshaler.MustMarshalJSON(&bankState)
	genState[stakingtypes.ModuleName] = cdc.Marshaler.MustMarshalJSON(&oldStakeGenesis)
	genState[distributiontypes.ModuleName] = cdc.Marshaler.MustMarshalJSON(&distrGenesis)

	bz, err := json.Marshal(genState)
	if err != nil {
		return err
	}
	doc.AppState = bz

	return doc.SaveAs(args[2])

}
