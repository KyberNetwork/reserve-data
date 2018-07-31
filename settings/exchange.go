package settings

import (
	"log"

	"github.com/KyberNetwork/reserve-data/common"
)

// GetFee returns a map[tokenID]exchangeFees and error if occur
func (setting *Settings) GetFee(ex ExchangeName) (common.ExchangeFees, error) {
	return setting.Exchange.Storage.GetFee(ex)
}

// UpdateFee will merge the current fee setting to the new fee setting,
// Any different will be overwriten from new fee to cufrent fee
// Afterwhich it stores the fee with exchangeName as key into database and return error if occur
func (setting *Settings) UpdateFee(exName ExchangeName, exFee common.ExchangeFees) error {
	currExFee, err := setting.GetFee(exName)
	var currFundingFee = common.NewFundingFee(make(map[string]float64), make(map[string]float64))
	if err != nil {
		log.Printf("UpdateExchangeFee: Can't get current exchange fee of %s (%s), overwrite it with new data", exName.String(), err)
		currExFee = common.NewExchangeFee(make(common.TradingFee), currFundingFee)
	}
	for tok, val := range exFee.Funding.Deposit {
		currExFee.Funding.Deposit[tok] = val
	}
	for tok, val := range exFee.Funding.Withdraw {
		currExFee.Funding.Withdraw[tok] = val
	}
	for tok, val := range exFee.Trading {
		currExFee.Trading[tok] = val
	}
	return setting.Exchange.Storage.StoreFee(exName, currExFee)
}

// GetMinDeposit returns a map[tokenID]MinDeposit and error if occur
func (setting *Settings) GetMinDeposit(ex ExchangeName) (common.ExchangesMinDeposit, error) {
	return setting.Exchange.Storage.GetMinDeposit(ex)
}

// UpdateMinDeposit will merge the current min Deposit to the new min Deposit,
// Any different will be overwriten from new minDeposit to cufrent minDeposit
// Afterwhich it stores the fee with exchangeName as key into database and return error if occur
func (setting *Settings) UpdateMinDeposit(exName ExchangeName, minDeposit common.ExchangesMinDeposit) error {
	currExMinDep, err := setting.GetMinDeposit(exName)
	if err != nil {
		log.Printf("UpdateMinDeposit: Can't get current min deposit of %s (%s), overwrite it with new data", exName.String(), err)
		currExMinDep = make(common.ExchangesMinDeposit)
	}
	for tok, val := range minDeposit {
		currExMinDep[tok] = val
	}
	return setting.Exchange.Storage.StoreMinDeposit(exName, currExMinDep)
}

// GetDepositAddresses returns a map[tokenID]DepositAddress and error if occur
func (setting *Settings) GetDepositAddresses(ex ExchangeName) (common.ExchangeAddresses, error) {
	return setting.Exchange.Storage.GetDepositAddresses(ex)
}

// Update get the deposit Addresses with exchangeName as key, change the desired deposit address
// then store into database and return error if occur
func (setting *Settings) UpdateDepositAddress(exName ExchangeName, addrs common.ExchangeAddresses) error {
	currAddrs, err := setting.GetDepositAddresses(exName)
	if err != nil {
		log.Printf("UpdateDepositAddress: Can't get current deposit address of %s (%s), overwrite it with new data", exName.String(), err)
		currAddrs = make(common.ExchangeAddresses)
	}
	for tokenID, address := range addrs {
		currAddrs.Update(tokenID, address)
	}
	return setting.Exchange.Storage.StoreDepositAddress(exName, currAddrs)
}

// GetExchangeInfor returns the an ExchangeInfo Object for each exchange
// and error if occur
func (setting *Settings) GetExchangeInfo(ex ExchangeName) (common.ExchangeInfo, error) {
	return setting.Exchange.Storage.GetExchangeInfo(ex)
}

// UpdateExchangeInfo will merge the new exchange info into current exchange info , the
// updates exchange info object using exchangeName as key
// returns error if occur
func (setting *Settings) UpdateExchangeInfo(exName ExchangeName, exInfo common.ExchangeInfo) error {
	currExInfo, err := setting.GetExchangeInfo(exName)
	if err != nil {
		log.Printf("UpdateExchangeInfo: Can't get exchange Info of %s (%s), overwrite it with new data", exName.String(), err)
		currExInfo = common.NewExchangeInfo()
	}
	for tokenPairID, exPreLim := range exInfo {
		currExInfo[tokenPairID] = exPreLim
	}
	return setting.Exchange.Storage.StoreExchangeInfo(exName, currExInfo)
}

func (setting *Settings) GetExchangeStatus() (common.ExchangesStatus, error) {
	return setting.Exchange.Storage.GetExchangeStatus()
}

func (setting *Settings) UpdateExchangeStatus(exStatus common.ExchangesStatus) error {
	return setting.Exchange.Storage.StoreExchangeStatus(exStatus)
}

func (setting *Settings) GetExchangeNotifications() (common.ExchangeNotifications, error) {
	return setting.Exchange.Storage.GetExchangeNotifications()
}

func (setting *Settings) UpdateExchangeNotification(exchange, action, tokenPair string, fromTime, toTime uint64, isWarning bool, msg string) error {
	return setting.Exchange.Storage.StoreExchangeNotification(exchange, action, tokenPair, fromTime, toTime, isWarning, msg)
}
