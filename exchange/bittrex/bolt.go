package bittrex

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/boltdb/bolt"
)

const (
	BITTREX_DEPOSIT_HISTORY string = "bittrex_deposit_history"
	TRADE_HISTORY           string = "trade_history"
	MAX_GET_TRADE_HISTORY   uint64 = 3 * 86400000
)

type BoltStorage struct {
	db *bolt.DB
}

func uint64ToBytes(u uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, u)
	return b
}

func bytesToUint64(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

func NewBoltStorage(path string) (*BoltStorage, error) {
	// init instance
	var err error
	var db *bolt.DB
	db, err = bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}
	// init buckets
	db.Update(func(tx *bolt.Tx) error {
		tx.CreateBucket([]byte(BITTREX_DEPOSIT_HISTORY))
		_, err = tx.CreateBucketIfNotExists([]byte(TRADE_HISTORY))
		if err != nil {
			return err
		}
		return nil
	})
	storage := &BoltStorage{db}
	return storage, nil
}

func (self *BoltStorage) IsNewBittrexDeposit(id uint64, actID common.ActivityID) bool {
	res := true
	self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BITTREX_DEPOSIT_HISTORY))
		v := b.Get(uint64ToBytes(id))
		if v != nil && string(v) != actID.String() {
			log.Printf("bolt: stored act id - current act id: %s - %s", string(v), actID.String())
			res = false
		}
		return nil
	})
	return res
}

func (self *BoltStorage) RegisterBittrexDeposit(id uint64, actID common.ActivityID) error {
	var err error
	self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BITTREX_DEPOSIT_HISTORY))
		// actIDBytes, _ := actID.MarshalText()
		actIDBytes, _ := actID.MarshalText()
		err = b.Put(uint64ToBytes(id), actIDBytes)
		return nil
	})
	return err
}

func (self *BoltStorage) StoreTradeHistory(data common.ExchangeTradeHistory) error {
	var err error
	err = self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TRADE_HISTORY))
		for pair, pairHistory := range data {
			pairBk, err := b.CreateBucketIfNotExists([]byte(pair))
			if err != nil {
				log.Printf("Cannot create pair history bucket: %s", err.Error())
				return err
			}
			for _, history := range pairHistory {
				idBytes := []byte(fmt.Sprintf("%s%s", strconv.FormatUint(history.Timestamp, 10), history.ID))
				dataJSON, err := json.Marshal(history)
				if err != nil {
					log.Printf("Cannot marshal history: %s", err.Error())
				}
				err = pairBk.Put(idBytes, dataJSON)
				if err != nil {
					log.Printf("Cannot put new data: %s", err.Error())
				}
			}
		}
		return err
	})
	return err
}

func (self *BoltStorage) GetTradeHistory(fromTime, toTime uint64) (common.ExchangeTradeHistory, error) {
	result := common.ExchangeTradeHistory{}
	var err error
	if toTime-fromTime > MAX_GET_TRADE_HISTORY {
		return result, errors.New(fmt.Sprintf("Time range is too broad, it must be smaller or equal to 3 days (miliseconds)"))
	}
	min := []byte(strconv.FormatUint(fromTime, 10))
	max := []byte(strconv.FormatUint(toTime, 10))
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TRADE_HISTORY))
		c := b.Cursor()
		exchangeHistory := common.ExchangeTradeHistory{}
		for key, value := c.First(); key != nil && value == nil; key, value = c.Next() {
			pairBk := b.Bucket(key)
			pairsHistory := []common.TradeHistory{}
			pairCursor := pairBk.Cursor()
			for pairKey, history := pairCursor.Seek(min); pairKey != nil && bytes.Compare(pairKey, max) <= 0; pairKey, history = pairCursor.Next() {
				pairHistory := common.TradeHistory{}
				err = json.Unmarshal(history, &pairHistory)
				if err != nil {
					log.Printf("Cannot unmarshal history: %s", err.Error())
					return err
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

func (self *BoltStorage) GetLastIDTradeHistory(exchange, pair string) (string, error) {
	history := common.TradeHistory{}
	err := self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TRADE_HISTORY))
		exchangeBk, err := b.CreateBucketIfNotExists([]byte(exchange))
		if err != nil {
			log.Printf("Cannot get exchange bucket: %s", err.Error())
			return err
		}
		pairBk, err := exchangeBk.CreateBucketIfNotExists([]byte(pair))
		if err != nil {
			log.Printf("Cannot get pair bucket: %s", err.Error())
			return err
		}
		k, v := pairBk.Cursor().Last()
		if k != nil {
			err = json.Unmarshal(v, &history)
			if err != nil {
				log.Printf("Cannot unmarshal history: %s", err.Error())
				return err
			}
		}
		return err
	})
	return history.ID, err
}
