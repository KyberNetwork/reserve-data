package world

import (
	"github.com/KyberNetwork/reserve-data/common"
)

func (tw *TheWorld) getCoinbaseInfo() common.CoinbaseData {
	var (
		url    = tw.endpoint.CoinbaseBTCEndpoint()
		result = common.CoinbaseData{
			Valid: true,
		}
	)
	if err := tw.getPublic(url, &result); err != nil {
		result.Valid = false
		result.Error = err.Error()
	}
	return result
}

func (tw *TheWorld) getGeminiInfo() common.GeminiData {
	var (
		url    = tw.endpoint.GeminiBTCEndpoint()
		result = common.GeminiData{
			Valid: true,
		}
	)
	if err := tw.getPublic(url, &result); err != nil {
		result.Valid = false
		result.Error = err.Error()
	}
	return result
}

func (tw *TheWorld) GetBTCInfo() (common.BTCData, error) {
	return common.BTCData{
		Coinbase: tw.getCoinbaseInfo(),
		Gemini:   tw.getGeminiInfo(),
	}, nil
}
