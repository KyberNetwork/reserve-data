package storagecontroller

import (
	"github.com/KyberNetwork/reserve-data/common/archive"
)

type StorageController struct {
	Runner StorageControllerRunner
	Arch   archive.Archive
}

func NewStorageController(storageControllerRunner StorageControllerRunner, arch archive.Archive) (StorageController, error) {
	storageController := StorageController{
		storageControllerRunner, arch,
	}
	return storageController, nil
}
