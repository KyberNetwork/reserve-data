package fetcher

import (
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/KyberNetwork/reserve-data/common"
	ethereum "github.com/ethereum/go-ethereum/common"
)

type Fetcher struct {
	storage                Storage
	globalStorage          GlobalStorage
	exchanges              []Exchange
	blockchain             Blockchain
	theworld               TheWorld
	runner                 FetcherRunner
	rmaddr                 ethereum.Address
	currentBlock           uint64
	currentBlockUpdateTime uint64
	simulationMode         bool
}

func NewFetcher(
	storage Storage,
	globalStorage GlobalStorage,
	theworld TheWorld,
	runner FetcherRunner,
	address ethereum.Address,
	simulationMode bool) *Fetcher {
	return &Fetcher{
		storage:        storage,
		globalStorage:  globalStorage,
		exchanges:      []Exchange{},
		blockchain:     nil,
		theworld:       theworld,
		runner:         runner,
		rmaddr:         address,
		simulationMode: simulationMode,
	}
}

func (self *Fetcher) SetBlockchain(blockchain Blockchain) {
	self.blockchain = blockchain
	self.FetchCurrentBlock(common.GetTimepoint())
}

func (self *Fetcher) AddExchange(exchange Exchange) {
	self.exchanges = append(self.exchanges, exchange)
	// initiate exchange status as up
	exchangeStatus, _ := self.storage.GetExchangeStatus()
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
	self.storage.UpdateExchangeStatus(exchangeStatus)
}

func (self *Fetcher) Stop() error {
	return self.runner.Stop()
}

func (self *Fetcher) Run() error {
	log.Printf("Fetcher runner is starting...")
	self.runner.Start()
	go self.RunOrderbookFetcher()
	go self.RunAuthDataFetcher()
	go self.RunRateFetcher()
	go self.RunBlockFetcher()
	// go self.RunTradeHistoryFetcher()
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
	data, _ := self.theworld.GetGoldInfo()
	data.Timestamp = common.GetTimepoint()
	err := self.globalStorage.StoreGoldInfo(data)
	if err != nil {
		log.Printf("Storing gold info failed: %s", err.Error())
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
	// only fetch rates 5s after the block number is updated
	if !self.simulationMode && self.currentBlockUpdateTime-timepoint <= 5000 {
		return
	}
	var err error
	var data common.AllRateEntry
	if self.simulationMode {
		data, err = self.blockchain.FetchRates(0, self.currentBlock)
	} else {
		data, err = self.blockchain.FetchRates(self.currentBlock-1, self.currentBlock)
	}
	if err != nil {
		log.Printf("Fetching rates from blockchain failed: %s. Will not store it to storage.", err.Error())
	} else {
		if data.Valid {
			log.Printf("Got rates from blockchain: %+v", data)
			if err = self.storage.StoreRate(data, timepoint); err != nil {
				log.Printf("Storing rates failed: %s", err.Error())
			}
		} else {
			log.Printf("Got invalid rates from blockchain: %s. Will not store it to storage.", data.Error)
		}
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
			activityStatus := status.(common.ActivityStatus)
			if activity.Result["tx"] != nil && activity.Result["tx"].(string) == "" {
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

// func (self *Fetcher) FetchTradeHistoryFromExchange(
// 	wait *sync.WaitGroup,
// 	exchange Exchange,
// 	data *sync.Map,
// 	timepoint uint64) {

// 	defer wait.Done()
// 	tokenPairs := exchange.TokenPairs()
// 	fromIDs := map[string]string{}
// 	for _, pair := range tokenPairs {
// 		id, _ := self.storage.GetLastIDTradeHistory(string(exchange.ID()), fmt.Sprintf("%s-%s", pair.Base.ID, pair.Quote.ID))
// 		fromIDs[pair.Base.ID+pair.Quote.ID] = id
// 	}
// 	tradeHistory, err := exchange.FetchTradeHistory(timepoint, fromIDs)
// 	if err != nil {
// 		log.Printf("Fetch trade history from exchange failed: %s", err.Error())
// 	}
// 	data.Store(exchange.ID(), tradeHistory)
// }

// func (self *Fetcher) FetchAllTradeHistory(timepoint uint64) {
// 	tradeHistory := common.AllTradeHistory{
// 		common.GetTimestamp(),
// 		map[common.ExchangeID]common.ExchangeTradeHistory{},
// 	}
// 	wait := sync.WaitGroup{}
// 	data := sync.Map{}
// 	for _, exchange := range self.exchanges {
// 		wait.Add(1)
// 		go self.FetchTradeHistoryFromExchange(&wait, exchange, &data, timepoint)
// 	}

// 	wait.Wait()
// 	data.Range(func(key, value interface{}) bool {
// 		tradeHistory.Data[key.(common.ExchangeID)] = value.(map[common.TokenPairID][]common.TradeHistory)
// 		return true
// 	})

// 	err := self.storage.StoreTradeHistory(tradeHistory, timepoint)
// 	if err != nil {
// 		log.Printf("Store trade history failed: %s", err.Error())
// 	}
// }

// func (self *Fetcher) RunTradeHistoryFetcher() {
// 	for {
// 		log.Printf("waiting for signal from runner trade history channel")
// 		t := <-self.runner.GetTradeHistoryTicker()
// 		log.Printf("got signal in trade history channel with timestamp %d", common.TimeToTimepoint(t))
// 		self.FetchAllTradeHistory(common.TimeToTimepoint(t))
// 		log.Printf("fetched trade history from exchanges")
// 	}
// }

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
	var statuses map[common.ActivityID]common.ActivityStatus
	var err error
	for {
		preStatuses := self.FetchStatusFromBlockchain(pendings)
		balances, err = self.FetchBalanceFromBlockchain()
		if err != nil {
			log.Printf("Fetching blockchain balances failed: %v", err)
			break
		}
		statuses = self.FetchStatusFromBlockchain(pendings)
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
	return self.blockchain.FetchBalanceData(self.rmaddr, 0)
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
		if act.Action != "set_rates" {
			return false
		}

		actNonce := act.Result["nonce"]
		if actNonce == nil {
			return false
		}
		nonce, _ := strconv.ParseUint(actNonce.(string), 10, 64)
		return nonce < minedNonce
	}
}

func (self *Fetcher) FetchStatusFromBlockchain(pendings []common.ActivityRecord) map[common.ActivityID]common.ActivityStatus {
	result := map[common.ActivityID]common.ActivityStatus{}
	nonceValidator := self.newNonceValidator()

	for _, activity := range pendings {
		if activity.IsBlockchainPending() && (activity.Action == "set_rates" || activity.Action == "deposit" || activity.Action == "withdraw") {
			var blockNum uint64
			var status string
			var err error
			tx := ethereum.HexToHash(activity.Result["tx"].(string))
			if tx.Big().IsInt64() && tx.Big().Int64() == 0 {
				continue
			}
			status, blockNum, err = self.blockchain.TxStatus(tx)
			if err != nil {
				log.Printf("Getting tx status failed, tx will be considered as pending: %s", err)
			}
			switch status {
			case "":
				if nonceValidator(activity) {
					result[activity.ID] = common.ActivityStatus{
						activity.ExchangeStatus,
						activity.Result["tx"].(string),
						blockNum,
						"failed",
						err,
					}
				}
			case "mined":
				result[activity.ID] = common.ActivityStatus{
					activity.ExchangeStatus,
					activity.Result["tx"].(string),
					blockNum,
					"mined",
					err,
				}
			case "failed":
				result[activity.ID] = common.ActivityStatus{
					activity.ExchangeStatus,
					activity.Result["tx"].(string),
					blockNum,
					"failed",
					err,
				}
			case "lost":
				var (
					// expiredDuration is the amount of time after that if a transaction doesn't appear,
					// it is considered failed
					expiredDuration = 15 * time.Minute
					txFailed        = false
				)
				if nonceValidator(activity) {
					txFailed = true
				} else {
					elapsed := common.GetTimepoint() - activity.Timestamp.ToUint64()
					if elapsed > uint64(expiredDuration/time.Millisecond) {
						log.Printf("Fetcher tx status: tx(%s) is lost, elapsed time: %d", activity.Result["tx"].(string), elapsed)
						txFailed = true
					}
				}

				if txFailed {
					result[activity.ID] = common.ActivityStatus{
						activity.ExchangeStatus,
						activity.Result["tx"].(string),
						blockNum,
						"failed",
						err,
					}
				}
			}
		}
	}
	return result
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
		v := value.(common.EBalanceEntry)
		allEBalances[key.(common.ExchangeID)] = v
		if !v.Valid {
			// get old auth data, because get balance error then we have to keep
			// balance to the latest version then analytic won't get exchange balance to zero
			authVersion, err := self.storage.CurrentAuthDataVersion(common.GetTimepoint())
			if err == nil {
				oldAuth, err := self.storage.GetAuthData(authVersion)
				if err != nil {
					allEBalances[key.(common.ExchangeID)] = common.EBalanceEntry{
						Error: err.Error(),
					}
				} else {
					// update old auth to current
					newEbalance := oldAuth.ExchangeBalances[key.(common.ExchangeID)]
					newEbalance.Error = v.Error
					newEbalance.Status = false
					allEBalances[key.(common.ExchangeID)] = newEbalance
				}
			}
			snapshot.Valid = false
			snapshot.Error = v.Error
		}
		return true
	})

	pendingActivities := []common.ActivityRecord{}
	for _, activity := range pendings {
		status, _ := estatuses.Load(activity.ID)
		var activityStatus common.ActivityStatus
		if status != nil {
			activityStatus := status.(common.ActivityStatus)
			log.Printf("In PersistSnapshot: exchange activity status for %+v: %+v", activity.ID, activityStatus)
			if activity.IsExchangePending() {
				activity.ExchangeStatus = activityStatus.ExchangeStatus
			}
			if activity.Result["tx"] != nil && activity.Result["tx"].(string) == "" {
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
		status, _ = bstatuses.Load(activity.ID)
		if status != nil {
			activityStatus = status.(common.ActivityStatus)
			log.Printf("In PersistSnapshot: blockchain activity status for %+v: %+v", activity.ID, activityStatus)

			if activity.IsBlockchainPending() {
				activity.MiningStatus = activityStatus.MiningStatus
			}
			if activityStatus.Error != nil {
				snapshot.Valid = false
				snapshot.Error = activityStatus.Error.Error()
				activity.Result["status_error"] = activityStatus.Error.Error()
			} else {
				activity.Result["status_error"] = ""
			}
		}
		log.Printf("Aggregate statuses, final activity: %+v", activity)
		if activity.IsPending() {
			pendingActivities = append(pendingActivities, activity)
		}
		activity.Result["blockNumber"] = activityStatus.BlockNumber
		err := self.storage.UpdateActivity(activity.ID, activity)
		if err != nil {
			snapshot.Valid = false
			snapshot.Error = err.Error()
		}
	}
	// note: only update status when it's pending status
	snapshot.ExchangeBalances = allEBalances
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
			if activity.Action == "trade" {
				orderID := id.EID
				base := activity.Params["base"].(string)
				quote := activity.Params["quote"].(string)
				// we ignore error of order status because it doesn't affect
				// authdata. Analytic will ignore order status anyway.
				status, _ = exchange.OrderStatus(orderID, base, quote)
			} else if activity.Action == "deposit" {
				txHash := activity.Result["tx"].(string)
				amountStr := activity.Params["amount"].(string)
				amount, _ := strconv.ParseFloat(amountStr, 64)
				currency := activity.Params["token"].(string)
				status, err = exchange.DepositStatus(id, txHash, currency, amount, timepoint)
				log.Printf("Got deposit status for %v: (%s), error(%v)", activity, status, err)
			} else if activity.Action == "withdraw" {
				amountStr := activity.Params["amount"].(string)
				amount, _ := strconv.ParseFloat(amountStr, 64)
				currency := activity.Params["token"].(string)
				tx = activity.Result["tx"].(string)
				status, tx, err = exchange.WithdrawStatus(id.EID, currency, amount, timepoint)
				log.Printf("Got withdraw status for %v: (%s), error(%v)", activity, status, err)
			} else {
				continue
			}
			result[id] = common.ActivityStatus{
				status, tx, blockNum, activity.MiningStatus, err,
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
	exdata, err := exchange.FetchPriceData(timepoint)
	if err != nil {
		log.Printf("Fetching data from %s failed: %v\n", exchange.Name(), err)
	}
	for pair, exchangeData := range exdata {
		data.SetOnePrice(exchange.ID(), pair, exchangeData)
	}
}
