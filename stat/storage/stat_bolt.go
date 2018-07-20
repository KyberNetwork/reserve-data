package storage

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/KyberNetwork/reserve-data/boltutil"
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/stat"
	"github.com/boltdb/bolt"
	ethereum "github.com/ethereum/go-ethereum/common"
)

const (
	tradeLogProcessorState string = "tradelog_processor_state"

	minuteBucket string = "minute"
	hourBucket   string = "hour"
	dayBucket    string = "day"

	walletAddressBucket string = "wallet_address"
	reserveRates        string = "reserve_rates"
	countryBucket       string = "country_stat_bucket"
	userFirstTradeEver  string = "user_first_trade_ever"
	userStatBucket      string = "user_stat_bucket"
	userListBucket      string = "user_list"
)

//BoltStatStorage object
type BoltStatStorage struct {
	db *bolt.DB
}

//NewBoltStatStorage return new storage instance
func NewBoltStatStorage(path string) (*BoltStatStorage, error) {
	// init instance
	var err error
	var db *bolt.DB
	db, err = bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}
	// init buckets
	err = db.Update(func(tx *bolt.Tx) error {
		if _, uErr := tx.CreateBucketIfNotExists([]byte(tradeLogProcessorState)); uErr != nil {
			return uErr
		}
		if _, uErr := tx.CreateBucketIfNotExists([]byte(walletAddressBucket)); uErr != nil {
			return uErr
		}
		if _, uErr := tx.CreateBucketIfNotExists([]byte(countryBucket)); uErr != nil {
			return uErr
		}
		if _, uErr := tx.CreateBucketIfNotExists([]byte(userFirstTradeEver)); uErr != nil {
			return uErr
		}
		if _, uErr := tx.CreateBucketIfNotExists([]byte(userStatBucket)); uErr != nil {
			return uErr
		}
		if _, uErr := tx.CreateBucketIfNotExists([]byte(walletAddressBucket)); uErr != nil {
			return uErr
		}
		if _, uErr := tx.CreateBucketIfNotExists([]byte(userListBucket)); uErr != nil {
			return uErr
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	storage := &BoltStatStorage{db}
	return storage, nil
}

func (self *BoltStatStorage) SetLastProcessedTradeLogTimepoint(statType string, timepoint uint64) error {
	var err error
	err = self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(tradeLogProcessorState))
		if b == nil {
			return errors.New("Cannot find last processed bucket")
		}
		err = b.Put([]byte(statType), boltutil.Uint64ToBytes(timepoint))
		return err
	})
	return err
}

func (self *BoltStatStorage) GetLastProcessedTradeLogTimepoint(statType string) (uint64, error) {
	var result uint64
	var err error
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(tradeLogProcessorState))
		if b == nil {
			return errors.New("Cannot find last processed bucket")
		}
		result = boltutil.BytesToUint64(b.Get([]byte(statType)))
		return nil
	})
	return result, err
}

func getBucketNameByFreq(freq string) (bucketName string, err error) {
	switch freq {
	case "m", "M":
		bucketName = minuteBucket
	case "h", "H":
		bucketName = hourBucket
	case "d", "D":
		bucketName = dayBucket
	default:
		offset, ok := strconv.ParseInt(strings.TrimPrefix(freq, "utc"), 10, 64)
		if (offset < stat.StartTimezone) || (offset > stat.EndTimezone) {
			err = errors.New("Frequency is wrong, can not get bucket name")
		}
		if ok != nil {
			err = ok
		}
		bucketName = freq
	}
	return
}

func getTimestampByFreq(t uint64, freq string) (result []byte) {
	ui64Day := uint64(time.Hour * 24)
	switch freq {
	case "m", "M":
		result = boltutil.Uint64ToBytes(t / uint64(time.Minute) * uint64(time.Minute))
	case "h", "H":
		result = boltutil.Uint64ToBytes(t / uint64(time.Hour) * uint64(time.Hour))
	case "d", "D":
		result = boltutil.Uint64ToBytes(t / ui64Day * ui64Day)
	default:
		// utc timezone
		offset, _ := strconv.ParseInt(strings.TrimPrefix(freq, "utc"), 10, 64)
		ui64offset := uint64(int64(time.Hour) * offset)
		if offset > 0 {
			result = boltutil.Uint64ToBytes((t+ui64offset)/ui64Day*ui64Day + ui64offset)
		} else {
			offset = 0 - offset
			result = boltutil.Uint64ToBytes((t-ui64offset)/ui64Day*ui64Day - ui64offset)
		}
	}

	return
}

func (self *BoltStatStorage) SetWalletAddress(ethWalletAddr ethereum.Address) (err error) {
	walletAddr := common.AddrToString(ethWalletAddr)
	err = self.db.Update(func(tx *bolt.Tx) error {
		walletBucket := tx.Bucket([]byte(walletAddressBucket))
		if walletBucket == nil {
			return fmt.Errorf("cannot find bucket %s", walletAddressBucket)
		}
		return walletBucket.Put([]byte(walletAddr), []byte("1"))
	})
	return
}

// GetWalletAddresses return a set of wallet address currently in core.
// Return empty result if there is no wallet address curently
func (self *BoltStatStorage) GetWalletAddresses() ([]string, error) {
	var result []string
	var err error
	err = self.db.View(func(tx *bolt.Tx) error {
		walletBucket := tx.Bucket([]byte(walletAddressBucket))
		if walletBucket == nil {
			return fmt.Errorf("GetWalletAddresses cannot get bucket %s", walletAddressBucket)
		}
		c := walletBucket.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			result = append(result, string(k[:]))
		}
		return nil
	})
	return result, err
}

func (self *BoltStatStorage) SetBurnFeeStat(burnFeeStats map[string]common.BurnFeeStatsTimeZone, lastProcessTimePoint uint64) error {
	err := self.db.Update(func(tx *bolt.Tx) error {
		for key, timezoneData := range burnFeeStats {
			key = strings.ToLower(key)
			// This is somewhat odd. There is no burnFee bucket? There should be a burnFee bucket and nest these timeZone bucket inside
			burnFeeBk, uErr := tx.CreateBucketIfNotExists([]byte(key))
			if uErr != nil {
				return uErr
			}
			for _, freq := range []string{"M", "H", "D"} {
				stats := timezoneData[freq]
				freqBkName, uErr := getBucketNameByFreq(freq)
				if uErr != nil {
					return uErr
				}
				freqBk, uErr := burnFeeBk.CreateBucketIfNotExists([]byte(freqBkName))
				if uErr != nil {
					return uErr
				}
				for timepoint, stat := range stats {
					timestamp := boltutil.Uint64ToBytes(timepoint)
					currentData := common.BurnFeeStats{}
					v := freqBk.Get(timestamp)
					if v != nil {
						if uErr := json.Unmarshal(v, &currentData); uErr != nil {
							return uErr
						}
					}
					currentData.TotalBurnFee += stat.TotalBurnFee
					dataJSON, uErr := json.Marshal(currentData)
					if uErr != nil {
						return uErr
					}
					if uErr := freqBk.Put(timestamp, dataJSON); uErr != nil {
						return uErr
					}
				}
			}
		}
		lastProcessBk := tx.Bucket([]byte(tradeLogProcessorState))
		if lastProcessBk == nil {
			return fmt.Errorf("cannot find Bucket %s", tradeLogProcessorState)
		}
		dataJSON := boltutil.Uint64ToBytes(lastProcessTimePoint)
		return lastProcessBk.Put([]byte(stat.BurnfeeAggregation), dataJSON)
	})
	return err
}

func (self *BoltStatStorage) SetVolumeStat(volumeStats map[string]common.VolumeStatsTimeZone, lastProcessTimePoint uint64) error {
	err := self.db.Update(func(tx *bolt.Tx) error {
		for asset, freqData := range volumeStats {
			asset = strings.ToLower(asset)
			// TODO: There should be a volume bucket and Nested these asset bucket inside
			volumeBk, uErr := tx.CreateBucketIfNotExists([]byte(asset))
			if uErr != nil {
				return uErr
			}
			for _, freq := range []string{"M", "H", "D"} {
				stats := freqData[freq]
				freqBkName, uErr := getBucketNameByFreq(freq)
				if uErr != nil {
					return uErr
				}
				freqBk, uErr := volumeBk.CreateBucketIfNotExists([]byte(freqBkName))
				if uErr != nil {
					return uErr
				}
				for timepoint, stat := range stats {
					timestamp := boltutil.Uint64ToBytes(timepoint)
					currentData := common.VolumeStats{}
					v := freqBk.Get(timestamp)
					if v != nil {
						if err := json.Unmarshal(v, &currentData); err != nil {
							return err
						}
					}
					currentData.ETHVolume += stat.ETHVolume
					currentData.USDAmount += stat.USDAmount
					currentData.Volume += stat.Volume

					dataJSON, uErr := json.Marshal(currentData)
					if uErr != nil {
						return uErr
					}
					if uErr := freqBk.Put(timestamp, dataJSON); uErr != nil {
						return uErr
					}
				}
			}
		}
		lastProcessBk := tx.Bucket([]byte(tradeLogProcessorState))
		if lastProcessBk == nil {
			return fmt.Errorf("cannot find Bucket %s", tradeLogProcessorState)
		}
		dataJSON := boltutil.Uint64ToBytes(lastProcessTimePoint)
		return lastProcessBk.Put([]byte(stat.VolumeStatAggregation), dataJSON)
	})
	return err
}

func addMetricData(currentData, stat common.MetricStats) common.MetricStats {
	currentData.ETHVolume += stat.ETHVolume
	currentData.USDVolume += stat.USDVolume
	currentData.BurnFee += stat.BurnFee
	currentData.TradeCount += stat.TradeCount
	currentData.UniqueAddr += stat.UniqueAddr
	currentData.NewUniqueAddresses += stat.NewUniqueAddresses
	currentData.KYCEd += stat.KYCEd
	if currentData.TradeCount > 0 {
		currentData.ETHPerTrade = currentData.ETHVolume / float64(currentData.TradeCount)
		currentData.USDPerTrade = currentData.USDVolume / float64(currentData.TradeCount)
	}
	return currentData
}

func (self *BoltStatStorage) SetWalletStat(stats map[string]common.MetricStatsTimeZone, lastProcessTimePoint uint64) error {
	err := self.db.Update(func(tx *bolt.Tx) error {
		for wallet, timeZoneStat := range stats {
			wallet = strings.ToLower(wallet)
			b, uErr := tx.CreateBucketIfNotExists([]byte(wallet))
			if uErr != nil {
				return uErr
			}
			for i := stat.StartTimezone; i <= stat.EndTimezone; i++ {
				stats := timeZoneStat[i]
				freq := fmt.Sprintf("%s%d", stat.TimezoneBucketPrefix, i)
				walletTzBucket, uErr := b.CreateBucketIfNotExists([]byte(freq))
				if uErr != nil {
					return uErr
				}
				for timepoint, stat := range stats {
					timestamp := boltutil.Uint64ToBytes(timepoint)
					currentData := common.MetricStats{}
					v := walletTzBucket.Get(timestamp)
					if v != nil {
						if uErr := json.Unmarshal(v, &currentData); uErr != nil {
							return uErr
						}
					}
					currentData = addMetricData(currentData, stat)
					dataJSON, uErr := json.Marshal(currentData)
					if uErr != nil {
						return uErr
					}
					if uErr := walletTzBucket.Put(timestamp, dataJSON); uErr != nil {
						return uErr
					}
				}
			}
		}
		lastProcessBk := tx.Bucket([]byte(tradeLogProcessorState))
		if lastProcessBk == nil {
			return fmt.Errorf("cannot find Bucket %s", tradeLogProcessorState)
		}
		dataJSON := boltutil.Uint64ToBytes(lastProcessTimePoint)
		return lastProcessBk.Put([]byte(stat.WalletAggregation), dataJSON)
	})
	return err
}

// GetWalletStats returns StatTicks for a specific address in a specific time range.
// If the wallet/timezone data isn't available, return empty result and no error.
// If the data is corrupted, error is returned.
func (self *BoltStatStorage) GetWalletStats(fromTime uint64, toTime uint64, ethWalletAddr ethereum.Address, timezone int64) (common.StatTicks, error) {
	walletAddr := common.AddrToString(ethWalletAddr)
	result := common.StatTicks{}
	tzstring := fmt.Sprintf("%s%d", stat.TimezoneBucketPrefix, timezone)
	err := self.db.View(func(tx *bolt.Tx) error {
		walletBk := tx.Bucket([]byte(walletAddr))
		if walletBk == nil {
			log.Printf("GetWalletStats cannot find bucket %s", walletAddr)
			return nil
		}
		timezoneBk := walletBk.Bucket([]byte(tzstring))
		if timezoneBk == nil {
			log.Printf("GetWalletStats cannot find bucket %s->%s", walletAddr, tzstring)
			return nil
		}
		min := boltutil.Uint64ToBytes(fromTime)
		max := boltutil.Uint64ToBytes(toTime)
		c := timezoneBk.Cursor()
		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			walletStat := common.MetricStats{}
			if uErr := json.Unmarshal(v, &walletStat); uErr != nil {
				return uErr
			}
			key := boltutil.BytesToUint64(k) / 1000000
			result[key] = walletStat
		}
		return nil
	})
	return result, err
}

func (self *BoltStatStorage) SetCountry(country string) error {
	var err error
	country = strings.ToUpper(country)
	err = self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(countryBucket))
		if b == nil {
			return fmt.Errorf("cannot find bucket %s", countryBucket)
		}
		return b.Put([]byte(country), []byte("1"))
	})
	return err
}

// GetCountries returns a list of string representing all countries available in stat
// It return err if ther is no data.
func (self *BoltStatStorage) GetCountries() ([]string, error) {
	countries := []string{}
	var err error
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(countryBucket))
		if b == nil {
			return fmt.Errorf("cannot find bucket %s", countryBucket)
		}
		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			countries = append(countries, string(k))
		}
		return nil
	})
	return countries, err
}

func (self *BoltStatStorage) SetCountryStat(stats map[string]common.MetricStatsTimeZone, lastProcessTimePoint uint64) error {
	var err error
	err = self.db.Update(func(tx *bolt.Tx) error {
		for country, timeZoneStat := range stats {
			country = strings.ToUpper(country)
			b, uErr := tx.CreateBucketIfNotExists([]byte(country))
			if uErr != nil {
				return uErr
			}
			for i := stat.StartTimezone; i <= stat.EndTimezone; i++ {
				var (
					uErr     error
					dataJSON []byte
				)
				stats := timeZoneStat[i]
				freq := fmt.Sprintf("%s%d", stat.TimezoneBucketPrefix, i)
				countryTzBucket, uErr := b.CreateBucketIfNotExists([]byte(freq))
				if uErr != nil {
					return uErr
				}
				for timepoint, stat := range stats {
					timestamp := boltutil.Uint64ToBytes(timepoint)
					currentData := common.MetricStats{}
					v := countryTzBucket.Get(timestamp)
					if v != nil {
						if uErr = json.Unmarshal(v, &currentData); uErr != nil {
							return uErr
						}
					}
					currentData = addMetricData(currentData, stat)
					if dataJSON, uErr = json.Marshal(currentData); uErr != nil {
						return uErr
					}
					if uErr = countryTzBucket.Put(timestamp, dataJSON); uErr != nil {
						return uErr
					}
				}
			}
		}
		lastProcessBk := tx.Bucket([]byte(tradeLogProcessorState))
		if lastProcessBk == nil {
			return fmt.Errorf("cannot find Bucket %s", tradeLogProcessorState)
		}
		dataJSON := boltutil.Uint64ToBytes(lastProcessTimePoint)
		return lastProcessBk.Put([]byte(stat.CountryAggregation), dataJSON)
	})
	return err
}

// GetCountryStats returns StatTicks for a specific country in a specific time range.
// If the data is not available, return empty result and no error.
// If the data is corrupted, error is returned.
func (self *BoltStatStorage) GetCountryStats(fromTime, toTime uint64, country string, timezone int64) (common.StatTicks, error) {
	result := common.StatTicks{}
	tzstring := fmt.Sprintf("%s%d", stat.TimezoneBucketPrefix, timezone)
	country = strings.ToUpper(country)
	err := self.db.View(func(tx *bolt.Tx) error {
		countryBk := tx.Bucket([]byte(country))
		if countryBk == nil {
			log.Printf("GetCountryStats cannot find bucket %s", country)
			return nil
		}
		timezoneBk := countryBk.Bucket([]byte(tzstring))
		if timezoneBk == nil {
			log.Printf("GetCountryStats cannot find bucket %s->%s", country, tzstring)
			return nil
		}
		min := boltutil.Uint64ToBytes(fromTime)
		max := boltutil.Uint64ToBytes(toTime)

		c := timezoneBk.Cursor()
		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			countryStat := common.MetricStats{}
			if vErr := json.Unmarshal(v, &countryStat); vErr != nil {
				return vErr
			}
			key := boltutil.BytesToUint64(k) / 1000000
			result[key] = countryStat
		}
		return nil
	})
	return result, err
}

func (self *BoltStatStorage) DidTrade(b *bolt.Bucket, userAddr string, timepoint uint64) bool {
	result := false
	v := b.Get([]byte(userAddr))
	if v != nil {
		savedTimepoint := boltutil.BytesToUint64(v)
		if savedTimepoint <= timepoint {
			result = true
		}
	}
	return result
}

func (self *BoltStatStorage) SetFirstTradeEver(userTradeLog *[]common.TradeLog) error {
	err := self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(userFirstTradeEver))
		if b == nil {
			return fmt.Errorf("cannot find bucket %s", userFirstTradeEver)
		}
		for _, trade := range *userTradeLog {
			userAddr := common.AddrToString(trade.UserAddress)
			timepoint := trade.Timestamp
			if !self.DidTrade(b, userAddr, timepoint) {
				timestampByte := boltutil.Uint64ToBytes(timepoint)
				if pErr := b.Put([]byte(userAddr), timestampByte); pErr != nil {
					log.Printf("Cannot put data: %s", pErr.Error())
				}
			}
		}
		return nil
	})
	return err
}

// GetAllFirstTradeEver returns a map of ethereumAddress to the timepoint where that address first traded
// If the data is not available, return empty result and no error.
// If the wrapper bucket is not available, error is returned.
func (self *BoltStatStorage) GetAllFirstTradeEver() (map[ethereum.Address]uint64, error) {
	result := map[ethereum.Address]uint64{}
	err := self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(userFirstTradeEver))
		if b == nil {
			return fmt.Errorf("canot find bucket %s", userFirstTradeEver)
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			value := boltutil.BytesToUint64(v)
			result[ethereum.HexToAddress(string(k))] = value
		}
		return nil
	})
	return result, err
}

func (self *BoltStatStorage) DidTradeInDay(userDailyBucket *bolt.Bucket, userAddr string, timepoint uint64) bool {
	result := false
	v := userDailyBucket.Get([]byte(userAddr))
	if v != nil {
		savedTimepoint := boltutil.BytesToUint64(v)
		if savedTimepoint <= timepoint {
			result = true
		}
	}
	return result
}

// GetFirstTradeInDay return the timepoint when a User first trade in a certain day.
// It return error if the data can not be found.
func (self *BoltStatStorage) GetFirstTradeInDay(ethUserAddr ethereum.Address, timepoint uint64, timezone int64) (uint64, error) {
	var result uint64
	userAddr := common.AddrToString(ethUserAddr)
	err := self.db.View(func(tx *bolt.Tx) error {
		userStatBk := tx.Bucket([]byte(userStatBucket))
		if userStatBk == nil {
			return fmt.Errorf("cannot find bucket %s", userStatBucket)
		}
		freq := fmt.Sprintf("%s%d", stat.TimezoneBucketPrefix, timezone)
		timestamp := getTimestampByFreq(timepoint, freq)
		timezoneBk := userStatBk.Bucket(boltutil.Uint64ToBytes(uint64(timezone)))
		if timezoneBk == nil {
			return errors.New("cannot find data record for the input")
		}

		userDailyBucket := timezoneBk.Bucket(timestamp)
		if userDailyBucket == nil {
			return errors.New("cannot find data record for the input")
		}

		v := userDailyBucket.Get([]byte(userAddr))
		if v == nil {
			return errors.New("cannot find data record for the input")
		}
		result = boltutil.BytesToUint64(v)
		return nil
	})
	return result, err
}

func (self *BoltStatStorage) SetFirstTradeInDay(tradeLogs *[]common.TradeLog) error {
	err := self.db.Update(func(tx *bolt.Tx) error {
		userStatBk := tx.Bucket([]byte(userStatBucket))
		if userStatBk == nil {
			return fmt.Errorf("cannot find bucket %s", userStatBucket)
		}
		for _, trade := range *tradeLogs {
			userAddr := common.AddrToString(trade.UserAddress)
			timepoint := trade.Timestamp
			for timezone := stat.StartTimezone; timezone <= stat.EndTimezone; timezone++ {
				freq := fmt.Sprintf("%s%d", stat.TimezoneBucketPrefix, timezone)
				timestamp := getTimestampByFreq(timepoint, freq)
				timezoneBk, uErr := userStatBk.CreateBucketIfNotExists(boltutil.Uint64ToBytes(uint64(timezone)))
				if uErr != nil {
					return uErr
				}
				userDailyBucket, uErr := timezoneBk.CreateBucketIfNotExists(timestamp)
				if uErr != nil {
					return uErr
				}
				if !self.DidTradeInDay(userDailyBucket, userAddr, timepoint) {
					timestampByte := boltutil.Uint64ToBytes(timepoint)
					if pErr := userDailyBucket.Put([]byte(userAddr), timestampByte); pErr != nil {
						log.Printf("Cannot put user daily first trade: %s", pErr.Error())
					}
				}
			}
		}
		return nil
	})
	return err
}

func (self *BoltStatStorage) SetUserList(userInfos map[string]common.UserInfoTimezone, lastProcessTimePoint uint64) error {
	err := self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(userListBucket))
		if b == nil {
			return fmt.Errorf("cannot find bucket %s", userListBucket)
		}
		for userAddr, userInfoData := range userInfos {
			var (
				timezoneBk *bolt.Bucket
				userAddr   = strings.ToLower(userAddr)
				uErr       error
			)
			for timezone := stat.StartTimezone; timezone <= stat.EndTimezone; timezone++ {
				freq := fmt.Sprintf("%s%d", stat.TimezoneBucketPrefix, timezone)
				if timezoneBk, uErr = b.CreateBucketIfNotExists([]byte(freq)); uErr != nil {
					return uErr
				}
				timezoneData := userInfoData[timezone]
				for timepoint, userData := range timezoneData {
					var timestampBk *bolt.Bucket
					timestamp := getTimestampByFreq(timepoint, freq)
					if timestampBk, uErr = timezoneBk.CreateBucketIfNotExists(timestamp); uErr != nil {
						return uErr
					}
					currentUserData := common.UserInfo{}
					currentValue := timestampBk.Get([]byte(userAddr))
					if currentValue != nil {
						if uErr = json.Unmarshal(currentValue, &currentUserData); uErr != nil {
							return uErr
						}
					}
					currentUserData.USDVolume += userData.USDVolume
					currentUserData.ETHVolume += userData.ETHVolume
					currentUserData.Addr = userData.Addr
					currentUserData.Email = userData.Email
					dataJSON, uErr := json.Marshal(currentUserData)
					if uErr != nil {
						return uErr
					}
					if uErr = timestampBk.Put([]byte(userAddr), dataJSON); uErr != nil {
						return uErr
					}
				}
			}
		}
		lastProcessBk := tx.Bucket([]byte(tradeLogProcessorState))
		if lastProcessBk == nil {
			return fmt.Errorf("cannot find Bucket %s", tradeLogProcessorState)
		}
		dataJSON := boltutil.Uint64ToBytes(lastProcessTimePoint)
		return lastProcessBk.Put([]byte(stat.UserInfoAggregation), dataJSON)
	})
	return err
}

// GetUserList returns a map of user address to UserInfo in a specific time range.
// If the data is not available, return empty result and no error.
// If the data is corrupted or wrapper bucket is not found, error is returned.
func (self *BoltStatStorage) GetUserList(fromTime, toTime uint64, timezone int64) (map[string]common.UserInfo, error) {
	result := map[string]common.UserInfo{}
	err := self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(userListBucket))
		if b == nil {
			return fmt.Errorf("GetUserList cannot find bucket %s", userListBucket)
		}
		min := boltutil.Uint64ToBytes(fromTime)
		max := boltutil.Uint64ToBytes(toTime)

		timezoneBkName := fmt.Sprintf("%s%d", stat.TimezoneBucketPrefix, timezone)
		timezoneBk := b.Bucket([]byte(timezoneBkName))
		if timezoneBk == nil {
			log.Printf("GetUserList cannot find bucket %s->%s", userListBucket, timezoneBkName)
			return nil
		}
		c := timezoneBk.Cursor()
		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			if v == nil {
				timestampBk := timezoneBk.Bucket(k)
				if timestampBk == nil {
					log.Printf("GetUserList cannot find bucket %s->%s->%d", userListBucket, timezoneBkName, boltutil.BytesToUint64(k))
					return nil
				}

				cursor := timestampBk.Cursor()
				for kk, vv := cursor.First(); kk != nil; kk, vv = cursor.Next() {
					value := common.UserInfo{}
					if uErr := json.Unmarshal(vv, &value); uErr != nil {
						return uErr
					}
					currentData, exist := result[value.Addr]
					if !exist {
						currentData = common.UserInfo{
							Email: value.Email,
							Addr:  value.Addr,
						}
					}
					currentData.ETHVolume += value.ETHVolume
					currentData.USDVolume += value.USDVolume
					result[value.Addr] = currentData
				}
			}
		}
		return nil
	})
	return result, err
}

//GetAssetVolume returns stat data for an asset address with specific frequency in a specific time range.
//If the data is not available, return empty result and no error.
//If the data is corrupted , error is returned.
func (self *BoltStatStorage) GetAssetVolume(fromTime uint64, toTime uint64, freq string, ethAssetAddr ethereum.Address) (common.StatTicks, error) {
	result := common.StatTicks{}
	assetAddr := common.AddrToString(ethAssetAddr)
	err := self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(assetAddr))
		if b == nil {
			log.Printf("GetAssetVolume cannot find bucket %s", assetAddr)
			return nil
		}
		freqBkName, vErr := getBucketNameByFreq(freq)
		if vErr != nil {
			return vErr
		}
		min := boltutil.Uint64ToBytes(fromTime)
		max := boltutil.Uint64ToBytes(toTime)

		freqBk := b.Bucket([]byte(freqBkName))
		if freqBk == nil {
			log.Printf("GetAssetVolume cannot find bucket %s->%s", assetAddr, freqBkName)
			return nil
		}
		c := freqBk.Cursor()
		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			value := common.VolumeStats{}
			if vErr := json.Unmarshal(v, &value); vErr != nil {
				return vErr
			}
			key := boltutil.BytesToUint64(k) / 1000000
			result[key] = value
		}
		return nil
	})
	return result, err
}

//GetBurnFee returns stat data for an address with specific frequency in a specific time range.
//If the data is not available, return empty result and no error.
//If the data is corrupted , error is returned.
func (self *BoltStatStorage) GetBurnFee(fromTime uint64, toTime uint64, freq string, ethReserveAddr ethereum.Address) (common.StatTicks, error) {
	result := common.StatTicks{}
	reserveAddr := common.AddrToString(ethReserveAddr)
	err := self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(reserveAddr))
		if b == nil {
			log.Printf("GetBurnFee cannot find bucket %s", reserveAddr)
			return nil
		}
		freqBkName, vErr := getBucketNameByFreq(freq)
		if vErr != nil {
			return vErr
		}

		freqBk := b.Bucket([]byte(freqBkName))
		if freqBk == nil {
			log.Printf("GetBurnFee cannot find bucket %s->%s", reserveAddr, freqBkName)
			return nil
		}
		min := boltutil.Uint64ToBytes(fromTime)
		max := boltutil.Uint64ToBytes(toTime)

		c := freqBk.Cursor()
		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			value := common.BurnFeeStats{}
			if vErr := json.Unmarshal(v, &value); vErr != nil {
				return vErr
			}
			key := boltutil.BytesToUint64(k) / 1000000
			result[key] = value.TotalBurnFee
		}
		return nil
	})
	return result, err
}

//GetWalletFee returns stat data for an address pair with specific frequency in a specific time range.
//If the data is not available, return empty result and no error.
//If the data is corrupted , error is returned.
func (self *BoltStatStorage) GetWalletFee(fromTime uint64, toTime uint64, freq string, reserveAddr ethereum.Address, walletAddr ethereum.Address) (common.StatTicks, error) {
	result := common.StatTicks{}

	err := self.db.View(func(tx *bolt.Tx) error {
		bucketName := fmt.Sprintf("%s_%s", common.AddrToString(reserveAddr), common.AddrToString(walletAddr))
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			log.Printf("GetWalletFee cannot find bucket %s", bucketName)
			return nil
		}
		freqBkName, vErr := getBucketNameByFreq(freq)
		if vErr != nil {
			return vErr
		}
		freqBk := b.Bucket([]byte(freqBkName))
		if freqBk == nil {
			log.Printf("GetWalletFee cannot find bucket %s->%s", bucketName, freqBkName)
			return nil
		}

		min := boltutil.Uint64ToBytes(fromTime)
		max := boltutil.Uint64ToBytes(toTime)

		c := freqBk.Cursor()
		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			value := common.BurnFeeStats{}
			if vErr := json.Unmarshal(v, &value); vErr != nil {
				return vErr
			}
			key := boltutil.BytesToUint64(k) / 1000000
			result[key] = value.TotalBurnFee
		}
		return nil
	})

	return result, err
}

//GetUserVolume returns stat data for a user Address with specific frequency in a specific time range.
//If the data is not available, return empty result and no error.
//If the data is corrupted , error is returned.
func (self *BoltStatStorage) GetUserVolume(fromTime uint64, toTime uint64, freq string, ethUserAddr ethereum.Address) (common.StatTicks, error) {
	result := common.StatTicks{}
	userAddr := common.AddrToString(ethUserAddr)
	err := self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(userAddr))
		if b == nil {
			log.Printf("GetUserVolume cannot find bucket %s", userAddr)
			return nil
		}
		freqBkName, vErr := getBucketNameByFreq(freq)
		if vErr != nil {
			return vErr
		}
		freqBk := b.Bucket([]byte(freqBkName))
		if freqBk == nil {
			log.Printf("GetUserVolume cannot find bucket %s->%s", userAddr, freqBkName)
		}
		min := boltutil.Uint64ToBytes(fromTime)
		max := boltutil.Uint64ToBytes(toTime)
		c := freqBk.Cursor()
		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			value := common.VolumeStats{}
			if vErr := json.Unmarshal(v, &value); vErr != nil {
				return vErr
			}
			key := boltutil.BytesToUint64(k) / 1000000
			result[key] = value
		}
		return nil
	})
	return result, err
}

//GetReserveVolume returns stat data for a reserve Address with specific frequency in a specific time range.
//If the data is not available, return empty result and no error.
//If the data is corrupted , error is returned.
func (self *BoltStatStorage) GetReserveVolume(fromTime uint64, toTime uint64, freq string, reserveAddr, token ethereum.Address) (common.StatTicks, error) {
	result := common.StatTicks{}
	err := self.db.View(func(tx *bolt.Tx) error {
		bucketKey := fmt.Sprintf("%s_%s", common.AddrToString(reserveAddr), common.AddrToString(token))
		b := tx.Bucket([]byte(bucketKey))
		if b == nil {
			log.Printf("GetReserveVolume cannot find bucket %s", bucketKey)
			return nil
		}
		freqBkName, vErr := getBucketNameByFreq(freq)
		if vErr != nil {
			return vErr
		}
		freqBk := b.Bucket([]byte(freqBkName))
		if freqBk == nil {
			log.Printf("GetReserveVolume cannot find bucket %s->%s", bucketKey, freqBkName)
			return nil
		}
		min := boltutil.Uint64ToBytes(fromTime)
		max := boltutil.Uint64ToBytes(toTime)
		c := freqBk.Cursor()
		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			value := common.VolumeStats{}
			if vErr := json.Unmarshal(v, &value); vErr != nil {
				return vErr
			}
			key := boltutil.BytesToUint64(k) / 1000000
			result[key] = value
		}
		return nil
	})
	return result, err
}

//GetTokenHeatmap returns stat data for a country as key with specific frequency in a specific time range.
//If the data is not available, return empty result and no error.
//If the data is corrupted , error is returned.
func (self *BoltStatStorage) GetTokenHeatmap(fromTime, toTime uint64, key, freq string) (common.StatTicks, error) {
	result := common.StatTicks{}
	err := self.db.View(func(tx *bolt.Tx) error {
		nkey := strings.ToLower(key)
		b := tx.Bucket([]byte(nkey))
		if b == nil {
			log.Printf("GetTokenHeatmap cannot find bucket %s", nkey)
			return nil
		}
		freqBkName, vErr := getBucketNameByFreq(freq)
		if vErr != nil {
			return vErr
		}

		freqBk := b.Bucket([]byte(freqBkName))
		if freqBk == nil {
			log.Printf("GetTokenHeatmap cannot find bucket %s->%s", nkey, freqBkName)
		}

		min := boltutil.Uint64ToBytes(fromTime)
		max := boltutil.Uint64ToBytes(toTime)
		c := freqBk.Cursor()
		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			value := common.VolumeStats{}
			if vErr := json.Unmarshal(v, &value); vErr != nil {
				return vErr
			}
			key := boltutil.BytesToUint64(k) / 1000000
			result[key] = value
		}
		return nil
	})
	return result, err
}

func (self *BoltStatStorage) SetTradeSummary(tradeSummary map[string]common.MetricStatsTimeZone, lastProcessTimePoint uint64) error {
	var err error
	err = self.db.Update(func(tx *bolt.Tx) error {
		for key, stats := range tradeSummary {
			key = strings.ToLower(key)
			b, uErr := tx.CreateBucketIfNotExists([]byte(key))
			if uErr != nil {
				return uErr
			}
			// update to timezone buckets
			for i := stat.StartTimezone; i <= stat.EndTimezone; i++ {
				stats := stats[i]
				freq := fmt.Sprintf("%s%d", stat.TimezoneBucketPrefix, i)
				tzBucket, uErr := b.CreateBucketIfNotExists([]byte(freq))
				if uErr != nil {
					return uErr
				}
				for timepoint, stat := range stats {
					timestamp := boltutil.Uint64ToBytes(timepoint)
					// try get data from this timestamp, if exist then add more data
					currentData := common.MetricStats{}
					v := tzBucket.Get(timestamp)
					if v != nil {
						if uErr := json.Unmarshal(v, &currentData); uErr != nil {
							return uErr
						}
					}
					currentData = addMetricData(currentData, stat)
					dataJSON, uErr := json.Marshal(currentData)
					if uErr != nil {
						return uErr
					}
					if uErr = tzBucket.Put(timestamp, dataJSON); uErr != nil {
						return uErr
					}
				}
			}
		}
		lastProcessBk := tx.Bucket([]byte(tradeLogProcessorState))
		if lastProcessBk == nil {
			return fmt.Errorf("cannot find Bucket %s", tradeLogProcessorState)
		}
		dataJSON := boltutil.Uint64ToBytes(lastProcessTimePoint)
		return lastProcessBk.Put([]byte(stat.TradeSummaryAggregation), dataJSON)
	})
	return err
}

//GetTradeSummary returns summary stat data for a timezone a specific time range.
//If the data is not available, return empty result and no error.
//If the data is corrupted , error is returned.
func (self *BoltStatStorage) GetTradeSummary(fromTime uint64, toTime uint64, timezone int64) (common.StatTicks, error) {
	result := common.StatTicks{}
	tzstring := fmt.Sprintf("%s%d", stat.TimezoneBucketPrefix, timezone)
	err := self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(stat.TradeSummaryKey))
		if b == nil {
			log.Printf("GetTradeSummary cannot find bucket %s", stat.TradeSummaryKey)
			return nil
		}
		timezoneBk := b.Bucket([]byte(tzstring))
		if timezoneBk == nil {
			log.Printf("GetTradeSummary cannot find bucket %s->%s", stat.TradeSummaryKey, tzstring)
			return nil
		}
		min := boltutil.Uint64ToBytes(fromTime)
		max := boltutil.Uint64ToBytes(toTime)

		c := timezoneBk.Cursor()
		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			summary := common.MetricStats{}
			if vErr := json.Unmarshal(v, &summary); vErr != nil {
				return vErr
			}
			key := boltutil.BytesToUint64(k) / 1000000
			result[key] = summary
		}
		return nil
	})

	return result, err
}
