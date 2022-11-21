package staking

import (
	"encoding/json"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
)

const (
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
	)

func TestGetVestingAccounts(t *testing.T) {
	// create a decoder
	stringReader := strings.NewReader(STAKING_ACCOUNTS)
	decoder := json.NewDecoder(stringReader)

	var appState = make(map[string]interface{})

	// decode the json file into the map
	err := decoder.Decode(&appState)
	if err != nil {
		t.Log("Error decoding json")
		t.FailNow()
	}

	stakedTokens := GetStakedTokens(appState)

	assert.Equal(t, sdk.NewDec(1000000), stakedTokens)
}