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
