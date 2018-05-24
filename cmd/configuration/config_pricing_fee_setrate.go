package configuration

import (
	"io/ioutil"
	"encoding/json"
)

type APIKey struct {
	ApiKey string `json:"api_key"`
}

func GetAPIKey(path string) string {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	result := APIKey{}
	err = json.Unmarshal(raw, &result)
	if err != nil {
		panic(err)
	}
	apiKey := result.ApiKey
	return apiKey
}

type BlockSetRate struct {
	BeginBlockSetRate uint64 `json:"begin_block_setrate"`
}

func GetBeginBlockSetRate(path string) uint64 {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	result := BlockSetRate{}
	err = json.Unmarshal(raw, &result)
	if err != nil {
		panic(err)
	}
	beginBlockSetRate := result.BeginBlockSetRate
	return beginBlockSetRate
}
