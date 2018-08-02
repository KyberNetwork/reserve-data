package mock

import (
	"github.com/KyberNetwork/reserve-data/common"
)

type TestReserveData struct {
}

func (trd TestReserveData) CurrentPriceVersion(timepoint uint64) (common.Version, error) {
	return common.Version(0), nil
}

func (trd TestReserveData) GetAllPrices(timepoint uint64) (common.AllPriceResponse, error) {
	return common.AllPriceResponse{}, nil
}

func (trd TestReserveData) GetOnePrice(pairID common.TokenPairID, timepoint uint64) (common.OnePriceResponse, error) {
	return common.OnePriceResponse{}, nil
}

func (trd TestReserveData) CurrentAuthDataVersion(timestamp uint64) (common.Version, error) {
	return common.Version(0), nil
}

func (trd TestReserveData) GetAuthData(timestamp uint64) (common.AuthDataResponse, error) {
	return common.AuthDataResponse{}, nil
}

func (trd TestReserveData) GetRate(timestamp uint64) (common.AllRateResponse, error) {
	return common.AllRateResponse{}, nil
}

// GetRates returns list of valid rates for all tokens that is collected between [fromTime, toTime).
func (trd TestReserveData) GetRates(fromTime, toTime uint64) ([]common.AllRateResponse, error) {
	return nil, nil
}

func (trd TestReserveData) GetRecords(fromTime, toTime uint64) ([]common.ActivityRecord, error) {
	return nil, nil
}
func (trd TestReserveData) GetPendingActivities() ([]common.ActivityRecord, error) {
	return nil, nil
}

func (trd TestReserveData) GetGoldData(timepoint uint64) (common.GoldData, error) {
	return common.GoldData{}, nil
}

func (trd TestReserveData) GetExchangeStatus() (common.ExchangesStatus, error) {
	return common.ExchangesStatus{}, nil
}
func (trd TestReserveData) UpdateExchangeStatus(exchange string, status bool, timestamp uint64) error {
	return nil
}

func (trd TestReserveData) UpdateExchangeNotification(exchange, action, tokenPair string, from, to uint64, isWarning bool, msg string) error {
	return nil
}
func (trd TestReserveData) GetNotifications() (common.ExchangeNotifications, error) {
	return common.ExchangeNotifications{}, nil
}

func (trd TestReserveData) GetTradeHistory(fromTime, toTime uint64) (common.AllTradeHistory, error) {
	return common.AllTradeHistory{}, nil
}
func (trd TestReserveData) CheckAndModifyAuthDataAfterTokenUpdate() error {
	return nil
}

func (trd TestReserveData) Run() error {
	return nil
}
func (trd TestReserveData) RunStorageController() error {
	return nil
}
func (trd TestReserveData) Stop() error {
	return nil
}

func NewTestReserveData() *TestReserveData {
	return &TestReserveData{}
}
