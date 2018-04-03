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
	MAX_GET_LOG_PERIOD uint64 = 86400000000000 //1 days in milisec

	TRADELOG_BUCKET string = "logs"
	CATLOG_BUCKET   string = "cat_logs"
)

type BoltLogStorage struct {
	mu    sync.RWMutex
	db    *bolt.DB
	block uint64
}

func NewBoltLogStorage(path string) (*BoltLogStorage, error) {
	// init instance
	var err error
	var db *bolt.DB
	db, err = bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}
	// init buckets
	db.Update(func(tx *bolt.Tx) error {
		tx.CreateBucket([]byte(TRADELOG_BUCKET))
		tx.CreateBucket([]byte(CATLOG_BUCKET))
		return nil
	})
	storage := &BoltLogStorage{sync.RWMutex{}, db, 0}
	storage.db.View(func(tx *bolt.Tx) error {
		block, _, err := storage.LoadLastLogIndex(tx)
		if err == nil {
			storage.block = block
		}
		return err
	})
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
	if k != nil {
		err := json.Unmarshal(v, &record)
		if err != nil {
			return record, err
		}
		return record, nil
	} else {
		return record, errors.New("Database is empty")
	}
}

func (self *BoltLogStorage) LoadLastLogIndex(tx *bolt.Tx) (uint64, uint, error) {
	b := tx.Bucket([]byte(TRADELOG_BUCKET))
	c := b.Cursor()
	k, v := c.Last()
	if k != nil {
		record := common.TradeLog{}
		json.Unmarshal(v, &record)
		return record.BlockNumber, record.Index, nil
	} else {
		return 0, 0, errors.New("Database is empty")
	}
}

func (self *BoltLogStorage) StoreCatLog(l common.SetCatLog) error {
	var err error
	self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(CATLOG_BUCKET))
		var dataJson []byte
		dataJson, err = json.Marshal(l)
		if err != nil {
			return err
		}
		// log.Printf("Storing cat log: %d", l.Timestamp)
		idByte := uint64ToBytes(l.Timestamp)
		err = b.Put(idByte, dataJson)
		return err
	})
	return err
}

func (self *BoltLogStorage) StoreTradeLog(stat common.TradeLog, timepoint uint64) error {
	var err error
	self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TRADELOG_BUCKET))
		var dataJson []byte
		block, index, berr := self.LoadLastLogIndex(tx)
		if berr == nil && (block > stat.BlockNumber || (block == stat.BlockNumber && index >= stat.Index)) {
			err = errors.New(
				fmt.Sprintf("Duplicated log %+v (new block number %d is smaller or equal to latest block number %d and tx index %d is smaller or equal to last log tx index %d)", stat, block, stat.BlockNumber, index, stat.Index))
			return err
		}
		dataJson, err = json.Marshal(stat)
		if err != nil {
			return err
		}
		// log.Printf("Storing log: %d", stat.Timestamp)
		idByte := uint64ToBytes(stat.Timestamp)
		err = b.Put(idByte, dataJson)
		return err
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
	var result common.SetCatLog
	var err error
	self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(CATLOG_BUCKET))
		c := b.Cursor()
		k, v := c.Last()
		if k == nil {
			err = errors.New("there is no catlog")
		} else {
			err = json.Unmarshal(v, &result)
		}
		return err
	})
	return result, err
}

func (self *BoltLogStorage) GetFirstCatLog() (common.SetCatLog, error) {
	var result common.SetCatLog
	var err error
	self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(CATLOG_BUCKET))
		c := b.Cursor()
		k, v := c.First()
		if k == nil {
			err = errors.New("there is no catlog")
		} else {
			err = json.Unmarshal(v, &result)
		}
		return err
	})
	return result, err
}

func (self *BoltLogStorage) GetCatLogs(fromTime uint64, toTime uint64) ([]common.SetCatLog, error) {
	result := []common.SetCatLog{}
	var err error
	if toTime-fromTime > MAX_GET_LOG_PERIOD {
		return result, errors.New(fmt.Sprintf("Time range is too broad, it must be smaller or equal to %d miliseconds", MAX_GET_RATES_PERIOD))
	}
	self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(CATLOG_BUCKET))
		c := b.Cursor()
		min := uint64ToBytes(fromTime)
		max := uint64ToBytes(toTime)
		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			record := common.SetCatLog{}
			err = json.Unmarshal(v, &record)
			if err != nil {
				return err
			}
			result = append([]common.SetCatLog{record}, result...)
		}
		return nil
	})
	return result, err
}

func (self *BoltLogStorage) GetLastTradeLog() (common.TradeLog, error) {
	var result common.TradeLog
	var err error
	self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TRADELOG_BUCKET))
		c := b.Cursor()
		k, v := c.Last()
		if k == nil {
			err = errors.New("there is no tradelog")
		} else {
			err = json.Unmarshal(v, &result)
		}
		return err
	})
	return result, err
}

func (self *BoltLogStorage) GetFirstTradeLog() (common.TradeLog, error) {
	var result common.TradeLog
	var err error
	self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TRADELOG_BUCKET))
		c := b.Cursor()
		k, v := c.First()
		if k == nil {
			err = errors.New("there is no tradelog")
		} else {
			err = json.Unmarshal(v, &result)
		}
		return err
	})
	return result, err
}

func (self *BoltLogStorage) GetTradeLogs(fromTime uint64, toTime uint64) ([]common.TradeLog, error) {
	result := []common.TradeLog{}
	var err error
	if toTime-fromTime > MAX_GET_LOG_PERIOD {
		return result, errors.New(fmt.Sprintf("Time range is too broad, it must be smaller or equal to %d miliseconds", MAX_GET_RATES_PERIOD))
	}
	self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TRADELOG_BUCKET))
		c := b.Cursor()
		min := uint64ToBytes(fromTime)
		max := uint64ToBytes(toTime)
		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			record := common.TradeLog{}
			err = json.Unmarshal(v, &record)
			if err != nil {
				return err
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
