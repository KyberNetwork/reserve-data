package stat

import (
	"github.com/KyberNetwork/reserve-data/common"
)

type StatStorage interface {
	GetLastProcessedTradeLogTimepoint(statType string) (timepoint uint64, err error)
	SetLastProcessedTradeLogTimepoint(statType string, timepoint uint64) error

	SetVolumeStat(volumeStat map[string]common.VolumeStatsTimeZone, lastProcessedTimepoint uint64) error
	GetAssetVolume(fromTime, toTime uint64, freq, asset string) (common.StatTicks, error)
	GetUserVolume(fromTime, toTime uint64, freq, userAddr string) (common.StatTicks, error)

	SetBurnFeeStat(burnFeeStat map[string]common.BurnFeeStatsTimeZone, lastProcessedTimepoint uint64) error
	GetBurnFee(fromTime, toTime uint64, freq, reserveAddr string) (common.StatTicks, error)
	GetWalletFee(fromTime, toTime uint64, freq, reserveAddr string, walletAddr string) (common.StatTicks, error)

	SetWalletAddress(walletAddr string) error
	GetWalletAddress() ([]string, error)

	SetWalletStat(walletStats map[string]common.MetricStatsTimeZone, lastProcessedTimepoint uint64) error
	GetWalletStats(fromTime, toTime uint64, walletAddr string, timezone int64) (common.StatTicks, error)

	SetCountry(country string) error
	GetCountries() ([]string, error)
	SetCountryStat(countryStats map[string]common.MetricStatsTimeZone, lastProcessedTimepoint uint64) error
	GetCountryStats(fromTime, toTime uint64, country string, tzparam int64) (common.StatTicks, error)

	SetFirstTradeEver(userAddrs map[string]uint64, lastProcessedTimepoint uint64) error
	GetFirstTradeEver(userAddr string) uint64
	GetAllFirstTradeEver() (map[string]uint64, error)

	SetFirstTradeInDay(userAddr map[string]uint64, lastProcessedTimepoint uint64) error
	GetFirstTradeInDay(userAddr string, timepoint uint64, timezone int64) uint64

	SetTradeSummary(stats map[string]common.MetricStatsTimeZone, lastProcessedTimepoint uint64) error
	GetTradeSummary(fromTime, toTime uint64, timezone int64) (common.StatTicks, error)
}
