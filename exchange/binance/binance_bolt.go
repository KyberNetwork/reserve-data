package binance

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/boltdb/bolt"
)

const (
	TRADE_HISTORY         string = "trade_history"
	MAX_GET_TRADE_HISTORY uint64 = 3 * 86400000
)

type BinanceStorage struct {
	mu sync.RWMutex
	db *bolt.DB
}

func NewBoltExchangeStorage(path string) (*BinanceStorage, error) {
	// init instance
	var err error
	var db *bolt.DB
	db, err = bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}
	// init buckets
	err = db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists([]byte(TRADE_HISTORY))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	storage := &BinanceStorage{sync.RWMutex{}, db}
	return storage, nil
}

func (self *BinanceStorage) StoreTradeHistory(data common.ExchangeTradeHistory) error {
	err := self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TRADE_HISTORY))
		for pair, pairHistory := range data {
			pairBk, vErr := b.CreateBucketIfNotExists([]byte(pair))
			if vErr != nil {
				return vErr
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

func (self *BinanceStorage) GetTradeHistory(fromTime, toTime uint64) (common.ExchangeTradeHistory, error) {
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

func (self *BinanceStorage) GetLastIDTradeHistory(exchange, pair string) (string, error) {
	history := common.TradeHistory{}
	err := self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TRADE_HISTORY))
		pairBk, err := b.CreateBucketIfNotExists([]byte(pair))
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
