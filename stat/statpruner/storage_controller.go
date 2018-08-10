package statpruner

import (
	"github.com/KyberNetwork/reserve-data/common/archive"
)

const (
	expiredAnalyticPath string = "expired-analytic-data"
	expiredRatePath     string = "expired-rate"
)

type StorageController struct {
	Runner                   ControllerRunner
	Arch                     archive.Archive
	ExpiredPriceAnalyticPath string
	ExpiredRatePath          string
}

func NewStorageController(storageControllerRunner ControllerRunner, arch archive.Archive) (StorageController, error) {
	storageController := StorageController{
		storageControllerRunner, arch, expiredAnalyticPath, expiredRatePath,
	}
	return storageController, nil
}
