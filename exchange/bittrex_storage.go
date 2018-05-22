package exchange

import (
	"github.com/KyberNetwork/reserve-data/common"
)

// This storage is used to store deposit histories
// to check if a deposit history id is registered for
// a different deposit already
type BittrexStorage interface {
	IsNewBittrexDeposit(id uint64, actID common.ActivityID) bool
	RegisterBittrexDeposit(id uint64, actID common.ActivityID) error
	StoreTradeHistory(data common.ExchangeTradeHistory) error

	GetTradeHistory(fromTime, toTime uint64) (common.ExchangeTradeHistory, error)
	GetLastIDTradeHistory(exchange, pair string) (string, error)
}
