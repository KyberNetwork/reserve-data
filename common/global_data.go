package common

import (
	"encoding/json"

	"github.com/KyberNetwork/reserve-data/common/feed"
)

// OneForgeGoldData oneforge gold data
type OneForgeGoldData struct {
	Value     json.Number
	Text      string
	Timestamp uint64
	Error     bool
	Message   string
}

// GDAXGoldData gdax gold data
type GDAXGoldData struct {
	Valid   bool
	Error   string
	TradeID uint64 `json:"trade_id"`
	Price   string `json:"price"`
	Size    string `json:"size"`
	Bid     string `json:"bid"`
	Ask     string `json:"ask"`
	Volume  string `json:"volume"`
	Time    string `json:"time"`
}

// KrakenGoldData kraken gold data
type KrakenGoldData struct {
	Valid           bool
	Error           string        `json:"network_error"`
	ErrorFromKraken []interface{} `json:"error"`
	Result          map[string]struct {
		A []string `json:"a"`
		B []string `json:"b"`
		C []string `json:"c"`
		V []string `json:"v"`
		P []string `json:"p"`
		T []uint64 `json:"t"`
		L []string `json:"l"`
		H []string `json:"h"`
		O string   `json:"o"`
	} `json:"result"`
}

// GeminiGoldData godl data from gemini
type GeminiGoldData struct {
	Valid  bool
	Error  string
	Bid    string `json:"bid"`
	Ask    string `json:"ask"`
	Volume struct {
		ETH       string `json:"ETH"`
		USD       string `json:"USD"`
		Timestamp uint64 `json:"timestamp"`
	} `json:"volume"`
	Last string `json:"last"`
}

// GoldData gold data info
type GoldData struct {
	Timestamp   uint64
	OneForgeETH OneForgeGoldData
	OneForgeUSD OneForgeGoldData
	GDAX        GDAXGoldData
	Kraken      KrakenGoldData
	Gemini      GeminiGoldData
}

func (d GoldData) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"Timestamp":                  d.Timestamp,
		feed.OneForgeXAUETH.String(): d.OneForgeETH,
		feed.OneForgeXAUUSD.String(): d.OneForgeUSD,
		feed.GDAXETHUSD.String():     d.GDAX,
		feed.KrakenETHUSD.String():   d.Kraken,
		feed.GeminiETHUSD.String():   d.Gemini,
	}
}
