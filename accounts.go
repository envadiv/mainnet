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

var genesisTime = time.Now().UTC().AddDate(0, 0, 15) // TODO: update genesis time
var airdropModuleAccountAmount = sdk.NewCoins(sdk.NewCoin(UPassageDenom, sdk.NewInt(18946800000000)))

const errorsAsWarnings = true

const removeAccount = "pasg197h5mwfpj3znhrcngjy36x4esaq8y0pmg7zp9q"
const airdropPoolAddress = "pasg1lel0s624jr9zsz4ml6yv9e5r4uzukfs7hwh22w"

var repalceDelegationMap = map[string]string{
	"pasg1qf755atr9rxy24t5ccnsctln04u8qzplt7x3qx": "pasg1t70qczjpxdtpwftyw750cmud7jzyc94gn90syj",
	"pasg12ktnvjqvv39x8pta82f55fc4n7k2rnn4r7sy8f": "pasg1t70qczjpxdtpwftyw750cmud7jzyc94gn90syj",
	"pasg1l3rh6794pnch3xz5sp7h4dcu0lees4puywjs5f": "pasg1y5cqly7q25den0av2wf7vyvfxlmu724md4qvsg",
	"pasg197h5mwfpj3znhrcngjy36x4esaq8y0pmg7zp9q": "pasg1y5cqly7q25den0av2wf7vyvfxlmu724md4qvsg",
}

var addDelegationAountMap = map[string]sdk.Int{
	"pasg1t70qczjpxdtpwftyw750cmud7jzyc94gn90syj": sdk.NewInt(7252711000000),
	"pasg1y5cqly7q25den0av2wf7vyvfxlmu724md4qvsg": sdk.NewInt(9302167960000),
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
				if pVestingAcc, ok := oldAccount.(*vesting.PeriodicVestingAccount); ok {
					if pVestingAcc.DelegatedVesting.IsAllGT(nVestingAcc.OriginalVesting) {
						nVestingAcc.AccountNumber = pVestingAcc.AccountNumber
						nVestingAcc.Sequence = pVestingAcc.Sequence
						nVestingAcc.PubKey = pVestingAcc.PubKey
						any, err := codectypes.NewAnyWithValue(nVestingAcc)
						if err != nil {
							return err
						}
						authState.Accounts[oldIndex] = any
					} else {
						if pVestingAcc.GetAddress().String() == airdropPoolAddress {
							pVestingAcc.OriginalVesting = nVestingAcc.OriginalVesting.Sub(airdropModuleAccountAmount)
							pVestingAcc.StartTime = nVestingAcc.StartTime
							pVestingAcc.EndTime = nVestingAcc.EndTime
							pVestingAcc.VestingPeriods = nVestingAcc.VestingPeriods
							pVestingAcc.DelegatedFree = sdk.Coins{}
						} else {
							pVestingAcc.OriginalVesting = nVestingAcc.OriginalVesting
							pVestingAcc.StartTime = nVestingAcc.StartTime
							pVestingAcc.EndTime = nVestingAcc.EndTime
							pVestingAcc.VestingPeriods = nVestingAcc.VestingPeriods
							pVestingAcc.DelegatedFree = sdk.Coins{}
						}
						any, err := codectypes.NewAnyWithValue(pVestingAcc)
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
				account.SetAccountNumber(oldAccount.GetAccountNumber())
				account.SetSequence(oldAccount.GetSequence())
				account.SetPubKey(oldAccount.GetPubKey())
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

	// pasg197h5mwfpj3znhrcngjy36x4esaq8y0pmg7zp9q account balance is set to remaining amount, so changing account type to base account
	for i, account := range authState.Accounts {
		accountI, ok := account.GetCachedValue().(authtypes.AccountI)
		if !ok {
			panic("failed to get account")
		}

		addr := accountI.GetAddress().String()
		if addr == removeAccount {
			x := authtypes.NewBaseAccount(accountI.GetAddress(), accountI.GetPubKey(), accountI.GetAccountNumber(), accountI.GetSequence())
			any, err := codectypes.NewAnyWithValue(x)
			if err != nil {
				return err
			}

			authState.Accounts[i] = any
		} else {
			vestigAmount, found := addDelegationAountMap[addr]
			if found {
				vestingAccount, ok := accountI.(*vesting.PeriodicVestingAccount)
				if ok {
					vestingAccount.DelegatedVesting = vestingAccount.DelegatedVesting.Add(sdk.NewCoins(sdk.NewCoin(UPassageDenom, vestigAmount))...)
				}

				any, err := codectypes.NewAnyWithValue(vestingAccount)
				if err != nil {
					return err
				}

				authState.Accounts[i] = any
			}
		}
	}

	// add new accounts to auth state
	for _, account := range newAccountsToAdd {
		vestigAmount, found := addDelegationAountMap[account.GetAddress().String()]
		if found {
			vestingAccount, ok := account.(*vesting.PeriodicVestingAccount)
			if ok {
				vestingAccount.DelegatedVesting = vestingAccount.DelegatedVesting.Add(sdk.NewCoins(sdk.NewCoin(UPassageDenom, vestigAmount))...)
			}

			any, err := codectypes.NewAnyWithValue(vestingAccount)
			if err != nil {
				return err
			}
			authState.Accounts = append(authState.Accounts, any)
		} else {
			any, err := codectypes.NewAnyWithValue(account)
			if err != nil {
				panic("failed to convert to any account")
			}
			authState.Accounts = append(authState.Accounts, any)
		}
	}

	var bankState banktypes.GenesisState
	cdc.Marshaler.MustUnmarshalJSON(genState[banktypes.ModuleName], &bankState)

	addressToBalanceIndex := make(map[string]int) // account address to balance index: old state
	for j := 0; j < len(bankState.Balances); j++ {
		balance := bankState.Balances[j]
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
	for _, delegation := range oldStakeGenesis.Delegations {
		repalceAddr, found := repalceDelegationMap[delegation.DelegatorAddress]
		if found {
			delegations = append(delegations, stakingtypes.Delegation{
				DelegatorAddress: repalceAddr,
				ValidatorAddress: delegation.ValidatorAddress,
				Shares:           delegation.Shares,
			})
		} else {
			delegations = append(delegations, delegation)
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
		replaceAddr, found := repalceDelegationMap[distr.DelegatorAddress]
		if found {
			startingInfoRecords = append(startingInfoRecords, distributiontypes.DelegatorStartingInfoRecord{
				DelegatorAddress: replaceAddr,
				ValidatorAddress: distr.ValidatorAddress,
				StartingInfo:     distr.StartingInfo,
			})
		} else {
			startingInfoRecords = append(startingInfoRecords, distr)
		}
	}
	distrGenesis.DelegatorStartingInfos = startingInfoRecords

	vestingAccountToRemaining := make(map[string]sdk.Coins)
	for _, acc := range authState.Accounts {
		if acc.TypeUrl == "/cosmos.vesting.v1beta1.PeriodicVestingAccount" {
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
		}
	}

	var newAccountsBalance []banktypes.Balance
	for _, balance := range balancesToAdd {
		newAccountsBalance = append(newAccountsBalance, balance)
	}
	bankState.Balances = append(bankState.Balances, newAccountsBalance...)

	var supply sdk.Coins
	// community pool has extra 21302upasg tokens than genesis supply.
	communityPoolBalance := sdk.NewCoins(sdk.NewCoin(UPassageDenom, sdk.NewInt(150000000000000)).Add(sdk.NewCoin(UPassageDenom, sdk.NewInt(21302))))
	const airdropPoolAddress = "pasg1lel0s624jr9zsz4ml6yv9e5r4uzukfs7hwh22w"
	const distributionModuleAddress = "pasg1jv65s3grqf6v6jl3dp4t6c9t9rk99cd8y8axyq"

	for index, balance := range bankState.Balances {
		remaining, found := vestingAccountToRemaining[balance.Address]
		if balance.Address == removeAccount {
			coins := sdk.NewCoins(sdk.NewCoin(UPassageDenom, sdk.NewInt(128488)))
			updateBalanceAndSupply(&bankState.Balances[index], coins, &supply)
		} else if balance.Address == airdropPoolAddress { // remove claim module account balance from airdrop pool
			coins := balance.Coins.Sub(airdropModuleAccountAmount)
			updateBalanceAndSupply(&bankState.Balances[index], coins, &supply)
		} else if found {
			updateBalanceAndSupply(&bankState.Balances[index], remaining, &supply)
		} else if balance.Address == distributionModuleAddress { // set distribution module account balance
			updateBalanceAndSupply(&bankState.Balances[index], communityPoolBalance, &supply)
		} else {
			updateBalanceAndSupply(&bankState.Balances[index], balance.Coins, &supply)
		}
	}

	bankState.Supply = supply
	distrGenesis.FeePool = distributiontypes.FeePool{
		CommunityPool: sdk.NewDecCoins(sdk.NewDecCoinFromCoin(sdk.NewCoin(UPassageDenom, sdk.NewInt(150000000000000).Add(sdk.NewInt(21302))))),
	}

	fmt.Println("Total Supply = ", supply.String())
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

// Function to update balance and supply
func updateBalanceAndSupply(balance *banktypes.Balance, coins sdk.Coins, supply *sdk.Coins) {
	*supply = supply.Add(coins...)
	balance.Coins = coins
}
