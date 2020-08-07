package world

import (
	"github.com/KyberNetwork/reserve-data/common"
)

func (tw *TheWorld) GetBTCInfo() (common.BTCData, error) {
	return common.BTCData{
		CoinbaseETHBTC3: tw.getFeedInfo(tw.endpoint.CoinbaseETHBTC3Endpoint()),
		BinanceETHBTC3:  tw.getFeedInfo(tw.endpoint.BinanceETHBTC3Endpoint()),
	}, nil
}
