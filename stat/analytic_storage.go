package stat

import "github.com/KyberNetwork/reserve-data/common"

type AnalyticStorage interface {
	UpdatePriceAnalyticData(timestamp uint64, value []byte) error
	GetPriceAnalyticData(fromTime uint64, toTime uint64) ([]common.AnalyticPriceResponse, error)
	//ExportExpiredPriceAnalyticData: Write all expired records into a predetermined filepath
	//each record will be represented in JSON format, and seperates by endline character
	//Return: Number of records exported (uint64) and error
	ExportExpiredPriceAnalyticData(currentTime uint64, fileName string) (uint64, error)
	PruneExpiredPriceAnalyticData(currentTime uint64) (uint64, error)
}
