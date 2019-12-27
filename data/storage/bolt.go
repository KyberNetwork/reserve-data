package storage

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"sync"

	"github.com/boltdb/bolt"

	"github.com/KyberNetwork/reserve-data/boltutil"
	"github.com/KyberNetwork/reserve-data/common"
	commonv3 "github.com/KyberNetwork/reserve-data/reservesetting/common"
)

const (
	priceBucket                     string = "prices"
	rateBucket                      string = "rates"
	activityBucket                  string = "activities"
	authDataBucket                  string = "auth_data"
	pendingActivityBucket           string = "pending_activities"
	enableRebalance                 string = "enable_rebalance"
	setrateControl                  string = "setrate_control"
	maxNumberVersion                int    = 1000
	maxGetRatesPeriod               uint64 = 86400000      //1 days in milisec
	authDataExpiredDuration         uint64 = 10 * 86400000 //10day in milisec
	stableTokenParamsBucket         string = "stable-token-params"
	pendingStatbleTokenParamsBucket string = "pending-stable-token-params"
	goldBucket                      string = "gold_feeds"
	btcBucket                       string = "btc_feeds"
	usdBucket                       string = "usd_feeds"
	disabledFeedsBucket             string = "disabled_feeds"

	//btcFetcherConfiguration stores configuration for btc fetcher
	fetcherConfigurationBucket = "btc_fetcher_configuration"
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
		buckets := []string{
			goldBucket,
			btcBucket,
			disabledFeedsBucket,
			priceBucket,
			rateBucket,
			activityBucket,
			pendingActivityBucket,
			authDataBucket,
			enableRebalance,
			setrateControl,
			pendingStatbleTokenParamsBucket,
			stableTokenParamsBucket,
			fetcherConfigurationBucket,
		}

		for _, bucket := range buckets {
			if _, cErr := tx.CreateBucketIfNotExists([]byte(bucket)); cErr != nil {
				return cErr
			}
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
			return 0, fmt.Errorf("there is no data before timepoint %d", timepoint)
		}
		return boltutil.BytesToUint64(version), nil
	}
	v := boltutil.BytesToUint64(version)
	if v == timepoint {
		return v, nil
	}
	version, _ = c.Prev()
	if version == nil {
		return 0, fmt.Errorf("there is no data before timepoint %d", timepoint)
	}
	return boltutil.BytesToUint64(version), nil
}

// CurrentGoldInfoVersion returns the most recent time point of gold info record.
// It implements data.GlobalStorage interface.
func (bs *BoltStorage) CurrentGoldInfoVersion(timepoint uint64) (common.Version, error) {
	var result uint64
	var err error
	err = bs.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(goldBucket)).Cursor()
		result, err = reverseSeek(timepoint, c)
		return nil
	})
	return common.Version(result), err
}

// GetGoldInfo returns gold info at given time point. It implements data.GlobalStorage interface.
func (bs *BoltStorage) GetGoldInfo(version common.Version) (common.GoldData, error) {
	result := common.GoldData{}
	var err error
	err = bs.db.View(func(tx *bolt.Tx) error {
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
func (bs *BoltStorage) StoreGoldInfo(data common.GoldData) error {
	var err error
	timepoint := data.Timestamp
	err = bs.db.Update(func(tx *bolt.Tx) error {
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

// CurrentBTCInfoVersion returns the most recent time point of gold info record.
// It implements data.GlobalStorage interface.
func (bs *BoltStorage) CurrentBTCInfoVersion(timepoint uint64) (common.Version, error) {
	var result uint64
	var err error
	err = bs.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(btcBucket)).Cursor()
		result, err = reverseSeek(timepoint, c)
		return nil
	})
	return common.Version(result), err
}

// CurrentUSDInfoVersion returns the most recent time point of gold info record.
// It implements data.GlobalStorage interface.
func (bs *BoltStorage) CurrentUSDInfoVersion(timepoint uint64) (common.Version, error) {
	var result uint64
	var err error
	err = bs.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(usdBucket)).Cursor()
		result, err = reverseSeek(timepoint, c)
		return nil
	})
	return common.Version(result), err
}

// GetBTCInfo returns BTC info at given time point. It implements data.GlobalStorage interface.
func (bs *BoltStorage) GetBTCInfo(version common.Version) (common.BTCData, error) {
	var (
		err    error
		result = common.BTCData{}
	)
	err = bs.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(btcBucket))
		data := b.Get(boltutil.Uint64ToBytes(uint64(version)))
		if data == nil {
			return fmt.Errorf("version %s doesn't exist", string(version))
		}
		return json.Unmarshal(data, &result)
	})
	return result, err
}

// StoreBTCInfo stores the given BTC information to database. It implements fetcher.GlobalStorage interface.
func (bs *BoltStorage) StoreBTCInfo(data common.BTCData) error {
	var (
		err       error
		timepoint = data.Timestamp
	)
	err = bs.db.Update(func(tx *bolt.Tx) error {
		var dataJSON []byte
		b := tx.Bucket([]byte(btcBucket))
		dataJSON, uErr := json.Marshal(data)
		if uErr != nil {
			return uErr
		}
		return b.Put(boltutil.Uint64ToBytes(timepoint), dataJSON)
	})
	return err
}

// StoreUSDInfo stores the given USD information to database. It implements fetcher.GlobalStorage interface.
func (bs *BoltStorage) StoreUSDInfo(data common.USDData) error {
	var (
		err       error
		timepoint = data.Timestamp
	)
	err = bs.db.Update(func(tx *bolt.Tx) error {
		var dataJSON []byte
		b := tx.Bucket([]byte(usdBucket))
		dataJSON, uErr := json.Marshal(data)
		if uErr != nil {
			return uErr
		}
		return b.Put(boltutil.Uint64ToBytes(timepoint), dataJSON)
	})
	return err
}

// GetUSDInfo returns USD info at given time point. It implements data.GlobalStorage interface.
func (bs *BoltStorage) GetUSDInfo(version common.Version) (common.USDData, error) {
	var (
		err    error
		result = common.USDData{}
	)
	err = bs.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(usdBucket))
		data := b.Get(boltutil.Uint64ToBytes(uint64(version)))
		if data == nil {
			return fmt.Errorf("version %s doesn't exist", string(version))
		}
		return json.Unmarshal(data, &result)
	})
	return result, err
}

// ExportExpiredAuthData export to back it up
func (bs *BoltStorage) ExportExpiredAuthData(currentTime uint64, fileName string) (nRecord uint64, err error) {
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

	err = bs.db.Update(func(tx *bolt.Tx) error {
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

// PruneExpiredAuthData from database
func (bs *BoltStorage) PruneExpiredAuthData(currentTime uint64) (nRecord uint64, err error) {
	expiredTimestampByte := boltutil.Uint64ToBytes(currentTime - authDataExpiredDuration)

	err = bs.db.Update(func(tx *bolt.Tx) error {
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
func (bs *BoltStorage) PruneOutdatedData(tx *bolt.Tx, bucket string) error {
	var err error
	b := tx.Bucket([]byte(bucket))
	c := b.Cursor()
	nExcess := bs.GetNumberOfVersion(tx, bucket) - maxNumberVersion
	for i := 0; i < nExcess; i++ {
		k, _ := c.First()
		if k == nil {
			err = fmt.Errorf("there is no previous version in %s", bucket)
			return err
		}
		err = b.Delete(k)
		if err != nil {
			return err
		}
	}

	return err
}

func (bs *BoltStorage) CurrentPriceVersion(timepoint uint64) (common.Version, error) {
	var result uint64
	var err error
	err = bs.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(priceBucket)).Cursor()
		result, err = reverseSeek(timepoint, c)
		return err
	})
	return common.Version(result), err
}

// GetNumberOfVersion return number of version storing in a bucket
func (bs *BoltStorage) GetNumberOfVersion(tx *bolt.Tx, bucket string) int {
	result := 0
	b := tx.Bucket([]byte(bucket))
	c := b.Cursor()
	for k, _ := c.First(); k != nil; k, _ = c.Next() {
		result++
	}
	return result
}

//GetAllPrices returns the corresponding AllPriceEntry to a particular Version
func (bs *BoltStorage) GetAllPrices(version common.Version) (common.AllPriceEntry, error) {
	result := common.AllPriceEntry{}
	var err error
	err = bs.db.View(func(tx *bolt.Tx) error {
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

func (bs *BoltStorage) GetOnePrice(pairID uint64, version common.Version) (common.OnePrice, error) {
	result := common.AllPriceEntry{}
	var err error
	err = bs.db.View(func(tx *bolt.Tx) error {
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
	dataPair, exist := result.Data[pairID]
	if exist {
		return dataPair, nil
	}
	return common.OnePrice{}, errors.New("pair of token is not supported")
}

func (bs *BoltStorage) StorePrice(data common.AllPriceEntry, timepoint uint64) error {
	err := bs.db.Update(func(tx *bolt.Tx) error {
		var (
			uErr     error
			dataJSON []byte
		)

		b := tx.Bucket([]byte(priceBucket))

		// remove outdated data from bucket
		log.Printf("Version number: %d\n", bs.GetNumberOfVersion(tx, priceBucket))
		if uErr = bs.PruneOutdatedData(tx, priceBucket); uErr != nil {
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

func (bs *BoltStorage) CurrentAuthDataVersion(timepoint uint64) (common.Version, error) {
	var result uint64
	var err error
	err = bs.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(authDataBucket)).Cursor()
		result, err = reverseSeek(timepoint, c)
		return err
	})
	return common.Version(result), err
}

func (bs *BoltStorage) GetAuthData(version common.Version) (common.AuthDataSnapshot, error) {
	result := common.AuthDataSnapshot{}
	var err error
	err = bs.db.View(func(tx *bolt.Tx) error {
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
func (bs *BoltStorage) CurrentRateVersion(timepoint uint64) (common.Version, error) {
	var result uint64
	var err error
	err = bs.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(rateBucket)).Cursor()
		result, err = reverseSeek(timepoint, c)
		return err
	})
	return common.Version(result), err
}

//GetRates return rates history
func (bs *BoltStorage) GetRates(fromTime, toTime uint64) ([]common.AllRateEntry, error) {
	result := []common.AllRateEntry{}
	if toTime-fromTime > maxGetRatesPeriod {
		return result, fmt.Errorf("time range is too broad, it must be smaller or equal to %d miliseconds", maxGetRatesPeriod)
	}
	var err error
	err = bs.db.View(func(tx *bolt.Tx) error {
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

func (bs *BoltStorage) GetRate(version common.Version) (common.AllRateEntry, error) {
	result := common.AllRateEntry{}
	var err error
	err = bs.db.View(func(tx *bolt.Tx) error {
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

func (bs *BoltStorage) StoreAuthSnapshot(
	data *common.AuthDataSnapshot, timepoint uint64) error {
	err := bs.db.Update(func(tx *bolt.Tx) error {
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
func (bs *BoltStorage) StoreRate(data common.AllRateEntry, timepoint uint64) error {
	log.Printf("Storing rate data to bolt: data(%v), timespoint(%v)", data, timepoint)
	err := bs.db.Update(func(tx *bolt.Tx) error {
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
		} else if uErr = json.Unmarshal(lastValue, &lastEntry); uErr != nil {
			return uErr
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
func (bs *BoltStorage) Record(
	action string,
	id common.ActivityID,
	destination string,
	params common.ActivityParams,
	result common.ActivityResult,
	estatus string,
	mstatus string,
	timepoint uint64) error {

	var err error
	err = bs.db.Update(func(tx *bolt.Tx) error {
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
			// 	bs.RemoveStalePendingActivities(tx, stales)
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
func (bs *BoltStorage) GetActivity(id common.ActivityID) (common.ActivityRecord, error) {
	result := common.ActivityRecord{}

	err := bs.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(activityBucket))
		idBytes := id.ToBytes()
		v := b.Get(idBytes[:])
		if v == nil {
			return errors.New("can not find that activity")
		}
		return json.Unmarshal(v, &result)
	})
	return result, err
}

func (bs *BoltStorage) GetAllRecords(fromTime, toTime uint64) ([]common.ActivityRecord, error) {
	result := []common.ActivityRecord{}
	var err error
	if (toTime-fromTime)/1000000 > maxGetRatesPeriod {
		return result, fmt.Errorf("time range is too broad, it must be smaller or equal to %d miliseconds", maxGetRatesPeriod)
	}
	err = bs.db.View(func(tx *bolt.Tx) error {
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

func getFirstAndCountPendingSetrate(pendings []common.ActivityRecord, minedNonce uint64) (*common.ActivityRecord, uint64, error) {
	var minNonce uint64 = math.MaxUint64
	var minPrice uint64 = math.MaxUint64
	var result *common.ActivityRecord
	var count uint64
	for i, act := range pendings {
		if act.Action == common.ActionSetRate {
			log.Printf("looking for pending set_rates: %+v", act)
			nonce := act.Result.Nonce
			if nonce < minedNonce {
				log.Printf("NONCE_ISSUE: stalled pending set rate transaction, pending: %d, mined: %d",
					nonce, minedNonce)
				continue
			} else if nonce-minedNonce > 1 {
				log.Printf("NONCE_ISSUE: pending set rate transaction for inconsecutive nonce, mined nonce: %d, request nonce: %d",
					minedNonce, nonce)
			}

			gasPrice, err := strconv.ParseUint(act.Result.GasPrice, 10, 64)
			if err != nil {
				return nil, 0, err
			}
			if nonce == minNonce {
				if gasPrice < minPrice {
					minNonce = nonce
					result = &pendings[i]
					minPrice = gasPrice
				}
				count++
			} else if nonce < minNonce {
				minNonce = nonce
				result = &pendings[i]
				minPrice = gasPrice
				count = 1
			}
		}
	}

	if result == nil {
		log.Printf("NONCE_ISSUE: found no pending set rate transaction with nonce newer than equal to mined nonce: %d", minedNonce)
	} else {
		log.Printf("NONCE_ISSUE: unmined pending set rate, nonce: %d, count: %d, mined nonce: %d", result.Result.Nonce, count, minedNonce)
	}

	return result, count, nil
}

//RemoveStalePendingActivities remove it
func (bs *BoltStorage) RemoveStalePendingActivities(tx *bolt.Tx, stales []common.ActivityRecord) error {
	pb := tx.Bucket([]byte(pendingActivityBucket))
	for _, stale := range stales {
		idBytes := stale.ID.ToBytes()
		if err := pb.Delete(idBytes[:]); err != nil {
			return err
		}
	}
	return nil
}

//PendingSetRate return pending set rate activity
func (bs *BoltStorage) PendingSetRate(minedNonce uint64) (*common.ActivityRecord, uint64, error) {
	pendings, err := bs.GetPendingActivities()
	if err != nil {
		return nil, 0, err
	}
	return getFirstAndCountPendingSetrate(pendings, minedNonce)
}

//GetPendingActivities return pending activities
func (bs *BoltStorage) GetPendingActivities() ([]common.ActivityRecord, error) {
	result := []common.ActivityRecord{}
	var err error
	err = bs.db.View(func(tx *bolt.Tx) error {
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
func (bs *BoltStorage) UpdateActivity(id common.ActivityID, activity common.ActivityRecord) error {
	err := bs.db.Update(func(tx *bolt.Tx) error {
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
		return b.Put(idBytes[:], dataJSON)
	})
	return err
}

//HasPendingDeposit check if a deposit is pending
func (bs *BoltStorage) HasPendingDeposit(asset commonv3.Asset, exchange common.Exchange) (bool, error) {
	var (
		err    error
		result = false
	)
	err = bs.db.View(func(tx *bolt.Tx) error {
		pb := tx.Bucket([]byte(pendingActivityBucket))
		c := pb.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			record := common.ActivityRecord{}
			if uErr := json.Unmarshal(v, &record); uErr != nil {
				return uErr
			}
			if record.Action == common.ActionDeposit {
				assetID := record.Params.Asset
				if assetID == asset.ID && record.Destination == exchange.ID().String() {
					result = true
				}
			}
		}
		return nil
	})
	return result, err
}
