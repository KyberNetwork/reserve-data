package stat

import (
	"github.com/KyberNetwork/reserve-data/common"
)

type RateStorage interface {
	StoreReserveRates(reserveAddr string, rate common.ReserveRates, timepoint uint64) error

	GetReserveRates(fromTime, toTime uint64, reserveAddr string) ([]common.ReserveRates, error)
}
