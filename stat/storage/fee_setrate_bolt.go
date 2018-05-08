package storage

import (
	"bytes"
	"log"
	"strconv"
	"encoding/json"
	// "fmt"
	"time"
	"math/big"

	"github.com/boltdb/bolt"
	"github.com/jinzhu/now"
	"github.com/KyberNetwork/reserve-data/common"
)

const (
	TRANSACTION_INFO_BUCKET string = "transaction"
	INDEXED_TIMESTAMP_BUCKET string = "indexed_timestamp"

	MAX_TIME_DISTANCE uint64 = 86400
	ETH_TO_WEI float64 = 1000000000000000000
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
		_, err = tx.CreateBucketIfNotExists([]byte(TRANSACTION_INFO_BUCKET))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(INDEXED_TIMESTAMP_BUCKET))
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
		b := tx.Bucket([]byte(TRANSACTION_INFO_BUCKET))
		c := b.Cursor()
		k, _ := c.Last()

		if k != nil {
			keyUint := bytesToUint64(k)
			latestBlockChecked = keyUint / 1000
		}
		return nil
	})
	if err != nil {
		log.Println(err)
		return 0, err
	}
	log.Println("lastBlockChecked: ", latestBlockChecked)
	return latestBlockChecked, nil
}

func (self *BoltFeeSetRateStorage) StoreTransaction(txs []common.SetRateTxInfo) error {
	var err error
	err = self.db.Update(func(tx *bolt.Tx) error {
		var dataJson []byte
		b := tx.Bucket([]byte(TRANSACTION_INFO_BUCKET))
		bIndex := tx.Bucket([]byte(INDEXED_TIMESTAMP_BUCKET))
		for _, transaction := range txs {
			blockNumUint, err := strconv.ParseUint(transaction.BlockNumber, 10, 64)
			if err != nil {
				log.Printf("Cant convert %s to uint64", transaction.BlockNumber)
				return err
			}
			txIndexUint, err := strconv.ParseUint(transaction.TransactionIndex, 10, 64)
			if err != nil {
				log.Printf("Cant convert %s to uint64", transaction.TransactionIndex)
				return err
			}
			keyStoreUint := blockNumUint * 1000 + txIndexUint
			keyStore := uint64ToBytes(keyStoreUint)
			storeTx, err := common.GetStoreTx(transaction)
			if err != nil {
				log.Println(err)
				return err
			}
			err = bIndex.Put(uint64ToBytes(storeTx.TimeStamp), keyStore)
			if err != nil {
				log.Println(err)
				return err
			}
			dataJson, err = json.Marshal(storeTx)
			if err != nil {
				log.Println(err)
				return err
			}
			err = b.Put(keyStore, dataJson)
			if err != nil {
				log.Println(err)
				return err
			}
		}
		return nil
	})
	return err
}

func (self *BoltFeeSetRateStorage) GetFeeSetRateByDay(fromTime, toTime uint64) ([]common.FeeSetRate, error) {
	fromTimeSecond := fromTime / 1000
	toTimeSecond := toTime / 1000
	// if toTimeSecond - fromTimeSecond > MAX_FEE_SETRATE_TIME_RAGE {
	// 	return []common.FeeSetRate{}, fmt.Errorf("Time range is too broad, it must be smaller or equal to three months (%d seconds)", MAX_FEE_SETRATE_TIME_RAGE)
	// }

	seqFeeSetRate := []common.FeeSetRate{}
	var err error
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TRANSACTION_INFO_BUCKET))
		bIndex := tx.Bucket([]byte(INDEXED_TIMESTAMP_BUCKET))
		c := b.Cursor()
		cIndex := bIndex.Cursor()
		minUint := uint64(now.New(time.Unix(int64(fromTimeSecond), 0).UTC()).BeginningOfDay().Unix())
		maxUint := uint64(now.New(time.Unix(int64(toTimeSecond), 0).UTC()).BeginningOfDay().Unix())
		var tickTime []byte = uint64ToBytes(minUint)
		var nextTick []byte = uint64ToBytes(minUint + DAY)
		max := uint64ToBytes(maxUint)

		for {
			if bytes.Compare(nextTick, max) > 0 {
				break
			}
			_, tickBlock := cIndex.Seek(tickTime)
			_, nextTickBlock := cIndex.Seek(nextTick)
			if tickBlock != nil && nextTickBlock != nil {
				feeSetRate, err := getFeeSetRate(c, tickBlock, nextTickBlock, tickTime)
				if err != nil {
					return err
				}
				seqFeeSetRate = append(seqFeeSetRate, feeSetRate)
			}
			tickTime = nextTick
			nextTick = uint64ToBytes(bytesToUint64(nextTick) + DAY)
		}
		return nil
	})
	return seqFeeSetRate, err
}

func getFeeSetRate(c *bolt.Cursor, tickBlock, nextTickBlock, tickTime []byte) (common.FeeSetRate, error) {
	sumFee := big.NewFloat(0)
	gasInEther := big.NewFloat(0)
	var feeSetRate common.FeeSetRate

	for k, v := c.Seek(tickBlock); k != nil && bytes.Compare(k, nextTickBlock) < 0; k, v = c.Next() {
		record := common.StoreSetRateTx{}
		if err := json.Unmarshal(v, &record); err != nil {
			return feeSetRate, err
		}
		log.Println("record: ", record)
		gasInWei := big.NewFloat(float64(record.GasPrice * record.GasUsed))
		gasInEther.Quo(gasInWei, big.NewFloat(ETH_TO_WEI))
		sumFee.Add(sumFee, gasInEther)
	}

	feeSetRate = common.FeeSetRate{
		TimeStamp: bytesToUint64(tickTime),
		GasUsed:   sumFee,
	}
	return feeSetRate, nil
}
