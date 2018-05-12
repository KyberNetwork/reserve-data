package datapruner

import (
	"github.com/KyberNetwork/reserve-data/common/archive"
)

const (
	EXPIRED_AUTHDATA_PATH = "expired-auth-data/"
)

type StorageController struct {
	Runner              StorageControllerRunner
	Arch                archive.Archive
	ExpiredAuthDataPath string
}

func NewStorageController(storageControllerRunner StorageControllerRunner, arch archive.Archive) (StorageController, error) {
	storageController := StorageController{
		storageControllerRunner, arch, EXPIRED_AUTHDATA_PATH,
	}
	return storageController, nil
}
