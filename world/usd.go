package world

import (
	"github.com/KyberNetwork/reserve-data/common"
)

func (tw *TheWorld) GetUSDInfo() (common.USDData, error) {
	return common.USDData{
		Timestamp:             0,
		CoinbaseETHUSD10000:   tw.getFeedInfo(tw.endpoint.CoinbaseETHUSD10000Endpoint()),
		GeminiETHUSD10000:     tw.getFeedInfo(tw.endpoint.GeminiETHUSD10000Endpoint()),
		CoinbaseETHUSDC10000:  tw.getFeedInfo(tw.endpoint.CoinbaseETHUSDC10000Endpoint()),
		BinanceETHUSDC10000:   tw.getFeedInfo(tw.endpoint.BinanceETHUSDC10000Endpoint()),
		CoinbaseETHUSDDAI5000: tw.getFeedInfo(tw.endpoint.CoinbaseETHUSDDAI5000Endpoint()),
		BitfinexETHUSDT10000:  tw.getFeedInfo(tw.endpoint.BitfinexETHUSDT10000Endpoint()),
		BinanceETHPAX5000:     tw.getFeedInfo(tw.endpoint.BinanceETHPAX5000Endpoint()),
		BinanceETHUSDT10000:   tw.getFeedInfo(tw.endpoint.BinanceETHUSDT10000Endpoint()),
		BinanceETHBUSD10000:   tw.getFeedInfo(tw.endpoint.BinanceETHBUSD10000Endpoint()),
	}, nil
}
