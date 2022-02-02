package main

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func init() {
	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount("pasg", "pasgpub")
}
