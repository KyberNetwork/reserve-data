package common

import (
	"math/big"

	ethereum "github.com/ethereum/go-ethereum/common"

	"github.com/KyberNetwork/reserve-data/reservesetting/common"
)

//TestExchange is mock exchange for test
type TestExchange struct {
}

// ID return exchange id
func (te TestExchange) ID() ExchangeID {
	return Binance
}

//Address function for test
func (te TestExchange) Address(asset common.Asset) (address ethereum.Address, supported bool) {
	return ethereum.Address{}, true
}

//Withdraw mock function
func (te TestExchange) Withdraw(asset common.Asset, amount *big.Int, address ethereum.Address) (string, error) {
	return "withdrawid", nil
}

// Trade mock function
func (te TestExchange) Trade(tradeType string, pair common.TradingPairSymbols, rate float64, amount float64) (id string, done float64, remaining float64, finished bool, err error) {
	return "tradeid", 10, 5, false, nil
}

// CancelOrder mock function
func (te TestExchange) CancelOrder(id, base, quote string) error {
	return nil
}

// MarshalText mock function
func (te TestExchange) MarshalText() (text []byte, err error) {
	return []byte("bittrex"), nil
}

// GetTradeHistory mock function
func (te TestExchange) GetTradeHistory(fromTime, toTime uint64) (ExchangeTradeHistory, error) {
	return ExchangeTradeHistory{}, nil
}

// GetLiveExchangeInfos mock function
func (te TestExchange) GetLiveExchangeInfos(pair []common.TradingPairSymbols) (ExchangeInfo, error) {
	return ExchangeInfo{}, nil
}

// OpenOrders mock open orders from binance
func (te TestExchange) OpenOrders(pair common.TradingPairSymbols) ([]Order, error) {
	return nil, nil
}
