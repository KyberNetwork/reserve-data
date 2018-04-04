package stat

import "github.com/KyberNetwork/reserve-data/common"

type AnalyticStorage interface {
	UpdatePriceAnalyticData(timestamp uint64, value []byte) error
	GetPriceAnalyticData(fromTime uint64, toTime uint64) ([]common.AnalyticPriceResponse, error)
	ExportPruneExpired(timestamp uint64) (uint64, error)
}
