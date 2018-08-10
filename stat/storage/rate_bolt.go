package storage

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"github.com/KyberNetwork/reserve-data/boltutil"
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/boltdb/bolt"
	ethereum "github.com/ethereum/go-ethereum/common"
)

const (
	maxGetRatesPeriod uint64 = 86400000      //1 days in milisec
	rateExpired       uint64 = 30 * 86400000 //30 days in milisecond
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

func getEndOfDayTimestamp(timestamp uint64) uint64 {
	ui64Day := uint64(time.Hour*24) / 1000000
	log.Printf("StatPruner: %d", ui64Day)
	return (timestamp/ui64Day)*ui64Day + ui64Day
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

type ReserverRateRecord struct {
	ReserveAddress string
	Rate           common.ReserveRates
}

func getFirstRecordTimestamp(tx *bolt.Tx) (uint64, error) {
	var result uint64 = math.MaxUint64
	err := tx.ForEach(func(name []byte, b *bolt.Bucket) error {
		c := b.Cursor()
		k, _ := c.First()
		recordTimestamp := boltutil.BytesToUint64(k)
		if recordTimestamp != 0 && recordTimestamp < result {
			result = recordTimestamp
		}
		return nil
	})
	return result, err
}

func (self *BoltRateStorage) ExportExpiredRateData(currentTime uint64, fileName string) (fromTime uint64, toTime uint64, nRecord uint64, err error) {
	expiredTimestampByte := boltutil.Uint64ToBytes(currentTime - rateExpired)
	outFile, err := os.Create(fileName)
	if err != nil {
		return 0, 0, 0, err
	}
	zw := gzip.NewWriter(outFile)
	zw.Name = fileName
	defer func() {
		if cErr := zw.Close(); cErr != nil {
			log.Printf("StatPruner: closing gzip error %s", cErr.Error())
		}
		if cErr := outFile.Close(); cErr != nil {
			log.Printf("Expire file close error: %s", cErr.Error())
		}
	}()

	err = self.db.View(func(tx *bolt.Tx) error {
		var vErr error
		fromTime, vErr = getFirstRecordTimestamp(tx)
		if vErr != nil {
			return vErr
		}
		if fromTime == 0 {
			log.Printf("There are no first time record. return now")
			return nil
		}
		toTimeByte := boltutil.Uint64ToBytes(getEndOfDayTimestamp(fromTime))
		log.Printf("StatPruner:fromtime is %d, to time is %d", fromTime, boltutil.BytesToUint64(toTimeByte))
		//loop through each bucket
		bucketLoopErr := tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			addrStr := string(name)
			c := b.Cursor()
			for k, v := c.First(); k != nil && bytes.Compare(k, expiredTimestampByte) <= 0 && bytes.Compare(k, toTimeByte) == -1; k, v = c.Next() {
				rates := common.ReserveRates{}
				if bVErr := json.Unmarshal(v, &rates); bVErr != nil {
					return bVErr
				}
				record := common.NewExportedReserverRateRecord(ethereum.HexToAddress(addrStr), rates, boltutil.BytesToUint64(k))
				var output []byte
				output, bVErr := json.Marshal(record)
				if bVErr != nil {
					return bVErr
				}
				_, bVErr = zw.Write([]byte((string(output) + "\n")))
				if bVErr != nil {
					return bVErr
				}
				nRecord++
				if boltutil.BytesToUint64(k) > toTime {
					toTime = boltutil.BytesToUint64(k)
				}
			}
			return nil
		})
		return bucketLoopErr
	})

	return fromTime, toTime, nRecord, err
}

func (self *BoltRateStorage) PruneExpiredReserveRateData(toTime uint64) (nRecord uint64, err error) {
	toTimeByte := boltutil.Uint64ToBytes(toTime)
	err = self.db.Update(func(tx *bolt.Tx) error {
		bucketLoopErr := tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			c := b.Cursor()
			for k, _ := c.First(); k != nil && bytes.Compare(k, toTimeByte) <= 0; k, _ = c.Next() {
				if uErr := b.Delete(k); uErr != nil {
					return uErr
				}
				nRecord++
			}
			return nil
		})
		return bucketLoopErr
	})
	return nRecord, err
}
