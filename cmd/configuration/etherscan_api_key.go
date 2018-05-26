package configuration

import (
	"encoding/json"
	"io/ioutil"
)

type EtherscanAPIKey struct {
	ApiKey string `json:"etherscan_api_key"`
}

func GetEtherscanAPIKey(path string) string {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	result := EtherscanAPIKey{}
	err = json.Unmarshal(raw, &result)
	if err != nil {
		panic(err)
	}
	apiKey := result.ApiKey
	return apiKey
}
