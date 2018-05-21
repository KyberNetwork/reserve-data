package storage

import (
	"fmt"
	"github.com/KyberNetwork/reserve-data/common"
	"sync"
)

type RamPriceStorage struct {
	mu      sync.RWMutex
	version int64
	data    map[int64]common.AllPriceEntry
}

func NewRamPriceStorage() *RamPriceStorage {
	return &RamPriceStorage{
		mu:      sync.RWMutex{},
		version: 0,
		data:    map[int64]common.AllPriceEntry{},
	}
}

func (self *RamPriceStorage) CurrentVersion() (int64, error) {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return self.version, nil
}

func (self *RamPriceStorage) GetAllPrices(version int64) (common.AllPriceEntry, error) {
	self.mu.RLock()
	defer self.mu.RUnlock()
	all := self.data[version]
	if all.Data == nil {
		return common.AllPriceEntry{}, fmt.Errorf("Version doesn't exist")
	} else {
		return all, nil
	}
}

func (self *RamPriceStorage) GetOnePrice(pair common.TokenPairID, version int64) (common.OnePrice, error) {
	self.mu.RLock()
	defer self.mu.RUnlock()
	all := self.data[version]
	if all.Data == nil {
		return common.OnePrice{}, fmt.Errorf("Version doesn't exist")
	} else {
		data := all.Data[pair]
		if len(data) == 0 {
			return common.OnePrice{}, fmt.Errorf("Pair of token is not supported")
		} else {
			return data, nil
		}
	}
}

func (self *RamPriceStorage) StoreNewData(data common.AllPriceEntry) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.version = self.version + 1
	self.data[self.version] = data
	delete(self.data, self.version-1)
	return nil
}
