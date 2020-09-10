package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data/boltutil"
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/world"
)

const (
	priceBucket                     string = "prices"
	rateBucket                      string = "rates"
	activityBucket                  string = "activities"
	pendingActivityBucket           string = "pending_activities"
	authDataBucket                  string = "auth_data"
	metricBucket                    string = "metrics"
	metricTargetQuantity            string = "target_quantity"
	enableRebalance                 string = "enable_rebalance"
	setrateControl                  string = "setrate_control"
	pwiEquation                     string = "pwi_equation"
	exchangeStatus                  string = "exchange_status"
	exchangeNotifications           string = "exchange_notifications"
	maxNumberVersion                int    = 1000
	aDay                            uint64 = 86400000 //1 days in milisec
	maxGetRatesPeriod                      = aDay
	maxSearchRange                         = aDay
	authDataExpiredDuration         uint64 = 10 * 86400000 //10day in milisec
	stableTokenParamsBucket         string = "stable-token-params"
	pendingStatbleTokenParamsBucket string = "pending-stable-token-params"
	goldBucket                      string = "gold_feeds"
	btcBucket                       string = "btc_feeds"
	usdBucket                       string = "usd_feeds"
	disabledFeedsBucket             string = "disabled_feeds"
	feedSetting                     string = "feed_setting"
	pendingFeedSetting              string = "pending_feed_setting"
	generalBucket                   string = "general_bucket"

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

	//btcFetcherConfiguration stores configuration for btc fetcher
	fetcherConfigurationBucket = "btc_fetcher_configuration"
)

var buckets = []string{
	goldBucket,
	btcBucket,
	usdBucket,
	disabledFeedsBucket,
	priceBucket,
	rateBucket,
	activityBucket,
	pendingActivityBucket,
	authDataBucket,
	metricBucket,
	metricTargetQuantity,
	enableRebalance,
	setrateControl,
	pwiEquation,
	exchangeStatus,
	exchangeNotifications,
	pendingStatbleTokenParamsBucket,
	stableTokenParamsBucket,
	pendingTargetQuantityV2,
	targetQuantityV2,
	pendingPWIEquationV2,
	pwiEquationV2,
	pendingRebalanceQuadratic,
	rebalanceQuadratic,
	fetcherConfigurationBucket,
	feedSetting,
	pendingFeedSetting,
	generalBucket,
}

// BoltStorage is the storage implementation of data.Storage interface
// that uses BoltDB as its storage engine.
type BoltStorage struct {
	mu sync.RWMutex
	db *bolt.DB
	l  *zap.SugaredLogger
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
		l:  zap.S(),
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
		return err
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
			err = fmt.Errorf("version %d doesn't exist", version)
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
		return err
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
		return err
	})
	return common.Version(result), err
}

var gasThresholdKey = `gas_threshold`

func (bs *BoltStorage) SetGasThreshold(v common.GasThreshold) error {
	return bs.saveGeneralBucket(gasThresholdKey, v)
}

func (bs *BoltStorage) GetGasThreshold() (common.GasThreshold, error) {
	var res common.GasThreshold
	err := bs.getGeneralBucket(gasThresholdKey, &res)
	return res, err
}

var preferGasSourceKey = `prefer_gas_source`

func (bs *BoltStorage) SetPreferGasSource(v common.PreferGasSource) error {
	return bs.saveGeneralBucket(preferGasSourceKey, v)
}

func (bs *BoltStorage) GetPreferGasSource() (common.PreferGasSource, error) {
	var res common.PreferGasSource
	err := bs.getGeneralBucket(preferGasSourceKey, &res)
	return res, err
}

func (bs *BoltStorage) saveGeneralBucket(key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return bs.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(generalBucket))
		return b.Put([]byte(key), data)
	})
}

func (bs *BoltStorage) getGeneralBucket(key string, value interface{}) error {
	err := bs.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(generalBucket))
		d := b.Get([]byte(key))
		if d == nil {
			return errors.New("entry not found in storage")
		}
		return json.Unmarshal(d, value)
	})

	return err
}

func (bs *BoltStorage) UpdateFeedConfiguration(name string, enabled bool) error {
	const disableValue = "disabled"
	var (
		allFeeds = world.AllFeeds()
		exists   = false
	)
	// ignore case
	name = strings.ToLower(name)

	for _, feed := range allFeeds {
		if strings.EqualFold(name, feed) {
			exists = true
			break
		}
	}
	if !exists {
		return fmt.Errorf("unknown feed: %q", name)
	}

	return bs.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(disabledFeedsBucket))
		if v := b.Get([]byte(name)); v == nil {
			// feed does not exists in disabled feeds bucket yet
			if !enabled {
				return b.Put([]byte(name), []byte(disableValue))
			}
			return nil
		}

		if enabled {
			return b.Delete([]byte(name))
		}
		return nil
	})
}

func (bs *BoltStorage) GetFeedConfiguration() ([]common.FeedConfiguration, error) {
	var (
		err      error
		allFeeds = world.AllFeeds()
		results  []common.FeedConfiguration
	)

	for _, feed := range allFeeds {
		results = append(results, common.FeedConfiguration{Name: feed, Enabled: true})
	}

	err = bs.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(disabledFeedsBucket))
		return b.ForEach(func(k, _ []byte) error {
			for i := range results {
				if strings.ToLower(results[i].Name) == string(k) {
					results[i].Enabled = false
					break
				}
			}
			return nil
		})
	})

	return results, err
}

func checkFeedExist(pendingFeed string, allFeeds []string) bool {
	for _, feed := range allFeeds {
		if feed == pendingFeed {
			return true
		}
	}
	return false
}

func (bs *BoltStorage) StorePendingFeedSetting(value []byte) error {
	var (
		err         error
		allFeeds    = world.AllFeeds()
		pendingData common.MapFeedSetting
	)

	if err = json.Unmarshal(value, &pendingData); err != nil {
		return fmt.Errorf("rejected: Data could not be unmarshalled to defined format: %v", err)
	}

	for pendingFeed := range pendingData {
		if isExist := checkFeedExist(pendingFeed, allFeeds); !isExist {
			return fmt.Errorf("rejected: feed doesn't exist, feed=%s", pendingFeed)
		}
	}

	err = bs.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(pendingFeedSetting))
		k := []byte("current_pending_feed_setting")
		if b.Get(k) != nil {
			return fmt.Errorf("currently there is a pending record")
		}
		return b.Put(k, value)
	})
	return err
}

func (bs *BoltStorage) ConfirmPendingFeedSetting(value []byte) error {
	var confirmFeedSetting common.MapFeedSetting
	err := json.Unmarshal(value, &confirmFeedSetting)
	if err != nil {
		return fmt.Errorf("rejected: Data could not be unmarshalled to defined format: %s", err)
	}

	pending, err := bs.GetPendingFeedSetting()
	if err != nil {
		return err
	}

	if eq := reflect.DeepEqual(pending, confirmFeedSetting); !eq {
		return fmt.Errorf("rejected: confiming data isn't consistent")
	}

	err = bs.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(feedSetting))
		for k, v := range confirmFeedSetting {
			dataJSON, err := json.Marshal(v)
			if err != nil {
				return err
			}
			if err := b.Put([]byte(k), dataJSON); err != nil {
				return err
			}
		}
		pendingBk := tx.Bucket([]byte(pendingFeedSetting))
		pendingKey := []byte("current_pending_feed_setting")
		return pendingBk.Delete(pendingKey)
	})
	return err
}

func (bs *BoltStorage) RejectPendingFeedSetting() error {
	return bs.db.Update(func(tx *bolt.Tx) error {
		pendingBk := tx.Bucket([]byte(pendingFeedSetting))
		pendingKey := []byte("current_pending_feed_setting")
		if pendingData := pendingBk.Get(pendingKey); pendingData == nil {
			return errors.New("there are no pending feed setting")
		}
		return pendingBk.Delete(pendingKey)
	})
}

func (bs *BoltStorage) GetPendingFeedSetting() (common.MapFeedSetting, error) {
	var (
		pendingData common.MapFeedSetting
		err         error
	)
	if err = bs.db.View(func(tx *bolt.Tx) error {
		pendingBk := tx.Bucket([]byte(pendingFeedSetting))
		pendingKey := []byte("current_pending_feed_setting")
		record := pendingBk.Get(pendingKey)
		if record == nil {
			return errors.New("there are no pending feed setting")
		}
		if err := json.Unmarshal(record, &pendingData); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return pendingData, err
	}
	return pendingData, nil
}

func (bs *BoltStorage) GetFeedSetting() (common.MapFeedSetting, error) {
	var (
		err      error
		allFeeds = world.AllFeeds()
		results  = make(common.MapFeedSetting)
	)

	err = bs.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(feedSetting))
		for _, feed := range allFeeds {
			var setting common.FeedSetting
			record := b.Get([]byte(feed))
			if record != nil {
				if errU := json.Unmarshal(record, &setting); errU != nil {
					return errU
				}
			}
			results[feed] = setting
		}
		return nil
	})
	return results, err
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
			return fmt.Errorf("version %d doesn't exist", version)
		}
		return json.Unmarshal(data, &result)
	})
	return result, err
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
			return fmt.Errorf("version %d doesn't exist", version)
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

func (bs *BoltStorage) ExportExpiredAuthData(currentTime uint64, fileName string) (nRecord uint64, err error) {
	expiredTimestampByte := boltutil.Uint64ToBytes(currentTime - authDataExpiredDuration)
	outFile, err := os.Create(fileName)
	if err != nil {
		return 0, err
	}
	defer func() {
		if cErr := outFile.Close(); cErr != nil {
			bs.l.Warnw("Close file", "err", cErr)
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
			err = fmt.Errorf("version %d doesn't exist", version)
		} else {
			err = json.Unmarshal(data, &result)
		}
		return err
	})
	return result, err
}

func (bs *BoltStorage) GetOnePrice(pair common.TokenPairID, version common.Version) (common.OnePrice, error) {
	result := common.AllPriceEntry{}
	var err error
	err = bs.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(priceBucket))
		data := b.Get(boltutil.Uint64ToBytes(uint64(version)))
		if data == nil {
			err = fmt.Errorf("version %d doesn't exist", version)
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
		bs.l.Infof("Version number: %d", bs.GetNumberOfVersion(tx, priceBucket))
		if uErr = bs.PruneOutdatedData(tx, priceBucket); uErr != nil {
			bs.l.Warnw("Prune out data", "err", uErr)
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
			err = fmt.Errorf("version %d doesn't exist", version)
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
			err = fmt.Errorf("version %d doesn't exist", version)
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
	bs.l.Infof("Storing rate data to bolt: data(%v), timespoint(%v)", data, timepoint)
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
			bs.l.Infof("Bucket %s is empty", rateBucket)
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

		// It is common that two nodes return different block number
		// But if the gap is too big we should log it
		if lastEntry.BlockNumber >= data.BlockNumber+2 {
			return errors.Errorf("rejected storing rates with smaller block number: %d, stored: %d",
				data.BlockNumber,
				lastEntry.BlockNumber)
		}
		return nil
	})
	return err
}

// interfaceConverstionToUint64 will assert the interface as string
// and parse it to uint64. Return 0 if anything goes wrong)
func interfaceConverstionToUint64(l *zap.SugaredLogger, intf interface{}) uint64 {
	numString, ok := intf.(string)
	if !ok {
		l.Warnw("can't be converted to type string", "value", fmt.Sprintf("%+v", intf))
		return 0
	}
	num, err := strconv.ParseUint(numString, 10, 64)
	if err != nil {
		l.Warnw("ERROR: parsing error, interface conversion to uint64 will set to 0", "err", err)
		return 0
	}
	return num
}

//StoreMetric store metric info
func (bs *BoltStorage) StoreMetric(data *common.MetricEntry, timepoint uint64) error {
	var err error
	err = bs.db.Update(func(tx *bolt.Tx) error {
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
func (bs *BoltStorage) GetMetric(tokens []common.Token, fromTime, toTime uint64) (map[string]common.MetricList, error) {
	imResult := map[string]*common.MetricList{}
	for _, tok := range tokens {
		imResult[tok.ID] = &common.MetricList{}
	}

	var err error
	err = bs.db.View(func(tx *bolt.Tx) error {
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

//GetTokenTargetQty get target quantity
func (bs *BoltStorage) GetTokenTargetQty() (common.TokenTargetQty, error) {
	var (
		tokenTargetQty = common.TokenTargetQty{}
		err            error
	)
	err = bs.db.View(func(tx *bolt.Tx) error {
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

func (bs *BoltStorage) GetRebalanceControl() (common.RebalanceControl, error) {
	var err error
	var result common.RebalanceControl
	var setDefault = false
	err = bs.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(enableRebalance))
		_, data := b.Cursor().First()
		// if data == nil set default value in bolt db
		if data == nil {
			result = common.RebalanceControl{
				Status: false,
			}
			setDefault = true
			return nil
		}
		return json.Unmarshal(data, &result)
	})
	if err != nil {
		return result, err
	}
	if setDefault {
		err = bs.StoreRebalanceControl(false)
	}
	return result, err
}

func (bs *BoltStorage) StoreRebalanceControl(status bool) error {
	err := bs.db.Update(func(tx *bolt.Tx) error {
		var (
			uErr     error
			dataJSON []byte
		)
		b := tx.Bucket([]byte(enableRebalance))
		// prune out old data
		c := b.Cursor()
		k, _ := c.First()
		if k != nil {
			if uErr = b.Delete(k); uErr != nil {
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

func (bs *BoltStorage) GetSetrateControl() (common.SetrateControl, error) {
	var err error
	var result common.SetrateControl
	var setDefault = false
	err = bs.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(setrateControl))
		_, data := b.Cursor().First()
		// if data == nil set default value in bolt db
		if data == nil {
			result = common.SetrateControl{
				Status: false,
			}
			setDefault = true
			return nil
		}
		return json.Unmarshal(data, &result)
	})
	if err != nil {
		return result, err
	}
	if setDefault {
		err = bs.StoreSetrateControl(false)
	}
	return result, err
}

func (bs *BoltStorage) StoreSetrateControl(status bool) error {
	err := bs.db.Update(func(tx *bolt.Tx) error {
		var (
			uErr     error
			dataJSON []byte
		)
		b := tx.Bucket([]byte(setrateControl))
		// prune out old data
		c := b.Cursor()
		k, _ := c.First()
		if k != nil {
			if uErr = b.Delete(k); uErr != nil {
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

// GetExchangeStatus get exchange status to dashboard and analytics
func (bs *BoltStorage) GetExchangeStatus() (common.ExchangesStatus, error) {
	result := make(common.ExchangesStatus)
	err := bs.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(exchangeStatus))
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

func (bs *BoltStorage) UpdateExchangeStatus(data common.ExchangesStatus) error {
	err := bs.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(exchangeStatus))
		for k, v := range data {
			dataJSON, uErr := json.Marshal(v)
			if uErr != nil {
				return uErr
			}
			if uErr := b.Put([]byte(k), dataJSON); uErr != nil {
				return uErr
			}
		}
		return nil
	})
	return err
}

func (bs *BoltStorage) UpdateExchangeNotification(
	exchange, action, token string, fromTime, toTime uint64, isWarning bool, msg string) error {
	err := bs.db.Update(func(tx *bolt.Tx) error {
		exchangeBk := tx.Bucket([]byte(exchangeNotifications))
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

func (bs *BoltStorage) GetExchangeNotifications() (common.ExchangeNotifications, error) {
	result := common.ExchangeNotifications{}
	err := bs.db.View(func(tx *bolt.Tx) error {
		exchangeBks := tx.Bucket([]byte(exchangeNotifications))
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
					if uErr := json.Unmarshal(v, &notiContent); uErr != nil {
						return uErr
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

func (bs *BoltStorage) SetStableTokenParams(value []byte) error {
	var err error
	k := boltutil.Uint64ToBytes(1)
	temp := make(map[string]interface{})

	if err = json.Unmarshal(value, &temp); err != nil {
		return fmt.Errorf("rejected: Data could not be unmarshalled to defined format: %s", err)
	}
	err = bs.db.Update(func(tx *bolt.Tx) error {
		b, uErr := tx.CreateBucketIfNotExists([]byte(pendingStatbleTokenParamsBucket))
		if uErr != nil {
			return uErr
		}
		if b.Get(k) != nil {
			return errors.New("currently there is a pending record")
		}
		return b.Put(k, value)
	})
	return err
}

func (bs *BoltStorage) ConfirmStableTokenParams(value []byte) error {
	var err error
	k := boltutil.Uint64ToBytes(1)
	temp := make(map[string]interface{})

	if err = json.Unmarshal(value, &temp); err != nil {
		return fmt.Errorf("rejected: Data could not be unmarshalled to defined format: %s", err)
	}

	pending, err := bs.GetPendingStableTokenParams()
	if err != nil {
		return err
	}

	if eq := reflect.DeepEqual(pending, temp); !eq {
		return errors.New("rejected: confiming data isn't consistent")
	}

	err = bs.db.Update(func(tx *bolt.Tx) error {
		b, uErr := tx.CreateBucketIfNotExists([]byte(stableTokenParamsBucket))
		if uErr != nil {
			return uErr
		}
		return b.Put(k, value)
	})
	if err != nil {
		return err
	}
	return bs.RemovePendingStableTokenParams()
}

func (bs *BoltStorage) GetStableTokenParams() (map[string]interface{}, error) {
	k := boltutil.Uint64ToBytes(1)
	result := make(map[string]interface{})
	err := bs.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(stableTokenParamsBucket))
		if b == nil {
			return errors.New("bucket hasn't exist yet")
		}
		record := b.Get(k)
		if record != nil {
			return json.Unmarshal(record, &result)
		}
		return nil
	})
	return result, err
}

func (bs *BoltStorage) GetPendingStableTokenParams() (map[string]interface{}, error) {
	k := boltutil.Uint64ToBytes(1)
	result := make(map[string]interface{})
	err := bs.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(pendingStatbleTokenParamsBucket))
		if b == nil {
			return errors.New("bucket hasn't exist yet")
		}
		record := b.Get(k)
		if record != nil {
			return json.Unmarshal(record, &result)
		}
		return nil
	})
	return result, err
}

func (bs *BoltStorage) RemovePendingStableTokenParams() error {
	k := boltutil.Uint64ToBytes(1)
	err := bs.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(pendingStatbleTokenParamsBucket))
		if b == nil {
			return errors.New("bucket hasn't existed yet")
		}
		record := b.Get(k)
		if record == nil {
			return errors.New("bucket is empty")
		}
		return b.Delete(k)
	})
	return err
}

//StorePendingTargetQtyV2 store value into pending target qty v2 bucket
func (bs *BoltStorage) StorePendingTargetQtyV2(value []byte) error {
	var (
		err         error
		pendingData common.TokenTargetQtyV2
	)

	if err = json.Unmarshal(value, &pendingData); err != nil {
		return fmt.Errorf("rejected: Data could not be unmarshalled to defined format: %v", err)
	}
	err = bs.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(pendingTargetQuantityV2))
		k := []byte("current_pending_target_qty")
		if b.Get(k) != nil {
			return fmt.Errorf("currently there is a pending record")
		}
		return b.Put(k, value)
	})
	return err
}

//GetPendingTargetQtyV2 return current pending target quantity
func (bs *BoltStorage) GetPendingTargetQtyV2() (common.TokenTargetQtyV2, error) {
	result := common.TokenTargetQtyV2{}
	err := bs.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(pendingTargetQuantityV2))
		k := []byte("current_pending_target_qty")
		record := b.Get(k)
		if record == nil {
			return errors.New("there is no pending target qty")
		}
		return json.Unmarshal(record, &result)
	})
	return result, err
}

//ConfirmTargetQtyV2 check if confirm data match pending data and save it to confirm bucket
//remove pending data from pending bucket
func (bs *BoltStorage) ConfirmTargetQtyV2(value []byte) error {
	confirmTargetQty := common.TokenTargetQtyV2{}
	err := json.Unmarshal(value, &confirmTargetQty)
	if err != nil {
		return fmt.Errorf("rejected: Data could not be unmarshalled to defined format: %s", err)
	}

	pending, err := bs.GetPendingTargetQtyV2()
	if err != nil {
		return err
	}

	if eq := reflect.DeepEqual(pending, confirmTargetQty); !eq {
		return fmt.Errorf("rejected: confiming data isn't consistent")
	}

	err = bs.db.Update(func(tx *bolt.Tx) error {
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
func (bs *BoltStorage) RemovePendingTargetQtyV2() error {
	err := bs.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(pendingTargetQuantityV2))
		if b == nil {
			return fmt.Errorf("bucket hasn't existed yet")
		}
		k := []byte("current_pending_target_qty")
		return b.Delete(k)
	})
	return err
}

// GetTargetQtyV2 return the current target quantity
func (bs *BoltStorage) GetTargetQtyV2() (common.TokenTargetQtyV2, error) {
	result := common.TokenTargetQtyV2{}
	err := bs.db.View(func(tx *bolt.Tx) error {
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
		targetQty, err = bs.GetTokenTargetQty()
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
func (bs *BoltStorage) StorePendingPWIEquationV2(data []byte) error {
	timepoint := common.GetTimepoint()
	err := bs.db.Update(func(tx *bolt.Tx) error {
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
func (bs *BoltStorage) GetPendingPWIEquationV2() (common.PWIEquationRequestV2, error) {
	var (
		err    error
		result common.PWIEquationRequestV2
	)

	err = bs.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(pendingPWIEquationV2))
		c := b.Cursor()
		_, v := c.First()
		if v == nil {
			return errors.New("there is no pending equation")
		}
		return json.Unmarshal(v, &result)
	})
	return result, err
}

// RemovePendingPWIEquationV2 deletes the pending equation request.
func (bs *BoltStorage) RemovePendingPWIEquationV2() error {
	err := bs.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(pendingPWIEquationV2))
		c := b.Cursor()
		k, _ := c.First()
		if k == nil {
			return errors.New("there is no pending data")
		}
		return b.Delete(k)
	})
	return err
}

// StorePWIEquationV2 moved the pending equation request to
// pwiEquationV2 bucket and remove it from pending bucket if the
// given data matched what stored.
func (bs *BoltStorage) StorePWIEquationV2(data string) error {
	err := bs.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(pendingPWIEquationV2))
		c := b.Cursor()
		k, v := c.First()
		if v == nil {
			return errors.New("there is no pending equation")
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
			return errors.New("confirm data does not match pending data")
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
		return nil, errors.New("there is no equation")
	}
	if err := json.Unmarshal(v, &eqv1); err != nil {
		return nil, err
	}
	return convertPWIEquationV1toV2(eqv1.Data)
}

// GetPWIEquationV2 returns the current PWI equations from database.
func (bs *BoltStorage) GetPWIEquationV2() (common.PWIEquationRequestV2, error) {
	var (
		err    error
		result common.PWIEquationRequestV2
	)
	err = bs.db.View(func(tx *bolt.Tx) error {
		var vErr error // convert pwi v1 to v2 error
		b := tx.Bucket([]byte(pwiEquationV2))
		c := b.Cursor()
		_, v := c.Last()
		if v == nil {
			bs.l.Infof("there no equation in pwiEquationV2, getting from pwiEquation")
			result, vErr = pwiEquationV1toV2(tx)
			return vErr
		}
		return json.Unmarshal(v, &result)
	})
	return result, err
}

//StorePendingRebalanceQuadratic store pending data (stand for rebalance quadratic equation) to db
//data byte for json {"KNC": {"a": 0.9, "b": 1.2, "c": 1.4}}
func (bs *BoltStorage) StorePendingRebalanceQuadratic(data []byte) error {
	timepoint := common.GetTimepoint()
	err := bs.db.Update(func(tx *bolt.Tx) error {
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
func (bs *BoltStorage) GetPendingRebalanceQuadratic() (common.RebalanceQuadraticRequest, error) {
	var result common.RebalanceQuadraticRequest
	err := bs.db.View(func(tx *bolt.Tx) error {
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
func (bs *BoltStorage) ConfirmRebalanceQuadratic(data []byte) error {
	err := bs.db.Update(func(tx *bolt.Tx) error {
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
func (bs *BoltStorage) RemovePendingRebalanceQuadratic() error {
	err := bs.db.Update(func(tx *bolt.Tx) error {
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
func (bs *BoltStorage) GetRebalanceQuadratic() (common.RebalanceQuadraticRequest, error) {
	var result common.RebalanceQuadraticRequest
	err := bs.db.View(func(tx *bolt.Tx) error {
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

//storeJSONByteArray store key-value into a bucket under one TX.
func (bs *BoltStorage) storeJSONByteArray(tx *bolt.Tx, bucketName string, key, value []byte) error {
	b := tx.Bucket([]byte(bucketName))
	if b == nil {
		return fmt.Errorf("bucket %s hasn't existed yet", bucketName)
	}
	return b.Put(key, value)
}

// ConfirmTokenUpdateInfo confirm update token info
func (bs *BoltStorage) ConfirmTokenUpdateInfo(tarQty common.TokenTargetQtyV2, pwi common.PWIEquationRequestV2, quadEq common.RebalanceQuadraticRequest) error {
	timeStampKey := boltutil.Uint64ToBytes(common.GetTimepoint())
	err := bs.db.Update(func(tx *bolt.Tx) error {
		dataJSON, uErr := json.Marshal(tarQty)
		if uErr != nil {
			return uErr
		}
		if uErr = bs.storeJSONByteArray(tx, targetQuantityV2, []byte("current_target_qty"), dataJSON); uErr != nil {
			return uErr
		}
		if dataJSON, uErr = json.Marshal(pwi); uErr != nil {
			return uErr
		}
		if uErr = bs.storeJSONByteArray(tx, pwiEquationV2, timeStampKey, dataJSON); uErr != nil {
			return uErr
		}
		if dataJSON, uErr = json.Marshal(quadEq); uErr != nil {
			return uErr
		}
		return bs.storeJSONByteArray(tx, rebalanceQuadratic, timeStampKey, dataJSON)
	})
	return err
}
