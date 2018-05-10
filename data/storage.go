package data

import (
	"github.com/KyberNetwork/reserve-data/common"
)

type Storage interface {
	CurrentPriceVersion(timepoint uint64) (common.Version, error)
	GetAllPrices(common.Version) (common.AllPriceEntry, error)
	GetOnePrice(common.TokenPairID, common.Version) (common.OnePrice, error)

	CurrentAuthDataVersion(timepoint uint64) (common.Version, error)
	GetAuthData(common.Version) (common.AuthDataSnapshot, error)
	ExportExpiredAuthData(timepoint uint64, fileName string) (uint64, error)

	CurrentRateVersion(timepoint uint64) (common.Version, error)
	GetRate(common.Version) (common.AllRateEntry, error)
	GetRates(fromTime, toTime uint64) ([]common.AllRateEntry, error)

	GetAllRecords(fromTime, toTime uint64) ([]common.ActivityRecord, error)
	GetPendingActivities() ([]common.ActivityRecord, error)

	GetTradeHistory(timepoint uint64) (common.AllTradeHistory, error)
	GetExchangeStatus() (common.ExchangesStatus, error)
	UpdateExchangeStatus(data common.ExchangesStatus) error

	UpdateExchangeNotification(exchange, action, tokenPair string, fromTime, toTime uint64, isWarning bool, msg string) error
	GetExchangeNotifications() (common.ExchangeNotifications, error)
}
