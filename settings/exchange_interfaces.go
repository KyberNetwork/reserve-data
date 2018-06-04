package settings

import (
	"github.com/KyberNetwork/reserve-data/common"
	ethereum "github.com/ethereum/go-ethereum/common"
)

// GetFee returns a map[tokenID]exchangeFees and error if occur
func (setting *Settings) GetFee(ex ExchangeName) (common.ExchangeFees, error) {
	return setting.Exchange.Storage.GetFee(ex)
}

// StoreFee stores the fee with exchangeName as key into database and return error if occur
func (setting *Settings) StoreFee(ex ExchangeName, data common.ExchangeFees) error {
	return setting.Exchange.Storage.StoreFee(ex, data)
}

// GetMinDeposit returns a map[tokenID]MinDeposit and error if occur
func (setting *Settings) GetMinDeposit(ex ExchangeName) (common.ExchangesMinDeposit, error) {
	return setting.Exchange.Storage.GetMinDeposit(ex)
}

// StoreMinDeposit stores the minDeposit with exchangeName as key into database and return error if occur
func (setting *Settings) StoreMinDeposit(ex ExchangeName, minDeposit common.ExchangesMinDeposit) error {
	return setting.Exchange.Storage.StoreMinDeposit(ex, minDeposit)
}

// GetDepositAddress returns a map[tokenID]DepositAddress and error if occur
func (setting *Settings) GetDepositAddress(ex ExchangeName) (common.ExchangeAddresses, error) {
	return setting.Exchange.Storage.GetDepositAddress(ex)
}

// Update get the deposit Addresses with exchangeName as key, change the desired deposit address
// then store into database and return error if occur
func (setting *Settings) UpdateDepositAddress(ex ExchangeName, token common.Token) error {
	addrs, err := setting.GetDepositAddress(ex)
	if err != nil {
		return err
	}
	addrs.Update(token.ID, ethereum.HexToAddress(token.Address))
	return setting.Exchange.Storage.StoreDepositAddress(ex, addrs)
}

// GetTokenPairs returns a list of TokenPairs available at current exchange
// return error if occur
func (setting *Settings) GetTokenPairs(ex ExchangeName) ([]common.TokenPair, error) {
	return setting.Exchange.Storage.GetTokenPairs(ex)
}

// StoreTokenPairs store the list of TokenPairs with exchangeName as key into database and
// return error if occur
func (setting *Settings) UpdateTokenPairs(ex ExchangeName, data []common.TokenPair) error {
	return setting.Exchange.Storage.StoreTokenPairs(ex, data)
}

// GetExchangeInfor returns the an ExchangeInfo Object for each exchange
// and error if occur
func (setting *Settings) GetExchangeInfo(ex ExchangeName) (*common.ExchangeInfo, error) {
	return setting.Exchange.Storage.GetExchangeInfo(ex)
}

// UpdateExchangeInfo updates exchange info object using exchangeName as key
// returns error if occur
func (setting *Settings) UpdateExchangeInfo(ex ExchangeName, exInfo *common.ExchangeInfo) error {
	return setting.Exchange.Storage.StoreExchangeInfo(ex, exInfo)
}
