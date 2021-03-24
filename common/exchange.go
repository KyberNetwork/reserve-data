package common

import (
	"fmt"
	"math/big"

	ethereum "github.com/ethereum/go-ethereum/common"

	"github.com/KyberNetwork/reserve-data/lib/rtypes"
	"github.com/KyberNetwork/reserve-data/reservesetting/common"
)

// ValidExchangeNames returns all valid exchange names.
var ValidExchangeNames = map[string]rtypes.ExchangeID{
	rtypes.Binance.String():  rtypes.Binance,
	rtypes.Huobi.String():    rtypes.Huobi,
	rtypes.Binance2.String(): rtypes.Binance2,
}

// Exchange represents a centralized exchange like Binance, Huobi...
type Exchange interface {
	ID() rtypes.ExchangeID
	// Address return the deposit address of an asset and return true
	// if token is supported in the exchange, otherwise return false.
	// This function will prioritize live address from exchange above the current stored address.
	Address(asset common.Asset) (address ethereum.Address, supported bool)
	Withdraw(asset common.Asset, amount *big.Int, address ethereum.Address) (string, error)
	Trade(tradeType string, pair common.TradingPairSymbols, rate, amount float64) (id string, done, remaining float64, finished bool, err error)

	// OpenOrders return open orders from exchange
	OpenOrders(pair common.TradingPairSymbols) (orders []Order, err error)
	CancelOrder(id, symbol string) error
	CancelAllOrders(symbol string) error
	MarshalText() (text []byte, err error)

	GetTradeHistory(fromTime, toTime uint64) (ExchangeTradeHistory, error)

	LiveExchange
	Transfer(fromAccount string, toAccount string, asset common.Asset, amount *big.Int, runAsync bool, referenceID string) (string, error)
}

// LiveExchange interface
// TODO: choose a better name as this interface for activity which does not affect
//
type LiveExchange interface {
	// GetLiveExchangeInfo querry the Exchange Endpoint for exchange precision and limit of a list of tokenPairIDs
	// It return error if occurs.
	GetLiveExchangeInfos([]common.TradingPairSymbols) (ExchangeInfo, error)
	// GetLiveWithdrawFee return withdraw fee of asset
	GetLiveWithdrawFee(asset string) (float64, error)
}

// SupportedExchanges map exchange id to its exchange
var SupportedExchanges = map[rtypes.ExchangeID]Exchange{}

// GetExchange return exchange by its name
func GetExchange(name string) (Exchange, error) {
	var (
		ex Exchange
	)
	exchangeID, exist := ValidExchangeNames[name]
	if !exist {
		return ex, fmt.Errorf("exchange %s does not exist", name)
	}
	ex = SupportedExchanges[exchangeID]
	if ex == nil {
		return ex, fmt.Errorf("exchange %s is not supported", name)
	}
	return ex, nil
}
