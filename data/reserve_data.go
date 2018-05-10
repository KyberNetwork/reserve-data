package data

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/archive"
	"github.com/KyberNetwork/reserve-data/data/storagecontroller"
)

type ReserveData struct {
	storage           Storage
	fetcher           Fetcher
	storageController storagecontroller.StorageController
	globalStorage     GlobalStorage
}

func (self ReserveData) CurrentGoldInfoVersion(timepoint uint64) (common.Version, error) {
	return self.globalStorage.CurrentGoldInfoVersion(timepoint)
}

func (self ReserveData) GetGoldData(timestamp uint64) (common.GoldData, error) {
	version, err := self.CurrentGoldInfoVersion(timestamp)
	if err != nil {
		return common.GoldData{}, nil
	}
	return self.globalStorage.GetGoldInfo(version)
}

func (self ReserveData) CurrentPriceVersion(timepoint uint64) (common.Version, error) {
	return self.storage.CurrentPriceVersion(timepoint)
}

func (self ReserveData) GetAllPrices(timepoint uint64) (common.AllPriceResponse, error) {
	timestamp := common.GetTimestamp()
	version, err := self.storage.CurrentPriceVersion(timepoint)
	if err != nil {
		return common.AllPriceResponse{}, err
	} else {
		result := common.AllPriceResponse{}
		data, err := self.storage.GetAllPrices(version)
		returnTime := common.GetTimestamp()
		result.Version = version
		result.Timestamp = timestamp
		result.ReturnTime = returnTime
		result.Data = data.Data
		result.Block = data.Block
		return result, err
	}
}

func (self ReserveData) GetOnePrice(pairID common.TokenPairID, timepoint uint64) (common.OnePriceResponse, error) {
	timestamp := common.GetTimestamp()
	version, err := self.storage.CurrentPriceVersion(timepoint)
	if err != nil {
		return common.OnePriceResponse{}, err
	} else {
		result := common.OnePriceResponse{}
		data, err := self.storage.GetOnePrice(pairID, version)
		returnTime := common.GetTimestamp()
		result.Version = version
		result.Timestamp = timestamp
		result.ReturnTime = returnTime
		result.Data = data
		return result, err
	}
}

func (self ReserveData) CurrentAuthDataVersion(timepoint uint64) (common.Version, error) {
	return self.storage.CurrentAuthDataVersion(timepoint)
}

func (self ReserveData) GetAuthData(timepoint uint64) (common.AuthDataResponse, error) {
	timestamp := common.GetTimestamp()
	version, err := self.storage.CurrentAuthDataVersion(timepoint)
	if err != nil {
		return common.AuthDataResponse{}, err
	} else {
		result := common.AuthDataResponse{}
		data, err := self.storage.GetAuthData(version)
		returnTime := common.GetTimestamp()
		result.Version = version
		result.Timestamp = timestamp
		result.ReturnTime = returnTime
		result.Data.Valid = data.Valid
		result.Data.Error = data.Error
		result.Data.Timestamp = data.Timestamp
		result.Data.ReturnTime = data.ReturnTime
		result.Data.ExchangeBalances = data.ExchangeBalances
		result.Data.PendingActivities = data.PendingActivities
		result.Data.Block = data.Block
		result.Data.ReserveBalances = map[string]common.BalanceResponse{}
		for tokenID, balance := range data.ReserveBalances {
			result.Data.ReserveBalances[tokenID] = balance.ToBalanceResponse(
				common.MustGetInternalToken(tokenID).Decimal,
			)
		}
		return result, err
	}
}

func (self ReserveData) CurrentRateVersion(timepoint uint64) (common.Version, error) {
	return self.storage.CurrentRateVersion(timepoint)
}

func isDuplicated(oldData, newData map[string]common.RateResponse) bool {
	for tokenID, oldElem := range oldData {
		newelem, ok := newData[tokenID]
		if !ok {
			return false
		}
		if oldElem.BaseBuy != newelem.BaseBuy {
			return false
		}
		if oldElem.CompactBuy != newelem.CompactBuy {
			return false
		}
		if oldElem.BaseSell != newelem.BaseSell {
			return false
		}
		if oldElem.CompactSell != newelem.CompactSell {
			return false
		}
		if oldElem.Rate != newelem.Rate {
			return false
		}
	}
	return true
}

func getOneRateData(rate common.AllRateEntry) map[string]common.RateResponse {
	//get data from rate object and return the data.
	data := map[string]common.RateResponse{}
	for tokenID, r := range rate.Data {
		data[tokenID] = common.RateResponse{
			Valid:       rate.Valid,
			Error:       rate.Error,
			Timestamp:   rate.Timestamp,
			ReturnTime:  rate.ReturnTime,
			BaseBuy:     common.BigToFloat(r.BaseBuy, 18),
			CompactBuy:  r.CompactBuy,
			BaseSell:    common.BigToFloat(r.BaseSell, 18),
			CompactSell: r.CompactSell,
			Block:       r.Block,
		}
	}
	return data
}

func (self ReserveData) GetRates(fromTime, toTime uint64) ([]common.AllRateResponse, error) {
	result := []common.AllRateResponse{}
	rates, err := self.storage.GetRates(fromTime, toTime)
	if err != nil {
		return result, err
	}
	//current: the unchanged one so far
	current := common.AllRateResponse{}
	for _, rate := range rates {
		one := common.AllRateResponse{}
		one.Timestamp = rate.Timestamp
		one.ReturnTime = rate.ReturnTime
		one.Error = rate.Error
		one.Valid = rate.Valid
		one.Data = getOneRateData(rate)
		one.BlockNumber = rate.BlockNumber
		//if one is the same as current
		if isDuplicated(one.Data, current.Data) {
			if len(result) > 0 {
				result[len(result)-1].ToBlockNumber = one.BlockNumber
			} else {
				one.ToBlockNumber = one.BlockNumber
			}
		} else {
			one.ToBlockNumber = rate.BlockNumber
			result = append(result, one)
			current = one
		}
	}

	return result, nil
}
func (self ReserveData) GetRate(timepoint uint64) (common.AllRateResponse, error) {
	timestamp := common.GetTimestamp()
	version, err := self.storage.CurrentRateVersion(timepoint)
	if err != nil {
		return common.AllRateResponse{}, err
	} else {
		result := common.AllRateResponse{}
		rates, err := self.storage.GetRate(version)
		returnTime := common.GetTimestamp()
		result.Version = version
		result.Timestamp = timestamp
		result.ReturnTime = returnTime
		data := map[string]common.RateResponse{}
		for tokenID, rate := range rates.Data {
			data[tokenID] = common.RateResponse{
				Valid:       rates.Valid,
				Error:       rates.Error,
				Timestamp:   rates.Timestamp,
				ReturnTime:  rates.ReturnTime,
				BaseBuy:     common.BigToFloat(rate.BaseBuy, 18),
				CompactBuy:  rate.CompactBuy,
				BaseSell:    common.BigToFloat(rate.BaseSell, 18),
				CompactSell: rate.CompactSell,
				Block:       rate.Block,
			}
		}
		result.Data = data
		return result, err
	}
}

func (self ReserveData) GetExchangeStatus() (common.ExchangesStatus, error) {
	data, err := self.storage.GetExchangeStatus()
	return data, err
}

func (self ReserveData) UpdateExchangeStatus(exchange string, status bool, timestamp uint64) error {
	currentExchangeStatus, err := self.storage.GetExchangeStatus()
	if err != nil {
		return err
	}
	currentExchangeStatus[exchange] = common.ExStatus{
		Timestamp: timestamp,
		Status:    status,
	}
	return self.storage.UpdateExchangeStatus(currentExchangeStatus)
}

func (self ReserveData) UpdateExchangeNotification(
	exchange, action, tokenPair string, fromTime, toTime uint64, isWarning bool, msg string) error {
	err := self.storage.UpdateExchangeNotification(exchange, action, tokenPair, fromTime, toTime, isWarning, msg)
	return err
}

func (self ReserveData) GetRecords(fromTime, toTime uint64) ([]common.ActivityRecord, error) {
	return self.storage.GetAllRecords(fromTime, toTime)
}

func (self ReserveData) GetPendingActivities() ([]common.ActivityRecord, error) {
	return self.storage.GetPendingActivities()
}

func (self ReserveData) GetTradeHistory(timepoint uint64) (common.AllTradeHistory, error) {
	data, err := self.storage.GetTradeHistory(timepoint)
	return data, err
}

func (self ReserveData) GetNotifications() (common.ExchangeNotifications, error) {
	return self.storage.GetExchangeNotifications()
}

func (self ReserveData) Run() error {
	return self.fetcher.Run()
}

func (self ReserveData) Stop() error {
	return self.fetcher.Stop()
}

func (self ReserveData) ControlAuthDataSize() error {
	for {
		log.Printf("StorageController: waiting for signal from runner AuthData controller channel")
		t := <-self.storageController.Runner.GetAuthBucketTicker()
		timepoint := common.TimeToTimepoint(t)
		log.Printf("StorageController: got signal in AuthData controller channel with timestamp %d", common.TimeToTimepoint(t))
		fileName := fmt.Sprintf("ExpiredAuthData_at_%s", time.Unix(int64(timepoint/1000), 0).UTC())
		nRecord, err := self.storage.ExportExpiredAuthData(common.TimeToTimepoint(t), fileName)
		if err != nil {
			log.Printf("ERROR: StorageController export and prune AuthData operation failed: %s, err")
		} else {
			if nRecord > 0 {
				err = self.storageController.Arch.BackupFile(self.storageController.Arch.GetReserveDataBucketName(), self.storageController.Arch.GetAuthDataPath(), fileName)
				if err != nil {
					log.Printf("StorageController: Back up file failed: %s", err)
				}
			}
			if err == nil {
				os.Remove(fileName)
				log.Printf("StorageController: exported and pruned %d expired records from AuthData", nRecord)
			}
		}
	}
}

func (self ReserveData) RunStorageController() error {
	self.storageController.Runner.Start()
	go self.ControlAuthDataSize()
	return nil
}

func NewReserveData(storage Storage, fetcher Fetcher, storageControllerRunner storagecontroller.StorageControllerRunner, arch archive.Archive, globalStorage GlobalStorage) *ReserveData {
	storageController, err := storagecontroller.NewStorageController(storageControllerRunner, arch)
	if err != nil {
		panic(err)
	}
	return &ReserveData{storage, fetcher, storageController, globalStorage}
}
