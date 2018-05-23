package world

import (
	"encoding/json"
	"io/ioutil"
)

type Endpoint interface {
	GoldDataEndpoint() string
	OneForgeGoldETHDataEndpoint() string
	OneForgeGoldUSDDataEndpoint() string
	GDAXDataEndpoint() string
	KrakenDataEndpoint() string
	GeminiDataEndpoint() string
}

type RealEndpoint struct {
	OneForgeKey string `json:"oneforge"`
}

func (self RealEndpoint) GoldDataEndpoint() string {
	return "https://datafeed.digix.global/tick/"
}

func (self RealEndpoint) OneForgeGoldETHDataEndpoint() string {
	return "https://forex.1forge.com/1.0.3/convert?from=XAU&to=ETH&quantity=1&api_key=" + self.OneForgeKey
}

func (self RealEndpoint) OneForgeGoldUSDDataEndpoint() string {
	return "https://forex.1forge.com/1.0.3/convert?from=XAU&to=USD&quantity=1&api_key=" + self.OneForgeKey
}

func (self RealEndpoint) GDAXDataEndpoint() string {
	return "https://api.gdax.com/products/eth-usd/ticker"
}

func (self RealEndpoint) KrakenDataEndpoint() string {
	return "https://api.kraken.com/0/public/Ticker?pair=ETHUSD"
}

func (self RealEndpoint) GeminiDataEndpoint() string {
	return "https://api.gemini.com/v1/pubticker/ethusd"
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
	return "http://simulator:5400/tick"
}

func (self SimulatedEndpoint) OneForgeGoldETHDataEndpoint() string {
	return "http://simulator:5500/1.0.3/convert?from=XAU&to=ETH&quantity=1&api_key="
}

func (self SimulatedEndpoint) OneForgeGoldUSDDataEndpoint() string {
	return "http://simulator:5500/1.0.3/convert?from=XAU&to=USD&quantity=1&api_key="
}

func (self SimulatedEndpoint) GDAXDataEndpoint() string {
	return "http://simulator:5600/products/eth-usd/ticker"
}

func (self SimulatedEndpoint) KrakenDataEndpoint() string {
	return "http://simulator:5700/0/public/Ticker?pair=ETHUSD"
}

func (self SimulatedEndpoint) GeminiDataEndpoint() string {
	return "http://simulator:5800/v1/pubticker/ethusd"
}
