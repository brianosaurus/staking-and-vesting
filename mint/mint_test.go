package mint

import (
	"encoding/json"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
)

const (
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

func TestGetParamsAndMinter(t *testing.T) {
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

	params, minter := GetParamsAndMinter(minterJson)
	
	paramsJson = paramsJson["params"].(map[string]interface{})
	inflationRateChange := sdk.MustNewDecFromStr(paramsJson["inflation_rate_change"].(string))
	inflationMax := sdk.MustNewDecFromStr(paramsJson["inflation_max"].(string))
	inflationMin := sdk.MustNewDecFromStr(paramsJson["inflation_min"].(string))
	goalBonded := sdk.MustNewDecFromStr(paramsJson["goal_bonded"].(string))

	assert.Equal(t, paramsJson["mint_denom"].(string), "uumee")
	assert.Equal(t, inflationRateChange, params.InflationRateChange)
	assert.Equal(t, inflationMax, params.InflationMax)
	assert.Equal(t, inflationMin, params.InflationMin)
	assert.Equal(t, goalBonded, params.GoalBonded)

	// 5 second blocks as opposed to what is in genesis.json
	assert.Equal(t, uint64((60 / SECONDS_PER_BLOCK) * 60 * 24 * 365), params.BlocksPerYear) 

	minterJson = minterJson["mint"].(map[string]interface{})["minter"].(map[string]interface{})
	inflation := sdk.MustNewDecFromStr(minterJson["inflation"].(string))
	annualProvisions := sdk.MustNewDecFromStr(minterJson["annual_provisions"].(string))

	assert.Equal(t, inflation, minter.Inflation)
	assert.Equal(t, annualProvisions, minter.AnnualProvisions)
}