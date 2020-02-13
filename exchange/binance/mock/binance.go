package mock

import (
	"errors"
	"math/big"

	ethereum "github.com/ethereum/go-ethereum/common"

	"github.com/KyberNetwork/reserve-data/common"
	commonv3 "github.com/KyberNetwork/reserve-data/reservesetting/common"
)

func init() {
	common.SupportedExchanges[common.Binance] = &BinanceTestExchange{}
}

// BinanceTestExchange is the mock implementation of binance exchange, for testing purpose.
type BinanceTestExchange struct{}

// Address binance mock address function
func (bte *BinanceTestExchange) Address(_ commonv3.Asset) (address ethereum.Address, supported bool) {
	return ethereum.Address{}, true
}

// Withdraw binance mock withdraw function
func (bte *BinanceTestExchange) Withdraw(asset commonv3.Asset, amount *big.Int, address ethereum.Address) (string, error) {
	return "withdrawid", nil
}

// Trade binance mock trade function
func (bte *BinanceTestExchange) Trade(tradeType string, pair commonv3.TradingPairSymbols, rate float64, amount float64) (id string, done float64, remaining float64, finished bool, err error) {
	return "tradeid", 10, 5, false, nil
}

// CancelOrder binance mock cancel order function
func (bte *BinanceTestExchange) CancelOrder(id, base, quote string) error {
	return nil
}

// MarshalText mock function
func (bte *BinanceTestExchange) MarshalText() (text []byte, err error) {
	return []byte("binance"), nil
}

// GetTradeHistory mock function
func (bte *BinanceTestExchange) GetTradeHistory(fromTime, toTime uint64) (common.ExchangeTradeHistory, error) {
	return common.ExchangeTradeHistory{}, nil
}

// OpenOrders get open orders for one pair
func (bte *BinanceTestExchange) OpenOrders(pair commonv3.TradingPairSymbols) ([]common.Order, error) {
	return nil, nil
}

// ID return binance exchange id
func (bte *BinanceTestExchange) ID() common.ExchangeID {
	return common.Binance
}

// GetLiveExchangeInfos of TestExchangeForSetting return a valid result for
func (bte *BinanceTestExchange) GetLiveExchangeInfos(tokenPairIDs []commonv3.TradingPairSymbols) (common.ExchangeInfo, error) {
	result := make(common.ExchangeInfo)
	for _, pairID := range tokenPairIDs {
		if pairID.ID != 1 {
			return result, errors.New("token pair ID is not support")
		}
		result[1] = common.ExchangePrecisionLimit{
			AmountLimit: common.TokenPairAmountLimit{
				Min: 1,
				Max: 900000,
			},
			Precision: common.TokenPairPrecision{
				Amount: 0,
				Price:  7,
			},
			PriceLimit: common.TokenPairPriceLimit{
				Min: 0.000192,
				Max: 0.019195,
			},
			MinNotional: 0.01,
		}
	}
	return result, nil
}
