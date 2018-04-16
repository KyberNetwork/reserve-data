package stat

import (
	"github.com/KyberNetwork/reserve-data/common"
	ethereum "github.com/ethereum/go-ethereum/common"
)

type StatStorage interface {
	GetAssetVolume(fromTime, toTime uint64, freq string, assetAddr ethereum.Address) (common.StatTicks, error)
	GetBurnFee(fromTime, toTime uint64, freq string, reserveAddr ethereum.Address) (common.StatTicks, error)
	GetWalletFee(fromTime, toTime uint64, freq string, reserveAddr ethereum.Address, walletAddr ethereum.Address) (common.StatTicks, error)
	GetUserVolume(fromTime, toTime uint64, freq string, userAddr ethereum.Address) (common.StatTicks, error)
	GetWalletStats(fromTime, toTime uint64, walletAddr ethereum.Address, timezone int64) (common.StatTicks, error)
	GetLastProcessedTradeLogTimepoint(statType string) (timepoint uint64, err error)
	SetLastProcessedTradeLogTimepoint(statType string, timepoint uint64) error

	SetVolumeStat(volumeStat map[string]common.VolumeStatsTimeZone, lastProcessedTimepoint uint64) error
	SetBurnFeeStat(burnFeeStat map[string]common.BurnFeeStatsTimeZone, lastProcessedTimepoint uint64) error

	SetWalletAddress(walletAddr ethereum.Address) error
	GetWalletAddress() ([]string, error)

	SetWalletStat(walletStats map[string]common.MetricStatsTimeZone, lastProcessedTimepoint uint64) error
	SetCountry(country string) error
	GetCountries() ([]string, error)
	SetCountryStat(countryStats map[string]common.MetricStatsTimeZone, lastProcessedTimepoint uint64) error
	GetCountryStats(fromTime, toTime uint64, country string, tzparam int64) (common.StatTicks, error)

	SetFirstTradeEver(tradeLogs *[]common.TradeLog, lastProcessTimePoint uint64) error
	GetFirstTradeEver(userAddr ethereum.Address) uint64
	GetAllFirstTradeEver() (map[ethereum.Address]uint64, error)
	SetFirstTradeInDay(tradeLogs *[]common.TradeLog, lastProcessTimePoint uint64) error
	GetFirstTradeInDay(userAddr ethereum.Address, timepoint uint64, timezone int64) uint64

	SetTradeSummary(stats map[string]common.MetricStatsTimeZone, lastProcessedTimepoint uint64) error
	GetTradeSummary(fromTime, toTime uint64, timezone int64) (common.StatTicks, error)
}
