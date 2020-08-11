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
	"github.com/KyberNetwork/reserve-data/settings"
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
	setting                Setting
	l                      *zap.SugaredLogger
}

func NewFetcher(
	storage Storage,
	globalStorage GlobalStorage,
	theworld TheWorld,
	runner Runner,
	simulationMode bool, setting Setting) *Fetcher {
	return &Fetcher{
		storage:        storage,
		globalStorage:  globalStorage,
		exchanges:      []Exchange{},
		blockchain:     nil,
		theworld:       theworld,
		runner:         runner,
		simulationMode: simulationMode,
		setting:        setting,
		l:              zap.S(),
	}
}

func (f *Fetcher) SetBlockchain(blockchain Blockchain) {
	f.blockchain = blockchain
	f.FetchCurrentBlock(common.GetTimepoint())
}

func (f *Fetcher) AddExchange(exchange Exchange) {
	f.exchanges = append(f.exchanges, exchange)
	// initiate exchange status as up
	exchangeStatus, _ := f.setting.GetExchangeStatus()
	if exchangeStatus == nil {
		exchangeStatus = map[string]common.ExStatus{}
	}
	exchangeID := string(exchange.ID())
	_, exist := exchangeStatus[exchangeID]
	if !exist {
		exchangeStatus[exchangeID] = common.ExStatus{
			Timestamp: common.GetTimepoint(),
			Status:    true,
		}
	}
	if err := f.setting.UpdateExchangeStatus(exchangeStatus); err != nil {
		f.l.Warnw("Update exchange status", "err", err)
	}
}

func (f *Fetcher) Stop() error {
	return f.runner.Stop()
}

func (f *Fetcher) Run() error {
	f.l.Infof("Fetcher runner is starting...")
	if err := f.runner.Start(); err != nil {
		return err
	}
	go f.RunOrderbookFetcher()
	go f.RunAuthDataFetcher()
	go f.RunRateFetcher()
	go f.RunBlockFetcher()
	go f.RunGlobalDataFetcher()
	f.l.Infof("Fetcher runner is running...")
	return nil
}

func (f *Fetcher) RunGlobalDataFetcher() {
	for {
		f.l.Infof("waiting for signal from global data channel")
		t := <-f.runner.GetGlobalDataTicker()
		f.l.Infof("got signal in global data channel with timestamp %d", common.TimeToTimepoint(t))
		timepoint := common.TimeToTimepoint(t)
		f.FetchGlobalData(timepoint)
		f.l.Infof("fetched block from blockchain")
	}
}

func (f *Fetcher) FetchGlobalData(timepoint uint64) {
	if goldData, err := f.theworld.GetGoldInfo(); err != nil {
		f.l.Warnw("failed to fetch Gold Info", "err", err)
	} else {
		goldData.Timestamp = common.GetTimepoint()
		if err = f.globalStorage.StoreGoldInfo(goldData); err != nil {
			f.l.Warnw("Storing gold info failed", "err", err)
		}
	}

	if btcData, err := f.theworld.GetBTCInfo(); err != nil {
		f.l.Warnw("failed to fetch BTC Info", "err", err)
	} else {
		btcData.Timestamp = common.GetTimepoint()
		if err = f.globalStorage.StoreBTCInfo(btcData); err != nil {
			f.l.Warnw("Storing BTC info failed", "err", err)
		}
	}

	if usdData, err := f.theworld.GetUSDInfo(); err != nil {
		f.l.Warnw("failed to fetch USD info", "err", err)
	} else {
		usdData.Timestamp = common.GetTimepoint()
		if err = f.globalStorage.StoreUSDInfo(usdData); err != nil {
			f.l.Warnw("Store USD info failed", "err", err)
		}
	}
}

func (f *Fetcher) RunBlockFetcher() {
	for {
		f.l.Infof("waiting for signal from block channel")
		t := <-f.runner.GetBlockTicker()
		f.l.Infof("got signal in block channel with timestamp %d", common.TimeToTimepoint(t))
		timepoint := common.TimeToTimepoint(t)
		f.FetchCurrentBlock(timepoint)
		f.l.Infof("fetched block from blockchain")
	}
}

func (f *Fetcher) RunRateFetcher() {
	for {
		f.l.Infof("waiting for signal from runner rate channel")
		t := <-f.runner.GetRateTicker()
		f.l.Infof("got signal in rate channel with timestamp %d", common.TimeToTimepoint(t))
		f.FetchRate(common.TimeToTimepoint(t))
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
		f.l.Warnw("Fetching rates from blockchain failed", "err", err, "at_block", atBlock, "current_block", f.currentBlock)
		return
	}

	f.l.Infof("Got rates from blockchain: %+v", data)
	if err = f.storage.StoreRate(data, timepoint); err != nil {
		f.l.Warnw("Storing rates failed", "err", err)
	}
}

func (f *Fetcher) RunAuthDataFetcher() {
	for {
		f.l.Infof("waiting for signal from runner auth data channel")
		t := <-f.runner.GetAuthDataTicker()
		f.l.Infof("got signal in auth data channel with timestamp %d", common.TimeToTimepoint(t))
		f.FetchAllAuthData(common.TimeToTimepoint(t))
		f.l.Infof("fetched data from exchanges")
	}
}

func (f *Fetcher) FetchAllAuthData(timepoint uint64) {
	f.l.Infof("start to fetch auth data")
	snapshot := common.AuthDataSnapshot{
		Valid:             true,
		Timestamp:         common.GetTimestamp(),
		ExchangeBalances:  map[common.ExchangeID]common.EBalanceEntry{},
		ReserveBalances:   map[string]common.BalanceEntry{},
		PendingActivities: []common.ActivityRecord{},
		Block:             0,
	}
	bbalances := map[string]common.BalanceEntry{}
	ebalances := sync.Map{}
	estatuses := sync.Map{}
	bstatuses := sync.Map{}
	pendings, err := f.storage.GetPendingActivities()
	if err != nil {
		f.l.Errorw("AuthData - getting pending activities failed", "err", err)
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
			txID := activity.Result[common.ResultTx]
			if txID == nil {
				continue
			}
			//Set activity result tx to tx from cexs if currently result tx is not nil an is an empty string
			resultTx, ok := txID.(string)
			if !ok {
				f.l.Warnw("Activity Result Tx cannot be asserted to string", common.ResultTx, fmt.Sprintf("%+v", txID))
				continue
			}
			if resultTx == "" {
				activity.Result[common.ResultTx] = activityStatus.Tx
			}
		}
	}

	if err = f.FetchAuthDataFromBlockchain(bbalances, &bstatuses, pendings); err != nil {
		snapshot.Error = err.Error()
		snapshot.Valid = false
		f.l.Errorw("AuthData - fetch from blockchain failed", "err", err)
	}
	snapshot.Block = f.currentBlock
	snapshot.ReturnTime = common.GetTimestamp()
	err = f.PersistSnapshot(
		&ebalances, bbalances, &estatuses, &bstatuses,
		pendings, &snapshot, timepoint)
	if err != nil {
		f.l.Errorw("AuthData - storing auth data failed", "err", err)
		return
	}
}

func (f *Fetcher) FetchAuthDataFromBlockchain(
	allBalances map[string]common.BalanceEntry,
	allStatuses *sync.Map,
	pendings []common.ActivityRecord) error {
	// we apply double check strategy to mitigate race condition on exchange side like this:
	// 1. Get list of pending activity status (A)
	// 2. Get list of balances (B)
	// 3. Get list of pending activity status again (C)
	// 4. if C != A, repeat 1, otherwise return A, B
	var balances map[string]common.BalanceEntry
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
		f.l.Infow("preStatuses and statuses are not match exactly, retrying")
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
		f.l.Warnw("Fetching current block failed", "err", err)
	} else {
		// update currentBlockUpdateTime first to avoid race condition
		// where fetcher is trying to fetch new rate
		f.currentBlockUpdateTime = common.GetTimepoint()
		f.currentBlock = block
	}
}

func (f *Fetcher) FetchBalanceFromBlockchain() (map[string]common.BalanceEntry, error) {
	reserveAddr, err := f.setting.GetAddress(settings.Reserve)
	if err != nil {
		return nil, err
	}
	return f.blockchain.FetchBalanceData(reserveAddr, 0)
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

		actNonce, ok := act.Result["nonce"].(string)
		// interface assertion also return false if actNonce is nil
		if !ok {
			return false
		}
		nonce, err := strconv.ParseUint(actNonce, 10, 64)
		if err != nil {
			f.l.Warnw("convert act.Result[nonce] to Uint64 failed", "err", err)
			return false
		}
		return nonce < minedNonce
	}
}

func (f *Fetcher) FetchStatusFromBlockchain(pendings []common.ActivityRecord) (map[common.ActivityID]common.ActivityStatus, error) {
	result := map[common.ActivityID]common.ActivityStatus{}
	nonceValidator := f.newNonceValidator()

	for _, activity := range pendings {
		if activity.IsBlockchainPending() && (activity.Action == common.ActionSetRate || activity.Action == common.ActionDeposit || activity.Action == common.ActionWithdraw) {
			var (
				blockNum uint64
				status   string
				err      error
			)
			txID := activity.Result[common.ResultTx]
			if txID == nil {
				continue
			}
			txStr, ok := txID.(string)
			if !ok {
				f.l.Warnw("TX_STATUS:cannot convert activity.Result[tx] to string type", common.ResultTx, fmt.Sprintf("%+v", txID))
				continue
			}
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
					0, // default withdraw fee
					err,
				)
			case common.MiningStatusFailed:
				result[activity.ID] = common.NewActivityStatus(
					activity.ExchangeStatus,
					txStr,
					blockNum,
					common.MiningStatusFailed,
					0, // default withdraw fee
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
					elapsed := common.GetTimepoint() - activity.Timestamp.MustToUint64()
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
						0, // default withdraw fee
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
		f.l.Warnw("ERROR: status cannot be asserted to common.ActivityStatus", "status", status)
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
		activity.Result[common.ResultStatusError] = activityStatus.Error.Error()
	} else {
		activity.Result[common.ResultStatusError] = ""
	}
	activity.Result[common.ResultBlockNumber] = activityStatus.BlockNumber
}

func (f *Fetcher) updateActivitywithExchangeStatus(activity *common.ActivityRecord, estatuses *sync.Map, snapshot *common.AuthDataSnapshot) {
	status, ok := estatuses.Load(activity.ID)
	if !ok || status == nil {
		f.l.Infow("exchange status for %s is nil or not existed ", "activity", activity.ID.String())
		return
	}
	activityStatus, ok := status.(common.ActivityStatus)
	if !ok {
		f.l.Warnw("status cannot be asserted to common.ActivityStatus", "status", status)
		return
	}
	f.l.Infof("In PersistSnapshot: exchange activity status for %+v: %+v", activity.ID, activityStatus)
	if activity.IsExchangePending() {
		activity.ExchangeStatus = activityStatus.ExchangeStatus
		if activity.Action == common.ActionWithdraw {
			activity.Result[common.WithdrawFee] = activityStatus.Fee
		}
	} else if activityStatus.ExchangeStatus == common.ExchangeStatusFailed {
		activity.ExchangeStatus = activityStatus.ExchangeStatus
		if activity.Action == common.ActionWithdraw {
			activity.Result[common.WithdrawFee] = activityStatus.Fee
		}
	}

	if resultTx, ok := activity.Result[common.ResultTx].(string); ok && resultTx == "" {
		activity.Result[common.ResultTx] = activityStatus.Tx
	}

	if activityStatus.Error != nil {
		snapshot.Valid = false
		snapshot.Error = activityStatus.Error.Error()
		activity.Result[common.ResultStatusError] = activityStatus.Error.Error()
		if activity.Action == common.ActionWithdraw {
			activity.Result[common.WithdrawFee] = activityStatus.Fee
		}
	} else {
		activity.Result[common.ResultStatusError] = ""
		if activity.Action == common.ActionWithdraw {
			activity.Result[common.WithdrawFee] = activityStatus.Fee
		}
	}
}

func (f *Fetcher) PersistSnapshot(
	ebalances *sync.Map,
	bbalances map[string]common.BalanceEntry,
	estatuses *sync.Map,
	bstatuses *sync.Map,
	pendings []common.ActivityRecord,
	snapshot *common.AuthDataSnapshot,
	timepoint uint64) error {

	allEBalances := map[common.ExchangeID]common.EBalanceEntry{}
	ebalances.Range(func(key, value interface{}) bool {
		//if type conversion went wrong, continue to the next record
		v, ok := value.(common.EBalanceEntry)
		if !ok {
			f.l.Warnw("value cannot be asserted to common.EbalanceEntry", "value", v)
			return true
		}
		exID, ok := key.(common.ExchangeID)
		if !ok {
			f.l.Warnw("key cannot be asserted to common.ExchangeID", "key", key)
			return true
		}
		allEBalances[exID] = v
		if !v.Valid {
			// get old auth data, because get balance error then we have to keep
			// balance to the latest version then analytic won't get exchange balance to zero
			authVersion, err := f.storage.CurrentAuthDataVersion(common.GetTimepoint())
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
	f.l.Infof("save auth data with timepoint %d", timepoint)
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
	var tokenAddress map[string]ethereum.Address
	startCheck := time.Now()
	for {
		preStatuses := f.FetchStatusFromExchange(exchange, pendings, timepoint)
		balances, err = exchange.FetchEBalanceData(timepoint)
		if err != nil {
			f.l.Errorw("AuthData - fetching exchange balances failed", "exchange", exchange.Name(), "err", err)
			break
		}
		//Remove all token which is not in this exchange's token addresses
		tokenAddress, err = exchange.TokenAddresses()
		if err != nil {
			f.l.Errorw("AuthData - getting token address failed", "exchange", exchange.Name(), "err", err)
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
	dur := time.Since(startCheck)
	checkThreshold := 30.0
	if dur.Seconds() > checkThreshold {
		f.l.Errorw("AuthData - fetch status from blockchain", "duration", dur.String())
	}
	if err == nil {
		allBalances.Store(exchange.ID(), balances)
		for id, activityStatus := range statuses {
			allStatuses.Store(id, activityStatus)
		}
	}
}

func (f *Fetcher) FetchStatusFromExchange(exchange Exchange, pendings []common.ActivityRecord, timepoint uint64) map[common.ActivityID]common.ActivityStatus {
	result := map[common.ActivityID]common.ActivityStatus{}
	for _, activity := range pendings {
		if activity.IsExchangePending() && activity.Destination == string(exchange.ID()) {
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
				base, ok := activity.Params[common.ParamBase].(string)
				if !ok {
					f.l.Warnw("activity Params base can't be converted to type string", common.ParamBase, activity.Params[common.ParamBase])
					continue
				}
				quote, ok := activity.Params[common.ParamQuote].(string)
				if !ok {
					f.l.Warnw("activity Params quote can't be converted to type string", common.ParamQuote, activity.Params[common.ParamQuote])
					continue
				}
				// we ignore error of order status because it doesn't affect
				// authdata. Analytic will ignore order status anyway.
				status, _ = exchange.OrderStatus(orderID, base, quote)
			case common.ActionDeposit:
				txID := activity.Result[common.ResultTx]
				if txID == nil {
					continue
				}
				txHash, ok := txID.(string)
				if !ok {
					f.l.Warnw("activity Result tx can't be converted to type string", common.ResultTx, fmt.Sprintf("%+v", txID))
					continue
				}
				amountStr, ok := activity.Params[common.ParamAmount].(string)
				if !ok {
					f.l.Warnw("activity Params amount can't be converted to type string", common.ParamAmount, activity.Params[common.ParamAmount])
					continue
				}
				amount, uErr := strconv.ParseFloat(amountStr, 64)
				if uErr != nil {
					f.l.Warnw("can't parse activity Params amount to float64", common.ParamAmount, amountStr)
					continue
				}
				currency, ok := activity.Params[common.ParamToken].(string)
				if !ok {
					f.l.Warnw("activity Params token can't be converted to type string", common.ParamToken, activity.Params[common.ParamToken])
					continue
				}
				status, err = exchange.DepositStatus(id, txHash, currency, amount, timepoint)
				if err == nil {
					f.l.Infof("Got deposit status for %v: (%s), error(%s)", activity, status, common.ErrorToString(err))
				} else {
					f.l.Errorf("Got deposit status for %v: (%s), error(%s)", activity, status, common.ErrorToString(err))
				}
			case common.ActionWithdraw:
				amountStr, ok := activity.Params[common.ParamAmount].(string)
				if !ok {
					f.l.Warnw("activity Params amount can't be converted to type string", common.ParamAmount, activity.Params[common.ParamAmount])
					continue
				}
				amount, uErr := strconv.ParseFloat(amountStr, 64)
				if uErr != nil {
					f.l.Warnw("can't parse activity Params amount to float64", common.ParamAmount, amountStr)
					continue
				}
				currency, ok := activity.Params[common.ParamToken].(string)
				if !ok {
					f.l.Warnw("activity Params token can't be converted to type string", common.ParamToken, activity.Params[common.ParamToken])
					continue
				}
				txID := activity.Result[common.ResultTx]
				if txID == nil {
					continue
				}
				if _, ok = txID.(string); !ok {
					f.l.Warnw("activity Result tx can't be converted to type string", common.ResultTx, fmt.Sprintf("%+v", txID))
					continue
				}
				status, tx, fee, err = exchange.WithdrawStatus(id.EID, currency, amount, timepoint)
				if err == nil {
					f.l.Infof("Got withdraw status for %v: (%s), error(%s)", activity, status, common.ErrorToString(err))
				} else {
					f.l.Warnf("Got withdraw status for %v: (%s), error(%s)", activity, status, common.ErrorToString(err))
				}

			default:
				continue
			}

			// in case there is something wrong with the cex and the activity is stuck for a very
			// long time. We will just consider it as a failed activity.
			timepoint, err1 := strconv.ParseUint(string(activity.Timestamp), 10, 64)
			if err1 != nil {
				f.l.Warnw("Activity has invalid timestamp. Just ignore it.", "activity", activity,
					"err", err1, "timestamp", activity.Timestamp)
			} else {
				if common.GetTimepoint()-timepoint > maxActivityLifeTime*uint64(time.Hour)/uint64(time.Millisecond) {
					result[id] = common.NewActivityStatus(common.ExchangeStatusFailed, tx, blockNum, activity.MiningStatus, fee, err)
				} else {
					result[id] = common.NewActivityStatus(status, tx, blockNum, activity.MiningStatus, fee, err)
				}
			}
		} else {
			f.l.Warnw("Activity has strange status", "is exchange pending", activity.IsExchangePending(), "exchange status", activity.ExchangeStatus)
		}
	}
	return result
}

func (f *Fetcher) RunOrderbookFetcher() {
	for {
		f.l.Infof("waiting for signal from runner orderbook channel")
		t := <-f.runner.GetOrderbookTicker()
		f.l.Infof("got signal in orderbook channel with timestamp %d", common.TimeToTimepoint(t))
		f.FetchOrderbook(common.TimeToTimepoint(t))
		f.l.Infof("fetched data from exchanges")
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
	fetchConfig, err := f.globalStorage.GetAllFetcherConfiguration()
	if err != nil {
		f.l.Warnw("cannot get btc fetcher configuration", "err", err)
		return
	}
	exdata, err := exchange.FetchPriceData(timepoint, fetchConfig.BTC)
	if err != nil {
		f.l.Warnw("Fetching data failed", "exchange", exchange.Name(), "err", err)
		return
	}
	for pair, exchangeData := range exdata {
		data.SetOnePrice(exchange.ID(), pair, exchangeData)
	}
}
