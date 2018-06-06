package storage

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/KyberNetwork/reserve-data/boltutil"
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
	ENABLE_REBALANCE                   string = "enable_rebalance"
	SETRATE_CONTROL                    string = "setrate_control"
	PENDING_PWI_EQUATION               string = "pending_pwi_equation"
	PWI_EQUATION                       string = "pwi_equation"
	INTERMEDIATE_TX                    string = "intermediate_tx"
	EXCHANGE_STATUS                    string = "exchange_status"
	EXCHANGE_NOTIFICATIONS             string = "exchange_notifications"
	MAX_NUMBER_VERSION                 int    = 1000
	MAX_GET_RATES_PERIOD               uint64 = 86400000      //1 days in milisec
	AUTH_DATA_EXPIRED_DURATION         uint64 = 10 * 86400000 //10day in milisec
	STABLE_TOKEN_PARAMS_BUCKET         string = "stable-token-params"
	PENDING_STABLE_TOKEN_PARAMS_BUCKET string = "pending-stable-token-params"
	GOLD_BUCKET                        string = "gold_feeds"

	// PENDING_TARGET_QUANTITY_V2 constant for bucket name for pending target quantity v2
	PENDING_TARGET_QUANTITY_V2 string = "pending_target_qty_v2"
	// TARGET_QUANTITY_V2 constant for bucet name for target quantity v2
	TARGET_QUANTITY_V2 string = "target_quantity_v2"

	// PENDING_PWI_EQUATION_V2 is the bucket name for storing pending
	// pwi equation for later approval.
	PENDING_PWI_EQUATION_V2 string = "pending_pwi_equation_v2"
	// PWI_EQUATION_V2 stores the PWI equations after confirmed.
	PWI_EQUATION_V2 string = "pwi_equation_v2"
)

// BoltStorage is the storage implementation of data.Storage interface
// that uses BoltDB as its storage engine.
type BoltStorage struct {
	mu sync.RWMutex
	db *bolt.DB
}

// NewBoltStorage creates a new BoltStorage instance with the database
// filename given in parameter.
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
		if _, vErr := tx.CreateBucketIfNotExists([]byte(GOLD_BUCKET)); vErr != nil {
			return vErr
		}

		if _, vErr := tx.CreateBucketIfNotExists([]byte(PRICE_BUCKET)); vErr != nil {
			return vErr
		}
		if _, vErr := tx.CreateBucketIfNotExists([]byte(RATE_BUCKET)); vErr != nil {
			return vErr
		}
		if _, vErr := tx.CreateBucketIfNotExists([]byte(ORDER_BUCKET)); vErr != nil {
			return vErr
		}
		if _, vErr := tx.CreateBucketIfNotExists([]byte(ACTIVITY_BUCKET)); vErr != nil {
			return vErr
		}
		if _, vErr := tx.CreateBucketIfNotExists([]byte(PENDING_ACTIVITY_BUCKET)); vErr != nil {
			return vErr
		}
		if _, vErr := tx.CreateBucketIfNotExists([]byte(AUTH_DATA_BUCKET)); vErr != nil {
			return vErr
		}
		if _, vErr := tx.CreateBucketIfNotExists([]byte(METRIC_BUCKET)); vErr != nil {
			return vErr
		}
		if _, vErr := tx.CreateBucketIfNotExists([]byte(METRIC_TARGET_QUANTITY)); vErr != nil {
			return vErr
		}
		if _, vErr := tx.CreateBucketIfNotExists([]byte(PENDING_TARGET_QUANTITY)); vErr != nil {
			return vErr
		}
		if _, vErr := tx.CreateBucketIfNotExists([]byte(ENABLE_REBALANCE)); vErr != nil {
			return vErr
		}
		if _, vErr := tx.CreateBucketIfNotExists([]byte(SETRATE_CONTROL)); vErr != nil {
			return vErr
		}
		if _, vErr := tx.CreateBucketIfNotExists([]byte(PENDING_PWI_EQUATION)); vErr != nil {
			return vErr
		}
		if _, vErr := tx.CreateBucketIfNotExists([]byte(PWI_EQUATION)); vErr != nil {
			return vErr
		}
		if _, vErr := tx.CreateBucketIfNotExists([]byte(INTERMEDIATE_TX)); vErr != nil {
			return vErr
		}
		if _, vErr := tx.CreateBucketIfNotExists([]byte(EXCHANGE_STATUS)); vErr != nil {
			return vErr
		}
		if _, vErr := tx.CreateBucketIfNotExists([]byte(EXCHANGE_NOTIFICATIONS)); vErr != nil {
			return vErr
		}
		if _, vErr := tx.CreateBucketIfNotExists([]byte(PENDING_STABLE_TOKEN_PARAMS_BUCKET)); vErr != nil {
			return vErr
		}
		if _, vErr := tx.CreateBucketIfNotExists([]byte(STABLE_TOKEN_PARAMS_BUCKET)); vErr != nil {
			return vErr
		}

		if _, vErr := tx.CreateBucketIfNotExists([]byte(PENDING_TARGET_QUANTITY_V2)); vErr != nil {
			return vErr
		}

		if _, vErr := tx.CreateBucketIfNotExists([]byte(TARGET_QUANTITY_V2)); vErr != nil {
			return vErr
		}

		if _, vErr := tx.CreateBucketIfNotExists([]byte(PENDING_PWI_EQUATION_V2)); vErr != nil {
			return vErr
		}
		if _, vErr := tx.CreateBucketIfNotExists([]byte(PWI_EQUATION_V2)); vErr != nil {
			return vErr
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	storage := &BoltStorage{
		mu: sync.RWMutex{},
		db: db,
	}
	return storage, nil
}

// reverseSeek returns the most recent time point to the given one in parameter.
// It returns an error if no there is no record exists before the given time point.
func reverseSeek(timepoint uint64, c *bolt.Cursor) (uint64, error) {
	version, _ := c.Seek(boltutil.Uint64ToBytes(timepoint))
	if version == nil {
		version, _ = c.Prev()
		if version == nil {
			return 0, fmt.Errorf("There is no data before timepoint %d", timepoint)
		}
		return boltutil.BytesToUint64(version), nil
	}
	v := boltutil.BytesToUint64(version)
	if v == timepoint {
		return v, nil
	}
	version, _ = c.Prev()
	if version == nil {
		return 0, fmt.Errorf("There is no data before timepoint %d", timepoint)
	}
	return boltutil.BytesToUint64(version), nil
}

// CurrentGoldInfoVersion returns the most recent time point of gold info record.
// It implements data.GlobalStorage interface.
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

// GetGoldInfo returns gold info at given time point. It implements data.GlobalStorage interface.
func (self *BoltStorage) GetGoldInfo(version common.Version) (common.GoldData, error) {
	result := common.GoldData{}
	var err error
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(GOLD_BUCKET))
		data := b.Get(boltutil.Uint64ToBytes(uint64(version)))
		if data == nil {
			err = fmt.Errorf("version %s doesn't exist", string(version))
		} else {
			err = json.Unmarshal(data, &result)
		}
		return err
	})
	return result, err
}

// StoreGoldInfo stores the given gold information to database. It implements fetcher.GlobalStorage interface.
func (self *BoltStorage) StoreGoldInfo(data common.GoldData) error {
	var err error
	timepoint := data.Timestamp
	err = self.db.Update(func(tx *bolt.Tx) error {
		var dataJSON []byte
		b := tx.Bucket([]byte(GOLD_BUCKET))
		dataJSON, vErr := json.Marshal(data)
		if vErr != nil {
			return vErr
		}
		return b.Put(boltutil.Uint64ToBytes(timepoint), dataJSON)
	})
	return err
}

func (self *BoltStorage) ExportExpiredAuthData(currentTime uint64, fileName string) (nRecord uint64, err error) {
	expiredTimestampByte := boltutil.Uint64ToBytes(currentTime - AUTH_DATA_EXPIRED_DURATION)
	outFile, err := os.Create(fileName)
	if err != nil {
		return 0, err
	}
	defer func() {
		if vErr := outFile.Close(); vErr != nil {
			log.Printf("Close file error: %s", vErr.Error())
		}
	}()

	err = self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(AUTH_DATA_BUCKET))
		c := b.Cursor()

		for k, v := c.First(); k != nil && bytes.Compare(k, expiredTimestampByte) <= 0; k, v = c.Next() {
			timestamp := boltutil.BytesToUint64(k)

			temp := common.AuthDataSnapshot{}
			if vErr := json.Unmarshal(v, &temp); vErr != nil {
				return vErr
			}
			record := common.NewAuthDataRecord(
				common.Timestamp(strconv.FormatUint(timestamp, 10)),
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
			if err != nil {
				return err
			}
		}
		return nil
	})

	return nRecord, err
}

func (self *BoltStorage) PruneExpiredAuthData(currentTime uint64) (nRecord uint64, err error) {
	expiredTimestampByte := boltutil.Uint64ToBytes(currentTime - AUTH_DATA_EXPIRED_DURATION)

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
		return err
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
		data := b.Get(boltutil.Uint64ToBytes(uint64(version)))
		if data == nil {
			err = fmt.Errorf("version %s doesn't exist", string(version))
		} else {
			err = json.Unmarshal(data, &result)
		}
		return err
	})
	return result, err
}

func (self *BoltStorage) GetOnePrice(pair common.TokenPairID, version common.Version) (common.OnePrice, error) {
	result := common.AllPriceEntry{}
	var err error
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(PRICE_BUCKET))
		data := b.Get(boltutil.Uint64ToBytes(uint64(version)))
		if data == nil {
			err = fmt.Errorf("version %s doesn't exist", string(version))
		} else {
			err = json.Unmarshal(data, &result)
		}
		return err
	})
	if err != nil {
		return common.OnePrice{}, err
	}
	dataPair, exist := result.Data[pair]
	if exist {
		return dataPair, nil
	}
	return common.OnePrice{}, errors.New("Pair of token is not supported")
}

func (self *BoltStorage) StorePrice(data common.AllPriceEntry, timepoint uint64) error {
	var err error
	err = self.db.Update(func(tx *bolt.Tx) error {
		var dataJSON []byte
		b := tx.Bucket([]byte(PRICE_BUCKET))

		// remove outdated data from bucket
		log.Printf("Version number: %d\n", self.GetNumberOfVersion(tx, PRICE_BUCKET))
		if vErr := self.PruneOutdatedData(tx, PRICE_BUCKET); vErr != nil {
			log.Printf("Prune out data: %s", vErr.Error())
			return vErr
		}

		dataJSON, err = json.Marshal(data)
		if err != nil {
			return err
		}
		return b.Put(boltutil.Uint64ToBytes(timepoint), dataJSON)
	})
	return err
}

func (self *BoltStorage) CurrentAuthDataVersion(timepoint uint64) (common.Version, error) {
	var result uint64
	var err error
	err = self.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(AUTH_DATA_BUCKET)).Cursor()
		result, err = reverseSeek(timepoint, c)
		return err
	})
	return common.Version(result), err
}

func (self *BoltStorage) GetAuthData(version common.Version) (common.AuthDataSnapshot, error) {
	result := common.AuthDataSnapshot{}
	var err error
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(AUTH_DATA_BUCKET))
		data := b.Get(boltutil.Uint64ToBytes(uint64(version)))
		if data == nil {
			err = fmt.Errorf("version %s doesn't exist", string(version))
		} else {
			err = json.Unmarshal(data, &result)
		}
		return err
	})
	return result, err
}

//CurrentRateVersion return current rate version
func (self *BoltStorage) CurrentRateVersion(timepoint uint64) (common.Version, error) {
	var result uint64
	var err error
	err = self.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(RATE_BUCKET)).Cursor()
		result, err = reverseSeek(timepoint, c)
		return err
	})
	return common.Version(result), err
}

//GetRates return rates history
func (self *BoltStorage) GetRates(fromTime, toTime uint64) ([]common.AllRateEntry, error) {
	result := []common.AllRateEntry{}
	if toTime-fromTime > MAX_GET_RATES_PERIOD {
		return result, fmt.Errorf("Time range is too broad, it must be smaller or equal to %d miliseconds", MAX_GET_RATES_PERIOD)
	}
	var err error
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(RATE_BUCKET))
		c := b.Cursor()
		min := boltutil.Uint64ToBytes(fromTime)
		max := boltutil.Uint64ToBytes(toTime)

		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			data := common.AllRateEntry{}
			err = json.Unmarshal(v, &data)
			if err != nil {
				return err
			}
			result = append([]common.AllRateEntry{data}, result...)
		}
		return err
	})
	return result, err
}

func (self *BoltStorage) GetRate(version common.Version) (common.AllRateEntry, error) {
	result := common.AllRateEntry{}
	var err error
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(RATE_BUCKET))
		data := b.Get(boltutil.Uint64ToBytes(uint64(version)))
		if data == nil {
			err = fmt.Errorf("version %s doesn't exist", string(version))
		} else {
			err = json.Unmarshal(data, &result)
		}
		return err
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
		err = b.Put(boltutil.Uint64ToBytes(timepoint), dataJson)
		return err
	})
	return err
}

//StoreRate store rate history
func (self *BoltStorage) StoreRate(data common.AllRateEntry, timepoint uint64) error {
	log.Printf("Storing rate data to bolt: data(%v), timespoint(%v)", data, timepoint)
	err := self.db.Update(func(tx *bolt.Tx) error {
		var (
			uErr      error
			lastEntry common.AllRateEntry
			dataJson  []byte
		)

		b := tx.Bucket([]byte(RATE_BUCKET))
		c := b.Cursor()
		lastKey, lastValue := c.Last()
		if lastKey == nil {
			log.Printf("Bucket %s is empty", RATE_BUCKET)
		} else {
			if uErr = json.Unmarshal(lastValue, &lastEntry); uErr != nil {
				return uErr
			}
		}
		// we still update when blocknumber is not changed because we want
		// to update the version and timestamp so api users will get
		// the newest data even it is identical to the old one.
		if lastEntry.BlockNumber <= data.BlockNumber {
			if dataJson, uErr = json.Marshal(data); uErr != nil {
				return uErr
			}
			return b.Put(boltutil.Uint64ToBytes(timepoint), dataJson)
		}
		return fmt.Errorf("rejected storing rates with smaller block number: %d, stored: %d",
			data.BlockNumber,
			lastEntry.BlockNumber)
	})
	return err
}

//Record save activity
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
	}
	activityID := common.NewActivityID(timepoint, "")
	byteID := activityID.ToBytes()
	return byteID[:]
}

//GetActivity get activity
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
		return result, fmt.Errorf("Time range is too broad, it must be smaller or equal to %d miliseconds", MAX_GET_RATES_PERIOD)
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
		return err
	})
	return result, err
}

func getLastAndCountPendingSetrate(pendings []common.ActivityRecord, minedNonce uint64) (*common.ActivityRecord, uint64, error) {
	var maxNonce uint64
	var maxPrice uint64
	var result *common.ActivityRecord
	var count uint64
	for i, act := range pendings {
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
					result = &pendings[i]
					maxPrice = gasPrice
				}
				count++
			} else if nonce > maxNonce {
				maxNonce = nonce
				result = &pendings[i]
				maxPrice = gasPrice
				count = 1
			}
		}
	}
	return result, count, nil
}

//RemovePendingActivities remove it
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

//PendingSetrate return pending set rate activity
func (self *BoltStorage) PendingSetrate(minedNonce uint64) (*common.ActivityRecord, uint64, error) {
	pendings, err := self.GetPendingActivities()
	if err != nil {
		return nil, 0, err
	}
	return getLastAndCountPendingSetrate(pendings, minedNonce)
}

//GetPendingActivities return pending activities
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
		return err
	})
	return result, err
}

//UpdateActivity update activity info
func (self *BoltStorage) UpdateActivity(id common.ActivityID, activity common.ActivityRecord) error {
	var err error
	err = self.db.Update(func(tx *bolt.Tx) error {
		pb := tx.Bucket([]byte(PENDING_ACTIVITY_BUCKET))
		idBytes := id.ToBytes()
		dataJSON, uErr := json.Marshal(activity)
		if uErr != nil {
			return uErr
		}
		// only update when it exists in pending activity bucket because
		// It might be deleted if it is replaced by another activity
		found := pb.Get(idBytes[:])
		if found != nil {
			uErr = pb.Put(idBytes[:], dataJSON)
			if uErr != nil {
				return uErr
			}
			if !activity.IsPending() {
				uErr = pb.Delete(idBytes[:])
				if uErr != nil {
					return uErr
				}
			}
		}
		b := tx.Bucket([]byte(ACTIVITY_BUCKET))
		if uErr != nil {
			return uErr
		}
		return b.Put(idBytes[:], dataJSON)
	})
	return err
}

//HasPendingDeposit check if a deposit is pending
func (self *BoltStorage) HasPendingDeposit(token common.Token, exchange common.Exchange) (bool, error) {
	result := false
	var err error
	err = self.db.View(func(tx *bolt.Tx) error {
		pb := tx.Bucket([]byte(PENDING_ACTIVITY_BUCKET))
		c := pb.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			record := common.ActivityRecord{}
			if vErr := json.Unmarshal(v, &record); vErr != nil {
				return vErr
			}
			if record.Action == "deposit" && record.Params["token"].(string) == token.ID && record.Destination == string(exchange.ID()) {
				result = true
			}
		}
		return nil
	})
	return result, err
}

//StoreMetric store metric info
func (self *BoltStorage) StoreMetric(data *metric.MetricEntry, timepoint uint64) error {
	var err error
	err = self.db.Update(func(tx *bolt.Tx) error {
		var dataJson []byte
		b := tx.Bucket([]byte(METRIC_BUCKET))
		dataJson, err = json.Marshal(data)
		if err != nil {
			return err
		}
		idByte := boltutil.Uint64ToBytes(data.Timestamp)
		err = b.Put(idByte, dataJson)
		return err
	})
	return err
}

//GetMetric return metric data
func (self *BoltStorage) GetMetric(tokens []common.Token, fromTime, toTime uint64) (map[string]metric.MetricList, error) {
	imResult := map[string]*metric.MetricList{}
	for _, tok := range tokens {
		imResult[tok.ID] = &metric.MetricList{}
	}

	var err error
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(METRIC_BUCKET))
		c := b.Cursor()
		min := boltutil.Uint64ToBytes(fromTime)
		max := boltutil.Uint64ToBytes(toTime)

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

//GetPendingTargetQty return pending target quantity if available
func (self *BoltStorage) GetPendingTargetQty() (metric.TokenTargetQty, error) {
	var err error
	var tokenTargetQty metric.TokenTargetQty
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(PENDING_TARGET_QUANTITY))
		_, data := b.Cursor().Last()
		if data == nil {
			return errors.New("There no pending target quantity")
		}
		if err = json.Unmarshal(data, &tokenTargetQty); err != nil {
			return err
		}
		return err
	})
	return tokenTargetQty, err
}

//StorePendingTargetQty store pending target quantity if there is none
func (self *BoltStorage) StorePendingTargetQty(data, dataType string) error {
	var err error
	timepoint := common.GetTimepoint()
	tokenTargetQty := metric.TokenTargetQty{}
	err = self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(PENDING_TARGET_QUANTITY))
		_, lastPending := b.Cursor().Last()
		if lastPending != nil {
			err = errors.New("there is another pending target quantity. please confirm or cancel it before setting new target")
			return err
		}

		tokenTargetQty.ID = timepoint
		tokenTargetQty.Status = "unconfirmed"
		tokenTargetQty.Data = data
		tokenTargetQty.Type, _ = strconv.ParseInt(dataType, 10, 64)
		idByte := boltutil.Uint64ToBytes(timepoint)
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

//RemovePendingTargetQty remove pending target quantity
func (self *BoltStorage) RemovePendingTargetQty() error {
	var err error
	err = self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(PENDING_TARGET_QUANTITY))
		k, lastPending := b.Cursor().Last()
		log.Printf("Last key: %s", k)
		if lastPending == nil {
			return errors.New("there is no pending target quantity")
		}
		return b.Delete([]byte(k))
	})
	return err
}

//CurrentTargetQtyVersion return current target quantity version
func (self *BoltStorage) CurrentTargetQtyVersion(timepoint uint64) (common.Version, error) {
	var result uint64
	var err error
	err = self.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(METRIC_TARGET_QUANTITY)).Cursor()
		result, err = reverseSeek(timepoint, c)
		return err
	})
	return common.Version(result), err
}

//GetTokenTargetQty get target quantity
func (self *BoltStorage) GetTokenTargetQty() (metric.TokenTargetQty, error) {
	var (
		tokenTargetQty = metric.TokenTargetQty{}
		err            error
	)
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(METRIC_TARGET_QUANTITY))
		c := b.Cursor()
		result, vErr := reverseSeek(common.GetTimepoint(), c)
		if vErr != nil {
			return vErr
		}
		data := b.Get(boltutil.Uint64ToBytes(result))
		// be defensive, but this should never happen
		if data == nil {
			return fmt.Errorf("version %d doesn't exist", result)
		}
		return json.Unmarshal(data, &tokenTargetQty)
	})
	return tokenTargetQty, err
}

//StoreTokenTargetQty store token target quantity in db
func (self *BoltStorage) StoreTokenTargetQty(id, data string) error {
	var err error
	var tokenTargetQty metric.TokenTargetQty
	var dataJSON []byte
	err = self.db.Update(func(tx *bolt.Tx) error {
		pending := tx.Bucket([]byte(PENDING_TARGET_QUANTITY))
		_, pendingTargetQty := pending.Cursor().Last()

		if pendingTargetQty == nil {
			return errors.New("there is no pending target activity to confirm")
		}
		// verify confirm data
		if vErr := json.Unmarshal(pendingTargetQty, &tokenTargetQty); vErr != nil {
			return vErr
		}
		pendingData := tokenTargetQty.Data
		idInt, vErr := strconv.ParseUint(id, 10, 64)
		if vErr != nil {
			return vErr
		}
		if tokenTargetQty.ID != idInt {
			return errors.New("pending target quantity ID does not match")
		}
		if data != pendingData {
			return errors.New("pending target quantity data does not match")
		}

		// Save to confirmed target quantity
		tokenTargetQty.Status = "confirmed"
		b := tx.Bucket([]byte(METRIC_TARGET_QUANTITY))
		dataJSON, err = json.Marshal(tokenTargetQty)
		if err != nil {
			return err
		}
		idByte := boltutil.Uint64ToBytes(common.GetTimepoint())
		return b.Put(idByte, dataJSON)
	})
	if err == nil {
		// remove pending target qty
		return self.RemovePendingTargetQty()
	}
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
			return self.StoreRebalanceControl(false)
		}
		if vErr := json.Unmarshal(data, &result); vErr != nil {
			return vErr
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
		idByte := boltutil.Uint64ToBytes(common.GetTimepoint())
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
			return self.StoreSetrateControl(false)
		}
		if vErr := json.Unmarshal(data, &result); vErr != nil {
			return vErr
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
		idByte := boltutil.Uint64ToBytes(common.GetTimepoint())
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
			return errors.New("There is another pending equation, please confirm or reject to set new equation")
		}
		idByte := boltutil.Uint64ToBytes(timepoint)
		saveData.ID = timepoint
		saveData.Data = data
		dataJSON, uErr := json.Marshal(saveData)
		if uErr != nil {
			return uErr
		}
		return b.Put(idByte, dataJSON)
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
		}
		if vErr := json.Unmarshal(v, &result); vErr != nil {
			return vErr
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
			return errors.New("There no pending equation")
		}
		p := tx.Bucket([]byte(PWI_EQUATION))
		idByte := boltutil.Uint64ToBytes(common.GetTimepoint())
		pending := metric.PWIEquation{}
		json.Unmarshal(v, &pending)
		if pending.Data != data {
			return errors.New("Confirm data does not match pending data")
		}
		saveData, uErr := json.Marshal(pending)
		if uErr != nil {
			return uErr
		}
		return p.Put(idByte, saveData)
	})
	if err == nil {
		return self.RemovePendingPWIEquation()
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
		}
		return json.Unmarshal(v, &result)
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

// GetExchangeStatus get exchange status to dashboard and analytics
func (self *BoltStorage) GetExchangeStatus() (common.ExchangesStatus, error) {
	result := make(common.ExchangesStatus)
	var err error
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(EXCHANGE_STATUS))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var exstat common.ExStatus
			if _, vErr := common.GetExchange(strings.ToLower(string(k))); vErr != nil {
				continue
			}
			if vErr := json.Unmarshal(v, &exstat); vErr != nil {
				return vErr
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
			dataJSON, uErr := json.Marshal(v)
			if uErr != nil {
				return uErr
			}
			if uErr = b.Put([]byte(k), dataJSON); uErr != nil {
				return uErr
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
		b, uErr := exchangeBk.CreateBucketIfNotExists([]byte(exchange))
		if uErr != nil {
			return uErr
		}
		key := fmt.Sprintf("%s_%s", action, token)
		noti := common.ExchangeNotiContent{
			FromTime:  fromTime,
			ToTime:    toTime,
			IsWarning: isWarning,
			Message:   msg,
		}

		// update new value
		dataJSON, uErr := json.Marshal(noti)
		if uErr != nil {
			return uErr
		}
		return b.Put([]byte(key), dataJSON)
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
					if vErr := json.Unmarshal(v, &notiContent); vErr != nil {
						return vErr
					}
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
		return nil
	})
	return result, err
}

func (self *BoltStorage) SetStableTokenParams(value []byte) error {
	var err error
	k := boltutil.Uint64ToBytes(1)
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
			return errors.New("Currently there is a pending record")
		}
		return b.Put(k, value)
	})
	return err
}

func (self *BoltStorage) ConfirmStableTokenParams(value []byte) error {
	var err error
	k := boltutil.Uint64ToBytes(1)
	temp := make(map[string]interface{})
	vErr := json.Unmarshal(value, &temp)
	if vErr != nil {
		return fmt.Errorf("Rejected: Data could not be unmarshalled to defined format: %s", vErr)
	}
	pending, err := self.GetPendingStableTokenParams()
	if eq := reflect.DeepEqual(pending, temp); !eq {
		return errors.New("Rejected: confiming data isn't consistent")
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
	k := boltutil.Uint64ToBytes(1)
	result := make(map[string]interface{})
	err := self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(STABLE_TOKEN_PARAMS_BUCKET))
		if b == nil {
			return errors.New("Bucket hasn't exist yet")
		}
		record := b.Get(k)
		if record != nil {
			return json.Unmarshal(record, &result)
		}
		return nil
	})
	return result, err
}

func (self *BoltStorage) GetPendingStableTokenParams() (map[string]interface{}, error) {
	k := boltutil.Uint64ToBytes(1)
	result := make(map[string]interface{})
	err := self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(PENDING_STABLE_TOKEN_PARAMS_BUCKET))
		if b == nil {
			return errors.New("Bucket hasn't exist yet")
		}
		record := b.Get(k)
		if record != nil {
			return json.Unmarshal(record, &result)
		}
		return nil
	})
	return result, err
}

func (self *BoltStorage) RemovePendingStableTokenParams() error {
	k := boltutil.Uint64ToBytes(1)
	err := self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(PENDING_STABLE_TOKEN_PARAMS_BUCKET))
		if b == nil {
			return errors.New("Bucket hasn't existed yet")
		}
		record := b.Get(k)
		if record == nil {
			return errors.New("Bucket is empty")
		}
		return b.Delete(k)
	})
	return err
}

func (self *BoltStorage) StorePendingTargetQtyV2(value []byte) error {
	temp := make(map[string]interface{})
	vErr := json.Unmarshal(value, &temp)
	if vErr != nil {
		return fmt.Errorf("Rejected: Data could not be unmarshalled to defined format: %s", vErr.Error())
	}
	err := self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(PENDING_TARGET_QUANTITY_V2))
		k := []byte("current_pending_target_qty")
		if b.Get(k) != nil {
			return fmt.Errorf("Currently there is a pending record")
		}
		return b.Put(k, value)
	})
	return err
}

func (self *BoltStorage) GetPendingTargetQtyV2() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	err := self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(PENDING_TARGET_QUANTITY_V2))
		k := []byte("current_pending_target_qty")
		record := b.Get(k)
		if record == nil {
			return errors.New("There is no pending target qty")
		}
		return json.Unmarshal(record, &result)
	})
	return result, err
}

func (self *BoltStorage) ConfirmTargetQtyV2(value []byte) error {
	temp := make(map[string]interface{})
	err := json.Unmarshal(value, &temp)
	if err != nil {
		return fmt.Errorf("Rejected: Data could not be unmarshalled to defined format: %s", err)
	}
	pending, err := self.GetPendingTargetQtyV2()
	if eq := reflect.DeepEqual(pending, temp); !eq {
		return fmt.Errorf("Rejected: confiming data isn't consistent")
	}

	err = self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TARGET_QUANTITY_V2))
		targetKey := []byte("current_target_qty")
		if vErr := b.Put(targetKey, value); vErr != nil {
			return vErr
		}
		pendingBk := tx.Bucket([]byte(PENDING_TARGET_QUANTITY_V2))
		pendingKey := []byte("current_pending_target_qty")
		return pendingBk.Delete(pendingKey)
	})
	return err
}

// RemovePendingTargetQtyV2 remove pending data from db
func (self *BoltStorage) RemovePendingTargetQtyV2() error {
	err := self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(PENDING_TARGET_QUANTITY_V2))
		if b == nil {
			return fmt.Errorf("Bucket hasn't existed yet")
		}
		k := []byte("current_pending_target_qty")
		return b.Delete(k)
	})
	return err
}

// GetTargetQtyV2 return the current target quantity
func (self *BoltStorage) GetTargetQtyV2() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	err := self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TARGET_QUANTITY_V2))
		k := []byte("current_target_qty")
		record := b.Get(k)
		if record == nil {
			return nil
		}
		return json.Unmarshal(record, &result)
	})

	// This block below is for backward compatible for api v1
	// when the result is empty it means there is not target quantity is set
	// we need to get current target quantity from v1 bucket and return it as v2 form.
	if len(result) == 0 {
		// target qty v1
		targetQty, vErr := self.GetTokenTargetQty()
		if vErr != nil {
			return result, vErr
		}
		result = convertTargetQtyV1toV2(targetQty)
	}
	return result, err
}

// This function convert target quantity from v1 to v2
// TokenTargetQty v1 should be follow this format:
// token_totalTarget_reserveTarget_rebalanceThreshold_transferThreshold|token_totalTarget_reserveTarget_rebalanceThreshold_transferThreshold|...
// while token is a string, it is validated before it saved then no need to validate again here
// totalTarget, reserveTarget, rebalanceThreshold and transferThreshold are float numbers
// and they are also no need to check to error here also (so we can ignore as below)
func convertTargetQtyV1toV2(target metric.TokenTargetQty) map[string]interface{} {
	result := make(map[string]interface{})
	strTargets := strings.Split(target.Data, "|")
	for _, target := range strTargets {
		elements := strings.Split(target, "_")
		if len(elements) != 5 {
			continue
		}
		token := elements[0]
		totalTarget, _ := strconv.ParseFloat(elements[1], 10)
		reserveTarget, _ := strconv.ParseFloat(elements[2], 10)
		rebalance, _ := strconv.ParseFloat(elements[3], 10)
		withdraw, _ := strconv.ParseFloat(elements[4], 10)
		result[token] = metric.TargetQtyStruct{
			SetTarget: metric.TargetQtySet{
				TotalTarget:        totalTarget,
				ReserveTarget:      reserveTarget,
				RebalanceThreshold: rebalance,
				TransferThreshold:  withdraw,
			},
		}
	}
	return result
}

// StorePendingPWIEquationV2 stores the given PWIs equation data for later approval.
func (self *BoltStorage) StorePendingPWIEquationV2(data []byte) error {
	timepoint := common.GetTimepoint()
	err := self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(PENDING_PWI_EQUATION_V2))
		c := b.Cursor()
		_, v := c.First()
		if v != nil {
			return errors.New("pending PWI equation exists")
		}
		return b.Put(boltutil.Uint64ToBytes(timepoint), data)
	})
	return err
}

// GetPendingPWIEquationV2 returns the stored PWIEquationRequestV2 in database.
func (self *BoltStorage) GetPendingPWIEquationV2() (metric.PWIEquationRequestV2, error) {
	var (
		err    error
		result metric.PWIEquationRequestV2
	)

	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(PENDING_PWI_EQUATION_V2))
		c := b.Cursor()
		_, v := c.First()
		if v == nil {
			return errors.New("There is no pending equation")
		}
		return json.Unmarshal(v, &result)
	})
	return result, err
}

// RemovePendingPWIEquationV2 deletes the pending equation request.
func (self *BoltStorage) RemovePendingPWIEquationV2() error {
	err := self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(PENDING_PWI_EQUATION_V2))
		c := b.Cursor()
		k, _ := c.First()
		if k == nil {
			return errors.New("There is no pending data")
		}
		return b.Delete(k)
	})
	return err
}

// StorePWIEquationV2 moved the pending equation request to
// PWI_EQUATION_V2 bucket and remove it from pending bucket if the
// given data matched what stored.
func (self *BoltStorage) StorePWIEquationV2(data string) error {
	err := self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(PENDING_PWI_EQUATION_V2))
		c := b.Cursor()
		k, v := c.First()
		if v == nil {
			return errors.New("There is no pending equation")
		}
		confirmData := make(map[string]interface{})
		if err := json.Unmarshal([]byte(data), &confirmData); err != nil {
			return err
		}
		currentData := make(map[string]interface{})
		if err := json.Unmarshal(v, &currentData); err != nil {
			return err
		}
		if eq := reflect.DeepEqual(currentData, confirmData); !eq {
			return errors.New("Confirm data does not match pending data")
		}
		id := boltutil.Uint64ToBytes(common.GetTimepoint())
		if uErr := tx.Bucket([]byte(PWI_EQUATION_V2)).Put(id, v); uErr != nil {
			return uErr
		}
		// remove pending PWI equations request
		return b.Delete(k)
	})
	return err
}

func convertPWIEquationV1toV2(data string) (metric.PWIEquationRequestV2, error) {
	result := metric.PWIEquationRequestV2{}
	for _, dataConfig := range strings.Split(data, "|") {
		dataParts := strings.Split(dataConfig, "_")
		if len(dataParts) != 4 {
			return nil, errors.New("malform data")
		}

		a, err := strconv.ParseFloat(dataParts[1], 64)
		if err != nil {
			return nil, err
		}
		b, err := strconv.ParseFloat(dataParts[2], 64)
		if err != nil {
			return nil, err
		}
		c, err := strconv.ParseFloat(dataParts[3], 64)
		if err != nil {
			return nil, err
		}
		eq := metric.PWIEquationV2{
			A: a,
			B: b,
			C: c,
		}
		result[dataParts[0]] = metric.PWIEquationTokenV2{
			"bid": eq,
			"ask": eq,
		}
	}
	return result, nil
}

func pwiEquationV1toV2(tx *bolt.Tx) (metric.PWIEquationRequestV2, error) {
	var eqv1 metric.PWIEquation
	b := tx.Bucket([]byte(PWI_EQUATION))
	c := b.Cursor()
	_, v := c.Last()
	if v == nil {
		return nil, errors.New("There is no equation")
	}
	if err := json.Unmarshal(v, &eqv1); err != nil {
		return nil, err
	}
	return convertPWIEquationV1toV2(eqv1.Data)
}

// GetPWIEquationV2 returns the current PWI equations from database.
func (self *BoltStorage) GetPWIEquationV2() (metric.PWIEquationRequestV2, error) {
	var (
		err    error
		result metric.PWIEquationRequestV2
	)
	err = self.db.View(func(tx *bolt.Tx) error {
		var vErr error
		b := tx.Bucket([]byte(PWI_EQUATION_V2))
		c := b.Cursor()
		_, v := c.Last()
		if v == nil {
			log.Println("there no equation in PWI_EQUATION_V2, getting from PWI_EQUATION")
			result, vErr = pwiEquationV1toV2(tx)
			return vErr
		}
		return json.Unmarshal(v, &result)
	})
	return result, err
}
