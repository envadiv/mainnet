package main

import (
	"os"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func ExamplePrintAccountAudit() {
	addr0, err := sdk.AccAddressFromBech32("pasg1ktvsz07ca4amg5cja68pph3t65lj4wdulrz464")
	if err != nil {
		panic(err)
	}
	addr1, err := sdk.AccAddressFromBech32("pasg16ll5l3zdu086ug96cau00k3rllqg9eeyz7ss7t")
	if err != nil {
		panic(err)
	}
	genesisTime, err := time.Parse(time.RFC3339, "2021-04-08T16:00:00Z")
	if err != nil {
		panic(err)
	}
	t0, err := time.Parse(time.RFC3339, "2021-05-03T16:00:00Z")
	if err != nil {
		panic(err)
	}
	t1, err := time.Parse(time.RFC3339, "2021-05-08T16:00:00Z")
	if err != nil {
		panic(err)
	}
	five, _ := NewDecFromString("5")
	ten, _ := NewDecFromString("10")
	fifteen, _ := NewDecFromString("15")

	PrintAccountAudit([]Account{
		{
			Address:      addr0,
			TotalPassage: ten,
			Distributions: []Distribution{
				{
					Time:    genesisTime,
					Passage: five,
				},
				{
					Time:    t1,
					Passage: five,
				},
			},
		},
		{
			Address:      addr1,
			TotalPassage: fifteen,
			Distributions: []Distribution{
				{
					Time:    t0,
					Passage: ten,
				},
				{
					Time:    t1,
					Passage: five,
				},
			},
		},
	}, genesisTime, os.Stdout)
	//Output:
	//pasg1ktvsz07ca4amg5cja68pph3t65lj4wdulrz464	10	2
	//	5	MAINNET
	//	5	2021-05-08 16:00:00
	//pasg16ll5l3zdu086ug96cau00k3rllqg9eeyz7ss7t	15	2
	//	10	2021-05-03 16:00:00
	//	5	2021-05-08 16:00:00
}
