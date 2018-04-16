package stat

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/stat/util"
	ethereum "github.com/ethereum/go-ethereum/common"
)

const (
	REORG_BLOCK_SAFE       uint64 = 7
	TIMEZONE_BUCKET_PREFIX string = "utc"
	START_TIMEZONE         int64  = -11
	END_TIMEZONE           int64  = 14

	TRADE_SUMMARY_AGGREGATION  string = "trade_summary_aggregation"
	WALLET_AGGREGATION         string = "wallet_aggregation"
	COUNTRY_AGGREGATION        string = "country_aggregation"
	USER_AGGREGATION           string = "user_aggregation"
	VOLUME_STAT_AGGREGATION    string = "volume_stat_aggregation"
	BURNFEE_AGGREGATION        string = "burn_fee_aggregation"
	RESERVE_VOLUME_AGGREGATION string = "reserve_volume_aggregation"
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

func (self *Fetcher) GetTradeLogTimeRange(fromTime uint64, t time.Time) (uint64, uint64) {
	fromTime += 1
	if fromTime == 1 {
		// there is no trade log being processed before
		// load the first log we have and set the fromTime to it's timestamp
		l, err := self.logStorage.GetFirstTradeLog()
		if err != nil {
			log.Printf("can't get first trade log: err(%s)", err)
			// continue
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
	return fromTime, toTime
}

func (self *Fetcher) RunCountryStatAggregation(t time.Time) {
	// get trade log from db
	fromTime, err := self.statStorage.GetLastProcessedTradeLogTimepoint(COUNTRY_AGGREGATION)
	if err != nil {
		log.Printf("get trade log processor state from db failed: %v", err)
		return
	}
	fromTime, toTime := self.GetTradeLogTimeRange(fromTime, t)
	tradeLogs, err := self.logStorage.GetTradeLogs(fromTime, toTime)
	if err != nil {
		log.Printf("get trade log from db failed: %v", err)
		return
	}
	if len(tradeLogs) > 0 {
		var last uint64
		countryStats := map[string]common.MetricStatsTimeZone{}
		allFirstTradeEver, _ := self.statStorage.GetAllFirstTradeEver()
		for _, trade := range tradeLogs {
			if err := self.aggregateCountryStats(trade, countryStats, allFirstTradeEver); err == nil {
				if trade.Timestamp > last {
					last = trade.Timestamp
				}
			}
		}
		self.statStorage.SetCountryStat(countryStats, last)
		// self.statStorage.SetLastProcessedTradeLogTimepoint(COUNTRY_AGGREGATION, last)
	} else {
		l, err := self.logStorage.GetLastTradeLog()
		if err != nil {
			log.Printf("can't get last trade log: err(%s)", err)
			return
		} else {
			if toTime < l.Timestamp {
				// if we are querying on past logs, store toTime as the last
				// processed trade log timepoint
				self.statStorage.SetLastProcessedTradeLogTimepoint(COUNTRY_AGGREGATION, toTime)
			}
		}
	}
}

func (self *Fetcher) RunTradeSummaryAggregation(t time.Time) {
	// get trade log from db
	fromTime, err := self.statStorage.GetLastProcessedTradeLogTimepoint(TRADE_SUMMARY_AGGREGATION)
	if err != nil {
		log.Printf("get trade log processor state from db failed: %v", err)
		return
	}
	fromTime, toTime := self.GetTradeLogTimeRange(fromTime, t)
	tradeLogs, err := self.logStorage.GetTradeLogs(fromTime, toTime)
	if err != nil {
		log.Printf("get trade log from db failed: %v", err)
		return
	}
	if len(tradeLogs) > 0 {
		var last uint64

		tradeSummary := map[string]common.MetricStatsTimeZone{}
		allFirstTradeEver, _ := self.statStorage.GetAllFirstTradeEver()
		for _, trade := range tradeLogs {
			if err := self.aggregateTradeSumary(trade, tradeSummary, allFirstTradeEver); err == nil {
				if trade.Timestamp > last {
					last = trade.Timestamp
				}
			}
		}
		self.statStorage.SetTradeSummary(tradeSummary, last)
		// self.statStorage.SetLastProcessedTradeLogTimepoint(TRADE_SUMMARY_AGGREGATION, last)
	} else {
		l, err := self.logStorage.GetLastTradeLog()
		if err != nil {
			log.Printf("can't get last trade log: err(%s)", err)
			return
		} else {
			if toTime < l.Timestamp {
				// if we are querying on past logs, store toTime as the last
				// processed trade log timepoint
				self.statStorage.SetLastProcessedTradeLogTimepoint(TRADE_SUMMARY_AGGREGATION, toTime)
			}
		}
	}
}

func (self *Fetcher) RunWalletStatAggregation(t time.Time) {
	// get trade log from db
	fromTime, err := self.statStorage.GetLastProcessedTradeLogTimepoint(WALLET_AGGREGATION)
	if err != nil {
		log.Printf("get trade log processor state from db failed: %v", err)
		return
	}
	fromTime, toTime := self.GetTradeLogTimeRange(fromTime, t)
	tradeLogs, err := self.logStorage.GetTradeLogs(fromTime, toTime)
	if err != nil {
		log.Printf("get trade log from db failed: %v", err)
		return
	}
	if len(tradeLogs) > 0 {
		var last uint64

		walletStats := map[string]common.MetricStatsTimeZone{}
		allFirstTradeEver, _ := self.statStorage.GetAllFirstTradeEver()
		for _, trade := range tradeLogs {
			if err := self.aggregateWalletStats(trade, walletStats, allFirstTradeEver); err == nil {
				if trade.Timestamp > last {
					last = trade.Timestamp
				}
			}
		}
		self.statStorage.SetWalletStat(walletStats, last)
		// self.statStorage.SetLastProcessedTradeLogTimepoint(WALLET_AGGREGATION, last)
	} else {
		l, err := self.logStorage.GetLastTradeLog()
		if err != nil {
			log.Printf("can't get last trade log: err(%s)", err)
			return
		} else {
			if toTime < l.Timestamp {
				// if we are querying on past logs, store toTime as the last
				// processed trade log timepoint
				self.statStorage.SetLastProcessedTradeLogTimepoint(WALLET_AGGREGATION, toTime)
			}
		}
	}
}

func (self *Fetcher) RunBurnFeeAggregation(t time.Time) {
	// get trade log from db
	fromTime, err := self.statStorage.GetLastProcessedTradeLogTimepoint(BURNFEE_AGGREGATION)
	if err != nil {
		log.Printf("get trade log processor state from db failed: %v", err)
		return
	}
	fromTime, toTime := self.GetTradeLogTimeRange(fromTime, t)
	tradeLogs, err := self.logStorage.GetTradeLogs(fromTime, toTime)
	if err != nil {
		log.Printf("get trade log from db failed: %v", err)
		return
	}
	if len(tradeLogs) > 0 {
		var last uint64

		burnFeeStats := map[string]common.BurnFeeStatsTimeZone{}
		for _, trade := range tradeLogs {
			if err := self.aggregateBurnFeeStats(trade, burnFeeStats); err == nil {
				if trade.Timestamp > last {
					last = trade.Timestamp
				}
			}
		}
		self.statStorage.SetBurnFeeStat(burnFeeStats, last)
		// self.statStorage.SetLastProcessedTradeLogTimepoint(BURNFEE_AGGREGATION, last)
	} else {
		l, err := self.logStorage.GetLastTradeLog()
		if err != nil {
			log.Printf("can't get last trade log: err(%s)", err)
			return
		} else {
			if toTime < l.Timestamp {
				self.statStorage.SetLastProcessedTradeLogTimepoint(BURNFEE_AGGREGATION, toTime)
			}
		}
	}
}

func (self *Fetcher) RunVolumeStatAggregation(t time.Time) {
	// get trade log from db
	fromTime, err := self.statStorage.GetLastProcessedTradeLogTimepoint(VOLUME_STAT_AGGREGATION)
	if err != nil {
		log.Printf("get trade log processor state from db failed: %v", err)
		return
	}
	fromTime, toTime := self.GetTradeLogTimeRange(fromTime, t)
	tradeLogs, err := self.logStorage.GetTradeLogs(fromTime, toTime)
	if err != nil {
		log.Printf("get trade log from db failed: %v", err)
		return
	}
	if len(tradeLogs) > 0 {
		var last uint64

		volumeStats := map[string]common.VolumeStatsTimeZone{}
		for _, trade := range tradeLogs {
			if err := self.aggregateVolumeStats(trade, volumeStats); err == nil {
				if trade.Timestamp > last {
					last = trade.Timestamp
				}
			}
		}
		self.statStorage.SetVolumeStat(volumeStats, last)
		// self.statStorage.SetLastProcessedTradeLogTimepoint(VOLUME_STAT_AGGREGATION, last)
	} else {
		l, err := self.logStorage.GetLastTradeLog()
		if err != nil {
			log.Printf("can't get last trade log: err(%s)", err)
			return
		} else {
			if toTime < l.Timestamp {
				self.statStorage.SetLastProcessedTradeLogTimepoint(VOLUME_STAT_AGGREGATION, toTime)
			}
		}
	}
	return
}

func (self *Fetcher) RunUserAggregation(t time.Time) {
	// get trade log from db
	fromTime, err := self.statStorage.GetLastProcessedTradeLogTimepoint(USER_AGGREGATION)
	if err != nil {
		log.Printf("get trade log processor state from db failed: %v", err)
		return
	}
	fromTime, toTime := self.GetTradeLogTimeRange(fromTime, t)
	tradeLogs, err := self.logStorage.GetTradeLogs(fromTime, toTime)
	if err != nil {
		log.Printf("get trade log from db failed: %v", err)
		return
	}
	if len(tradeLogs) > 0 {
		var last uint64
		userTradeList := map[string]uint64{} // map of user address and fist trade timestamp
		for _, trade := range tradeLogs {
			userAddr := common.AddrToString(trade.UserAddress)
			key := fmt.Sprintf("%s_%d", userAddr, trade.Timestamp)
			userTradeList[key] = trade.Timestamp
			if trade.Timestamp > last {
				last = trade.Timestamp
			}
		}
		self.statStorage.SetFirstTradeEver(userTradeList, last)
		self.statStorage.SetFirstTradeInDay(userTradeList, last)
		self.statStorage.SetLastProcessedTradeLogTimepoint(USER_AGGREGATION, last)
	} else {
		l, err := self.logStorage.GetLastTradeLog()
		if err != nil {
			log.Printf("can't get last trade log: err(%s)", err)
			return
		} else {
			if toTime < l.Timestamp {
				self.statStorage.SetLastProcessedTradeLogTimepoint(USER_AGGREGATION, toTime)
			}
		}
	}
}

func runAggregationInParallel(wg *sync.WaitGroup, t time.Time, f func(t time.Time)) {
	defer wg.Done()
	f(t)
}

func (self *Fetcher) RunTradeLogProcessor() {
	for {
		t := <-self.runner.GetTradeLogProcessorTicker()
		self.RunUserAggregation(t)
		wg := sync.WaitGroup{}
		wg.Add(1)
		go runAggregationInParallel(&wg, t, self.RunBurnFeeAggregation)
		wg.Add(1)
		go runAggregationInParallel(&wg, t, self.RunVolumeStatAggregation)
		wg.Add(1)
		go runAggregationInParallel(&wg, t, self.RunTradeSummaryAggregation)
		wg.Add(1)
		go runAggregationInParallel(&wg, t, self.RunWalletStatAggregation)
		wg.Add(1)
		go runAggregationInParallel(&wg, t, self.RunCountryStatAggregation)
		wg.Wait()
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
			nextBlock, err := self.FetchLogs(lastBlock+1, toBlock, timepoint)
			if err != nil {
				// in case there is error, we roll back and try it again.
				// dont have to do anything here. just continute with the loop.
				log.Printf("LogFetcher - continue with the loop to try it again")
			} else {
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
			}
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
		country, err := util.IPToCountry(response.Data.IP)
		if err != nil {
			return "", "", err
		}
		return response.Data.IP, country, err
	}
	return "", "unknown", err
}

// return block number that we just fetched the logs
func (self *Fetcher) FetchLogs(fromBlock uint64, toBlock uint64, timepoint uint64) (uint64, error) {
	logs, err := self.blockchain.GetLogs(fromBlock, toBlock)
	if err != nil {
		log.Printf("LogFetcher - fetching logs data from block %d failed, error: %v", fromBlock, err)
		if fromBlock == 0 {
			return 0, err
		} else {
			return fromBlock - 1, err
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
			return max, nil
		} else {
			return fromBlock - 1, nil
		}
	}
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

func getTimestampFromTimeZone(t uint64, freq string) uint64 {
	result := uint64(0)
	ui64Day := uint64(time.Hour * 24)
	switch freq {
	case "m", "M":
		result = t / uint64(time.Minute) * uint64(time.Minute)
	case "h", "H":
		result = t / uint64(time.Hour) * uint64(time.Hour)
	case "d", "D":
		result = t / ui64Day * ui64Day
	default:
		offset, _ := strconv.ParseInt(strings.TrimPrefix(freq, "utc"), 10, 64)
		ui64offset := uint64(int64(time.Hour) * offset)
		if offset > 0 {
			result = (t+ui64offset)/ui64Day*ui64Day + ui64offset
		} else {
			offset = 0 - offset
			result = (t-ui64offset)/ui64Day*ui64Day - ui64offset
		}
	}
	return result
}

func (self *Fetcher) getTradeInfo(trade common.TradeLog) (float64, float64, float64, float64, bool, error) {

	srcAddr := common.AddrToString(trade.SrcAddress)
	dstAddr := common.AddrToString(trade.DestAddress)
	userAddr := common.AddrToString(trade.UserAddress)

	var srcAmount, destAmount, ethAmount, burnFee float64
	for _, token := range common.SupportedTokens {
		if strings.ToLower(token.Address) == srcAddr {
			srcAmount = common.BigToFloat(trade.SrcAmount, token.Decimal)
			if token.IsETH() {
				ethAmount = srcAmount
			}
		}

		if strings.ToLower(token.Address) == dstAddr {
			destAmount = common.BigToFloat(trade.DestAmount, token.Decimal)
			if token.IsETH() {
				ethAmount = destAmount
			}
		}
	}

	eth := common.SupportedTokens["ETH"]
	if trade.BurnFee != nil {
		burnFee = common.BigToFloat(trade.BurnFee, eth.Decimal)
	}
	// stats on user
	userAddr = strings.ToLower(trade.UserAddress.String())
	email, regTime, err := self.userStorage.GetUserOfAddress(userAddr)
	if err != nil {
		return srcAmount, destAmount, ethAmount, burnFee, false, err
	}

	var kycEd bool
	if email != "" && email != userAddr && trade.Timestamp > regTime {
		kycEd = true
	}

	return srcAmount, destAmount, ethAmount, burnFee, kycEd, nil
}

func (self *Fetcher) aggregateCountryStats(trade common.TradeLog,
	countryStats map[string]common.MetricStatsTimeZone, allFirstTradeEver map[string]uint64) error {

	err := self.statStorage.SetCountry(trade.Country)
	if err != nil {
		log.Printf("Cannot store country: %s", err.Error())
		return err
	}
	_, _, ethAmount, burnFee, kycEd, _ := self.getTradeInfo(trade)
	self.aggregateMetricStat(trade, trade.Country, ethAmount, burnFee, countryStats, kycEd, allFirstTradeEver)
	return err
}

func (self *Fetcher) aggregateWalletStats(trade common.TradeLog,
	walletStats map[string]common.MetricStatsTimeZone, allFirstTradeEver map[string]uint64) error {

	walletAddr := common.AddrToString(trade.WalletAddress)
	if checkWalletAddress(walletAddr) {
		self.statStorage.SetWalletAddress(walletAddr)
	}
	_, _, ethAmount, burnFee, kycEd, _ := self.getTradeInfo(trade)
	self.aggregateMetricStat(trade, walletAddr, ethAmount, burnFee, walletStats, kycEd, allFirstTradeEver)
	return nil
}

func (self *Fetcher) aggregateTradeSumary(trade common.TradeLog,
	tradeSummary map[string]common.MetricStatsTimeZone, allFirstTradeEver map[string]uint64) error {

	_, _, ethAmount, burnFee, kycEd, _ := self.getTradeInfo(trade)
	self.aggregateMetricStat(trade, "trade_summary", ethAmount, burnFee, tradeSummary, kycEd, allFirstTradeEver)
	return nil
}

func (self *Fetcher) aggregateVolumeStats(trade common.TradeLog, volumeStats map[string]common.VolumeStatsTimeZone) error {

	srcAddr := common.AddrToString(trade.SrcAddress)
	dstAddr := common.AddrToString(trade.DestAddress)
	userAddr := common.AddrToString(trade.UserAddress)
	reserveAddr := common.AddrToString(trade.ReserveAddress)

	srcAmount, destAmount, ethAmount, _, _, _ := self.getTradeInfo(trade)
	// token volume
	self.aggregateVolumeStat(trade, srcAddr, srcAmount, ethAmount, trade.FiatAmount, volumeStats)
	self.aggregateVolumeStat(trade, dstAddr, destAmount, ethAmount, trade.FiatAmount, volumeStats)

	//user volume
	self.aggregateVolumeStat(trade, userAddr, srcAmount, ethAmount, trade.FiatAmount, volumeStats)

	// reserve volume
	self.aggregateVolumeStat(trade, userAddr, srcAmount, ethAmount, trade.FiatAmount, volumeStats)

	// reserve volume
	eth := common.MustGetToken("ETH")
	var assetAddr string
	var assetAmount float64
	if srcAddr != eth.Address {
		assetAddr = srcAddr
		assetAmount = srcAmount
	} else {
		assetAddr = dstAddr
		assetAmount = destAmount
	}
	key := fmt.Sprintf("%s_%s", reserveAddr, assetAddr)
	self.aggregateVolumeStat(trade, key, assetAmount, ethAmount, trade.FiatAmount, volumeStats)
	return nil
}

func (self *Fetcher) aggregateBurnFeeStats(trade common.TradeLog, burnFeeStats map[string]common.BurnFeeStatsTimeZone) error {

	reserveAddr := common.AddrToString(trade.ReserveAddress)
	walletAddr := common.AddrToString(trade.WalletAddress)
	_, _, _, burnFee, _, _ := self.getTradeInfo(trade)
	// reserve burn fee
	self.aggregateBurnfee(reserveAddr, burnFee, trade, burnFeeStats)

	// wallet fee
	var walletFee float64
	eth := common.MustGetToken("ETH")
	if trade.WalletFee != nil {
		walletFee = common.BigToFloat(trade.WalletFee, eth.Decimal)
	}
	self.aggregateBurnfee(fmt.Sprintf("%s_%s", reserveAddr, walletAddr), walletFee, trade, burnFeeStats)
	return nil
}

func (self *Fetcher) aggregateBurnfee(key string, fee float64, trade common.TradeLog, burnFeeStats map[string]common.BurnFeeStatsTimeZone) {
	for _, freq := range []string{"M", "H", "D"} {
		timestamp := getTimestampFromTimeZone(trade.Timestamp, freq)

		currentVolume, exist := burnFeeStats[key]
		if !exist {
			currentVolume = common.BurnFeeStatsTimeZone{}
		}
		dataTimeZone, exist := currentVolume[freq]
		if !exist {
			dataTimeZone = map[uint64]common.BurnFeeStats{}
		}
		data, exist := dataTimeZone[timestamp]
		if !exist {
			data = common.BurnFeeStats{}
		}
		data.TotalBurnFee += fee
		dataTimeZone[timestamp] = data
		currentVolume[freq] = dataTimeZone
		burnFeeStats[key] = currentVolume
	}
}

func (self *Fetcher) aggregateVolumeStat(
	trade common.TradeLog,
	assetAddr string,
	assetAmount, ethAmount, fiatAmount float64,
	assetVolumetStats map[string]common.VolumeStatsTimeZone) {
	for _, freq := range []string{"M", "H", "D"} {
		timestamp := getTimestampFromTimeZone(trade.Timestamp, freq)

		currentVolume, exist := assetVolumetStats[assetAddr]
		if !exist {
			currentVolume = common.VolumeStatsTimeZone{}
		}
		dataTimeZone, exist := currentVolume[freq]
		if !exist {
			dataTimeZone = map[uint64]common.VolumeStats{}
		}
		data, exist := dataTimeZone[timestamp]
		if !exist {
			data = common.VolumeStats{}
		}
		data.ETHVolume += ethAmount
		data.USDAmount += fiatAmount
		data.Volume += assetAmount
		dataTimeZone[timestamp] = data
		currentVolume[freq] = dataTimeZone
		assetVolumetStats[assetAddr] = currentVolume
	}
}

func (self *Fetcher) aggregateMetricStat(trade common.TradeLog, statKey string, ethAmount, burnFee float64,
	metricStats map[string]common.MetricStatsTimeZone,
	kycEd bool,
	allFirstTradeEver map[string]uint64) {
	userAddr := common.AddrToString(trade.UserAddress)

	for i := START_TIMEZONE; i <= END_TIMEZONE; i++ {
		freq := fmt.Sprintf("%s%d", TIMEZONE_BUCKET_PREFIX, i)
		timestamp := getTimestampFromTimeZone(trade.Timestamp, freq)
		currentMetricData, exist := metricStats[statKey]
		if !exist {
			currentMetricData = common.MetricStatsTimeZone{}
		}
		dataTimeZone, exist := currentMetricData[i]
		if !exist {
			dataTimeZone = map[uint64]common.MetricStats{}
		}
		data, exist := dataTimeZone[timestamp]
		if !exist {
			data = common.MetricStats{}
		}
		timeFirstTrade := allFirstTradeEver[userAddr]
		if timeFirstTrade == trade.Timestamp {
			data.NewUniqueAddresses++
			data.UniqueAddr++
			if kycEd {
				data.KYCEd++
			}
		} else {
			firstTradeInday := self.statStorage.GetFirstTradeInDay(userAddr, trade.Timestamp, i)
			if firstTradeInday == trade.Timestamp {
				data.UniqueAddr++
				if kycEd {
					data.KYCEd++
				}
			}
		}

		data.ETHVolume += ethAmount
		data.BurnFee += burnFee
		data.TradeCount++
		data.USDVolume += trade.FiatAmount
		dataTimeZone[timestamp] = data
		currentMetricData[i] = dataTimeZone
		metricStats[statKey] = currentMetricData
	}
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
