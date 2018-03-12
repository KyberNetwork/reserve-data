package storage

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/boltdb/bolt"
)

const (
	KYC_CATEGORY       string = "0x0000000000000000000000000000000000000000000000000000000000000004"
	MAX_NUMBER_VERSION int    = 1000

	LOG_BUCKET         string = "logs"
	TRADE_STATS_BUCKET string = "trade_stats"
	MINUTE_BUCKET      string = "minute"
	HOUR_BUCKET        string = "hour"
	DAY_BUCKET         string = "day"

	ADDRESS_BUCKET       string = "address"
	DAILY_ADDRESS_BUCKET string = "daily_address"
	DAILY_USER_BUCKET    string = "daily_user"

	ADDRESS_CATEGORY  string = "address_category"
	ADDRESS_ID        string = "address_id"
	ID_ADDRESSES      string = "id_addresses"
	PENDING_ADDRESSES string = "pending_addresses"
	RESERVE_RATES     string = "reserve_rates"
)

type BoltStorage struct {
	mu    sync.RWMutex
	db    *bolt.DB
	block uint64
	index uint
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
		tx.CreateBucket([]byte(LOG_BUCKET))
		tx.CreateBucket([]byte(ADDRESS_ID))
		tx.CreateBucket([]byte(ID_ADDRESSES))
		tx.CreateBucket([]byte(ADDRESS_CATEGORY))
		tx.CreateBucket([]byte(TRADE_STATS_BUCKET))
		tx.CreateBucket([]byte(PENDING_ADDRESSES))
		tx.CreateBucket([]byte(RESERVE_RATES))
		tx.CreateBucket([]byte(ADDRESS_BUCKET))
		tx.CreateBucket([]byte(DAILY_ADDRESS_BUCKET))
		tx.CreateBucket([]byte(DAILY_USER_BUCKET))

		tradeStatsBk := tx.Bucket([]byte(TRADE_STATS_BUCKET))
		frequencies := []string{MINUTE_BUCKET, HOUR_BUCKET, DAY_BUCKET}

		for _, freq := range frequencies {
			tradeStatsBk.CreateBucket([]byte(freq))
		}

		return nil
	})
	storage := &BoltStorage{sync.RWMutex{}, db, 0, 0}
	storage.db.View(func(tx *bolt.Tx) error {
		block, index, err := storage.LoadLastLogIndex(tx)
		if err == nil {
			storage.block = block
			storage.index = index
		}
		return err
	})
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

// PruneOutdatedData Remove first version out of database
func (self *BoltStorage) PruneOutdatedData(tx *bolt.Tx, bucket string) error {
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

func (self *BoltStorage) UpdateLogBlock(block, timepoint uint64) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.block = block
	return nil
}

func (self *BoltStorage) LastBlock() (uint64, error) {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return self.block, nil
}

func (self *BoltStorage) LoadLastLogIndex(tx *bolt.Tx) (uint64, uint, error) {
	b := tx.Bucket([]byte(LOG_BUCKET))
	c := b.Cursor()
	k, v := c.Last()
	if k != nil {
		record := common.TradeLog{}
		json.Unmarshal(v, &record)
		return record.BlockNumber, record.TransactionIndex, nil
	} else {
		return 0, 0, errors.New("Database is empty")
	}
}

func (self *BoltStorage) GetTradeLogs(fromTime uint64, toTime uint64) ([]common.TradeLog, error) {
	result := []common.TradeLog{}
	var err error
	self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(LOG_BUCKET))
		c := b.Cursor()
		min := uint64ToBytes(fromTime * 1000000)
		max := uint64ToBytes(toTime * 1000000)
		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			record := common.TradeLog{}
			err = json.Unmarshal(v, &record)
			if err != nil {
				return err
			}
			result = append([]common.TradeLog{record}, result...)
		}
		return nil
	})
	return result, err
}

func (self *BoltStorage) StoreTradeLog(stat common.TradeLog, timepoint uint64) error {
	var err error
	self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(LOG_BUCKET))
		var dataJson []byte
		block, txindex, berr := self.LoadLastLogIndex(tx)
		if berr == nil && (block > stat.BlockNumber || (block == stat.BlockNumber && txindex >= stat.TransactionIndex)) {
			err = errors.New(
				fmt.Sprintf("Duplicated log (new block number %s is smaller or equal to latest block number %s)", block, stat.BlockNumber))
			return err
		}
		dataJson, err = json.Marshal(stat)
		if err != nil {
			return err
		}
		log.Printf("Storing log: %d", stat.Timestamp)
		idByte := uint64ToBytes(stat.Timestamp)
		err = b.Put(idByte, dataJson)
		return err
	})
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
	}
	return
}

func getTimestampByFreq(t uint64, freq string) (result []byte) {
	switch freq {
	case "m", "M":
		result = uint64ToBytes(t / uint64(time.Minute) * uint64(time.Minute))
	case "h", "H":
		result = uint64ToBytes(t / uint64(time.Hour) * uint64(time.Hour))
	case "d", "D":
		result = uint64ToBytes(t / uint64(time.Hour*24) * uint64(time.Hour*24))
	}
	return
}

func (self *BoltStorage) SetTradeStats(freq string, t uint64, tradeStats common.TradeStats) (err error) {
	self.db.Update(func(tx *bolt.Tx) error {
		tradeStatsBk := tx.Bucket([]byte(TRADE_STATS_BUCKET))

		freqBkName, err := getBucketNameByFreq(freq)
		if err != nil {
			return err
		}
		freqBk := tradeStatsBk.Bucket([]byte(freqBkName))

		timestamp := getTimestampByFreq(t, freq)

		rawStats := freqBk.Get(timestamp)
		var stats common.TradeStats
		if rawStats != nil {
			json.Unmarshal(rawStats, &stats)
		} else {
			stats = common.TradeStats{}
		}

		for key, value := range tradeStats {
			sum, ok := stats[key]
			if ok {
				stats[key] = sum + value
			} else {
				stats[key] = value
			}
		}

		dataJSON, err := json.Marshal(stats)
		if err != nil {
			return err
		}

		if err := freqBk.Put(timestamp, dataJSON); err != nil {
			return err
		}

		return err
	})
	return
}

func (self *BoltStorage) getTradeStats(fromTime, toTime uint64, freq string) (map[uint64]common.TradeStats, error) {
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

func (self *BoltStorage) SaveUserAddress(timestamp uint64, addr string) (common.TradeStats, error) {
	stats := common.TradeStats{}
	var err error
	var kycEd bool

	self.db.Update(func(tx *bolt.Tx) error {
		// CHECK IF USER HAD TRADED EVER BEFORE
		addrBk := tx.Bucket([]byte(ADDRESS_BUCKET))
		v := addrBk.Get([]byte(addr))
		if v == nil {
			stats["first_trade_ever"] = 1
			stats["first_trade_in_day"] = 1

			if err := addrBk.Put([]byte(addr), []byte("1")); err != nil {
				return err
			}
		} else {
			// IF USER HAD TRADED, CHECK IF USER HAD TRADED IN DAY BEFORE
			dailyAddrBk := tx.Bucket([]byte(DAILY_ADDRESS_BUCKET))

			timestamp := getTimestampByFreq(timestamp, "D")
			raw := dailyAddrBk.Get(timestamp)

			var addrTradedInDay map[string]bool
			if raw != nil {
				json.Unmarshal(raw, &addrTradedInDay)
			} else {
				addrTradedInDay = map[string]bool{}
			}

			if _, traded := addrTradedInDay[addr]; !traded {
				stats["first_trade_in_day"] = 1

				addrTradedInDay[addr] = true
				dataJSON, err := json.Marshal(addrTradedInDay)
				if err != nil {
					return err
				}

				if err := dailyAddrBk.Put(timestamp, dataJSON); err != nil {
					return err
				}
			}
		}

		// CHECK IF KYCED USER HAD TRADED IN DAY BEFORE
		// MEANS: first trade from addr in day, user is kyced and also first trade from user in day
		if stats["first_trade_in_day"] == 1 {
			user, err := self.GetUserOfAddress(addr)
			if err != nil {
				return err
			}
			kycEd = (user != addr && user != "")

			if kycEd {
				dailyUserBk := tx.Bucket([]byte(DAILY_USER_BUCKET))

				timestamp := getTimestampByFreq(timestamp, "D")
				raw := dailyUserBk.Get(timestamp)

				var userTradedInDay map[string]bool
				if raw != nil {
					json.Unmarshal(raw, &userTradedInDay)
				} else {
					userTradedInDay = map[string]bool{}
				}

				if _, traded := userTradedInDay[user]; !traded {
					stats["kyced_in_day"] = 1

					userTradedInDay[user] = true
					dataJSON, err := json.Marshal(userTradedInDay)
					if err != nil {
						return err
					}

					if err := dailyUserBk.Put(timestamp, dataJSON); err != nil {
						return err
					}
				}
			}
		}

		return nil
	})

	return stats, err
}

func (self *BoltStorage) GetAssetVolume(fromTime uint64, toTime uint64, freq string, asset string) (common.StatTicks, error) {
	result := common.StatTicks{}

	stats, err := self.getTradeStats(fromTime, toTime, freq)
	eth := common.SupportedTokens["ETH"]

	for timestamp, stat := range stats {
		if strings.ToLower(eth.Address) == asset {
			result[timestamp] = map[string]float64{
				"volume":     stat[strings.Join([]string{"assets_volume", asset}, "_")],
				"usd_amount": stat[strings.Join([]string{"assets_usd_amount", asset}, "_")],
			}
		} else {
			result[timestamp] = map[string]float64{
				"volume":     stat[strings.Join([]string{"assets_volume", asset}, "_")],
				"eth_amount": stat[strings.Join([]string{"assets_eth_amount", asset}, "_")],
				"usd_amount": stat[strings.Join([]string{"assets_usd_amount", asset}, "_")],
			}
		}
	}

	return result, err
}

func (self *BoltStorage) GetBurnFee(fromTime uint64, toTime uint64, freq string, reserveAddr string) (common.StatTicks, error) {
	result := common.StatTicks{}

	stats, err := self.getTradeStats(fromTime, toTime, freq)
	for timestamp, stat := range stats {
		result[timestamp] = stat[strings.Join([]string{"burn_fee", reserveAddr}, "_")]
	}

	return result, err
}

func (self *BoltStorage) GetWalletFee(fromTime uint64, toTime uint64, freq string, reserveAddr string, walletAddr string) (common.StatTicks, error) {
	result := common.StatTicks{}

	stats, err := self.getTradeStats(fromTime, toTime, freq)
	for timestamp, stat := range stats {
		result[timestamp] = stat[strings.Join([]string{"wallet_fee", reserveAddr, walletAddr}, "_")]
	}

	return result, err
}

func (self *BoltStorage) GetUserVolume(fromTime uint64, toTime uint64, freq string, userAddr string) (common.StatTicks, error) {
	result := common.StatTicks{}

	stats, err := self.getTradeStats(fromTime, toTime, freq)
	for timestamp, stat := range stats {
		result[timestamp] = stat[strings.Join([]string{"user_volume", userAddr}, "_")]
	}
	return result, err
}

func (self *BoltStorage) GetTradeSummary(fromTime uint64, toTime uint64) (common.StatTicks, error) {
	result := common.StatTicks{}

	stats, err := self.getTradeStats(fromTime, toTime, "D")
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

func (self *BoltStorage) GetAddressesOfUser(user string) ([]string, error) {
	var err error
	result := []string{}
	self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ID_ADDRESSES))
		userBucket := b.Bucket([]byte(user))
		if userBucket != nil {
			userBucket.ForEach(func(k, v []byte) error {
				result = append(result, string(k))
				return nil
			})
		}
		return nil
	})
	return result, err
}

func (self *BoltStorage) GetUserOfAddress(addr string) (string, error) {
	addr = strings.ToLower(addr)
	var err error
	var result string
	self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ADDRESS_ID))
		id := b.Get([]byte(addr))
		if id == nil {
			log.Println("Get user = nil for user address %s", addr)
		}
		result = string(id)
		return nil
	})
	return result, err
}

func (self *BoltStorage) GetCategory(addr string) (string, error) {
	addr = strings.ToLower(addr)
	var err error
	var result string
	self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ADDRESS_CATEGORY))
		cat := b.Get([]byte(addr))
		result = string(cat)
		return nil
	})
	return result, err
}

func (self *BoltStorage) GetPendingAddresses() ([]string, error) {
	var err error
	result := []string{}
	self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(PENDING_ADDRESSES))
		b.ForEach(func(k, v []byte) error {
			result = append(result, string(k))
			return nil
		})
		return nil
	})
	return result, err
}

func (self *BoltStorage) UpdateUserAddresses(user string, addrs []string) error {
	user = strings.ToLower(user)
	addresses := []string{}
	for _, addr := range addrs {
		addresses = append(addresses, strings.ToLower(addr))
	}
	var err error
	self.db.Update(func(tx *bolt.Tx) error {
		for _, address := range addresses {
			// get temp user identity
			b := tx.Bucket([]byte(ADDRESS_ID))
			oldID := b.Get([]byte(address))
			// remove the addresses bucket assocciated to this temp user
			b = tx.Bucket([]byte(ID_ADDRESSES))
			b.DeleteBucket(oldID)
			// update user to each address => user
			b = tx.Bucket([]byte(ADDRESS_ID))
			b.Put([]byte(address), []byte(user))
		}
		// remove old addresses from pending bucket
		pendingBk := tx.Bucket([]byte(PENDING_ADDRESSES))
		oldAddrs, err := self.GetAddressesOfUser(user)
		if err != nil {
			return err
		}
		for _, oldAddr := range oldAddrs {
			pendingBk.Delete([]byte(oldAddr))
		}
		// update addresses bucket for real user
		// add new addresses to pending bucket
		b := tx.Bucket([]byte(ID_ADDRESSES))
		b, err = b.CreateBucketIfNotExists([]byte(user))
		if err != nil {
			return err
		}
		catBk := tx.Bucket([]byte(ADDRESS_CATEGORY))
		for _, address := range addresses {
			b.Put([]byte(address), []byte{1})
			cat := catBk.Get([]byte(address))
			log.Printf("category of %s: %s", address, cat)
			if string(cat) != KYC_CATEGORY {
				pendingBk.Put([]byte(address), []byte{1})
			}
		}
		return nil
	})
	return err
}

func (self *BoltStorage) StoreCatLog(l common.SetCatLog) error {
	var err error
	self.db.Update(func(tx *bolt.Tx) error {
		// map address to category
		b := tx.Bucket([]byte(ADDRESS_CATEGORY))
		addrBytes := []byte(strings.ToLower(l.Address.Hex()))
		b.Put(addrBytes, []byte(strings.ToLower(l.Category)))
		// get the user of it
		b = tx.Bucket([]byte(ADDRESS_ID))
		user := b.Get(addrBytes)
		if len(user) == 0 {
			// if the user doesn't exist, we set the user to its address
			user = addrBytes
		}
		// add address to its user addresses
		b = tx.Bucket([]byte(ID_ADDRESSES))
		b, err = b.CreateBucketIfNotExists(user)
		if err != nil {
			return err
		}
		b.Put(addrBytes, []byte{1})
		// add user to map
		b = tx.Bucket([]byte(ADDRESS_ID))
		b.Put(addrBytes, user)
		// remove address from pending list
		b = tx.Bucket([]byte(PENDING_ADDRESSES))
		b.Delete(addrBytes)
		return nil
	})
	return err
}

func (self *BoltStorage) StoreReserveRates(reserveAddr string, reserveRates common.ReserveRates, timepoint uint64) error {
	var err error
	self.db.Update(func(tx *bolt.Tx) error {
		b, _ := tx.CreateBucketIfNotExists([]byte(reserveAddr))
		c := b.Cursor()
		var prevDataJSON common.ReserveRates
		_, prevData := c.Last()
		json.Unmarshal(prevData, &prevDataJSON)
		if prevDataJSON.BlockNumber < reserveRates.BlockNumber {
			idByte := uint64ToBytes(timepoint)
			dataJson, err := json.Marshal(reserveRates)
			if err != nil {
				return err
			}
			err = b.Put(idByte, dataJson)
			if err != nil {
				return err
			}
			log.Printf("Save rates to db %s successfully", reserveAddr)
		}
		return nil
	})
	return err
}

func (self *BoltStorage) GetReserveRates(fromTime, toTime uint64, reserveAddr string) ([]common.ReserveRates, error) {
	var err error
	var result []common.ReserveRates
	self.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(reserveAddr))
		if err != nil {
			log.Println("Cannot get bucket: ", err.Error())
			return err
		}
		c := b.Cursor()
		min := uint64ToBytes(fromTime)
		max := uint64ToBytes(toTime)
		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			rates := common.ReserveRates{}
			err := json.Unmarshal(v, &rates)
			if err != nil {
				return err
			}
			result = append(result, rates)
		}
		return err
	})
	return result, err
}
