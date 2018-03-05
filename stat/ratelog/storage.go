package ratelog

import (
	"github.com/KyberNetwork/reserve-data/common"
)

type Storage interface {
	GetEthRateLog(time uint64) []common.EthRateLog
	StoreEthRateLog(bulkEthRateLog []common.EthRateLog, timePoint uint64) error
}
