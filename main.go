package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"

	sdkTypes "github.com/cosmos/cosmos-sdk/types"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingTypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	"github.com/cosmos/cosmos-sdk/codec"
)

func NewBaseAccount(baseAccountJson map[string]interface{}) *authTypes.BaseAccount {
	if baseAccountJson["address"] == nil {
		panic(fmt.Sprintln("Error: address is nil", baseAccountJson))
	}
	address := baseAccountJson["address"].(string)

	// if baseAccountJson["account_number"] == nil {
	// 	panic(fmt.Sprintln("Error: account_number is nil", baseAccountJson))
	// }
  accountNumber, err := strconv.ParseUint(baseAccountJson["account_number"].(string), 10, 64)
	if err != nil {
		panic(fmt.Sprintln("parseUint accountNumber:", baseAccountJson, err))
	}

	// if baseAccountJson["sequence"] == nil {
	// 	panic(fmt.Sprintln("Error: sequence is nil", baseAccountJson))
	// }
	sequence, err := strconv.ParseUint(baseAccountJson["sequence"].(string), 10, 64)
	if err != nil {
		panic(fmt.Sprintln("ParseUint sequence:", baseAccountJson, err))
	}

	return &authTypes.BaseAccount{
		Address: address,
		AccountNumber: accountNumber,
		Sequence: sequence,
	}
}

func NewCoins(coinsJson []interface{}) sdkTypes.Coins {
	// create coins from array of maps (for original vesting)
	coins := make([]sdkTypes.Coin, len(coinsJson))
	for i, coin := range coinsJson {
		// if coin.(map[string]interface{})["amount"] == nil {
		// 	panic(fmt.Sprintln("Error: amount is nil", coin))
		// }
		// if coin.(map[string]interface{})["denom"] == nil {
		// 	panic(fmt.Sprintln("Error: denom is nil", coin))
		// }

		amount, _ := strconv.ParseInt(coin.(map[string]interface{})["amount"].(string), 10, 64)
		coins[i] = sdkTypes.NewInt64Coin(coin.(map[string]interface{})["denom"].(string), amount)
	}

	return coins
}

func NewTime(strTime string) int64 {
	time, err := strconv.ParseInt(strTime, 10, 64) 
	if err != nil {
		fmt.Println("Error converting time")
		panic(fmt.Sprintln("Error converting time", time, err))
	}

	return time
}

func NewDelayedVestingAccount(baseAccount *authTypes.BaseAccount, baseVestingAccountJson map[string]interface{}) *vestingTypes.DelayedVestingAccount {
	originalVestingJson := baseVestingAccountJson["original_vesting"].([]interface{})
	coins := NewCoins(originalVestingJson)

	if baseVestingAccountJson["end_time"] == nil {
		panic(fmt.Sprintln("Error: end_time is nil", baseVestingAccountJson))
	}

	endTime := NewTime(baseVestingAccountJson["end_time"].(string))

	return vestingTypes.NewDelayedVestingAccount(baseAccount, coins, endTime)
}

func NewContinuousVestingAccount(baseAccount *authTypes.BaseAccount, startTime int64, baseVestingAccountJson map[string]interface{}) *vestingTypes.ContinuousVestingAccount {
	// if baseVestingAccountJson["original_vesting"] == nil || baseVestingAccountJson["original_vesting"].([]interface{}) == nil {
	// 	panic(fmt.Sprintln("Error: original_vesting is nil", baseVestingAccountJson))
	// }

	originalVestingJson := baseVestingAccountJson["original_vesting"].([]interface{})
	coins := NewCoins(originalVestingJson)

	// if baseVestingAccountJson["end_time"] == nil {
	// 	panic(fmt.Sprintln("Error: end_time is nil", baseVestingAccountJson))
	// }

	endTime := NewTime(baseVestingAccountJson["end_time"].(string))

	// if baseVestingAccountJson["end_time"] == nil {
	// 	panic(fmt.Sprintln("Error: end_time is nil", baseVestingAccountJson))
	// }

	return vestingTypes.NewContinuousVestingAccount(baseAccount, coins, startTime, endTime)
}

// there are no periodic vesting accounts in the genesis file so I am not implementing this
// func NewPeriodicVestingAccount(baseAccount *authTypes.BaseAccount, baseVestingAccountJson map[string]interface{}) vestingTypes.NewPeriodicVestingAccount {
// }

func main() {
	// get the data from the json file
	flag.Parse()
	args := flag.Args()
	if len(args) != 1 {
		fmt.Println("Please provide a genesis.json as the only argument")
		return
	}

	jsonFile := args[0]
	// read file with the io package
	file, err := os.Open(jsonFile)
	if err != nil {
		fmt.Println("Error opening json file")
		return
	}
	defer file.Close()

	// create a decoder
	decoder := json.NewDecoder(file)

	// create a map to store the data
	var data map[string]interface{}

	// decode the json file into the map
	err = decoder.Decode(&data)
	if err != nil {
		fmt.Println("Error decoding json")
		return
	}

	// get accounts
	appState := (data["app_state"].(map[string]interface{}))
	auth := (appState["auth"]).(map[string]interface{})
	accounts := (auth["accounts"]).([]interface{})

	// two arrays for delayed and continuous vesting accounts
	continuousAccounts := make([]*vestingTypes.ContinuousVestingAccount, 0)
	delayedAccounts := make([]*vestingTypes.DelayedVestingAccount, 0)

	cdc := codec.NewLegacyAmino()

	for idx, account := range accounts {
		// skip these because they aren't vesting accounts
		if account.(map[string]interface{}) == nil || account.(map[string]interface{})["@type"] == nil {
			fmt.Println("no type!", account)
			continue
		}
		
		if account.(map[string]interface{})["@type"] == "/cosmos.auth.v1beta1.BaseAccount" { 
			continue
		} 

		// get relevant data from account
		baseVestingAccountJson := account.(map[string]interface{})["base_vesting_account"]
		if baseVestingAccountJson == nil || baseVestingAccountJson.(map[string]interface{}) == nil { // apparently this can happen in the genesis file
			fmt.Println("Skipping account with no base_vesting_account: ", account)
			continue
		}

		baseAccountJson := baseVestingAccountJson.(map[string]interface{})["base_account"]
		if baseAccountJson == nil || baseAccountJson.(map[string]interface{}) == nil { // apparently this can happen in the genesis file
			fmt.Println("Skipping account with no base_account: ", account)
			continue
		}

		// create composed structs

		baseAccount := NewBaseAccount(baseAccountJson.(map[string]interface{}))

		paintedDelayed := false
		
		if account.(map[string]interface{})["@type"] == "/cosmos.vesting.v1beta1.ContinuousVestingAccount" {
			if account.(map[string]interface{})["start_time"] == nil {
				panic(fmt.Sprintln("Error: start_time is nil", account))
			}

			startTime := NewTime(account.(map[string]interface{})["start_time"].(string))

			continuousAccount := NewContinuousVestingAccount(baseAccount, startTime, baseVestingAccountJson.(map[string]interface{}))

			if continuousAccount != nil {
				continuousAccounts = append(continuousAccounts, continuousAccount)
			}
			fmt.Printf("continuousAccount %#+v %#+v %v\n\n", continuousAccount, continuousAccount.BaseVestingAccount, continuousAccount.BaseVestingAccount.OriginalVesting[0].Amount)
			jsonStr, _ := json.Marshal(continuousAccount) 
			rawMsgStr, _ := json.RawMessage.MarshalJSON((jsonStr))
			// fmt.Printf("\n\nJSON RAW STR %s\n\n", string(rawMsgStr))
			var foo vestingTypes.ContinuousVestingAccount
			cdc.UnmarshalJSON(rawMsgStr, &foo)
			fmt.Printf("\n\nFOO %#+v %#+v %v\n\n", foo, foo.BaseVestingAccount, foo.BaseVestingAccount.OriginalVesting[0].Amount)

			break
		} else {
			if paintedDelayed {
				continue
			}

			delayedAccount := NewDelayedVestingAccount(baseAccount, baseVestingAccountJson.(map[string]interface{}))

			if delayedAccount != nil {
				delayedAccounts = append(delayedAccounts, delayedAccount)
			}
			if idx == 0 {
				fmt.Printf("delayedAccount %#+v %#+v %v\n\n", delayedAccount, delayedAccount.BaseVestingAccount, delayedAccount.BaseVestingAccount.OriginalVesting[0].Amount)
				jsonStr, _ := json.Marshal(delayedAccount) 
				rawMsgStr, _ := json.RawMessage.MarshalJSON((jsonStr))
				// fmt.Printf("\n\nJSON RAW STR %s\n\n", string(rawMsgStr))
				var foo vestingTypes.DelayedVestingAccount
				cdc.UnmarshalJSON(account.([]byte), &foo)
				fmt.Printf("\n\nFOO %#+v %#+v %v\n\n", foo, foo.BaseVestingAccount, foo.BaseVestingAccount.OriginalVesting[0].Amount)
				// fmt.Printf("ACCOUNT %s\n\n", account)

				paintedDelayed = true
			}
		}

	}


	for _, account := range delayedAccounts {
		fmt.Printf("Delayed %#+v %+v\n", *account.BaseVestingAccount, *account.BaseAccount)
		break
	}

	for _, account := range continuousAccounts {
		fmt.Printf("Continuous %#+v %#+v %+v\n", *account, *account.BaseVestingAccount, *account.BaseAccount)
		break
	}

	fmt.Printf("\nDone\n")
}