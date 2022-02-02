package main

import (
	"testing"
	"time"

	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/stretchr/testify/require"
)

func TestPassage3DToCoins(t *testing.T) {
	one, err := NewDecFromString("1")
	require.NoError(t, err)

	coins, err := Passage3DToCoins(one)
	require.NoError(t, err)
	require.Equal(t, "1000000upasg", coins.String())
}

func TestToCosmosAccount(t *testing.T) {
	addr0 := sdk.AccAddress("abcdefg012345")
	addr1 := sdk.AccAddress("012345abcdefg")
	five, err := NewDecFromString("5")
	require.NoError(t, err)
	ten, err := NewDecFromString("10")
	require.NoError(t, err)
	thirty, err := NewDecFromString("30")
	require.NoError(t, err)

	time0, err := time.Parse(time.RFC3339, "2021-05-21T00:00:00Z")
	require.NoError(t, err)

	time1, err := time.Parse(time.RFC3339, "2021-06-21T00:00:00Z")
	require.NoError(t, err)

	time2, err := time.Parse(time.RFC3339, "2021-07-21T00:00:00Z")
	require.NoError(t, err)

	time3, err := time.Parse(time.RFC3339, "2021-08-21T00:00:00Z")
	require.NoError(t, err)

	tests := []struct {
		name    string
		acc     Account
		want    auth.AccountI
		want1   *bank.Balance
		wantErr bool
	}{
		{
			"one distribution at mainnet",
			Account{
				Address:      addr0,
				TotalPassage: ten,
				Distributions: []Distribution{
					{
						Time:    time0,
						Passage: ten,
					},
				},
			},
			&auth.BaseAccount{
				Address: addr0.String(),
			},
			&bank.Balance{
				Address: addr0.String(),
				Coins:   sdk.NewCoins(sdk.NewInt64Coin(UPassageDenom, 10000000)),
			},
			false,
		},
		{
			"one distribution at mainnet bad balance",
			Account{
				Address:      addr0,
				TotalPassage: ten,
				Distributions: []Distribution{
					{
						Time:    time0,
						Passage: five,
					},
				},
			},
			nil,
			nil,
			true,
		},
		{
			"one distribution after mainnet",
			Account{
				Address:      addr1,
				TotalPassage: five,
				Distributions: []Distribution{
					{
						Time:    time1,
						Passage: five,
					},
				},
			},
			&vesting.DelayedVestingAccount{
				BaseVestingAccount: &vesting.BaseVestingAccount{
					BaseAccount: &auth.BaseAccount{
						Address: addr1.String(),
					},
					OriginalVesting: sdk.NewCoins(sdk.NewInt64Coin(UPassageDenom, 5000000)),
					EndTime:         time1.Unix(),
				},
			},
			&bank.Balance{
				Address: addr1.String(),
				Coins:   sdk.NewCoins(sdk.NewInt64Coin(UPassageDenom, 5000000)),
			},
			false,
		},
		{
			"two distributions",
			Account{
				Address:      addr1,
				TotalPassage: ten,
				Distributions: []Distribution{
					{
						Time:    time0,
						Passage: five,
					},
					{
						Time:    time1,
						Passage: five,
					},
				},
			},
			&vesting.PeriodicVestingAccount{
				BaseVestingAccount: &vesting.BaseVestingAccount{
					BaseAccount: &auth.BaseAccount{
						Address: addr1.String(),
					},
					OriginalVesting: sdk.NewCoins(sdk.NewInt64Coin(UPassageDenom, 10000000)),
					EndTime:         time1.Unix(),
				},
				StartTime: time0.Unix(),
				VestingPeriods: []vesting.Period{
					{
						Length: 0,
						Amount: sdk.NewCoins(sdk.NewInt64Coin(UPassageDenom, 5000000)),
					},
					{
						Length: int64(time1.Sub(time0).Seconds()),
						Amount: sdk.NewCoins(sdk.NewInt64Coin(UPassageDenom, 5000000)),
					},
				},
			},
			&bank.Balance{
				Address: addr1.String(),
				Coins:   sdk.NewCoins(sdk.NewInt64Coin(UPassageDenom, 10000000)),
			},
			false,
		},
		{
			"four distributions",
			Account{
				Address:      addr1,
				TotalPassage: thirty,
				Distributions: []Distribution{
					{
						Time:    time0,
						Passage: five,
					},
					{
						Time:    time1,
						Passage: ten,
					},
					{
						Time:    time2,
						Passage: five,
					},
					{
						Time:    time3,
						Passage: ten,
					},
				},
			},
			&vesting.PeriodicVestingAccount{
				BaseVestingAccount: &vesting.BaseVestingAccount{
					BaseAccount: &auth.BaseAccount{
						Address: addr1.String(),
					},
					OriginalVesting: sdk.NewCoins(sdk.NewInt64Coin(UPassageDenom, 30000000)),
					EndTime:         time3.Unix(),
				},
				StartTime: time0.Unix(),
				VestingPeriods: []vesting.Period{
					{
						Length: 0,
						Amount: sdk.NewCoins(sdk.NewInt64Coin(UPassageDenom, 5000000)),
					},
					{
						Length: int64(time1.Sub(time0).Seconds()),
						Amount: sdk.NewCoins(sdk.NewInt64Coin(UPassageDenom, 10000000)),
					},
					{
						Length: int64(time2.Sub(time1).Seconds()),
						Amount: sdk.NewCoins(sdk.NewInt64Coin(UPassageDenom, 5000000)),
					},
					{
						Length: int64(time3.Sub(time2).Seconds()),
						Amount: sdk.NewCoins(sdk.NewInt64Coin(UPassageDenom, 10000000)),
					},
				},
			},
			&bank.Balance{
				Address: addr1.String(),
				Coins:   sdk.NewCoins(sdk.NewInt64Coin(UPassageDenom, 30000000)),
			},
			false,
		},
		{
			"two distributions bad balance",
			Account{
				Address:      addr0,
				TotalPassage: ten,
				Distributions: []Distribution{
					{
						Time:    time0,
						Passage: five,
					},
					{
						Time:    time1,
						Passage: ten,
					},
				},
			},
			nil,
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := ToCosmosAccount(tt.acc, time0)
			if tt.wantErr {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.want, got)
			require.Equal(t, tt.want1, got1)
			require.NoError(t, ValidateVestingAccount(got))
		})
	}
}

func TestRecordToAccount(t *testing.T) {
	addr0 := sdk.AccAddress("abcdefg012345")

	five, err := NewDecFromString("5")
	require.NoError(t, err)
	five000001, err := NewDecFromString("5.000001")
	require.NoError(t, err)
	ten000001, err := NewDecFromString("10.000001")
	require.NoError(t, err)
	seven500001, err := NewDecFromString("7.500001")
	require.NoError(t, err)
	two5, err := NewDecFromString("2.5")
	require.NoError(t, err)
	three00000, err := NewDecFromString("300000.0")
	require.NoError(t, err)
	one50000, err := NewDecFromString("150000")
	require.NoError(t, err)

	genesisTime, err := time.Parse(time.RFC3339, "2021-04-08T00:00:00Z")
	require.NoError(t, err)

	oneMonthAfterGenesis := genesisTime.Add(OneMonth)

	start0, err := time.Parse(time.RFC3339, "2021-01-05T00:00:00Z")
	require.NoError(t, err)

	twoMonthsBeforeGenesis := genesisTime.Add(-OneMonth * 2)

	tests := []struct {
		name    string
		record  Record
		want    Account
		wantErr bool
	}{
		{
			"no dists, no start time",
			Record{
				Address:     addr0,
				TotalAmount: five,
			},
			Account{},
			true,
		},
		{
			"no dists, genesis start time",
			Record{
				Address:     addr0,
				TotalAmount: five,
				StartTime:   genesisTime,
			},
			Account{},
			true,
		},
		{
			"one dist at genesis",
			Record{
				Address:                addr0,
				TotalAmount:            five,
				StartTime:              genesisTime,
				NumWeeklyDistributions: 1,
			},
			Account{
				Address:      addr0,
				TotalPassage: five,
				Distributions: []Distribution{
					{
						Time:    genesisTime,
						Passage: five,
					},
				},
			},
			false,
		},
		{
			"two dists from genesis",
			Record{
				Address:                addr0,
				TotalAmount:            ten000001,
				StartTime:              genesisTime,
				NumWeeklyDistributions: 2,
			},
			Account{
				Address:      addr0,
				TotalPassage: ten000001,
				Distributions: []Distribution{
					{
						Time:    genesisTime,
						Passage: five000001,
					},
					{
						Time:    oneMonthAfterGenesis,
						Passage: five,
					},
				},
			},
			false,
		},
		{
			"two dists all from before genesis",
			Record{
				Address:                addr0,
				TotalAmount:            ten000001,
				StartTime:              start0,
				NumWeeklyDistributions: 2,
			},
			Account{
				Address:      addr0,
				TotalPassage: ten000001,
				Distributions: []Distribution{
					{
						Time:    genesisTime,
						Passage: ten000001,
					},
				},
			},
			false,
		},
		{
			"four dists starting before genesis",
			Record{
				Address:                addr0,
				TotalAmount:            ten000001,
				StartTime:              twoMonthsBeforeGenesis,
				NumWeeklyDistributions: 4,
			},
			Account{
				Address:      addr0,
				TotalPassage: ten000001,
				Distributions: []Distribution{
					{
						Time:    genesisTime,
						Passage: seven500001,
					},
					{
						Time:    oneMonthAfterGenesis,
						Passage: two5,
					},
				},
			},
			false,
		},
		{
			"one dist after genesis",
			Record{
				Address:                addr0,
				TotalAmount:            ten000001,
				StartTime:              oneMonthAfterGenesis,
				NumWeeklyDistributions: 1,
			},
			Account{
				Address:      addr0,
				TotalPassage: ten000001,
				Distributions: []Distribution{
					{
						Time:    oneMonthAfterGenesis,
						Passage: ten000001,
					},
				},
			},
			false,
		},
		{
			"two dists after genesis",
			Record{
				Address:                addr0,
				TotalAmount:            ten000001,
				StartTime:              oneMonthAfterGenesis,
				NumWeeklyDistributions: 2,
			},
			Account{
				Address:      addr0,
				TotalPassage: ten000001,
				Distributions: []Distribution{
					{
						Time:    oneMonthAfterGenesis,
						Passage: five000001,
					},
					{
						Time:    oneMonthAfterGenesis.Add(OneMonth),
						Passage: five,
					},
				},
			},
			false,
		},
		{
			"two dists after genesis, 300000.0",
			Record{
				Address:                addr0,
				TotalAmount:            three00000,
				StartTime:              oneMonthAfterGenesis,
				NumWeeklyDistributions: 2,
			},
			Account{
				Address:      addr0,
				TotalPassage: three00000,
				Distributions: []Distribution{
					{
						Time:    oneMonthAfterGenesis,
						Passage: one50000,
					},
					{
						Time:    oneMonthAfterGenesis.Add(OneMonth),
						Passage: one50000,
					},
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RecordToAccount(tt.record, genesisTime)
			if tt.wantErr {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}
			require.NotNil(t, got)
			RequireAccountEqual(t, tt.want, got)
			require.NoError(t, got.Validate())
		})
	}
}

func Test_distAmountAndDust(t *testing.T) {
	five01, err := NewDecFromString("5.01")
	require.NoError(t, err)
	ten02, err := NewDecFromString("10.02")
	require.NoError(t, err)
	ten020001, err := NewDecFromString("10.020001")
	require.NoError(t, err)
	point000001, err := NewDecFromString("0.000001")
	require.NoError(t, err)
	ten, err := NewDecFromString("10")
	require.NoError(t, err)
	three333333, err := NewDecFromString("3.333333")
	require.NoError(t, err)
	three00000, err := NewDecFromString("300000.0")
	require.NoError(t, err)
	twelve500, err := NewDecFromString("12500")
	require.NoError(t, err)

	type args struct {
		amount  Dec
		numDist int
	}
	tests := []struct {
		name           string
		args           args
		wantDistAmount Dec
		wantDust       Dec
		wantErr        bool
	}{
		{
			name: "0 dist",
			args: args{
				amount:  ten02,
				numDist: 0,
			},
			wantErr: true,
		},
		{
			name: "1 dist",
			args: args{
				amount:  ten02,
				numDist: 1,
			},
			wantDistAmount: ten02,
			wantDust:       Dec{},
			wantErr:        false,
		},
		{
			name: "2 dist, no dust",
			args: args{
				amount:  ten02,
				numDist: 2,
			},
			wantDistAmount: five01,
			wantDust:       Dec{},
			wantErr:        false,
		},
		{
			name: "2 dist, dust",
			args: args{
				amount:  ten020001,
				numDist: 2,
			},
			wantDistAmount: five01,
			wantDust:       point000001,
			wantErr:        false,
		},
		{
			name: "3 dists, dust",
			args: args{
				amount:  ten,
				numDist: 3,
			},
			wantDistAmount: three333333,
			wantDust:       point000001,
			wantErr:        false,
		},
		{
			name: "24 dists, no dust",
			args: args{
				amount:  three00000,
				numDist: 24,
			},
			wantDistAmount: twelve500,
			wantDust:       Dec{},
			wantErr:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDistAmount, gotDust, err := distAmountAndDust(tt.args.amount, tt.args.numDist)
			if (err != nil) != tt.wantErr {
				t.Errorf("distAmountAndDust() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			RequireDecEqual(t, tt.wantDistAmount, gotDistAmount, "distAmount")
			RequireDecEqual(t, tt.wantDust, gotDust, "distAmount")
		})
	}
}
