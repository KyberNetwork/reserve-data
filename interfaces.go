package reserve

import (
	"math/big"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/lib/rtypes"
	commonv3 "github.com/KyberNetwork/reserve-data/reservesetting/common"
)

// Data is the interface of of all data query methods.
// All methods' implementations must support concurrency.
type Data interface {
	CurrentPriceVersion(timestamp uint64) (common.Version, error)
	GetAllPrices(timestamp uint64) (common.AllPriceResponse, error)
	GetOnePrice(id rtypes.TradingPairID, timestamp uint64) (common.OnePriceResponse, error)

	CurrentAuthDataVersion(timestamp uint64) (common.Version, error)
	GetAuthData(timestamp uint64) (common.AuthDataResponseV3, error)

	// GetRate returns latest valid rates for all tokens that is before timestamp.
	GetRate(timestamp uint64) (common.AllRateResponse, error)
	// GetRates returns list of valid rates for all tokens that is collected between [fromTime, toTime).
	GetRates(fromTime, toTime uint64) ([]common.AllRateResponse, error)

	GetRecords(fromTime, toTime uint64) ([]common.ActivityRecord, error)
	GetPendingActivities() ([]common.ActivityRecord, error)

	GetGoldData(timepoint uint64) (common.GoldData, error)
	GetBTCData(timepoint uint64) (common.BTCData, error)
	GetUSDData(timepoint uint64) (common.USDData, error)

	GetTradeHistory(fromTime, toTime uint64) (common.AllTradeHistory, error)

	Run() error
	RunStorageController() error
	Stop() error
	GetAssetRateTriggers(fromTime uint64, toTime uint64) (map[rtypes.AssetID]int, error)
}

// Core is the interface that wrap around all interactions
// with exchanges and blockchain.
type Core interface {
	// place order
	Trade(
		exchange common.Exchange,
		tradeType string,
		pair commonv3.TradingPairSymbols,
		rate float64,
		amount float64) (id common.ActivityID, done float64, remaining float64, finished bool, err error)

	Deposit(
		exchange common.Exchange,
		asset commonv3.Asset,
		amount *big.Int,
		timestamp uint64) (common.ActivityID, error)

	Withdraw(
		exchange common.Exchange,
		token commonv3.Asset,
		amount *big.Int) (common.ActivityID, error)

	CancelOrders(orders []common.RequestOrder, exchange common.Exchange) map[string]common.CancelOrderResult

	// blockchain related action
	SetRates(tokens []commonv3.Asset, buys, sells []*big.Int, block *big.Int, afpMid []*big.Int, msgs []string, triggers []bool) (common.ActivityID, error)
	CancelSetRate() (common.ActivityID, error)
}
