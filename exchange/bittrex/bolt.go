package bittrex

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/KyberNetwork/reserve-data/boltutil"
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/boltdb/bolt"
)

const (
	bittrexDepositHistory string = "bittrex_deposit_history"
	tradeHistory          string = "trade_history"
	maxGetTradeHistory    uint64 = 3 * 86400000
)

//BoltStorage storage object for bittrex
type BoltStorage struct {
	db *bolt.DB
}

//NewBoltStorage return new storage instance for using bittrex
func NewBoltStorage(path string) (*BoltStorage, error) {
	// init instance
	var err error
	var db *bolt.DB
	db, err = bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}
	// init buckets
	err = db.Update(func(tx *bolt.Tx) error {
		if _, uErr := tx.CreateBucketIfNotExists([]byte(bittrexDepositHistory)); uErr != nil {
			return uErr
		}
		if _, uErr := tx.CreateBucketIfNotExists([]byte(tradeHistory)); uErr != nil {
			return uErr
		}
		return nil
	})
	storage := &BoltStorage{db}
	return storage, err
}

func (self *BoltStorage) IsNewBittrexDeposit(id uint64, actID common.ActivityID) bool {
	res := true
	err := self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bittrexDepositHistory))
		v := b.Get(boltutil.Uint64ToBytes(id))
		if v != nil && string(v) != actID.String() {
			log.Printf("bolt: stored act id - current act id: %s - %s", string(v), actID.String())
			res = false
		}
		return nil
	})
	if err != nil {
		log.Printf("Check new bittrex deposit error: %s", err.Error())
	}
	return res
}

func (self *BoltStorage) RegisterBittrexDeposit(id uint64, actID common.ActivityID) error {
	var err error
	err = self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bittrexDepositHistory))
		// actIDBytes, _ := actID.MarshalText()
		actIDBytes, _ := actID.MarshalText()
		return b.Put(boltutil.Uint64ToBytes(id), actIDBytes)
	})
	return err
}

func (self *BoltStorage) StoreTradeHistory(data common.ExchangeTradeHistory) error {
	var err error
	err = self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(tradeHistory))
		for pair, pairHistory := range data {
			pairBk, vErr := b.CreateBucketIfNotExists([]byte(pair))
			if vErr != nil {
				return err
			}
			for _, history := range pairHistory {
				idBytes := []byte(fmt.Sprintf("%s%s", strconv.FormatUint(history.Timestamp, 10), history.ID))
				dataJSON, vErr := json.Marshal(history)
				if vErr != nil {
					return vErr
				}
				vErr = pairBk.Put(idBytes, dataJSON)
				if vErr != nil {
					return vErr
				}
			}
		}
		return nil
	})
	return err
}

func (self *BoltStorage) GetTradeHistory(fromTime, toTime uint64) (common.ExchangeTradeHistory, error) {
	result := common.ExchangeTradeHistory{}
	var err error
	if toTime-fromTime > maxGetTradeHistory {
		return result, errors.New("Time range is too broad, it must be smaller or equal to 3 days (miliseconds)")
	}
	min := []byte(strconv.FormatUint(fromTime, 10))
	max := []byte(strconv.FormatUint(toTime, 10))
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(tradeHistory))
		c := b.Cursor()
		exchangeHistory := common.ExchangeTradeHistory{}
		for key, value := c.First(); key != nil && value == nil; key, value = c.Next() {
			pairBk := b.Bucket(key)
			pairsHistory := []common.TradeHistory{}
			pairCursor := pairBk.Cursor()
			for pairKey, history := pairCursor.Seek(min); pairKey != nil && bytes.Compare(pairKey, max) <= 0; pairKey, history = pairCursor.Next() {
				pairHistory := common.TradeHistory{}
				if vErr := json.Unmarshal(history, &pairHistory); vErr != nil {
					return vErr
				}
				pairsHistory = append(pairsHistory, pairHistory)
			}
			exchangeHistory[common.TokenPairID(key)] = pairsHistory
		}
		result = exchangeHistory
		return nil
	})
	return result, err
}
