package storage

import (
	"bytes"
	"log"
	"strconv"
	"encoding/json"
	"fmt"
	"time"

	"github.com/boltdb/bolt"
	"github.com/jinzhu/now"
	"github.com/KyberNetwork/reserve-data/common"
)

const (
	MAX_TIME_DISTANCE uint64 = 86400
	TRANSACTION_INFO string = "transaction"
	GWEI uint64 = 1000000000
	DAY uint64 = 86400 // a day in seconds
	MAX_FEE_SETRATE_TIME_RAGE uint64 = 7776000 // 3 months in seconds
)

type BoltFeeSetRateStorage struct {
	db *bolt.DB
}

func NewBoltFeeSetRateStorage(path string) (*BoltFeeSetRateStorage, error) {
	var err error
	var db *bolt.DB
	db, err = bolt.Open(path, 0600, nil)
	if err != nil {
		panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists([]byte(TRANSACTION_INFO))
		if err != nil {
			return err
		}
		return nil
	})
	storage := &BoltFeeSetRateStorage{db}
	return storage, err
}

func (self *BoltFeeSetRateStorage) GetLastBlockChecked() (uint64, error) {
	var latestBlockChecked uint64
	var err error
	err = self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TRANSACTION_INFO))
		c := b.Cursor()
		_, v := c.Last()

		if v != nil {
			var txs common.StoreTransaction
			err = json.Unmarshal(v, &txs)
			if err != nil {
				log.Println(err)
				return err
			}
			log.Println("last block checked: ", txs.BlockNumber)
			latestBlockChecked, err = strconv.ParseUint(txs.BlockNumber, 10, 64)
			if err != nil {
				log.Println(err)
				return err
			}
		}
		return nil
	})
	return latestBlockChecked, err
}

func (self *BoltFeeSetRateStorage) StoreTransaction(txs common.TransactionInfo) error {
	// log.Println("data save: ", txs.BlockNumber)
	var err error
	err = self.db.Update(func(tx *bolt.Tx) error {
		var dataJson []byte
		b := tx.Bucket([]byte(TRANSACTION_INFO))

		storeTxs, err := common.GetStoreTx(txs)
		if err != nil {
			log.Println(err)
			return err
		}
		dataJson, err = json.Marshal(storeTxs)
		if err != nil {
			log.Println(err)
			return err
		}
		timeStampStr := txs.TimeStamp
		timeStamp, err := strconv.ParseUint(timeStampStr, 10, 64)
		if err != nil {
			log.Println(err)
			return err
		}
		return b.Put(uint64ToBytes(timeStamp), dataJson)
	})
	return err
}

func (self *BoltFeeSetRateStorage) GetFeeSetRateByDay(fromTime, toTime uint64) ([]common.FeeSetRate, error) {
	fromTimeSecond := fromTime / 1000
	toTimeSecond := toTime / 1000
	if toTimeSecond - fromTimeSecond > MAX_FEE_SETRATE_TIME_RAGE {
		return []common.FeeSetRate{}, fmt.Errorf("Time range is too broad, it must be smaller or equal to three months (%d seconds)", MAX_FEE_SETRATE_TIME_RAGE)
	}

	seqFeeSetRate := []common.FeeSetRate{}
	var err error
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TRANSACTION_INFO))
		c := b.Cursor()
		minUint := uint64(now.New(time.Unix(int64(fromTimeSecond), 0).UTC()).BeginningOfDay().Unix())
		maxUint := uint64(now.New(time.Unix(int64(toTimeSecond), 0).UTC()).BeginningOfDay().Unix())
		var tickTime []byte = uint64ToBytes(minUint)
		var nextTick []byte = uint64ToBytes(minUint + DAY)
		var sumFee float64 = 0
		var feeSetRate common.FeeSetRate
		min := uint64ToBytes(minUint)
		max := uint64ToBytes(maxUint)

		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			record := common.StoreTransaction{}
			if vErr := json.Unmarshal(v, &record); vErr != nil {
				return vErr
			}
			if bytes.Compare(k, nextTick) < 0 {
				gasInGWei := float64(record.GasPrice * record.GasUsed) / float64(GWEI)
				sumFee += gasInGWei
				continue
			}
			feeSetRate = common.FeeSetRate{
				TimeStamp: bytesToUint64(tickTime),
				GasUsed:   sumFee,
			}
			seqFeeSetRate = append(seqFeeSetRate, feeSetRate)
			sumFee = 0
			tickTime = nextTick
			nextTick = uint64ToBytes(bytesToUint64(nextTick) + DAY)
		}
		feeSetRate = common.FeeSetRate{
			TimeStamp: bytesToUint64(tickTime),
			GasUsed:   sumFee,
		}
		seqFeeSetRate = append(seqFeeSetRate, feeSetRate)
		return nil
	})
	return seqFeeSetRate, err
}
