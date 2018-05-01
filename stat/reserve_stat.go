package stat

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/KyberNetwork/reserve-data/common"
	ethereum "github.com/ethereum/go-ethereum/common"
)

const (
	MAX_GET_RATES_PERIOD uint64 = 86400000 //7 days in milisec
)

type ReserveStats struct {
	analyticStorage   AnalyticStorage
	statStorage       StatStorage
	logStorage        LogStorage
	userStorage       UserStorage
	rateStorage       RateStorage
	feeSetRateStorage FeeSetRateStorage
	fetcher           *Fetcher
	controllerRunner  ControllerRunner
}

func NewReserveStats(
	analyticStorage AnalyticStorage,
	statStorage StatStorage,
	logStorage LogStorage,
	rateStorage RateStorage,
	userStorage UserStorage,
	feeSetRateStorage FeeSetRateStorage,
	controllerRunner ControllerRunner,
	fetcher *Fetcher) *ReserveStats {
	return &ReserveStats{
		analyticStorage:   analyticStorage,
		statStorage:       statStorage,
		logStorage:        logStorage,
		rateStorage:       rateStorage,
		userStorage:       userStorage,
		feeSetRateStorage: feeSetRateStorage,
		fetcher:           fetcher,
		controllerRunner:  controllerRunner,
	}
}

func validateTimeWindow(fromTime, toTime uint64, freq string) (uint64, uint64, error) {
	var from = fromTime * 1000000
	var to = toTime * 1000000
	switch freq {
	case "m", "M":
		if to-from > uint64((time.Hour * 24).Nanoseconds()) {
			return 0, 0, errors.New("Minute frequency limit is 1 day")
		}
	case "h", "H":
		if to-from > uint64((time.Hour * 24 * 180).Nanoseconds()) {
			return 0, 0, errors.New("Hour frequency limit is 180 days")
		}
	case "d", "D":
		if to-from > uint64((time.Hour * 24 * 365 * 3).Nanoseconds()) {
			return 0, 0, errors.New("Day frequency limit is 3 years")
		}
	default:
		return 0, 0, errors.New("Invalid frequencies")
	}
	return from, to, nil
}

func (self ReserveStats) GetAssetVolume(fromTime, toTime uint64, freq, asset string) (common.StatTicks, error) {
	data := common.StatTicks{}

	fromTime, toTime, err := validateTimeWindow(fromTime, toTime, freq)
	if err != nil {
		return data, err
	}

	token, err := common.GetNetworkToken(asset)
	if err != nil {
		return data, errors.New(fmt.Sprintf("assets %s is not supported", asset))
	}

	data, err = self.statStorage.GetAssetVolume(fromTime, toTime, freq, ethereum.HexToAddress(token.Address))
	return data, err
}

func (self ReserveStats) GetBurnFee(fromTime, toTime uint64, freq, reserveAddr string) (common.StatTicks, error) {
	data := common.StatTicks{}

	fromTime, toTime, err := validateTimeWindow(fromTime, toTime, freq)
	if err != nil {
		return data, err
	}

	data, err = self.statStorage.GetBurnFee(fromTime, toTime, freq, ethereum.HexToAddress(reserveAddr))

	return data, err
}

func (self ReserveStats) GetWalletFee(fromTime, toTime uint64, freq, reserveAddr, walletAddr string) (common.StatTicks, error) {
	data := common.StatTicks{}

	fromTime, toTime, err := validateTimeWindow(fromTime, toTime, freq)
	if err != nil {
		return data, err
	}

	data, err = self.statStorage.GetWalletFee(fromTime, toTime, freq, ethereum.HexToAddress(reserveAddr), ethereum.HexToAddress(walletAddr))

	return data, err
}

func (self ReserveStats) GetUserVolume(fromTime, toTime uint64, freq, userAddr string) (common.StatTicks, error) {
	data := common.StatTicks{}

	fromTime, toTime, err := validateTimeWindow(fromTime, toTime, freq)
	if err != nil {
		return data, err
	}

	data, err = self.statStorage.GetUserVolume(fromTime, toTime, freq, ethereum.HexToAddress(userAddr))

	return data, err
}

func (self ReserveStats) GetUsersVolume(fromTime, toTime uint64, freq string, userAddrs []string) (common.UsersVolume, error) {
	data := common.StatTicks{}
	result := common.UsersVolume{}

	fromTime, toTime, err := validateTimeWindow(fromTime, toTime, freq)
	if err != nil {
		return result, err
	}
	for _, userAddr := range userAddrs {
		data, err = self.statStorage.GetUserVolume(fromTime, toTime, freq, ethereum.HexToAddress(userAddr))
		result[userAddr] = data
	}

	return result, err
}

func (self ReserveStats) GetReserveVolume(fromTime, toTime uint64, freq, reserveAddr, tokenAddr string) (common.StatTicks, error) {
	data := common.StatTicks{}

	fromTime, toTime, err := validateTimeWindow(fromTime, toTime, freq)
	if err != nil {
		return data, err
	}

	reserveAddr = strings.ToLower(reserveAddr)
	tokenAddr = strings.ToLower(tokenAddr)
	data, err = self.statStorage.GetReserveVolume(fromTime, toTime, freq, ethereum.HexToAddress(reserveAddr), ethereum.HexToAddress(tokenAddr))
	return data, err
}

func (self ReserveStats) GetTradeSummary(fromTime, toTime uint64, timezone int64) (common.StatTicks, error) {
	data := common.StatTicks{}

	fromTime, toTime, err := validateTimeWindow(fromTime, toTime, "D")
	if err != nil {
		return data, err
	}

	data, err = self.statStorage.GetTradeSummary(fromTime, toTime, timezone)
	return data, err
}

func (self ReserveStats) GetTradeLogs(fromTime uint64, toTime uint64) ([]common.TradeLog, error) {
	result := []common.TradeLog{}

	if toTime-fromTime > MAX_GET_RATES_PERIOD {
		return result, errors.New(fmt.Sprintf("Time range is too broad, it must be smaller or equal to %d miliseconds", MAX_GET_RATES_PERIOD))
	}

	result, err := self.logStorage.GetTradeLogs(fromTime*1000000, toTime*1000000)
	return result, err
}

func (self ReserveStats) GetGeoData(fromTime, toTime uint64, country string, tzparam int64) (common.StatTicks, error) {
	var err error
	result := common.StatTicks{}
	fromTime, toTime, err = validateTimeWindow(fromTime, toTime, "D")
	if err != nil {
		return result, err
	}
	result, err = self.statStorage.GetCountryStats(fromTime, toTime, country, tzparam)
	return result, err
}

func (self ReserveStats) GetHeatMap(fromTime, toTime uint64, tzparam int64) (common.HeatmapResponse, error) {
	result := common.Heatmap{}
	var arrResult common.HeatmapResponse
	var err error
	fromTime, toTime, err = validateTimeWindow(fromTime, toTime, "D")
	if err != nil {
		return arrResult, err
	}
	countries, err := self.statStorage.GetCountries()
	if err != nil {
		return arrResult, err
	}

	// get stats
	for _, c := range countries {
		cStats, err := self.statStorage.GetCountryStats(fromTime, toTime, c, tzparam)
		if err != nil {
			return arrResult, err
		}
		for _, stat := range cStats {
			s := stat.(common.MetricStats)
			current := result[c]
			result[c] = common.HeatmapType{
				TotalETHValue:        current.TotalETHValue + s.ETHVolume,
				TotalFiatValue:       current.TotalFiatValue + s.USDVolume,
				ToTalBurnFee:         current.ToTalBurnFee + s.BurnFee,
				TotalTrade:           current.TotalTrade + s.TradeCount,
				TotalUniqueAddresses: current.TotalUniqueAddresses + s.UniqueAddr,
				TotalKYCUser:         current.TotalKYCUser + s.KYCEd,
			}
		}
	}

	// sort heatmap
	for k, v := range result {
		arrResult = append(arrResult, common.HeatmapObject{
			Country:              k,
			TotalETHValue:        v.TotalETHValue,
			TotalFiatValue:       v.TotalFiatValue,
			ToTalBurnFee:         v.ToTalBurnFee,
			TotalTrade:           v.TotalTrade,
			TotalUniqueAddresses: v.TotalUniqueAddresses,
			TotalKYCUser:         v.TotalKYCUser,
		})
	}
	sort.Sort(sort.Reverse(arrResult))
	return arrResult, err
}

func (self ReserveStats) GetTokenHeatmap(fromTime, toTime uint64, tokenStr, freq string) (common.TokenHeatmapResponse, error) {
	result := common.CountryTokenHeatmap{}
	var arrResult common.TokenHeatmapResponse
	fromTime, toTime, err := validateTimeWindow(fromTime, toTime, "D")
	if err != nil {
		return arrResult, err
	}
	countries, err := self.statStorage.GetCountries()
	if err != nil {
		return arrResult, err
	}
	token, err := common.GetNetworkToken(tokenStr)
	if err != nil {
		return arrResult, err
	}
	for _, country := range countries {
		key := fmt.Sprintf("%s_%s", country, strings.ToLower(token.Address))
		stats, err := self.statStorage.GetTokenHeatmap(fromTime, toTime, key, freq)
		if err != nil {
			return arrResult, err
		}
		for _, stat := range stats {
			s := stat.(common.VolumeStats)
			current := result[country]
			result[country] = common.VolumeStats{
				Volume:    current.Volume + s.Volume,
				ETHVolume: current.ETHVolume + s.ETHVolume,
				USDAmount: current.USDAmount + s.USDAmount,
			}
		}
	}
	for k, v := range result {
		arrResult = append(arrResult, common.TokenHeatmap{
			Country:   k,
			Volume:    v.Volume,
			ETHVolume: v.ETHVolume,
			USDVolume: v.USDAmount,
		})
	}
	sort.Sort(sort.Reverse(arrResult))
	return arrResult, err
}

func (self ReserveStats) GetCountries() ([]string, error) {
	result, _ := self.statStorage.GetCountries()
	return result, nil
}

func (self ReserveStats) GetCatLogs(fromTime uint64, toTime uint64) ([]common.SetCatLog, error) {
	return self.logStorage.GetCatLogs(fromTime, toTime)
}

func (self ReserveStats) GetPendingAddresses() ([]string, error) {
	addresses, err := self.userStorage.GetPendingAddresses()
	if err != nil {
		return nil, err
	}
	result := []string{}
	for _, addr := range addresses {
		result = append(result, common.AddrToString(addr))
	}
	return result, nil
}

func (self ReserveStats) RunAnalyticStorageController() {
	for {
		log.Printf("waiting for signal from analytic storage control channel")
		t := <-self.controllerRunner.GetAnalyticStorageControlTicker()
		timepoint := common.TimeToTimepoint(t)
		log.Printf("got signal in analytic storage control channel with timestamp %d", timepoint)
		fileName := fmt.Sprintf("ExpiredPriceAnalyticData_%s", time.Unix(int64(timepoint/1000), 0).UTC())
		nRecord, err := self.analyticStorage.ExportPruneExpired(common.GetTimepoint(), fileName)
		if err != nil {
			log.Printf("export and prune operation failed: %s", err)
		} else {
			if nRecord > 0 {
				err := self.analyticStorage.BackupFile(fileName)
				if err != nil {
					log.Printf("AnalyticPriceData: Back up file failed: %s", err)
				} else {
					log.Printf("AnalyticPriceData: Back up file successfully.")
				}

			} else {
				//remove the empty file
				os.Remove(fileName)
			}
			log.Printf("AnalyticPriceData: exported and pruned %d expired records from storage controll block from blockchain", nRecord)
		}
	}
}

func (self ReserveStats) RunDBController() error {
	err := self.controllerRunner.Start()
	if err != nil {
		return err
	}
	go self.RunAnalyticStorageController()
	return err
}

func (self ReserveStats) Run() error {
	return self.fetcher.Run()
}

func (self ReserveStats) Stop() error {
	return self.fetcher.Stop()
}

func (self ReserveStats) GetCapByAddress(addr ethereum.Address) (*common.UserCap, error) {
	category, err := self.userStorage.GetCategory(addr)
	if err != nil {
		return nil, err
	}
	if category == "0x4" {
		return common.KycedCap(), nil
	} else {
		return common.NonKycedCap(), nil
	}
}

func (self ReserveStats) GetCapByUser(userID string) (*common.UserCap, error) {
	addresses, _, err := self.userStorage.GetAddressesOfUser(userID)
	if err != nil {
		return nil, err
	}
	if len(addresses) == 0 {
		log.Printf("Couldn't find any associated addresses. User %s is not kyced.", userID)
		return common.NonKycedCap(), nil
	} else {
		return self.GetCapByAddress(addresses[0])
	}
}

func isDuplicate(currentRate, latestRate common.ReserveRates) bool {
	currentData := currentRate.Data
	latestData := latestRate.Data
	for key, _ := range currentData {
		if currentData[key].BuyReserveRate != latestData[key].BuyReserveRate ||
			currentData[key].BuySanityRate != latestData[key].BuySanityRate ||
			currentData[key].SellReserveRate != latestData[key].SellReserveRate ||
			currentData[key].SellSanityRate != latestData[key].SellSanityRate {
			return false
		}
	}
	return true
}
func (self ReserveStats) GetWalletStats(fromTime uint64, toTime uint64, walletAddr string, timezone int64) (common.StatTicks, error) {
	fromTime, toTime, err := validateTimeWindow(fromTime, toTime, "D")
	if err != nil {
		return nil, err
	}
	walletAddr = strings.ToLower(walletAddr)
	return self.statStorage.GetWalletStats(fromTime, toTime, ethereum.HexToAddress(walletAddr), timezone)
}

func (self ReserveStats) GetWalletAddress() ([]string, error) {
	return self.statStorage.GetWalletAddress()
}

func (self ReserveStats) GetReserveRates(fromTime, toTime uint64, reserveAddr ethereum.Address) ([]common.ReserveRates, error) {
	var result []common.ReserveRates
	var err error
	var rates []common.ReserveRates
	rates, err = self.rateStorage.GetReserveRates(fromTime, toTime, reserveAddr)
	latest := common.ReserveRates{}
	for _, rate := range rates {
		if !isDuplicate(rate, latest) {
			result = append(result, rate)
		} else {
			if len(result) > 0 {
				result[len(result)-1].ToBlockNumber = rate.BlockNumber
			}
		}
		latest = rate
	}
	log.Printf("Get reserve rate: %v", result)
	return result, err
}

func (self ReserveStats) GetUserList(fromTime, toTime uint64, timezone int64) (common.UserListResponse, error) {
	fromTime, toTime, err := validateTimeWindow(fromTime, toTime, "D")
	if err != nil {
		return []common.UserInfo{}, err
	}
	result := common.UserListResponse{}
	data, err := self.statStorage.GetUserList(fromTime, toTime, timezone)
	for _, v := range data {
		result = append(result, v)
	}
	sort.Sort(sort.Reverse(result))
	return result, err
}

func (self ReserveStats) UpdateUserAddresses(userID string, addrs []ethereum.Address, timestamps []uint64) error {
	addresses := []ethereum.Address{}
	for _, addr := range addrs {
		addresses = append(addresses, addr)
	}
	return self.userStorage.UpdateUserAddresses(userID, addresses, timestamps)
}

func (self ReserveStats) ExceedDailyLimit(address ethereum.Address) (bool, error) {
	user, _, err := self.userStorage.GetUserOfAddress(address)
	log.Printf("got user %s for address %s", user, strings.ToLower(address.Hex()))
	if err != nil {
		return false, err
	}
	addrs := []string{}
	if user == "" {
		// address is not associated to any users
		addrs = append(addrs, strings.ToLower(address.Hex()))
	} else {
		addrs, _, err := self.userStorage.GetAddressesOfUser(user)
		log.Printf("got addresses %v for address %s", addrs, strings.ToLower(address.Hex()))
		if err != nil {
			return false, err
		}
	}
	today := common.GetTimepoint() / uint64(24*time.Hour/time.Millisecond) * uint64(24*time.Hour/time.Millisecond)
	var totalVolume float64 = 0.0
	for _, addr := range addrs {
		volumeStats, err := self.GetUserVolume(today-1, today, "D", addr)
		if err == nil {
			log.Printf("volumes: %+v", volumeStats)
			if len(volumeStats) == 0 {
			} else if len(volumeStats) > 1 {
				log.Printf("Got more than 1 day stats. This is a bug in GetUserVolume")
			} else {
				for _, volume := range volumeStats {
					volumeValue := volume.(common.VolumeStats)
					totalVolume += volumeValue.USDAmount
					break
				}
			}
		} else {
			log.Printf("Getting volumes for %s failed, err: %s", strings.ToLower(address.Hex()), err.Error())
		}
	}
	cap, err := self.GetCapByAddress(address)
	if err == nil && totalVolume >= cap.DailyLimit {
		return true, nil
	} else {
		return false, nil
	}
}

func (self ReserveStats) UpdatePriceAnalyticData(timestamp uint64, value []byte) error {
	return self.analyticStorage.UpdatePriceAnalyticData(timestamp, value)
}

func (self ReserveStats) GetPriceAnalyticData(fromTime uint64, toTime uint64) ([]common.AnalyticPriceResponse, error) {
	return self.analyticStorage.GetPriceAnalyticData(fromTime, toTime)
}

func (self ReserveStats) GetFeeSetRateByDay(fromTime uint64, toTime uint64) ([]common.FeeSetRate, error) {
	return self.feeSetRateStorage.GetFeeSetRateByDay(fromTime, toTime)
}
