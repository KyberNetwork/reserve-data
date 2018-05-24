package common

import (
	"math/big"

	ethereum "github.com/ethereum/go-ethereum/common"
)

type TestExchange struct {
}

func (self TestExchange) ID() ExchangeID {
	return "binance"
}
func (self TestExchange) Address(token Token) (address ethereum.Address, supported bool) {
	return ethereum.Address{}, true
}
func (self TestExchange) Withdraw(token Token, amount *big.Int, address ethereum.Address, timepoint uint64) (string, error) {
	return "withdrawid", nil
}
func (self TestExchange) Trade(tradeType string, base Token, quote Token, rate float64, amount float64, timepoint uint64) (id string, done float64, remaining float64, finished bool, err error) {
	return "tradeid", 10, 5, false, nil
}
func (self TestExchange) CancelOrder(id, base, quote string) error {
	return nil
}
func (self TestExchange) MarshalText() (text []byte, err error) {
	return []byte("bittrex"), nil
}
func (self TestExchange) GetExchangeInfo(pair TokenPairID) (ExchangePrecisionLimit, error) {
	return ExchangePrecisionLimit{}, nil
}
func (self TestExchange) GetFee() ExchangeFees {
	return ExchangeFees{}
}
func (self TestExchange) GetMinDeposit() ExchangesMinDeposit {
	return ExchangesMinDeposit{}
}
func (self TestExchange) GetInfo() (*ExchangeInfo, error) {
	return &ExchangeInfo{}, nil
}
func (self TestExchange) TokenAddresses() map[string]ethereum.Address {
	return map[string]ethereum.Address{}
}
func (self TestExchange) UpdateDepositAddress(token Token, address string) {
}
