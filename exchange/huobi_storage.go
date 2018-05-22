package exchange

import (
	"github.com/KyberNetwork/reserve-data/common"
)

type HuobiStorage interface {
	StoreIntermediateTx(id common.ActivityID, data common.TXEntry) error
	StorePendingIntermediateTx(id common.ActivityID, data common.TXEntry) error

	// get intermediate tx corresponding to the id. It can be done, failed or pending
	GetIntermedatorTx(id common.ActivityID) (common.TXEntry, error)
	GetPendingIntermediateTXs() (map[common.ActivityID]common.TXEntry, error)
	RemovePendingIntermediateTx(id common.ActivityID) error

	StoreTradeHistory(data common.ExchangeTradeHistory) error

	GetTradeHistory(fromTime, toTime uint64) (common.ExchangeTradeHistory, error)
	GetLastIDTradeHistory(exchange, pair string) (string, error)
}
