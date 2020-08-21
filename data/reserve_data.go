package data

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/archive"
	"github.com/KyberNetwork/reserve-data/data/datapruner"
	"github.com/KyberNetwork/reserve-data/lib/rtypes"
	v3 "github.com/KyberNetwork/reserve-data/reservesetting/common"
	"github.com/KyberNetwork/reserve-data/reservesetting/storage"
)

//ReserveData struct for reserve data
type ReserveData struct {
	storage           Storage
	fetcher           Fetcher
	storageController datapruner.StorageController
	globalStorage     GlobalStorage
	exchanges         []common.Exchange
	settingStorage    storage.Interface
	l                 *zap.SugaredLogger
}

// CurrentGoldInfoVersion get current godl info version
func (rd ReserveData) CurrentGoldInfoVersion(timepoint uint64) (common.Version, error) {
	return rd.globalStorage.CurrentGoldInfoVersion(timepoint)
}

// CurrentBTCInfoVersion return
func (rd ReserveData) CurrentBTCInfoVersion(timepoint uint64) (common.Version, error) {
	return rd.globalStorage.CurrentBTCInfoVersion(timepoint)
}

// CurrentUSDInfoVersion return
func (rd ReserveData) CurrentUSDInfoVersion(timepoint uint64) (common.Version, error) {
	return rd.globalStorage.CurrentUSDInfoVersion(timepoint)
}

// GetGoldData return gold data
func (rd ReserveData) GetGoldData(timestamp uint64) (common.GoldData, error) {
	version, err := rd.CurrentGoldInfoVersion(timestamp)
	if err != nil {
		rd.l.Errorw("cannot get gold data version", "error", err)
		return common.GoldData{}, err
	}
	return rd.globalStorage.GetGoldInfo(version)
}

// GetBTCData return BTC data
func (rd ReserveData) GetBTCData(timestamp uint64) (common.BTCData, error) {
	version, err := rd.CurrentBTCInfoVersion(timestamp)
	if err != nil {
		rd.l.Errorw("cannot get BTC data version", "error", err)
		return common.BTCData{}, err
	}
	return rd.globalStorage.GetBTCInfo(version)
}

// GetUSDData return USD data
func (rd ReserveData) GetUSDData(timestamp uint64) (common.USDData, error) {
	version, err := rd.CurrentUSDInfoVersion(timestamp)
	if err != nil {
		rd.l.Errorw("cannot get USD data version", "error", err)
		return common.USDData{}, err
	}
	return rd.globalStorage.GetUSDInfo(version)
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
func (rd ReserveData) GetOnePrice(pairID rtypes.TradingPairID, timepoint uint64) (common.OnePriceResponse, error) {
	timestamp := common.GetTimestamp()
	version, err := rd.storage.CurrentPriceVersion(timepoint)
	if err != nil {
		return common.OnePriceResponse{}, err
	}
	result := common.OnePriceResponse{}
	data, err := rd.storage.GetOnePrice(pairID, version)
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
	assets := make(map[rtypes.AssetID]v3.Asset)
	exchanges := make(map[string]v3.Exchange)
	// get id from exchange balance asset name
	for exchangeID, balances := range data.ExchangeBalances {
		exchange, err := rd.settingStorage.GetExchangeByName(exchangeID.String())
		if err != nil {
			return result, errors.Wrapf(err, "failed to get exchange by name: %s", exchangeID.String())
		}
		exchanges[exchangeID.String()] = exchange
		for assetID := range balances.AvailableBalance {
			//* cos symbol of token in an exchange can be different then we need to use GetAssetExchangeBySymbol
			token, err := rd.settingStorage.GetAsset(assetID)
			//* it seems this token have balance in exchange but have not configured
			//* in core, just ignore it
			if err != nil {
				rd.l.Warnw("failed to get token by name", "symbol", assetID, "err", err)
				continue
			}
			assets[token.ID] = token
		}
	}

	for assetID := range data.ReserveBalances {
		if _, exist := assets[assetID]; !exist {
			token, err := rd.settingStorage.GetAsset(assetID)
			//* it seems this token have balance in exchange but have not configured
			//* in core, just ignore it
			if err != nil {
				rd.l.Warnw("failed to get token by id", "assetID", assetID, "err", err)
				continue
			}
			assets[assetID] = token
		}
	}

	var balances []common.AuthdataBalance
	for assetID, token := range assets {
		tokenBalance := common.AuthdataBalance{
			Valid: true,
		}
		tokenBalance.AssetID = token.ID
		tokenBalance.Symbol = token.Symbol
		var exchangeBalances []common.ExchangeBalance
		for exchangeID, balances := range data.ExchangeBalances {
			if _, exist := balances.AvailableBalance[token.ID]; !exist {
				continue
			}

			exchangeBalance := common.ExchangeBalance{
				ExchangeID: exchanges[exchangeID.String()].ID,
				Name:       exchangeID.String(),
			}
			if balances.Error != "" {
				exchangeBalance.Error = balances.Error
				tokenBalance.Valid = false
			}
			exchangeBalance.ExchangeID = exchanges[exchangeID.String()].ID
			exchangeBalance.Available = balances.AvailableBalance[token.ID]
			exchangeBalance.Locked = balances.LockedBalance[token.ID]
			exchangeBalances = append(exchangeBalances, exchangeBalance)

		}
		tokenBalance.Exchanges = exchangeBalances
		if balance, exist := data.ReserveBalances[assetID]; exist {
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

func isDuplicated(oldData, newData map[rtypes.AssetID]common.RateResponse) bool {
	if len(oldData) != len(newData) {
		return false
	}
	for tokenID, oldElem := range oldData {
		newElem, ok := newData[tokenID]
		if !ok {
			return false
		}
		if oldElem.BaseBuy != newElem.BaseBuy {
			return false
		}
		if oldElem.CompactBuy != newElem.CompactBuy {
			return false
		}
		if oldElem.BaseSell != newElem.BaseSell {
			return false
		}
		if oldElem.CompactSell != newElem.CompactSell {
			return false
		}
		if oldElem.Rate != newElem.Rate {
			return false
		}
	}
	return true
}

func getOneRateData(rate common.AllRateEntry) map[rtypes.AssetID]common.RateResponse {
	//get data from rate object and return the data.
	data := map[rtypes.AssetID]common.RateResponse{}
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

// GetAssetRateTriggers query count of setRate with trigger=true, for each asset
func (rd ReserveData) GetAssetRateTriggers(fromTime uint64, toTime uint64) (map[rtypes.AssetID]int, error) {
	triggers, err := rd.storage.GetAssetRateTriggers(fromTime, toTime)
	if err != nil {
		return nil, err
	}
	res := make(map[rtypes.AssetID]int)
	for _, t := range triggers {
		res[t.AssetID] = t.Count
	}
	return res, nil
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
	data := map[rtypes.AssetID]common.RateResponse{}
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
// params: fromTime, toTime milisecond
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
			rd.l.Errorw("failed to cleanup temp dir", "tmpdir", tmpDir, "err", rErr)
		}
	}()

	for {
		rd.l.Info("DataPruner: waiting for signal from runner AuthData controller channel")
		t := <-rd.storageController.Runner.GetAuthBucketTicker()
		timepoint := common.TimeToMillis(t)
		rd.l.Infow("DataPruner: got signal in AuthData controller channel", "timestamp", common.TimeToMillis(t))
		fileName := filepath.Join(tmpDir, fmt.Sprintf("ExpiredAuthData_at_%s", time.Unix(int64(timepoint/1000), 0).UTC()))
		nRecord, err := rd.storage.ExportExpiredAuthData(common.TimeToMillis(t), fileName)
		if err != nil {
			rd.l.Errorw("ERROR: DataPruner export AuthData operation failed", "err", err, "file", fileName)
		} else {
			var integrity bool
			if nRecord > 0 {
				err = rd.storageController.Arch.UploadFile(rd.storageController.Arch.GetReserveDataBucketName(), rd.storageController.ExpiredAuthDataPath, fileName)
				if err != nil {
					rd.l.Errorw("DataPruner: Upload file failed", "err", err, "file", fileName)
				} else {
					integrity, err = rd.storageController.Arch.CheckFileIntergrity(rd.storageController.Arch.GetReserveDataBucketName(), rd.storageController.ExpiredAuthDataPath, fileName)
					if err != nil {
						rd.l.Errorw("ERROR: DataPruner: error in file integrity check", "err", err)
					} else if !integrity {
						rd.l.Errorw("ERROR: DataPruner: file upload corrupted")

					}
					if err != nil || !integrity {
						//if the intergrity check failed, remove the remote file.
						removalErr := rd.storageController.Arch.RemoveFile(rd.storageController.Arch.GetReserveDataBucketName(), rd.storageController.ExpiredAuthDataPath, fileName)
						if removalErr != nil {
							rd.l.Warnw("ERROR: DataPruner: cannot remove remote file", "err", removalErr, "file", fileName)
							return err
						}
					}
				}
			}
			if integrity && err == nil {
				nPrunedRecords, err := rd.storage.PruneExpiredAuthData(common.TimeToMillis(t))
				switch {
				case err != nil:
					rd.l.Errorw("DataPruner: Can not prune Auth Data", "err", err)
					return err
				case nPrunedRecords != nRecord:
					rd.l.Infof("DataPruner: Number of Exported Data is %d, which is different from number of pruned data %d", nRecord, nPrunedRecords)
				default:
					rd.l.Infof("DataPruner: exported and pruned %d expired records from AuthData", nRecord)
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
	data.Data = map[rtypes.ExchangeID]common.ExchangeTradeHistory{}
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
		rd.l.Fatalw("Storage controller runner error", "err", err)
	}
	go func() {
		if err := rd.ControlAuthDataSize(); err != nil {
			rd.l.Errorw("Control auth data size failed", "err", err)
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
		l:                 zap.S(),
	}
}
