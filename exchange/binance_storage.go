package exchange

import (
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/lib/rtypes"
)

// BinanceStorage is the interface that wraps all database operation of Binance exchange.
type BinanceStorage interface {
	StoreTradeHistory(data common.ExchangeTradeHistory) error

	GetTradeHistory(exchangeID rtypes.ExchangeID, fromTime, toTime uint64) (common.ExchangeTradeHistory, error)
	GetLastIDTradeHistory(pairID rtypes.TradingPairID) (string, error)

	StoreIntermediateDeposit(id common.ActivityID, activity common.TXEntry, isPending bool) error
	GetPendingIntermediateTx(id common.ActivityID) (common.TXEntry, error)
}
