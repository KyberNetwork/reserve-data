package fetcher

import (
	"github.com/KyberNetwork/reserve-data/common"
)

// Storage is the interface that wraps all database operations of fetcher.
type Storage interface {
	StorePrice(data common.AllPriceEntry, timepoint uint64) error
	StoreRate(data common.AllRateEntry, timepoint uint64) error
	StoreAuthSnapshot(data *common.AuthDataSnapshot, timepoint uint64) error

	GetPendingActivities() ([]common.ActivityRecord, error)
	UpdateActivity(id common.ActivityID, act common.ActivityRecord) error

	CurrentAuthDataVersion(timepoint uint64) (common.Version, error)
	GetAuthData(common.Version) (common.AuthDataSnapshot, error)
	FindReplacedTx(action string, nonce uint64) (string, error)
}
