package data

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/archive"
	"github.com/KyberNetwork/reserve-data/data/datapruner"
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

func (rd ReserveData) CurrentGoldInfoVersion(timepoint uint64) (common.Version, error) {
	return rd.globalStorage.CurrentGoldInfoVersion(timepoint)
}

func (rd ReserveData) GetGoldData(timestamp uint64) (common.GoldData, error) {
	version, err := rd.CurrentGoldInfoVersion(timestamp)
	if err != nil {
		return common.GoldData{}, nil
	}
	return rd.globalStorage.GetGoldInfo(version)
}

func (rd ReserveData) CurrentPriceVersion(timepoint uint64) (common.Version, error) {
	return rd.storage.CurrentPriceVersion(timepoint)
}

func (rd ReserveData) GetAllPrices(timepoint uint64) (common.AllPriceResponse, error) {
	timestamp := common.GetTimestamp()
	version, err := rd.storage.CurrentPriceVersion(timepoint)
	if err != nil {
		return common.AllPriceResponse{}, err
	} else {
		result := common.AllPriceResponse{}
		data, err := rd.storage.GetAllPrices(version)
		returnTime := common.GetTimestamp()
		result.Version = version
		result.Timestamp = timestamp
		result.ReturnTime = returnTime
		result.Data = data.Data
		result.Block = data.Block
		return result, err
	}
}

func (rd ReserveData) GetOnePrice(pairID common.TokenPairID, timepoint uint64) (common.OnePriceResponse, error) {
	timestamp := common.GetTimestamp()
	version, err := rd.storage.CurrentPriceVersion(timepoint)
	if err != nil {
		return common.OnePriceResponse{}, err
	} else {
		result := common.OnePriceResponse{}
		data, err := rd.storage.GetOnePrice(pairID, version)
		returnTime := common.GetTimestamp()
		result.Version = version
		result.Timestamp = timestamp
		result.ReturnTime = returnTime
		result.Data = data
		return result, err
	}
}

func (rd ReserveData) CurrentAuthDataVersion(timepoint uint64) (common.Version, error) {
	return rd.storage.CurrentAuthDataVersion(timepoint)
}

func (rd ReserveData) GetAuthData(timepoint uint64) (common.AuthDataResponse, error) {
	timestamp := common.GetTimestamp()
	version, err := rd.storage.CurrentAuthDataVersion(timepoint)
	if err != nil {
		return common.AuthDataResponse{}, err
	} else {
		result := common.AuthDataResponse{}
		data, err := rd.storage.GetAuthData(version)
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
			token, uErr := rd.setting.GetInternalTokenByID(tokenID)
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

func (rd ReserveData) GetRates(fromTime, toTime uint64) ([]common.AllRateResponse, error) {
	result := []common.AllRateResponse{}
	rates, err := rd.storage.GetRates(fromTime, toTime)
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
func (rd ReserveData) GetRate(timepoint uint64) (common.AllRateResponse, error) {
	timestamp := common.GetTimestamp()
	version, err := rd.storage.CurrentRateVersion(timepoint)
	if err != nil {
		return common.AllRateResponse{}, err
	} else {
		result := common.AllRateResponse{}
		rates, err := rd.storage.GetRate(version)
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

func (rd ReserveData) GetExchangeStatus() (common.ExchangesStatus, error) {
	return rd.setting.GetExchangeStatus()
}

func (rd ReserveData) UpdateExchangeStatus(exchange string, status bool, timestamp uint64) error {
	currentExchangeStatus, err := rd.setting.GetExchangeStatus()
	if err != nil {
		return err
	}
	currentExchangeStatus[exchange] = common.ExStatus{
		Timestamp: timestamp,
		Status:    status,
	}
	return rd.setting.UpdateExchangeStatus(currentExchangeStatus)
}

func (rd ReserveData) UpdateExchangeNotification(
	exchange, action, tokenPair string, fromTime, toTime uint64, isWarning bool, msg string) error {
	return rd.setting.UpdateExchangeNotification(exchange, action, tokenPair, fromTime, toTime, isWarning, msg)
}

func (rd ReserveData) GetRecords(fromTime, toTime uint64) ([]common.ActivityRecord, error) {
	return rd.storage.GetAllRecords(fromTime, toTime)
}

func (rd ReserveData) GetPendingActivities() ([]common.ActivityRecord, error) {
	return rd.storage.GetPendingActivities()
}

func (rd ReserveData) GetNotifications() (common.ExchangeNotifications, error) {
	return rd.setting.GetExchangeNotifications()
}

//Run run fetcher
func (rd ReserveData) Run() error {
	return rd.fetcher.Run()
}

//Stop stop the fetcher
func (rd ReserveData) Stop() error {
	return rd.fetcher.Stop()
}

//ControlAuthDataSize pack old data to file, push to S3 and prune outdated data
func (rd ReserveData) ControlAuthDataSize() error {
	for {
		log.Printf("DataPruner: waiting for signal from runner AuthData controller channel")
		t := <-rd.storageController.Runner.GetAuthBucketTicker()
		timepoint := common.TimeToTimepoint(t)
		log.Printf("DataPruner: got signal in AuthData controller channel with timestamp %d", common.TimeToTimepoint(t))
		fileName := fmt.Sprintf("./exported/ExpiredAuthData_at_%s", time.Unix(int64(timepoint/1000), 0).UTC())
		nRecord, err := rd.storage.ExportExpiredAuthData(common.TimeToTimepoint(t), fileName)
		if err != nil {
			log.Printf("ERROR: DataPruner export AuthData operation failed: %s", err)
		} else {
			var integrity bool
			if nRecord > 0 {
				err = rd.storageController.Arch.UploadFile(rd.storageController.Arch.GetReserveDataBucketName(), rd.storageController.ExpiredAuthDataPath, fileName)
				if err != nil {
					log.Printf("DataPruner: Upload file failed: %s", err)
				} else {
					integrity, err = rd.storageController.Arch.CheckFileIntergrity(rd.storageController.Arch.GetReserveDataBucketName(), rd.storageController.ExpiredAuthDataPath, fileName)
					if err != nil {
						log.Printf("ERROR: DataPruner: error in file integrity check (%s):", err)
					} else if !integrity {
						log.Printf("ERROR: DataPruner: file upload corrupted")

					}
					if err != nil || !integrity {
						//if the intergrity check failed, remove the remote file.
						removalErr := rd.storageController.Arch.RemoveFile(rd.storageController.Arch.GetReserveDataBucketName(), rd.storageController.ExpiredAuthDataPath, fileName)
						if removalErr != nil {
							log.Printf("ERROR: DataPruner: cannot remove remote file :(%s)", removalErr)
							return err
						}
					}
				}
			}
			if integrity && err == nil {
				nPrunedRecords, err := rd.storage.PruneExpiredAuthData(common.TimeToTimepoint(t))
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

func (rd ReserveData) GetTradeHistory(fromTime, toTime uint64) (common.AllTradeHistory, error) {
	data := common.AllTradeHistory{}
	data.Data = map[common.ExchangeID]common.ExchangeTradeHistory{}
	for _, ex := range rd.exchanges {
		history, err := ex.GetTradeHistory(fromTime, toTime)
		if err != nil {
			return data, err
		}
		data.Data[ex.ID()] = history
	}
	data.Timestamp = common.GetTimestamp()
	return data, nil
}

func (rd ReserveData) RunStorageController() error {
	if err := rd.storageController.Runner.Start(); err != nil {
		log.Fatalf("Storage controller runner error: %s", err.Error())
	}
	go func() {
		if err := rd.ControlAuthDataSize(); err != nil {
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
