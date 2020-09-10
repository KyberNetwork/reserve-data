package data

import (
	"github.com/KyberNetwork/reserve-data/common"
)

// GlobalStorage is the interfaces that wraps database operations of real world
// pricing information of ReserveData.
type GlobalStorage interface {
	GetGoldInfo(version common.Version) (common.GoldData, error)
	CurrentGoldInfoVersion(timepoint uint64) (common.Version, error)

	GetBTCInfo(version common.Version) (common.BTCData, error)
	GetUSDInfo(version common.Version) (common.USDData, error)
	CurrentBTCInfoVersion(timepoint uint64) (common.Version, error)
	CurrentUSDInfoVersion(timepoint uint64) (common.Version, error)

	UpdateFeedConfiguration(string, bool) error
	GetFeedConfiguration() ([]common.FeedConfiguration, error)
	StorePendingFeedSetting(value []byte) error
	ConfirmPendingFeedSetting(value []byte) error
	RejectPendingFeedSetting() error
	GetPendingFeedSetting() (common.MapFeedSetting, error)
	GetFeedSetting() (common.MapFeedSetting, error)

	UpdateFetcherConfiguration(common.FetcherConfiguration) error
	GetAllFetcherConfiguration() (common.FetcherConfiguration, error)

	SetGasThreshold(v common.GasThreshold) error
	GetGasThreshold() (common.GasThreshold, error)

	SetPreferGasSource(v common.PreferGasSource) error
	GetPreferGasSource() (common.PreferGasSource, error)
}
