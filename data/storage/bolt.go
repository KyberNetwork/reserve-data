package storage

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/metric"
	"github.com/boltdb/bolt"
)

const (
	PRICE_BUCKET                       string = "prices"
	RATE_BUCKET                        string = "rates"
	ORDER_BUCKET                       string = "orders"
	ACTIVITY_BUCKET                    string = "activities"
	AUTH_DATA_BUCKET                   string = "auth_data"
	PENDING_ACTIVITY_BUCKET            string = "pending_activities"
	METRIC_BUCKET                      string = "metrics"
	METRIC_TARGET_QUANTITY             string = "target_quantity"
	PENDING_TARGET_QUANTITY            string = "pending_target_quantity"
	TRADE_HISTORY                      string = "trade_history"
	ENABLE_REBALANCE                   string = "enable_rebalance"
	SETRATE_CONTROL                    string = "setrate_control"
	PENDING_PWI_EQUATION               string = "pending_pwi_equation"
	PWI_EQUATION                       string = "pwi_equation"
	INTERMEDIATE_TX                    string = "intermediate_tx"
	EXCHANGE_STATUS                    string = "exchange_status"
	EXCHANGE_NOTIFICATIONS             string = "exchange_notifications"
	MAX_NUMBER_VERSION                 int    = 1000
	MAX_GET_RATES_PERIOD               uint64 = 86400000 //1 days in milisec
	UI64DAY                            uint64 = uint64(time.Hour * 24)
	EXPORT_BATCH                       int    = 1
	AUTH_DATA_EXPIRED_DURATION         uint64 = 1 * 86400000 //10day in milisec
	STABLE_TOKEN_PARAMS_BUCKET         string = "stable-token-params"
	PENDING_STABLE_TOKEN_PARAMS_BUCKET string = "pending-stable-token-params"
	GOLD_BUCKET                        string = "gold_feeds"
)

type BoltStorage struct {
	mu sync.RWMutex
	db *bolt.DB
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
		_, err = tx.CreateBucketIfNotExists([]byte(GOLD_BUCKET))
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists([]byte(PRICE_BUCKET))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(RATE_BUCKET))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(ORDER_BUCKET))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(ACTIVITY_BUCKET))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(PENDING_ACTIVITY_BUCKET))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(AUTH_DATA_BUCKET))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(METRIC_BUCKET))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(METRIC_TARGET_QUANTITY))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(PENDING_TARGET_QUANTITY))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(TRADE_HISTORY))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(ENABLE_REBALANCE))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(SETRATE_CONTROL))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(PENDING_PWI_EQUATION))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(PWI_EQUATION))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(INTERMEDIATE_TX))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(EXCHANGE_STATUS))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(EXCHANGE_NOTIFICATIONS))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(PENDING_STABLE_TOKEN_PARAMS_BUCKET))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(STABLE_TOKEN_PARAMS_BUCKET))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	storage := &BoltStorage{sync.RWMutex{}, db}
	return storage, nil
}

func uint64ToBytes(u uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, u)
	return b
}

func bytesToUint64(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

func reverseSeek(timepoint uint64, c *bolt.Cursor) (uint64, error) {
	version, _ := c.Seek(uint64ToBytes(timepoint))
	if version == nil {
		version, _ = c.Prev()
		if version == nil {
			return 0, errors.New(fmt.Sprintf("There is no data before timepoint %d", timepoint))
		} else {
			return bytesToUint64(version), nil
		}
	} else {
		v := bytesToUint64(version)
		if v == timepoint {
			return v, nil
		} else {
			version, _ = c.Prev()
			if version == nil {
				return 0, errors.New(fmt.Sprintf("There is no data before timepoint %d", timepoint))
			} else {
				return bytesToUint64(version), nil
			}
		}
	}
}

func (self *BoltStorage) CurrentGoldInfoVersion(timepoint uint64) (common.Version, error) {
	var result uint64
	var err error
	err = self.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(GOLD_BUCKET)).Cursor()
		result, err = reverseSeek(timepoint, c)
		return nil
	})
	return common.Version(result), err
}

func (self *BoltStorage) GetGoldInfo(version common.Version) (common.GoldData, error) {
	result := common.GoldData{}
	var err error
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(GOLD_BUCKET))
		data := b.Get(uint64ToBytes(uint64(version)))
		if data == nil {
			err = errors.New(fmt.Sprintf("version %s doesn't exist", string(version)))
		} else {
			err = json.Unmarshal(data, &result)
		}
		return nil
	})
	return result, err
}

func (self *BoltStorage) StoreGoldInfo(data common.GoldData) error {
	var err error
	timepoint := data.Timestamp
	err = self.db.Update(func(tx *bolt.Tx) error {
		var dataJson []byte
		b := tx.Bucket([]byte(GOLD_BUCKET))
		dataJson, err = json.Marshal(data)
		if err != nil {
			return err
		}
		err = b.Put(uint64ToBytes(timepoint), dataJson)
		return err
	})
	return err
}

func (self *BoltStorage) ExportExpiredAuthData(currentTime uint64, fileName string) (nRecord uint64, err error) {
	expiredTimestampByte := uint64ToBytes(currentTime - AUTH_DATA_EXPIRED_DURATION)
	outFile, err := os.Create(fileName)
	defer outFile.Close()
	if err != nil {
		return 0, err
	}

	err = self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(AUTH_DATA_BUCKET))
		c := b.Cursor()

		for k, v := c.First(); k != nil && bytes.Compare(k, expiredTimestampByte) <= 0; k, v = c.Next() {
			timestamp := bytesToUint64(k)

			temp := common.AuthDataSnapshot{}
			err = json.Unmarshal(v, &temp)
			if err != nil {
				return err
			}
			record := common.AuthDataRecord{
				common.Timestamp(strconv.FormatUint(timestamp, 10)),
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
			if err != nil {
				return err
			}
		}
		return nil
	})

	return nRecord, err
}

func (self *BoltStorage) PruneExpiredAuthData(currentTime uint64) (nRecord uint64, err error) {
	expiredTimestampByte := uint64ToBytes(currentTime - AUTH_DATA_EXPIRED_DURATION)

	err = self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(AUTH_DATA_BUCKET))
		c := b.Cursor()
		for k, _ := c.First(); k != nil && bytes.Compare(k, expiredTimestampByte) <= 0; k, _ = c.Next() {
			err = b.Delete(k)
			if err != nil {
				return err
			}
			nRecord++
		}
		return err
	})

	return nRecord, err
}

// PruneOutdatedData Remove first version out of database
func (self *BoltStorage) PruneOutdatedData(tx *bolt.Tx, bucket string) error {
	var err error
	b := tx.Bucket([]byte(bucket))
	c := b.Cursor()
	nExcess := self.GetNumberOfVersion(tx, bucket) - MAX_NUMBER_VERSION
	for i := 0; i < nExcess; i++ {
		k, _ := c.First()
		if k == nil {
			err = fmt.Errorf("There is no previous version in %s", bucket)
			return err
		}
		err = b.Delete([]byte(k))
		if err != nil {
			panic(err)
		}
	}

	return err
}

func (self *BoltStorage) CurrentPriceVersion(timepoint uint64) (common.Version, error) {
	var result uint64
	var err error
	err = self.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(PRICE_BUCKET)).Cursor()
		result, err = reverseSeek(timepoint, c)
		return nil
	})
	return common.Version(result), err
}

// GetNumberOfVersion return number of version storing in a bucket
func (self *BoltStorage) GetNumberOfVersion(tx *bolt.Tx, bucket string) int {
	result := 0
	b := tx.Bucket([]byte(bucket))
	c := b.Cursor()
	for k, _ := c.First(); k != nil; k, _ = c.Next() {
		result++
	}
	return result
}

//GetAllPrices returns the corresponding AllPriceEntry to a particular Version
func (self *BoltStorage) GetAllPrices(version common.Version) (common.AllPriceEntry, error) {
	result := common.AllPriceEntry{}
	var err error
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(PRICE_BUCKET))
		data := b.Get(uint64ToBytes(uint64(version)))
		if data == nil {
			err = errors.New(fmt.Sprintf("version %s doesn't exist", string(version)))
		} else {
			err = json.Unmarshal(data, &result)
		}
		return nil
	})
	return result, err
}

func (self *BoltStorage) GetOnePrice(pair common.TokenPairID, version common.Version) (common.OnePrice, error) {
	result := common.AllPriceEntry{}
	var err error
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(PRICE_BUCKET))
		data := b.Get(uint64ToBytes(uint64(version)))
		if data == nil {
			err = errors.New(fmt.Sprintf("version %s doesn't exist", string(version)))
		} else {
			err = json.Unmarshal(data, &result)
		}
		return nil
	})
	if err != nil {
		return common.OnePrice{}, err
	} else {
		pair, exist := result.Data[pair]
		if exist {
			return pair, nil
		} else {
			return common.OnePrice{}, errors.New("Pair of token is not supported")
		}
	}
}

func (self *BoltStorage) StorePrice(data common.AllPriceEntry, timepoint uint64) error {
	var err error
	err = self.db.Update(func(tx *bolt.Tx) error {
		var dataJson []byte
		b := tx.Bucket([]byte(PRICE_BUCKET))

		// remove outdated data from bucket
		log.Printf("Version number: %d\n", self.GetNumberOfVersion(tx, PRICE_BUCKET))
		self.PruneOutdatedData(tx, PRICE_BUCKET)
		log.Printf("After prune number version: %d\n", self.GetNumberOfVersion(tx, PRICE_BUCKET))

		dataJson, err = json.Marshal(data)
		if err != nil {
			return err
		}
		return b.Put(uint64ToBytes(timepoint), dataJson)
	})
	return err
}

func (self *BoltStorage) CurrentAuthDataVersion(timepoint uint64) (common.Version, error) {
	var result uint64
	var err error
	err = self.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(AUTH_DATA_BUCKET)).Cursor()
		result, err = reverseSeek(timepoint, c)
		return nil
	})
	return common.Version(result), err
}

func (self *BoltStorage) GetAuthData(version common.Version) (common.AuthDataSnapshot, error) {
	result := common.AuthDataSnapshot{}
	var err error
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(AUTH_DATA_BUCKET))
		data := b.Get(uint64ToBytes(uint64(version)))
		if data == nil {
			err = errors.New(fmt.Sprintf("version %s doesn't exist", string(version)))
		} else {
			err = json.Unmarshal(data, &result)
		}
		return nil
	})
	return result, err
}

func (self *BoltStorage) CurrentRateVersion(timepoint uint64) (common.Version, error) {
	var result uint64
	var err error
	err = self.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(RATE_BUCKET)).Cursor()
		result, err = reverseSeek(timepoint, c)
		return nil
	})
	return common.Version(result), err
}

func (self *BoltStorage) GetRates(fromTime, toTime uint64) ([]common.AllRateEntry, error) {
	result := []common.AllRateEntry{}
	if toTime-fromTime > MAX_GET_RATES_PERIOD {
		return result, errors.New(fmt.Sprintf("Time range is too broad, it must be smaller or equal to %d miliseconds", MAX_GET_RATES_PERIOD))
	}
	var err error
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(RATE_BUCKET))
		c := b.Cursor()
		min := uint64ToBytes(fromTime)
		max := uint64ToBytes(toTime)

		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			data := common.AllRateEntry{}
			err = json.Unmarshal(v, &data)
			if err != nil {
				return err
			}
			result = append([]common.AllRateEntry{data}, result...)
		}
		return nil
	})
	return result, err
}

func (self *BoltStorage) GetRate(version common.Version) (common.AllRateEntry, error) {
	result := common.AllRateEntry{}
	var err error
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(RATE_BUCKET))
		data := b.Get(uint64ToBytes(uint64(version)))
		if data == nil {
			err = errors.New(fmt.Sprintf("version %s doesn't exist", string(version)))
		} else {
			err = json.Unmarshal(data, &result)
		}
		return nil
	})
	return result, err
}

func (self *BoltStorage) StoreAuthSnapshot(
	data *common.AuthDataSnapshot, timepoint uint64) error {

	var err error
	err = self.db.Update(func(tx *bolt.Tx) error {
		var dataJson []byte
		b := tx.Bucket([]byte(AUTH_DATA_BUCKET))
		dataJson, err = json.Marshal(data)
		if err != nil {
			return err
		}
		err = b.Put(uint64ToBytes(timepoint), dataJson)
		return err
	})
	return err
}

func (self *BoltStorage) StoreRate(data common.AllRateEntry, timepoint uint64) error {
	log.Printf("Storing rate data to bolt: data(%v), timespoint(%v)", data, timepoint)
	var err error
	var lastEntryjson common.AllRateEntry
	err = self.db.Update(func(tx *bolt.Tx) error {
		var dataJson []byte
		b := tx.Bucket([]byte(RATE_BUCKET))
		c := b.Cursor()
		_, lastEntry := c.Last()
		json.Unmarshal(lastEntry, &lastEntryjson)
		if lastEntryjson.BlockNumber < data.BlockNumber {
			dataJson, err = json.Marshal(data)
			if err != nil {
				return err
			}
			return b.Put(uint64ToBytes(timepoint), dataJson)
		}
		return err
	})
	return err
}

func (self *BoltStorage) Record(
	action string,
	id common.ActivityID,
	destination string,
	params map[string]interface{}, result map[string]interface{},
	estatus string,
	mstatus string,
	timepoint uint64) error {

	var err error
	err = self.db.Update(func(tx *bolt.Tx) error {
		var dataJson []byte
		b := tx.Bucket([]byte(ACTIVITY_BUCKET))
		record := common.ActivityRecord{
			Action:         action,
			ID:             id,
			Destination:    destination,
			Params:         params,
			Result:         result,
			ExchangeStatus: estatus,
			MiningStatus:   mstatus,
			Timestamp:      common.Timestamp(strconv.FormatUint(timepoint, 10)),
		}
		dataJson, err = json.Marshal(record)
		if err != nil {
			return err
		}
		// idByte, _ := id.MarshalText()
		idByte := id.ToBytes()
		err = b.Put(idByte[:], dataJson)
		if err != nil {
			return err
		}
		if record.IsPending() {
			pb := tx.Bucket([]byte(PENDING_ACTIVITY_BUCKET))
			// all other pending set rates should be staled now
			// remove all of them
			// AFTER EXPERIMENT, THIS WILL NOT WORK
			// log.Printf("===> Trying to remove staled set rates")
			// if record.Action == "set_rates" {
			// 	stales := []common.ActivityRecord{}
			// 	c := pb.Cursor()
			// 	for k, v := c.First(); k != nil; k, v = c.Next() {
			// 		record := common.ActivityRecord{}
			// 		log.Printf("===> staled act: %+v", record)
			// 		err = json.Unmarshal(v, &record)
			// 		if err != nil {
			// 			return err
			// 		}
			// 		if record.Action == "set_rates" {
			// 			stales = append(stales, record)
			// 		}
			// 	}
			// 	log.Printf("===> removing staled acts: %+v", stales)
			// 	self.RemoveStalePendingActivities(tx, stales)
			// }
			// after remove all of them, put new set rate activity
			err = pb.Put(idByte[:], dataJson)
		}
		return err
	})
	return err
}

func formatTimepointToActivityID(timepoint uint64, id []byte) []byte {
	if timepoint == 0 {
		return id
	} else {
		activityID := common.NewActivityID(timepoint, "")
		byteID := activityID.ToBytes()
		return byteID[:]
	}
}

func (self *BoltStorage) GetActivity(id common.ActivityID) (common.ActivityRecord, error) {
	result := common.ActivityRecord{}

	err := self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ACTIVITY_BUCKET))
		idBytes := id.ToBytes()
		v := b.Get(idBytes[:])
		if v == nil {
			return errors.New("Cannot find that activity")
		}
		return json.Unmarshal(v, &result)
	})
	return result, err
}

func (self *BoltStorage) GetAllRecords(fromTime, toTime uint64) ([]common.ActivityRecord, error) {
	result := []common.ActivityRecord{}
	var err error
	if (toTime-fromTime)/1000000 > MAX_GET_RATES_PERIOD {
		return result, errors.New(fmt.Sprintf("Time range is too broad, it must be smaller or equal to %d miliseconds", MAX_GET_RATES_PERIOD))
	}
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ACTIVITY_BUCKET))
		c := b.Cursor()
		fkey, _ := c.First()
		lkey, _ := c.Last()
		min := formatTimepointToActivityID(fromTime, fkey)
		max := formatTimepointToActivityID(toTime, lkey)

		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			record := common.ActivityRecord{}
			err = json.Unmarshal(v, &record)
			if err != nil {
				return err
			}
			result = append([]common.ActivityRecord{record}, result...)
		}
		return nil
	})
	return result, err
}

func getLastPendingSetrate(pendings []common.ActivityRecord, minedNonce uint64) (*common.ActivityRecord, error) {
	var maxNonce uint64 = 0
	var maxPrice uint64 = 0
	var result *common.ActivityRecord
	for _, act := range pendings {
		if act.Action == "set_rates" {
			log.Printf("looking for pending set_rates: %+v", act)
			var nonce uint64
			actNonce := act.Result["nonce"]
			if actNonce != nil {
				nonce, _ = strconv.ParseUint(actNonce.(string), 10, 64)
			} else {
				nonce = 0
			}
			if nonce < minedNonce {
				// this is a stale actitivity, ignore it
				continue
			}
			var gasPrice uint64
			actPrice := act.Result["gasPrice"]
			if actPrice != nil {
				gasPrice, _ = strconv.ParseUint(actPrice.(string), 10, 64)
			} else {
				gasPrice = 0
			}
			if nonce == maxNonce {
				if gasPrice > maxPrice {
					maxNonce = nonce
					result = &act
					maxPrice = gasPrice
				}
			} else if nonce > maxNonce {
				maxNonce = nonce
				result = &act
				maxPrice = gasPrice
			}
		}
	}
	return result, nil
}

func (self *BoltStorage) RemoveStalePendingActivities(tx *bolt.Tx, stales []common.ActivityRecord) error {
	pb := tx.Bucket([]byte(PENDING_ACTIVITY_BUCKET))
	for _, stale := range stales {
		idBytes := stale.ID.ToBytes()
		if err := pb.Delete(idBytes[:]); err != nil {
			return err
		}
	}
	return nil
}

func (self *BoltStorage) PendingSetrate(minedNonce uint64) (*common.ActivityRecord, error) {
	pendings, err := self.GetPendingActivities()
	if err != nil {
		return nil, err
	} else {
		return getLastPendingSetrate(pendings, minedNonce)
	}
}

func (self *BoltStorage) GetPendingActivities() ([]common.ActivityRecord, error) {
	result := []common.ActivityRecord{}
	var err error
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(PENDING_ACTIVITY_BUCKET))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			record := common.ActivityRecord{}
			err = json.Unmarshal(v, &record)
			if err != nil {
				return err
			}
			result = append(
				[]common.ActivityRecord{record}, result...)
		}
		return nil
	})
	return result, err
}

func (self *BoltStorage) UpdateActivity(id common.ActivityID, activity common.ActivityRecord) error {
	var err error
	err = self.db.Update(func(tx *bolt.Tx) error {
		pb := tx.Bucket([]byte(PENDING_ACTIVITY_BUCKET))
		// idBytes, _ := id.MarshalText()
		idBytes := id.ToBytes()
		dataJson, err := json.Marshal(activity)
		if err != nil {
			return err
		}
		// only update when it exists in pending activity bucket because
		// It might be deleted if it is replaced by another activity
		found := pb.Get(idBytes[:])
		if found != nil {
			err = pb.Put(idBytes[:], dataJson)
			if err != nil {
				return err
			}
			if !activity.IsPending() {
				err = pb.Delete(idBytes[:])
				if err != nil {
					return err
				}
			}
		}
		b := tx.Bucket([]byte(ACTIVITY_BUCKET))
		if err != nil {
			return err
		}
		return b.Put(idBytes[:], dataJson)
	})
	return err
}

func (self *BoltStorage) HasPendingDeposit(token common.Token, exchange common.Exchange) (bool, error) {
	result := false
	var err error
	err = self.db.View(func(tx *bolt.Tx) error {
		pb := tx.Bucket([]byte(PENDING_ACTIVITY_BUCKET))
		c := pb.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			record := common.ActivityRecord{}
			json.Unmarshal(v, &record)
			if record.Action == "deposit" && record.Params["token"].(string) == token.ID && record.Destination == string(exchange.ID()) {
				result = true
			}
		}
		return nil
	})
	return result, err
}

func (self *BoltStorage) StoreMetric(data *metric.MetricEntry, timepoint uint64) error {
	var err error
	err = self.db.Update(func(tx *bolt.Tx) error {
		var dataJson []byte
		b := tx.Bucket([]byte(METRIC_BUCKET))
		dataJson, err = json.Marshal(data)
		if err != nil {
			return err
		}
		idByte := uint64ToBytes(data.Timestamp)
		err = b.Put(idByte, dataJson)
		return err
	})
	return err
}

func (self *BoltStorage) GetMetric(tokens []common.Token, fromTime, toTime uint64) (map[string]metric.MetricList, error) {
	imResult := map[string]*metric.MetricList{}
	for _, tok := range tokens {
		imResult[tok.ID] = &metric.MetricList{}
	}

	var err error
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(METRIC_BUCKET))
		c := b.Cursor()
		min := uint64ToBytes(fromTime)
		max := uint64ToBytes(toTime)

		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			data := metric.MetricEntry{}
			err = json.Unmarshal(v, &data)
			if err != nil {
				return err
			}
			for tok, m := range data.Data {
				metricList, found := imResult[tok]
				if found {
					*metricList = append(*metricList, metric.TokenMetricResponse{
						Timestamp: data.Timestamp,
						AfpMid:    m.AfpMid,
						Spread:    m.Spread,
					})
				}
			}
		}
		return nil
	})
	result := map[string]metric.MetricList{}
	for k, v := range imResult {
		result[k] = *v
	}
	return result, err
}

func (self *BoltStorage) GetPendingTargetQty() (metric.TokenTargetQty, error) {
	var err error
	var tokenTargetQty metric.TokenTargetQty
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(PENDING_TARGET_QUANTITY))
		_, data := b.Cursor().Last()
		if data == nil {
			err = errors.New("There no pending target quantity")
		} else {
			err = json.Unmarshal(data, &tokenTargetQty)
			if err != nil {
				log.Printf("Cannot unmarshal: %s", err.Error())
			}
		}
		return nil
	})
	return tokenTargetQty, err
}

func (self *BoltStorage) StorePendingTargetQty(data, dataType string) error {
	var err error
	timepoint := common.GetTimepoint()
	tokenTargetQty := metric.TokenTargetQty{}
	err = self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(PENDING_TARGET_QUANTITY))
		_, lastPending := b.Cursor().Last()
		if lastPending != nil {
			err = errors.New("There is another pending target quantity. Please confirm or cancel it before setting new target.")
			return err
		}

		tokenTargetQty.ID = timepoint
		tokenTargetQty.Status = "unconfirmed"
		tokenTargetQty.Data = data
		tokenTargetQty.Type, _ = strconv.ParseInt(dataType, 10, 64)
		idByte := uint64ToBytes(timepoint)
		var dataJson []byte
		dataJson, err = json.Marshal(tokenTargetQty)
		if err != nil {
			return err
		}
		log.Printf("Target to save: %v", dataJson)
		return b.Put(idByte, dataJson)
	})
	return err
}

func (self *BoltStorage) RemovePendingTargetQty() error {
	var err error
	err = self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(PENDING_TARGET_QUANTITY))
		k, lastPending := b.Cursor().Last()
		log.Printf("Last key: %s", k)
		if lastPending == nil {
			return errors.New("There is no pending target quantity.")
		}

		b.Delete([]byte(k))
		return nil
	})
	return err
}

func (self *BoltStorage) CurrentTargetQtyVersion(timepoint uint64) (common.Version, error) {
	var result uint64
	var err error
	err = self.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(METRIC_TARGET_QUANTITY)).Cursor()
		result, err = reverseSeek(timepoint, c)
		return nil
	})
	return common.Version(result), err
}

func (self *BoltStorage) GetTokenTargetQty() (metric.TokenTargetQty, error) {
	tokenTargetQty := metric.TokenTargetQty{}
	version, err := self.CurrentTargetQtyVersion(common.GetTimepoint())
	if err != nil {
		log.Printf("Cannot get version: %s", err.Error())
	}
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(METRIC_TARGET_QUANTITY))
		data := b.Get(uint64ToBytes(uint64(version)))
		if data == nil {
			err = errors.New(fmt.Sprintf("version %s doesn't exist", string(version)))
		} else {
			err = json.Unmarshal(data, &tokenTargetQty)
			if err != nil {
				log.Printf("Cannot unmarshal: %s", err.Error())
			}
		}
		return nil
	})
	return tokenTargetQty, err
}

func (self *BoltStorage) StoreTokenTargetQty(id, data string) error {
	var err error
	var tokenTargetQty metric.TokenTargetQty
	var dataJson []byte
	err = self.db.Update(func(tx *bolt.Tx) error {
		pending := tx.Bucket([]byte(PENDING_TARGET_QUANTITY))
		_, pendingTargetQty := pending.Cursor().Last()

		if pendingTargetQty == nil {
			err = errors.New("There is no pending target activity to confirm.")
			return err
		} else {
			// verify confirm data
			json.Unmarshal(pendingTargetQty, &tokenTargetQty)
			pendingData := tokenTargetQty.Data
			idInt, _ := strconv.ParseUint(id, 10, 64)
			if tokenTargetQty.ID != idInt {
				err = errors.New("Pending target quantity ID does not match")
				return err
			}
			if data != pendingData {
				err = errors.New("Pending target quantity data does not match")
				return err
			}

			// Save to confirmed target quantity
			tokenTargetQty.Status = "confirmed"
			b := tx.Bucket([]byte(METRIC_TARGET_QUANTITY))
			dataJson, err = json.Marshal(tokenTargetQty)
			if err != nil {
				return err
			}
			idByte := uint64ToBytes(common.GetTimepoint())
			return b.Put(idByte, dataJson)
		}
	})
	if err == nil {
		// remove pending target qty
		return self.RemovePendingTargetQty()
	}
	return err
}

func (self *BoltStorage) GetTradeHistory(timepoint uint64) (common.AllTradeHistory, error) {
	result := common.AllTradeHistory{}
	var err error
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TRADE_HISTORY))
		_, data := b.Cursor().First()
		if data == nil {
			err = errors.New(fmt.Sprintf("There no data before timepoint %d", timepoint))
		} else {
			err = json.Unmarshal(data, &result)
		}
		return err
	})
	return result, err
}

func (self *BoltStorage) StoreTradeHistory(data common.AllTradeHistory, timepoint uint64) error {
	var err error
	err = self.db.Update(func(tx *bolt.Tx) error {
		var dataJson []byte
		b := tx.Bucket([]byte(TRADE_HISTORY))
		c := b.Cursor()
		k, v := c.First()
		currentData := common.AllTradeHistory{
			Timestamp: common.GetTimestamp(),
			Data:      map[common.ExchangeID]common.ExchangeTradeHistory{},
		}
		if k != nil {
			json.Unmarshal(v, &currentData)
		}

		// override old data
		for exchange, dataHistory := range data.Data {
			currentDataHistory, exist := currentData.Data[exchange]
			if !exist {
				currentData.Data[exchange] = dataHistory
			} else {
				for pair, pairHistory := range dataHistory {
					currentDataHistory[pair] = pairHistory
				}
			}
		}
		log.Printf("History: %+v", data)

		// add new data
		dataJson, err = json.Marshal(currentData)
		if err != nil {
			return err
		}
		idByte := uint64ToBytes(timepoint)
		return b.Put(idByte, dataJson)
	})
	return err
}

func (self *BoltStorage) GetRebalanceControl() (metric.RebalanceControl, error) {
	var err error
	var result metric.RebalanceControl
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ENABLE_REBALANCE))
		_, data := b.Cursor().First()
		if data == nil {
			result = metric.RebalanceControl{
				Status: false,
			}
			self.StoreRebalanceControl(false)
		} else {
			json.Unmarshal(data, &result)
		}
		return nil
	})
	return result, err
}

func (self *BoltStorage) StoreRebalanceControl(status bool) error {
	var err error
	err = self.db.Update(func(tx *bolt.Tx) error {
		var dataJson []byte
		b := tx.Bucket([]byte(ENABLE_REBALANCE))
		// prune out old data
		c := b.Cursor()
		k, _ := c.First()
		if k != nil {
			err = b.Delete([]byte(k))
			if err != nil {
				return err
			}
		}

		// add new data
		data := metric.RebalanceControl{
			Status: status,
		}
		dataJson, err = json.Marshal(data)
		if err != nil {
			return err
		}
		idByte := uint64ToBytes(common.GetTimepoint())
		return b.Put(idByte, dataJson)
	})
	return err
}

func (self *BoltStorage) GetSetrateControl() (metric.SetrateControl, error) {
	var err error
	var result metric.SetrateControl
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(SETRATE_CONTROL))
		_, data := b.Cursor().First()
		if data == nil {
			result = metric.SetrateControl{
				Status: false,
			}
			self.StoreSetrateControl(false)
		} else {
			json.Unmarshal(data, &result)
		}
		return nil
	})
	return result, err
}

func (self *BoltStorage) StoreSetrateControl(status bool) error {
	var err error
	err = self.db.Update(func(tx *bolt.Tx) error {
		var dataJson []byte
		b := tx.Bucket([]byte(SETRATE_CONTROL))
		// prune out old data
		c := b.Cursor()
		k, _ := c.First()
		if k != nil {
			err = b.Delete([]byte(k))
			if err != nil {
				return err
			}
		}

		// add new data
		data := metric.SetrateControl{
			Status: status,
		}
		dataJson, err = json.Marshal(data)
		if err != nil {
			return err
		}
		idByte := uint64ToBytes(common.GetTimepoint())
		return b.Put(idByte, dataJson)
	})
	return err
}

func (self *BoltStorage) StorePendingPWIEquation(data string) error {
	var err error
	timepoint := common.GetTimepoint()
	saveData := metric.PWIEquation{}
	err = self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(PENDING_PWI_EQUATION))
		c := b.Cursor()
		_, v := c.First()
		if v != nil {
			err = errors.New("There is another pending equation, please confirm or reject to set new equation")
			return err
		} else {
			idByte := uint64ToBytes(timepoint)
			saveData.ID = timepoint
			saveData.Data = data
			if err != nil {
				return err
			}
			dataJson, err := json.Marshal(saveData)
			if err != nil {
				return err
			}
			err = b.Put(idByte, dataJson)
		}
		return err
	})
	return err
}

func (self *BoltStorage) GetPendingPWIEquation() (metric.PWIEquation, error) {
	var err error
	var result metric.PWIEquation
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(PENDING_PWI_EQUATION))
		c := b.Cursor()
		_, v := c.First()
		if v == nil {
			return errors.New("There no pending equation")
		} else {
			json.Unmarshal(v, &result)
		}
		return nil
	})
	return result, err
}

func (self *BoltStorage) StorePWIEquation(data string) error {
	var err error
	err = self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(PENDING_PWI_EQUATION))
		c := b.Cursor()
		_, v := c.First()
		if v == nil {
			err = errors.New("There no pending equation")
			return err
		} else {
			p := tx.Bucket([]byte(PWI_EQUATION))
			idByte := uint64ToBytes(common.GetTimepoint())
			pending := metric.PWIEquation{}
			json.Unmarshal(v, &pending)
			if pending.Data != data {
				err = errors.New("Confirm data does not match pending data")
				return err
			}
			saveData, err := json.Marshal(pending)
			if err != nil {
				return err
			}
			err = p.Put(idByte, saveData)
		}
		return err
	})
	if err == nil {
		self.RemovePendingPWIEquation()
	}
	return err
}

func (self *BoltStorage) GetPWIEquation() (metric.PWIEquation, error) {
	var err error
	var result metric.PWIEquation
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(PWI_EQUATION))
		c := b.Cursor()
		_, v := c.Last()
		if v == nil {
			err = errors.New("There is no equation")
			return err
		} else {
			json.Unmarshal(v, &result)
		}
		return err
	})
	return result, err
}

func (self *BoltStorage) RemovePendingPWIEquation() error {
	var err error
	err = self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(PENDING_PWI_EQUATION))
		c := b.Cursor()
		k, _ := c.First()
		if k != nil {
			err = b.Delete([]byte(k))
			if err != nil {
				return err
			}
		} else {
			err = errors.New("There is no pending data")
		}
		return err
	})
	return err
}

func (self *BoltStorage) GetExchangeStatus() (common.ExchangesStatus, error) {
	result := make(common.ExchangesStatus)
	var err error
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(EXCHANGE_STATUS))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var exstat common.ExStatus
			err := json.Unmarshal(v, &exstat)
			if err != nil {
				return err
			}
			result[string(k)] = exstat
		}
		return nil
	})
	return result, err
}

func (self *BoltStorage) UpdateExchangeStatus(data common.ExchangesStatus) error {
	var err error
	err = self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(EXCHANGE_STATUS))
		for k, v := range data {
			dataJson, err := json.Marshal(v)
			if err != nil {
				return err
			}
			err = b.Put([]byte(k), dataJson)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func (self *BoltStorage) UpdateExchangeNotification(
	exchange, action, token string, fromTime, toTime uint64, isWarning bool, msg string) error {
	var err error
	err = self.db.Update(func(tx *bolt.Tx) error {
		exchangeBk := tx.Bucket([]byte(EXCHANGE_NOTIFICATIONS))
		b, err := exchangeBk.CreateBucketIfNotExists([]byte(exchange))
		if err != nil {
			return err
		}
		key := fmt.Sprintf("%s_%s", action, token)
		noti := common.ExchangeNotiContent{
			FromTime:  fromTime,
			ToTime:    toTime,
			IsWarning: isWarning,
			Message:   msg,
		}

		// update new value
		dataJSON, err := json.Marshal(noti)
		if err != nil {
			return err
		}
		err = b.Put([]byte(key), dataJSON)
		return err
	})
	return err
}

func (self *BoltStorage) GetExchangeNotifications() (common.ExchangeNotifications, error) {
	result := common.ExchangeNotifications{}
	var err error
	err = self.db.View(func(tx *bolt.Tx) error {
		exchangeBks := tx.Bucket([]byte(EXCHANGE_NOTIFICATIONS))
		c := exchangeBks.Cursor()
		for name, bucket := c.First(); name != nil; name, bucket = c.Next() {
			// if bucket == nil, then name is a child bucket name (according to bolt docs)
			if bucket == nil {
				b := exchangeBks.Bucket(name)
				c := b.Cursor()
				actionContent := common.ExchangeActionNoti{}
				for k, v := c.First(); k != nil; k, v = c.Next() {
					actionToken := strings.Split(string(k), "_")
					action := actionToken[0]
					token := actionToken[1]
					notiContent := common.ExchangeNotiContent{}
					json.Unmarshal(v, &notiContent)
					tokenContent, exist := actionContent[action]
					if !exist {
						tokenContent = common.ExchangeTokenNoti{}
					}
					tokenContent[token] = notiContent
					actionContent[action] = tokenContent
				}
				result[string(name)] = actionContent
			}
		}
		return err
	})
	return result, err
}

func (self *BoltStorage) SetStableTokenParams(value []byte) error {
	var err error
	k := uint64ToBytes(1)
	temp := make(map[string]interface{})
	vErr := json.Unmarshal(value, &temp)
	if vErr != nil {
		return fmt.Errorf("Rejected: Data could not be unmarshalled to defined format: %s", vErr)
	}
	err = self.db.Update(func(tx *bolt.Tx) error {
		b, uErr := tx.CreateBucketIfNotExists([]byte(PENDING_STABLE_TOKEN_PARAMS_BUCKET))
		if uErr != nil {
			return uErr
		}
		if b.Get(k) != nil {
			return fmt.Errorf("Currently there is a pending record")
		}
		return b.Put(k, value)
	})
	return err
}

func (self *BoltStorage) ConfirmStableTokenParams(value []byte) error {
	var err error
	k := uint64ToBytes(1)
	temp := make(map[string]interface{})
	vErr := json.Unmarshal(value, &temp)
	if vErr != nil {
		return fmt.Errorf("Rejected: Data could not be unmarshalled to defined format: %s", vErr)
	}
	pending, err := self.GetPendingStableTokenParams()
	if eq := reflect.DeepEqual(pending, temp); !eq {
		return fmt.Errorf("Rejected: confiming data isn't consistent")
	}

	err = self.db.Update(func(tx *bolt.Tx) error {
		b, uErr := tx.CreateBucketIfNotExists([]byte(STABLE_TOKEN_PARAMS_BUCKET))
		if uErr != nil {
			return uErr
		}
		return b.Put(k, value)
	})
	if err != nil {
		return err
	}
	err = self.RemovePendingStableTokenParams()
	return err
}

func (self *BoltStorage) GetStableTokenParams() (map[string]interface{}, error) {
	k := uint64ToBytes(1)
	result := make(map[string]interface{})
	err := self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(STABLE_TOKEN_PARAMS_BUCKET))
		if b == nil {
			return fmt.Errorf("Bucket hasn't exist yet")
		}
		record := b.Get(k)
		if record == nil {
			return nil
		}
		vErr := json.Unmarshal(record, &result)
		if vErr != nil {
			return vErr
		}
		return nil
	})
	return result, err
}

func (self *BoltStorage) GetPendingStableTokenParams() (map[string]interface{}, error) {
	k := uint64ToBytes(1)
	result := make(map[string]interface{})
	err := self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(PENDING_STABLE_TOKEN_PARAMS_BUCKET))
		if b == nil {
			return fmt.Errorf("Bucket hasn't exist yet")
		}
		record := b.Get(k)
		if record == nil {
			return nil
		}
		vErr := json.Unmarshal(record, &result)
		if vErr != nil {
			return vErr
		}
		return nil
	})
	return result, err
}

func (self *BoltStorage) RemovePendingStableTokenParams() error {
	k := uint64ToBytes(1)
	err := self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(PENDING_STABLE_TOKEN_PARAMS_BUCKET))
		if b == nil {
			return fmt.Errorf("Bucket hasn't existed yet")
		}
		record := b.Get(k)
		if record == nil {
			return fmt.Errorf("Bucket is empty")
		}
		return b.Delete(k)
	})
	return err
}
