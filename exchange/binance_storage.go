package exchange

import (
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/lib/rtypes"
	common2 "github.com/ethereum/go-ethereum/common"
)

// BinanceStorage is the interface that wraps all database operation of Binance exchange.
type BinanceStorage interface {
	StoreTradeHistory(data common.ExchangeTradeHistory) error

	GetTradeHistory(exchangeID rtypes.ExchangeID, fromTime, toTime uint64) (common.ExchangeTradeHistory, error)
	GetLastIDTradeHistory(pairID rtypes.TradingPairID) (string, error)

	StoreIntermediateDeposit(id common.ActivityID, activity common.IntermediateTX) error
	GetIntermediateTX(id common.ActivityID) ([]common.IntermediateTX, error)
	SetDoneIntermediateTX(id common.ActivityID, hash common2.Hash, status string) error
}
