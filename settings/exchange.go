package settings

import (
	"github.com/KyberNetwork/reserve-data/common"
)

// GetFee returns a map[tokenID]exchangeFees and error if occur
func (s *Settings) GetFee(ex ExchangeName) (common.ExchangeFees, error) {
	return s.Exchange.Storage.GetFee(ex)
}

// UpdateFee will merge the current fee setting to the new fee setting,
// Any different will be overwriten from new fee to cufrent fee
// Afterwhich it stores the fee with exchangeName as key into database and return error if occur
func (s *Settings) UpdateFee(exName ExchangeName, exFee common.ExchangeFees, timestamp uint64) error {
	if timestamp == 0 {
		timestamp = common.GetTimepoint()
	}
	currExFee, err := s.GetFee(exName)
	if err != nil {
		if err != ErrExchangeRecordNotFound {
			return err
		}
		s.l.Infof("UpdateExchangeFee: the current exchange fee of %s hasn't existed yet, overwrite it with new data", exName.String())
		currExFee = common.NewExchangeFee(common.TradingFee{}, common.NewFundingFee(make(map[string]float64), make(map[string]float64)))
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
	return s.Exchange.Storage.StoreFee(exName, currExFee, timestamp)
}

// GetMinDeposit returns a map[tokenID]MinDeposit and error if occur
func (s *Settings) GetMinDeposit(ex ExchangeName) (common.ExchangesMinDeposit, error) {
	return s.Exchange.Storage.GetMinDeposit(ex)
}

// UpdateMinDeposit will merge the current min Deposit to the new min Deposit,
// Any different will be overwriten from new minDeposit to cufrent minDeposit
// Afterwhich it stores the fee with exchangeName as key into database and return error if occur
func (s *Settings) UpdateMinDeposit(exName ExchangeName, minDeposit common.ExchangesMinDeposit, timestamp uint64) error {
	if timestamp == 0 {
		timestamp = common.GetTimepoint()
	}
	currExMinDep, err := s.GetMinDeposit(exName)
	if err != nil {
		if err != ErrExchangeRecordNotFound {
			return err
		}
		s.l.Infof("UpdateMinDeposit: Can't get current min deposit of %s, overwrite it with new data", exName.String())
		currExMinDep = make(common.ExchangesMinDeposit)
	}
	for tok, val := range minDeposit {
		currExMinDep[tok] = val
	}
	return s.Exchange.Storage.StoreMinDeposit(exName, currExMinDep, timestamp)
}

// GetDepositAddresses returns a map[tokenID]DepositAddress and error if occur
func (s *Settings) GetDepositAddresses(ex ExchangeName) (common.ExchangeAddresses, error) {
	return s.Exchange.Storage.GetDepositAddresses(ex)
}

// Update get the deposit Addresses with exchangeName as key, change the desired deposit address
// then store into database and return error if occur
func (s *Settings) UpdateDepositAddress(exName ExchangeName, addrs common.ExchangeAddresses, timestamp uint64) error {
	if timestamp == 0 {
		timestamp = common.GetTimepoint()
	}
	currAddrs, err := s.GetDepositAddresses(exName)
	if err != nil {
		if err != ErrExchangeRecordNotFound {
			return err
		}
		s.l.Infof("UpdateDepositAddress: the current exchange deposit addresses for %s hasn't existed yet. Overwrite new setting instead", exName.String())
		currAddrs = make(common.ExchangeAddresses)
	}
	for tokenID, address := range addrs {
		currAddrs.Update(tokenID, address)
	}
	return s.Exchange.Storage.StoreDepositAddress(exName, currAddrs, timestamp)
}

// GetExchangeInfor returns the an ExchangeInfo Object for each exchange
// and error if occur
func (s *Settings) GetExchangeInfo(ex ExchangeName) (common.ExchangeInfo, error) {
	return s.Exchange.Storage.GetExchangeInfo(ex)
}

// UpdateExchangeInfo will merge the new exchange info into current exchange info , the
// updates exchange info object using exchangeName as key
// returns error if occur
func (s *Settings) UpdateExchangeInfo(exName ExchangeName, exInfo common.ExchangeInfo, timestamp uint64) error {
	if timestamp == 0 {
		timestamp = common.GetTimepoint()
	}
	currExInfo, err := s.GetExchangeInfo(exName)
	if err != nil {
		if err != ErrExchangeRecordNotFound {
			return err
		}
		s.l.Infof("UpdateExchangeInfo: the current exchange Info for %s hasn't existed yet. Overwrite new setting instead", exName.String())
		currExInfo = common.NewExchangeInfo()
	}
	for tokenPairID, exPreLim := range exInfo {
		currExInfo[tokenPairID] = exPreLim
	}
	return s.Exchange.Storage.StoreExchangeInfo(exName, currExInfo, timestamp)
}

func (s *Settings) GetExchangeStatus() (common.ExchangesStatus, error) {
	return s.Exchange.Storage.GetExchangeStatus()
}

func (s *Settings) UpdateExchangeStatus(exStatus common.ExchangesStatus) error {
	return s.Exchange.Storage.StoreExchangeStatus(exStatus)
}

func (s *Settings) GetExchangeNotifications() (common.ExchangeNotifications, error) {
	return s.Exchange.Storage.GetExchangeNotifications()
}

func (s *Settings) UpdateExchangeNotification(exchange, action, tokenPair string, fromTime, toTime uint64, isWarning bool, msg string) error {
	return s.Exchange.Storage.StoreExchangeNotification(exchange, action, tokenPair, fromTime, toTime, isWarning, msg)
}

func (s *Settings) GetExchangeVersion() (uint64, error) {
	return s.Exchange.Storage.GetExchangeVersion()
}
