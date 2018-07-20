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
	"github.com/boltdb/bolt"
)

const (
	priceBucket                     string = "prices"
	rateBucket                      string = "rates"
	activityBucket                  string = "activities"
	authDataBucket                  string = "auth_data"
	pendingActivityBucket           string = "pending_activities"
	metricBucket                    string = "metrics"
	metricTargetQuantity            string = "target_quantity"
	pendingTargetQuantity           string = "pending_target_quantity"
	enableRebalance                 string = "enable_rebalance"
	setrateControl                  string = "setrate_control"
	pendingPWIEquation              string = "pending_pwi_equation"
	pwiEquation                     string = "pwi_equation"
	exchangeStatus                  string = "exchange_status"
	exchangeNotifications           string = "exchange_notifications"
	maxNumberVersion                int    = 1000
	maxGetRatesPeriod               uint64 = 86400000      //1 days in milisec
	authDataExpiredDuration         uint64 = 10 * 86400000 //10day in milisec
	stableTokenParamsBucket         string = "stable-token-params"
	pendingStatbleTokenParamsBucket string = "pending-stable-token-params"
	goldBucket                      string = "gold_feeds"

	// pendingTargetQuantityV2 constant for bucket name for pending target quantity v2
	pendingTargetQuantityV2 string = "pending_target_qty_v2"
	// targetQuantityV2 constant for bucet name for target quantity v2
	targetQuantityV2 string = "target_quantity_v2"

	// pendingPWIEquationV2 is the bucket name for storing pending
	// pwi equation for later approval.
	pendingPWIEquationV2 string = "pending_pwi_equation_v2"
	// pwiEquationV2 stores the PWI equations after confirmed.
	pwiEquationV2 string = "pwi_equation_v2"

	// pendingRebalanceQuadratic stores pending rebalance quadratic equation
	pendingRebalanceQuadratic = "pending_rebalance_quadratic"
	// rebalanceQuadratic stores rebalance quadratic equation
	rebalanceQuadratic = "rebalance_quadratic"
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
		if _, cErr := tx.CreateBucketIfNotExists([]byte(goldBucket)); cErr != nil {
			return cErr
		}

		if _, cErr := tx.CreateBucketIfNotExists([]byte(priceBucket)); cErr != nil {
			return cErr
		}
		if _, cErr := tx.CreateBucketIfNotExists([]byte(rateBucket)); cErr != nil {
			return cErr
		}
		if _, cErr := tx.CreateBucketIfNotExists([]byte(activityBucket)); cErr != nil {
			return cErr
		}
		if _, cErr := tx.CreateBucketIfNotExists([]byte(pendingActivityBucket)); cErr != nil {
			return cErr
		}
		if _, cErr := tx.CreateBucketIfNotExists([]byte(authDataBucket)); cErr != nil {
			return cErr
		}
		if _, cErr := tx.CreateBucketIfNotExists([]byte(metricBucket)); cErr != nil {
			return cErr
		}
		if _, cErr := tx.CreateBucketIfNotExists([]byte(metricTargetQuantity)); cErr != nil {
			return cErr
		}
		if _, cErr := tx.CreateBucketIfNotExists([]byte(pendingTargetQuantity)); cErr != nil {
			return cErr
		}
		if _, cErr := tx.CreateBucketIfNotExists([]byte(enableRebalance)); cErr != nil {
			return cErr
		}
		if _, cErr := tx.CreateBucketIfNotExists([]byte(setrateControl)); cErr != nil {
			return cErr
		}
		if _, cErr := tx.CreateBucketIfNotExists([]byte(pendingPWIEquation)); cErr != nil {
			return cErr
		}
		if _, cErr := tx.CreateBucketIfNotExists([]byte(pwiEquation)); cErr != nil {
			return cErr
		}
		if _, cErr := tx.CreateBucketIfNotExists([]byte(exchangeStatus)); cErr != nil {
			return cErr
		}
		if _, cErr := tx.CreateBucketIfNotExists([]byte(exchangeNotifications)); cErr != nil {
			return cErr
		}
		if _, cErr := tx.CreateBucketIfNotExists([]byte(pendingStatbleTokenParamsBucket)); cErr != nil {
			return cErr
		}
		if _, cErr := tx.CreateBucketIfNotExists([]byte(stableTokenParamsBucket)); cErr != nil {
			return cErr
		}

		if _, cErr := tx.CreateBucketIfNotExists([]byte(pendingTargetQuantityV2)); cErr != nil {
			return cErr
		}

		if _, cErr := tx.CreateBucketIfNotExists([]byte(targetQuantityV2)); cErr != nil {
			return cErr
		}

		if _, cErr := tx.CreateBucketIfNotExists([]byte(pendingPWIEquationV2)); cErr != nil {
			return cErr
		}
		if _, cErr := tx.CreateBucketIfNotExists([]byte(pwiEquationV2)); cErr != nil {
			return cErr
		}

		if _, cErr := tx.CreateBucketIfNotExists([]byte(pendingRebalanceQuadratic)); cErr != nil {
			return cErr
		}
		if _, cErr := tx.CreateBucketIfNotExists([]byte(rebalanceQuadratic)); cErr != nil {
			return cErr
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
		c := tx.Bucket([]byte(goldBucket)).Cursor()
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
		b := tx.Bucket([]byte(goldBucket))
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
		b := tx.Bucket([]byte(goldBucket))
		dataJSON, uErr := json.Marshal(data)
		if uErr != nil {
			return uErr
		}
		return b.Put(boltutil.Uint64ToBytes(timepoint), dataJSON)
	})
	return err
}

func (self *BoltStorage) ExportExpiredAuthData(currentTime uint64, fileName string) (nRecord uint64, err error) {
	expiredTimestampByte := boltutil.Uint64ToBytes(currentTime - authDataExpiredDuration)
	outFile, err := os.Create(fileName)
	if err != nil {
		return 0, err
	}
	defer func() {
		if cErr := outFile.Close(); cErr != nil {
			log.Printf("Close file error: %s", cErr.Error())
		}
	}()

	err = self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(authDataBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil && bytes.Compare(k, expiredTimestampByte) <= 0; k, v = c.Next() {
			timestamp := boltutil.BytesToUint64(k)

			temp := common.AuthDataSnapshot{}
			if uErr := json.Unmarshal(v, &temp); uErr != nil {
				return uErr
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
	expiredTimestampByte := boltutil.Uint64ToBytes(currentTime - authDataExpiredDuration)

	err = self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(authDataBucket))
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
	nExcess := self.GetNumberOfVersion(tx, bucket) - maxNumberVersion
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
		c := tx.Bucket([]byte(priceBucket)).Cursor()
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
		b := tx.Bucket([]byte(priceBucket))
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
		b := tx.Bucket([]byte(priceBucket))
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
		var (
			uErr     error
			dataJSON []byte
		)

		b := tx.Bucket([]byte(priceBucket))

		// remove outdated data from bucket
		log.Printf("Version number: %d\n", self.GetNumberOfVersion(tx, priceBucket))
		if uErr = self.PruneOutdatedData(tx, priceBucket); uErr != nil {
			log.Printf("Prune out data: %s", uErr.Error())
			return uErr
		}

		if dataJSON, uErr = json.Marshal(data); uErr != nil {
			return uErr
		}
		return b.Put(boltutil.Uint64ToBytes(timepoint), dataJSON)
	})
	return err
}

func (self *BoltStorage) CurrentAuthDataVersion(timepoint uint64) (common.Version, error) {
	var result uint64
	var err error
	err = self.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(authDataBucket)).Cursor()
		result, err = reverseSeek(timepoint, c)
		return err
	})
	return common.Version(result), err
}

func (self *BoltStorage) GetAuthData(version common.Version) (common.AuthDataSnapshot, error) {
	result := common.AuthDataSnapshot{}
	var err error
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(authDataBucket))
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
		c := tx.Bucket([]byte(rateBucket)).Cursor()
		result, err = reverseSeek(timepoint, c)
		return err
	})
	return common.Version(result), err
}

//GetRates return rates history
func (self *BoltStorage) GetRates(fromTime, toTime uint64) ([]common.AllRateEntry, error) {
	result := []common.AllRateEntry{}
	if toTime-fromTime > maxGetRatesPeriod {
		return result, fmt.Errorf("Time range is too broad, it must be smaller or equal to %d miliseconds", maxGetRatesPeriod)
	}
	var err error
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(rateBucket))
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
		b := tx.Bucket([]byte(rateBucket))
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
		var (
			uErr     error
			dataJSON []byte
		)
		b := tx.Bucket([]byte(authDataBucket))

		if dataJSON, uErr = json.Marshal(data); uErr != nil {
			return uErr
		}
		return b.Put(boltutil.Uint64ToBytes(timepoint), dataJSON)
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
			dataJSON  []byte
		)

		b := tx.Bucket([]byte(rateBucket))
		c := b.Cursor()
		lastKey, lastValue := c.Last()
		if lastKey == nil {
			log.Printf("Bucket %s is empty", rateBucket)
		} else {
			if uErr = json.Unmarshal(lastValue, &lastEntry); uErr != nil {
				return uErr
			}
		}
		// we still update when blocknumber is not changed because we want
		// to update the version and timestamp so api users will get
		// the newest data even it is identical to the old one.
		if lastEntry.BlockNumber <= data.BlockNumber {
			if dataJSON, uErr = json.Marshal(data); uErr != nil {
				return uErr
			}
			return b.Put(boltutil.Uint64ToBytes(timepoint), dataJSON)
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
		var dataJSON []byte
		b := tx.Bucket([]byte(activityBucket))
		record := common.NewActivityRecord(
			action,
			id,
			destination,
			params,
			result,
			estatus,
			mstatus,
			common.Timestamp(strconv.FormatUint(timepoint, 10)),
		)
		dataJSON, err = json.Marshal(record)
		if err != nil {
			return err
		}

		idByte := id.ToBytes()
		err = b.Put(idByte[:], dataJSON)
		if err != nil {
			return err
		}
		if record.IsPending() {
			pb := tx.Bucket([]byte(pendingActivityBucket))
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
			err = pb.Put(idByte[:], dataJSON)
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
		b := tx.Bucket([]byte(activityBucket))
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
	if (toTime-fromTime)/1000000 > maxGetRatesPeriod {
		return result, fmt.Errorf("Time range is too broad, it must be smaller or equal to %d miliseconds", maxGetRatesPeriod)
	}
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(activityBucket))
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

// interfaceConverstionToUint64 will assert the interface as string
// and parse it to uint64. Return 0 if anything goes wrong)
func interfaceConverstionToUint64(intf interface{}) uint64 {
	numString, ok := intf.(string)
	if !ok {
		log.Printf("(%v) can't be converted to type string", intf)
		return 0
	}
	num, err := strconv.ParseUint(numString, 10, 64)
	if err != nil {
		log.Printf("ERROR: parsing error %s, inteface conversion to uint64 will set to 0", err)
		return 0
	}
	return num
}

func getLastAndCountPendingSetrate(pendings []common.ActivityRecord, minedNonce uint64) (*common.ActivityRecord, uint64, error) {
	var maxNonce uint64
	var maxPrice uint64
	var result *common.ActivityRecord
	var count uint64
	for i, act := range pendings {
		if act.Action == "set_rates" {
			log.Printf("looking for pending set_rates: %+v", act)
			nonce := interfaceConverstionToUint64(act.Result["nonce"])
			if nonce < minedNonce {
				// this is a stale actitivity, ignore it
				continue
			}
			gasPrice := interfaceConverstionToUint64(act.Result["gasPrice"])
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
	pb := tx.Bucket([]byte(pendingActivityBucket))
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
		b := tx.Bucket([]byte(pendingActivityBucket))
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
		pb := tx.Bucket([]byte(pendingActivityBucket))
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
		b := tx.Bucket([]byte(activityBucket))
		if uErr != nil {
			return uErr
		}
		return b.Put(idBytes[:], dataJSON)
	})
	return err
}

//HasPendingDeposit check if a deposit is pending
func (self *BoltStorage) HasPendingDeposit(token common.Token, exchange common.Exchange) (bool, error) {
	var (
		err    error
		result = false
	)
	err = self.db.View(func(tx *bolt.Tx) error {
		pb := tx.Bucket([]byte(pendingActivityBucket))
		c := pb.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			record := common.ActivityRecord{}
			if uErr := json.Unmarshal(v, &record); uErr != nil {
				return uErr
			}
			if record.Action == "deposit" {
				tokenID, ok := record.Params["token"].(string)
				if !ok {
					log.Printf("ERROR: record Params token (%v) can not be converted to string", record.Params["token"])
					continue
				}
				if tokenID == token.ID && record.Destination == string(exchange.ID()) {
					result = true
				}
			}
		}
		return nil
	})
	return result, err
}

//StoreMetric store metric info
func (self *BoltStorage) StoreMetric(data *common.MetricEntry, timepoint uint64) error {
	var err error
	err = self.db.Update(func(tx *bolt.Tx) error {
		var dataJSON []byte
		b := tx.Bucket([]byte(metricBucket))
		dataJSON, mErr := json.Marshal(data)
		if mErr != nil {
			return mErr
		}
		idByte := boltutil.Uint64ToBytes(data.Timestamp)
		err = b.Put(idByte, dataJSON)
		return err
	})
	return err
}

//GetMetric return metric data
func (self *BoltStorage) GetMetric(tokens []common.Token, fromTime, toTime uint64) (map[string]common.MetricList, error) {
	imResult := map[string]*common.MetricList{}
	for _, tok := range tokens {
		imResult[tok.ID] = &common.MetricList{}
	}

	var err error
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(metricBucket))
		c := b.Cursor()
		min := boltutil.Uint64ToBytes(fromTime)
		max := boltutil.Uint64ToBytes(toTime)

		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			data := common.MetricEntry{}
			err = json.Unmarshal(v, &data)
			if err != nil {
				return err
			}
			for tok, m := range data.Data {
				metricList, found := imResult[tok]
				if found {
					*metricList = append(*metricList, common.TokenMetricResponse{
						Timestamp: data.Timestamp,
						AfpMid:    m.AfpMid,
						Spread:    m.Spread,
					})
				}
			}
		}
		return nil
	})
	result := map[string]common.MetricList{}
	for k, v := range imResult {
		result[k] = *v
	}
	return result, err
}

//GetPendingTargetQty return pending target quantity if available
func (self *BoltStorage) GetPendingTargetQty() (common.TokenTargetQty, error) {
	var err error
	var tokenTargetQty common.TokenTargetQty
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(pendingTargetQuantity))
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
	tokenTargetQty := common.TokenTargetQty{}
	err = self.db.Update(func(tx *bolt.Tx) error {
		var (
			uErr     error
			dataJSON []byte
		)
		b := tx.Bucket([]byte(pendingTargetQuantity))
		_, lastPending := b.Cursor().Last()
		if lastPending != nil {
			return errors.New("there is another pending target quantity. please confirm or cancel it before setting new target")
		}

		tokenTargetQty.ID = timepoint
		tokenTargetQty.Status = "unconfirmed"
		tokenTargetQty.Data = data
		tokenTargetQty.Type, _ = strconv.ParseInt(dataType, 10, 64)
		idByte := boltutil.Uint64ToBytes(timepoint)

		if dataJSON, uErr = json.Marshal(tokenTargetQty); uErr != nil {
			return uErr
		}

		log.Printf("Target to save: %v", dataJSON)
		return b.Put(idByte, dataJSON)
	})
	return err
}

//RemovePendingTargetQty remove pending target quantity
func (self *BoltStorage) RemovePendingTargetQty() error {
	var err error
	err = self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(pendingTargetQuantity))
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
		c := tx.Bucket([]byte(metricTargetQuantity)).Cursor()
		result, err = reverseSeek(timepoint, c)
		return err
	})
	return common.Version(result), err
}

//GetTokenTargetQty get target quantity
func (self *BoltStorage) GetTokenTargetQty() (common.TokenTargetQty, error) {
	var (
		tokenTargetQty = common.TokenTargetQty{}
		err            error
	)
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(metricTargetQuantity))
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
	var (
		err            error
		tokenTargetQty common.TokenTargetQty
		dataJSON       []byte
	)
	err = self.db.Update(func(tx *bolt.Tx) error {
		var (
			uErr  error
			idInt uint64
		)
		pending := tx.Bucket([]byte(pendingTargetQuantity))
		_, pendingTargetQty := pending.Cursor().Last()

		if pendingTargetQty == nil {
			return errors.New("there is no pending target activity to confirm")
		}
		// verify confirm data
		if uErr = json.Unmarshal(pendingTargetQty, &tokenTargetQty); uErr != nil {
			return uErr
		}
		pendingData := tokenTargetQty.Data

		if idInt, uErr = strconv.ParseUint(id, 10, 64); uErr != nil {
			return uErr
		}
		if tokenTargetQty.ID != idInt {
			return errors.New("pending target quantity ID does not match")
		}
		if data != pendingData {
			return errors.New("pending target quantity data does not match")
		}

		// Save to confirmed target quantity
		tokenTargetQty.Status = "confirmed"
		b := tx.Bucket([]byte(metricTargetQuantity))
		dataJSON, uErr = json.Marshal(tokenTargetQty)
		if uErr != nil {
			return uErr
		}
		idByte := boltutil.Uint64ToBytes(common.GetTimepoint())
		return b.Put(idByte, dataJSON)
	})
	if err != nil {
		return err
	}
	// remove pending target qty
	return self.RemovePendingTargetQty()
}

func (self *BoltStorage) GetRebalanceControl() (common.RebalanceControl, error) {
	var err error
	var result common.RebalanceControl
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(enableRebalance))
		_, data := b.Cursor().First()
		if data == nil {
			result = common.RebalanceControl{
				Status: false,
			}
			return self.StoreRebalanceControl(false)
		}
		return json.Unmarshal(data, &result)
	})
	return result, err
}

func (self *BoltStorage) StoreRebalanceControl(status bool) error {
	var err error
	err = self.db.Update(func(tx *bolt.Tx) error {
		var (
			uErr     error
			dataJSON []byte
		)
		b := tx.Bucket([]byte(enableRebalance))
		// prune out old data
		c := b.Cursor()
		k, _ := c.First()
		if k != nil {
			if uErr = b.Delete([]byte(k)); uErr != nil {
				return uErr
			}
		}

		// add new data
		data := common.RebalanceControl{
			Status: status,
		}
		if dataJSON, uErr = json.Marshal(data); uErr != nil {
			return uErr
		}
		idByte := boltutil.Uint64ToBytes(common.GetTimepoint())
		return b.Put(idByte, dataJSON)
	})
	return err
}

func (self *BoltStorage) GetSetrateControl() (common.SetrateControl, error) {
	var err error
	var result common.SetrateControl
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(setrateControl))
		_, data := b.Cursor().First()
		if data == nil {
			result = common.SetrateControl{
				Status: false,
			}
			return self.StoreSetrateControl(false)
		}
		return json.Unmarshal(data, &result)
	})
	return result, err
}

func (self *BoltStorage) StoreSetrateControl(status bool) error {
	var err error
	err = self.db.Update(func(tx *bolt.Tx) error {
		var (
			uErr     error
			dataJSON []byte
		)
		b := tx.Bucket([]byte(setrateControl))
		// prune out old data
		c := b.Cursor()
		k, _ := c.First()
		if k != nil {
			if uErr = b.Delete([]byte(k)); uErr != nil {
				return uErr
			}
		}

		// add new data
		data := common.SetrateControl{
			Status: status,
		}

		if dataJSON, uErr = json.Marshal(data); uErr != nil {
			return uErr
		}
		idByte := boltutil.Uint64ToBytes(common.GetTimepoint())
		return b.Put(idByte, dataJSON)
	})
	return err
}

func (self *BoltStorage) StorePendingPWIEquation(data string) error {
	var err error
	timepoint := common.GetTimepoint()
	saveData := common.PWIEquation{}
	err = self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(pendingPWIEquation))
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

func (self *BoltStorage) GetPendingPWIEquation() (common.PWIEquation, error) {
	var err error
	var result common.PWIEquation
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(pendingPWIEquation))
		c := b.Cursor()
		_, v := c.First()
		if v == nil {
			return errors.New("There no pending equation")
		}
		return json.Unmarshal(v, &result)
	})
	return result, err
}

func (self *BoltStorage) StorePWIEquation(data string) error {
	var err error
	err = self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(pendingPWIEquation))
		c := b.Cursor()
		_, v := c.First()
		if v == nil {
			return errors.New("There no pending equation")
		}
		p := tx.Bucket([]byte(pwiEquation))
		idByte := boltutil.Uint64ToBytes(common.GetTimepoint())
		pending := common.PWIEquation{}
		if uErr := json.Unmarshal(v, &pending); uErr != nil {
			return uErr
		}
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

func (self *BoltStorage) GetPWIEquation() (common.PWIEquation, error) {
	var err error
	var result common.PWIEquation
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(pwiEquation))
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
		b := tx.Bucket([]byte(pendingPWIEquation))
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

func (self *BoltStorage) SetStableTokenParams(value []byte) error {
	var err error
	k := boltutil.Uint64ToBytes(1)
	temp := make(map[string]interface{})

	if err = json.Unmarshal(value, &temp); err != nil {
		return fmt.Errorf("Rejected: Data could not be unmarshalled to defined format: %s", err)
	}
	err = self.db.Update(func(tx *bolt.Tx) error {
		b, uErr := tx.CreateBucketIfNotExists([]byte(pendingStatbleTokenParamsBucket))
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

	if err = json.Unmarshal(value, &temp); err != nil {
		return fmt.Errorf("Rejected: Data could not be unmarshalled to defined format: %s", err)
	}
	pending, err := self.GetPendingStableTokenParams()
	if eq := reflect.DeepEqual(pending, temp); !eq {
		return errors.New("Rejected: confiming data isn't consistent")
	}

	err = self.db.Update(func(tx *bolt.Tx) error {
		b, uErr := tx.CreateBucketIfNotExists([]byte(stableTokenParamsBucket))
		if uErr != nil {
			return uErr
		}
		return b.Put(k, value)
	})
	if err != nil {
		return err
	}
	return self.RemovePendingStableTokenParams()
}

func (self *BoltStorage) GetStableTokenParams() (map[string]interface{}, error) {
	k := boltutil.Uint64ToBytes(1)
	result := make(map[string]interface{})
	err := self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(stableTokenParamsBucket))
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
		b := tx.Bucket([]byte(pendingStatbleTokenParamsBucket))
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
		b := tx.Bucket([]byte(pendingStatbleTokenParamsBucket))
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

//StorePendingTargetQtyV2 store value into pending target qty v2 bucket
func (self *BoltStorage) StorePendingTargetQtyV2(value []byte) error {
	var (
		err         error
		pendingData common.TokenTargetQtyV2
	)

	if err = json.Unmarshal(value, &pendingData); err != nil {
		return fmt.Errorf("Rejected: Data could not be unmarshalled to defined format: %s", err.Error())
	}
	err = self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(pendingTargetQuantityV2))
		k := []byte("current_pending_target_qty")
		if b.Get(k) != nil {
			return fmt.Errorf("Currently there is a pending record")
		}
		return b.Put(k, value)
	})
	return err
}

//GetPendingTargetQtyV2 return current pending target quantity
func (self *BoltStorage) GetPendingTargetQtyV2() (common.TokenTargetQtyV2, error) {
	result := common.TokenTargetQtyV2{}
	err := self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(pendingTargetQuantityV2))
		k := []byte("current_pending_target_qty")
		record := b.Get(k)
		if record == nil {
			return errors.New("There is no pending target qty")
		}
		return json.Unmarshal(record, &result)
	})
	return result, err
}

//ConfirmTargetQtyV2 check if confirm data match pending data and save it to confirm bucket
//remove pending data from pending bucket
func (self *BoltStorage) ConfirmTargetQtyV2(value []byte) error {
	confirmTargetQty := common.TokenTargetQtyV2{}
	err := json.Unmarshal(value, &confirmTargetQty)
	if err != nil {
		return fmt.Errorf("Rejected: Data could not be unmarshalled to defined format: %s", err)
	}
	pending, err := self.GetPendingTargetQtyV2()
	if eq := reflect.DeepEqual(pending, confirmTargetQty); !eq {
		return fmt.Errorf("Rejected: confiming data isn't consistent")
	}

	err = self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(targetQuantityV2))
		targetKey := []byte("current_target_qty")
		if uErr := b.Put(targetKey, value); uErr != nil {
			return uErr
		}
		pendingBk := tx.Bucket([]byte(pendingTargetQuantityV2))
		pendingKey := []byte("current_pending_target_qty")
		return pendingBk.Delete(pendingKey)
	})
	return err
}

// RemovePendingTargetQtyV2 remove pending data from db
func (self *BoltStorage) RemovePendingTargetQtyV2() error {
	err := self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(pendingTargetQuantityV2))
		if b == nil {
			return fmt.Errorf("Bucket hasn't existed yet")
		}
		k := []byte("current_pending_target_qty")
		return b.Delete(k)
	})
	return err
}

// GetTargetQtyV2 return the current target quantity
func (self *BoltStorage) GetTargetQtyV2() (common.TokenTargetQtyV2, error) {
	result := common.TokenTargetQtyV2{}
	err := self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(targetQuantityV2))
		k := []byte("current_target_qty")
		record := b.Get(k)
		if record == nil {
			return nil
		}
		return json.Unmarshal(record, &result)
	})
	if err != nil {
		return result, err
	}

	// This block below is for backward compatible for api v1
	// when the result is empty it means there is not target quantity is set
	// we need to get current target quantity from v1 bucket and return it as v2 form.
	if len(result) == 0 {
		// target qty v1
		var targetQty common.TokenTargetQty
		targetQty, err = self.GetTokenTargetQty()
		if err != nil {
			return result, err
		}
		result = convertTargetQtyV1toV2(targetQty)
	}
	return result, nil
}

// This function convert target quantity from v1 to v2
// TokenTargetQty v1 should be follow this format:
// token_totalTarget_reserveTarget_rebalanceThreshold_transferThreshold|token_totalTarget_reserveTarget_rebalanceThreshold_transferThreshold|...
// while token is a string, it is validated before it saved then no need to validate again here
// totalTarget, reserveTarget, rebalanceThreshold and transferThreshold are float numbers
// and they are also no need to check to error here also (so we can ignore as below)
func convertTargetQtyV1toV2(target common.TokenTargetQty) common.TokenTargetQtyV2 {
	result := common.TokenTargetQtyV2{}
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
		result[token] = common.TargetQtyV2{
			SetTarget: common.TargetQtySet{
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
// Return error if occur or there is no pending PWIEquation
func (self *BoltStorage) StorePendingPWIEquationV2(data []byte) error {
	timepoint := common.GetTimepoint()
	err := self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(pendingPWIEquationV2))
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
func (self *BoltStorage) GetPendingPWIEquationV2() (common.PWIEquationRequestV2, error) {
	var (
		err    error
		result common.PWIEquationRequestV2
	)

	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(pendingPWIEquationV2))
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
		b := tx.Bucket([]byte(pendingPWIEquationV2))
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
// pwiEquationV2 bucket and remove it from pending bucket if the
// given data matched what stored.
func (self *BoltStorage) StorePWIEquationV2(data string) error {
	err := self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(pendingPWIEquationV2))
		c := b.Cursor()
		k, v := c.First()
		if v == nil {
			return errors.New("There is no pending equation")
		}
		confirmData := common.PWIEquationRequestV2{}
		if err := json.Unmarshal([]byte(data), &confirmData); err != nil {
			return err
		}
		currentData := common.PWIEquationRequestV2{}
		if err := json.Unmarshal(v, &currentData); err != nil {
			return err
		}
		if eq := reflect.DeepEqual(currentData, confirmData); !eq {
			return errors.New("Confirm data does not match pending data")
		}
		id := boltutil.Uint64ToBytes(common.GetTimepoint())
		if uErr := tx.Bucket([]byte(pwiEquationV2)).Put(id, v); uErr != nil {
			return uErr
		}
		// remove pending PWI equations request
		return b.Delete(k)
	})
	return err
}

func convertPWIEquationV1toV2(data string) (common.PWIEquationRequestV2, error) {
	result := common.PWIEquationRequestV2{}
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
		eq := common.PWIEquationV2{
			A: a,
			B: b,
			C: c,
		}
		result[dataParts[0]] = common.PWIEquationTokenV2{
			"bid": eq,
			"ask": eq,
		}
	}
	return result, nil
}

func pwiEquationV1toV2(tx *bolt.Tx) (common.PWIEquationRequestV2, error) {
	var eqv1 common.PWIEquation
	b := tx.Bucket([]byte(pwiEquation))
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
func (self *BoltStorage) GetPWIEquationV2() (common.PWIEquationRequestV2, error) {
	var (
		err    error
		result common.PWIEquationRequestV2
	)
	err = self.db.View(func(tx *bolt.Tx) error {
		var vErr error // convert pwi v1 to v2 error
		b := tx.Bucket([]byte(pwiEquationV2))
		c := b.Cursor()
		_, v := c.Last()
		if v == nil {
			log.Println("there no equation in pwiEquationV2, getting from pwiEquation")
			result, vErr = pwiEquationV1toV2(tx)
			return vErr
		}
		return json.Unmarshal(v, &result)
	})
	return result, err
}

//StorePendingRebalanceQuadratic store pending data (stand for rebalance quadratic equation) to db
//data byte for json {"KNC": {"a": 0.9, "b": 1.2, "c": 1.4}}
func (self *BoltStorage) StorePendingRebalanceQuadratic(data []byte) error {
	timepoint := common.GetTimepoint()
	err := self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(pendingRebalanceQuadratic))
		c := b.Cursor()
		k, _ := c.First()
		if k != nil {
			return errors.New("pending rebalance quadratic equation exists")
		}
		return b.Put(boltutil.Uint64ToBytes(timepoint), data)
	})
	return err
}

//GetPendingRebalanceQuadratic return pending rebalance quadratic equation
//Return err if occur, or if the DB is empty
func (self *BoltStorage) GetPendingRebalanceQuadratic() (common.RebalanceQuadraticRequest, error) {
	var result common.RebalanceQuadraticRequest
	err := self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(pendingRebalanceQuadratic))
		c := b.Cursor()
		k, v := c.First()
		if k == nil {
			return errors.New("there is no pending rebalance quadratic equation")
		}
		return json.Unmarshal(v, &result)
	})
	return result, err
}

//ConfirmRebalanceQuadratic confirm pending equation save it to confirmed bucket
//and remove pending equation
func (self *BoltStorage) ConfirmRebalanceQuadratic(data []byte) error {
	err := self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(pendingRebalanceQuadratic))
		c := b.Cursor()
		k, v := c.First()
		if v == nil {
			return errors.New("there is no pending rebalance quadratic equation")
		}
		confirmData := common.RebalanceQuadraticRequest{}
		if err := json.Unmarshal(data, &confirmData); err != nil {
			return err
		}
		currentData := common.RebalanceQuadraticRequest{}
		if err := json.Unmarshal(v, &currentData); err != nil {
			return err
		}
		if eq := reflect.DeepEqual(currentData, confirmData); !eq {
			return errors.New("confirm data does not match rebalance quadratic pending data")
		}
		id := boltutil.Uint64ToBytes(common.GetTimepoint())
		if uErr := tx.Bucket([]byte(rebalanceQuadratic)).Put(id, v); uErr != nil {
			return uErr
		}
		// remove pending rebalance quadratic equation
		return b.Delete(k)
	})
	return err
}

//RemovePendingRebalanceQuadratic remove pending rebalance quadratic equation
//use when admin want to reject a config for rebalance quadratic equation
func (self *BoltStorage) RemovePendingRebalanceQuadratic() error {
	err := self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(pendingRebalanceQuadratic))
		c := b.Cursor()
		k, _ := c.First()
		if k == nil {
			return errors.New("there no pending rebalance quadratic equation to delete")
		}
		return b.Delete(k)
	})
	return err
}

//GetRebalanceQuadratic return current confirm rebalance quadratic equation
func (self *BoltStorage) GetRebalanceQuadratic() (common.RebalanceQuadraticRequest, error) {
	var result common.RebalanceQuadraticRequest
	err := self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(rebalanceQuadratic))
		c := b.Cursor()
		k, v := c.Last()
		if k == nil {
			return errors.New("there is no rebalance quadratic equation")
		}
		return json.Unmarshal(v, &result)
	})
	return result, err
}

//storeJSONByteArrayToPendingBucket store key-value into a bucket under one TX.
//This function is use to store data to pendingbucket, it only update if the bucket is empty
func (self *BoltStorage) storeJSONByteArrayToPendingBucket(tx *bolt.Tx, bucketName string, key, value []byte) error {
	b := tx.Bucket([]byte(bucketName))
	if b == nil {
		return fmt.Errorf("Bucket %s hasn't existed yet", bucketName)
	}
	c := b.Cursor()
	_, v := c.First()
	if v != nil {
		return fmt.Errorf("Bucket %s has a pending record", bucketName)
	}
	return b.Put(key, value)
}

//storeJSONByteArray store key-value into a bucket under one TX.
func (self *BoltStorage) storeJSONByteArray(tx *bolt.Tx, bucketName string, key, value []byte) error {
	b := tx.Bucket([]byte(bucketName))
	if b == nil {
		return fmt.Errorf("Bucket %s hasn't existed yet", bucketName)
	}
	return b.Put(key, value)
}

func (self *BoltStorage) StorePendingTokenUpdateInfo(tarQty common.TokenTargetQtyV2, pwi common.PWIEquationRequestV2, quadEq common.RebalanceQuadraticRequest) error {
	timeStampKey := boltutil.Uint64ToBytes(common.GetTimepoint())
	err := self.db.Update(func(tx *bolt.Tx) error {
		dataJSON, uErr := json.Marshal(tarQty)
		if uErr != nil {
			return uErr
		}
		if uErr = self.storeJSONByteArrayToPendingBucket(tx, pendingTargetQuantityV2, []byte("current_pending_target_qty"), dataJSON); uErr != nil {
			return uErr
		}
		if dataJSON, uErr = json.Marshal(pwi); uErr != nil {
			return uErr
		}
		if uErr = self.storeJSONByteArrayToPendingBucket(tx, pendingPWIEquationV2, timeStampKey, dataJSON); uErr != nil {
			return uErr
		}
		if dataJSON, uErr = json.Marshal(quadEq); uErr != nil {
			return uErr
		}
		return self.storeJSONByteArrayToPendingBucket(tx, pendingRebalanceQuadratic, timeStampKey, dataJSON)
	})
	return err
}

func (self *BoltStorage) deleteTheOnlyObjectFromBucket(tx *bolt.Tx, bucketName string) error {
	b := tx.Bucket([]byte(bucketName))
	if b == nil {
		return fmt.Errorf("Bucket %s hasn't existed yet", bucketName)
	}
	c := b.Cursor()
	k, _ := c.First()
	if k == nil {
		return fmt.Errorf("Bucket %s is empty", bucketName)
	}
	return b.Delete(k)
}

func (self *BoltStorage) RemovePendingTokenUpdateInfo() error {
	err := self.db.Update(func(tx *bolt.Tx) error {
		if uErr := self.deleteTheOnlyObjectFromBucket(tx, pendingTargetQuantityV2); uErr != nil {
			return uErr
		}
		if uErr := self.deleteTheOnlyObjectFromBucket(tx, pendingPWIEquationV2); uErr != nil {
			return uErr
		}
		if uErr := self.deleteTheOnlyObjectFromBucket(tx, pendingRebalanceQuadratic); uErr != nil {
			return uErr
		}
		return nil
	})
	return err
}

func (self *BoltStorage) ConfirmTokenUpdateInfo(tarQty common.TokenTargetQtyV2, pwi common.PWIEquationRequestV2, quadEq common.RebalanceQuadraticRequest) error {
	timeStampKey := boltutil.Uint64ToBytes(common.GetTimepoint())
	err := self.db.Update(func(tx *bolt.Tx) error {
		dataJSON, uErr := json.Marshal(tarQty)
		if uErr != nil {
			return uErr
		}
		if uErr = self.storeJSONByteArray(tx, targetQuantityV2, []byte("current_target_qty"), dataJSON); uErr != nil {
			return uErr
		}
		if dataJSON, uErr = json.Marshal(pwi); uErr != nil {
			return uErr
		}
		if uErr = self.storeJSONByteArray(tx, pwiEquationV2, timeStampKey, dataJSON); uErr != nil {
			return uErr
		}
		if dataJSON, uErr = json.Marshal(quadEq); uErr != nil {
			return uErr
		}
		return self.storeJSONByteArray(tx, rebalanceQuadratic, timeStampKey, dataJSON)
	})
	return err
}
