package world

import (
	"encoding/json"
	"io/ioutil"

	"github.com/KyberNetwork/reserve-data/common/config"
)

var (
	allFeeds = []string{
		"DGX",
		"OneForgeETH",
		"OneForgeUSD",
		"GDAX",
		"Kraken",
		"Gemini",

		"CoinbaseBTC",
		"GeminiBTC",

		"CoinbaseUSD",
		"BinanceUSD",

		"CoinbaseUSDC",
		"BinanceUSDC",

		"CoinbaseDAI",
		"HitDAI",

		"BitFinexUSD",
		"BinanceUSDT",
		"BinancePAX",
		"BinanceTUSD",
	}
	// remove unused feeds
)

// AllFeeds returns all configured feed sources.
func AllFeeds() []string {
	return allFeeds
}

// Endpoint returns all API endpoints to use in TheWorld struct.
type Endpoint interface {
	GoldDataEndpoint() string
	OneForgeGoldETHDataEndpoint() string
	OneForgeGoldUSDDataEndpoint() string
	GDAXDataEndpoint() string
	KrakenDataEndpoint() string
	GeminiDataEndpoint() string

	CoinbaseBTCEndpoint() string
	GeminiBTCEndpoint() string

	CoinbaseUSDCEndpoint() string
	BinanceUSDCEndpoint() string
	CoinbaseUSDEndpoint() string
	CoinbaseDAIEndpoint() string
	HitDaiEndpoint() string

	BitFinexUSDTEndpoint() string
	BinanceUSDTEndpoint() string
	BinancePAXEndpoint() string
	BinanceTUSDEndpoint() string
}

type RealEndpoint struct {
	OneForgeKey string `json:"oneforge"`
}

func (re RealEndpoint) BitFinexUSDTEndpoint() string {
	return "https://api-pub.bitfinex.com/v2/ticker/tETHUSD"
}

func (re RealEndpoint) BinanceUSDTEndpoint() string {
	return "https://api.binance.com/api/v3/ticker/bookTicker?symbol=ETHUSDT"
}

func (re RealEndpoint) BinancePAXEndpoint() string {
	return "https://api.binance.com/api/v3/ticker/bookTicker?symbol=ETHPAX"
}

func (re RealEndpoint) BinanceTUSDEndpoint() string {
	return "https://api.binance.com/api/v3/ticker/bookTicker?symbol=ETHTUSD"
}

func (re RealEndpoint) CoinbaseDAIEndpoint() string {
	return "https://api.pro.coinbase.com/products/eth-dai/ticker"
}

func (re RealEndpoint) HitDaiEndpoint() string {
	return "https://api.hitbtc.com/api/2/public/ticker/ETHDAI"
}

func (re RealEndpoint) CoinbaseUSDEndpoint() string {
	return "https://api.pro.coinbase.com/products/eth-usd/ticker"
}

func (re RealEndpoint) CoinbaseUSDCEndpoint() string {
	return "https://api.pro.coinbase.com/products/eth-usdc/ticker"
}

func (re RealEndpoint) BinanceUSDCEndpoint() string {
	return "https://api.binance.com/api/v3/ticker/bookTicker?symbol=ETHUSDC"
}

func (re RealEndpoint) GoldDataEndpoint() string {
	return "https://datafeed.digix.global/tick/"
}

func (re RealEndpoint) OneForgeGoldETHDataEndpoint() string {
	return "https://api.1forge.com/convert?from=XAU&to=ETH&quantity=1&api_key=" + re.OneForgeKey
}

func (re RealEndpoint) OneForgeGoldUSDDataEndpoint() string {
	return "https://api.1forge.com/convert?from=XAU&to=USD&quantity=1&api_key=" + re.OneForgeKey
}

func (re RealEndpoint) GDAXDataEndpoint() string {
	return "https://api.pro.coinbase.com/products/eth-usd/ticker"
}

func (re RealEndpoint) KrakenDataEndpoint() string {
	return "https://api.kraken.com/0/public/Ticker?pair=ETHUSD"
}

func (re RealEndpoint) GeminiDataEndpoint() string {
	return "https://api.gemini.com/v1/pubticker/ethusd"
}

func (re RealEndpoint) CoinbaseBTCEndpoint() string {
	return "https://api.pro.coinbase.com/products/eth-btc/ticker"
}

func (re RealEndpoint) GeminiBTCEndpoint() string {
	return "https://api.gemini.com/v1/pubticker/ethbtc"
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

// SimulatedEndpoint implement endpoint for testing in simulate.
type SimulatedEndpoint struct {
	eps config.WorldEndpoints
}

// NewSimulationEndpoint ...
func NewSimulationEndpoint(eps config.WorldEndpoints) *SimulatedEndpoint {
	return &SimulatedEndpoint{eps: eps}
}

// NewSimulationEndpointFromFile create new simulation with config endpoints from file
func NewSimulationEndpointFromFile(file string) (*SimulatedEndpoint, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	result := config.WorldEndpoints{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}
	return NewSimulationEndpoint(result), nil
}

func (se SimulatedEndpoint) BitFinexUSDTEndpoint() string {
	return se.eps.BitFinexUSDT.URL
}

func (se SimulatedEndpoint) BinanceUSDTEndpoint() string {
	return se.eps.BinanceUSDT.URL
}

func (se SimulatedEndpoint) BinancePAXEndpoint() string {
	return se.eps.BinancePAX.URL
}

func (se SimulatedEndpoint) BinanceTUSDEndpoint() string {
	return se.eps.BinanceTUSD.URL
}

func (se SimulatedEndpoint) CoinbaseDAIEndpoint() string {
	return se.eps.CoinbaseDAI.URL
}

func (se SimulatedEndpoint) HitDaiEndpoint() string {
	return se.eps.HitDai.URL
}

func (se SimulatedEndpoint) CoinbaseUSDEndpoint() string {
	return se.eps.CoinbaseUSD.URL
}

// TODO: support simulation
func (se SimulatedEndpoint) CoinbaseUSDCEndpoint() string {
	return se.eps.CoinbaseUSDC.URL
}

func (se SimulatedEndpoint) BinanceUSDCEndpoint() string {
	return se.eps.BinanceUSDC.URL
}

func (se SimulatedEndpoint) GoldDataEndpoint() string {
	return se.eps.GoldData.URL
}

func (se SimulatedEndpoint) OneForgeGoldETHDataEndpoint() string {
	return se.eps.OneForgeGoldETH.URL
}

func (se SimulatedEndpoint) OneForgeGoldUSDDataEndpoint() string {
	return se.eps.OneForgeGoldUSD.URL
}

func (se SimulatedEndpoint) GDAXDataEndpoint() string {
	return se.eps.GDAXData.URL
}

func (se SimulatedEndpoint) KrakenDataEndpoint() string {
	return se.eps.KrakenData.URL
}

func (se SimulatedEndpoint) GeminiDataEndpoint() string {
	return se.eps.GeminiData.URL
}

func (se SimulatedEndpoint) CoinbaseBTCEndpoint() string {
	return se.eps.CoinbaseBTC.URL
}

func (se SimulatedEndpoint) GeminiBTCEndpoint() string {
	return se.eps.GeminiBTC.URL
}
