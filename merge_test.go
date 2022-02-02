package main

import (
	"reflect"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestMergeAccounts(t *testing.T) {
	addr0 := sdk.AccAddress("abcdefg012345")
	addr1 := sdk.AccAddress("012345abcdefg")
	five, err := NewDecFromString("5")
	require.NoError(t, err)
	ten, err := NewDecFromString("10")
	require.NoError(t, err)
	fifteen, err := NewDecFromString("15")
	require.NoError(t, err)
	twenty, err := NewDecFromString("20")
	require.NoError(t, err)
	thirty, err := NewDecFromString("30")
	require.NoError(t, err)
	forty, err := NewDecFromString("40")
	require.NoError(t, err)

	time0, err := time.Parse(time.RFC3339, "2021-05-21T00:00:00Z")
	require.NoError(t, err)

	time1, err := time.Parse(time.RFC3339, "2021-06-21T00:00:00Z")
	require.NoError(t, err)

	time2, err := time.Parse(time.RFC3339, "2021-07-21T00:00:00Z")
	require.NoError(t, err)

	time3, err := time.Parse(time.RFC3339, "2021-08-21T00:00:00Z")
	require.NoError(t, err)

	time4, err := time.Parse(time.RFC3339, "2021-09-21T00:00:00Z")
	require.NoError(t, err)

	tests := []struct {
		name     string
		accounts []Account
		want     []Account
		wantErr  bool
	}{
		{
			"merging two accounts basic",
			[]Account{
				{
					Address:      addr0,
					TotalPassage: ten,
					Distributions: []Distribution{
						{
							Time:    time0,
							Passage: ten,
						},
					},
				},
				{
					Address:      addr0,
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
			},
			[]Account{
				{
					Address:      addr0,
					TotalPassage: twenty,
					Distributions: []Distribution{
						{
							Time:    time0,
							Passage: fifteen,
						},
						{
							Time:    time1,
							Passage: five,
						},
					},
				},
			},
			false,
		},
		{
			"merging two accounts, no common distr times",
			[]Account{
				{
					Address:      addr0,
					TotalPassage: twenty,
					Distributions: []Distribution{
						{
							Time:    time0,
							Passage: ten,
						},
						{
							Time:    time2,
							Passage: ten,
						},
					},
				},
				{
					Address:      addr0,
					TotalPassage: ten,
					Distributions: []Distribution{
						{
							Time:    time1,
							Passage: five,
						},
						{
							Time:    time3,
							Passage: five,
						},
					},
				},
			},
			[]Account{
				{
					Address:      addr0,
					TotalPassage: thirty,
					Distributions: []Distribution{
						{
							Time:    time0,
							Passage: ten,
						},
						{
							Time:    time1,
							Passage: five,
						},
						{
							Time:    time2,
							Passage: ten,
						},
						{
							Time:    time3,
							Passage: five,
						},
					},
				},
			},
			false,
		},
		{
			"merge accounts with all different addresses",
			[]Account{
				{
					Address:      addr0,
					TotalPassage: twenty,
					Distributions: []Distribution{
						{
							Time:    time0,
							Passage: ten,
						},
						{
							Time:    time3,
							Passage: ten,
						},
					},
				},
				{
					Address:      addr1,
					TotalPassage: thirty,
					Distributions: []Distribution{
						{
							Time:    time0,
							Passage: ten,
						},
						{
							Time:    time1,
							Passage: ten,
						},
						{
							Time:    time2,
							Passage: ten,
						},
					},
				},
			},
			[]Account{
				{
					Address:      addr0,
					TotalPassage: twenty,
					Distributions: []Distribution{
						{
							Time:    time0,
							Passage: ten,
						},
						{
							Time:    time3,
							Passage: ten,
						},
					},
				},
				{
					Address:      addr1,
					TotalPassage: thirty,
					Distributions: []Distribution{
						{
							Time:    time0,
							Passage: ten,
						},
						{
							Time:    time1,
							Passage: ten,
						},
						{
							Time:    time2,
							Passage: ten,
						},
					},
				},
			},
			false,
		},
		{
			"merge complex accounts with overlapping distributions",
			[]Account{
				{
					Address:      addr0,
					TotalPassage: fifteen,
					Distributions: []Distribution{
						{
							Time:    time1,
							Passage: five,
						},
						{
							Time:    time2,
							Passage: five,
						},
						{
							Time:    time3,
							Passage: five,
						},
					},
				},
				{
					Address:      addr0,
					TotalPassage: twenty,
					Distributions: []Distribution{
						{
							Time:    time0,
							Passage: ten,
						},
						{
							Time:    time1,
							Passage: five,
						},
						{
							Time:    time2,
							Passage: five,
						},
					},
				},
				{
					Address:      addr0,
					TotalPassage: five,
					Distributions: []Distribution{
						{
							Time:    time1,
							Passage: five,
						},
					},
				},
				{
					Address:      addr1,
					TotalPassage: thirty,
					Distributions: []Distribution{
						{
							Time:    time4,
							Passage: thirty,
						},
					},
				},
			},
			[]Account{
				{
					Address:      addr0,
					TotalPassage: forty,
					Distributions: []Distribution{
						{
							Time:    time0,
							Passage: ten,
						},
						{
							Time:    time1,
							Passage: fifteen,
						},
						{
							Time:    time2,
							Passage: ten,
						},
						{
							Time:    time3,
							Passage: five,
						},
					},
				},
				{
					Address:      addr1,
					TotalPassage: thirty,
					Distributions: []Distribution{
						{
							Time:    time4,
							Passage: thirty,
						},
					},
				},
			},
			false,
		},
		{
			"error on account with no distributions",
			[]Account{
				{
					Address:       addr0,
					TotalPassage:  twenty,
					Distributions: []Distribution{},
				},
				{
					Address:      addr1,
					TotalPassage: thirty,
					Distributions: []Distribution{
						{
							Time:    time0,
							Passage: ten,
						},
						{
							Time:    time1,
							Passage: ten,
						},
						{
							Time:    time2,
							Passage: ten,
						},
					},
				},
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAccMap, err := MergeAccounts(tt.accounts)
			if (err != nil) != tt.wantErr {
				t.Errorf("MergeAccounts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			got := SortAccounts(gotAccMap)

			require.Equal(t, len(tt.want), len(got))
			for i, gotAcc := range got {
				wantAcc := tt.want[i]
				require.Equal(t, wantAcc.Address, gotAcc.Address)
				RequireAccountEqual(t, wantAcc, gotAcc)
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MergeAccounts() got = %v, want %v", got, tt.want)
			}
			for _, acc := range got {
				require.NoError(t, acc.Validate())
			}
		})
	}
}
