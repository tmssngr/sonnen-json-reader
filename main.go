package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const url = "http://192.168.92.216:80/api/v2/status"

func main() {
	for {
		keyValues, err := readData(url)
		if err != nil {
			break
		}

		printValues(keyValues)

		time.Sleep(30 * time.Second)
	}
}

func printValues(keyValues map[string]interface{}) {
	consumptionAvg := keyValues["Consumption_Avg"].(float64)
	productionW := keyValues["Production_W"].(float64)
	gridFeedInW := keyValues["GridFeedIn_W"].(float64)
	pacTotalW := keyValues["Pac_total_W"].(float64)
	usoc := keyValues["USOC"].(float64)
	fmt.Printf("\n")
	fmt.Printf("Consumption: %4.0f W\n", consumptionAvg)
	fmt.Printf("Production:  %4.0f W\n", productionW)
	if gridFeedInW < 0.0 {
		fmt.Printf("Buying:      %4.0f W\n", -gridFeedInW)
	} else {
		fmt.Printf("Selling:     %4.0f W\n", gridFeedInW)
	}
	fmt.Printf("Battery:     %4.0f %% ", usoc)
	fmt.Printf("%4.0f W ", -pacTotalW)
	if pacTotalW < 0.0 {
		fmt.Printf("(charging)\n")
	} else {
		fmt.Printf("(discharging)\n")
	}
}

func readData(url string) (map[string]interface{}, error) {
	response, err := http.Get(url)
	if err != nil {
		fmt.Printf("Failed to read from %s: %v\n", url, err)
		return nil, err
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("Failed to read from %s: %v\n", url, err)
		return nil, err
	}

	var keyValues map[string]interface{}
	err = json.Unmarshal(body, &keyValues)
	if err != nil {
		fmt.Printf("Failed to parse %s: %v\n", body, err)
		return nil, err
	}

	return keyValues, nil
}
