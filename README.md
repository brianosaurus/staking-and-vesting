# challenge2

## Running the challenge

To run the code 
```sh
make; ./genesisAnalyzer
```

genesisAnalyzer takes several arguments who have defaults. They are listed here
```sh
./genesisAnalyzer -h
Usage of ./genesisAnalyzer:
  -csv string
    	the csv file to output the data to (default "genesis_analysis.csv")
  -genesis string
    	the genesis file to analyze (default "genesis.json")
```

getData will overwrite the output files on subsequent runs (for convenience).

To Test 
```sh
go test ./...
```

## Output file formats

### validators.csv

The columns are labeled on the first line of the csv: 

```Days Since Genesis Analyzed, Tokens Unvesting, Inflation, Staking Rewards, Circulating Supply, and Total Supply.```

Days Since Genesis Analyzed starts at Day zero and increases the day from there. Each day Inflation changes, rewards are given,
tokens unvest, and the Circulating plus the Total supplies increase. It is worth noting that rewards are actually given each block but inflation
only changes hourly. This is taken into account within the algorithm to paint the CSV. 

Furthermore, each day (by the day) new tokens are unvested (granted to the owner to transfer) and this is a daily calculation.

```csv
Days Since Genesis Analyzed,Tokens Unvesting,Inflation,Staking Rewards,Circulating Supply,Total Supply
0,0,0.130000000000000000,0,3178325438516160,10000000000000000
0,7540923156480,0.130003640132970884,56888352540,3185923250025180,10000056888352540
1,7540923156480,0.130007280265978706,113778621636,3193521063450756,10000113778621636
2,7540923156480,0.130010920399023466,170670807300,3201118878792900,10000170670807300
3,7540923156480,0.130014560532105164,227564909568,3208716696051648,10000227564909568
4,7540923156480,0.130018200665223800,284460928488,3216314515227048,10000284460928488
```
