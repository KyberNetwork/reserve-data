package exchange

import "github.com/KyberNetwork/reserve-data/common"

type BinanceStorage interface {
	StoreTradeHistory(data common.ExchangeTradeHistory) error

	GetTradeHistory(fromTime, toTime uint64) (common.ExchangeTradeHistory, error)
	GetLastIDTradeHistory(exchange, pair string) (string, error)
}
