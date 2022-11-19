package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"math/big"
	"os"
	"sort"
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	vestingTypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	mintingTypes "github.com/cosmos/cosmos-sdk/x/mint/types"

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

func GetStakedTokens(appState map[string]interface{}) sdk.Dec {
	genutil := (appState["genutil"]).(map[string]interface{})
	genTxs := (genutil["gen_txs"]).([]interface{})
	stakedTokens := sdk.NewDec(0)

	for _, genTx := range genTxs {
		body := (genTx.(map[string]interface{})["body"]).(map[string]interface{})
		messages := (body["messages"]).([]interface{})

		for _, message := range messages {
			if message.(map[string]interface{})["@type"] == "/cosmos.staking.v1beta1.MsgCreateValidator" {
				amount := message.(map[string]interface{})["value"].(map[string]interface{})["amount"].(string)
				amountI, err := strconv.ParseInt(amount, 10, 64)
				if err != nil {
					panic(fmt.Sprintln("Error parsing amount", err))
				}

				stakedTokens = stakedTokens.Add(sdk.NewDec(amountI))
			}
		}
	}

	return stakedTokens
}

func GetParamsAndMinter(appState map[string]interface{}) (mintingTypes.Params, mintingTypes.Minter) {
	mint := (appState["mint"]).(map[string]interface{})
	minterJson := (mint["minter"]).(map[string]interface{})
	paramsJson := (mint["params"]).(map[string]interface{})

	params := mintingTypes.Params{
		MintDenom: paramsJson["mint_denom"].(string),
		InflationRateChange: sdk.MustNewDecFromStr(paramsJson["inflation_rate_change"].(string)),
		InflationMax:        sdk.MustNewDecFromStr(paramsJson["inflation_max"].(string)),
		InflationMin:        sdk.MustNewDecFromStr(paramsJson["inflation_min"].(string)),
		GoalBonded:          sdk.MustNewDecFromStr(paramsJson["goal_bonded"].(string)),
		BlocksPerYear:       uint64((60 / SECONDS_PER_BLOCK) * 60 * 24 *365), // change this as we're assuming 5 second blocks
	}

	minter := mintingTypes.Minter{
		Inflation:          sdk.MustNewDecFromStr(minterJson["inflation"].(string)),
		AnnualProvisions:   sdk.MustNewDecFromStr(minterJson["annual_provisions"].(string)),
	}

	return params, minter
}


func WriteCSV(writer *csv.Writer, vestingOnDays *map[int]sdk.Dec, totalSupply sdk.Dec, stakedTokens sdk.Dec, 
	minter mintingTypes.Minter, params mintingTypes.Params) {
	// write the header
	err := writer.Write([]string{"Days Since Genesis Analyzed", "Tokens Unvesting", "Inflation", "Staking rewards", "Circulating Supply", "Total Supply"})
	if err != nil {
		fmt.Println("Error writing to csv")
		return
	}

	days := make([]int, 0, len(*vestingOnDays))

	for day := range *vestingOnDays {
		days = append(days, day)
	}

	sort.Ints(days)

	// is total in circulation really correct here? Vesting tokens are unaccessible however they came from an
	// account initially. So they have been minted already. In any case, I'll go with the assumption that
	// tokens that haven't yet been vested are not in circulation however staked tokens are in circulation
	// because staked tokens can be retrived even if there is a lockout period. Excluding staked yet to vest tokens.
	totalInCirculation := totalSupply

	for day := range days {
		if _, ok := (*vestingOnDays)[day]; !ok {
			continue
		}

		totalInCirculation = totalInCirculation.Sub((*vestingOnDays)[day])
	}

	stakingRewards := sdk.NewDec(0)

	csvStr := []string{"0", sdk.NewDec(0).RoundInt().String(), minter.Inflation.String(), stakingRewards.RoundInt().String(), 
		totalInCirculation.RoundInt().String(), totalSupply.RoundInt().String()}
	writer.Write(csvStr)

	for day := range days {
		if _, ok := (*vestingOnDays)[day]; !ok {
			continue
		}

		// calculate inflation for the previous day (rewards are calculated and rewarded every block)
		stakingRatio := stakedTokens.Quo(totalSupply)


		// calculate rewards for each hour of the previous day
		for i := 1; i < 24; i++ {
			// inflation changes hourly so make these calculations hourly
			minter.Inflation = minter.NextInflationRate(params, stakingRatio)
			minter.AnnualProvisions = minter.NextAnnualProvisions(params, sdk.NewInt(totalSupply.RoundInt64()))
			coins := minter.BlockProvision(params)

			// rewards are distributed every block
			for j := 0; j < (60 / SECONDS_PER_BLOCK); j++ {
				stakingRewards = stakingRewards.Add(sdk.NewDecFromBigInt(coins.Amount.BigInt()))
				totalInCirculation = totalInCirculation.Add(sdk.NewDecFromBigInt(coins.Amount.BigInt()))
				totalSupply = totalSupply.Add(sdk.NewDecFromBigInt(coins.Amount.BigInt()))
			}
		}

		totalInCirculation = totalInCirculation.Add((*vestingOnDays)[day]) // add recently unvested tokens to total in circulation
		csvStr = []string{strconv.Itoa(day), (*vestingOnDays)[day].RoundInt().String(), minter.Inflation.String(), stakingRewards.RoundInt().String(), 
			totalInCirculation.RoundInt().String(), totalSupply.RoundInt().String()}

		writer.Write(csvStr)
	}

	writer.Flush()
}

func main() {
	// get the data from the json file
	// flag for the output csv file
	var csvStr string
	var genesisFile string
	flag.StringVar(&csvStr, "csv", "genesis_analysis.csv", "the csv file to output the data to")
	flag.StringVar(&genesisFile, "genesis", "genesis.json", "the genesis file to analyze")
	flag.Parse()

	// read file with the io package
	file, err := os.Open(genesisFile)
	if err != nil {
		fmt.Println("Error opening json file")
		return
	}
	defer file.Close()

	// create a decoder
	decoder := json.NewDecoder(file)

	// create a map to store the data
	var genesis map[string]interface{}

	// decode the json file into the map
	err = decoder.Decode(&genesis)
	if err != nil {
		fmt.Println("Error decoding json")
		return
	}

	// get accounts
	appState := (genesis["app_state"].(map[string]interface{}))

	continuousVestingAccounts, delayedVestingAccounts := GetVestingAccounts(appState)

	// totalSupply here matches the total supply in the genesis.json from the banking module. A good verification that math is correct
	totalSupply, vestingOnDays := GetTotalSupplyAndVestingSchedule(appState, continuousVestingAccounts, delayedVestingAccounts)
	stakedTokens := GetStakedTokens(appState)
	params, minter := GetParamsAndMinter(appState)

	// write the data to a csv file
	file, err = os.Create(csvStr)
	if err != nil {
		fmt.Println("Error creating csv file")
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)

	WriteCSV(writer, vestingOnDays, totalSupply, stakedTokens, minter, params)
	fmt.Printf("\nDone\n")
}
