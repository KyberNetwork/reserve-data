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
)

const (
	MAX_NUMBER_VERSION int = 1000

	TRADELOG_PROCESSOR_STATE string = "tradelog_processor_state"

	TRADE_STATS_BUCKET string = "trade_stats"

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
	COUNTRY_BUCKET              string = "country_bucket"
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
	db.Update(func(tx *bolt.Tx) error {
		tx.CreateBucketIfNotExists([]byte(TRADE_STATS_BUCKET))
		tx.CreateBucketIfNotExists([]byte(TRADELOG_PROCESSOR_STATE))
		tx.CreateBucketIfNotExists([]byte(WALLET_ADDRESS_BUCKET))
		tx.CreateBucketIfNotExists([]byte(COUNTRY_BUCKET))
		//create timezone buckets
		tradeStatsBk := tx.Bucket([]byte(TRADE_STATS_BUCKET))
		frequencies := []string{MINUTE_BUCKET, HOUR_BUCKET, DAY_BUCKET}

		for _, freq := range frequencies {
			tradeStatsBk.CreateBucketIfNotExists([]byte(freq))
		}
		for i := START_TIMEZONE; i <= END_TIMEZONE; i++ {
			tzstring := fmt.Sprintf("%s%d", TIMEZONE_BUCKET_PREFIX, i)
			tx.CreateBucketIfNotExists([]byte(tzstring))
			tradeStatsBk.CreateBucketIfNotExists([]byte(tzstring))
			dailyAddrBkname := fmt.Sprintf("%s%d", DAILY_ADDRESS_BUCKET_PREFIX, i)
			tx.CreateBucketIfNotExists([]byte(dailyAddrBkname))
			dailyUserBkname := fmt.Sprintf("%s%d", DAILY_USER_BUCKET_PREFIX, i)
			tx.CreateBucketIfNotExists([]byte(dailyUserBkname))
			addrBucketName := fmt.Sprintf("%s%d", ADDRESS_BUCKET_PREFIX, i)
			tx.CreateBucketIfNotExists([]byte(addrBucketName))
		}

		return nil
	})
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

func (self *BoltStatStorage) SetLastProcessedTradeLogTimepoint(timepoint uint64) error {
	var err error
	self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TRADELOG_PROCESSOR_STATE))
		err = b.Put([]byte("last_timepoint"), uint64ToBytes(timepoint))
		return err
	})
	return err
}

func (self *BoltStatStorage) GetLastProcessedTradeLogTimepoint() (uint64, error) {
	var result uint64
	self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TRADELOG_PROCESSOR_STATE))
		result = bytesToUint64(b.Get([]byte("last_timepoint")))
		return nil
	})
	return result, nil
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

func (self *BoltStatStorage) SetTradeStats(freq string, t uint64, tradeStats common.TradeStats) (err error) {
	self.db.Update(func(tx *bolt.Tx) error {
		tradeStatsBk := tx.Bucket([]byte(TRADE_STATS_BUCKET))

		freqBkName, err := getBucketNameByFreq(freq)
		if err != nil {
			return err
		}
		freqBk := tradeStatsBk.Bucket([]byte(freqBkName))
		timestamp := getTimestampByFreq(t, freq)
		log.Printf("AGGREGATE SetTradeStats, getting raw stat")
		rawStats := freqBk.Get(timestamp)
		var stats common.TradeStats

		if rawStats != nil {
			json.Unmarshal(rawStats, &stats)
		} else {
			stats = common.TradeStats{}
		}
		log.Printf("AGGREGATE SetTradeStats, unmarshaled stat, len(raw)=%d", len(rawStats))
		for key, value := range tradeStats {
			sum, ok := stats[key]
			if ok {
				stats[key] = sum + value
			} else {
				stats[key] = value
			}
		}
		dataJSON, err := json.Marshal(stats)
		log.Printf("AGGREGATE SetTradeStats, marshal updated stat, len(raw)=%d", len(dataJSON))
		if err != nil {
			return err
		}

		if err := freqBk.Put(timestamp, dataJSON); err != nil {
			return err
		}
		log.Printf("AGGREGATE SetTradeStats, finish")

		return err
	})
	log.Printf("AGGREGATE SetTradeStats, updated the db")
	return
}

func (self *BoltStatStorage) getTradeStats(fromTime, toTime uint64, freq string) (map[uint64]common.TradeStats, error) {
	result := map[uint64]common.TradeStats{}
	var err error
	self.db.View(func(tx *bolt.Tx) error {
		// Get trade stats bucket
		tradeStatsBk := tx.Bucket([]byte(TRADE_STATS_BUCKET))

		var freqBkName string
		freqBkName, err = getBucketNameByFreq(freq)
		if err != nil {
			return err
		}
		freqBk := tradeStatsBk.Bucket([]byte(freqBkName))
		c := freqBk.Cursor()
		// min := getTimestampByFreq(fromTime, freq)
		// max := getTimestampByFreq(toTime, freq)
		// log.Printf("from %d to %d", min, max)

		min := uint64ToBytes(fromTime)
		max := uint64ToBytes(toTime)

		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			stats := common.TradeStats{}
			err = json.Unmarshal(v, &stats)
			// log.Printf("%v", stats)
			if err != nil {
				return err
			}

			timestamp := bytesToUint64(k) / 1000000 // to milis
			result[timestamp] = stats
		}
		return nil
	})
	return result, err
}

func getUserStatsBucket(tx *bolt.Tx,
	stats common.TradeStats, timestamp uint64, timezone int64,
	firstTradeBucket, firstTradeEverBucket, kycInDay string,
	dailyAddrKey, addrKey, dailyUserKey string, kycEd bool) {
	dailyAddrBkname := fmt.Sprintf("%s%d", DAILY_ADDRESS_BUCKET_PREFIX, timezone)
	dailyAddrBk := tx.Bucket([]byte(dailyAddrBkname))
	if v := dailyAddrBk.Get([]byte(dailyAddrKey)); v == nil {
		stats[firstTradeBucket] = 1 // FIRST TRADE IN DAY
		addrBucketName := fmt.Sprintf("%s%d", ADDRESS_BUCKET_PREFIX, timezone)
		addrBk := tx.Bucket([]byte(addrBucketName))
		if v := addrBk.Get([]byte(addrKey)); v == nil {
			stats[firstTradeEverBucket] = 1 // FIRST TRADE EVER
		}

		if kycEd {
			dailyUserBkname := fmt.Sprintf("%s%d", DAILY_USER_BUCKET_PREFIX, timezone)
			dailyUserBk := tx.Bucket([]byte(dailyUserBkname))
			if v := dailyUserBk.Get([]byte(dailyUserKey)); v == nil {
				stats[kycInDay] = 1
			}
		}
	}
}

func (self *BoltStatStorage) GetUserStats(timestamp uint64,
	addr, email, wallet, country string, kycEd bool, timezone int64) (common.TradeStats, error) {
	stats := common.TradeStats{}
	var err error

	self.db.View(func(tx *bolt.Tx) error {
		freq := fmt.Sprintf("%s%d", TIMEZONE_BUCKET_PREFIX, timezone)
		dailyTimestamp := string(getTimestampByFreq(timestamp, freq))
		dailyAddrKey := fmt.Sprintf("%s_%s", dailyTimestamp, addr)
		dailyUserKey := fmt.Sprintf("%s_%s", dailyTimestamp, email)
		getUserStatsBucket(tx, stats, timestamp, timezone,
			"first_trade_in_day", "first_trade_ever", "kyced_in_day", dailyAddrKey, addr, dailyUserKey, kycEd)

		//Get user stat for the current Wallet address
		dailyAddrWalletKey := fmt.Sprintf("%s_%s_%s", dailyTimestamp, addr, wallet)
		walletFirstTrade := fmt.Sprintf("wallet_first_trade_in_day_%s", wallet)
		walletFirstTradeEver := fmt.Sprintf("wallet_first_trade_ever_%s", wallet)
		walletKycedInday := fmt.Sprintf("wallet_kyced_in_day_%s", wallet)
		addrWalletKey := fmt.Sprintf("%s_%s", addr, wallet)
		dailyUserWalletKey := fmt.Sprintf("%s_%s_%s", dailyTimestamp, email, wallet)
		getUserStatsBucket(tx, stats, timestamp, timezone,
			walletFirstTrade, walletFirstTradeEver, walletKycedInday, dailyAddrWalletKey, addrWalletKey, dailyUserWalletKey, kycEd)

		//Get user stat for the current Wallet address
		geoFirstTradeInDay := fmt.Sprintf("geo_first_trade_in_day_%s", country)
		geoFirstTradeEver := fmt.Sprintf("geo_first_trade_ever_%s", country)
		geoKycedInDay := fmt.Sprintf("geo_kyced_in_day_%s", country)
		addrCountryKey := fmt.Sprintf("%s_%s", addr, country)
		dailyAddrCountryKey := fmt.Sprintf("%s_%s_%s", dailyTimestamp, addr, country)
		dailyUserCountryKey := fmt.Sprintf("%s_%s_%s", dailyTimestamp, email, wallet)
		getUserStatsBucket(tx, stats, timestamp, timezone,
			geoFirstTradeInDay, geoFirstTradeEver, geoKycedInDay, dailyAddrCountryKey, addrCountryKey, dailyUserCountryKey, kycEd)

		return nil
	})

	return stats, err
}

func isEalier(k, timestamp []byte) bool {
	temps := strings.Split(string(k), "_")
	return (bytes.Compare([]byte(temps[0]), timestamp) <= 0)
}

func (self *BoltStatStorage) PruneDailyBucket(timepoint uint64, timezone int64) (err error) {
	freq := fmt.Sprintf("%s%d", TIMEZONE_BUCKET_PREFIX, timezone)
	currentUintTimeStamp := bytesToUint64(getTimestampByFreq(timepoint, freq))
	expiredByteTimeStamp := uint64ToBytes(currentUintTimeStamp - EXPIRED)

	//update daily Address bucket
	dailyAddrBkname := fmt.Sprintf("%s%d", DAILY_ADDRESS_BUCKET_PREFIX, timezone)
	self.db.Update(func(tx *bolt.Tx) error {
		dailyAddrBk := tx.Bucket([]byte(dailyAddrBkname))
		c := dailyAddrBk.Cursor()
		for k, _ := c.First(); k != nil && isEalier(k, expiredByteTimeStamp); k, _ = c.Next() {
			err = dailyAddrBk.Delete([]byte(k))
		}
		return err
	})
	//update daily User bucket
	dailyUserBkname := fmt.Sprintf("%s%d", DAILY_USER_BUCKET_PREFIX, timezone)
	self.db.Update(func(tx *bolt.Tx) error {
		dailyAddrBk := tx.Bucket([]byte(dailyUserBkname))
		c := dailyAddrBk.Cursor()
		for k, _ := c.First(); k != nil && isEalier(k, expiredByteTimeStamp); k, _ = c.Next() {
			err = dailyAddrBk.Delete([]byte(k))
		}
		return err
	})
	return
}

func setUserBucket(tx *bolt.Tx,
	stats common.TradeStats, firstTradeBucket, firstTradeEverBucket, addr, dailyTimestamp, email string,
	dailyAddrBkname, dailyAddrKey, dailyUserBkname, dailyUserKey, addrBucketKey string,
	timezone int64, kycEd bool) error {
	if _, traded := stats[firstTradeBucket]; traded {
		dailyAddrBk := tx.Bucket([]byte(dailyAddrBkname))
		if err := dailyAddrBk.Put([]byte(dailyAddrKey), []byte("1")); err != nil {
			return err
		}
		if _, traded := stats[firstTradeEverBucket]; traded {
			addrBucketName := fmt.Sprintf("%s%d", ADDRESS_BUCKET_PREFIX, timezone)
			addrBk := tx.Bucket([]byte(addrBucketName))
			if err := addrBk.Put([]byte(addrBucketKey), []byte("1")); err != nil {
				return err
			}
		}

		if kycEd {
			dailyUserBk := tx.Bucket([]byte(dailyUserBkname))
			if err := dailyUserBk.Put([]byte(dailyUserKey), []byte("1")); err != nil {
				return err
			}
		}
	}
	return nil
}

func (self *BoltStatStorage) SetUserStats(timestamp uint64,
	addr, email, wallet, country string, kycEd bool, timezone int64, stats common.TradeStats) error {
	var err error
	freq := fmt.Sprintf("%s%d", TIMEZONE_BUCKET_PREFIX, timezone)
	if err = self.SetTradeStats(freq, timestamp, stats); err != nil {
		return err
	}
	self.PruneDailyBucket(timestamp, timezone)

	self.db.Update(func(tx *bolt.Tx) error {

		dailyTimestamp := string(getTimestampByFreq(timestamp, freq))
		dailyAddrBkname := fmt.Sprintf("%s%d", DAILY_ADDRESS_BUCKET_PREFIX, timezone)
		dailyAddrKey := fmt.Sprintf("%s_%s", dailyTimestamp, addr)
		dailyUserBkname := fmt.Sprintf("%s%d", DAILY_USER_BUCKET_PREFIX, timezone)
		dailyUserKey := fmt.Sprintf("%s_%s", dailyTimestamp, email)
		setUserBucket(tx, stats, "first_trade_in_day", "first_trade_ever", addr, dailyTimestamp, email, dailyAddrBkname, dailyAddrKey, dailyUserBkname, dailyUserKey, addr, timezone, kycEd)

		//Set user stat for the current Wallet address
		walletFirstTradeBucket := fmt.Sprintf("wallet_first_trade_in_day_%s", wallet)
		walletFirstTradeEverBucket := fmt.Sprintf("wallet_first_trade_ever_%s", wallet)
		dailyAddrWalletKey := fmt.Sprintf("%s_%s_%s", dailyTimestamp, addr, wallet)
		dailyUserWalletKey := fmt.Sprintf("%s_%s_%s", dailyTimestamp, email, wallet)
		addrWalletKey := fmt.Sprintf("%s_%s", addr, wallet)
		setUserBucket(tx, stats, walletFirstTradeBucket, walletFirstTradeEverBucket, addr, dailyTimestamp, email,
			dailyAddrBkname, dailyAddrWalletKey, dailyUserBkname, dailyUserWalletKey, addrWalletKey, timezone, kycEd)

		// Set user stat for the current country
		countryFirstTradeBucket := fmt.Sprintf("geo_first_trade_in_day_%s", country)
		countryFirstTradeEverBucket := fmt.Sprintf("geo_first_trade_ever_%s", country)
		dailyAddrCountryKey := fmt.Sprintf("%s_%s_%s", dailyTimestamp, addr, country)
		dailyUserCountryKey := fmt.Sprintf("%s_%s_%s", dailyTimestamp, email, country)
		addrCountryKey := fmt.Sprintf("%s_%s", addr, country)
		setUserBucket(tx, stats, countryFirstTradeBucket, countryFirstTradeEverBucket, addr, dailyTimestamp, email,
			dailyAddrBkname, dailyAddrCountryKey, dailyUserBkname, dailyUserCountryKey, addrCountryKey, timezone, kycEd)
		return nil
	})
	return err
}

func (self *BoltStatStorage) SetWalletAddress(walletAddr string) (err error) {
	self.db.Update(func(tx *bolt.Tx) error {
		walletBucket := tx.Bucket([]byte(WALLET_ADDRESS_BUCKET))

		if err := walletBucket.Put([]byte(walletAddr), []byte("1")); err != nil {
			return err
		}

		return nil
	})
	return
}

func (self *BoltStatStorage) GetWalletAddress() ([]string, error) {
	var result []string
	self.db.View(func(tx *bolt.Tx) error {
		walletBucket := tx.Bucket([]byte(WALLET_ADDRESS_BUCKET))
		c := walletBucket.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			result = append(result, string(k[:]))
		}
		return nil
	})
	return result, nil
}

func (self *BoltStatStorage) GetWalletStats(fromTime uint64, toTime uint64, walletAddr string, timezone int64) (common.StatTicks, error) {

	walletAddr = strings.ToLower(walletAddr)
	log.Printf("fromtime %v toTime %v fre %v", fromTime, toTime, timezone)
	result := common.StatTicks{}
	tzstring := fmt.Sprintf("%s%d", TIMEZONE_BUCKET_PREFIX, timezone)
	stats, err := self.getTradeStats(fromTime, toTime, tzstring)
	if err != nil {
		return result, err
	}

	for timestamp, stat := range stats {
		trade_countstr, found := stat[fmt.Sprintf("wallet_trade_count_%s", walletAddr)]
		if !found {
			continue
		}
		tradeCount := float64(trade_countstr)
		var avgEth, avgUsd float64
		if tradeCount > 0 {
			avgEth = float64(stat[fmt.Sprintf("wallet_eth_volume_%s", walletAddr)]) / tradeCount
			avgUsd = float64(stat[fmt.Sprintf("wallet_usd_volume_%s", walletAddr)]) / tradeCount
		}
		result[timestamp] = map[string]float64{
			"total_eth_volume":     stat[fmt.Sprintf("wallet_eth_volume_%s", walletAddr)],
			"total_usd_amount":     stat[fmt.Sprintf("wallet_usd_volume_%s", walletAddr)],
			"total_burn_fee":       stat[fmt.Sprintf("wallet_burn_fee_%s", walletAddr)],
			"unique_addresses":     stat[fmt.Sprintf("wallet_first_trade_in_day_%s", walletAddr)],
			"new_unique_addresses": stat[fmt.Sprintf("wallet_first_trade_ever_%s", walletAddr)],
			"kyced_addresses":      stat[fmt.Sprintf("wallet_kyced_in_day_%s", walletAddr)],
			"total_trade":          tradeCount,
			"eth_per_trade":        avgEth,
			"usd_per_trade":        avgUsd,
		}
	}
	return result, nil
}

func (self *BoltStatStorage) SetCountry(country string) error {
	var err error
	self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(COUNTRY_BUCKET))
		err = b.Put([]byte(country), []byte("1"))
		return err
	})
	return err
}

func (self *BoltStatStorage) GetCountries() ([]string, error) {
	countries := []string{}
	var err error
	self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(COUNTRY_BUCKET))
		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			countries = append(countries, string(k))
		}
		return err
	})
	return countries, err
}

func (self *BoltStatStorage) GetCountryStats(fromTime, toTime uint64, country string, timezone int64) (common.StatTicks, error) {
	result := common.StatTicks{}
	tzstring := fmt.Sprintf("%s%d", TIMEZONE_BUCKET_PREFIX, timezone)
	stats, err := self.getTradeStats(fromTime, toTime, tzstring)
	if err != nil {
		return result, err
	}

	for timestamp, stat := range stats {
		trade_countstr, found := stat[fmt.Sprintf("geo_trade_count_%s", country)]
		if !found {
			continue
		}
		tradeCount := float64(trade_countstr)
		var avgEth, avgUsd float64
		if tradeCount > 0 {
			avgEth = float64(stat[fmt.Sprintf("geo_eth_volume_%s", country)]) / tradeCount
			avgUsd = float64(stat[fmt.Sprintf("geo_usd_volume_%s", country)]) / tradeCount
		}
		result[timestamp] = map[string]float64{
			"total_eth_volume":     stat[fmt.Sprintf("geo_eth_volume_%s", country)],
			"total_usd_amount":     stat[fmt.Sprintf("geo_usd_volume_%s", country)],
			"total_burn_fee":       stat[fmt.Sprintf("geo_burn_fee_%s", country)],
			"unique_addresses":     stat[fmt.Sprintf("geo_first_trade_in_day_%s", country)],
			"new_unique_addresses": stat[fmt.Sprintf("geo_first_trade_ever_%s", country)],
			"kyced_addresses":      stat[fmt.Sprintf("geo_kyced_in_day_%s", country)],
			"total_trade":          tradeCount,
			"eth_per_trade":        avgEth,
			"usd_per_trade":        avgUsd,
		}
	}
	return result, nil

}

func (self *BoltStatStorage) GetAssetVolume(fromTime uint64, toTime uint64, freq string, asset string) (common.StatTicks, error) {
	result := common.StatTicks{}

	stats, err := self.getTradeStats(fromTime, toTime, freq)
	eth := common.SupportedTokens["ETH"]

	for timestamp, stat := range stats {
		if strings.ToLower(eth.Address) == asset {
			result[timestamp] = map[string]float64{
				"volume":     stat[fmt.Sprintf("assets_volume_%s", asset)],
				"usd_amount": stat[fmt.Sprintf("assets_usd_amount_%s", asset)],
			}
		} else {
			result[timestamp] = map[string]float64{
				"volume":     stat[fmt.Sprintf("assets_volume_%s", asset)],
				"eth_amount": stat[fmt.Sprintf("assets_eth_amount_%s", asset)],
				"usd_amount": stat[fmt.Sprintf("assets_usd_amount_%s", asset)],
			}
		}
	}

	return result, err
}

func (self *BoltStatStorage) GetBurnFee(fromTime uint64, toTime uint64, freq string, reserveAddr string) (common.StatTicks, error) {
	result := common.StatTicks{}

	stats, err := self.getTradeStats(fromTime, toTime, freq)
	for timestamp, stat := range stats {
		result[timestamp] = stat[fmt.Sprintf("burn_fee_%s", reserveAddr)]
	}

	return result, err
}

func (self *BoltStatStorage) GetWalletFee(fromTime uint64, toTime uint64, freq string, reserveAddr string, walletAddr string) (common.StatTicks, error) {
	result := common.StatTicks{}

	stats, err := self.getTradeStats(fromTime, toTime, freq)
	for timestamp, stat := range stats {
		result[timestamp] = stat[fmt.Sprintf("wallet_fee_%s_%s", reserveAddr, walletAddr)]
	}

	return result, err
}

func (self *BoltStatStorage) GetUserVolume(fromTime uint64, toTime uint64, freq string, userAddr string) (common.StatTicks, error) {
	result := common.StatTicks{}

	stats, err := self.getTradeStats(fromTime, toTime, freq)
	for timestamp, stat := range stats {
		result[timestamp] = stat[fmt.Sprintf("user_volume_%s", userAddr)]
	}
	return result, err
}

func (self *BoltStatStorage) GetTradeSummary(fromTime uint64, toTime uint64, timezone int64) (common.StatTicks, error) {
	log.Printf("fromtime %v toTime %v fre %v", fromTime, toTime, timezone)
	result := common.StatTicks{}
	tzstring := fmt.Sprintf("%s%d", TIMEZONE_BUCKET_PREFIX, timezone)
	stats, err := self.getTradeStats(fromTime, toTime, tzstring)
	if err != nil {
		return result, err
	}

	for timestamp, stat := range stats {
		tradeCount := float64(stat["trade_count"])
		var avgEth, avgUsd float64
		if tradeCount > 0 {
			avgEth = float64(stat["eth_volume"]) / tradeCount
			avgUsd = float64(stat["usd_volume"]) / tradeCount
		}

		result[timestamp] = map[string]float64{
			"total_eth_volume":     stat["eth_volume"],
			"total_usd_amount":     stat["usd_volume"],
			"total_burn_fee":       stat["burn_fee"],
			"unique_addresses":     stat["first_trade_in_day"],
			"new_unique_addresses": stat["first_trade_ever"],
			"kyced_addresses":      stat["kyced_in_day"],
			"total_trade":          tradeCount,
			"eth_per_trade":        avgEth,
			"usd_per_trade":        avgUsd,
		}
	}

	return result, err
}
