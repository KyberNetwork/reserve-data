package storage

import (
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/boltdb/bolt"
)

const (
	TokenBucketByID          string = "token_by_id"
	TokenBucketByAddress     string = "token_by_addr"
	ExchangeFeeBucket        string = "exchange_fee"
	ExchangeMinDepositBucket string = "exchange_min_deposit"
	ExchangeDepositAddress   string = "exchange_deposit_address"
	ExchangeTokenPairs       string = "exchange_token_pairs"
	ExchangeInfo             string = "exchange_info"
	ExchangeStatus           string = "exchange_status"
	ExchangeNotifications    string = "exchange_notifications"
	PendingTokenRequest      string = "pending_token_request"
)

type FilterFunction func(common.Token) bool

func isActive(t common.Token) bool {
	return t.Active
}

func isToken(_ common.Token) bool {
	return true
}

func isInternal(t common.Token) bool {
	return t.Active && t.Internal
}

func isExternal(t common.Token) bool {
	return t.Active && !t.Internal
}

type BoltSettingStorage struct {
	db *bolt.DB
}

func NewBoltSettingStorage(dbPath string) (*BoltSettingStorage, error) {
	var err error
	var db *bolt.DB
	db, err = bolt.Open(dbPath, 0600, nil)
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		if _, uErr := tx.CreateBucketIfNotExists([]byte(TokenBucketByID)); uErr != nil {
			return uErr
		}
		if _, uErr := tx.CreateBucketIfNotExists([]byte(TokenBucketByAddress)); uErr != nil {
			return uErr
		}
		if _, uErr := tx.CreateBucketIfNotExists([]byte(ExchangeFeeBucket)); uErr != nil {
			return uErr
		}
		if _, uErr := tx.CreateBucketIfNotExists([]byte(ExchangeMinDepositBucket)); uErr != nil {
			return uErr
		}
		if _, uErr := tx.CreateBucketIfNotExists([]byte(ExchangeDepositAddress)); uErr != nil {
			return uErr
		}
		if _, uErr := tx.CreateBucketIfNotExists([]byte(ExchangeTokenPairs)); uErr != nil {
			return uErr
		}
		if _, uErr := tx.CreateBucketIfNotExists([]byte(ExchangeInfo)); uErr != nil {
			return uErr
		}
		if _, uErr := tx.CreateBucketIfNotExists([]byte(ExchangeStatus)); uErr != nil {
			return uErr
		}
		if _, uErr := tx.CreateBucketIfNotExists([]byte(ExchangeNotifications)); uErr != nil {
			return uErr
		}
		if _, uErr := tx.CreateBucketIfNotExists([]byte(PendingTokenRequest)); uErr != nil {
			return uErr
		}
		if _, uErr := tx.CreateBucketIfNotExists([]byte(tokenVersion)); uErr != nil {
			return uErr
		}
		if _, uErr := tx.CreateBucketIfNotExists([]byte(exchangeVersion)); uErr != nil {
			return uErr
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	storage := BoltSettingStorage{db}
	return &storage, nil
}
