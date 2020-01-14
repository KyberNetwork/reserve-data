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

		"CoinbaseBTC",
		"GeminiBTC", // this no longer use and replace by BinanceBTC below.
		"BinanceBTC",

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

	CoinbaseUSDCEndpoint() string
	BinanceUSDCEndpoint() string
	CoinbaseUSDEndpoint() string
	CoinbaseDAIEndpoint() string
	HitDaiEndpoint() string

	BitFinexUSDTEndpoint() string
	BinanceUSDTEndpoint() string
	BinancePAXEndpoint() string
	BinanceTUSDEndpoint() string
	BinanceBTCEndpoint() string
}

// Endpoints implement endpoint for testing in simulate.
type Endpoints struct {
	eps config.WorldEndpoints
}

// NewWorldEndpoint ...
func NewWorldEndpoint(eps config.WorldEndpoints) *Endpoints {
	return &Endpoints{eps: eps}
}

func (ep Endpoints) BitFinexUSDTEndpoint() string {
	return ep.eps.BitFinexUSDT.URL
}

func (ep Endpoints) BinanceUSDTEndpoint() string {
	return ep.eps.BinanceUSDT.URL
}

func (ep Endpoints) BinancePAXEndpoint() string {
	return ep.eps.BinancePAX.URL
}

func (ep Endpoints) BinanceTUSDEndpoint() string {
	return ep.eps.BinanceTUSD.URL
}

func (ep Endpoints) CoinbaseDAIEndpoint() string {
	return ep.eps.CoinbaseDAI.URL
}

func (ep Endpoints) HitDaiEndpoint() string {
	return ep.eps.HitDai.URL
}

func (ep Endpoints) CoinbaseUSDEndpoint() string {
	return ep.eps.CoinbaseUSD.URL
}

// TODO: support simulation
func (ep Endpoints) CoinbaseUSDCEndpoint() string {
	return ep.eps.CoinbaseUSDC.URL
}

func (ep Endpoints) BinanceUSDCEndpoint() string {
	return ep.eps.BinanceUSDC.URL
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

func (ep Endpoints) CoinbaseBTCEndpoint() string {
	return ep.eps.CoinbaseBTC.URL
}

func (ep Endpoints) BinanceBTCEndpoint() string {
	return ep.eps.BinanceBTC.URL
}
