package exchange

import "github.com/KyberNetwork/reserve-data/common"

// BinanceStorage is the interface that wraps all database operation of Binance exchange.
type BinanceStorage interface {
	StoreTradeHistory(data common.ExchangeTradeHistory) error

	GetTradeHistory(exchangeID, fromTime, toTime uint64) (common.ExchangeTradeHistory, error)
	GetLastIDTradeHistory(pairID uint64) (string, error)
}
