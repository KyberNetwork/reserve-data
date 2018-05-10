package statpruner

import (
	"github.com/KyberNetwork/reserve-data/common/archive"
)

type StorageController struct {
	Runner ControllerRunner
	Arch   archive.Archive
}

func NewStorageController(storageControllerRunner ControllerRunner, arch archive.Archive) (StorageController, error) {
	storageController := StorageController{
		storageControllerRunner, arch,
	}
	return storageController, nil
}
