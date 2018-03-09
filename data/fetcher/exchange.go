package fetcher

import (
	"github.com/KyberNetwork/reserve-data/common"
	ethereum "github.com/ethereum/go-ethereum/common"
)

type Exchange interface {
	ID() common.ExchangeID
	Name() string
	TokenPairs() []common.TokenPair
	FetchPriceData(timepoint uint64) (map[common.TokenPairID]common.ExchangePrice, error)
	FetchEBalanceData(timepoint uint64) (common.EBalanceEntry, error)
	FetchTradeHistory(timepoint uint64) (map[common.TokenPairID][]common.TradeHistory, error)
	// FetchOrderData(timepoint uint64) (common.OrderEntry, error)
	OrderStatus(id common.ActivityID, timepoint uint64) (string, error)
	DepositStatus(id common.ActivityID, timepoint uint64) (string, error)
	WithdrawStatus(id common.ActivityID, timepoint uint64) (string, string, error)
	TokenAddresses() map[string]ethereum.Address
}
