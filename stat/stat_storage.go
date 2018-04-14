package stat

import (
	"github.com/KyberNetwork/reserve-data/common"
)

type StatStorage interface {
	GetAssetVolume(fromTime, toTime uint64, freq, asset string) (common.StatTicks, error)
	GetBurnFee(fromTime, toTime uint64, freq, reserveAddr string) (common.StatTicks, error)
	GetWalletFee(fromTime, toTime uint64, freq, reserveAddr string, walletAddr string) (common.StatTicks, error)
	GetUserVolume(fromTime, toTime uint64, freq, userAddr string) (common.StatTicks, error)
	GetLastProcessedTradeLogTimepoint() (timepoint uint64, err error)
	GetWalletStats(fromTime, toTime uint64, walletAddr string, timezone int64) (common.StatTicks, error)
	SetWalletStat(walletStats map[string]common.MetricStatsTimeZone) error

	SetVolumeStat(volumeStat map[string]common.VolumeStatsTimeZone) error
	SetBurnFeeStat(burnFeeStat map[string]common.BurnFeeStatsTimeZone) error

	SetTradeStats(freq string, t uint64, tradeStats common.TradeStats) error
	SetWalletAddress(walletAddr string) error
	GetWalletAddress() ([]string, error)
	SetLastProcessedTradeLogTimepoint(timepoint uint64) error

	SetCountry(country string) error
	GetCountries() ([]string, error)
	SetCountryStat(countryStats map[string]common.MetricStatsTimeZone) error
	GetCountryStats(fromTime, toTime uint64, country string, tzparam int64) (common.StatTicks, error)

	SetFirstTradeEver(userAddrs map[string]uint64) error
	GetFirstTradeEver(userAddr string) uint64
	GetAllFirstTradeEver() (map[string]uint64, error)
	SetFirstTradeInDay(userAddr map[string]uint64) error
	GetFirstTradeInDay(userAddr string, timepoint uint64, timezone int64) uint64

	SetUserList(userInfos map[string]common.UserInfoTimezone) error
	GetUserList(fromTime, toTime uint64, timezone int64) (map[string]common.UserInfo, error)

	SetTradeSummary(stats map[string]common.MetricStatsTimeZone) error
	GetTradeSummary(fromTime, toTime uint64, timezone int64) (common.StatTicks, error)
}
