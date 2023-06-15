package main

import (
	"encoding/json"
	"fmt"

	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	feegranttypes "github.com/cosmos/cosmos-sdk/x/feegrant"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	passage "github.com/envadiv/Passage3D/app"
	"github.com/spf13/cobra"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmos "github.com/tendermint/tendermint/libs/os"
	"github.com/tendermint/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var vpDecreaseValidator1 = "pasgvaloper1xwdc6lvp30s0uesaufxhhxd9p3prhqss8mt38y"
var vpDecreaseForRemovedAccount1 = sdk.NewDecFromInt(sdk.NewInt(2515753000000))

var vpDecreaseValidator2 = "pasgvaloper10x0s5tzu6203c2zkqy37ar7pcfpdft9aepuahq"
var vpDecreaseForRemovedAccount2 = sdk.NewDecFromInt(sdk.NewInt(984244960000))

var removeDelegations = map[string][]string{
	"pasg197h5mwfpj3znhrcngjy36x4esaq8y0pmg7zp9q": {
		"pasgvaloper1xwdc6lvp30s0uesaufxhhxd9p3prhqss8mt38y",
		"pasgvaloper10x0s5tzu6203c2zkqy37ar7pcfpdft9aepuahq",
	},
}

var deductDelegation = map[string]int64{
	"pasg1qf755atr9rxy24t5ccnsctln04u8qzplt7x3qx": 1160432000000,
	"pasg12ktnvjqvv39x8pta82f55fc4n7k2rnn4r7sy8f": 1175541000000,
	"pasg1l3rh6794pnch3xz5sp7h4dcu0lees4puywjs5f": 1160430000000,
}

func MigrateRemainingState() *cobra.Command {
	cmd := &cobra.Command{
		Use: "migrate-genesis-state [old-genesis-file] [new-genesis-file] [destination-genesis-file]",
		Long: `Copy remaining module state from old-genesis.json file to new-genesis.json file and
		updates delegations and distributions`,
		Args: cobra.ExactArgs(3),
		Example: `
		go run main.go migrate-genesis-state old-genesis-file.json new-genesis-file.json dest-genesis.json
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			oldGenesisFilePath := args[0]
			newGenesisFilePath := args[1]

			oldGenesis, err := types.GenesisDocFromFile(oldGenesisFilePath)
			if err != nil {
				return err
			}

			var oldGenState map[string]json.RawMessage
			err = json.Unmarshal(oldGenesis.AppState, &oldGenState)
			if err != nil {
				return err
			}

			cdc := passage.MakeEncodingConfig()
			var oldStakeGenesis stakingtypes.GenesisState
			err = cdc.Marshaler.UnmarshalJSON(oldGenState[stakingtypes.ModuleName], &oldStakeGenesis)
			if err != nil {
				return err
			}

			validatorsMap := make(map[string]sdk.Dec)
			var delegations stakingtypes.Delegations
			temp := make(map[string]int64)
			for key, value := range deductDelegation {
				temp[key] = value
			}

			for i := 0; i < len(oldStakeGenesis.Delegations); i++ {
				delegation := oldStakeGenesis.Delegations[i]
				if value, ok := removeDelegations[delegation.DelegatorAddress]; ok {
					for i := 0; i < len(value); i++ {
						if delegation.ValidatorAddress == value[i] {
							continue
						}
					}
				} else if ds, ok := temp[delegation.DelegatorAddress]; ok {
					result := delegation.Shares.Sub(sdk.NewDecFromInt(sdk.NewInt(ds)))
					if result.IsPositive() {
						delegations = append(delegations, stakingtypes.Delegation{
							DelegatorAddress: delegation.DelegatorAddress,
							ValidatorAddress: delegation.ValidatorAddress,
							Shares:           result,
						},
						)
						if vpReduction, ok := validatorsMap[delegation.ValidatorAddress]; ok {
							validatorsMap[delegation.ValidatorAddress] = vpReduction.Add(sdk.NewDecFromInt(sdk.NewInt(ds)))
						} else {
							validatorsMap[delegation.ValidatorAddress] = sdk.NewDecFromInt(sdk.NewInt(ds))
						}
						delete(temp, delegation.DelegatorAddress)
					} else {
						delegations = append(delegations, oldStakeGenesis.Delegations[i])
					}
				} else {
					delegations = append(delegations, oldStakeGenesis.Delegations[i])
				}
			}

			var validators []stakingtypes.Validator
			for i := 0; i < len(oldStakeGenesis.Validators); i++ {
				validator := oldStakeGenesis.Validators[i]
				if value, ok := validatorsMap[validator.OperatorAddress]; ok {
					validator.Tokens = validator.Tokens.Sub(value.RoundInt())
					validator.DelegatorShares = validator.DelegatorShares.Sub(value)
				}

				if validator.OperatorAddress == vpDecreaseValidator1 {
					validator.Tokens = validator.Tokens.Sub(vpDecreaseForRemovedAccount1.RoundInt())
					validator.DelegatorShares = validator.DelegatorShares.Sub(vpDecreaseForRemovedAccount1)
				}

				if validator.OperatorAddress == vpDecreaseValidator2 {
					validator.Tokens = validator.Tokens.Sub(vpDecreaseForRemovedAccount2.RoundInt())
					validator.DelegatorShares = validator.DelegatorShares.Sub(vpDecreaseForRemovedAccount2)

				}

				validators = append(validators, validator)
			}

			var lastPowers []stakingtypes.LastValidatorPower
			for i := 0; i < len(oldStakeGenesis.LastValidatorPowers); i++ {
				if value, ok := validatorsMap[oldStakeGenesis.LastValidatorPowers[i].Address]; ok {
					lastPowers = append(lastPowers, stakingtypes.LastValidatorPower{
						Address: oldStakeGenesis.LastValidatorPowers[i].Address,
						Power:   oldStakeGenesis.LastValidatorPowers[i].Power - value.TruncateInt64(),
					})
				} else if oldStakeGenesis.LastValidatorPowers[i].Address == vpDecreaseValidator1 {
					lastPowers = append(lastPowers, stakingtypes.LastValidatorPower{
						Address: oldStakeGenesis.LastValidatorPowers[i].Address,
						Power:   oldStakeGenesis.LastValidatorPowers[i].Power - vpDecreaseForRemovedAccount1.TruncateInt64(),
					})
				} else if oldStakeGenesis.LastValidatorPowers[i].Address == vpDecreaseValidator2 {
					lastPowers = append(lastPowers, stakingtypes.LastValidatorPower{
						Address: oldStakeGenesis.LastValidatorPowers[i].Address,
						Power:   oldStakeGenesis.LastValidatorPowers[i].Power - vpDecreaseForRemovedAccount2.TruncateInt64(),
					})
				} else {
					lastPowers = append(lastPowers, oldStakeGenesis.LastValidatorPowers[i])
				}
			}

			newStakeGenesis := stakingtypes.GenesisState{}
			newStakeGenesis.Delegations = delegations
			newStakeGenesis.Validators = validators
			newStakeGenesis.Exported = oldStakeGenesis.Exported
			newStakeGenesis.LastTotalPower = oldStakeGenesis.LastTotalPower
			newStakeGenesis.LastValidatorPowers = lastPowers
			newStakeGenesis.Params = oldStakeGenesis.Params
			newStakeGenesis.Redelegations = oldStakeGenesis.Redelegations
			newStakeGenesis.UnbondingDelegations = oldStakeGenesis.UnbondingDelegations

			slashingState := oldGenState[slashingtypes.ModuleName]

			newGenesis, err := types.GenesisDocFromFile(newGenesisFilePath)
			if err != nil {
				return err
			}

			var newGenesisState map[string]json.RawMessage
			err = json.Unmarshal(newGenesis.AppState, &newGenesisState)
			if err != nil {
				return err
			}

			bz, err := cdc.Marshaler.MarshalJSON(&newStakeGenesis)
			if err != nil {
				return err
			}
			newGenesisState[stakingtypes.ModuleName] = bz
			newGenesisState[slashingtypes.ModuleName] = slashingState

			authzState := oldGenState[authztypes.ModuleName]
			newGenesisState[authztypes.ModuleName] = authzState

			feegrantState := oldGenState[feegranttypes.ModuleName]
			newGenesisState[feegranttypes.ModuleName] = feegrantState

			mintState := oldGenState[minttypes.ModuleName]
			newGenesisState[minttypes.ModuleName] = mintState

			capabilityState := oldGenState[capabilitytypes.ModuleName]
			newGenesisState[capabilitytypes.ModuleName] = capabilityState

			var distrGenesis distributiontypes.GenesisState
			var startingInfoRecords []distributiontypes.DelegatorStartingInfoRecord
			err = cdc.Marshaler.UnmarshalJSON(oldGenState[distributiontypes.ModuleName], &distrGenesis)
			if err != nil {
				return err
			}

			for i := 0; i < len(distrGenesis.DelegatorStartingInfos); i++ {
				distr := distrGenesis.DelegatorStartingInfos[i]
				if _, ok := removeDelegations[distr.DelegatorAddress]; ok {
					continue
				} else if delegation, ok := deductDelegation[distr.DelegatorAddress]; ok {
					if _, ok := validatorsMap[distr.ValidatorAddress]; ok {
						startingInfoRecords = append(startingInfoRecords, distributiontypes.DelegatorStartingInfoRecord{
							DelegatorAddress: distr.DelegatorAddress,
							ValidatorAddress: distr.ValidatorAddress,
							StartingInfo: distributiontypes.DelegatorStartingInfo{
								PreviousPeriod: distr.StartingInfo.PreviousPeriod,
								Height:         distr.StartingInfo.Height,
								Stake:          distr.StartingInfo.Stake.Sub(sdk.NewDecFromInt(sdk.NewInt(delegation))),
							},
						})
					}

				} else {
					startingInfoRecords = append(startingInfoRecords, distr)
				}
			}

			distrGenesis.DelegatorStartingInfos = startingInfoRecords
			distrGenesis.FeePool.CommunityPool = sdk.NewDecCoins(
				sdk.NewDecCoin(UPassageDenom, sdk.NewInt(CommunityPoolPassage3DAmount).Mul(sdk.NewInt(1000000))),
			)
			bz, err = cdc.Marshaler.MarshalJSON(&distrGenesis)
			if err != nil {
				return err
			}

			newGenesisState[distributiontypes.ModuleName] = bz

			// memoize validators
			validatorToStatus := make(map[string]stakingtypes.BondStatus)
			for _, val := range newStakeGenesis.Validators {
				validatorToStatus[val.OperatorAddress] = val.Status
			}

			bondedAccountToStake := make(map[string]sdk.Dec)
			notbondedAccountToStake := make(map[string]sdk.Dec)
			bondedBalance := sdk.NewInt(0)
			notBondedBalance := sdk.NewInt(0)
			for _, delegation := range newStakeGenesis.Delegations {
				if status, ok := validatorToStatus[delegation.ValidatorAddress]; ok && status == stakingtypes.Bonded {
					bondedAccountToStake[delegation.DelegatorAddress] = delegation.Shares
					bondedBalance.Add(delegation.Shares.TruncateInt())
				} else {
					notbondedAccountToStake[delegation.DelegatorAddress] = delegation.Shares
					notBondedBalance.Add(delegation.Shares.TruncateInt())
				}
			}

			bankState := oldGenState[banktypes.ModuleName]
			var bankGenesis banktypes.GenesisState
			err = cdc.Marshaler.UnmarshalJSON(bankState, &bankGenesis)
			if err != nil {
				return err
			}

			bondedPoolAddress := "pasg1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3krs45g"
			notBondedPoolAddress := "pasg1tygms3xhhs3yv487phx3dw4a95jn7t7lzrvyzu"
			for index, balance := range bankGenesis.Balances {
				if balance.Address == bondedPoolAddress {
					bankGenesis.Balances[index] = banktypes.Balance{
						Address: balance.Address,
						Coins:   sdk.NewCoins(sdk.NewCoin(UPassageDenom, bondedBalance)),
					}
				} else if balance.Address == notBondedPoolAddress {
					bankGenesis.Balances[index] = banktypes.Balance{
						Address: balance.Address,
						Coins:   sdk.NewCoins(sdk.NewCoin(UPassageDenom, notBondedBalance)),
					}
				} else {
					remainingAmount := balance.Coins
					bondedBalance, ok := bondedAccountToStake[balance.Address]
					if ok {
						remainingAmount, ok = remainingAmount.SafeSub(sdk.NewCoins(sdk.NewCoin(UPassageDenom, bondedBalance.TruncateInt())))
						if !ok {
							fmt.Println(bondedBalance, "  ", balance.Address, "  ", balance.Coins.String(), "  ", remainingAmount)
							panic("failed to deduct staked amount 1")
						}
					}

					notBondedBalance, ok := notbondedAccountToStake[balance.Address]
					if ok {
						remainingAmount, ok = balance.Coins.SafeSub(sdk.NewCoins(sdk.NewCoin(UPassageDenom, notBondedBalance.RoundInt())))
						if !ok {
							fmt.Println(notBondedBalance, "   ", balance.Address)
							panic("failed to deduct staked amount 2")
						}

					}
					bankGenesis.Balances[index] = banktypes.Balance{
						Address: balance.Address,
						Coins:   remainingAmount,
					}
				}
			}

			bz, err = cdc.Marshaler.MarshalJSON(&bankGenesis)
			if err != nil {
				return err
			}
			newGenesisState[banktypes.ModuleName] = bz

			bz, err = tmjson.Marshal(newGenesisState)
			if err != nil {
				return err
			}
			newGenesis.AppState = bz
			genDocBytes, err := tmjson.MarshalIndent(newGenesis, "", "  ")
			if err != nil {
				return err
			}

			return tmos.WriteFile(args[2], genDocBytes, 0644)
		},
	}

	return cmd
}
