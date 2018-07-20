package stat

import (
	"errors"
	"fmt"
	"strings"

	ethereum "github.com/ethereum/go-ethereum/common"

	"github.com/KyberNetwork/reserve-data/common"
)

type StatStorageTest struct {
	storage StatStorage
}

const (
	testWalletAddr string = "0xdd61803d4a56c597e0fC864f7a20Ec7158c6Cba5"
	testCountry    string = "unKnown"
	testAssetAddr  string = "0x2aab2b157a03915c8A73aDae735d0cf51c872f31"
	testUserAddr   string = "0x778599Dd7893C8166D313F0F9B5F6cbF7536c293"
)

func NewStatStorageTest(storage StatStorage) *StatStorageTest {
	return &StatStorageTest{storage}
}

func (self *StatStorageTest) TestTradeStatsSummary() error {
	var err error
	mtStat := common.NewMetricStats(
		10.0,
		4567.8,
		11.1,
		3,
		2,
		4,
		5,
		0,
		0,
	)
	tzmtStat := common.MetricStatsTimeZone{0: {0: mtStat}}
	updates := map[string]common.MetricStatsTimeZone{"trade_summary": tzmtStat}
	if err = self.storage.SetTradeSummary(updates, 0); err != nil {
		return err
	}
	tradeSum, err := self.storage.GetTradeSummary(0, 86400000, 0)
	if err != nil {
		return err
	}
	if tradeSum == nil || len(tradeSum) == 0 {
		return errors.New("Can't find such record")
	}
	result, ok := (tradeSum[0]).(common.MetricStats)
	if !ok {
		return errors.New("Type mismatched: get trade stat summary return wrong type")
	}
	usdVol := (result.USDVolume)
	if !ok {
		return errors.New("Type mismatched: get trade stat summary return missing field")
	}
	if usdVol != 4567.8 {
		return fmt.Errorf("Wrong usd volume value returned: %v expected 4567.8 ", usdVol)

	}
	return nil
}

func (self *StatStorageTest) TestWalletStats() error {
	var err error

	mtStat := common.NewMetricStats(
		10.0,
		4567.8,
		11.1,
		3,
		2,
		4,
		5,
		0,
		0,
	)
	testWallet := ethereum.HexToAddress(testAssetAddr)

	tzmtStat := common.MetricStatsTimeZone{0: {0: mtStat}}
	updates := map[string]common.MetricStatsTimeZone{testAssetAddr: tzmtStat}
	err = self.storage.SetWalletStat(updates, 0)
	if err != nil {
		return err
	}
	walletStat, err := self.storage.GetWalletStats(0, 86400000, testWallet, 0)
	if walletStat == nil || len(walletStat) == 0 {
		return errors.New("Can't find such record, addressess might not be in the correct case")
	}
	result, ok := (walletStat[0]).(common.MetricStats)
	if !ok {
		return errors.New("Type mismatched: get wallet stat return wrong type (UPPER CASE ADDR)")
	}
	usdVol := (result.USDVolume)
	if !ok {
		return errors.New("Type mismatched: get wallet stat return missing field (UPPER CASE ADDR)")
	}
	if usdVol != 4567.8 {
		return fmt.Errorf("Wrong usd volume value returned: %v expected 4567.8 (UPPER CASE ADDR)", usdVol)
	}
	return nil
}

func (self *StatStorageTest) TestCountryStats() error {
	var err error
	mtStat := common.NewMetricStats(
		10.0,
		4567.8,
		11.1,
		3,
		2,
		4,
		5,
		0,
		0,
	)
	tzmtStat := common.MetricStatsTimeZone{0: {0: mtStat}}
	updates := map[string]common.MetricStatsTimeZone{testCountry: tzmtStat}
	err = self.storage.SetCountryStat(updates, 0)
	if err != nil {
		return err
	}

	countryStat, err := self.storage.GetCountryStats(0, 86400000, testCountry, 0)
	if countryStat == nil || len(countryStat) == 0 {
		return errors.New("Can't find such record, addressess might not be in the correct case")
	}
	result, ok := (countryStat[0]).(common.MetricStats)
	if !ok {
		return errors.New("Type mismatched: get country stats return wrong type (LOWER CASE COUNTRY) ")
	}
	usdVol := (result.USDVolume)
	if !ok {
		return errors.New("Type mismatched: get country stats return missing field (LOWER CASE COUNTRY)")
	}
	if usdVol != 4567.8 {
		return fmt.Errorf("Wrong usd volume value returned: %v expected 4567.8 (LOWER CASE COUNTRY)", usdVol)

	}
	countryStat, err = self.storage.GetCountryStats(0, 86400000, strings.ToUpper(testCountry), 0)
	if countryStat == nil || len(countryStat) == 0 {
		return errors.New("Can't find such record, addressess might not be in the correct case")
	}
	result, ok = (countryStat[0]).(common.MetricStats)
	if !ok {
		return errors.New("Type mismatched: get country stats return wrong type (UPPER CASE COUNTRY) ")
	}
	usdVol = (result.USDVolume)
	if !ok {
		return errors.New("Type mismatched: get country stats return missing field (UPPER CASE COUNTRY)")
	}
	if usdVol != 4567.8 {
		return fmt.Errorf("Wrong usd volume value returned: %v expected 4567.8 (UPPER CASE COUNTRY)", usdVol)
	}
	return nil
}

func (self *StatStorageTest) TestVolumeStats() error {
	var err error
	vlStat := common.NewVolumeStats(
		10.0,
		4567.8,
		11.1,
	)
	testAsset := ethereum.HexToAddress(testAssetAddr)

	tzvlStat := common.VolumeStatsTimeZone{"D": {0: vlStat}}
	updates := map[string]common.VolumeStatsTimeZone{testAssetAddr: tzvlStat}
	err = self.storage.SetVolumeStat(updates, 0)
	if err != nil {
		return err
	}
	assetVol, err := self.storage.GetAssetVolume(0, 86400000, "D", testAsset)
	if assetVol == nil || len(assetVol) == 0 {
		return errors.New("Can't find such record, addressess might not be in the correct case")
	}
	result, ok := (assetVol[0]).(common.VolumeStats)
	if !ok {
		return errors.New("Type mismatched: get volume stat return wrong type (LOWER CASE ADDR)")
	}
	usdVol := (result.USDAmount)
	if !ok {
		return errors.New("Type mismatched: get volume stat return missing field (LOWER CASE ADDR)")
	}
	if usdVol != 4567.8 {
		return fmt.Errorf("Wrong usd volume value returned: %v expected 4567.8 (LOWER CASE ADDR)", usdVol)
	}

	assetVol, err = self.storage.GetAssetVolume(0, 86400000, "D", testAsset)
	if assetVol == nil || len(assetVol) == 0 {
		return errors.New("Can't find such record, addressess might not be in the correct case")
	}
	result, ok = (assetVol[0]).(common.VolumeStats)
	if !ok {
		return errors.New("Type mismatched: get volume stat return wrong type (UPPER CASE ADDR)")
	}
	usdVol = (result.USDAmount)
	if !ok {
		return errors.New("Type mismatched: get volume stat return missing field (UPPER CASE ADDR)")
	}
	if usdVol != 4567.8 {
		return fmt.Errorf("Wrong usd volume value returned: %v expected 4567.8 (UPPER CASE ADDR)", usdVol)
	}

	//test user volume
	testUser := ethereum.HexToAddress(testUserAddr)
	updates = map[string]common.VolumeStatsTimeZone{testUserAddr: tzvlStat}
	err = self.storage.SetVolumeStat(updates, 0)
	if err != nil {
		return err
	}
	userVol, err := self.storage.GetUserVolume(0, 86400000, "D", testUser)
	if (userVol == nil) || len(userVol) == 0 {
		return errors.New("Test uservolume failed. Can't find such record, addressess might not be in the correct case")
	}
	result, ok = (userVol[0]).(common.VolumeStats)
	if !ok {
		return errors.New("Type mismatched: get user volume summary return wrong type")
	}
	usdVol = (result.USDAmount)
	if !ok {
		return errors.New("Type mismatched: get user volume summary return missing field")
	}
	if usdVol != 4567.8 {
		return fmt.Errorf("Wrong usd volume value returned: %v expected 4567.8", usdVol)
	}

	updates = map[string]common.VolumeStatsTimeZone{fmt.Sprintf("%s_%s", testAssetAddr, testUserAddr): tzvlStat}
	err = self.storage.SetVolumeStat(updates, 0)
	if err != nil {
		return err
	}
	reserveVol, err := self.storage.GetReserveVolume(0, 86400000, "D", testAsset, testUser)
	if (reserveVol == nil) || len(reserveVol) == 0 {
		return errors.New("Can't find such record, addressess might not be in the correct case")
	}
	result, ok = (reserveVol[0]).(common.VolumeStats)
	if !ok {
		return errors.New("Type mismatched: get user volume summary return wrong type")
	}
	usdVol = (result.USDAmount)
	if !ok {
		return errors.New("Type mismatched: get user volume summary return missing field")
	}
	if usdVol != 4567.8 {
		return fmt.Errorf("Wrong usd volume value returned: %v expected 4567.8", usdVol)
	}
	return nil

}

func (self *StatStorageTest) TestBurnFee() error {
	var err error
	bfStat := common.NewBurnFeeStats(4567.8)

	tzbfStat := common.BurnFeeStatsTimeZone{"D": {0: bfStat}}
	updates := map[string]common.BurnFeeStatsTimeZone{testAssetAddr: tzbfStat}
	err = self.storage.SetBurnFeeStat(updates, 0)
	testAsset := ethereum.HexToAddress(testAssetAddr)
	if err != nil {
		return err
	}
	burnFee, err := self.storage.GetBurnFee(0, 86400000, "D", testAsset)
	if err != nil {
		return err
	}
	if (burnFee == nil) || len(burnFee) == 0 {
		return errors.New("Can't find such record, addressess might not be in the correct case")
	}
	//Note : This is only temporary, burn fee return needs to be casted to common.BurnFeeStats for consistent in design
	result, ok := (burnFee[0]).(float64)
	if !ok {
		return errors.New(" Type mismatched: get burn fee return wrong type")
	}
	burnVol := (result)
	if !ok {
		return errors.New("Type mismatched: get burn fee return missing field ")
	}
	if burnVol != 4567.8 {
		return fmt.Errorf("Wrong burn fee value returned: %v expected 4567.8 ", burnVol)
	}

	testWallet := ethereum.HexToAddress(testWalletAddr)
	updates = map[string]common.BurnFeeStatsTimeZone{fmt.Sprintf("%s_%s", testAssetAddr, testWalletAddr): tzbfStat}
	err = self.storage.SetBurnFeeStat(updates, 0)
	if err != nil {
		return err
	}
	burnFee, err = self.storage.GetWalletFee(0, 86400000, "D", testAsset, testWallet)
	if burnFee == nil || len(burnFee) == 0 {
		return errors.New("Can't find such record, addressess might not be in the correct case ")
	}
	result, ok = (burnFee[0]).(float64)
	if !ok {
		return errors.New("Type mismatched: get burn fee return wrong type ")
	}
	burnVol = (result)
	if !ok {
		return errors.New("Type mismatched: get burn fee return missing field ")
	}
	if burnVol != 4567.8 {
		return fmt.Errorf("Wrong wallet fee value returned: %v expected 4567.8", burnVol)

	}
	return nil
}

func (self *StatStorageTest) TestWalletAddress() error {
	var err error
	walletAddr := ethereum.HexToAddress("0xdd61803d4A56C597e0fc864f7a20ec7158c6cba5")
	err = self.storage.SetWalletAddress(walletAddr)
	if err != nil {
		return err
	}
	walletaddresses, err := self.storage.GetWalletAddresses()
	if err != nil {
		return err
	}
	if len(walletaddresses) != 1 {
		return fmt.Errorf("expected 1 record, got %d record of wallet addresses returned", len(walletaddresses))
	}
	if walletaddresses[0] != "0xdd61803d4a56c597e0fc864f7a20ec7158c6cba5" {
		return fmt.Errorf("expected address 0xdd61803d4a56c597e0fc864f7a20ec7158c6cba5, got %s instead", walletaddresses[0])
	}
	return err
}

func (self *StatStorageTest) TestLastProcessedTradeLogTimePoint() error {
	var err error
	err = self.storage.SetLastProcessedTradeLogTimepoint(TradeSummaryAggregation, 45678)
	if err != nil {
		return err
	}
	lastTimePoint, err := self.storage.GetLastProcessedTradeLogTimepoint(TradeSummaryAggregation)
	if err != nil {
		return err
	}
	if lastTimePoint != 45678 {
		return fmt.Errorf("expected last time point to be 45678, got %d instead", lastTimePoint)
	}
	return err
}

func (self *StatStorageTest) TestCountries() error {
	var err error
	err = self.storage.SetCountry("Bunny")
	if err != nil {
		return err
	}
	countries, err := self.storage.GetCountries()
	if err != nil {
		return err
	}
	if len(countries) != 1 {
		return fmt.Errorf("wrong countries len, expect 1, got %d", len(countries))
	}
	if countries[0] != "BUNNY" {
		return fmt.Errorf("wrong country result, expect BUNNY, got %s", countries[0])
	}
	return err

}

func (self *StatStorageTest) TestFirstTradeEver() error {
	var err error
	tradelog := common.TradeLog{
		Timestamp:   45678,
		UserAddress: ethereum.HexToAddress(testUserAddr),
	}
	userAddrs := []common.TradeLog{tradelog}
	err = self.storage.SetFirstTradeEver(&userAddrs)
	if err != nil {
		return err
	}

	tradelog = common.TradeLog{
		Timestamp:   45678,
		UserAddress: ethereum.HexToAddress(testWalletAddr),
	}
	userAddrs = []common.TradeLog{tradelog}
	err = self.storage.SetFirstTradeEver(&userAddrs)
	if err != nil {
		return err
	}
	allFirstTradeEver, err := self.storage.GetAllFirstTradeEver()
	if len(allFirstTradeEver) != 2 {
		return fmt.Errorf("wrong all first trade ever  len, expect 2, got %d", len(allFirstTradeEver))
	}
	return err

}

func (self *StatStorageTest) TestFirstTradeInDay() error {
	var err error
	tradelog := common.TradeLog{
		Timestamp:   45678,
		UserAddress: ethereum.HexToAddress(testUserAddr),
	}
	userAddrs := []common.TradeLog{tradelog}
	err = self.storage.SetFirstTradeInDay(&userAddrs)
	if err != nil {
		return err
	}
	testUserAddr := ethereum.HexToAddress(testUserAddr)
	timepoint, err := self.storage.GetFirstTradeInDay(testUserAddr, 0, 0)
	if err != nil {
		return err
	}
	if timepoint != 45678 {
		return fmt.Errorf("first trade in day error, expect timepoint 45678, got %d", timepoint)
	}
	return err
}
