package stat

import "github.com/KyberNetwork/reserve-data/common"

type AnalyticStorage interface {
	UpdatePriceAnalyticData(timestamp uint64, value []byte) error
	GetPriceAnalyticData(fromTime uint64, toTime uint64) ([]common.AnalyticPriceResponse, error)
	ExportPruneExpired(currentTime uint64, fileName string) (uint64, error)
	BackupFile(fileName string) error
}
