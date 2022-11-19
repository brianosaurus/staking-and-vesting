package staking

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

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
