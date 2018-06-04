package settings

import (
	"github.com/KyberNetwork/reserve-data/common"
)

type ExchangeStorage interface {
	// GetFee returns a map[tokenID]exchangeFees and error if occur
	// If there is no exchangeFee matched with key param, error is returned as well
	GetFee(ex ExchangeName) (common.ExchangeFees, error)
	// StoreFee stores the fee with exchangeName as key into database and return error if occur
	StoreFee(ex ExchangeName, data common.ExchangeFees) error
	// GetMinDeposit returns a map[tokenID]MinDeposit and error if occur
	GetMinDeposit(ex ExchangeName) (common.ExchangesMinDeposit, error)
	// StoreMinDeposit stores the minDeposit with exchangeName as key into database and return error if occur
	StoreMinDeposit(ex ExchangeName, minDeposit common.ExchangesMinDeposit) error
	// GetDepositAddress returns a map[tokenID]DepositAddress and error if occur
	GetDepositAddress(ex ExchangeName) (common.ExchangeAddresses, error)
	// StoreDepositAddress stores the depositAddress with exchangeName as key into database and
	// return error if occur
	StoreDepositAddress(ex ExchangeName, addrs common.ExchangeAddresses) error
	// GetTokenPairs returns a list of TokenPairs available at current exchange
	// return error if occur
	GetTokenPairs(ex ExchangeName) ([]common.TokenPair, error)
	// StoreTokenPairs store the list of TokenPairs with exchangeName as key into database and
	// return error if occur
	StoreTokenPairs(ex ExchangeName, data []common.TokenPair) error
	// GetExchangeInfo returns the an ExchangeInfo Object for each exchange
	// and error if occur
	GetExchangeInfo(ex ExchangeName) (*common.ExchangeInfo, error)
	// StoreExchangeInfo store the ExchangeInfo Object using the exchangeName as the key into database
	// return error if occur
	StoreExchangeInfo(ex ExchangeName, exInfo *common.ExchangeInfo) error
}
