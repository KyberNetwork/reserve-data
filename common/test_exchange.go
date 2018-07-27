package common

import (
	"math/big"

	ethereum "github.com/ethereum/go-ethereum/common"
)

//TestExchange an exchange object for testing purpose
type TestExchange struct {
}

//ID return ID of TestExchange
//in this case is binance
func (testExchange TestExchange) ID() ExchangeID {
	return "binance"
}

//Address return deposit address of a token
//in this test is empty (cause we do not test deposit here)
func (testExchange TestExchange) Address(token Token) (address ethereum.Address, supported bool) {
	return ethereum.Address{}, true
}

//Withdraw return a withdraw id
func (testExchange TestExchange) Withdraw(token Token, amount *big.Int, address ethereum.Address, timepoint uint64) (string, error) {
	return "withdrawid", nil
}

//Trade calling trade activity
func (testExchange TestExchange) Trade(tradeType string, base Token, quote Token, rate float64, amount float64, timepoint uint64) (id string, done float64, remaining float64, finished bool, err error) {
	return "tradeid", 10, 5, false, nil
}

//CancelOrder cancel an order
func (testExchange TestExchange) CancelOrder(id, base, quote string) error {
	return nil
}
func (testExchange TestExchange) MarshalText() (text []byte, err error) {
	return []byte("bittrex"), nil
}
func (testExchange TestExchange) GetExchangeInfo(pair TokenPairID) (ExchangePrecisionLimit, error) {
	return ExchangePrecisionLimit{}, nil
}
func (testExchange TestExchange) GetFee() (ExchangeFees, error) {
	return ExchangeFees{}, nil
}
func (testExchange TestExchange) GetMinDeposit() (ExchangesMinDeposit, error) {
	return ExchangesMinDeposit{}, nil
}
func (testExchange TestExchange) GetInfo() (ExchangeInfo, error) {
	return ExchangeInfo{}, nil
}
func (testExchange TestExchange) TokenAddresses() (map[string]ethereum.Address, error) {
	return map[string]ethereum.Address{}, nil
}
func (testExchange TestExchange) UpdateDepositAddress(token Token, address string) error {
	return nil
}
func (testExchange TestExchange) GetTradeHistory(fromTime, toTime uint64) (ExchangeTradeHistory, error) {
	return ExchangeTradeHistory{}, nil
}

func (testExchange TestExchange) GetLiveExchangeInfos(tokenPairIDs []TokenPairID) (ExchangeInfo, error) {
	return ExchangeInfo{}, nil
}
