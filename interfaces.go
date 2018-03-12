package reserve

import (
	"math/big"

	"github.com/KyberNetwork/reserve-data/common"
	ethereum "github.com/ethereum/go-ethereum/common"
)

// all of the functions must support concurrency
type ReserveStats interface {
	GetTradeLogs(fromTime uint64, toTime uint64) ([]common.TradeLog, error)
	GetAssetVolume(fromTime, toTime uint64, freq, asset string) (common.StatTicks, error)
	GetBurnFee(fromTime, toTime uint64, freq, reserveAddr string) (common.StatTicks, error)
	GetWalletFee(fromTime, toTime uint64, freq, reserveAddr, walletAddr string) (common.StatTicks, error)
	GetUserVolume(fromTime, toTime uint64, freq, userAddr string) (common.StatTicks, error)
	GetTradeSummary(fromTime, toTime uint64) (common.StatTicks, error)

	GetCapByUser(userID string) (*common.UserCap, error)
	GetCapByAddress(addr ethereum.Address) (*common.UserCap, error)
	ExceedDailyLimit(addr ethereum.Address) (bool, error)
	GetPendingAddresses() ([]string, error)

	GetReserveRates(fromTime, toTime uint64, reserveAddr ethereum.Address) ([]common.ReserveRates, error)

	UpdateUserAddresses(userID string, addresses []ethereum.Address) error

	Run() error
	Stop() error
}

// all of the functions must support concurrency
type ReserveData interface {
	CurrentPriceVersion(timestamp uint64) (common.Version, error)
	GetAllPrices(timestamp uint64) (common.AllPriceResponse, error)
	GetOnePrice(id common.TokenPairID, timestamp uint64) (common.OnePriceResponse, error)

	CurrentAuthDataVersion(timestamp uint64) (common.Version, error)
	GetAuthData(timestamp uint64) (common.AuthDataResponse, error)

	CurrentRateVersion(timestamp uint64) (common.Version, error)
	GetRate(timestamp uint64) (common.AllRateResponse, error)
	GetRates(fromTime, toTime uint64) ([]common.AllRateResponse, error)

	GetRecords(fromTime, toTime uint64) ([]common.ActivityRecord, error)
	GetPendingActivities() ([]common.ActivityRecord, error)

	GetTradeHistory(timepoint uint64) (common.AllTradeHistory, error)

	Run() error
	Stop() error
}

type ReserveCore interface {
	// place order
	Trade(
		exchange common.Exchange,
		tradeType string,
		base common.Token,
		quote common.Token,
		rate float64,
		amount float64,
		timestamp uint64) (id common.ActivityID, done float64, remaining float64, finished bool, err error)

	Deposit(
		exchange common.Exchange,
		token common.Token,
		amount *big.Int,
		timestamp uint64) (common.ActivityID, error)

	Withdraw(
		exchange common.Exchange,
		token common.Token,
		amount *big.Int,
		timestamp uint64) (common.ActivityID, error)

	CancelOrder(id common.ActivityID, exchange common.Exchange) error

	// blockchain related action
	SetRates(tokens []common.Token, buys, sells []*big.Int, block *big.Int, afpMid []*big.Int) (common.ActivityID, error)

	GetAddresses() *common.Addresses
}
