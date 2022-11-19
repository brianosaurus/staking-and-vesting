package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	mintingTypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	vestingModule "github.com/brianosaurus/challenge2/vesting"
	stakingModule "github.com/brianosaurus/challenge2/staking"
	mintModule "github.com/brianosaurus/challenge2/mint"
)

const (
	SECONDS_PER_BLOCK = 5
)


func WriteCSV(writer *csv.Writer, vestingOnDays *map[int]sdk.Dec, totalSupply sdk.Dec, stakedTokens sdk.Dec, 
	minter mintingTypes.Minter, params mintingTypes.Params) {
	// write the header
	err := writer.Write([]string{"Days Since Genesis Analyzed", "Tokens Unvesting", "Inflation", "Staking Rewards", "Circulating Supply", "Total Supply"})
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

	continuousVestingAccounts, delayedVestingAccounts := vestingModule.GetVestingAccounts(appState)

	// totalSupply here matches the total supply in the genesis.json from the banking module. A good verification that math is correct
	totalSupply, vestingOnDays := vestingModule.GetTotalSupplyAndVestingSchedule(appState, continuousVestingAccounts, delayedVestingAccounts)
	stakedTokens := stakingModule.GetStakedTokens(appState)
	params, minter := mintModule.GetParamsAndMinter(appState)

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
