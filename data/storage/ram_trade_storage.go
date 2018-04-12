package storage

import (
	"sync"

	"github.com/KyberNetwork/reserve-data/common"
)

type RamTradeStorage struct {
	mu   sync.RWMutex
	data common.AllTradeHistory
}

func NewRamTradeStorage() *RamTradeStorage {
	return &RamTradeStorage{
		mu:   sync.RWMutex{},
		data: common.AllTradeHistory{},
	}
}

func (self *RamTradeStorage) GetTradeHistory() (common.AllTradeHistory, error) {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return self.data, nil
}

func (self *RamTradeStorage) StoreTradeHistory(data common.AllTradeHistory) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.data = data
	return nil
}
