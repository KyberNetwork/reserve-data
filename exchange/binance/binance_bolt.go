package binance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/boltdb/bolt"
	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data/common"
)

const (
	tradeHistory       string = "trade_history"
	maxGetTradeHistory uint64 = 3 * 86400000
)

//Storage storage binance information
//including trade history
type Storage struct {
	db *bolt.DB
	l  *zap.SugaredLogger
}

//NewBoltStorage create database and related bucket for binance storage
func NewBoltStorage(path string) (*Storage, error) {
	// init instance
	var err error
	var db *bolt.DB
	db, err = bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}
	// init buckets
	err = db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists([]byte(tradeHistory))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	storage := &Storage{db: db, l: zap.S()}
	return storage, nil
}

//StoreTradeHistory store binance trade history
func (bs *Storage) StoreTradeHistory(data common.ExchangeTradeHistory) error {
	err := bs.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(tradeHistory))
		for pair, pairHistory := range data {
			pairBk, uErr := b.CreateBucketIfNotExists([]byte(pair))
			if uErr != nil {
				return uErr
			}
			for _, history := range pairHistory {
				idBytes := []byte(fmt.Sprintf("%s%s", strconv.FormatUint(history.Timestamp, 10), history.ID))
				dataJSON, uErr := json.Marshal(history)
				if uErr != nil {
					return uErr
				}
				uErr = pairBk.Put(idBytes, dataJSON)
				if uErr != nil {
					return uErr
				}
			}
		}
		return nil
	})
	return err
}

//GetTradeHistory return trade history from binance from time to time
func (bs *Storage) GetTradeHistory(fromTime, toTime uint64) (common.ExchangeTradeHistory, error) {
	result := common.ExchangeTradeHistory{}
	var err error
	if toTime-fromTime > maxGetTradeHistory {
		return result, fmt.Errorf("time range is too broad, it must be smaller or equal to 3 days (miliseconds)")
	}
	min := []byte(strconv.FormatUint(fromTime, 10))
	max := []byte(strconv.FormatUint(toTime, 10))
	err = bs.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(tradeHistory))
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
					bs.l.Warnf("Cannot unmarshal history: %+v", err)
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

//GetLastIDTradeHistory return last id of trade history of a token
//using for query trade history from binance
func (bs *Storage) GetLastIDTradeHistory(pair string) (string, error) {
	history := common.TradeHistory{}
	err := bs.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(tradeHistory))
		pairBk, err := b.CreateBucketIfNotExists([]byte(pair))
		if err != nil {
			bs.l.Warnf("Cannot get pair bucket: %s", err)
			return err
		}
		k, v := pairBk.Cursor().Last()
		if k != nil {
			err = json.Unmarshal(v, &history)
			if err != nil {
				bs.l.Warnf("Cannot unmarshal history: %s", err)
				return err
			}
		}
		return err
	})
	return history.ID, err
}
