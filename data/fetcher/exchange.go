package fetcher

import (
	"math/big"

	ethereum "github.com/ethereum/go-ethereum/common"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/lib/rtypes"
	commonv3 "github.com/KyberNetwork/reserve-data/reservesetting/common"
)

// Exchange is the common interface of centralized exchanges.
type Exchange interface {
	ID() rtypes.ExchangeID
	FetchPriceData(timepoint uint64) (map[rtypes.TradingPairID]common.ExchangePrice, error)
	FetchEBalanceData(timepoint uint64) (common.EBalanceEntry, error)
	// FetchTradeHistory gets history and save data in the exchange db
	FetchTradeHistory()

	OrderStatus(id, base, quote string) (string, float64, error)
	DepositStatus(id common.ActivityID, txHash string, assetID rtypes.AssetID, amount float64, timepoint uint64) (string, error)
	WithdrawStatus(id string, assetID rtypes.AssetID, amount float64, timepoint uint64) (string, string, float64, error)
	TokenAddresses() (map[rtypes.AssetID]ethereum.Address, error)
	Transfer(fromAccount string, toAccount string, asset commonv3.Asset, amount *big.Int, runAsync bool, referenceID string) (string, error)
}
