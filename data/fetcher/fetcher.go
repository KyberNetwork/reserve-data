package fetcher

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/settings"
	ethereum "github.com/ethereum/go-ethereum/common"
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
	runner                 FetcherRunner
	currentBlock           uint64
	currentBlockUpdateTime uint64
	simulationMode         bool
	setting                Setting
}

func NewFetcher(
	storage Storage,
	globalStorage GlobalStorage,
	theworld TheWorld,
	runner FetcherRunner,
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
	}
}

func (self *Fetcher) SetBlockchain(blockchain Blockchain) {
	self.blockchain = blockchain
	self.FetchCurrentBlock(common.GetTimepoint())
}

func (self *Fetcher) AddExchange(exchange Exchange) {
	self.exchanges = append(self.exchanges, exchange)
	// initiate exchange status as up
	exchangeStatus, _ := self.setting.GetExchangeStatus()
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
	if err := self.setting.UpdateExchangeStatus(exchangeStatus); err != nil {
		log.Printf("Update exchange status error: %s", err.Error())
	}
}

func (self *Fetcher) Stop() error {
	return self.runner.Stop()
}

func (self *Fetcher) Run() error {
	log.Printf("Fetcher runner is starting...")
	if err := self.runner.Start(); err != nil {
		return err
	}
	go self.RunOrderbookFetcher()
	go self.RunAuthDataFetcher()
	go self.RunRateFetcher()
	go self.RunBlockFetcher()
	go self.RunGlobalDataFetcher()
	log.Printf("Fetcher runner is running...")
	return nil
}

func (self *Fetcher) RunGlobalDataFetcher() {
	for {
		log.Printf("waiting for signal from global data channel")
		t := <-self.runner.GetGlobalDataTicker()
		log.Printf("got signal in global data channel with timestamp %d", common.TimeToTimepoint(t))
		timepoint := common.TimeToTimepoint(t)
		self.FetchGlobalData(timepoint)
		log.Printf("fetched block from blockchain")
	}
}

func (self *Fetcher) FetchGlobalData(timepoint uint64) {
	goldData, err := self.theworld.GetGoldInfo()
	if err != nil {
		log.Printf("failed to fetch Gold Info: %s", err.Error())
		return
	}
	goldData.Timestamp = common.GetTimepoint()

	if err = self.globalStorage.StoreGoldInfo(goldData); err != nil {
		log.Printf("Storing gold info failed: %s", err.Error())
	}

	btcData, err := self.theworld.GetBTCInfo()
	if err != nil {
		log.Printf("failed to fetch BTC Info: %s", err.Error())
		return
	}
	btcData.Timestamp = common.GetTimepoint()
	if err = self.globalStorage.StoreBTCInfo(btcData); err != nil {
		log.Printf("Storing BTC info failed: %s", err.Error())
	}
}

func (self *Fetcher) RunBlockFetcher() {
	for {
		log.Printf("waiting for signal from block channel")
		t := <-self.runner.GetBlockTicker()
		log.Printf("got signal in block channel with timestamp %d", common.TimeToTimepoint(t))
		timepoint := common.TimeToTimepoint(t)
		self.FetchCurrentBlock(timepoint)
		log.Printf("fetched block from blockchain")
	}
}

func (self *Fetcher) RunRateFetcher() {
	for {
		log.Printf("waiting for signal from runner rate channel")
		t := <-self.runner.GetRateTicker()
		log.Printf("got signal in rate channel with timestamp %d", common.TimeToTimepoint(t))
		self.FetchRate(common.TimeToTimepoint(t))
		log.Printf("fetched rates from blockchain")
	}
}

func (self *Fetcher) FetchRate(timepoint uint64) {
	var (
		err  error
		data common.AllRateEntry
	)
	// only fetch rates 5s after the block number is updated
	if !self.simulationMode && self.currentBlockUpdateTime-timepoint <= 5000 {
		return
	}

	var atBlock = self.currentBlock - 1
	// in simulation mode, just fetches from latest known block
	if self.simulationMode {
		atBlock = 0
	}

	data, err = self.blockchain.FetchRates(atBlock, self.currentBlock)
	if err != nil {
		log.Printf("Fetching rates from blockchain failed: %s. Will not store it to storage.", err.Error())
		return
	}

	log.Printf("Got rates from blockchain: %+v", data)
	if err = self.storage.StoreRate(data, timepoint); err != nil {
		log.Printf("Storing rates failed: %s", err.Error())
	}
}

func (self *Fetcher) RunAuthDataFetcher() {
	for {
		log.Printf("waiting for signal from runner auth data channel")
		t := <-self.runner.GetAuthDataTicker()
		log.Printf("got signal in auth data channel with timestamp %d", common.TimeToTimepoint(t))
		self.FetchAllAuthData(common.TimeToTimepoint(t))
		log.Printf("fetched data from exchanges")
	}
}

func (self *Fetcher) FetchAllAuthData(timepoint uint64) {
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
	pendings, err := self.storage.GetPendingActivities()
	if err != nil {
		log.Printf("Getting pending activites failed: %s\n", err)
		return
	}
	wait := sync.WaitGroup{}
	for _, exchange := range self.exchanges {
		wait.Add(1)
		go self.FetchAuthDataFromExchange(
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
				log.Print("WARNING: status from cexs cannot be asserted to common.ActivityStatus")
				continue
			}
			//Set activity result tx to tx from cexs if currently result tx is not nil an is an empty string
			resultTx, ok := activity.Result["tx"].(string)
			if !ok {
				log.Printf("WARNING: Activity Result Tx (value %v) cannot be asserted to string", activity.Result["tx"])
				continue
			}
			if resultTx == "" {
				activity.Result["tx"] = activityStatus.Tx
			}
		}
	}

	self.FetchAuthDataFromBlockchain(
		bbalances, &bstatuses, pendings)
	snapshot.Block = self.currentBlock
	snapshot.ReturnTime = common.GetTimestamp()
	err = self.PersistSnapshot(
		&ebalances, bbalances, &estatuses, &bstatuses,
		pendings, &snapshot, timepoint)
	if err != nil {
		log.Printf("Storing exchange balances failed: %s\n", err)
		return
	}
}

func (self *Fetcher) FetchAuthDataFromBlockchain(
	allBalances map[string]common.BalanceEntry,
	allStatuses *sync.Map,
	pendings []common.ActivityRecord) {
	// we apply double check strategy to mitigate race condition on exchange side like this:
	// 1. Get list of pending activity status (A)
	// 2. Get list of balances (B)
	// 3. Get list of pending activity status again (C)
	// 4. if C != A, repeat 1, otherwise return A, B
	var balances map[string]common.BalanceEntry
	var preStatuses, statuses map[common.ActivityID]common.ActivityStatus
	var err error
	for {
		preStatuses, err = self.FetchStatusFromBlockchain(pendings)
		if err != nil {
			log.Printf("Fetching blockchain status failed:  %v", err)
			break
		}
		balances, err = self.FetchBalanceFromBlockchain()
		if err != nil {
			log.Printf("Fetching blockchain balances failed: %v", err)
			break
		}
		statuses, err = self.FetchStatusFromBlockchain(pendings)
		if err != nil {
			log.Printf("Fetching blockchain status failed:  %v", err)
			break
		}
		if unchanged(preStatuses, statuses) {
			break
		}
	}
	if err == nil {
		for k, v := range balances {
			allBalances[k] = v
		}
		for id, activityStatus := range statuses {
			allStatuses.Store(id, activityStatus)
		}
	}
}

func (self *Fetcher) FetchCurrentBlock(timepoint uint64) {
	block, err := self.blockchain.CurrentBlock()
	if err != nil {
		log.Printf("Fetching current block failed: %v. Ignored.", err)
	} else {
		// update currentBlockUpdateTime first to avoid race condition
		// where fetcher is trying to fetch new rate
		self.currentBlockUpdateTime = common.GetTimepoint()
		self.currentBlock = block
	}
}

func (self *Fetcher) FetchBalanceFromBlockchain() (map[string]common.BalanceEntry, error) {
	reserveAddr, err := self.setting.GetAddress(settings.Reserve)
	if err != nil {
		return nil, err
	}
	return self.blockchain.FetchBalanceData(reserveAddr, 0)
}

func (self *Fetcher) newNonceValidator() func(common.ActivityRecord) bool {
	// SetRateMinedNonce might be slow, use closure to not invoke it every time
	minedNonce, err := self.blockchain.SetRateMinedNonce()
	if err != nil {
		log.Printf("Getting mined nonce failed: %s", err)
	}

	return func(act common.ActivityRecord) bool {
		// this check only works with set rate transaction as:
		//   - account nonce is record in result field of activity
		//   - the SetRateMinedNonce method is available
		if act.Action != common.ActionSetrate {
			return false
		}

		actNonce, ok := act.Result["nonce"].(string)
		// interface assertion also return false if actNonce is nil
		if !ok {
			return false
		}
		nonce, err := strconv.ParseUint(actNonce, 10, 64)
		if err != nil {
			log.Printf("ERROR convert act.Result[nonce] to Uint64 failed %s", err.Error())
			return false
		}
		return nonce < minedNonce
	}
}

func (self *Fetcher) FetchStatusFromBlockchain(pendings []common.ActivityRecord) (map[common.ActivityID]common.ActivityStatus, error) {
	result := map[common.ActivityID]common.ActivityStatus{}
	nonceValidator := self.newNonceValidator()

	for _, activity := range pendings {
		if activity.IsBlockchainPending() && (activity.Action == common.ActionSetrate || activity.Action == common.ActionDeposit || activity.Action == common.ActionWithdraw) {
			var blockNum uint64
			var status string
			var err error
			txStr, ok := activity.Result["tx"].(string)
			if !ok {
				log.Printf("TX_STATUS: WARNING cannot convert activity.Result[tx] (value %v) to string type", activity.Result["tx"])
				continue
			}
			tx := ethereum.HexToHash(txStr)
			if tx.Big().IsInt64() && tx.Big().Int64() == 0 {
				continue
			}
			status, blockNum, err = self.blockchain.TxStatus(tx)
			if err != nil {
				return result, fmt.Errorf("TX_STATUS: ERROR Getting tx status failed: %s", err)
			}

			switch status {
			case common.MiningStatusPending:
				log.Printf("TX_STATUS: tx (%s) status is pending", tx)
				if nonceValidator(activity) {
					result[activity.ID] = common.NewActivityStatus(
						activity.ExchangeStatus,
						txStr,
						blockNum,
						common.MiningStatusFailed,
						err,
					)
				}
			case common.MiningStatusMined:
				if activity.Action == common.ActionSetrate {
					log.Printf("TX_STATUS set rate transaction is mined, id: %s", activity.ID.EID)
				}
				result[activity.ID] = common.NewActivityStatus(
					activity.ExchangeStatus,
					txStr,
					blockNum,
					common.MiningStatusMined,
					err,
				)
			case common.MiningStatusFailed:
				result[activity.ID] = common.NewActivityStatus(
					activity.ExchangeStatus,
					txStr,
					blockNum,
					common.MiningStatusFailed,
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
						log.Printf("TX_STATUS: tx(%s) is lost, elapsed time: %d", txStr, elapsed)
						txFailed = true
					}
				}

				if txFailed {
					result[activity.ID] = common.NewActivityStatus(
						activity.ExchangeStatus,
						txStr,
						blockNum,
						common.MiningStatusFailed,
						err,
					)
				}
			default:
				log.Printf("TX_STATUS: tx (%s) status is not available. Wait till next try", tx)
			}
		}
	}
	return result, nil
}

func unchanged(pre, post map[common.ActivityID]common.ActivityStatus) bool {
	if len(pre) != len(post) {
		return false
	} else {
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
	}
	return true
}

func updateActivitywithBlockchainStatus(activity *common.ActivityRecord, bstatuses *sync.Map, snapshot *common.AuthDataSnapshot) {
	status, ok := bstatuses.Load(activity.ID)
	if !ok || status == nil {
		log.Printf("block chain status for %s is nil or not existed ", activity.ID.String())
		return
	}

	activityStatus, ok := status.(common.ActivityStatus)
	if !ok {
		log.Printf("ERROR: status (%v) cannot be asserted to common.ActivityStatus", status)
		return
	}
	log.Printf("In PersistSnapshot: blockchain activity status for %+v: %+v", activity.ID, activityStatus)
	if activity.IsBlockchainPending() {
		activity.MiningStatus = activityStatus.MiningStatus
	}

	if activityStatus.ExchangeStatus == common.ExchangeStatusFailed {
		activity.ExchangeStatus = activityStatus.ExchangeStatus
	}

	if activityStatus.Error != nil {
		snapshot.Valid = false
		snapshot.Error = activityStatus.Error.Error()
		activity.Result["status_error"] = activityStatus.Error.Error()
	} else {
		activity.Result["status_error"] = ""
	}
	activity.Result["blockNumber"] = activityStatus.BlockNumber
}

func updateActivitywithExchangeStatus(activity *common.ActivityRecord, estatuses *sync.Map, snapshot *common.AuthDataSnapshot) {
	status, ok := estatuses.Load(activity.ID)
	if !ok || status == nil {
		log.Printf("exchange status for %s is nil or not existed ", activity.ID.String())
		return
	}
	activityStatus, ok := status.(common.ActivityStatus)
	if !ok {
		log.Printf("ERROR: status (%v) cannot be asserted to common.ActivityStatus", status)
		return
	}
	log.Printf("In PersistSnapshot: exchange activity status for %+v: %+v", activity.ID, activityStatus)
	if activity.IsExchangePending() {
		activity.ExchangeStatus = activityStatus.ExchangeStatus
	} else {
		if activityStatus.ExchangeStatus == common.ExchangeStatusFailed {
			activity.ExchangeStatus = activityStatus.ExchangeStatus
		}
	}
	resultTx, ok := activity.Result["tx"].(string)
	if !ok {
		log.Printf("WARNING: activity.Result[tx] (value %v) cannot be asserted to string type", activity.Result["tx"])
	} else if ok && resultTx == "" {
		activity.Result["tx"] = activityStatus.Tx
	}

	if activityStatus.Error != nil {
		snapshot.Valid = false
		snapshot.Error = activityStatus.Error.Error()
		activity.Result["status_error"] = activityStatus.Error.Error()
	} else {
		activity.Result["status_error"] = ""
	}
}

func (self *Fetcher) PersistSnapshot(
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
			log.Printf("ERROR: value (%v) cannot be asserted to common.EbalanceEntry", v)
			return true
		}
		exID, ok := key.(common.ExchangeID)
		if !ok {
			log.Printf("ERROR: key (%v) cannot be asserted to common.ExchangeID", key)
			return true
		}
		allEBalances[exID] = v
		if !v.Valid {
			// get old auth data, because get balance error then we have to keep
			// balance to the latest version then analytic won't get exchange balance to zero
			authVersion, err := self.storage.CurrentAuthDataVersion(common.GetTimepoint())
			if err == nil {
				oldAuth, err := self.storage.GetAuthData(authVersion)
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
		updateActivitywithExchangeStatus(&activity, estatuses, snapshot)
		updateActivitywithBlockchainStatus(&activity, bstatuses, snapshot)
		log.Printf("Aggregate statuses, final activity: %+v", activity)
		if activity.IsPending() {
			pendingActivities = append(pendingActivities, activity)
		}
		err := self.storage.UpdateActivity(activity.ID, activity)
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
	return self.storage.StoreAuthSnapshot(snapshot, timepoint)
}

func (self *Fetcher) FetchAuthDataFromExchange(
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
	for {
		preStatuses := self.FetchStatusFromExchange(exchange, pendings, timepoint)
		balances, err = exchange.FetchEBalanceData(timepoint)
		if err != nil {
			log.Printf("Fetching exchange balances from %s failed: %v\n", exchange.Name(), err)
			break
		}
		statuses = self.FetchStatusFromExchange(exchange, pendings, timepoint)
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

func (self *Fetcher) FetchStatusFromExchange(exchange Exchange, pendings []common.ActivityRecord, timepoint uint64) map[common.ActivityID]common.ActivityStatus {
	result := map[common.ActivityID]common.ActivityStatus{}
	for _, activity := range pendings {
		if activity.IsExchangePending() && activity.Destination == string(exchange.ID()) {
			var err error
			var status string
			var tx string
			var blockNum uint64

			id := activity.ID
			//These type conversion errors can be ignore since if happens, it will be reflected in activity.error
			if activity.Action == common.ActionTrade {
				orderID := id.EID
				base, ok := activity.Params["base"].(string)
				if !ok {
					log.Printf("WARNING: activity Params base (%v) can't be converted to type string", activity.Params["base"])
					continue
				}
				quote, ok := activity.Params["quote"].(string)
				if !ok {
					log.Printf("WARNING: activity Params quote (%v) can't be converted to type string", activity.Params["quote"])
					continue
				}
				// we ignore error of order status because it doesn't affect
				// authdata. Analytic will ignore order status anyway.
				status, _ = exchange.OrderStatus(orderID, base, quote)
			} else if activity.Action == common.ActionDeposit {
				txHash, ok := activity.Result["tx"].(string)
				if !ok {
					log.Printf("WARNING: activity Result tx (%v) can't be converted to type string", activity.Result["tx"])
					continue
				}
				amountStr, ok := activity.Params["amount"].(string)
				if !ok {
					log.Printf("WARNING: activity Params amount (%v) can't be converted to type string", activity.Params["amount"])
					continue
				}
				amount, uErr := strconv.ParseFloat(amountStr, 64)
				if uErr != nil {
					log.Printf("WARNING: can't parse activity Params amount %s to float64", amountStr)
					continue
				}
				currency, ok := activity.Params["token"].(string)
				if !ok {
					log.Printf("WARNING: activity Params token (%v) can't be converted to type string", activity.Params["token"])
					continue
				}
				status, err = exchange.DepositStatus(id, txHash, currency, amount, timepoint)
				log.Printf("Got deposit status for %v: (%s), error(%s)", activity, status, common.ErrorToString(err))
			} else if activity.Action == common.ActionWithdraw {
				amountStr, ok := activity.Params["amount"].(string)
				if !ok {
					log.Printf("WARNING: activity Params amount (%v) can't be converted to type string", activity.Params["amount"])
					continue
				}
				amount, uErr := strconv.ParseFloat(amountStr, 64)
				if uErr != nil {
					log.Printf("WARNING: can't parse activity Params amount %s to float64", amountStr)
					continue
				}
				currency, ok := activity.Params["token"].(string)
				if !ok {
					log.Printf("WARNING: activity Params token (%v) can't be converted to type string", activity.Params["token"])
					continue
				}
				tx, ok = activity.Result["tx"].(string)
				if !ok {
					log.Printf("WARNING: activity Result tx (%v) can't be converted to type string", activity.Result["tx"])
					continue
				}
				status, tx, err = exchange.WithdrawStatus(id.EID, currency, amount, timepoint)
				log.Printf("Got withdraw status for %v: (%s), error(%s)", activity, status, common.ErrorToString(err))
			} else {
				continue
			}
			// in case there is something wrong with the cex and the activity is stuck for a very
			// long time. We will just consider it as a failed activity.
			timepoint, err1 := strconv.ParseUint(string(activity.Timestamp), 10, 64)
			if err1 != nil {
				log.Printf("Activity %v has invalid timestamp. Just ignore it.", activity)
			} else {
				if common.GetTimepoint()-timepoint > uint64(maxActivityLifeTime*uint64(time.Hour))/uint64(time.Millisecond) {
					result[id] = common.NewActivityStatus(common.ExchangeStatusFailed, tx, blockNum, activity.MiningStatus, err)
				} else {
					result[id] = common.NewActivityStatus(status, tx, blockNum, activity.MiningStatus, err)
				}
			}
		} else {
			timepoint, err1 := strconv.ParseUint(string(activity.Timestamp), 10, 64)
			if err1 != nil {
				log.Printf("Activity %v has invalid timestamp. Just ignore it.", activity)
			} else {
				if activity.Destination == string(exchange.ID()) &&
					activity.ExchangeStatus == common.ExchangeStatusDone &&
					common.GetTimepoint()-timepoint > uint64(maxActivityLifeTime*uint64(time.Hour))/uint64(time.Millisecond) {
					// the activity is still pending but its exchange status is done and it is stuck there for more than
					// maxActivityLifeTime. This activity is considered failed.
					result[activity.ID] = common.NewActivityStatus(common.ExchangeStatusFailed, "", 0, activity.MiningStatus, nil)
				}
			}
		}
	}
	return result
}

func (self *Fetcher) RunOrderbookFetcher() {
	for {
		log.Printf("waiting for signal from runner orderbook channel")
		t := <-self.runner.GetOrderbookTicker()
		log.Printf("got signal in orderbook channel with timestamp %d", common.TimeToTimepoint(t))
		self.FetchOrderbook(common.TimeToTimepoint(t))
		log.Printf("fetched data from exchanges")
	}
}

func (self *Fetcher) FetchOrderbook(timepoint uint64) {
	data := NewConcurrentAllPriceData()
	// start fetching
	wait := sync.WaitGroup{}
	for _, exchange := range self.exchanges {
		wait.Add(1)
		go self.fetchPriceFromExchange(&wait, exchange, data, timepoint)
	}
	wait.Wait()
	data.SetBlockNumber(self.currentBlock)
	err := self.storage.StorePrice(data.GetData(), timepoint)
	if err != nil {
		log.Printf("Storing data failed: %s\n", err)
	}
}

func (self *Fetcher) fetchPriceFromExchange(wg *sync.WaitGroup, exchange Exchange, data *ConcurrentAllPriceData, timepoint uint64) {
	defer wg.Done()
	fetchConfig, err := self.globalStorage.GetAllFetcherConfiguration()
	if err != nil {
		log.Fatal("cannot get btc fetcher configuration")
	}
	exdata, err := exchange.FetchPriceData(timepoint, fetchConfig.BTC)
	if err != nil {
		log.Printf("Fetching data from %s failed: %v\n", exchange.Name(), err)
	}
	for pair, exchangeData := range exdata {
		data.SetOnePrice(exchange.ID(), pair, exchangeData)
	}
}
