package stat

import (
	"github.com/KyberNetwork/reserve-data/common"
)

type Blockchain interface {
	CurrentBlock() (uint64, error)
	GetLogs(fromBlock uint64, toBlock uint64, ethRate float64) ([]common.KNLog, error)
}
