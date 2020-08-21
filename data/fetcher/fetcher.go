package fetcher

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	ethereum "github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/blockchain"
	"github.com/KyberNetwork/reserve-data/lib/rtypes"
)

// maxActivityLifeTime is the longest time of an activity. If the
// activity is pending for more than MAX_ACVITY_LIFE_TIME, it will be
// considered as failed.
const maxActivityLifeTime uint64 = 6 // activity max life time in hour

type Fetcher struct {
	storage                Storage
	globalStorage          GlobalStorage
	exchanges              []Exchange
	blockchain             Blockchain
	theworld               TheWorld
	runner                 Runner
	currentBlock           uint64
	currentBlockUpdateTime uint64
	simulationMode         bool
	contractAddressConf    *common.ContractAddressConfiguration
	l                      *zap.SugaredLogger
}

func NewFetcher(
	storage Storage,
	globalStorage GlobalStorage,
	theworld TheWorld,
	runner Runner,
	simulationMode bool,
	contractAddressConf *common.ContractAddressConfiguration) *Fetcher {
	return &Fetcher{
		storage:             storage,
		globalStorage:       globalStorage,
		exchanges:           []Exchange{},
		blockchain:          nil,
		theworld:            theworld,
		runner:              runner,
		simulationMode:      simulationMode,
		contractAddressConf: contractAddressConf,
		l:                   zap.S(),
	}
}

func (f *Fetcher) SetBlockchain(blockchain Blockchain) {
	f.blockchain = blockchain
	f.FetchCurrentBlock(common.NowInMillis())
}

func (f *Fetcher) AddExchange(exchange Exchange) {
	f.exchanges = append(f.exchanges, exchange)
}

func (f *Fetcher) Stop() error {
	return f.runner.Stop()
}

func (f *Fetcher) Run() error {
	f.l.Info("Fetcher runner is starting...")
	if err := f.runner.Start(); err != nil {
		return err
	}
	go f.RunOrderbookFetcher()
	go f.RunAuthDataFetcher()
	go f.RunRateFetcher()
	go f.RunBlockFetcher()
	go f.RunGlobalDataFetcher()
	go f.RunFetchExchangeHistory()
	f.l.Infof("Fetcher runner is running...")
	return nil
}

func (f *Fetcher) RunGlobalDataFetcher() {
	for {
		f.l.Info("waiting for signal from global data channel")
		t := <-f.runner.GetGlobalDataTicker()
		f.l.Infof("got signal in global data channel with timestamp %d", common.TimeToMillis(t))
		timepoint := common.TimeToMillis(t)
		f.FetchGlobalData(timepoint)
		f.l.Info("fetched block from blockchain")
	}
}

func (f *Fetcher) FetchGlobalData(timepoint uint64) {
	goldData, err := f.theworld.GetGoldInfo()
	if err != nil {
		f.l.Infof("failed to fetch Gold Info: %s", err.Error())
		return
	}
	goldData.Timestamp = common.NowInMillis()

	if err = f.globalStorage.StoreGoldInfo(goldData); err != nil {
		f.l.Infof("Storing gold info failed: %s", err.Error())
	}

	btcData, err := f.theworld.GetBTCInfo()
	if err != nil {
		f.l.Infof("failed to fetch BTC Info: %s", err.Error())
		return
	}
	btcData.Timestamp = common.NowInMillis()
	if err = f.globalStorage.StoreBTCInfo(btcData); err != nil {
		f.l.Infof("Storing BTC info failed: %s", err.Error())
	}

	usdData, err := f.theworld.GetUSDInfo()
	if err != nil {
		f.l.Warnw("failed to fetch USD info", "err", err)
		return
	}
	usdData.Timestamp = common.NowInMillis()
	if err = f.globalStorage.StoreUSDInfo(usdData); err != nil {
		f.l.Warnw("Store USD info failed", "err", err)
	}
}

func (f *Fetcher) RunBlockFetcher() {
	for {
		f.l.Info("waiting for signal from block channel")
		t := <-f.runner.GetBlockTicker()
		f.l.Infof("got signal in block channel with timestamp %d", common.TimeToMillis(t))
		timepoint := common.TimeToMillis(t)
		f.FetchCurrentBlock(timepoint)
		f.l.Info("fetched block from blockchain")
	}
}

func (f *Fetcher) RunRateFetcher() {
	for {
		f.l.Infof("waiting for signal from runner rate channel")
		t := <-f.runner.GetRateTicker()
		f.l.Infof("got signal in rate channel with timestamp %d", common.TimeToMillis(t))
		f.FetchRate(common.TimeToMillis(t))
		f.l.Infof("fetched rates from blockchain")
	}
}

func (f *Fetcher) FetchRate(timepoint uint64) {
	var (
		err  error
		data common.AllRateEntry
	)
	// only fetch rates 5s after the block number is updated
	if !f.simulationMode && f.currentBlockUpdateTime-timepoint <= 5000 {
		return
	}

	var atBlock = f.currentBlock - 1
	// in simulation mode, just fetches from latest known block
	if f.simulationMode {
		atBlock = 0
	}

	data, err = f.blockchain.FetchRates(atBlock, f.currentBlock)
	if err != nil {
		f.l.Warnw("Fetching rates from blockchain failed. Will not store it to storage.", "err", err)
		return
	}

	f.l.Infof("Got rates from blockchain: %+v", data)
	if err = f.storage.StoreRate(data, timepoint); err != nil {
		f.l.Errorw("Storing rates failed", "err", err)
	}
}

func (f *Fetcher) RunAuthDataFetcher() {
	for {
		f.l.Infof("waiting for signal from runner auth data channel")
		t := <-f.runner.GetAuthDataTicker()
		f.l.Infof("got signal in auth data channel with timestamp %d", common.TimeToMillis(t))
		f.FetchAllAuthData(common.TimeToMillis(t))
		f.l.Infof("fetched data from exchanges")
	}
}

func (f *Fetcher) FetchAllAuthData(timepoint uint64) {
	snapshot := common.AuthDataSnapshot{
		Valid:             true,
		Timestamp:         common.GetTimestamp(),
		ExchangeBalances:  map[rtypes.ExchangeID]common.EBalanceEntry{},
		ReserveBalances:   map[rtypes.AssetID]common.BalanceEntry{},
		PendingActivities: []common.ActivityRecord{},
		Block:             0,
	}
	bbalances := map[rtypes.AssetID]common.BalanceEntry{}
	ebalances := sync.Map{}
	estatuses := sync.Map{}
	bstatuses := sync.Map{}
	pendings, err := f.storage.GetPendingActivities()
	if err != nil {
		f.l.Errorw("Getting pending activities failed", "err", err)
		return
	}
	wait := sync.WaitGroup{}
	for _, exchange := range f.exchanges {
		wait.Add(1)
		go f.FetchAuthDataFromExchange(
			&wait, exchange, &ebalances, &estatuses,
			pendings, timepoint)
	}
	wait.Wait()
	// if we got tx info of withdrawals from the cexs, we have to
	// update them to pending activities in order to also check
	// their mining status.
	// otherwise, if the txs are already mined and the reserve
	// balances are already changed, their mining statuses will
	// still be "", which can lead analytic to intepret the balances
	// wrongly.
	for _, activity := range pendings {
		status, found := estatuses.Load(activity.ID)
		if found {
			activityStatus, ok := status.(common.ActivityStatus)
			if !ok {
				f.l.Warnw("status from cexs cannot be asserted to common.ActivityStatus")
				continue
			}
			//Set activity result tx to tx from cexs if currently result tx is not nil an is an empty string
			if activity.Result.Tx == "" {
				activity.Result.Tx = activityStatus.Tx
			}
		}
	}

	if err = f.FetchAuthDataFromBlockchain(bbalances, &bstatuses, pendings); err != nil {
		snapshot.Error = err.Error()
		snapshot.Valid = false
	}
	snapshot.Block = f.currentBlock
	snapshot.ReturnTime = common.GetTimestamp()
	err = f.PersistSnapshot(
		&ebalances, bbalances, &estatuses, &bstatuses,
		pendings, &snapshot, timepoint)
	if err != nil {
		f.l.Warnw("Storing exchange balances failed", "err", err)
		return
	}
}

func (f *Fetcher) FetchAuthDataFromBlockchain(
	allBalances map[rtypes.AssetID]common.BalanceEntry,
	allStatuses *sync.Map,
	pendings []common.ActivityRecord) error {
	// we apply double check strategy to mitigate race condition on exchange side like this:
	// 1. Get list of pending activity status (A)
	// 2. Get list of balances (B)
	// 3. Get list of pending activity status again (C)
	// 4. if C != A, repeat 1, otherwise return A, B

	/*
		we try to build a consistent view of pending activities and balances,
		activities update(eg, some txs become complete) can make balances result looks wrong
		so, we verify activities status before and after we collect balances, make sure it does not change.
	*/
	var balances map[rtypes.AssetID]common.BalanceEntry
	var preStatuses, statuses map[common.ActivityID]common.ActivityStatus
	var err error
	for {
		preStatuses, err = f.FetchStatusFromBlockchain(pendings)
		if err != nil {
			f.l.Warnw("Fetching blockchain pre statuses failed, retrying", "err", err)
		}
		balances, err = f.FetchBalanceFromBlockchain()
		if err != nil {
			f.l.Warnw("Fetching blockchain balances failed", "err", err)
			return err
		}
		statuses, err = f.FetchStatusFromBlockchain(pendings)
		if err != nil {
			f.l.Warnw("Fetching blockchain statuses failed, retrying", "err", err)
		}
		if unchanged(preStatuses, statuses) {
			break
		}
	}
	for k, v := range balances {
		allBalances[k] = v
	}
	for id, activityStatus := range statuses {
		allStatuses.Store(id, activityStatus)
	}
	return nil
}

func (f *Fetcher) FetchCurrentBlock(timepoint uint64) {
	block, err := f.blockchain.CurrentBlock()
	if err != nil {
		f.l.Warnw("Fetching current block failed, ignored.", "err", err)
	} else {
		// update currentBlockUpdateTime first to avoid race condition
		// where fetcher is trying to fetch new rate
		f.currentBlockUpdateTime = common.NowInMillis()
		f.currentBlock = block
	}
}

func (f *Fetcher) FetchBalanceFromBlockchain() (map[rtypes.AssetID]common.BalanceEntry, error) {
	return f.blockchain.FetchBalanceData(f.contractAddressConf.Reserve, 0)
}

func (f *Fetcher) newNonceValidator() func(common.ActivityRecord) bool {
	// GetMinedNonceWithOP might be slow, use closure to not invoke it every time
	minedNonce, err := f.blockchain.GetMinedNonceWithOP(blockchain.PricingOP)
	if err != nil {
		f.l.Warnw("Getting mined nonce failed", "err", err)
	}

	return func(act common.ActivityRecord) bool {
		// this check only works with set rate transaction as:
		//   - account nonce is record in result field of activity
		//   - the GetMinedNonceWithOP method is available
		if act.Action != common.ActionSetRate {
			return false
		}
		return act.Result.Nonce < minedNonce
	}
}

func (f *Fetcher) FetchStatusFromBlockchain(pendings []common.ActivityRecord) (map[common.ActivityID]common.ActivityStatus, error) {
	result := map[common.ActivityID]common.ActivityStatus{}
	nonceValidator := f.newNonceValidator()

	for _, activity := range pendings {
		if activity.IsBlockchainPending() && (activity.Action == common.ActionSetRate || activity.Action == common.ActionDeposit || activity.Action == common.ActionWithdraw || activity.Action == common.ActionCancelSetRate) {
			var (
				blockNum uint64
				status   string
				err      error
			)
			txStr := activity.Result.Tx
			tx := ethereum.HexToHash(txStr)
			if tx.Big().IsInt64() && tx.Big().Int64() == 0 {
				continue
			}
			status, blockNum, err = f.blockchain.TxStatus(tx)
			if err != nil {
				return result, fmt.Errorf("TX_STATUS: ERROR Getting tx %s status failed: %s", txStr, err)
			}

			switch status {
			case common.MiningStatusPending:
				f.l.Infof("TX_STATUS: tx (%s) status is pending", tx)
			case common.MiningStatusMined:
				if activity.Action == common.ActionSetRate {
					f.l.Infof("TX_STATUS set rate transaction is mined, id: %s", activity.ID.EID)
				}
				result[activity.ID] = common.NewActivityStatus(
					activity.ExchangeStatus,
					txStr,
					blockNum,
					common.MiningStatusMined,
					0,
					err,
				)
			case common.MiningStatusFailed:
				f.l.Warnw("transaction failed to mine", "tx", tx)
				result[activity.ID] = common.NewActivityStatus(
					activity.ExchangeStatus,
					txStr,
					blockNum,
					common.MiningStatusFailed,
					0,
					err,
				)
			case common.MiningStatusLost:
				var (
					// expiredDuration is the amount of time after that if a transaction doesn't appear,
					// it is considered failed
					expiredDuration = 15 * time.Minute / time.Millisecond
					txFailed        = false
				)
				if nonceValidator(activity) {
					txFailed = true
				} else {
					elapsed := common.NowInMillis() - activity.Timestamp.Millis()
					if elapsed > uint64(expiredDuration) {
						f.l.Infof("TX_STATUS: tx(%s) is lost, elapsed time: %d", txStr, elapsed)
						txFailed = true
					}
				}

				if txFailed {
					result[activity.ID] = common.NewActivityStatus(
						activity.ExchangeStatus,
						txStr,
						blockNum,
						common.MiningStatusFailed,
						0,
						err,
					)
				}
			default:
				f.l.Infof("TX_STATUS: tx (%s) status is not available. Wait till next try", tx)
			}
		}
	}
	return result, nil
}

func unchanged(pre, post map[common.ActivityID]common.ActivityStatus) bool {
	if len(pre) != len(post) {
		return false
	}
	for k, v := range pre {
		vpost, found := post[k]
		if !found {
			return false
		}
		if v.ExchangeStatus != vpost.ExchangeStatus ||
			v.MiningStatus != vpost.MiningStatus ||
			v.Tx != vpost.Tx {
			return false
		}
	}
	return true
}

func (f *Fetcher) updateActivitywithBlockchainStatus(activity *common.ActivityRecord, bstatuses *sync.Map, snapshot *common.AuthDataSnapshot) {
	status, ok := bstatuses.Load(activity.ID)
	if !ok || status == nil {
		f.l.Infof("block chain status for %s is nil or not existed ", activity.ID.String())
		return
	}

	activityStatus, ok := status.(common.ActivityStatus)
	if !ok {
		f.l.Errorw("ERROR: status cannot be asserted to common.ActivityStatus", "status", status)
		return
	}
	f.l.Infof("In PersistSnapshot: blockchain activity status for %+v: %+v", activity.ID, activityStatus)
	if activity.IsBlockchainPending() {
		activity.MiningStatus = activityStatus.MiningStatus
	}

	if activityStatus.ExchangeStatus == common.ExchangeStatusFailed {
		activity.ExchangeStatus = activityStatus.ExchangeStatus
	}

	if activityStatus.Error != nil {
		snapshot.Valid = false
		snapshot.Error = activityStatus.Error.Error()
		activity.Result.StatusError = activityStatus.Error.Error()
	} else {
		activity.Result.StatusError = ""
	}
	activity.Result.BlockNumber = activityStatus.BlockNumber
}

func (f *Fetcher) updateActivitywithExchangeStatus(activity *common.ActivityRecord, estatuses *sync.Map, snapshot *common.AuthDataSnapshot) {
	status, ok := estatuses.Load(activity.ID)
	if !ok || status == nil {
		f.l.Infof("exchange status for %s is nil or not existed ", activity.ID.String())
		return
	}
	activityStatus, ok := status.(common.ActivityStatus)
	if !ok {
		f.l.Errorw("ERROR: status cannot be asserted to common.ActivityStatus", "status", status)
		return
	}
	f.l.Infof("In PersistSnapshot: exchange activity status for %+v: %+v", activity.ID, activityStatus)
	if activity.IsExchangePending() {
		activity.ExchangeStatus = activityStatus.ExchangeStatus
	} else if activityStatus.ExchangeStatus == common.ExchangeStatusFailed {
		activity.ExchangeStatus = activityStatus.ExchangeStatus
		activity.Result.WithdrawFee = activityStatus.WithdrawFee
	}

	if activity.Result.Tx == "" {
		activity.Result.Tx = activityStatus.Tx
	}

	if activityStatus.Error != nil {
		snapshot.Valid = false
		snapshot.Error = activityStatus.Error.Error()
		activity.Result.StatusError = activityStatus.Error.Error()
		activity.Result.WithdrawFee = activityStatus.WithdrawFee
	} else {
		activity.Result.StatusError = ""
		activity.Result.WithdrawFee = activityStatus.WithdrawFee
	}
}

// PersistSnapshot save a authdata snapshot into db
func (f *Fetcher) PersistSnapshot(
	ebalances *sync.Map,
	bbalances map[rtypes.AssetID]common.BalanceEntry,
	estatuses *sync.Map,
	bstatuses *sync.Map,
	pendings []common.ActivityRecord,
	snapshot *common.AuthDataSnapshot,
	timepoint uint64) error {

	allEBalances := map[rtypes.ExchangeID]common.EBalanceEntry{}
	ebalances.Range(func(key, value interface{}) bool {
		//if type conversion went wrong, continue to the next record
		v, ok := value.(common.EBalanceEntry)
		if !ok {
			f.l.Errorw("ERROR: value cannot be asserted to common.EbalanceEntry", "value", v)
			return true
		}
		exID, ok := key.(rtypes.ExchangeID)
		if !ok {
			f.l.Errorw("key cannot be asserted to common.ExchangeID", "key", key)
			return true
		}
		allEBalances[exID] = v
		if !v.Valid {
			// get old auth data, because get balance error then we have to keep
			// balance to the latest version then analytic won't get exchange balance to zero
			authVersion, err := f.storage.CurrentAuthDataVersion(common.NowInMillis())
			if err == nil {
				oldAuth, err := f.storage.GetAuthData(authVersion)
				if err != nil {
					allEBalances[exID] = common.EBalanceEntry{
						Error: err.Error(),
					}
				} else {
					// update old auth to current
					newEbalance := oldAuth.ExchangeBalances[exID]
					newEbalance.Error = v.Error
					newEbalance.Status = false
					allEBalances[exID] = newEbalance
				}
			}
			snapshot.Valid = false
			snapshot.Error = v.Error
		}
		return true
	})

	pendingActivities := []common.ActivityRecord{}
	for _, activity := range pendings {
		activity := activity
		f.updateActivitywithExchangeStatus(&activity, estatuses, snapshot)
		f.updateActivitywithBlockchainStatus(&activity, bstatuses, snapshot)
		f.l.Infof("Aggregate statuses, final activity: %+v", activity)
		if activity.IsPending() {
			pendingActivities = append(pendingActivities, activity)
		}
		err := f.storage.UpdateActivity(activity.ID, activity)
		if err != nil {
			snapshot.Valid = false
			snapshot.Error = err.Error()
		}
	}
	// note: only update status when it's pending status
	snapshot.ExchangeBalances = allEBalances

	// persist blockchain balance
	// if blockchain balance is not valid then auth snapshot will also not valid
	for _, balance := range bbalances {
		if !balance.Valid {
			snapshot.Valid = false
			if balance.Error != "" {
				if snapshot.Error != "" {
					snapshot.Error += "; " + balance.Error
				} else {
					snapshot.Error = balance.Error
				}
			}
		}
	}
	// persist blockchain balances
	snapshot.ReserveBalances = bbalances
	snapshot.PendingActivities = pendingActivities
	return f.storage.StoreAuthSnapshot(snapshot, timepoint)
}

func (f *Fetcher) FetchAuthDataFromExchange(
	wg *sync.WaitGroup, exchange Exchange,
	allBalances *sync.Map, allStatuses *sync.Map,
	pendings []common.ActivityRecord,
	timepoint uint64) {
	defer wg.Done()
	// we apply double check strategy to mitigate race condition on exchange side like this:
	// 1. Get list of pending activity status (A)
	// 2. Get list of balances (B)
	// 3. Get list of pending activity status again (C)
	// 4. if C != A, repeat 1, otherwise return A, B
	var balances common.EBalanceEntry
	var statuses map[common.ActivityID]common.ActivityStatus
	var err error
	var tokenAddress map[rtypes.AssetID]ethereum.Address
	for {
		preStatuses := f.FetchStatusFromExchange(exchange, pendings, timepoint)
		balances, err = exchange.FetchEBalanceData(timepoint)
		if err != nil {
			f.l.Warnw("Fetching exchange balances failed", "exchange", exchange.ID().String(), "err", err)
			break
		}
		//Remove all token which is not in this exchange's token addresses
		tokenAddress, err = exchange.TokenAddresses()
		if err != nil {
			f.l.Warnw("getting token address failed: %v", "exchange", exchange.ID().String(), "err", err)
			break
		}
		for tokenID := range balances.AvailableBalance {
			if _, ok := tokenAddress[tokenID]; !ok {
				delete(balances.AvailableBalance, tokenID)
			}
		}

		for tokenID := range balances.LockedBalance {
			if _, ok := tokenAddress[tokenID]; !ok {
				delete(balances.LockedBalance, tokenID)
			}
		}

		for tokenID := range balances.DepositBalance {
			if _, ok := tokenAddress[tokenID]; !ok {
				delete(balances.DepositBalance, tokenID)
			}
		}

		statuses = f.FetchStatusFromExchange(exchange, pendings, timepoint)
		if unchanged(preStatuses, statuses) {
			break
		}
	}
	if err == nil {
		allBalances.Store(exchange.ID(), balances)
		for id, activityStatus := range statuses {
			allStatuses.Store(id, activityStatus)
		}
	}
}

// FetchStatusFromExchange return status of activity from exchange
func (f *Fetcher) FetchStatusFromExchange(exchange Exchange, pendings []common.ActivityRecord, timepoint uint64) map[common.ActivityID]common.ActivityStatus {
	result := map[common.ActivityID]common.ActivityStatus{}
	for _, activity := range pendings {
		if activity.Destination != exchange.ID().String() {
			continue
		}
		if activity.IsExchangePending() {
			var (
				err        error
				status, tx string
				blockNum   uint64
				fee        float64
			)

			id := activity.ID
			//These type conversion errors can be ignore since if happens, it will be reflected in activity.error

			switch activity.Action {
			case common.ActionTrade:
				orderID := id.EID
				base := activity.Params.Base
				quote := activity.Params.Quote
				// we ignore error of order status because it doesn't affect
				// authdata. Analytic will ignore order status anyway.
				status, _ = exchange.OrderStatus(orderID, base, quote)
			case common.ActionDeposit:
				txHash := activity.Result.Tx
				amount := activity.Params.Amount
				assetID := activity.Params.Asset

				status, err = exchange.DepositStatus(id, txHash, assetID, amount, timepoint)
				f.l.Infof("Got deposit status for %v: (%s), error(%v)", activity, status, err)
			case common.ActionWithdraw:
				amount := activity.Params.Amount
				assetID := activity.Params.Asset

				status, tx, fee, err = exchange.WithdrawStatus(id.EID, assetID, amount, timepoint)
				f.l.Infof("Got withdraw status for %v: (%s), error(%v)", activity, status, err)
			default:
				continue
			}

			// in case there is something wrong with the cex and the activity is stuck for a very
			// long time. We will just consider it as a failed activity.
			timepoint, err1 := strconv.ParseUint(string(activity.Timestamp), 10, 64)
			if err1 != nil {
				f.l.Infof("Activity %+v has invalid timestamp. Just ignore it.", activity)
			} else {
				if common.NowInMillis()-timepoint > maxActivityLifeTime*uint64(time.Hour)/uint64(time.Millisecond) {
					result[id] = common.NewActivityStatus(common.ExchangeStatusFailed, tx, blockNum, activity.MiningStatus, fee, err)
				} else {
					result[id] = common.NewActivityStatus(status, tx, blockNum, activity.MiningStatus, fee, err)
				}
			}
		} else {
			timepoint, err1 := strconv.ParseUint(string(activity.Timestamp), 10, 64)
			if err1 != nil {
				f.l.Infof("Activity %+v has invalid timestamp. Just ignore it.", activity)
				continue
			}
			if activity.ExchangeStatus == common.ExchangeStatusDone &&
				common.NowInMillis()-timepoint > maxActivityLifeTime*uint64(time.Hour)/uint64(time.Millisecond) {
				// the activity is still pending but its exchange status is done and it is stuck there for more than
				// maxActivityLifeTime. This activity is considered failed.
				result[activity.ID] = common.NewActivityStatus(common.ExchangeStatusFailed, "", 0, activity.MiningStatus, 0, nil)
			}
		}
	}
	return result
}

func (f *Fetcher) RunOrderbookFetcher() {
	for {
		f.l.Infof("waiting for signal from runner orderbook channel")
		t := <-f.runner.GetOrderbookTicker()
		f.l.Infof("got signal in orderbook channel with timestamp %d", common.TimeToMillis(t))
		f.FetchOrderbook(common.TimeToMillis(t))
		f.l.Info("fetched data from exchanges")
	}
}

func (f *Fetcher) FetchOrderbook(timepoint uint64) {
	data := NewConcurrentAllPriceData()
	// start fetching
	wait := sync.WaitGroup{}
	for _, exchange := range f.exchanges {
		wait.Add(1)
		go f.fetchPriceFromExchange(&wait, exchange, data, timepoint)
	}
	wait.Wait()
	data.SetBlockNumber(f.currentBlock)
	err := f.storage.StorePrice(data.GetData(), timepoint)
	if err != nil {
		f.l.Warnw("Storing data failed", "err", err)
	}
}

func (f *Fetcher) fetchPriceFromExchange(wg *sync.WaitGroup, exchange Exchange, data *ConcurrentAllPriceData, timepoint uint64) {
	defer wg.Done()
	exdata, err := exchange.FetchPriceData(timepoint)
	if err != nil {
		f.l.Warnw("Fetching data failed", "exchange", exchange.ID().String(), "err", err)
		return
	}
	for pair, exchangeData := range exdata {
		data.SetOnePrice(exchange.ID(), pair, exchangeData)
	}
}

// RunFetchExchangeHistory starts a fetcher to get exchange trade history
func (f *Fetcher) RunFetchExchangeHistory() {
	for ; ; <-f.runner.GetExchangeHistoryTicker() {
		f.l.Info("got signal in orderbook channel with exchange-history")
		f.fetchExchangeTradeHistory()
		f.l.Info("fetched data from exchanges")
	}
}

func (f *Fetcher) fetchExchangeTradeHistory() {
	wait := sync.WaitGroup{}
	for _, exchange := range f.exchanges {
		wait.Add(1)
		go func(exchange Exchange) {
			defer wait.Done()
			exchange.FetchTradeHistory()
		}(exchange)
	}
	wait.Wait()
}
