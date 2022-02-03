package main

import (
	"strings"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/require"
)

func TestParseAccountsCsv(t *testing.T) {
	genesisTime, err := time.Parse(time.RFC3339, "2021-04-08T16:00:00Z")
	require.NoError(t, err)

	records, err := ParseAccountsCsv(strings.NewReader(`
pasg10rk2v8pxjnldtxuy9ds0s5na9qjcmh5ymplz87,100000,MAINNET,1
pasg1lusdjktpk3f2v33cda5uwnya5qcyv04cwvnkwz,200000.301,2020-06-19,24
pasg1lusdjktpk3f2v33cda5uwnya5qcyv04cwvnkwz,300000.0,MAINNET+1YEAR,24`), genesisTime, false)
	require.NoError(t, err)

	addr0, err := sdk.AccAddressFromBech32("pasg10rk2v8pxjnldtxuy9ds0s5na9qjcmh5ymplz87")
	require.NoError(t, err)

	addr1, err := sdk.AccAddressFromBech32("pasg1lusdjktpk3f2v33cda5uwnya5qcyv04cwvnkwz")
	require.NoError(t, err)

	addr2, err := sdk.AccAddressFromBech32("pasg1lusdjktpk3f2v33cda5uwnya5qcyv04cwvnkwz")
	require.NoError(t, err)

	t0, err := time.Parse(time.RFC3339, "2020-06-19T00:00:00Z")
	require.NoError(t, err)

	d0 := NewDecFromInt64(100000)
	d1, err := NewDecFromString("200000.301")
	require.NoError(t, err)
	d2 := NewDecFromInt64(300000)

	require.True(t, records[0].Equal(Record{
		Address:                 addr0,
		TotalAmount:             d0,
		StartTime:               genesisTime,
		NumWeeklyDistributions: 1,
	}))

	require.True(t, records[1].Equal(Record{
		Address:                 addr1,
		TotalAmount:             d1,
		StartTime:               t0,
		NumWeeklyDistributions: 24,
	}))

	require.True(t, records[2].Equal(Record{
		Address:                 addr2,
		TotalAmount:             d2,
		StartTime:               genesisTime.Add(OneYear),
		NumWeeklyDistributions: 24,
	}))
}
