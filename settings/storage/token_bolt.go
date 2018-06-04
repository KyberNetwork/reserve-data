package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/KyberNetwork/reserve-data/boltutil"
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/settings"
	"github.com/boltdb/bolt"
	ethereum "github.com/ethereum/go-ethereum/common"
)

func addTokenByID(tx *bolt.Tx, t common.Token) error {
	b, uErr := tx.CreateBucketIfNotExists([]byte(TOKEN_BUCKET_BY_ID))
	if uErr != nil {
		return uErr
	}
	dataJSON, uErr := json.Marshal(t)
	if uErr != nil {
		return uErr
	}
	return b.Put([]byte(strings.ToLower(t.ID)), dataJSON)
}

func addTokenByAddress(tx *bolt.Tx, t common.Token) error {
	b, uErr := tx.CreateBucketIfNotExists([]byte(TOKEN_BUCKET_BY_ADDRESS))
	if uErr != nil {
		return uErr
	}
	dataJson, uErr := json.Marshal(t)
	if uErr != nil {
		return uErr
	}
	return b.Put([]byte(strings.ToLower(t.Address)), dataJson)
}

func (boltSettingStorage *BoltSettingStorage) UpdateToken(t common.Token) error {
	err := boltSettingStorage.db.Update(func(tx *bolt.Tx) error {
		if uErr := addTokenByID(tx, t); uErr != nil {
			return uErr
		}
		if uErr := addTokenByAddress(tx, t); uErr != nil {
			return uErr
		}
		return nil
	})
	return err
}

func (boltSettingStorage *BoltSettingStorage) AddTokenByID(t common.Token) error {
	err := boltSettingStorage.db.Update(func(tx *bolt.Tx) error {
		b, uErr := tx.CreateBucketIfNotExists([]byte(TOKEN_BUCKET_BY_ID))
		if uErr != nil {
			return uErr
		}
		dataJSON, uErr := json.Marshal(t)
		if uErr != nil {
			return uErr
		}
		return b.Put([]byte(strings.ToLower(t.ID)), dataJSON)
	})
	return err
}

func (boltSettingStorage *BoltSettingStorage) AddTokenByAddress(t common.Token) error {
	err := boltSettingStorage.db.Update(func(tx *bolt.Tx) error {
		b, uErr := tx.CreateBucketIfNotExists([]byte(TOKEN_BUCKET_BY_ADDRESS))
		if uErr != nil {
			return uErr
		}
		dataJson, uErr := json.Marshal(t)
		if uErr != nil {
			return uErr
		}
		return b.Put([]byte(strings.ToLower(t.Address)), dataJson)
	})
	return err
}

func (boltSettingStorage *BoltSettingStorage) getTokensWithFilter(filter FilterFunction) (result []common.Token, err error) {
	err = boltSettingStorage.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TOKEN_BUCKET_BY_ID))
		if b == nil {
			return fmt.Errorf("Bucket doesn't exist yet")
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var temp common.Token
			uErr := json.Unmarshal(v, &temp)
			if uErr != nil {
				return uErr
			}
			if filter(temp) {
				result = append(result, temp)
			}
		}
		return nil
	})
	return result, err
}

func (boltSettingStorage *BoltSettingStorage) GetAllTokens() (result []common.Token, err error) {
	return boltSettingStorage.getTokensWithFilter(isToken)
}

func (boltSettingStorage *BoltSettingStorage) GetActiveTokens() (result []common.Token, err error) {
	return boltSettingStorage.getTokensWithFilter(isActive)
}

func (boltSettingStorage *BoltSettingStorage) GetInternalTokens() (result []common.Token, err error) {
	return boltSettingStorage.getTokensWithFilter(isInternal)
}

func (boltSettingStorage *BoltSettingStorage) GetExternalTokens() (result []common.Token, err error) {
	return boltSettingStorage.getTokensWithFilter(isExternal)
}

func (boltSettingStorage *BoltSettingStorage) getTokenByIDWithFiltering(id string, filter FilterFunction) (t common.Token, err error) {
	id = strings.ToLower(id)
	err = boltSettingStorage.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TOKEN_BUCKET_BY_ID))
		if b == nil {
			return fmt.Errorf("Bucket doesn't exist yet")
		}
		c := b.Cursor()
		k, v := c.Seek([]byte(id))
		if bytes.Compare(k, []byte(id)) != 0 {
			return fmt.Errorf("Token %s is not found in current setting", id)
		}
		uErr := json.Unmarshal(v, &t)
		if uErr != nil {
			return uErr
		}
		if !filter(t) {
			return fmt.Errorf("Token %s status is not as querried", id)
		}
		return nil
	})
	return t, err
}

func (boltSettingStorage *BoltSettingStorage) GetTokenByID(id string) (common.Token, error) {
	return boltSettingStorage.getTokenByIDWithFiltering(id, isToken)
}

func (boltSettingStorage *BoltSettingStorage) GetActiveTokenByID(id string) (common.Token, error) {
	return boltSettingStorage.getTokenByIDWithFiltering(id, isActive)
}

func (boltSettingStorage *BoltSettingStorage) GetInternalTokenByID(id string) (common.Token, error) {
	return boltSettingStorage.getTokenByIDWithFiltering(id, isInternal)
}

func (boltSettingStorage *BoltSettingStorage) GetExternalTokenByID(id string) (common.Token, error) {
	return boltSettingStorage.getTokenByIDWithFiltering(id, isExternal)
}

func (boltSettingStorage *BoltSettingStorage) getTokenByAddressWithFiltering(addr string, filter FilterFunction) (t common.Token, err error) {
	addr = strings.ToLower(addr)
	err = boltSettingStorage.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TOKEN_BUCKET_BY_ADDRESS))
		if b == nil {
			return fmt.Errorf("Bucket doesn't exist yet")
		}
		c := b.Cursor()
		k, v := c.Seek([]byte(addr))
		if bytes.Compare(k, []byte(addr)) != 0 {
			return fmt.Errorf("Token %s is not found in current setting", addr)
		}
		uErr := json.Unmarshal(v, &t)
		if uErr != nil {
			return uErr
		}
		if !filter(t) {
			return fmt.Errorf("Token %s status is not as querried", t.ID)
		}
		return nil
	})
	return t, err
}

func (boltSettingStorage *BoltSettingStorage) GetActiveTokenByAddress(Addr ethereum.Address) (common.Token, error) {
	return boltSettingStorage.getTokenByAddressWithFiltering(Addr.Hex(), isActive)
}

func (boltSettingStorage *BoltSettingStorage) GetTokenByAddress(Addr ethereum.Address) (common.Token, error) {
	return boltSettingStorage.getTokenByAddressWithFiltering(Addr.Hex(), isToken)
}

func (boltSettingStorage *BoltSettingStorage) GetInternalTokenByAddress(Addr ethereum.Address) (common.Token, error) {
	return boltSettingStorage.getTokenByAddressWithFiltering(Addr.Hex(), isInternal)
}

func (boltSettingStorage *BoltSettingStorage) GetExternalTokenByAddress(Addr ethereum.Address) (common.Token, error) {
	return boltSettingStorage.getTokenByAddressWithFiltering(Addr.Hex(), isExternal)
}

func (boltSettingStorage *BoltSettingStorage) UpdateOneAddress(name settings.AddressName, address string) error {
	address = strings.ToLower(address)
	err := boltSettingStorage.db.Update(func(tx *bolt.Tx) error {
		b, uErr := tx.CreateBucketIfNotExists([]byte(ADDRESS_SETTING_BUCKET))
		if uErr != nil {
			return uErr
		}
		return b.Put(boltutil.Uint64ToBytes(uint64(name)), []byte(address))
	})
	return err
}
