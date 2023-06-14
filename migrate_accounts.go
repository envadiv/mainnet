package main

import (
	"encoding/json"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	passage "github.com/envadiv/Passage3D/app"
	"github.com/spf13/cobra"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmos "github.com/tendermint/tendermint/libs/os"
	"github.com/tendermint/tendermint/types"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

type AccountItem struct {
	accountNumber uint64
	sequence      uint64
	pubkey        cryptotypes.PubKey
}

func MigrateAccounts() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "migrate-accounts [old-genesis-file] [new-genesis-file] [destination-file]",
		Long: "migrate account number and sequence from old-genesis.json to file new-genesis.json file",
		Args: cobra.ExactArgs(3),
		Example: `
		go run main.go migrate-accounts old-genesis-file.json new-genesis-file.json migrated-genesis-file.json 
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			sourceGenesisFilePath := args[0]
			destinationGenesisFilePath := args[1]

			source, err := types.GenesisDocFromFile(sourceGenesisFilePath)
			if err != nil {
				return err
			}

			var sourceGenState map[string]json.RawMessage
			err = json.Unmarshal(source.AppState, &sourceGenState)
			if err != nil {
				return err
			}

			cdc := passage.MakeEncodingConfig()
			var sourceAuthGenesis authtypes.GenesisState
			err = cdc.Marshaler.UnmarshalJSON(sourceGenState[authtypes.ModuleName], &sourceAuthGenesis)
			if err != nil {
				return err
			}

			oldAccountsMap := make(map[string]AccountItem)
			var moduleAccounts []*cdctypes.Any
			for i := 0; i < len(sourceAuthGenesis.Accounts); i++ {
				account := sourceAuthGenesis.Accounts[i]
				acc, ok := account.GetCachedValue().(authtypes.AccountI)
				if !ok {
					panic("failed to decode account")
				}
				acc.GetPubKey()
				oldAccountsMap[acc.GetAddress().String()] = AccountItem{
					accountNumber: acc.GetAccountNumber(),
					sequence:      acc.GetSequence(),
					pubkey:        acc.GetPubKey(),
				}

				if account.TypeUrl == "/cosmos.auth.v1beta1.ModuleAccount" &&
					acc.GetAddress().String() != "pasg1jv65s3grqf6v6jl3dp4t6c9t9rk99cd8y8axyq" { //ignore community pool as it is already created at genesis
					moduleAccounts = append(moduleAccounts, account)
				}
			}

			// destination
			destination, err := types.GenesisDocFromFile(destinationGenesisFilePath)
			if err != nil {
				return err
			}

			var destGenState map[string]json.RawMessage
			err = json.Unmarshal(destination.AppState, &destGenState)
			if err != nil {
				return err
			}

			var destAuthGenesis authtypes.GenesisState
			err = cdc.Marshaler.UnmarshalJSON(destGenState[authtypes.ModuleName], &destAuthGenesis)
			if err != nil {
				return err
			}

			for i := 0; i < len(destAuthGenesis.Accounts); i++ {
				account := destAuthGenesis.Accounts[i]
				acc, ok := account.GetCachedValue().(authtypes.AccountI)
				if !ok {
					panic("failed to decode account")
				}

				address := acc.GetAddress().String()
				if v, ok := oldAccountsMap[address]; ok {
					acc.SetAccountNumber(v.accountNumber)
					acc.SetSequence(v.sequence)
					acc.SetPubKey(v.pubkey)
					a, err := cdctypes.NewAnyWithValue(acc)
					if err != nil {
						return err
					}
					destAuthGenesis.Accounts[i] = a
				}
			}

			temp := destAuthGenesis.Accounts
			for i := 0; i < len(moduleAccounts); i++ {
				temp = append(temp, moduleAccounts[i])
			}

			destAuthGenesis.Accounts = temp
			bz, err := cdc.Marshaler.MarshalJSON(&destAuthGenesis)
			if err != nil {
				panic(err)
			}

			destGenState[authtypes.ModuleName] = bz
			appState, err := tmjson.Marshal(destGenState)
			if err != nil {
				panic(err)
			}

			destination.AppState = appState
			genDocBytes, err := tmjson.MarshalIndent(destination, "", "  ")
			if err != nil {
				return err
			}

			return tmos.WriteFile(args[2], genDocBytes, 0644)
		},
	}

	return cmd
}
