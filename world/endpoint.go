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

// WorldEndpoint implement endpoint for testing in simulate.
type WorldEndpoint struct {
	eps config.WorldEndpoints
}

// NewWorldEndpoint ...
func NewWorldEndpoint(eps config.WorldEndpoints) *WorldEndpoint {
	return &WorldEndpoint{eps: eps}
}

func (ep WorldEndpoint) BitFinexUSDTEndpoint() string {
	return ep.eps.BitFinexUSDT.URL
}

func (ep WorldEndpoint) BinanceUSDTEndpoint() string {
	return ep.eps.BinanceUSDT.URL
}

func (ep WorldEndpoint) BinancePAXEndpoint() string {
	return ep.eps.BinancePAX.URL
}

func (ep WorldEndpoint) BinanceTUSDEndpoint() string {
	return ep.eps.BinanceTUSD.URL
}

func (ep WorldEndpoint) CoinbaseDAIEndpoint() string {
	return ep.eps.CoinbaseDAI.URL
}

func (ep WorldEndpoint) HitDaiEndpoint() string {
	return ep.eps.HitDai.URL
}

func (ep WorldEndpoint) CoinbaseUSDEndpoint() string {
	return ep.eps.CoinbaseUSD.URL
}

// TODO: support simulation
func (ep WorldEndpoint) CoinbaseUSDCEndpoint() string {
	return ep.eps.CoinbaseUSDC.URL
}

func (ep WorldEndpoint) BinanceUSDCEndpoint() string {
	return ep.eps.BinanceUSDC.URL
}

func (ep WorldEndpoint) GoldDataEndpoint() string {
	return ep.eps.GoldData.URL
}

func (ep WorldEndpoint) OneForgeGoldETHDataEndpoint() string {
	return ep.eps.OneForgeGoldETH.URL
}

func (ep WorldEndpoint) OneForgeGoldUSDDataEndpoint() string {
	return ep.eps.OneForgeGoldUSD.URL
}

func (ep WorldEndpoint) GDAXDataEndpoint() string {
	return ep.eps.GDAXData.URL
}

func (ep WorldEndpoint) KrakenDataEndpoint() string {
	return ep.eps.KrakenData.URL
}

func (ep WorldEndpoint) GeminiDataEndpoint() string {
	return ep.eps.GeminiData.URL
}

func (ep WorldEndpoint) CoinbaseBTCEndpoint() string {
	return ep.eps.CoinbaseBTC.URL
}

func (ep WorldEndpoint) GeminiBTCEndpoint() string {
	return ep.eps.GeminiBTC.URL
}
