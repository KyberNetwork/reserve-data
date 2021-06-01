package fetcher

import (
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/lib/rtypes"
)

// Storage is the interface that wraps all database operations of fetcher.
type Storage interface {
	StorePrice(data common.AllPriceEntry, timepoint uint64) error
	StoreRate(data common.AllRateEntry, timepoint uint64) error
	StoreAuthSnapshot(data *common.AuthDataSnapshot, timepoint uint64) error

	GetPendingActivities() ([]common.ActivityRecord, error)
	UpdateActivity(id common.ActivityID, act common.ActivityRecord) error
	Record(action string, id common.ActivityID, destination string, params common.ActivityParams,
		result common.ActivityResult, estatus string, mstatus string, timepoint uint64, isPending bool, orgTime uint64) error
	GetActivity(exchangeID rtypes.ExchangeID, id string) (common.ActivityRecord, error)
	GetLatestSetRatesActivityMined() (common.ActivityRecord, error)

	CurrentAuthDataVersion(timepoint uint64) (common.Version, error)
	GetAuthData(common.Version) (common.AuthDataSnapshot, error)
	FindReplacedTx(actions []string, nonce uint64) (string, error)
}
