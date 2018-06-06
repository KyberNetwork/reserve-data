package storage

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/KyberNetwork/reserve-data/boltutil"
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

//NewBoltAnalyticStorage return new storage instance
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
	k := boltutil.Uint64ToBytes(timestamp)
	err = self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(PRICE_ANALYTIC_BUCKET))
		c := b.Cursor()
		existedKey, _ := c.Seek(k)
		if existedKey != nil {
			return errors.New("timestamp is already existed")
		}
		return b.Put(k, value)
	})
	return err
}

func (self *BoltAnalyticStorage) ExportExpiredPriceAnalyticData(currentTime uint64, fileName string) (nRecord uint64, err error) {
	expiredTimestampByte := boltutil.Uint64ToBytes(currentTime - PRICE_ANALYTIC_EXPIRED)
	outFile, err := os.Create(fileName)
	defer func() {
		if vErr := outFile.Close(); vErr != nil {
			log.Printf("Expire file close error: %s", vErr.Error())
		}
	}()
	if err != nil {
		return 0, err
	}
	err = self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(PRICE_ANALYTIC_BUCKET))
		c := b.Cursor()
		for k, v := c.First(); k != nil && bytes.Compare(k, expiredTimestampByte) <= 0; k, v = c.Next() {
			timestamp := boltutil.BytesToUint64(k)
			temp := make(map[string]interface{})
			if err = json.Unmarshal(v, &temp); err != nil {
				return err
			}
			record := common.NewAnalyticPriceResponse(
				timestamp,
				temp,
			)
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
	expiredTimestampByte := boltutil.Uint64ToBytes(currentTime - PRICE_ANALYTIC_EXPIRED)
	err = self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(PRICE_ANALYTIC_BUCKET))
		c := b.Cursor()
		for k, _ := c.First(); k != nil && bytes.Compare(k, expiredTimestampByte) <= 0; k, _ = c.Next() {
			if err = b.Delete(k); err != nil {
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
	min := boltutil.Uint64ToBytes(fromTime)
	max := boltutil.Uint64ToBytes(toTime)
	var result []common.AnalyticPriceResponse
	if toTime-fromTime > MAX_GET_ANALYTIC_PERIOD {
		return result, fmt.Errorf("Time range is too broad, it must be smaller or equal to %d miliseconds", MAX_GET_RATES_PERIOD)
	}

	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(PRICE_ANALYTIC_BUCKET))
		c := b.Cursor()
		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			timestamp := boltutil.BytesToUint64(k)
			temp := make(map[string]interface{})
			if vErr := json.Unmarshal(v, &temp); vErr != nil {
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
