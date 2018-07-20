package storage

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/KyberNetwork/reserve-data/boltutil"
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/boltdb/bolt"
	ethereum "github.com/ethereum/go-ethereum/common"
)

const (
	maxGetRatesPeriod uint64 = 86400000 //1 days in milisec
)

type BoltRateStorage struct {
	db *bolt.DB
}

func NewBoltRateStorage(path string) (*BoltRateStorage, error) {
	// init instance
	var err error
	var db *bolt.DB
	db, err = bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}
	// init buckets
	err = db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists([]byte(reserveRates))
		return err
	})
	storage := &BoltRateStorage{db}
	return storage, nil
}

func (self *BoltRateStorage) StoreReserveRates(ethReserveAddr ethereum.Address, rate common.ReserveRates, timepoint uint64) error {
	var err error
	reserveAddr := common.AddrToString(ethReserveAddr)
	err = self.db.Update(func(tx *bolt.Tx) error {
		b, _ := tx.CreateBucketIfNotExists([]byte(reserveAddr))
		c := b.Cursor()
		var prevDataJSON common.ReserveRates
		_, prevData := c.Last()
		if prevData != nil {
			if uErr := json.Unmarshal(prevData, &prevDataJSON); uErr != nil {
				return uErr
			}
		}
		if prevDataJSON.BlockNumber < rate.BlockNumber {
			idByte := boltutil.Uint64ToBytes(timepoint)
			dataJSON, uErr := json.Marshal(rate)
			if uErr != nil {
				return uErr
			}
			return b.Put(idByte, dataJSON)
		}
		return nil
	})
	return err
}

func (self *BoltRateStorage) GetReserveRates(fromTime, toTime uint64, ethReserveAddr ethereum.Address) ([]common.ReserveRates, error) {
	var err error
	reserveAddr := common.AddrToString(ethReserveAddr)
	var result []common.ReserveRates
	if toTime-fromTime > maxGetRatesPeriod {
		return result, fmt.Errorf("Time range is too broad, it must be smaller or equal to %d miliseconds", maxGetRatesPeriod)
	}
	err = self.db.Update(func(tx *bolt.Tx) error {
		b, uErr := tx.CreateBucketIfNotExists([]byte(reserveAddr))
		if uErr != nil {
			return uErr
		}
		c := b.Cursor()
		min := boltutil.Uint64ToBytes(fromTime)
		max := boltutil.Uint64ToBytes(toTime)
		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			rates := common.ReserveRates{}
			if uErr = json.Unmarshal(v, &rates); uErr != nil {
				return uErr
			}
			result = append(result, rates)
		}
		return nil
	})
	return result, err
}
