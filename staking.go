package main

import (
	"encoding/json"

	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	feegranttypes "github.com/cosmos/cosmos-sdk/x/feegrant"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"

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

var removeAccount = map[string]bool{
	"pasg197h5mwfpj3znhrcngjy36x4esaq8y0pmg7zp9q": true,
}

var deductDelegation = map[string]int64{
	"pasg1qf755atr9rxy24t5ccnsctln04u8qzplt7x3qx": 1160432,
	"pasg12ktnvjqvv39x8pta82f55fc4n7k2rnn4r7sy8f": 1175541,
	"pasg1l3rh6794pnch3xz5sp7h4dcu0lees4puywjs5f": 1160432,
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
				if _, ok := removeAccount[delegation.DelegatorAddress]; ok {
					continue
				} else if ds, ok := temp[delegation.DelegatorAddress]; ok {
					result := delegation.Shares.Sub(sdk.NewDecFromInt(sdk.NewInt(ds)))
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
			}

			var validators []stakingtypes.Validator
			for i := 0; i < len(oldStakeGenesis.Validators); i++ {
				validator := oldStakeGenesis.Validators[i]
				if value, ok := validatorsMap[validator.OperatorAddress]; ok {
					validator.Tokens = validator.Tokens.Sub(value.RoundInt())
					validator.DelegatorShares = validator.DelegatorShares.Sub(value)
				}

				validators = append(validators, validator)
			}

			newStakeGenesis := stakingtypes.GenesisState{}
			newStakeGenesis.Delegations = delegations
			newStakeGenesis.Validators = validators
			newStakeGenesis.Exported = oldStakeGenesis.Exported
			newStakeGenesis.LastTotalPower = oldStakeGenesis.LastTotalPower
			newStakeGenesis.LastValidatorPowers = oldStakeGenesis.LastValidatorPowers
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
				if _, ok := removeAccount[distr.DelegatorAddress]; ok {
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
