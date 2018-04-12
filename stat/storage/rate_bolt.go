package storage

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/boltdb/bolt"
)

const (
	MAX_GET_RATES_PERIOD uint64 = 86400000 //1 days in milisec
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
	db.Update(func(tx *bolt.Tx) error {
		tx.CreateBucketIfNotExists([]byte(RESERVE_RATES))
		return nil
	})
	storage := &BoltRateStorage{db}
	return storage, nil
}

func (self *BoltRateStorage) StoreReserveRates(reserveAddr string, rate common.ReserveRates, timepoint uint64) error {
	var err error
	err = self.db.Update(func(tx *bolt.Tx) error {
		b, _ := tx.CreateBucketIfNotExists([]byte(reserveAddr))
		c := b.Cursor()
		var prevDataJSON common.ReserveRates
		_, prevData := c.Last()
		json.Unmarshal(prevData, &prevDataJSON)
		if prevDataJSON.BlockNumber < rate.BlockNumber {
			idByte := uint64ToBytes(timepoint)
			dataJson, err := json.Marshal(rate)
			if err != nil {
				return err
			}
			err = b.Put(idByte, dataJson)
			if err != nil {
				log.Printf("Saving rates to db failed: err(%+v)", err)
				return err
			}
			log.Printf("Save rates to db %s successfully", reserveAddr)
		}
		return nil
	})
	return err
}

func (self *BoltRateStorage) GetReserveRates(fromTime, toTime uint64, reserveAddr string) ([]common.ReserveRates, error) {
	var err error
	var result []common.ReserveRates
	if toTime-fromTime > MAX_GET_RATES_PERIOD {
		return result, errors.New(fmt.Sprintf("Time range is too broad, it must be smaller or equal to %d miliseconds", MAX_GET_RATES_PERIOD))
	}
	err = self.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(reserveAddr))
		if err != nil {
			log.Println("Cannot get bucket: ", err.Error())
			return err
		}
		c := b.Cursor()
		min := uint64ToBytes(fromTime)
		max := uint64ToBytes(toTime)
		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			rates := common.ReserveRates{}
			err := json.Unmarshal(v, &rates)
			if err != nil {
				return err
			}
			result = append(result, rates)
		}
		return err
	})
	return result, err
}
