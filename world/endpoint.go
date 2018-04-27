package world

import (
	"encoding/json"
	"io/ioutil"
)

type Endpoint interface {
	GoldDataEndpoint() string
	BackupGoldDataEndpoint() string
}

type RealEndpoint struct {
	OneForgeKey string `json:"oneforge"`
}

func (self RealEndpoint) GoldDataEndpoint() string {
	return "https://datafeed.digix.global/tick/"
}

func (self RealEndpoint) BackupGoldDataEndpoint() string {
	return "https://forex.1forge.com/1.0.3/convert?from=XAU&to=ETH&quantity=1&api_key=" + self.OneForgeKey
}

func NewRealEndpointFromFile(path string) (*RealEndpoint, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	result := RealEndpoint{}
	err = json.Unmarshal(data, &result)
	return &result, err
}

type SimulatedEndpoint struct {
}

func (self SimulatedEndpoint) GoldDataEndpoint() string {
	return ""
}

func (self SimulatedEndpoint) BackupGoldDataEndpoint() string {
	return ""
}
