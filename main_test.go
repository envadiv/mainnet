package main

import (
	"bytes"
	"strings"
	"testing"

	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	distribution "github.com/cosmos/cosmos-sdk/x/distribution/types"
	passage "github.com/envadiv/Passage3D/app"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/libs/json"
	"github.com/tendermint/tendermint/types"
)

const testTemplate = `
{
  "genesis_time": "2021-03-26T16:00:00Z",
  "chain_id": "passage-prelaunch-1",
  "initial_height": "1",
  "app_hash": "",
  "app_state": {
    "auth":{},
    "bank":{},
    "distribution":{}
  }
}
`

const testAccounts = `
pasg1zxcfa9nrjamf5kt3q7ruwh2nscmm7su2temk4u,100000,MAINNET,1
pasg1lusdjktpk3f2v33cda5uwnya5qcyv04cwvnkwz,200000.301,2020-06-19,2
pasg1lusdjktpk3f2v33cda5uwnya5qcyv04cwvnkwz,300000.0,MAINNET+1YEAR,2
`

const testCommunityPoolRegenAmount = 100000

func processStrings(tmplJson string, accountsCsv string, communityPoolRegen int) (string, string, error) {
	doc, err := types.GenesisDocFromJSON([]byte(tmplJson))
	if err != nil {
		return "", "", err
	}

	auditOut := new(bytes.Buffer)
	err = Process(doc, strings.NewReader(accountsCsv), communityPoolRegen, auditOut, false)
	if err != nil {
		return "", "", err
	}

	bz, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", "", err
	}

	return string(bz), auditOut.String(), nil
}

func TestProcess(t *testing.T) {
	json, auditOut, err := processStrings(testTemplate, testAccounts, testCommunityPoolRegenAmount)
	require.NoError(t, err)
	require.Equal(t,
		`{
  "genesis_time": "2021-03-26T16:00:00Z",
  "chain_id": "passage-prelaunch-1",
  "initial_height": "1",
  "consensus_params": {
    "block": {
      "max_bytes": "22020096",
      "max_gas": "-1",
      "time_iota_ms": "1000"
    },
    "evidence": {
      "max_age_num_blocks": "100000",
      "max_age_duration": "172800000000000",
      "max_bytes": "1048576"
    },
    "validator": {
      "pub_key_types": [
        "ed25519"
      ]
    },
    "version": {}
  },
  "app_hash": "",
  "app_state": {
    "auth": {
      "params": {
        "max_memo_characters": "0",
        "tx_sig_limit": "0",
        "tx_size_cost_per_byte": "0",
        "sig_verify_cost_ed25519": "0",
        "sig_verify_cost_secp256k1": "0"
      },
      "accounts": [
        {
          "@type": "/cosmos.vesting.v1beta1.PeriodicVestingAccount",
          "base_vesting_account": {
            "base_account": {
              "address": "pasg1lusdjktpk3f2v33cda5uwnya5qcyv04cwvnkwz",
              "pub_key": null,
              "account_number": "0",
              "sequence": "0"
            },
            "original_vesting": [
              {
                "denom": "upasg",
                "amount": "500001301000"
              }
            ],
            "delegated_free": [],
            "delegated_vesting": [],
            "end_time": "1650961098"
          },
          "start_time": "1616774400",
          "vesting_periods": [
            {
              "length": "0",
              "amount": [
                {
                  "denom": "upasg",
                  "amount": "200001301000"
                }
              ]
            },
            {
              "length": "31556952",
              "amount": [
                {
                  "denom": "upasg",
                  "amount": "150000000000"
                }
              ]
            },
            {
              "length": "2629746",
              "amount": [
                {
                  "denom": "upasg",
                  "amount": "150000000000"
                }
              ]
            }
          ]
        },
        {
          "@type": "/cosmos.auth.v1beta1.BaseAccount",
          "address": "pasg1zxcfa9nrjamf5kt3q7ruwh2nscmm7su2temk4u",
          "pub_key": null,
          "account_number": "0",
          "sequence": "0"
        },
        {
          "@type": "/cosmos.auth.v1beta1.ModuleAccount",
          "base_account": {
            "address": "pasg1jv65s3grqf6v6jl3dp4t6c9t9rk99cd8ca0qlm",
            "pub_key": null,
            "account_number": "0",
            "sequence": "0"
          },
          "name": "distribution",
          "permissions": []
        }
      ]
    },
    "bank": {
      "params": {
        "send_enabled": [],
        "default_send_enabled": false
      },
      "balances": [
        {
          "address": "pasg1lusdjktpk3f2v33cda5uwnya5qcyv04cwvnkwz",
          "coins": [
            {
              "denom": "upasg",
              "amount": "500001301000"
            }
          ]
        },
        {
          "address": "pasg1zxcfa9nrjamf5kt3q7ruwh2nscmm7su2temk4u",
          "coins": [
            {
              "denom": "upasg",
              "amount": "99999000000"
            }
          ]
        },
        {
          "address": "pasg1jv65s3grqf6v6jl3dp4t6c9t9rk99cd8ca0qlm",
          "coins": [
            {
              "denom": "upasg",
              "amount": "100000000000"
            }
          ]
        }
      ],
      "supply": [
        {
          "denom": "upasg",
          "amount": "700000301000"
        }
      ],
      "denom_metadata": []
    },
    "distribution": {
      "params": {
        "community_tax": "0",
        "base_proposer_reward": "0",
        "bonus_proposer_reward": "0",
        "withdraw_addr_enabled": false
      },
      "fee_pool": {
        "community_pool": [
          {
            "denom": "upasg",
            "amount": "100000000000.000000000000000000"
          }
        ]
      },
      "delegator_withdraw_infos": [],
      "previous_proposer": "",
      "outstanding_rewards": [],
      "validator_accumulated_commissions": [],
      "validator_historical_rewards": [],
      "validator_current_rewards": [],
      "delegator_starting_infos": [],
      "validator_slash_events": []
    }
  }
}`, json)
	require.Equal(t, `pasg1lusdjktpk3f2v33cda5uwnya5qcyv04cwvnkwz	500001.301	3
	200001.3010	MAINNET
	150000.0	2022-03-26 21:49:12
	150000	2022-04-26 08:18:18
pasg1zxcfa9nrjamf5kt3q7ruwh2nscmm7su2temk4u	99999	1
	99999	MAINNET
`, auditOut)
}

func TestBuildDistrMaccAndBalance(t *testing.T) {

	testRegenAmount := 1000
	testCoins, err := RegenToCoins(NewDecFromInt64(int64(testRegenAmount)))
	require.NoError(t, err)

	maccPerms := passage.GetMaccPerms()
	testMacc := auth.NewEmptyModuleAccount(distribution.ModuleName, maccPerms[distribution.ModuleName]...)

	testBalance := &bank.Balance{
		Coins:   testCoins,
		Address: testMacc.Address,
	}

	distrMacc, distrBalance, err := buildDistrMaccAndBalance(testRegenAmount)
	require.NoError(t, err)

	require.Equal(t, testBalance, distrBalance)
	require.Equal(t, testMacc, distrMacc)
}
