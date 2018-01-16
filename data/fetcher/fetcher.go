package fetcher

import (
	"log"
	"sync"

	"github.com/KyberNetwork/reserve-data/common"
	ethereum "github.com/ethereum/go-ethereum/common"
)

type Fetcher struct {
	storage      Storage
	exchanges    []Exchange
	blockchain   Blockchain
	runner       FetcherRunner
	rmaddr       ethereum.Address
	currentBlock uint64
}

func NewFetcher(
	storage Storage,
	runner FetcherRunner,
	address ethereum.Address) *Fetcher {
	return &Fetcher{
		storage:    storage,
		exchanges:  []Exchange{},
		blockchain: nil,
		runner:     runner,
		rmaddr:     address,
	}
}

func (self *Fetcher) SetBlockchain(blockchain Blockchain) {
	self.blockchain = blockchain
}

func (self *Fetcher) AddExchange(exchange Exchange) {
	self.exchanges = append(self.exchanges, exchange)
}

func (self *Fetcher) Stop() error {
	return self.runner.Stop()
}

func (self *Fetcher) Run() error {
	log.Printf("Fetcher runner is starting...")
	self.runner.Start()
	// go self.RunOrderbookFetcher()
	// go self.RunAuthDataFetcher()
	// go self.RunRateFetcher()
	go self.RunBlockAndLogFetcher()
	log.Printf("Fetcher runner is running...")
	return nil
}

func (self *Fetcher) RunBlockAndLogFetcher() {
	for {
		log.Printf("waiting for signal from block channel")
		t := <-self.runner.GetBlockTicker()
		log.Printf("got signal in block channel with timestamp %d", common.TimeToTimepoint(t))
		timepoint := common.TimeToTimepoint(t)
		self.FetchCurrentBlock(timepoint)
		log.Printf("fetched block from blockchain")
		lastBlock, err := self.storage.LastBlock()
		if err == nil {
			nextBlock := self.FetchLogs(lastBlock+1, timepoint)
			self.storage.UpdateLogBlock(nextBlock, timepoint)
			log.Printf("nextBlock: %d", nextBlock)
		} else {
			log.Printf("failed to get last fetched log block, err: %+v", err)
		}
	}
}

// return block number that we just fetched the logs
func (self *Fetcher) FetchLogs(fromBlock uint64, timepoint uint64) uint64 {
	logs, err := self.blockchain.GetLogs(fromBlock, timepoint)
	if err != nil {
		log.Printf("fetching logs data from block %d failed, error: %v", fromBlock, err)
		if fromBlock == 0 {
			return 0
		} else {
			return fromBlock - 1
		}
	} else {
		log.Printf("Got logs: %+v", logs)
		if len(logs) > 0 {
			for _, l := range logs {
				err = self.storage.StoreTradeLog(l, timepoint)
				if err != nil {
					log.Printf("storing trade log failed, abort storing process and return latest stored log block number, err: %+v", err)
					return l.BlockNumber
				}
			}
			return logs[len(logs)-1].BlockNumber
		} else {
			return fromBlock - 1
		}
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
	data, err := self.blockchain.FetchRates(timepoint)
	if err != nil {
		log.Printf("Fetching rates from blockchain failed: %s\n", err)
	}
	err = self.storage.StoreRate(data, timepoint)
	// fmt.Printf("balance data: %v\n", data)
	if err != nil {
		log.Printf("Storing rates failed: %s\n", err)
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
	self.FetchAuthDataFromBlockchain(
		bbalances, &bstatuses, pendings, timepoint)
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
	pendings []common.ActivityRecord,
	timepoint uint64) {
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
		balances, err = self.FetchBalanceFromBlockchain(timepoint)
		if err != nil {
			log.Printf("Fetching blockchain balances failed: %v\n", err)
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
		self.currentBlock = block
	}
}

func (self *Fetcher) FetchBalanceFromBlockchain(timepoint uint64) (map[string]common.BalanceEntry, error) {
	return self.blockchain.FetchBalanceData(self.rmaddr, timepoint)
}

func (self *Fetcher) FetchStatusFromBlockchain(pendings []common.ActivityRecord) map[common.ActivityID]common.ActivityStatus {
	result := map[common.ActivityID]common.ActivityStatus{}
	for _, activity := range pendings {
		if activity.IsBlockchainPending() && (activity.Action == "set_rates" || activity.Action == "deposit" || activity.Action == "withdraw") {
			tx := ethereum.HexToHash(activity.Result["tx"].(string))
			if tx.Big().IsInt64() && tx.Big().Int64() == 0 {
				continue
			}
			isMined, err := self.blockchain.IsMined(tx)
			if isMined {
				result[activity.ID] = common.ActivityStatus{
					activity.ExchangeStatus,
					activity.Result["tx"].(string),
					"mined",
					err,
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
			if activityStatus.Error == nil {
				if activity.IsExchangePending() {
					activity.ExchangeStatus = activityStatus.ExchangeStatus
				}
				if activity.Result["tx"] != nil && activity.Result["tx"].(string) == "" {
					activity.Result["tx"] = activityStatus.Tx
				}
			} else {
				snapshot.Valid = false
				snapshot.Error = activityStatus.Error.Error()
			}
		}
		status, _ = bstatuses.Load(activity.ID)
		if status != nil {
			activityStatus = status.(common.ActivityStatus)
			log.Printf("In PersistSnapshot: blockchain activity status for %+v: %+v", activity.ID, activityStatus)
			if activityStatus.Error == nil {
				if activity.IsBlockchainPending() {
					activity.MiningStatus = activityStatus.MiningStatus
				}
			} else {
				snapshot.Valid = false
				snapshot.Error = activityStatus.Error.Error()
			}
		}
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
			id := activity.ID
			if activity.Action == "trade" {
				status, err = exchange.OrderStatus(id, timepoint)
			} else if activity.Action == "deposit" {
				status, err = exchange.DepositStatus(id, timepoint)
				log.Printf("Got deposit status for %v: (%s), error(%v)", activity, status, err)
			} else if activity.Action == "withdraw" {
				log.Printf("Activity: %+v", activity)
				tx = activity.Result["tx"].(string)
				status, tx, err = exchange.WithdrawStatus(id, timepoint)
				log.Printf("Got withdraw status for %v: (%s), error(%v)", activity, status, err)
			} else {
				continue
			}
			result[id] = common.ActivityStatus{
				status, tx, activity.MiningStatus, err,
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
