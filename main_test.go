package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"io"
	"strings"

	"testing"

	"github.com/stretchr/testify/assert"

	mintModule "github.com/brianosaurus/challenge2/mint"
	stakingModule "github.com/brianosaurus/challenge2/staking"
	vestingModule "github.com/brianosaurus/challenge2/vesting"
)

const (
	BANK_BALANCES = 
`{
	"bank": {
		"balances": [
			{
				"address": "umee1qqq8fdsrcsz4jnlvrcqa6ds2ruflgrgeguyeh0",
				"coins": [
					{
						"denom": "uumee",
						"amount": "8333000000"
					}
				]
			},
			{
				"address": "umee1qqqk0gxu4he52m0t2w6f6vfag6uvyaegprmj58",
				"coins": [
					{
						"denom": "uumee",
						"amount": "7500000000"
					}
				]
			},
			{
				"address": "umee1qqp7t9xsw4m2qsjetspahrgz9q8h02n7wqkcf2",
				"coins": [
					{
						"denom": "uumee",
						"amount": "7143000000"
					}
				]
			}
		]
	}
}`

	AUTH_VESTING_ACCOUNTS = 
`{
	"auth": {
 		"accounts": [
 			{
 				"@type": "/cosmos.vesting.v1beta1.DelayedVestingAccount",
 				"base_vesting_account": {
 					"base_account": {
 						"address": "umee1agzky2ak6xs5vve3c2wzjtqdq7fwadcgj2mxf9",
 						"pub_key": null,
 						"account_number": "0",
 						"sequence": "0"
 					},
 					"original_vesting": [
 						{
 							"denom": "uumee",
 							"amount": "309282000000"
 						}
 					],
 					"delegated_free": [],
 					"delegated_vesting": [],
 					"end_time": "1676480400"
 				}
			},
 			{
 				"@type": "/cosmos.vesting.v1beta1.ContinuousVestingAccount",
 				"base_vesting_account": {
 					"base_account": {
 						"address": "umee1wqr08242ysrepqgzm6q0mn7ndcnjlsf6vdxd0v",
 						"pub_key": null,
 						"account_number": "0",
 						"sequence": "0"
 					},
 					"original_vesting": [
 						{
 							"denom": "uumee",
 							"amount": "11250000000000"
 						}
 					],
 					"delegated_free": [],
 					"delegated_vesting": [],
 					"end_time": "1739638800"
 				},
 				"start_time": "1660582800"
 			}
 		]
 	}
 }`

	STAKING_ACCOUNTS = 
`{
	"genutil": {
		"gen_txs": [
			{
				"body": {
					"messages": [
						{
							"@type": "/cosmos.staking.v1beta1.MsgCreateValidator",
							"description": {
								"moniker": "0base.vc",
								"identity": "67A577430DBBCEE0",
								"website": "https://0base.vc",
								"security_contact": "0@0base.vc",
								"details": "0base.vc is a validator who doesn't trust any blockchain; we validate it ourselves."
							},
							"commission": {
								"rate": "0.020000000000000000",
								"max_rate": "0.100000000000000000",
								"max_change_rate": "0.010000000000000000"
							},
							"min_self_delegation": "1",
							"delegator_address": "umee1n3mhyp9fvcmuu8l0q8qvjy07x0rql8q4dtsqwh",
							"validator_address": "umeevaloper1n3mhyp9fvcmuu8l0q8qvjy07x0rql8q4d0h0la",
							"pubkey": {
								"@type": "/cosmos.crypto.ed25519.PubKey",
								"key": "PZ3V+lSY4TFFMvr0drY4ARBKvh/ZHUgW0ByL45yyQUk="
							},
							"value": {
								"denom": "uumee",
								"amount": "1000000"
							}
						},
						{
							"@type": "/gravity.v1.MsgSetOrchestratorAddress",
							"validator": "umeevaloper1n3mhyp9fvcmuu8l0q8qvjy07x0rql8q4d0h0la",
							"orchestrator": "umee1fx2la0tx67xxrnlzf03khk2fjzs9kfyqvl67y9",
							"eth_address": "0x6D588c5ddB0FfF0C2723e0cFDc019b885DaBa474"
						}
					],
					"memo": "8ba2333604c540dc3d9dfe13c619e5c78144d61d@192.168.2.133:26656",
					"timeout_height": "0",
					"extension_options": [],
					"non_critical_extension_options": []
				},
				"auth_info": {
					"signer_infos": [
						{
							"public_key": {
								"@type": "/cosmos.crypto.secp256k1.PubKey",
								"key": "ArscgfwUlatB4SKqaROqnzMzvj95XgAbNMy2Tp8bLAQ5"
							},
							"mode_info": {
								"single": {
									"mode": "SIGN_MODE_LEGACY_AMINO_JSON"
								}
							},
							"sequence": "0"
						}
					],
					"fee": {
						"amount": [],
						"gas_limit": "200000",
						"payer": "",
						"granter": ""
					}
				},
				"signatures": [
					"MiczRmpNRSiHbPlxoTDDtmEafLJfBwJuKIwEb5kDHWd46jDpOTqUQKPsnlRaRSEaBH3PGuFTp5S1J0brWjxCQQ=="
				]
			}
		]
	}
}`

	PARAMS =
`{
	"params": {
		"mint_denom": "uumee",
		"inflation_rate_change": "1.000000000000000000",
		"inflation_max": "0.140000000000000000",
		"inflation_min": "0.070000000000000000",
		"goal_bonded": "0.330000000000000000",
		"blocks_per_year": "4360000"
	}
}`	

MINTER =
`{
	"mint": {
		"minter": {
			"inflation": "0.130000000000000000",
			"annual_provisions": "0.000000000000000000"
		}
	}
}`
	)


func TestWriteCSV(t *testing.T) {
	// create a decoder
	stringReader := strings.NewReader(MINTER)
	decoder := json.NewDecoder(stringReader)

	var minterJson = make(map[string]interface{})

	// decode the json file into the map
	err := decoder.Decode(&minterJson)
	if err != nil {
		t.Log("Error decoding json")
		t.FailNow()
	}

	stringReader = strings.NewReader(PARAMS)
	decoder = json.NewDecoder(stringReader)

	var paramsJson = make(map[string]interface{})

	// decode the json file into the map
	err = decoder.Decode(&paramsJson)
	if err != nil {
		t.Log("Error decoding json")
		t.FailNow()
	}

	minterJson["mint"].(map[string]interface{})["params"] = paramsJson["params"]

	params, minter := mintModule.GetParamsAndMinter(minterJson)

	var appState = make(map[string]interface{})

	stringReader = strings.NewReader(AUTH_VESTING_ACCOUNTS)
	decoder = json.NewDecoder(stringReader)

	// decode the json file into the map
	err = decoder.Decode(&appState)
	if err != nil {
		t.Log("Error decoding json")
		t.FailNow()
	}

	continuousVestingAccounts, delayedVestingAccounts := vestingModule.GetVestingAccounts(appState)

	stringReader = strings.NewReader(BANK_BALANCES)
	decoder = json.NewDecoder(stringReader)

	appState = make(map[string]interface{})

	// decode the json file into the map
	err = decoder.Decode(&appState)
	if err != nil {
		t.Log("Error decoding json")
		t.FailNow()
	}

	totalSupply, vestingOnDays := vestingModule.GetTotalSupplyAndVestingSchedule(appState, continuousVestingAccounts, delayedVestingAccounts)

	// create a decoder
	stringReader = strings.NewReader(STAKING_ACCOUNTS)
	decoder = json.NewDecoder(stringReader)

	appState = make(map[string]interface{})

	// decode the json file into the map
	err = decoder.Decode(&appState)
	if err != nil {
		t.Log("Error decoding json")
		t.FailNow()
	}

	stakedTokens := stakingModule.GetStakedTokens(appState)

	var buf bytes.Buffer
	bufWriter := io.Writer(&buf)
	writer := csv.NewWriter(bufWriter)
	WriteCSV(writer, vestingOnDays, totalSupply, stakedTokens, minter, params)

	bufString := strings.Split(buf.String(), "\n")

	// I reaize this is obnoxiously long ... short on time to do this better
	assert.Equal(t, "Days Since Genesis Analyzed,Tokens Unvesting,Inflation,Staking Rewards,Circulating Supply,Total Supply", bufString[0])
	assert.Equal(t, "0,0,0.130000000000000000,0,1240202470400,11582258000000", bufString[1])
	assert.Equal(t, "815,12295065600,0.132975646103044134,54508050696,11636766050696,11636766050696", bufString[len(bufString)-2])
}