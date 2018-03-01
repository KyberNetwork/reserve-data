package common

import (
	"errors"
	"fmt"
	"math/big"

	ethereum "github.com/ethereum/go-ethereum/common"
)

type Exchange interface {
	ID() ExchangeID
	Address(token Token) (address ethereum.Address, supported bool)
	UpdateDepositAddress(token Token, addr string)
	Withdraw(token Token, amount *big.Int, address ethereum.Address, timepoint uint64) (string, error)
	Trade(tradeType string, base Token, quote Token, rate float64, amount float64, timepoint uint64) (id string, done float64, remaining float64, finished bool, err error)
	CancelOrder(id ActivityID) error
	MarshalText() (text []byte, err error)
	GetInfo() (ExchangeInfo, error)
	GetExchangeInfo(TokenPairID) (ExchangePrecisionLimit, error)
	GetFee() ExchangeFees
	TokenAddresses() map[string]ethereum.Address
	GetImtorMode() bool
}

var SupportedExchanges = map[ExchangeID]Exchange{}

func GetExchange(id string) (Exchange, error) {
	ex := SupportedExchanges[ExchangeID(id)]
	if ex == nil {
		return ex, errors.New(fmt.Sprintf("Exchange %s is not supported", id))
	} else {
		return ex, nil
	}
}

func MustGetExchange(id string) Exchange {
	result, err := GetExchange(id)
	if err != nil {
		panic(err)
	}
	return result
}
