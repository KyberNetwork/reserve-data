package data

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/archive"
	"github.com/KyberNetwork/reserve-data/data/datapruner"
	v3 "github.com/KyberNetwork/reserve-data/v3/common"
	"github.com/KyberNetwork/reserve-data/v3/storage"
)

//ReserveData struct for reserve data
type ReserveData struct {
	storage           Storage
	fetcher           Fetcher
	storageController datapruner.StorageController
	globalStorage     GlobalStorage
	exchanges         []common.Exchange
	settingStorage    storage.Interface
}

// CurrentGoldInfoVersion get current godl info version
func (rd ReserveData) CurrentGoldInfoVersion(timepoint uint64) (common.Version, error) {
	return rd.globalStorage.CurrentGoldInfoVersion(timepoint)
}

// CurrentBTCInfoVersion return
func (rd ReserveData) CurrentBTCInfoVersion(timepoint uint64) (common.Version, error) {
	return rd.globalStorage.CurrentBTCInfoVersion(timepoint)
}

// GetGoldData return gold data
func (rd ReserveData) GetGoldData(timestamp uint64) (common.GoldData, error) {
	version, err := rd.CurrentGoldInfoVersion(timestamp)
	if err != nil {
		return common.GoldData{}, nil
	}
	return rd.globalStorage.GetGoldInfo(version)
}

// GetBTCData return BTC data
func (rd ReserveData) GetBTCData(timestamp uint64) (common.BTCData, error) {
	version, err := rd.CurrentBTCInfoVersion(timestamp)
	if err != nil {
		return common.BTCData{}, nil
	}
	return rd.globalStorage.GetBTCInfo(version)
}

// UpdateFeedConfiguration update feed configuration
func (rd ReserveData) UpdateFeedConfiguration(name string, enabled bool) error {
	return rd.globalStorage.UpdateFeedConfiguration(name, enabled)
}

// GetFeedConfiguration return configuration for feed (gold, btc)
func (rd ReserveData) GetFeedConfiguration() ([]common.FeedConfiguration, error) {
	return rd.globalStorage.GetFeedConfiguration()
}

// CurrentPriceVersion return current price version
func (rd ReserveData) CurrentPriceVersion(timepoint uint64) (common.Version, error) {
	return rd.storage.CurrentPriceVersion(timepoint)
}

// GetAllPrices return all price
func (rd ReserveData) GetAllPrices(timepoint uint64) (common.AllPriceResponse, error) {
	timestamp := common.GetTimestamp()
	version, err := rd.storage.CurrentPriceVersion(timepoint)
	if err != nil {
		return common.AllPriceResponse{}, err
	}
	result := common.AllPriceResponse{}
	data, err := rd.storage.GetAllPrices(version)
	if err != nil {
		return common.AllPriceResponse{}, err
	}

	returnTime := common.GetTimestamp()
	result.Version = version
	result.Timestamp = timestamp
	result.ReturnTime = returnTime
	result.Data = data.Data
	result.Block = data.Block
	return result, err
}

// GetOnePrice return price of one pair tokens
func (rd ReserveData) GetOnePrice(pairIDs []uint64, timepoint uint64) (common.OnePriceResponse, error) {
	timestamp := common.GetTimestamp()
	version, err := rd.storage.CurrentPriceVersion(timepoint)
	if err != nil {
		return common.OnePriceResponse{}, err
	}
	result := common.OnePriceResponse{}
	data := make(common.OnePrice)
	for _, pairID := range pairIDs {
		r, err := rd.storage.GetOnePrice(pairID, version)
		if err != nil {
			return common.OnePriceResponse{}, errors.Wrapf(err, "failed to get one price of pair %d", pairID)
		}
		for k, v := range r {
			data[k] = v
		}
	}
	returnTime := common.GetTimestamp()
	result.Version = version
	result.Timestamp = timestamp
	result.ReturnTime = returnTime
	result.Data = data
	return result, err
}

// CurrentAuthDataVersion return current version of auth data
func (rd ReserveData) CurrentAuthDataVersion(timepoint uint64) (common.Version, error) {
	return rd.storage.CurrentAuthDataVersion(timepoint)
}

// GetAuthData return current auth data
// TODO: save AuthData using new format
func (rd ReserveData) GetAuthData(timepoint uint64) (common.AuthDataResponseV3, error) {
	version, err := rd.storage.CurrentAuthDataVersion(timepoint)
	if err != nil {
		return common.AuthDataResponseV3{}, err
	}
	result := common.AuthDataResponseV3{}
	data, err := rd.storage.GetAuthData(version)
	if err != nil {
		return common.AuthDataResponseV3{}, err
	}
	result.Version = version
	// result.PendingActivities = data.Pendingctivities
	pendingSetRate := []common.ActivityRecord{}
	pendingWithdraw := []common.ActivityRecord{}
	pendingDeposit := []common.ActivityRecord{}
	for _, activity := range data.PendingActivities {
		switch activity.Action {
		case common.ActionSetRate:
			pendingSetRate = append(pendingSetRate, activity)
		case common.ActionDeposit:
			pendingDeposit = append(pendingDeposit, activity)
		case common.ActionWithdraw:
			pendingWithdraw = append(pendingWithdraw, activity)
		}
	}
	result.PendingActivities.SetRates = pendingSetRate
	result.PendingActivities.Withdraw = pendingWithdraw
	result.PendingActivities.Deposit = pendingDeposit
	// map of token
	tokens := make(map[string]v3.Asset)
	exchanges := make(map[string]v3.Exchange)
	// get id from exchange balance asset name
	for exchangeID, balances := range data.ExchangeBalances {
		exchange, err := rd.settingStorage.GetExchangeByName(exchangeID.String())
		if err != nil {
			return result, errors.Wrapf(err, "failed to get exchange by name: %s", exchangeID.String())
		}
		exchanges[exchangeID.String()] = exchange
		for exchangeTokenSymbol := range balances.AvailableBalance {
			//* cos symbol of token in an exchange can be different then we need to use GetAssetExchangeBySymbol
			token, err := rd.settingStorage.GetAssetExchangeBySymbol(exchange.ID, exchangeTokenSymbol)
			//* it seems this token have balance in exchange but have not configured
			//* in core, just ignore it
			if err != nil {
				log.Printf("failed to get token by name: %s", exchangeTokenSymbol)
				continue
			}
			tokens[exchangeTokenSymbol] = token
		}
	}

	for tokenSymbol := range data.ReserveBalances {
		if _, exist := tokens[tokenSymbol]; !exist {
			token, err := rd.settingStorage.GetAssetBySymbol(tokenSymbol)
			//* it seems this token have balance in exchange but have not configured
			//* in core, just ignore it
			if err != nil {
				log.Printf("failed to get token by name: %s", tokenSymbol)
				continue
			}
			tokens[tokenSymbol] = token
		}
	}

	balances := []common.AuthdataBalance{}
	for tokenSymbol, token := range tokens {
		tokenBalance := common.AuthdataBalance{
			Valid: true,
		}
		tokenBalance.AssetID = token.ID
		tokenBalance.Symbol = tokenSymbol
		exchangeBalances := []common.ExchangeBalance{}
		for exchangeID, balances := range data.ExchangeBalances {
			exchangeBalance := common.ExchangeBalance{
				ExchangeID: exchanges[exchangeID.String()].ID,
			}
			if !balances.Valid {
				exchangeBalance.Error = balances.Error
			}
			if _, exist := balances.AvailableBalance[tokenSymbol]; exist {
				exchangeBalance.ExchangeID = exchanges[exchangeID.String()].ID
				exchangeBalance.Available = balances.AvailableBalance[tokenSymbol]
				exchangeBalance.Locked = balances.LockedBalance[tokenSymbol]
				exchangeBalance.Name = exchangeID.String()
			}
			exchangeBalances = append(exchangeBalances, exchangeBalance)
		}
		tokenBalance.Exchanges = exchangeBalances
		if balance, exist := data.ReserveBalances[tokenSymbol]; exist {
			tokenBalance.Reserve = balance.Balance.ToFloat(int64(token.Decimals))
			if !balance.Valid {
				tokenBalance.Valid = balance.Valid
				tokenBalance.ReserveError = balance.Error
			}
		}
		balances = append(balances, tokenBalance)
	}
	result.Balances = balances

	return result, err
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

// GetRates return all rates version
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

// GetRate return all rate
func (rd ReserveData) GetRate(timepoint uint64) (common.AllRateResponse, error) {
	timestamp := common.GetTimestamp()
	version, err := rd.storage.CurrentRateVersion(timepoint)
	if err != nil {
		return common.AllRateResponse{}, err
	}
	result := common.AllRateResponse{}
	rates, err := rd.storage.GetRate(version)
	if err != nil {
		return common.AllRateResponse{}, err
	}

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

// GetRecords return all records
func (rd ReserveData) GetRecords(fromTime, toTime uint64) ([]common.ActivityRecord, error) {
	return rd.storage.GetAllRecords(fromTime, toTime)
}

// GetPendingActivities return all pending activities
func (rd ReserveData) GetPendingActivities() ([]common.ActivityRecord, error) {
	return rd.storage.GetPendingActivities()
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
	tmpDir, err := ioutil.TempDir("", "ExpiredAuthData")
	if err != nil {
		return err
	}

	defer func() {
		if rErr := os.RemoveAll(tmpDir); rErr != nil {
			log.Printf("failed to cleanup temp dir: %s, err : %s", tmpDir, rErr.Error())
		}
	}()

	for {
		log.Printf("DataPruner: waiting for signal from runner AuthData controller channel")
		t := <-rd.storageController.Runner.GetAuthBucketTicker()
		timepoint := common.TimeToTimepoint(t)
		log.Printf("DataPruner: got signal in AuthData controller channel with timestamp %d", common.TimeToTimepoint(t))
		fileName := filepath.Join(tmpDir, fmt.Sprintf("ExpiredAuthData_at_%s", time.Unix(int64(timepoint/1000), 0).UTC()))
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
				switch {
				case err != nil:
					log.Printf("DataPruner: Can not prune Auth Data (%s)", err)
					return err
				case nPrunedRecords != nRecord:
					log.Printf("DataPruner: Number of Exported Data is %d, which is different from number of pruned data %d", nRecord, nPrunedRecords)
				default:
					log.Printf("DataPruner: exported and pruned %d expired records from AuthData", nRecord)
				}
			}
		}
		if err := os.Remove(fileName); err != nil {
			return err
		}
	}
}

// GetTradeHistory return trade history
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

// RunStorageController run storage controller
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
	exchanges []common.Exchange,
	settingStorage storage.Interface) *ReserveData {
	storageController, err := datapruner.NewStorageController(storageControllerRunner, arch)
	if err != nil {
		panic(err)
	}
	return &ReserveData{
		storage:           storage,
		fetcher:           fetcher,
		storageController: storageController,
		globalStorage:     globalStorage,
		exchanges:         exchanges,
		settingStorage:    settingStorage,
	}
}
