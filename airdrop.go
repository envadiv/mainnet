package main

import (
	"fmt"
	"time"
)

const communityBootstrappingAddr = "pasg18h2lu29rwwkr5nfukghn80vx6gm9xxvv62zgyx"

var one = NewDecFromInt64(1)

var airdropDenyList = []string{
	"pasg1fe64cfja86k4rr8e2qrvc499dulfpxt3anuscs",
	"pasg1yz3c47fkm5e2yd4dtxk30sug4ntcxc9ng2cf69",
	"pasg1xgplygxjyzh7ngzc8wal8w40wr6m7qn2sm9mzx",
	"pasg1ma0pk6g6wsxuksgnefsrtgsceqx6x0nksez2ts",
	"pasg1p62d756unkyn83wa9ls8dzmlg3d2u24fz65lt2",
	"pasg1wts9f35khx5r94p3rdwgk3evv0v06ewrfl90th",
	"pasg1xv7pkdjtumtvakyfgwc7daknzxrdk8yhh9kng7",
}

func AirdropPassage3DForMinFees(accMap map[string]Account, genesisTime time.Time) error {
	var totalAirdrop Dec
	var err error
	airdropDenyMap := make(map[string]bool)

	for _, addr := range airdropDenyList {
		airdropDenyMap[addr] = true
	}

	for addr, acc := range accMap {
		if addr == communityBootstrappingAddr || airdropDenyMap[addr] {
			continue
		}

		accMap[addr], err = airdrop1(acc, genesisTime)
		if err != nil {
			return err
		}

		totalAirdrop, err = totalAirdrop.Add(one)
		if err != nil {
			return err
		}
	}

	bootstrapAcc := accMap[communityBootstrappingAddr]
	bootstrapAcc.TotalPassage, err = bootstrapAcc.TotalPassage.Sub(totalAirdrop)
	if err != nil {
		return err
	}

	if len(bootstrapAcc.Distributions) < 1 || !bootstrapAcc.Distributions[0].Time.Equal(genesisTime) {
		return fmt.Errorf("problem with community bootstrap account")
	}

	bootstrapAcc.Distributions[0].Passage, err = bootstrapAcc.Distributions[0].Passage.Sub(totalAirdrop)
	if err != nil {
		return err
	}

	accMap[communityBootstrappingAddr] = bootstrapAcc

	return nil
}

func airdrop1(acc Account, genesisTime time.Time) (Account, error) {
	var err error
	acc.TotalPassage, err = acc.TotalPassage.Add(one)
	if err != nil {
		return Account{}, err
	}

	if len(acc.Distributions) < 1 {
		return Account{}, fmt.Errorf("expected at least 1 distribution")
	}

	if acc.Distributions[0].Time.Equal(genesisTime) {
		acc.Distributions[0].Passage, err = acc.Distributions[0].Passage.Add(one)
		if err != nil {
			return Account{}, err
		}
	} else {
		genesisDist := Distribution{Time: genesisTime, Passage: one}
		acc.Distributions = append([]Distribution{genesisDist}, acc.Distributions...)
	}

	return acc, nil
}
