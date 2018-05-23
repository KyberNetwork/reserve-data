package stat

import (
	"github.com/KyberNetwork/reserve-data/common"
)

type LogStorage interface {
	StoreCatLog(l common.SetCatLog) error
	GetCatLogs(fromTime uint64, toTime uint64) ([]common.SetCatLog, error)
	GetFirstCatLog() (common.SetCatLog, error)
	GetLastCatLog() (common.SetCatLog, error)

	StoreTradeLog(stat common.TradeLog, timepoint uint64) error
	GetTradeLogs(fromTime uint64, toTime uint64) ([]common.TradeLog, error)
	GetFirstTradeLog() (common.TradeLog, error)
	GetLastTradeLog() (common.TradeLog, error)

	UpdateLogBlock(block uint64, timepoint uint64) error
	MaxRange() uint64
	LastBlock() (uint64, error)
	LoadLastTradeLogIndex() (block uint64, index uint, err error)
	LoadLastCatLogIndex() (block uint64, index uint, err error)
}
