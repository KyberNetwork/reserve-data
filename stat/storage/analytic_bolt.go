package storage

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/KyberNetwork/reserve-data/common"

	"github.com/boltdb/bolt"
)

const (
	PRICE_ANALYTIC_BUCKET   string = "price_analytic"
	MAX_GET_ANALYTIC_PERIOD uint64 = 86400000      //1 day in milisecond
	PRICE_ANALYTIC_EXPIRED  uint64 = 30 * 86400000 //30 days in milisecond
)

type BoltAnalyticStorage struct {
	db *bolt.DB
}

func NewBoltAnalyticStorage(dbPath string) (*BoltAnalyticStorage, error) {
	var err error
	var db *bolt.DB
	db, err = bolt.Open(dbPath, 0600, nil)
	if err != nil {
		panic(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists([]byte(PRICE_ANALYTIC_BUCKET))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	storage := BoltAnalyticStorage{db}
	return &storage, nil
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

func (self *BoltAnalyticStorage) ExportExpiredPriceAnalyticData(currentTime uint64, fileName string) (nRecord uint64, err error) {
	expiredTimestampByte := uint64ToBytes(currentTime - PRICE_ANALYTIC_EXPIRED)
	outFile, err := os.Create(fileName)
	defer outFile.Close()
	if err != nil {
		return 0, err
	}
	err = self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(PRICE_ANALYTIC_BUCKET))
		c := b.Cursor()
		for k, v := c.First(); k != nil && bytes.Compare(k, expiredTimestampByte) <= 0; k, v = c.Next() {
			timestamp := bytesToUint64(k)
			temp := make(map[string]interface{})
			err = json.Unmarshal(v, &temp)
			if err != nil {
				return err
			}
			record := common.AnalyticPriceResponse{
				timestamp,
				temp,
			}
			var output []byte
			output, err = json.Marshal(record)
			if err != nil {
				return err
			}
			_, err = outFile.WriteString(string(output) + "\n")
			if err != nil {
				return err
			}
			nRecord++
		}
		return nil
	})
	return nRecord, err
}

func (self *BoltAnalyticStorage) PruneExpiredPriceAnalyticData(currentTime uint64) (nRecord uint64, err error) {
	expiredTimestampByte := uint64ToBytes(currentTime - PRICE_ANALYTIC_EXPIRED)
	err = self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(PRICE_ANALYTIC_BUCKET))
		c := b.Cursor()
		for k, _ := c.First(); k != nil && bytes.Compare(k, expiredTimestampByte) <= 0; k, _ = c.Next() {
			err = b.Delete(k)
			if err != nil {
				return err
			}
			nRecord++
		}
		return nil
	})
	return nRecord, err
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
