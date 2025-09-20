package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const debug = false
const urlHttp = "http://"
const urlPortPath = ":80/api/v2/status"

func main() {
	urlString, err := determineUrl()
	if err != nil {
		return
	}

	fmt.Println("Reading from ", urlString)

	for {
		keyValues, err := readData(urlString)
		if err != nil {
			break
		}

		printValues(keyValues)

		time.Sleep(30 * time.Second)
	}
}

func determineUrl() (string, error) {
	netInterfaceAddresses, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, netInterfaceAddress := range netInterfaceAddresses {
		networkIp, ok := netInterfaceAddress.(*net.IPNet)
		if !ok {
			continue
		}
		ip := networkIp.IP
		if ip.IsLoopback() {
			continue
		}

		if ip.To4() == nil {
			continue
		}

		defaultMaskString := ip.DefaultMask().String()
		defaultMask, err := strconv.ParseUint(defaultMaskString, 16, 32)
		if err != nil {
			continue
		}
		if defaultMask != 0xff_ff_ff_00 {
			continue
		}

		ipString := ip.String()
		ip4Parts := strings.Split(ipString, ".")
		if len(ip4Parts) != 4 {
			continue
		}

		if debug {
			fmt.Print("Testing")
		}
		client := http.Client{Timeout: 50 * time.Millisecond}
		for lastIpSegment := 100; lastIpSegment < 254; lastIpSegment++ {
			if debug {
				if lastIpSegment%10 == 0 {
					fmt.Printf(" %d", lastIpSegment)
				} else {
					fmt.Print(".")
				}
			}
			myUrl := fmt.Sprintf("%s%s.%s.%s.%d%s", urlHttp, ip4Parts[0], ip4Parts[1], ip4Parts[2], lastIpSegment, urlPortPath)
			response, err := client.Get(myUrl)
			if err != nil {
				continue
			}
			response.Body.Close()
			if debug {
				fmt.Println("Found")
			}
			return myUrl, nil
		}
		fmt.Println()
	}
	return "", errors.New("nothing found")
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
