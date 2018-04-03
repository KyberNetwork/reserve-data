package stat

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"strings"
	"sync"

	// "sync"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/stat/util"
	ethereum "github.com/ethereum/go-ethereum/common"
)

const (
	REORG_BLOCK_SAFE       uint64 = 7
	TIMEZONE_BUCKET_PREFIX string = "utc"
	START_TIMEZONE         int64  = -11
	END_TIMEZONE           int64  = 14
)

type Fetcher struct {
	statStorage            StatStorage
	userStorage            UserStorage
	logStorage             LogStorage
	rateStorage            RateStorage
	blockchain             Blockchain
	runner                 FetcherRunner
	currentBlock           uint64
	currentBlockUpdateTime uint64
	deployBlock            uint64
	reserveAddress         ethereum.Address
	thirdPartyReserves     []ethereum.Address
}

func NewFetcher(
	statStorage StatStorage,
	logStorage LogStorage,
	rateStorage RateStorage,
	userStorage UserStorage,
	runner FetcherRunner,
	deployBlock uint64,
	reserve ethereum.Address,
	thirdPartyReserves []ethereum.Address) *Fetcher {
	return &Fetcher{
		statStorage:        statStorage,
		logStorage:         logStorage,
		rateStorage:        rateStorage,
		userStorage:        userStorage,
		blockchain:         nil,
		runner:             runner,
		deployBlock:        deployBlock,
		reserveAddress:     reserve,
		thirdPartyReserves: thirdPartyReserves,
	}
}

func (self *Fetcher) Stop() error {
	return self.runner.Stop()
}

func (self *Fetcher) SetBlockchain(blockchain Blockchain) {
	self.blockchain = blockchain
	self.FetchCurrentBlock()
}

func (self *Fetcher) Run() error {
	log.Printf("Fetcher runner is starting...")
	self.runner.Start()
	go self.RunBlockFetcher()
	go self.RunLogFetcher()
	go self.RunReserveRatesFetcher()
	go self.RunTradeLogProcessor()
	go self.RunCatLogProcessor()
	log.Printf("Fetcher runner is running...")
	return nil
}

func (self *Fetcher) RunCatLogProcessor() {
	for {
		t := <-self.runner.GetCatLogProcessorTicker()
		// get trade log from db
		fromTime, err := self.userStorage.GetLastProcessedCatLogTimepoint()
		if err != nil {
			log.Printf("get last processor state from db failed: %v", err)
			continue
		}
		fromTime += 1
		if fromTime == 1 {
			// there is no cat log being processed before
			// load the first log we have and set the fromTime to it's timestamp
			l, err := self.logStorage.GetFirstCatLog()
			if err != nil {
				log.Printf("can't get first cat log: err(%s)", err)
				continue
			} else {
				fromTime = l.Timestamp - 1
			}
		}
		toTime := common.TimeToTimepoint(t) * 1000000
		maxRange := self.logStorage.MaxRange()
		if toTime-fromTime > maxRange {
			toTime = fromTime + maxRange
		}
		catLogs, err := self.logStorage.GetCatLogs(fromTime, toTime)
		if err != nil {
			log.Printf("get cat log from db failed: %v", err)
			continue
		}
		log.Printf("PROCESS %d cat logs from %d to %d", len(catLogs), fromTime, toTime)
		if len(catLogs) > 0 {
			var last uint64
			for _, l := range catLogs {
				err := self.userStorage.UpdateAddressCategory(
					strings.ToLower(l.Address.Hex()),
					l.Category,
				)
				if err != nil {
					log.Printf("updating address and category failed: err(%s)", err)
				} else {
					if l.Timestamp > last {
						last = l.Timestamp
					}
				}
			}
			self.userStorage.SetLastProcessedCatLogTimepoint(last)
		} else {
			l, err := self.logStorage.GetLastCatLog()
			if err != nil {
				log.Printf("LogFetcher - can't get last cat log: err(%s)", err)
				continue
			} else {
				// log.Printf("LogFetcher - got last cat log: %+v", l)
				if toTime < l.Timestamp {
					// if we are querying on past logs, store toTime as the last
					// processed trade log timepoint
					self.userStorage.SetLastProcessedCatLogTimepoint(toTime)
				}
			}
		}

		log.Println("processed cat logs")
	}
}

func (self *Fetcher) RunTradeLogProcessor() {
	for {
		log.Printf("TradeLogProcessor - waiting for signal from trade log processor channel")
		t := <-self.runner.GetTradeLogProcessorTicker()
		// get trade log from db
		fromTime, err := self.statStorage.GetLastProcessedTradeLogTimepoint()
		if err != nil {
			log.Printf("get trade log processor state from db failed: %v", err)
			continue
		}
		fromTime += 1
		if fromTime == 1 {
			// there is no trade log being processed before
			// load the first log we have and set the fromTime to it's timestamp
			l, err := self.logStorage.GetFirstTradeLog()
			if err != nil {
				log.Printf("can't get first trade log: err(%s)", err)
				continue
			} else {
				log.Printf("got first trade: %+v", l)
				fromTime = l.Timestamp - 1
			}
		}
		toTime := common.TimeToTimepoint(t) * 1000000
		maxRange := self.logStorage.MaxRange()
		if toTime-fromTime > maxRange {
			toTime = fromTime + maxRange
		}
		tradeLogs, err := self.logStorage.GetTradeLogs(fromTime, toTime)
		if err != nil {
			log.Printf("get trade log from db failed: %v", err)
			continue
		}
		log.Printf("AGGREGATE %d trades from %d to %d", len(tradeLogs), fromTime, toTime)
		if len(tradeLogs) > 0 {
			var last uint64
			for _, trade := range tradeLogs {
				log.Printf("AGGREGATE going to aggregate log %d", i)
				if err := self.aggregateTradeLog(trade); err == nil {
					if trade.Timestamp > last {
						last = trade.Timestamp
					}
				}
				log.Printf("AGGREGATE aggregated log %d", i)
			}
			self.statStorage.SetLastProcessedTradeLogTimepoint(last)
			log.Printf("AGGREGATE finished all logs in the batch")
		} else {
			l, err := self.logStorage.GetLastTradeLog()
			if err != nil {
				log.Printf("can't get last trade log: err(%s)", err)
				continue
			} else {
				// log.Printf("got last trade: %+v", l)
				if toTime < l.Timestamp {
					// if we are querying on past logs, store toTime as the last
					// processed trade log timepoint
					self.statStorage.SetLastProcessedTradeLogTimepoint(toTime)
				}
			}
		}
		log.Println("aggregated trade stats")
	}
}

func (self *Fetcher) RunReserveRatesFetcher() {
	for {
		log.Printf("waiting for signal from reserve rate channel")
		t := <-self.runner.GetReserveRatesTicker()
		log.Printf("got signal in reserve rate channel with timstamp %d", common.GetTimepoint())
		timepoint := common.TimeToTimepoint(t)
		self.FetchReserveRates(timepoint)
		log.Printf("fetched reserve rate from blockchain")
	}
}

func (self *Fetcher) GetReserveRates(
	currentBlock uint64, reserveAddr ethereum.Address,
	tokens []common.Token, data *sync.Map, wg *sync.WaitGroup) {
	defer wg.Done()
	rates, err := self.blockchain.GetReserveRates(currentBlock-1, currentBlock, reserveAddr, tokens)
	if err != nil {
		log.Println(err.Error())
	}
	data.Store(string(reserveAddr.Hex()), rates)
}

func (self *Fetcher) FetchReserveRates(timepoint uint64) {
	log.Printf("Fetching reserve and sanity rate from blockchain")
	tokens := []common.Token{}
	for _, token := range common.SupportedTokens {
		if token.ID != "ETH" {
			tokens = append(tokens, token)
		}
	}
	supportedReserves := append(self.thirdPartyReserves, self.reserveAddress)
	data := sync.Map{}
	wg := sync.WaitGroup{}
	// get current block to use to fetch all reserve rates.
	// dont use self.currentBlock directly with self.GetReserveRates
	// because otherwise, rates from different reserves will not
	// be synced with block no
	block := self.currentBlock
	for _, reserveAddr := range supportedReserves {
		wg.Add(1)
		go self.GetReserveRates(block, reserveAddr, tokens, &data, &wg)
	}
	wg.Wait()
	data.Range(func(key, value interface{}) bool {
		reserveAddr := key.(string)
		rates := value.(common.ReserveRates)
		log.Printf("Storing reserve rates to db...")
		self.rateStorage.StoreReserveRates(reserveAddr, rates, common.GetTimepoint())
		return true
	})
}

func (self *Fetcher) RunLogFetcher() {
	for {
		log.Printf("LogFetcher - waiting for signal from log channel")
		t := <-self.runner.GetLogTicker()
		timepoint := common.TimeToTimepoint(t)
		log.Printf("LogFetcher - got signal in log channel with timestamp %d", timepoint)
		lastBlock, err := self.logStorage.LastBlock()
		if lastBlock == 0 {
			lastBlock = self.deployBlock
		}
		if err == nil {
			toBlock := lastBlock + 1 + 1440 // 1440 is considered as 6 hours
			if toBlock > self.currentBlock-REORG_BLOCK_SAFE {
				toBlock = self.currentBlock - REORG_BLOCK_SAFE
			}
			if lastBlock+1 > toBlock {
				continue
			}
			nextBlock := self.FetchLogs(lastBlock+1, toBlock, timepoint)
			if nextBlock == lastBlock && toBlock != 0 {
				// in case that we are querying old blocks (6 hours in the past)
				// and got no logs. we will still continue with next block
				// It is not the case if toBlock == 0, means we are querying
				// best window, we should keep querying it in order not to
				// miss any logs due to node inconsistency
				nextBlock = toBlock
			}
			log.Printf("LogFetcher - update log block: %d", nextBlock)
			self.logStorage.UpdateLogBlock(nextBlock, timepoint)
			log.Printf("LogFetcher - nextBlock: %d", nextBlock)
		} else {
			log.Printf("LogFetcher - failed to get last fetched log block, err: %+v", err)
		}
	}
}

func (self *Fetcher) RunBlockFetcher() {
	for {
		log.Printf("waiting for signal from block channel")
		t := <-self.runner.GetBlockTicker()
		timepoint := common.TimeToTimepoint(t)
		log.Printf("got signal in block channel with timestamp %d", timepoint)
		self.FetchCurrentBlock()
		log.Printf("fetched block from blockchain")
	}
}

func (self *Fetcher) GetTradeGeo(txHash string) (string, string, error) {
	url := fmt.Sprintf("https://broadcast.kyber.network/get-tx-info/%s", txHash)

	resp, err := http.Get(url)
	if err != nil {
		return "", "", err
	}
	response := common.TradeLogGeoInfoResp{}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", "", err
	}
	if response.Success {
		if response.Data.Country != "" {
			return response.Data.IP, response.Data.Country, err
		}
		country, err := util.IpToCountry(response.Data.IP)
		if err != nil {
			return "", "", err
		}
		return response.Data.IP, country, err
	}
	return "", "unknown", err
}

// return block number that we just fetched the logs
func (self *Fetcher) FetchLogs(fromBlock uint64, toBlock uint64, timepoint uint64) uint64 {
	logs, err := self.blockchain.GetLogs(fromBlock, toBlock)
	if err != nil {
		log.Printf("LogFetcher - fetching logs data from block %d failed, error: %v", fromBlock, err)
		if fromBlock == 0 {
			return 0
		} else {
			return fromBlock - 1
		}
	} else {
		if len(logs) > 0 {
			for _, il := range logs {
				if il.Type() == "TradeLog" {
					l := il.(common.TradeLog)
					txHash := il.TxHash()
					ip, country, err := self.GetTradeGeo(txHash.Hex())
					l.IP = ip
					l.Country = country
					err = self.statStorage.SetCountry(country)
					if err != nil {
						log.Printf("Cannot store country: %s", err.Error())
					}

					err = self.logStorage.StoreTradeLog(l, timepoint)
					if err != nil {
						log.Printf("LogFetcher - storing trade log failed, ignore that log and proceed with remaining logs, err: %+v", err)
					}
				} else if il.Type() == "SetCatLog" {
					l := il.(common.SetCatLog)
					err = self.logStorage.StoreCatLog(l)
					if err != nil {
						log.Printf("LogFetcher - storing cat log failed, ignore that log and proceed with remaining logs, err: %+v", err)
					}
				}
			}
			var max uint64 = 0
			for _, l := range logs {
				if l.BlockNo() > max {
					max = l.BlockNo()
				}
			}
			return max
		} else {
			return fromBlock - 1
		}
	}
}

func (self *Fetcher) updateTimeZoneBuckets(timestamp uint64, updates common.TradeStats) (err error) {
	//update to timezone buckets
	log.Printf("AGGREGATE start updaing timezone buckets")
	for i := START_TIMEZONE; i <= END_TIMEZONE; i++ {
		log.Printf("AGGREGATE updaing timezone buckets: %d", i)
		freq := fmt.Sprintf("%s%d", TIMEZONE_BUCKET_PREFIX, i)
		err = self.statStorage.SetTradeStats(freq, timestamp, updates)
		if err != nil {
			return err
		}
	}
	log.Printf("AGGREGATE finish updaing timezone buckets")
	return nil
}

func checkWalletAddress(wallet string) bool {
	walletAddr := ethereum.HexToAddress(wallet)
	cap := big.NewInt(0)
	cap.Exp(big.NewInt(2), big.NewInt(128), big.NewInt(0))
	if walletAddr.Big().Cmp(cap) < 0 {
		return false
	}
	return true
}

func (self *Fetcher) aggregateTradeLog(trade common.TradeLog) (err error) {
	srcAddr := common.AddrToString(trade.SrcAddress)
	dstAddr := common.AddrToString(trade.DestAddress)
	reserveAddr := common.AddrToString(trade.ReserveAddress)
	walletAddr := common.AddrToString(trade.WalletAddress)
	userAddr := common.AddrToString(trade.UserAddress)
	country := trade.Country

	log.Printf("AGGREGATE step 1")
	if checkWalletAddress(walletAddr) {
		self.statStorage.SetWalletAddress(walletAddr)
	}

	var srcAmount, destAmount, ethAmount, burnFee, walletFee float64
	var tokenAddr string
	log.Printf("AGGREGATE step 2")
	for _, token := range common.SupportedTokens {
		if strings.ToLower(token.Address) == srcAddr {
			srcAmount = common.BigToFloat(trade.SrcAmount, token.Decimal)
			if token.IsETH() {
				ethAmount = srcAmount
			} else {
				tokenAddr = strings.ToLower(token.Address)
			}
		}

		if strings.ToLower(token.Address) == dstAddr {
			destAmount = common.BigToFloat(trade.DestAmount, token.Decimal)
			if token.IsETH() {
				ethAmount = destAmount
			} else {
				tokenAddr = strings.ToLower(token.Address)
			}
		}
	}
	log.Printf("AGGREGATE step 3")

	eth := common.SupportedTokens["ETH"]
	if trade.BurnFee != nil {
		burnFee = common.BigToFloat(trade.BurnFee, eth.Decimal)
	}
	if trade.WalletFee != nil {
		walletFee = common.BigToFloat(trade.WalletFee, eth.Decimal)
	}

	updates := common.TradeStats{
		fmt.Sprintf("assets_volume_%s", srcAddr):                 srcAmount,
		fmt.Sprintf("assets_volume_%s", dstAddr):                 destAmount,
		fmt.Sprintf("assets_eth_amount_%s", tokenAddr):           ethAmount,
		fmt.Sprintf("assets_usd_amount_%s", srcAddr):             trade.FiatAmount,
		fmt.Sprintf("assets_usd_amount_%s", dstAddr):             trade.FiatAmount,
		fmt.Sprintf("burn_fee_%s", reserveAddr):                  burnFee,
		fmt.Sprintf("wallet_fee_%s_%s", reserveAddr, walletAddr): walletFee,
		fmt.Sprintf("user_volume_%s", userAddr):                  trade.FiatAmount,
	}

	log.Printf("AGGREGATE step 4")
	for _, freq := range []string{"M", "H", "D"} {
		err = self.statStorage.SetTradeStats(freq, trade.Timestamp, updates)
		if err != nil {
			return
		}
	}
	log.Printf("AGGREGATE step 5")
	if err = self.updateTimeZoneBuckets(trade.Timestamp, updates); err != nil {
		return
	}
	log.Printf("AGGREGATE step 6")
	// total stats on trading
	updates = common.TradeStats{
		"eth_volume":  ethAmount,
		"usd_volume":  trade.FiatAmount,
		"burn_fee":    burnFee,
		"trade_count": 1,
	}

	if err = self.updateTimeZoneBuckets(trade.Timestamp, updates); err != nil {
		return
	}
	log.Printf("AGGREGATE step 7")

	// stats on user
	userAddr = strings.ToLower(trade.UserAddress.String())
	email, regTime, err := self.userStorage.GetUserOfAddress(userAddr)
	if err != nil {
		return
	}

	log.Printf("AGGREGATE step 8")
	var kycEd bool
	if email != "" && email != userAddr && trade.Timestamp > regTime {
		kycEd = true
	}
	for i := START_TIMEZONE; i <= END_TIMEZONE; i++ {
		userStats, err := self.statStorage.GetUserStats(trade.Timestamp, userAddr, email, walletAddr, country, kycEd, i)
		if err != nil {
			return err
		}

		if len(userStats) > 0 {
			if err := self.statStorage.SetUserStats(trade.Timestamp, userAddr, email, walletAddr, country, kycEd, i, userStats); err != nil {
				log.Println("Set user stats failed: ", err)
				return err
			}
		}
	}
	log.Printf("AGGREGATE step 9")

	//stats on wallet address
	updates = common.TradeStats{
		fmt.Sprintf("wallet_eth_volume_%s", walletAddr):  ethAmount,
		fmt.Sprintf("wallet_usd_volume_%s", walletAddr):  trade.FiatAmount,
		fmt.Sprintf("wallet_burn_fee_%s", walletAddr):    burnFee,
		fmt.Sprintf("wallet_trade_count_%s", walletAddr): 1,
	}
	//update to timezone buckets
	if err = self.updateTimeZoneBuckets(trade.Timestamp, updates); err != nil {
		return
	}
	log.Printf("AGGREGATE step 10")

	// update geo stats
	updates = common.TradeStats{
		fmt.Sprintf("geo_eth_volume_%s", country):  ethAmount,
		fmt.Sprintf("geo_usd_volume_%s", country):  trade.FiatAmount,
		fmt.Sprintf("geo_burn_fee_%s", country):    burnFee,
		fmt.Sprintf("geo_trade_count_%s", country): 1,
	}

	if err = self.updateTimeZoneBuckets(trade.Timestamp, updates); err != nil {
		return
	}
	log.Printf("AGGREGATE step 11")

	return
}

func (self *Fetcher) FetchCurrentBlock() {
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
