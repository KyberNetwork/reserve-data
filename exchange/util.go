package exchange

import (
	"github.com/KyberNetwork/reserve-data/common"
)

func getExchangePairsAndFeesFromConfig(
	addressConfig map[string]string,
	feeConfig common.ExchangeFees,
	exchange string) ([]common.TokenPair, common.ExchangeFees) {

	pairs := []common.TokenPair{}
	fees := common.ExchangeFees{
		feeConfig.Trading,
		common.FundingFee{
			map[string]float64{},
			map[string]float64{},
		},
	}
	for tokenID := range addressConfig {
		if tokenID != "ETH" {
			pair := common.MustCreateTokenPair(tokenID, "ETH")
			pairs = append(pairs, pair)
		}
		if _, exist := feeConfig.Funding.Withdraw[tokenID]; exist {
			fees.Funding.Withdraw[tokenID] = feeConfig.Funding.Withdraw[tokenID] * 2
		} else {
			panic(tokenID + " is not found in " + exchange + " withdraw fee config file")
		}
		if _, exist := feeConfig.Funding.Deposit[tokenID]; exist {
			fees.Funding.Deposit[tokenID] = feeConfig.Funding.Deposit[tokenID]
		} else {
			panic(tokenID + " is not found in " + exchange + " binance deposit fee config file")
		}
	}
	return pairs, fees
}
