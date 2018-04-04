package storage

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/boltdb/bolt"
)

const (
	PRICE_ANALYTIC_BUCKET   string = "price_analytic"
	MAX_GET_ANALYTIC_PERIOD uint64 = 86400000 //1 sec in milisecond
)

type BoltAnalyticStorage struct {
	db *bolt.DB
}

func NewBoltAnalyticStorage(path string) (*BoltAnalyticStorage, error) {
	var (
		err error
		db  *bolt.DB
	)

	db, err = bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, uErr := tx.CreateBucket([]byte(PRICE_ANALYTIC_BUCKET))
		return uErr
	})
	if err != nil {
		return nil, err
	}
	storage := &BoltAnalyticStorage{db}
	return storage, nil
}

func (self *BoltAnalyticStorage) UpdatePriceAnalyticData(timestamp uint64, value []byte) error {
	var err error
	k := uint64ToBytes(timestamp)
	err = self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(PRICE_ANALYTIC_BUCKET))
		c := b.Cursor()
		existedKey, _ := c.Seek(k)
		if existedKey != nil {
			return errors.New("The timestamp is already existed.")
		}
		return b.Put(k, value)
	})
	return err
}

func (self *BoltAnalyticStorage) GetPriceAnalyticData(fromTime uint64, toTime uint64) ([]common.AnalyticPriceResponse, error) {
	var err error
	min := uint64ToBytes(fromTime)
	max := uint64ToBytes(toTime)
	var result []common.AnalyticPriceResponse
	if toTime-fromTime > MAX_GET_ANALYTIC_PERIOD {
		return result, errors.New(fmt.Sprintf("Time range is too broad, it must be smaller or equal to %d miliseconds", MAX_GET_RATES_PERIOD))
	}

	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(PRICE_ANALYTIC_BUCKET))
		c := b.Cursor()
		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			timestamp := bytesToUint64(k)
			temp := make(map[string]interface{})
			vErr := json.Unmarshal(v, &temp)
			if vErr != nil {
				return vErr
			}
			record := common.AnalyticPriceResponse{
				Timestamp: timestamp,
				Data:      temp,
			}
			result = append(result, record)
		}
		return nil
	})
	return result, err
}
