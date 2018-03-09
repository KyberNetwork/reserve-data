package storage

import (
	"errors"
	"github.com/KyberNetwork/reserve-data/common"
	"sync"
)

type RamAuthStorage struct {
	mu      sync.RWMutex
	version int64
	data    map[int64]common.AuthDataSnapshot
}

func NewRamAuthStorage() *RamAuthStorage {
	return &RamAuthStorage{
		mu:      sync.RWMutex{},
		version: 0,
		data:    map[int64]common.AuthDataSnapshot{},
	}
}

func (self *RamAuthStorage) CurrentVersion() (int64, error) {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return self.version, nil
}

func (self *RamAuthStorage) GetSnapshot(version int64) (common.AuthDataSnapshot, error) {
	self.mu.RLock()
	defer self.mu.RUnlock()
	all, found := self.data[version]
	if found {
		return all, nil
	} else {
		return common.AuthDataSnapshot{}, errors.New("Version doesn't exist")
	}
}

func (self *RamAuthStorage) StoreNewSnapshot(data *common.AuthDataSnapshot) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.version = self.version + 1
	self.data[self.version] = *data
	delete(self.data, self.version-1)
	return nil
}
