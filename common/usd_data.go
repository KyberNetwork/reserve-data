package common

import (
	"github.com/KyberNetwork/reserve-data/common/feed"
)

// BinanceData data return by https://api.binance.com/api/v3/ticker/bookTicker?symbol=ETHUSDC
type BinanceData struct {
	Valid    bool
	Error    string
	Symbol   string `json:"symbol"`
	BidPrice string `json:"bidPrice"`
	BidQty   string `json:"bidQty"`
	AskPrice string `json:"askPrice"`
	AskQty   string `json:"askQty"`
}

// USDData ...
type USDData struct {
	Timestamp             uint64
	CoinbaseETHUSDDAI5000 FeedProviderResponse `json:"CoinbaseETHUSDDAI5000"`
	CurveDAIUSDC10000     FeedProviderResponse `json:"CurveDAIUSDC10000"`
	BinanceETHUSDC10000   FeedProviderResponse `json:"BinanceETHUSDC10000"`
}

// ToMap convert to map result.
func (u USDData) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"Timestamp":                         u.Timestamp,
		feed.CoinbaseETHUSDDAI5000.String(): u.CoinbaseETHUSDDAI5000,
		feed.CurveDAIUSDC10000.String():     u.CurveDAIUSDC10000,
		feed.BinanceETHUSDC10000.String():   u.BinanceETHUSDC10000,
	}
}
