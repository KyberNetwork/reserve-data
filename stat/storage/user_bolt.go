package storage

import (
	"log"
	"strings"

	"github.com/KyberNetwork/reserve-data/boltutil"
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/boltdb/bolt"
	ethereum "github.com/ethereum/go-ethereum/common"
)

const (
	KYC_CATEGORY string = "0x0000000000000000000000000000000000000000000000000000000000000004"

	CATLOG_PROCESSOR_STATE string = "catlog_processor_state"

	ADDRESS_CATEGORY  string = "address_category"
	ADDRESS_ID        string = "address_id"
	ID_ADDRESSES      string = "id_addresses"
	ADDRESS_TIME      string = "address_time"
	PENDING_ADDRESSES string = "pending_addresses"
)

type BoltUserStorage struct {
	db *bolt.DB
}

func NewBoltUserStorage(path string) (*BoltUserStorage, error) {
	var err error
	var db *bolt.DB
	db, err = bolt.Open(path, 0600, nil)
	if db == nil {
		return nil, err
	}

	// init buckets
	err = db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists([]byte(ADDRESS_CATEGORY))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(ADDRESS_ID))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(ID_ADDRESSES))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(ADDRESS_TIME))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(PENDING_ADDRESSES))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(CATLOG_PROCESSOR_STATE))
		if err != nil {
			return err
		}
		return nil
	})
	storage := &BoltUserStorage{db}
	return storage, err
}

func (self *BoltUserStorage) SetLastProcessedCatLogTimepoint(timepoint uint64) error {
	var err error
	err = self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(CATLOG_PROCESSOR_STATE))
		err = b.Put([]byte("last_timepoint"), boltutil.Uint64ToBytes(timepoint))
		return err
	})
	return err
}

func (self *BoltUserStorage) GetLastProcessedCatLogTimepoint() (uint64, error) {
	var result uint64
	var err error
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(CATLOG_PROCESSOR_STATE))
		result = boltutil.BytesToUint64(b.Get([]byte("last_timepoint")))
		return nil
	})
	return result, err
}

func (self *BoltUserStorage) UpdateAddressCategory(address ethereum.Address, cat string) error {
	var err error
	err = self.db.Update(func(tx *bolt.Tx) error {
		// map address to category
		b := tx.Bucket([]byte(ADDRESS_CATEGORY))
		addrBytes := []byte(common.AddrToString(address))
		err = b.Put(addrBytes, []byte(strings.ToLower(cat)))
		if err != nil {
			return err
		}
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
		err = b.Put(addrBytes, []byte{1})
		if err != nil {
			return err
		}
		// add user to map
		b = tx.Bucket([]byte(ADDRESS_ID))
		err = b.Put(addrBytes, user)
		if err != nil {
			return err
		}
		// remove address from pending list
		b = tx.Bucket([]byte(PENDING_ADDRESSES))
		err = b.Delete(addrBytes)
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

func (self *BoltUserStorage) UpdateUserAddresses(user string, addrs []ethereum.Address, timestamps []uint64) error {
	user = strings.ToLower(user)
	addresses := []string{}
	for _, addr := range addrs {
		addresses = append(addresses, common.AddrToString(addr))
	}
	var err error
	err = self.db.Update(func(tx *bolt.Tx) error {
		timeBucket := tx.Bucket([]byte(ADDRESS_TIME))
		for _, address := range addresses {
			// get temp user identity
			b := tx.Bucket([]byte(ADDRESS_ID))
			oldID := b.Get([]byte(address))
			// remove the addresses bucket assocciated to this temp user
			b = tx.Bucket([]byte(ID_ADDRESSES))
			if oldID != nil {
				if _, vErr := b.CreateBucketIfNotExists(oldID); err != nil {
					return vErr
				}

				if vErr := b.DeleteBucket(oldID); vErr != nil {
					return vErr
				}
			}
			vErr := timeBucket.Delete([]byte(address))
			if vErr != nil {
				return vErr
			}
			// update user to each address => user
			b = tx.Bucket([]byte(ADDRESS_ID))
			if vErr := b.Put([]byte(address), []byte(user)); vErr != nil {
				return vErr
			}
		}
		// remove old addresses from pending bucket
		pendingBk := tx.Bucket([]byte(PENDING_ADDRESSES))
		oldAddrs, _, vErr := self.GetAddressesOfUser(user)
		if vErr != nil {
			return vErr
		}
		for _, oldAddr := range oldAddrs {
			if uErr := pendingBk.Delete([]byte(common.AddrToString(oldAddr))); uErr != nil {
				return uErr
			}
		}
		// update addresses bucket for real user
		// add new addresses to pending bucket
		b := tx.Bucket([]byte(ID_ADDRESSES))
		userBucket, vErr := b.CreateBucketIfNotExists([]byte(user))
		if vErr != nil {
			return vErr
		}
		catBk := tx.Bucket([]byte(ADDRESS_CATEGORY))
		for i, address := range addresses {
			if uErr := userBucket.Put([]byte(address), []byte{1}); uErr != nil {
				return uErr
			}
			cat := catBk.Get([]byte(address))
			if string(cat) != KYC_CATEGORY {
				if tErr := pendingBk.Put([]byte(address), []byte{1}); tErr != nil {
					return tErr
				}
			}
			log.Printf("storing timestamp for %s - %d", address, timestamps[i])
			if err = timeBucket.Put([]byte(address), boltutil.Uint64ToBytes(timestamps[i])); err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

// returns lowercased category of an address
func (self *BoltUserStorage) GetCategory(ethaddr ethereum.Address) (string, error) {
	addr := common.AddrToString(ethaddr)
	var err error
	var result string
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ADDRESS_CATEGORY))
		cat := b.Get([]byte(addr))
		result = string(cat)
		return nil
	})
	return result, err
}

func (self *BoltUserStorage) GetAddressesOfUser(user string) ([]ethereum.Address, []uint64, error) {
	var err error
	user = strings.ToLower(user)
	result := []ethereum.Address{}
	timestamps := []uint64{}
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ID_ADDRESSES))
		timeBucket := tx.Bucket([]byte(ADDRESS_TIME))
		userBucket := b.Bucket([]byte(user))
		if userBucket != nil {
			err = userBucket.ForEach(func(k, v []byte) error {
				addr := ethereum.HexToAddress(string(k))
				result = append(result, addr)
				timestamps = append(timestamps, boltutil.BytesToUint64(timeBucket.Get(k)))
				return nil
			})
		}
		return err
	})
	return result, timestamps, err
}

// returns lowercased user identity of the address
func (self *BoltUserStorage) GetUserOfAddress(ethaddr ethereum.Address) (string, uint64, error) {
	addr := common.AddrToString(ethaddr)
	var err error
	var result string
	var timestamp uint64
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ADDRESS_ID))
		timeBucket := tx.Bucket([]byte(ADDRESS_TIME))
		id := b.Get([]byte(addr))
		result = string(id)
		timestamp = boltutil.BytesToUint64(timeBucket.Get([]byte(addr)))
		return nil
	})
	return result, timestamp, err
}

func (self *BoltUserStorage) GetKycUsers() (map[string]uint64, error) {
	result := map[string]uint64{}
	err := self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ADDRESS_ID))
		timeBucket := tx.Bucket([]byte(ADDRESS_TIME))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			id := string(v)
			var timestamp uint64
			if id != "" && id != string(k) {
				timestamp = boltutil.BytesToUint64(timeBucket.Get(k))
				result[string(k)] = timestamp
			}
		}
		return nil
	})
	return result, err
}

// returns all of addresses that's not pushed to the chain
// for kyced category
func (self *BoltUserStorage) GetPendingAddresses() ([]ethereum.Address, error) {
	var err error
	result := []ethereum.Address{}
	err = self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(PENDING_ADDRESSES))
		err = b.ForEach(func(k, v []byte) error {
			result = append(result, ethereum.HexToAddress(string(k)))
			return nil
		})
		return err
	})
	return result, err
}
