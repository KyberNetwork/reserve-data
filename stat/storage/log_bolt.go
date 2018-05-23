package storage

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/boltdb/bolt"
)

const (
	MAX_GET_LOG_PERIOD uint64 = 86400000000000 //1 days in nanosecond
	TRADELOG_BUCKET    string = "logs"
	CATLOG_BUCKET      string = "cat_logs"
)

type BoltLogStorage struct {
	mu    sync.RWMutex
	db    *bolt.DB
	block uint64
}

func NewBoltLogStorage(path string) (*BoltLogStorage, error) {
	// init instance
	var (
		err error
		db  *bolt.DB
	)
	db, err = bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}
	// init buckets
	err = db.Update(func(tx *bolt.Tx) error {
		if _, uErr := tx.CreateBucketIfNotExists([]byte(TRADELOG_BUCKET)); uErr != nil {
			return uErr
		}
		_, uErr := tx.CreateBucketIfNotExists([]byte(CATLOG_BUCKET))
		return uErr
	})

	if err != nil {
		return nil, err
	}

	storage := &BoltLogStorage{sync.RWMutex{}, db, 0}
	err = storage.db.View(func(tx *bolt.Tx) error {
		block, vErr := storage.LoadLastLogIndex(tx)
		if vErr != nil {
			return vErr
		}
		storage.block = block
		return nil
	})
	if err != nil {
		return nil, err
	}
	return storage, nil
}

func (self *BoltLogStorage) MaxRange() uint64 {
	return MAX_GET_LOG_PERIOD
}

func (self *BoltLogStorage) LoadLastCatLog(tx *bolt.Tx) (common.SetCatLog, error) {
	b := tx.Bucket([]byte(CATLOG_BUCKET))
	c := b.Cursor()
	k, v := c.Last()
	record := common.SetCatLog{}
	if k == nil {
		return record, errors.New("Database is empty")
	}

	err := json.Unmarshal(v, &record)
	return record, err
}

func (self *BoltLogStorage) LoadLastLogIndex(tx *bolt.Tx) (uint64, error) {
	catBlock, _, err1 := self.LoadLastCatLogIndex()
	tradeBlock, _, err2 := self.LoadLastTradeLogIndex()
	if err1 != nil && err2 != nil {
		return 0, fmt.Errorf("last Cat Log err: %v and last Trade log err: %v ", err1, err2)
	}
	if catBlock > tradeBlock {
		return catBlock, nil
	}
	return tradeBlock, nil
}

// func (self *BoltLogStorage) LoadLastTradeLogIndex(tx *bolt.Tx) (uint64, uint, error) {
// 	b := tx.Bucket([]byte(TRADELOG_BUCKET))
// 	c := b.Cursor()
// 	k, v := c.Last()
// 	if k == nil {
// 		return 0, 0, nil
// 	}

// 	record := common.TradeLog{}
// 	if err := json.Unmarshal(v, &record); err != nil {
// 		return 0, 0, err
// 	}
// 	return record.BlockNumber, record.Index, nil
// }

func (self *BoltLogStorage) LoadLastTradeLogIndex() (block uint64, index uint, err error) {
	block = 0
	index = 0
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TRADELOG_BUCKET))
		c := b.Cursor()
		k, v := c.Last()
		if k == nil {
			return nil
		}
		record := common.TradeLog{}
		if err := json.Unmarshal(v, &record); err != nil {
			return err
		}
		block = record.BlockNumber
		index = record.Index
		return nil
	})
	return block, index, err
}

func (self *BoltLogStorage) LoadLastCatLogIndex() (block uint64, index uint, err error) {
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TRADELOG_BUCKET))
		c := b.Cursor()
		k, v := c.Last()
		if k == nil {
			return nil
		}
		record := common.SetCatLog{}
		if err := json.Unmarshal(v, &record); err != nil {
			return err
		}
		block = record.BlockNumber
		index = record.Index
		return nil
	})
	return block, index, err
}

// func (self *BoltLogStorage) LoadLastCatLogIndex(tx *bolt.Tx) (uint64, uint, error) {
// 	b := tx.Bucket([]byte(TRADELOG_BUCKET))
// 	c := b.Cursor()
// 	k, v := c.Last()
// 	if k == nil {
// 		return 0, 0, nil
// 	}

// 	record := common.SetCatLog{}
// 	if err := json.Unmarshal(v, &record); err != nil {
// 		return 0, 0, err
// 	}
// 	return record.BlockNumber, record.Index, nil
// }

func (self *BoltLogStorage) StoreCatLog(l common.SetCatLog) error {
	var err error
	err = self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(CATLOG_BUCKET))
		dataJson, uErr := json.Marshal(l)
		if uErr != nil {
			return uErr
		}
		block, index, uErr := self.LoadLastCatLogIndex()
		if uErr == nil && (block > l.BlockNumber || (block == l.BlockNumber && index >= l.Index)) {
			// TODO: logs the error, or embed on the returning error
			return fmt.Errorf("Duplicated cat log %+v (new block number %d is smaller or equal to latest block number %d and tx index %d is smaller or equal to last log tx index %d)", l, block, l.BlockNumber, index, l.Index)
		}
		// log.Printf("Storing cat log: %d", l.Timestamp)
		idByte := uint64ToBytes(l.Timestamp)
		return b.Put(idByte, dataJson)
	})
	return err
}

func (self *BoltLogStorage) StoreTradeLog(stat common.TradeLog, timepoint uint64) error {
	var err error
	err = self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TRADELOG_BUCKET))
		var dataJson []byte
		// block, index, uErr := self.LoadLastTradeLogIndex()
		// if uErr == nil && (block > stat.BlockNumber || (block == stat.BlockNumber && index >= stat.Index)) {
		// 	// TODO: logs the error, or embed on the returning error
		// 	return fmt.Errorf("Duplicated trade log %+v (new block number %d is smaller or equal to latest block number %d and tx index %d is smaller or equal to last log tx index %d)", stat, block, stat.BlockNumber, index, stat.Index)
		// }
		dataJson, uErr := json.Marshal(stat)
		if uErr != nil {
			return uErr
		}
		// log.Printf("Storing log: %d", stat.Timestamp)
		idByte := uint64ToBytes(stat.Timestamp)
		return b.Put(idByte, dataJson)
	})
	return err
}

func (self *BoltLogStorage) UpdateLogBlock(block uint64, timepoint uint64) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.block = block
	return nil
}

func (self *BoltLogStorage) GetLastCatLog() (common.SetCatLog, error) {
	var (
		err    error
		result common.SetCatLog
	)
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(CATLOG_BUCKET))
		c := b.Cursor()
		k, v := c.Last()
		if k == nil {
			return errors.New("there is no catlog")
		}
		return json.Unmarshal(v, &result)
	})
	return result, err
}

func (self *BoltLogStorage) GetFirstCatLog() (common.SetCatLog, error) {
	var (
		err    error
		result common.SetCatLog
	)
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(CATLOG_BUCKET))
		c := b.Cursor()
		k, v := c.First()
		if k == nil {
			return errors.New("there is no catlog")
		}
		return json.Unmarshal(v, &result)
	})
	return result, err
}

func (self *BoltLogStorage) GetCatLogs(fromTime uint64, toTime uint64) ([]common.SetCatLog, error) {
	var (
		err    error
		result = make([]common.SetCatLog, 0)
	)
	if toTime-fromTime > MAX_GET_LOG_PERIOD {
		err = fmt.Errorf("Time range is too broad, it must be smaller or equal to %d miliseconds", MAX_GET_RATES_PERIOD)
		return result, err
	}
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(CATLOG_BUCKET))
		c := b.Cursor()
		min := uint64ToBytes(fromTime)
		max := uint64ToBytes(toTime)
		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			record := common.SetCatLog{}
			if vErr := json.Unmarshal(v, &record); vErr != nil {
				return vErr
			}
			result = append([]common.SetCatLog{record}, result...)
		}
		return nil
	})
	return result, err
}

func (self *BoltLogStorage) GetLastTradeLog() (common.TradeLog, error) {
	var (
		err    error
		result common.TradeLog
	)
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TRADELOG_BUCKET))
		c := b.Cursor()
		k, v := c.Last()
		if k == nil {
			return errors.New("there is no tradelog")
		}
		return json.Unmarshal(v, &result)
	})
	return result, err
}

func (self *BoltLogStorage) GetFirstTradeLog() (common.TradeLog, error) {
	var (
		err    error
		result common.TradeLog
	)
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TRADELOG_BUCKET))
		c := b.Cursor()
		k, v := c.First()
		if k == nil {
			return errors.New("there is no tradelog")
		}
		return json.Unmarshal(v, &result)
	})
	return result, err
}

func (self *BoltLogStorage) GetTradeLogs(fromTime uint64, toTime uint64) ([]common.TradeLog, error) {
	var (
		err    error
		result = make([]common.TradeLog, 0)
	)
	if toTime-fromTime > MAX_GET_LOG_PERIOD {
		err = fmt.Errorf("Time range is too broad, it must be smaller or equal to %d miliseconds", MAX_GET_RATES_PERIOD)
		return result, err
	}
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TRADELOG_BUCKET))
		c := b.Cursor()
		min := uint64ToBytes(fromTime)
		max := uint64ToBytes(toTime)
		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			record := common.TradeLog{}
			if vErr := json.Unmarshal(v, &record); vErr != nil {
				return vErr
			}
			result = append(result, record)
		}
		return nil
	})
	return result, err
}

func (self *BoltLogStorage) LastBlock() (uint64, error) {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return self.block, nil
}
