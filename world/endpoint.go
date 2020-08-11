package world

import (
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

		"CoinbaseETHBTC3",
		"BinanceETHBTC3",

		"CoinbaseETHUSD10000",
		"CoinbaseETHUSDC10000",
		"BinanceETHUSDC10000",
		"CoinbaseETHUSDDAI5000",
		"BitfinexETHUSDT10000",
		"BinanceETHUSDT10000",
		"BinanceETHPAX5000",
		"BinanceETHBUSD10000",
		"GeminiETHUSD10000",
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

	BinanceETHBTC3Endpoint() string
	CoinbaseETHBTC3Endpoint() string

	CoinbaseETHUSDC10000Endpoint() string
	BinanceETHUSDC10000Endpoint() string
	CoinbaseETHUSD10000Endpoint() string
	CoinbaseETHUSDDAI5000Endpoint() string
	BitfinexETHUSDT10000Endpoint() string
	BinanceETHUSDT10000Endpoint() string
	BinanceETHPAX5000Endpoint() string
	BinanceETHBUSD10000Endpoint() string
	GeminiETHUSD10000Endpoint() string
}

// Endpoints implement endpoint for testing in simulate.
type Endpoints struct {
	eps config.WorldEndpoints
}

// NewWorldEndpoint ...
func NewWorldEndpoint(eps config.WorldEndpoints) *Endpoints {
	return &Endpoints{eps: eps}
}

func (ep Endpoints) BitfinexETHUSDT10000Endpoint() string {
	return ep.eps.BitfinexETHUSDT10000.URL
}

func (ep Endpoints) GeminiETHUSD10000Endpoint() string {
	return ep.eps.GeminiETHUSD10000.URL
}

func (ep Endpoints) BinanceETHUSDT10000Endpoint() string {
	return ep.eps.BinanceETHUSDT10000.URL
}

func (ep Endpoints) BinanceETHPAX5000Endpoint() string {
	return ep.eps.BinanceETHPAX5000.URL
}

func (ep Endpoints) BinanceETHBUSD10000Endpoint() string {
	return ep.eps.BinanceETHBUSD10000.URL
}

func (ep Endpoints) CoinbaseETHUSDDAI5000Endpoint() string {
	return ep.eps.CoinbaseETHUSDDAI5000.URL
}

func (ep Endpoints) CoinbaseETHUSD10000Endpoint() string {
	return ep.eps.CoinbaseETHUSD10000.URL
}

// TODO: support simulation
func (ep Endpoints) CoinbaseETHUSDC10000Endpoint() string {
	return ep.eps.CoinbaseETHUSDC10000.URL
}

func (ep Endpoints) BinanceETHUSDC10000Endpoint() string {
	return ep.eps.BinanceETHUSDC10000.URL
}

func (ep Endpoints) GoldDataEndpoint() string {
	return ep.eps.GoldData.URL
}

func (ep Endpoints) OneForgeGoldETHDataEndpoint() string {
	return ep.eps.OneForgeGoldETH.URL
}

func (ep Endpoints) OneForgeGoldUSDDataEndpoint() string {
	return ep.eps.OneForgeGoldUSD.URL
}

func (ep Endpoints) GDAXDataEndpoint() string {
	return ep.eps.GDAXData.URL
}

func (ep Endpoints) KrakenDataEndpoint() string {
	return ep.eps.KrakenData.URL
}

func (ep Endpoints) GeminiDataEndpoint() string {
	return ep.eps.GeminiData.URL
}

func (ep Endpoints) CoinbaseETHBTC3Endpoint() string {
	return ep.eps.CoinbaseETHBTC3.URL
}

func (ep Endpoints) BinanceETHBTC3Endpoint() string {
	return ep.eps.BinanceETHBTC3.URL
}
