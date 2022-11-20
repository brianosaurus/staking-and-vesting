
package mint

import (

	sdk "github.com/cosmos/cosmos-sdk/types"
	mintingTypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

const (
	SECONDS_PER_BLOCK = 5
)

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
		BlocksPerYear:       uint64((60 / SECONDS_PER_BLOCK) * 60 * 24 * 365), // change this as we're assuming 5 second blocks
	}

	minter := mintingTypes.Minter{
		Inflation:          sdk.MustNewDecFromStr(minterJson["inflation"].(string)),
		AnnualProvisions:   sdk.MustNewDecFromStr(minterJson["annual_provisions"].(string)),
	}

	return params, minter
}
