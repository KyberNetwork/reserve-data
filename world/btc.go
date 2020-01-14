package world

import (
	"github.com/KyberNetwork/reserve-data/common"
)

func (tw *TheWorld) getCoinbaseInfo(ep string) common.CoinbaseData {
	var (
		url    = ep
		result = common.CoinbaseData{}
	)

	err := tw.getPublic(url, &result)
	if err != nil {
		result.Error = err.Error()
		result.Valid = false
	} else {
		result.Valid = true
	}
	return result
}

func (tw *TheWorld) GetBTCInfo() (common.BTCData, error) {
	return common.BTCData{
		Coinbase: tw.getCoinbaseInfo(tw.endpoint.CoinbaseBTCEndpoint()),
		Binance:  tw.getBinanceInfo(tw.endpoint.BinanceBTCEndpoint()),
	}, nil
}
