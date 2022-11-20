package vesting

import (
	"encoding/json"
	"strings"
	big "math/big"

	"testing"

	"github.com/stretchr/testify/assert"
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
	)


func TestGetVestingAccounts(t *testing.T) {

	// create a decoder
	stringReader := strings.NewReader(AUTH_VESTING_ACCOUNTS)
	decoder := json.NewDecoder(stringReader)

	var appState = make(map[string]interface{})

	// decode the json file into the map
	err := decoder.Decode(&appState)
	if err != nil {
		t.Log("Error decoding json")
		t.FailNow()
	}

	continuousVestingAccounts, delayedVestingAccounts := GetVestingAccounts(appState)

	assert.Equal(t, 1, len(continuousVestingAccounts))
	assert.Equal(t, 1, len(delayedVestingAccounts))

	continuousVestingAccount := continuousVestingAccounts["umee1wqr08242ysrepqgzm6q0mn7ndcnjlsf6vdxd0v"]
	delayedVestingAccount := delayedVestingAccounts["umee1agzky2ak6xs5vve3c2wzjtqdq7fwadcgj2mxf9"]

	assert.Equal(t, big.NewInt(11250000000000), continuousVestingAccount.OriginalVesting[0].Amount.BigInt())
	assert.Equal(t, big.NewInt(309282000000), delayedVestingAccount.OriginalVesting[0].Amount.BigInt())
	assert.Equal(t, int64(1739638800), continuousVestingAccount.EndTime)
	assert.Equal(t, int64(1676480400), delayedVestingAccount.EndTime)
	assert.Equal(t, int64(1660582800), continuousVestingAccount.StartTime)
	assert.Equal(t, "umee1wqr08242ysrepqgzm6q0mn7ndcnjlsf6vdxd0v", continuousVestingAccount.BaseVestingAccount.BaseAccount.Address)
	assert.Equal(t, "umee1agzky2ak6xs5vve3c2wzjtqdq7fwadcgj2mxf9", delayedVestingAccount.BaseVestingAccount.BaseAccount.Address)
	assert.Equal(t, "uumee", continuousVestingAccount.OriginalVesting[0].Denom)
	assert.Equal(t, "uumee", delayedVestingAccount.OriginalVesting[0].Denom)
}

func TestGetTotalSupplyAndVestingSchedule(t *testing.T) {
	TheTime = 1668956141 // make this static for testing

	// create a decoder
	stringReader := strings.NewReader(AUTH_VESTING_ACCOUNTS)
	decoder := json.NewDecoder(stringReader)

	var appState = make(map[string]interface{})

	// decode the json file into the map
	err := decoder.Decode(&appState)
	if err != nil {
		t.Log("Error decoding json")
		t.FailNow()
	}

	continuousVestingAccounts, delayedVestingAccounts := GetVestingAccounts(appState)

	stringReader = strings.NewReader(BANK_BALANCES)
	decoder = json.NewDecoder(stringReader)

	appState = make(map[string]interface{})

	// decode the json file into the map
	err = decoder.Decode(&appState)
	if err != nil {
		t.Log("Error decoding json")
		t.FailNow()
	}

	totalSupply, vestingOnDays := GetTotalSupplyAndVestingSchedule(appState, continuousVestingAccounts, delayedVestingAccounts)

	// t.Log(vestingOnDays)
	// t.Log(totalSupply)
	
	bigTotalSupply, success := totalSupply.BigInt().SetString("11582258000000000000000000000000", 10)
	if !success {
		t.Log("Error converting string to big int")
		t.FailNow()
	}

	// day zero on the vesting schedule is really day 1 of vesting. How computers count which is to say
	// indexes start at zero.
	vestingStart, success := totalSupply.BigInt().SetString("12295065600000000000000000000", 10)
	if !success {
		t.Log("Error converting string to big int")
		t.FailNow()
	}

	vestingEnd, success := totalSupply.BigInt().SetString("12295065600000000000000000000", 10)
	if !success {
		t.Log("Error converting string to big int")
		t.FailNow()
	}

	assert.Equal(t, bigTotalSupply, totalSupply.BigInt())
	assert.Equal(t, 818, len(*vestingOnDays))
	assert.Equal(t, vestingStart, ((*vestingOnDays)[0].BigInt()))
	assert.Equal(t, vestingEnd, ((*vestingOnDays)[817].BigInt()))

	// make sure it is vesting for the proper amount of days
	address := "umee1wqr08242ysrepqgzm6q0mn7ndcnjlsf6vdxd0v"
	assert.Equal(t, int64(len(*vestingOnDays)), (continuousVestingAccounts[address].EndTime - TheTime) / 86400)
}

