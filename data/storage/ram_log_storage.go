package storage

import (
	"container/list"
	"errors"
	"sync"

	"github.com/KyberNetwork/reserve-data/common"
)

type RamLogStorage struct {
	mu      sync.RWMutex
	block   uint64
	records *list.List
}

func NewRamLogStorage() *RamLogStorage {
	return &RamLogStorage{
		sync.RWMutex{}, 0, list.New(),
	}
}

func (self *RamLogStorage) UpdateLogBlock(block uint64, timepoint uint64) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.block = block
	return nil
}

func (self *RamLogStorage) LastBlock() (uint64, error) {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return self.block, nil
}

func (self *RamLogStorage) GetTradeLogs(fromTime uint64, toTime uint64) ([]common.TradeLog, error) {
	self.mu.RLock()
	defer self.mu.RUnlock()
	result := []common.TradeLog{}
	ele := self.records.Back()
	for {
		if ele == nil {
			break
		} else {
			record := ele.Value.(common.TradeLog)
			if record.Timestamp >= fromTime && record.Timestamp <= toTime {
				result = append(result, record)
				ele = ele.Prev()
			} else if record.Timestamp < fromTime {
				break
			}
		}
	}
	return result, nil
}

func (self *RamLogStorage) StoreTradeLog(stat common.TradeLog, timepoint uint64) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	ele := self.records.Back()
	if ele != nil {
		record := ele.Value.(common.TradeLog)
		if record.BlockNumber >= stat.BlockNumber {
			return errors.New("Duplicated log (new block number is smaller or equal to latest block number)")
		}
	}
	self.records.PushBack(stat)
	return nil
}
