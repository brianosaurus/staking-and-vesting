package main

import (
	"flag"
	"fmt"
	"os"
	"encoding/json"
)

func main() {
	// get the data from the json file
	flag.Parse()
	args := flag.Args()
	if len(args) != 1 {
		fmt.Println("Please provide a json file")
		return
	}

	jsonFile := args[0]
	fmt.Println(jsonFile)
	// read file with the io package
	file, err := os.Open(jsonFile)
	if err != nil {
		fmt.Println("Error opening file")
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
	app_state := (data["app_state"].(map[string]interface{}))
	auth := (app_state["auth"]).(map[string]interface{})
	accounts := (auth["accounts"]).([]interface{})

	fmt.Println(accounts[0])
}