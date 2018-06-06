package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/KyberNetwork/reserve-data/boltutil"
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/boltdb/bolt"
	"github.com/jinzhu/now"
)

const (
	TRANSACTION_INFO_BUCKET  string = "transaction"
	INDEXED_TIMESTAMP_BUCKET string = "indexed_timestamp"
	TOTAL_GAS_SPENT_BUCKET   string = "total_gas_spent"

	ETH_TO_WEI                 float64 = 1000000000000000000
	DAY                        uint64  = 86400   // a day in seconds
	MAX_FEE_SETRATE_TIME_RANGE uint64  = 7776000 // 3 months in seconds
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
		_, err = tx.CreateBucketIfNotExists([]byte(TOTAL_GAS_SPENT_BUCKET))
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
			keyUint := boltutil.BytesToUint64(k)
			latestBlockChecked = keyUint / 1000000
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return latestBlockChecked, nil
}

func (self *BoltFeeSetRateStorage) StoreTransaction(txs []common.SetRateTxInfo) error {
	var err error
	err = self.db.Update(func(tx *bolt.Tx) error {
		var dataJSON []byte
		b := tx.Bucket([]byte(TRANSACTION_INFO_BUCKET))
		bIndex := tx.Bucket([]byte(INDEXED_TIMESTAMP_BUCKET))
		bTotal := tx.Bucket([]byte(TOTAL_GAS_SPENT_BUCKET))

		for _, transaction := range txs {
			blockNumUint, vErr := strconv.ParseUint(transaction.BlockNumber, 10, 64)
			if vErr != nil {
				return vErr
			}
			txIndexUint, vErr := strconv.ParseUint(transaction.TransactionIndex, 10, 64)
			if vErr != nil {
				return vErr
			}
			keyStoreUint := blockNumUint*1000000 + txIndexUint
			keyStore := boltutil.Uint64ToBytes(keyStoreUint)
			storeTx, vErr := common.GetStoreTx(transaction)
			if vErr != nil {
				return vErr
			}
			vErr = bIndex.Put(boltutil.Uint64ToBytes(storeTx.TimeStamp), keyStore)
			if vErr != nil {
				return vErr
			}
			vErr = storeTotalGasSpent(bTotal, storeTx)
			if vErr != nil {
				return vErr
			}
			dataJSON, vErr = json.Marshal(storeTx)
			if vErr != nil {
				return vErr
			}
			vErr = b.Put(keyStore, dataJSON)
			if vErr != nil {
				return vErr
			}
		}
		return nil
	})
	return err
}

func storeTotalGasSpent(b *bolt.Bucket, storeTx common.StoreSetRateTx) error {
	var err error
	totalGasSpent := big.NewInt(0)
	keyUint := uint64(now.New(time.Unix(int64(storeTx.TimeStamp), 0).UTC()).BeginningOfDay().Unix())
	keyStore := boltutil.Uint64ToBytes(keyUint)
	gasCost := big.NewInt(int64(storeTx.GasPrice * storeTx.GasUsed))
	totalGasSpentByte := b.Get(keyStore)
	if totalGasSpentByte == nil {
		c := b.Cursor()
		_, last := c.Last()
		if last != nil {
			totalGasSpent.SetBytes(last)
			totalGasSpent.Add(totalGasSpent, gasCost)
			err = b.Put(keyStore, totalGasSpent.Bytes())
			return err
		}
		totalGasSpent = gasCost
		err = b.Put(keyStore, totalGasSpent.Bytes())
		return err
	}
	totalGasSpent.SetBytes(totalGasSpentByte)
	totalGasSpent.Add(totalGasSpent, gasCost)
	err = b.Put(keyStore, totalGasSpent.Bytes())
	return err
}

func (self *BoltFeeSetRateStorage) GetFeeSetRateByDay(fromTime, toTime uint64) ([]common.FeeSetRate, error) {
	var seqFeeSetRate []common.FeeSetRate
	fromTimeSecond := fromTime / 1000
	toTimeSecond := toTime / 1000
	if toTimeSecond > (MAX_FEE_SETRATE_TIME_RANGE + fromTimeSecond) {
		return seqFeeSetRate, fmt.Errorf("Time range is too broad, it must be smaller or equal to three months (%d seconds)", MAX_FEE_SETRATE_TIME_RANGE)
	}

	var err error
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TRANSACTION_INFO_BUCKET))
		bIndex := tx.Bucket([]byte(INDEXED_TIMESTAMP_BUCKET))
		c := b.Cursor()
		cIndex := bIndex.Cursor()
		bTotal := tx.Bucket([]byte(TOTAL_GAS_SPENT_BUCKET))
		minUint := uint64(now.New(time.Unix(int64(fromTimeSecond), 0).UTC()).BeginningOfDay().Unix())
		maxUint := uint64(now.New(time.Unix(int64(toTimeSecond), 0).UTC()).BeginningOfDay().Unix())
		var tickTime []byte = boltutil.Uint64ToBytes(minUint)
		var nextTick []byte = boltutil.Uint64ToBytes(minUint + DAY)
		max := boltutil.Uint64ToBytes(maxUint)

		for {
			if bytes.Compare(nextTick, max) > 0 {
				break
			}
			_, tickBlock := cIndex.Seek(tickTime)
			_, nextTickBlock := cIndex.Seek(nextTick)
			if tickBlock != nil && nextTickBlock != nil {
				totalGasSpent := bTotal.Get(tickTime)
				feeSetRate, vErr := getFeeSetRate(c, tickBlock, nextTickBlock, tickTime, totalGasSpent)
				if vErr != nil {
					return vErr
				}
				// if timestamp = 0 means that there are no setrate activities on this day
				if feeSetRate.TimeStamp != 0 {
					seqFeeSetRate = append(seqFeeSetRate, feeSetRate)
				}
			} else {
				break
			}
			tickTime = nextTick
			nextTick = boltutil.Uint64ToBytes(boltutil.BytesToUint64(nextTick) + DAY)
		}
		return nil
	})
	return seqFeeSetRate, err
}

func getFeeSetRate(c *bolt.Cursor, tickBlock, nextTickBlock, tickTime, totalGasSpentByte []byte) (common.FeeSetRate, error) {
	var feeSetRate common.FeeSetRate
	totalGasSpentInt := big.NewInt(0)
	if totalGasSpentByte == nil {
		return feeSetRate, nil
	}
	totalGasSpent := big.NewFloat(0)
	totalGasSpentInt.SetBytes(totalGasSpentByte)
	totalGasSpent.SetInt(totalGasSpentInt)
	totalGasSpent.Quo(totalGasSpent, big.NewFloat(ETH_TO_WEI))
	sumFee := big.NewFloat(0)
	gasInEther := big.NewFloat(0)

	for k, v := c.Seek(tickBlock); k != nil && bytes.Compare(k, nextTickBlock) < 0; k, v = c.Next() {
		record := common.StoreSetRateTx{}
		if err := json.Unmarshal(v, &record); err != nil {
			return feeSetRate, err
		}
		gasInWei := big.NewFloat(float64(record.GasPrice * record.GasUsed))
		gasInEther.Quo(gasInWei, big.NewFloat(ETH_TO_WEI))
		sumFee.Add(sumFee, gasInEther)
	}

	feeSetRate = common.FeeSetRate{
		TimeStamp:     boltutil.BytesToUint64(tickTime),
		GasUsed:       sumFee,
		TotalGasSpent: totalGasSpent,
	}
	return feeSetRate, nil
}
