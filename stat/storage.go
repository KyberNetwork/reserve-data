package stat

import (
	"github.com/KyberNetwork/reserve-data/common"
)

type Storage interface {
	LastBlock() (uint64, error)
	GetTradeLogs(fromTime uint64, toTime uint64) ([]common.TradeLog, error)

	GetAssetVolume(fromTime uint64, toTime uint64, freq string, asset string) (common.StatTicks, error)
	GetBurnFee(fromTime uint64, toTime uint64, freq string, reserveAddr string) (common.StatTicks, error)
	GetWalletFee(fromTime uint64, toTime uint64, freq string, reserveAddr string, walletAddr string) (common.StatTicks, error)
	GetUserVolume(fromTime uint64, toTime uint64, freq string, userAddr string) (common.StatTicks, error)

	// all addr and user are case insensitive here
	StoreCatLog(l common.SetCatLog) error
	UpdateUserAddresses(user string, addresses []string) error
	// returns lowercased category of an address
	GetCategory(addr string) (string, error)
	GetAddressesOfUser(user string) ([]string, error)
	// returns lowercased user identity of the address
	GetUserOfAddress(addr string) (string, error)

	UpdateLogBlock(block uint64, timepoint uint64) error
	StoreTradeLog(stat common.TradeLog, timepoint uint64) error
	SetTradeStats(metric, freq string, t uint64, tradeStats common.TradeStats) error
}
