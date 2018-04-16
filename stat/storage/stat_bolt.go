package storage

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/boltdb/bolt"
	ethereum "github.com/ethereum/go-ethereum/common"
)

const (
	MAX_NUMBER_VERSION int = 1000

	TRADELOG_PROCESSOR_STATE string = "tradelog_processor_state"

	MINUTE_BUCKET               string = "minute"
	HOUR_BUCKET                 string = "hour"
	DAY_BUCKET                  string = "day"
	TIMEZONE_BUCKET_PREFIX      string = "utc"
	START_TIMEZONE              int64  = -11
	END_TIMEZONE                int64  = 14
	ADDRESS_BUCKET_PREFIX       string = "address"
	DAILY_ADDRESS_BUCKET_PREFIX string = "daily_address"
	DAILY_USER_BUCKET_PREFIX    string = "daily_user"
	WALLET_ADDRESS_BUCKET       string = "wallet_address"
	RESERVE_RATES               string = "reserve_rates"
	EXPIRED                     uint64 = uint64(7 * time.Hour * 24)
	COUNTRY_BUCKET              string = "country_stat_bucket"
	USER_FIRST_TRADE_EVER       string = "user_first_trade_ever"
	USER_STAT_BUCKET            string = "user_stat_bucket"
	VOLUME_STAT_BUCKET          string = "volume_stat_bucket"

	TRADE_SUMMARY_AGGREGATION string = "trade_summary_aggregation"
	WALLET_AGGREGATION        string = "wallet_aggregation"
	COUNTRY_AGGREGATION       string = "country_aggregation"
	USER_AGGREGATION          string = "user_aggregation"
	VOLUME_STAT_AGGREGATION   string = "volume_stat_aggregation"
	BURNFEE_AGGREGATION       string = "burn_fee_aggregation"
)

type BoltStatStorage struct {
	db *bolt.DB
}

func uint64ToBytes(u uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, u)
	return b
}

func bytesToUint64(b []byte) uint64 {
	if len(b) != 8 {
		return 0
	}
	return binary.BigEndian.Uint64(b)
}

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
		_, err = tx.CreateBucketIfNotExists([]byte(TRADELOG_PROCESSOR_STATE))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(WALLET_ADDRESS_BUCKET))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(COUNTRY_BUCKET))
		if err != nil {
			return err
		}

		return err
	})
	if err != nil {
		return nil, err
	}
	storage := &BoltStatStorage{db}
	return storage, nil
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

func (self *BoltStatStorage) SetLastProcessedTradeLogTimepoint(statType string, timepoint uint64) error {
	var err error
	err = self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TRADELOG_PROCESSOR_STATE))
		err = b.Put([]byte(statType), uint64ToBytes(timepoint))
		return err
	})
	return err
}

func (self *BoltStatStorage) GetLastProcessedTradeLogTimepoint(statType string) (uint64, error) {
	var result uint64
	var err error
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TRADELOG_PROCESSOR_STATE))
		if b == nil {
			return (errors.New("Can not find last processed bucket"))
		}
		result = bytesToUint64(b.Get([]byte(statType)))
		return nil
	})
	return result, err
}

// GetNumberOfVersion return number of version storing in a bucket
func (self *BoltStatStorage) GetNumberOfVersion(tx *bolt.Tx, bucket string) int {
	result := 0
	b := tx.Bucket([]byte(bucket))
	c := b.Cursor()
	for k, _ := c.First(); k != nil; k, _ = c.Next() {
		result++
	}
	return result
}

// PruneOutdatedData Remove first version out of database
func (self *BoltStatStorage) PruneOutdatedData(tx *bolt.Tx, bucket string) error {
	var err error
	b := tx.Bucket([]byte(bucket))
	c := b.Cursor()
	for self.GetNumberOfVersion(tx, bucket) >= MAX_NUMBER_VERSION {
		k, _ := c.First()
		if k == nil {
			err = errors.New(fmt.Sprintf("There no version in %s", bucket))
			return err
		}
		err = b.Delete([]byte(k))
		if err != nil {
			panic(err)
		}
	}
	return err
}

func getBucketNameByFreq(freq string) (bucketName string, err error) {
	switch freq {
	case "m", "M":
		bucketName = MINUTE_BUCKET
	case "h", "H":
		bucketName = HOUR_BUCKET
	case "d", "D":
		bucketName = DAY_BUCKET
	default:
		offset, ok := strconv.ParseInt(strings.TrimPrefix(freq, "utc"), 10, 64)
		if (offset < START_TIMEZONE) || (offset > END_TIMEZONE) {
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
		result = uint64ToBytes(t / uint64(time.Minute) * uint64(time.Minute))
	case "h", "H":
		result = uint64ToBytes(t / uint64(time.Hour) * uint64(time.Hour))
	case "d", "D":
		result = uint64ToBytes(t / ui64Day * ui64Day)
	default:
		// utc timezone
		offset, _ := strconv.ParseInt(strings.TrimPrefix(freq, "utc"), 10, 64)
		ui64offset := uint64(int64(time.Hour) * offset)
		if offset > 0 {
			result = uint64ToBytes((t+ui64offset)/ui64Day*ui64Day + ui64offset)
		} else {
			offset = 0 - offset
			result = uint64ToBytes((t-ui64offset)/ui64Day*ui64Day - ui64offset)
		}
	}

	return
}

func (self *BoltStatStorage) SetWalletAddress(walletAddr ethereum.Address) (err error) {
	err = self.db.Update(func(tx *bolt.Tx) error {
		walletBucket := tx.Bucket([]byte(WALLET_ADDRESS_BUCKET))

		if err := walletBucket.Put([]byte(common.AddrToString(walletAddr)), []byte("1")); err != nil {
			return err
		}

		return nil
	})
	return
}

func (self *BoltStatStorage) GetWalletAddress() ([]string, error) {
	var result []string
	var err error
	err = self.db.View(func(tx *bolt.Tx) error {
		walletBucket := tx.Bucket([]byte(WALLET_ADDRESS_BUCKET))
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
			burnFeeBk, _ := tx.CreateBucketIfNotExists([]byte(key))
			for _, freq := range []string{"M", "H", "D"} {
				stats := timezoneData[freq]
				freqBkName, _ := getBucketNameByFreq(freq)
				freqBk, _ := burnFeeBk.CreateBucketIfNotExists([]byte(freqBkName))
				for timepoint, stat := range stats {
					timestamp := uint64ToBytes(timepoint)
					currentData := common.BurnFeeStats{}
					v := freqBk.Get(timestamp)
					if v != nil {
						json.Unmarshal(v, &currentData)
					}
					currentData.TotalBurnFee += stat.TotalBurnFee

					dataJSON, _ := json.Marshal(currentData)
					freqBk.Put(timestamp, dataJSON)
				}
			}
		}
		lastProcessBk := tx.Bucket([]byte(TRADELOG_PROCESSOR_STATE))
		if lastProcessBk != nil {
			dataJSON := uint64ToBytes(lastProcessTimePoint)
			lastProcessBk.Put([]byte(BURNFEE_AGGREGATION), dataJSON)
		}
		return nil
	})
	return err
}

func (self *BoltStatStorage) SetVolumeStat(volumeStats map[ethereum.Address]common.VolumeStatsTimeZone, lastProcessTimePoint uint64) error {
	err := self.db.Update(func(tx *bolt.Tx) error {
		for asset, freqData := range volumeStats {
			volumeBk, _ := tx.CreateBucketIfNotExists([]byte(common.AddrToString(asset)))
			for _, freq := range []string{"M", "H", "D"} {
				stats := freqData[freq]
				freqBkName, _ := getBucketNameByFreq(freq)
				freqBk, _ := volumeBk.CreateBucketIfNotExists([]byte(freqBkName))
				for timepoint, stat := range stats {
					timestamp := uint64ToBytes(timepoint)
					currentData := common.VolumeStats{}
					v := freqBk.Get(timestamp)
					if v != nil {
						json.Unmarshal(v, &currentData)
					}
					currentData.ETHVolume += stat.ETHVolume
					currentData.USDAmount += stat.USDAmount
					currentData.Volume += stat.Volume

					dataJSON, _ := json.Marshal(currentData)
					freqBk.Put(timestamp, dataJSON)
				}
			}
		}
		lastProcessBk := tx.Bucket([]byte(TRADELOG_PROCESSOR_STATE))
		if lastProcessBk != nil {
			dataJSON := uint64ToBytes(lastProcessTimePoint)
			lastProcessBk.Put([]byte(VOLUME_STAT_AGGREGATION), dataJSON)
		}
		return nil
	})
	return err
}

func (self *BoltStatStorage) SetWalletStat(stats map[string]common.MetricStatsTimeZone, lastProcessTimePoint uint64) error {
	err := self.db.Update(func(tx *bolt.Tx) error {
		for wallet, timeZoneStat := range stats {
			wallet = strings.ToLower(wallet)
			b, err := tx.CreateBucketIfNotExists([]byte(wallet))
			if err != nil {
				return err
			}
			for i := START_TIMEZONE; i <= END_TIMEZONE; i++ {
				stats := timeZoneStat[i]
				freq := fmt.Sprintf("%s%d", TIMEZONE_BUCKET_PREFIX, i)
				walletTzBucket, err := b.CreateBucketIfNotExists([]byte(freq))
				if err != nil {
					return err
				}
				for timepoint, stat := range stats {
					timestamp := uint64ToBytes(timepoint)
					currentData := common.MetricStats{}
					v := walletTzBucket.Get(timestamp)
					if v != nil {
						json.Unmarshal(v, &currentData)
					}
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
					dataJSON, err := json.Marshal(currentData)
					if err != nil {
						return err
					}
					walletTzBucket.Put(timestamp, dataJSON)
				}
			}
		}
		lastProcessBk := tx.Bucket([]byte(TRADELOG_PROCESSOR_STATE))
		if lastProcessBk != nil {
			dataJSON := uint64ToBytes(lastProcessTimePoint)
			lastProcessBk.Put([]byte(WALLET_AGGREGATION), dataJSON)
		}
		return nil
	})
	return err
}

func (self *BoltStatStorage) GetWalletStats(fromTime uint64, toTime uint64, walletAddr ethereum.Address, timezone int64) (common.StatTicks, error) {

	result := common.StatTicks{}
	tzstring := fmt.Sprintf("%s%d", TIMEZONE_BUCKET_PREFIX, timezone)
	log.Printf("Get wallet stats: %s and %d", common.AddrToString(walletAddr), timezone)
	err := self.db.Update(func(tx *bolt.Tx) error {
		walletBk, _ := tx.CreateBucketIfNotExists([]byte(common.AddrToString(walletAddr)))
		timezoneBk, _ := walletBk.CreateBucketIfNotExists([]byte(tzstring))
		min := uint64ToBytes(fromTime)
		max := uint64ToBytes(toTime)
		c := timezoneBk.Cursor()
		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			walletStat := common.MetricStats{}
			json.Unmarshal(v, &walletStat)
			key := bytesToUint64(k) / 1000000
			result[key] = walletStat
		}
		return nil
	})
	return result, err
}

func (self *BoltStatStorage) SetCountry(country string) error {
	var err error
	country = strings.ToLower(country)
	err = self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(COUNTRY_BUCKET))
		err = b.Put([]byte(country), []byte("1"))
		return err
	})
	return err
}

func (self *BoltStatStorage) GetCountries() ([]string, error) {
	countries := []string{}
	var err error
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(COUNTRY_BUCKET))
		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			countries = append(countries, string(k))
		}
		return err
	})
	return countries, err
}

func (self *BoltStatStorage) SetCountryStat(stats map[string]common.MetricStatsTimeZone, lastProcessTimePoint uint64) error {
	var err error
	err = self.db.Update(func(tx *bolt.Tx) error {
		for country, timeZoneStat := range stats {
			country = strings.ToLower(country)
			b, err := tx.CreateBucketIfNotExists([]byte(country))
			if err != nil {
				return err
			}
			for i := START_TIMEZONE; i <= END_TIMEZONE; i++ {
				stats := timeZoneStat[i]
				freq := fmt.Sprintf("%s%d", TIMEZONE_BUCKET_PREFIX, i)
				countryTzBucket, err := b.CreateBucketIfNotExists([]byte(freq))
				if err != nil {
					return err
				}
				for timepoint, stat := range stats {
					timestamp := uint64ToBytes(timepoint)
					currentData := common.MetricStats{}
					v := countryTzBucket.Get(timestamp)
					if v != nil {
						json.Unmarshal(v, &currentData)
					}
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
					dataJSON, err := json.Marshal(currentData)
					if err != nil {
						return err
					}
					countryTzBucket.Put(timestamp, dataJSON)
				}
			}
		}
		lastProcessBk := tx.Bucket([]byte(TRADELOG_PROCESSOR_STATE))
		if lastProcessBk != nil {
			dataJSON := uint64ToBytes(lastProcessTimePoint)
			lastProcessBk.Put([]byte(COUNTRY_AGGREGATION), dataJSON)
		}
		return nil
	})
	return err
}

func (self *BoltStatStorage) GetCountryStats(fromTime, toTime uint64, country string, timezone int64) (common.StatTicks, error) {
	result := common.StatTicks{}
	country = strings.ToLower(country)
	tzstring := fmt.Sprintf("%s%d", TIMEZONE_BUCKET_PREFIX, timezone)
	self.db.Update(func(tx *bolt.Tx) error {
		countryBk, _ := tx.CreateBucketIfNotExists([]byte(country))
		timezoneBk, _ := countryBk.CreateBucketIfNotExists([]byte(tzstring))
		min := uint64ToBytes(fromTime)
		max := uint64ToBytes(toTime)

		c := timezoneBk.Cursor()
		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			countryStat := common.MetricStats{}
			json.Unmarshal(v, &countryStat)
			key := bytesToUint64(k) / 1000000
			result[key] = countryStat
		}
		return nil
	})

	return result, nil
}

func (self *BoltStatStorage) DidTrade(tx *bolt.Tx, userAddr string, timepoint uint64) bool {
	result := false
	b, _ := tx.CreateBucketIfNotExists([]byte(USER_FIRST_TRADE_EVER))
	v := b.Get([]byte(userAddr))
	if v != nil {
		savedTimepoint := bytesToUint64(v)
		if savedTimepoint <= timepoint {
			result = true
		}
	}
	return result
}

func (self *BoltStatStorage) SetFirstTradeEver(userTradeLog *[]common.TradeLog, lastProcessTimePoint uint64) error {
	err := self.db.Update(func(tx *bolt.Tx) error {
		b, _ := tx.CreateBucketIfNotExists([]byte(USER_FIRST_TRADE_EVER))
		for _, trade := range *userTradeLog {
			userAddr := common.AddrToString(trade.UserAddress)
			timepoint := trade.Timestamp
			if !self.DidTrade(tx, userAddr, timepoint) {
				timestampByte := uint64ToBytes(timepoint)
				b.Put([]byte(userAddr), timestampByte)
			}
		}
		lastProcessBk := tx.Bucket([]byte(TRADELOG_PROCESSOR_STATE))
		if lastProcessBk != nil {
			dataJSON := uint64ToBytes(lastProcessTimePoint)
			lastProcessBk.Put([]byte(USER_AGGREGATION), dataJSON)
		}
		return nil
	})
	return err
}

func (self *BoltStatStorage) GetFirstTradeEver(userAddr ethereum.Address) (uint64, error) {
	result := uint64(0)
	err := self.db.Update(func(tx *bolt.Tx) error {
		b, _ := tx.CreateBucketIfNotExists([]byte(USER_FIRST_TRADE_EVER))
		v := b.Get([]byte(common.AddrToString(userAddr)))
		result = bytesToUint64(v)
		return nil
	})
	return result, err
}

func (self *BoltStatStorage) GetAllFirstTradeEver() (map[ethereum.Address]uint64, error) {
	result := map[ethereum.Address]uint64{}
	err := self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(USER_FIRST_TRADE_EVER))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			value := bytesToUint64(v)
			result[ethereum.HexToAddress(string(k))] = value
		}
		return nil
	})
	return result, err
}

func (self *BoltStatStorage) DidTradeInDay(tx *bolt.Tx, userAddr string, timepoint uint64, timezone int64) bool {
	result := false
	userStatBk, _ := tx.CreateBucketIfNotExists([]byte(USER_STAT_BUCKET))
	freq := fmt.Sprintf("%s%d", TIMEZONE_BUCKET_PREFIX, timezone)
	timestamp := getTimestampByFreq(timepoint, freq)

	timezoneBk, _ := userStatBk.CreateBucketIfNotExists(uint64ToBytes(uint64(timezone)))
	userDailyBucket, _ := timezoneBk.CreateBucketIfNotExists(timestamp)

	v := userDailyBucket.Get([]byte(userAddr))
	if v != nil {
		savedTimepoint := bytesToUint64(v)
		if savedTimepoint <= timepoint {
			result = true
		}
	}
	return result
}

func (self *BoltStatStorage) GetFirstTradeInDay(userAddr ethereum.Address, timepoint uint64, timezone int64) (uint64, error) {
	result := uint64(0)
	err := self.db.Update(func(tx *bolt.Tx) error {
		userStatBk, _ := tx.CreateBucketIfNotExists([]byte(USER_STAT_BUCKET))
		freq := fmt.Sprintf("%s%d", TIMEZONE_BUCKET_PREFIX, timezone)
		timestamp := getTimestampByFreq(timepoint, freq)

		timezoneBk, _ := userStatBk.CreateBucketIfNotExists(uint64ToBytes(uint64(timezone)))
		userDailyBucket, _ := timezoneBk.CreateBucketIfNotExists(timestamp)

		v := userDailyBucket.Get([]byte(common.AddrToString(userAddr)))
		if v != nil {
			result = bytesToUint64(v)
		}
		return nil
	})
	return result, err
}

func (self *BoltStatStorage) SetFirstTradeInDay(tradeLogs *[]common.TradeLog, lastProcessTimePoint uint64) error {
	err := self.db.Update(func(tx *bolt.Tx) error {
		userStatBk, _ := tx.CreateBucketIfNotExists([]byte(USER_STAT_BUCKET))
		for _, trade := range *tradeLogs {
			userAddr := common.AddrToString(trade.UserAddress)
			timepoint := trade.Timestamp
			for timezone := START_TIMEZONE; timezone <= END_TIMEZONE; timezone++ {
				freq := fmt.Sprintf("%s%d", TIMEZONE_BUCKET_PREFIX, timezone)
				timestamp := getTimestampByFreq(timepoint, freq)
				timezoneBk, _ := userStatBk.CreateBucketIfNotExists(uint64ToBytes(uint64(timezone)))
				userDailyBucket, _ := timezoneBk.CreateBucketIfNotExists(timestamp)
				if !self.DidTradeInDay(tx, userAddr, timepoint, timezone) {
					timestampByte := uint64ToBytes(timepoint)
					userDailyBucket.Put([]byte(userAddr), timestampByte)
				}
			}
		}
		lastProcessBk := tx.Bucket([]byte(TRADELOG_PROCESSOR_STATE))
		if lastProcessBk != nil {
			dataJSON := uint64ToBytes(lastProcessTimePoint)
			lastProcessBk.Put([]byte(USER_AGGREGATION), dataJSON)
		}
		return nil
	})
	return err
}

func (self *BoltStatStorage) GetAssetVolume(fromTime uint64, toTime uint64, freq string, assetAddr ethereum.Address) (common.StatTicks, error) {
	result := common.StatTicks{}
	err := self.db.Update(func(tx *bolt.Tx) error {
		b, _ := tx.CreateBucketIfNotExists([]byte(common.AddrToString(assetAddr)))

		freqBkName, _ := getBucketNameByFreq(freq)
		min := uint64ToBytes(fromTime)
		max := uint64ToBytes(toTime)

		freqBk, _ := b.CreateBucketIfNotExists([]byte(freqBkName))
		c := freqBk.Cursor()
		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			value := common.VolumeStats{}
			json.Unmarshal(v, &value)
			key := bytesToUint64(k) / 1000000
			result[key] = value
		}

		return nil
	})
	return result, err
}

func (self *BoltStatStorage) GetBurnFee(fromTime uint64, toTime uint64, freq string, reserveAddr ethereum.Address) (common.StatTicks, error) {
	result := common.StatTicks{}
	err := self.db.Update(func(tx *bolt.Tx) error {
		b, _ := tx.CreateBucketIfNotExists([]byte(common.AddrToString(reserveAddr)))
		freqBkName, _ := getBucketNameByFreq(freq)

		freqBk, _ := b.CreateBucketIfNotExists([]byte(freqBkName))

		min := uint64ToBytes(fromTime)
		max := uint64ToBytes(toTime)

		c := freqBk.Cursor()
		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			value := common.BurnFeeStats{}
			json.Unmarshal(v, &value)
			key := bytesToUint64(k) / 1000000
			result[key] = value.TotalBurnFee
		}
		return nil
	})
	return result, err
}

func (self *BoltStatStorage) GetWalletFee(fromTime uint64, toTime uint64, freq string, reserveAddr ethereum.Address, walletAddr ethereum.Address) (common.StatTicks, error) {
	result := common.StatTicks{}

	err := self.db.Update(func(tx *bolt.Tx) error {
		bucketName := fmt.Sprintf("%s_%s", common.AddrToString(reserveAddr), common.AddrToString(walletAddr))
		b, _ := tx.CreateBucketIfNotExists([]byte(bucketName))
		freqBkName, _ := getBucketNameByFreq(freq)
		freqBk, _ := b.CreateBucketIfNotExists([]byte(freqBkName))

		min := uint64ToBytes(fromTime)
		max := uint64ToBytes(toTime)

		c := freqBk.Cursor()
		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			value := common.BurnFeeStats{}
			json.Unmarshal(v, &value)
			key := bytesToUint64(k) / 1000000
			result[key] = value.TotalBurnFee
		}
		return nil
	})

	return result, err
}

func (self *BoltStatStorage) GetUserVolume(fromTime uint64, toTime uint64, freq string, userAddr ethereum.Address) (common.StatTicks, error) {
	result := common.StatTicks{}
	err := self.db.Update(func(tx *bolt.Tx) error {
		b, _ := tx.CreateBucketIfNotExists([]byte(common.AddrToString(userAddr)))
		freqBkName, _ := getBucketNameByFreq(freq)
		freqBk, _ := b.CreateBucketIfNotExists([]byte(freqBkName))

		min := uint64ToBytes(fromTime)
		max := uint64ToBytes(toTime)
		c := freqBk.Cursor()
		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			value := common.VolumeStats{}
			json.Unmarshal(v, &value)
			key := bytesToUint64(k) / 1000000
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
			b, _ := tx.CreateBucketIfNotExists([]byte(key))
			// update to timezone buckets
			for i := START_TIMEZONE; i <= END_TIMEZONE; i++ {
				stats := stats[i]
				freq := fmt.Sprintf("%s%d", TIMEZONE_BUCKET_PREFIX, i)
				tzBucket, err := b.CreateBucketIfNotExists([]byte(freq))
				if err != nil {
					return err
				}
				for timepoint, stat := range stats {
					timestamp := uint64ToBytes(timepoint)
					// try get data from this timestamp, if exist then add more data
					currentData := common.MetricStats{}
					v := tzBucket.Get(timestamp)
					if v != nil {
						json.Unmarshal(v, &currentData)
					}
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
					dataJSON, err := json.Marshal(currentData)
					if err != nil {
						return err
					}
					tzBucket.Put(timestamp, dataJSON)
				}
			}
		}
		lastProcessBk := tx.Bucket([]byte(TRADELOG_PROCESSOR_STATE))
		if lastProcessBk != nil {
			dataJSON := uint64ToBytes(lastProcessTimePoint)
			lastProcessBk.Put([]byte(TRADE_SUMMARY_AGGREGATION), dataJSON)
		}
		return nil
	})
	return err
}

func (self *BoltStatStorage) GetTradeSummary(fromTime uint64, toTime uint64, timezone int64) (common.StatTicks, error) {
	result := common.StatTicks{}
	tzstring := fmt.Sprintf("%s%d", TIMEZONE_BUCKET_PREFIX, timezone)
	err := self.db.Update(func(tx *bolt.Tx) error {
		b, _ := tx.CreateBucketIfNotExists([]byte("trade_summary"))
		timezoneBk, _ := b.CreateBucketIfNotExists([]byte(tzstring))
		min := uint64ToBytes(fromTime)
		max := uint64ToBytes(toTime)

		c := timezoneBk.Cursor()
		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			summary := common.MetricStats{}
			json.Unmarshal(v, &summary)
			key := bytesToUint64(k) / 1000000
			result[key] = summary
		}
		return nil
	})

	return result, err
}
