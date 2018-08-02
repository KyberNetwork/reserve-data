package data

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/archive"
	"github.com/KyberNetwork/reserve-data/data/datapruner"
	"github.com/KyberNetwork/reserve-data/settings"
)

//ReserveData struct for reserve data
type ReserveData struct {
	storage           Storage
	fetcher           Fetcher
	storageController datapruner.StorageController
	globalStorage     GlobalStorage
	exchanges         []common.Exchange
	setting           Setting
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

func (self ReserveData) cointainAllInternalTokens(data common.AuthDataSnapshot) (bool, error) {
	tokens, err := self.setting.GetInternalTokens()
	if err != nil {
		return false, err
	}
	//go through all exchange in exchangebalances, if an exchange supports a token but its
	//available balance doesn't has that token, return false.
	for exID, ebalance := range data.ExchangeBalances {
		exName, ok := settings.ExchangeTypeValues()[string(exID)]
		if !ok {
			return false, fmt.Errorf("Exchange %s can not be found in current setting database", exID)
		}
		depositAddrs, vErr := self.setting.GetDepositAddresses(exName)
		if vErr != nil {
			return false, vErr
		}
		for _, token := range tokens {
			_, exchangeHasToken := depositAddrs[token.ID]
			_, balanceHasToken := ebalance.AvailableBalance[token.ID]
			if exchangeHasToken && !balanceHasToken {
				return false, nil
			}
		}
	}

	//check if all the internal token is avail in reserve balance
	for _, token := range tokens {
		if _, balanceHasToken := data.ReserveBalances[token.ID]; !balanceHasToken {

			return false, nil
		}
	}
	return true, nil
}

func (self ReserveData) CheckAndModifyAuthDataAfterTokenUpdate() error {
	timepoint := common.GetTimepoint()
	version, err := self.storage.CurrentAuthDataVersion(timepoint)
	data, err := self.storage.GetAuthData(version)
	if err != nil {
		return err
	}
	check, err := self.cointainAllInternalTokens(data)
	if err != nil {
		return err
	}
	if !check {
		data.Valid = false
		return self.storage.StoreAuthSnapshot(&data, timepoint)
	}
	return nil
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
			token, uErr := self.setting.GetInternalTokenByID(tokenID)
			//If the token is invalid, this must Panic
			if uErr != nil {
				return result, fmt.Errorf("Can't get Internal token %s: (%s)", tokenID, uErr)
			}
			result.Data.ReserveBalances[tokenID] = balance.ToBalanceResponse(
				token.Decimals,
			)
		}
		return result, err
	}
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
		one.Data = getOneRateData(rate)
		one.BlockNumber = rate.BlockNumber
		//if one is the same as current
		if isDuplicated(one.Data, current.Data) {
			if len(result) > 0 {
				result[len(result)-1].ToBlockNumber = one.BlockNumber
				result[len(result)-1].Timestamp = one.Timestamp
				result[len(result)-1].ReturnTime = one.ReturnTime
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
	return self.setting.GetExchangeStatus()
}

func (self ReserveData) UpdateExchangeStatus(exchange string, status bool, timestamp uint64) error {
	currentExchangeStatus, err := self.setting.GetExchangeStatus()
	if err != nil {
		return err
	}
	currentExchangeStatus[exchange] = common.ExStatus{
		Timestamp: timestamp,
		Status:    status,
	}
	return self.setting.UpdateExchangeStatus(currentExchangeStatus)
}

func (self ReserveData) UpdateExchangeNotification(
	exchange, action, tokenPair string, fromTime, toTime uint64, isWarning bool, msg string) error {
	return self.setting.UpdateExchangeNotification(exchange, action, tokenPair, fromTime, toTime, isWarning, msg)
}

func (self ReserveData) GetRecords(fromTime, toTime uint64) ([]common.ActivityRecord, error) {
	return self.storage.GetAllRecords(fromTime, toTime)
}

func (self ReserveData) GetPendingActivities() ([]common.ActivityRecord, error) {
	return self.storage.GetPendingActivities()
}

func (self ReserveData) GetNotifications() (common.ExchangeNotifications, error) {
	return self.setting.GetExchangeNotifications()
}

//Run run fetcher
func (self ReserveData) Run() error {
	return self.fetcher.Run()
}

//Stop stop the fetcher
func (self ReserveData) Stop() error {
	return self.fetcher.Stop()
}

//ControlAuthDataSize pack old data to file, push to S3 and prune outdated data
func (self ReserveData) ControlAuthDataSize() error {
	for {
		log.Printf("DataPruner: waiting for signal from runner AuthData controller channel")
		t := <-self.storageController.Runner.GetAuthBucketTicker()
		timepoint := common.TimeToTimepoint(t)
		log.Printf("DataPruner: got signal in AuthData controller channel with timestamp %d", common.TimeToTimepoint(t))
		fileName := fmt.Sprintf("./exported/ExpiredAuthData_at_%s", time.Unix(int64(timepoint/1000), 0).UTC())
		nRecord, err := self.storage.ExportExpiredAuthData(common.TimeToTimepoint(t), fileName)
		if err != nil {
			log.Printf("ERROR: DataPruner export AuthData operation failed: %s", err)
		} else {
			var integrity bool
			if nRecord > 0 {
				err = self.storageController.Arch.UploadFile(self.storageController.Arch.GetReserveDataBucketName(), self.storageController.ExpiredAuthDataPath, fileName)
				if err != nil {
					log.Printf("DataPruner: Upload file failed: %s", err)
				} else {
					integrity, err = self.storageController.Arch.CheckFileIntergrity(self.storageController.Arch.GetReserveDataBucketName(), self.storageController.ExpiredAuthDataPath, fileName)
					if err != nil {
						log.Printf("ERROR: DataPruner: error in file integrity check (%s):", err)
					} else if !integrity {
						log.Printf("ERROR: DataPruner: file upload corrupted")

					}
					if err != nil || !integrity {
						//if the intergrity check failed, remove the remote file.
						removalErr := self.storageController.Arch.RemoveFile(self.storageController.Arch.GetReserveDataBucketName(), self.storageController.ExpiredAuthDataPath, fileName)
						if removalErr != nil {
							log.Printf("ERROR: DataPruner: cannot remove remote file :(%s)", removalErr)
							return err
						}
					}
				}
			}
			if integrity && err == nil {
				nPrunedRecords, err := self.storage.PruneExpiredAuthData(common.TimeToTimepoint(t))
				if err != nil {
					log.Printf("DataPruner: Can not prune Auth Data (%s)", err)
					return err
				} else if nPrunedRecords != nRecord {
					log.Printf("DataPruner: Number of Exported Data is %d, which is different from number of pruned data %d", nRecord, nPrunedRecords)
				} else {
					log.Printf("DataPruner: exported and pruned %d expired records from AuthData", nRecord)
				}
			}
		}
		if err := os.Remove(fileName); err != nil {
			return err
		}
	}
}

func (self ReserveData) GetTradeHistory(fromTime, toTime uint64) (common.AllTradeHistory, error) {
	data := common.AllTradeHistory{}
	data.Data = map[common.ExchangeID]common.ExchangeTradeHistory{}
	for _, ex := range self.exchanges {
		history, err := ex.GetTradeHistory(fromTime, toTime)
		if err != nil {
			return data, err
		}
		data.Data[ex.ID()] = history
	}
	data.Timestamp = common.GetTimestamp()
	return data, nil
}

func (self ReserveData) RunStorageController() error {
	if err := self.storageController.Runner.Start(); err != nil {
		log.Fatalf("Storage controller runner error: %s", err.Error())
	}
	go func() {
		if err := self.ControlAuthDataSize(); err != nil {
			log.Printf("Control auth data size error: %s", err.Error())
		}
	}()
	return nil
}

//NewReserveData initiate a new reserve instance
func NewReserveData(storage Storage,
	fetcher Fetcher, storageControllerRunner datapruner.StorageControllerRunner,
	arch archive.Archive, globalStorage GlobalStorage,
	exchanges []common.Exchange, setting Setting) *ReserveData {
	storageController, err := datapruner.NewStorageController(storageControllerRunner, arch)
	if err != nil {
		panic(err)
	}
	return &ReserveData{storage, fetcher, storageController, globalStorage, exchanges, setting}
}
