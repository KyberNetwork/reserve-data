package world

import (
	"encoding/json"
	"io/ioutil"
)

// Endpoint returns all API endpoints to use in TheWorld struct.
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

func (re RealEndpoint) GoldDataEndpoint() string {
	return "https://datafeed.digix.global/tick/"
}

func (re RealEndpoint) OneForgeGoldETHDataEndpoint() string {
	return "https://forex.1forge.com/1.0.3/convert?from=XAU&to=ETH&quantity=1&api_key=" + re.OneForgeKey
}

func (re RealEndpoint) OneForgeGoldUSDDataEndpoint() string {
	return "https://forex.1forge.com/1.0.3/convert?from=XAU&to=USD&quantity=1&api_key=" + re.OneForgeKey
}

func (re RealEndpoint) GDAXDataEndpoint() string {
	return "https://api.gdax.com/products/eth-usd/ticker"
}

func (re RealEndpoint) KrakenDataEndpoint() string {
	return "https://api.kraken.com/0/public/Ticker?pair=ETHUSD"
}

func (re RealEndpoint) GeminiDataEndpoint() string {
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

func (se SimulatedEndpoint) GoldDataEndpoint() string {
	return "http://simulator:5400/tick"
}

func (se SimulatedEndpoint) OneForgeGoldETHDataEndpoint() string {
	return "http://simulator:5500/1.0.3/convert?from=XAU&to=ETH&quantity=1&api_key="
}

func (se SimulatedEndpoint) OneForgeGoldUSDDataEndpoint() string {
	return "http://simulator:5500/1.0.3/convert?from=XAU&to=USD&quantity=1&api_key="
}

func (se SimulatedEndpoint) GDAXDataEndpoint() string {
	return "http://simulator:5600/products/eth-usd/ticker"
}

func (se SimulatedEndpoint) KrakenDataEndpoint() string {
	return "http://simulator:5700/0/public/Ticker?pair=ETHUSD"
}

func (se SimulatedEndpoint) GeminiDataEndpoint() string {
	return "http://simulator:5800/v1/pubticker/ethusd"
}
