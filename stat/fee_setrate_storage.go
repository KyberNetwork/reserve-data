package stat

import (
	"github.com/KyberNetwork/reserve-data/common"
)

type FeeSetRateStorage interface {
	GetLastBlockChecked() (uint64, error)
	StoreTransaction(tx common.TransactionInfo) error
	GetFeeSetRateByDay(fromTime, toTime uint64) ([]common.FeeSetRate, error)
}
