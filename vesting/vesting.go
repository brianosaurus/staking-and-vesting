package vesting

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	vestingTypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	"github.com/cosmos/cosmos-sdk/codec"
)

const (
	SECONDS_PER_BLOCK = 5
)

func NewDelayedVestingAccount(account map[string]interface{}, codec *codec.LegacyAmino) *vestingTypes.DelayedVestingAccount {
	rawAccount, err := json.Marshal(account)
	if err != nil {
		fmt.Println("Error marshalling DelayedVestingAccount")
		return nil
	}

	var delayedVestingAccount vestingTypes.DelayedVestingAccount
	err = codec.UnmarshalJSON(rawAccount, &delayedVestingAccount)
	if err != nil {
		panic(fmt.Sprintln("Error unmarshalling DelayedVestingAccount", err))
	}

	return &delayedVestingAccount
}

func NewContinuousVestingAccount(account map[string]interface{}, codec *codec.LegacyAmino) *vestingTypes.ContinuousVestingAccount {
	rawAccount, err := json.Marshal(account)
	if err != nil {
		fmt.Println("Error marshalling DelayedVestingAccount")
		return nil
	}

	var continuousVestingAccount vestingTypes.ContinuousVestingAccount
	err = codec.UnmarshalJSON(rawAccount, &continuousVestingAccount)
	if err != nil {
		panic(fmt.Sprintln("Error unmarshalling ContinuousVestingAccount", err))
	}

	return &continuousVestingAccount
}

// there are no periodic vesting accounts in the genesis file so I am not implementing this
// func NewPeriodicVestingAccount(baseAccount *authTypes.BaseAccount, baseVestingAccountJson map[string]interface{}) vestingTypes.NewPeriodicVestingAccount {
// }
func GetVestingAccounts(appState map[string]interface{}) (map[string]*vestingTypes.ContinuousVestingAccount,
	map[string]*vestingTypes.DelayedVestingAccount,
) {
	auth := (appState["auth"]).(map[string]interface{})
	accounts := (auth["accounts"]).([]interface{})

	// two arrays for delayed and continuous vesting accounts
	continuousAccounts := make(map[string]*vestingTypes.ContinuousVestingAccount)
	delayedAccounts := make(map[string]*vestingTypes.DelayedVestingAccount)

	// genesis.json is in the amino format. We need to use the amino codec to unmarshal the accounts
	cdc := codec.NewLegacyAmino()

	for _, account := range accounts {
		if account.(map[string]interface{})["@type"] == "/cosmos.auth.v1beta1.BaseAccount" {
			continue
		}

		if account.(map[string]interface{})["@type"] == "/cosmos.vesting.v1beta1.ContinuousVestingAccount" {
			continuousAccount := NewContinuousVestingAccount(account.(map[string]interface{}), cdc)

			if continuousAccount != nil {
				continuousAccounts[continuousAccount.Address] = continuousAccount
			}
		} else {
			delayedAccount := NewDelayedVestingAccount(account.(map[string]interface{}), cdc)

			if delayedAccount != nil {
				delayedAccounts[delayedAccount.Address] = delayedAccount
			}
		}
	}

	return continuousAccounts, delayedAccounts
}

func GetTotalSupplyAndVestingSchedule(appState map[string]interface{}, continuousAccounts map[string]*vestingTypes.ContinuousVestingAccount,
	delayedAccounts map[string]*vestingTypes.DelayedVestingAccount,
) (sdk.Dec, *map[int]sdk.Dec) {
	bank := (appState["bank"]).(map[string]interface{})
	balances := (bank["balances"]).([]interface{})

	totalSupply := sdk.NewDec(0)

	for _, account := range balances {
		baseAccount := account.(map[string]interface{})
		address := baseAccount["address"].(string)

		// if it is a vesting account skip (for now)
		if _, ok := continuousAccounts[address]; ok {
			continue
		}

		if _, ok := delayedAccounts[address]; ok {
			continue
		}

		amountStr := baseAccount["coins"].([]interface{})[0].(map[string]interface{})["amount"].(string)
		amountI, err := strconv.ParseInt(amountStr, 10, 64)
		if err != nil {
			panic(fmt.Sprintln("Error parsing amount", err))
		}

		totalSupply = totalSupply.Add(sdk.NewDec(amountI))
	}

	theTime := time.Now().Unix()

	vestingOnDays := make(map[int]sdk.Dec)

	for _, account := range continuousAccounts {
		amount := account.OriginalVesting[0].Amount // there are only 1 coin in the array for all accounts in genesis.json
		startTime := account.StartTime
		endTime := account.EndTime

		totalSupply = totalSupply.Add(sdk.NewDecFromBigInt(amount.BigInt()))

		if startTime > theTime {
			continue
		}

		// math to get the number of tokens that have vested in a continuous vesting account
		secondsTokenHasBeenVesting := big.NewInt(0).Sub(big.NewInt(endTime), big.NewInt(startTime))
		numberOfFiveSecondChunksTokenHasBeenVesting := big.NewInt(0).Div(secondsTokenHasBeenVesting, big.NewInt(SECONDS_PER_BLOCK))
		numberOfTokensVestingInTotalDuringTimeQuanta := big.NewInt(amount.Int64())
		tokensVestedPerBlock := amount.BigInt().Div(numberOfTokensVestingInTotalDuringTimeQuanta, numberOfFiveSecondChunksTokenHasBeenVesting)
		tokensVestedPerDay := tokensVestedPerBlock.Mul(tokensVestedPerBlock, big.NewInt((60 / SECONDS_PER_BLOCK) * 60 * 24)) 

		// add the tokens that have not vested to the map by days since today
		daysLeft := int((endTime - theTime) / 86400)

		for vestingDay := 0; vestingDay <= daysLeft; vestingDay++ {
			if _, ok := vestingOnDays[vestingDay]; !ok {
				vestingOnDays[vestingDay] = sdk.NewDec(0)
			}

			vestingOnDays[vestingDay] = vestingOnDays[vestingDay].Add(sdk.NewDecFromBigInt(tokensVestedPerDay))
		}
	}

	for _, account := range delayedAccounts {
		amount := account.OriginalVesting[0].Amount // there are only 1 coin in the array for all accounts in genesis.json
		endTime := account.EndTime

		totalSupply = totalSupply.Add(sdk.NewDecFromBigInt(amount.BigInt()))

		if endTime < theTime {
			continue
		}

		// add the tokens that have not vested to the map by the Nth day since today
		vestingDay := int((endTime - theTime) / 86400)

		if _, ok := vestingOnDays[vestingDay]; !ok {
			vestingOnDays[vestingDay] = sdk.NewDec(0)
		}

		vestingOnDays[vestingDay] = vestingOnDays[vestingDay].Add(sdk.NewDecFromBigInt(amount.BigInt()))
	}

	return totalSupply, &vestingOnDays 
}