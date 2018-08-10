package stat

import (
	"github.com/KyberNetwork/reserve-data/common"
	ethereum "github.com/ethereum/go-ethereum/common"
)

// RateStorage is the storage interface of conversion rates.
type RateStorage interface {
	StoreReserveRates(reserveAddr ethereum.Address, rate common.ReserveRates, timepoint uint64) error
	GetReserveRates(fromTime, toTime uint64, reserveAddr ethereum.Address) ([]common.ReserveRates, error)
	// ExportExpiredRateData look from the first record in the database, find all record in that day,
	// zip and write it to the fileName, return first and last timepoint of these records, number of record and error if occurs.
	ExportExpiredRateData(currentTime uint64, fileName string) (uint64, uint64, uint64, error)
	PruneExpiredReserveRateData(toTimestamp uint64) (uint64, error)
}
