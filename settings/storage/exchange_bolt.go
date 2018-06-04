package storage

import (
	"encoding/json"
	"fmt"

	"github.com/KyberNetwork/reserve-data/boltutil"
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/settings"
	"github.com/boltdb/bolt"
)

// GetFee returns a map[tokenID]exchangeFees and error if occur
func (boltSettingStorage *BoltSettingStorage) GetFee(ex settings.ExchangeName) (common.ExchangeFees, error) {
	var result common.ExchangeFees
	err := boltSettingStorage.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(EXCHANGE_FEE_BUCKET))
		if b == nil {
			return fmt.Errorf("bucket %s hasn't existed yet", EXCHANGE_FEE_BUCKET)
		}
		data := b.Get(boltutil.Uint64ToBytes(uint64(ex)))
		if data == nil {
			return fmt.Errorf("key %s hasn't existed yet", ex.String())
		}
		uErr := json.Unmarshal(data, &result)
		if uErr != nil {
			return uErr
		}
		return nil
	})
	return result, err
}

// StoreFee stores the fee with exchangeName as key into database and return error if occur
func (boltSettingStorage *BoltSettingStorage) StoreFee(ex settings.ExchangeName, data common.ExchangeFees) error {
	err := boltSettingStorage.db.Update(func(tx *bolt.Tx) error {
		b, uErr := tx.CreateBucketIfNotExists([]byte(EXCHANGE_FEE_BUCKET))
		if uErr != nil {
			return uErr
		}
		dataJSON, uErr := json.Marshal(data)
		if uErr != nil {
			return uErr
		}
		return b.Put(boltutil.Uint64ToBytes(uint64(ex)), dataJSON)
	})
	return err
}

// GetMinDeposit returns a map[tokenID]MinDeposit and error if occur
func (boltSettingStorage *BoltSettingStorage) GetMinDeposit(ex settings.ExchangeName) (common.ExchangesMinDeposit, error) {
	result := make(common.ExchangesMinDeposit)
	err := boltSettingStorage.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(EXCHANGE_MIN_DEPOSIT_BUCKET))
		if b == nil {
			return fmt.Errorf("bucket %s hasn't existed yet", EXCHANGE_MIN_DEPOSIT_BUCKET)
		}
		data := b.Get(boltutil.Uint64ToBytes(uint64(ex)))
		if data == nil {
			return fmt.Errorf("key %s hasn't existed yet", ex.String())
		}
		uErr := json.Unmarshal(data, &result)
		if uErr != nil {
			return uErr
		}
		return nil
	})
	return result, err
}

// StoreMinDeposit stores the minDeposit with exchangeName as key into database and return error if occur
func (boltSettingStorage *BoltSettingStorage) StoreMinDeposit(ex settings.ExchangeName, data common.ExchangesMinDeposit) error {
	err := boltSettingStorage.db.Update(func(tx *bolt.Tx) error {
		b, uErr := tx.CreateBucketIfNotExists([]byte(EXCHANGE_MIN_DEPOSIT_BUCKET))
		if uErr != nil {
			return uErr
		}
		dataJSON, uErr := json.Marshal(data)
		if uErr != nil {
			return uErr
		}
		return b.Put(boltutil.Uint64ToBytes(uint64(ex)), dataJSON)
	})
	return err
}

// GetDepositAddress returns a map[tokenID]DepositAddress and error if occur
func (boltSettingStorage *BoltSettingStorage) GetDepositAddress(ex settings.ExchangeName) (common.ExchangeAddresses, error) {
	result := make(common.ExchangeAddresses)
	err := boltSettingStorage.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(EXCHANGE_DEPOSIT_ADDRESS))
		if b == nil {
			return fmt.Errorf("bucket %s hasn't existed yet", EXCHANGE_DEPOSIT_ADDRESS)
		}
		data := b.Get(boltutil.Uint64ToBytes(uint64(ex)))
		if data == nil {
			return fmt.Errorf("key %s hasn't existed yet", ex.String())
		}
		uErr := json.Unmarshal(data, &result)
		if uErr != nil {
			return uErr
		}
		return nil
	})
	return result, err

}

// StoreDepositAddress stores the depositAddress with exchangeName as key into database and
// return error if occur
func (boltSettingStorage *BoltSettingStorage) StoreDepositAddress(ex settings.ExchangeName, addrs common.ExchangeAddresses) error {
	err := boltSettingStorage.db.Update(func(tx *bolt.Tx) error {
		b, uErr := tx.CreateBucketIfNotExists([]byte(EXCHANGE_DEPOSIT_ADDRESS))
		if uErr != nil {
			return uErr
		}
		dataJSON, uErr := json.Marshal(addrs)
		if uErr != nil {
			return uErr
		}
		return b.Put(boltutil.Uint64ToBytes(uint64(ex)), dataJSON)
	})
	return err
}

// GetTokenPairs returns a list of TokenPairs available at current exchange
// return error if occur
func (boltSettingStorage *BoltSettingStorage) GetTokenPairs(ex settings.ExchangeName) ([]common.TokenPair, error) {
	var result []common.TokenPair
	err := boltSettingStorage.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(EXCHANGE_TOKEN_PAIRS))
		if b == nil {
			return fmt.Errorf("bucket %s hasn't existed yet", EXCHANGE_TOKEN_PAIRS)
		}
		data := b.Get(boltutil.Uint64ToBytes(uint64(ex)))
		if data == nil {
			return fmt.Errorf("key %s hasn't existed yet", ex.String())
		}
		if uErr := json.Unmarshal(data, &result); uErr != nil {
			return uErr
		}
		return nil
	})
	return result, err
}

// StoreTokenPairs store the list of TokenPairs with exchangeName as key into database and
// return error if occur
func (boltSettingStorage *BoltSettingStorage) StoreTokenPairs(ex settings.ExchangeName, data []common.TokenPair) error {
	err := boltSettingStorage.db.Update(func(tx *bolt.Tx) error {
		b, uErr := tx.CreateBucketIfNotExists([]byte(EXCHANGE_TOKEN_PAIRS))
		if uErr != nil {
			return uErr
		}
		dataJSON, uErr := json.Marshal(data)
		if uErr != nil {
			return uErr
		}
		return b.Put(boltutil.Uint64ToBytes(uint64(ex)), dataJSON)
	})
	return err
}

func (boltSettingStorage *BoltSettingStorage) GetExchangeInfo(ex settings.ExchangeName) (*common.ExchangeInfo, error) {
	var result common.ExchangeInfo
	err := boltSettingStorage.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(EXCHANGE_INFO))
		if b == nil {
			return fmt.Errorf("bucket %s hasn't existed yet", EXCHANGE_INFO)
		}
		data := b.Get(boltutil.Uint64ToBytes(uint64(ex)))
		if data == nil {
			return fmt.Errorf("key %s hasn't existed yet", ex.String())
		}
		if uErr := json.Unmarshal(data, &result); uErr != nil {
			return uErr
		}
		return nil
	})
	return &result, err
}

func (boltSettingStorage *BoltSettingStorage) StoreExchangeInfo(ex settings.ExchangeName, exInfo *common.ExchangeInfo) error {
	err := boltSettingStorage.db.Update(func(tx *bolt.Tx) error {
		b, uErr := tx.CreateBucketIfNotExists([]byte(EXCHANGE_INFO))
		if uErr != nil {
			return uErr
		}
		dataJson, uErr := json.Marshal(exInfo)
		if uErr != nil {
			return uErr
		}
		return b.Put(boltutil.Uint64ToBytes(uint64(ex)), dataJson)
	})
	return err
}
