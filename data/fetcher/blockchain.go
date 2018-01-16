package fetcher

import (
	"github.com/KyberNetwork/reserve-data/common"
	ethereum "github.com/ethereum/go-ethereum/common"
)

type Blockchain interface {
	FetchBalanceData(addr ethereum.Address, timepoint uint64) (map[string]common.BalanceEntry, error)
	FetchRates(timepoint uint64) (common.AllRateEntry, error)
	IsMined(tx ethereum.Hash) (bool, error)
	CurrentBlock() (uint64, error)
	GetLogs(fromBlock uint64, timepoint uint64) ([]common.TradeLog, error)
}
