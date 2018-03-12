package stat

import (
	"log"
	"strings"
	"sync"
	// "sync"

	"github.com/KyberNetwork/reserve-data/common"
	ethereum "github.com/ethereum/go-ethereum/common"
)

type CoinCapRateResponse []struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	Rank     string `json:"rank"`
	PriceUSD string `json:"price_usd"`
}

type Fetcher struct {
	storage                Storage
	blockchain             Blockchain
	runner                 FetcherRunner
	ethRate                EthUSDRate
	currentBlock           uint64
	currentBlockUpdateTime uint64
	deployBlock            uint64
	reserveAddress         ethereum.Address
	thirdPartyReserves     []ethereum.Address
}

func NewFetcher(
	storage Storage,
	ethUSDRate EthUSDRate,
	runner FetcherRunner,
	deployBlock uint64,
	reserve ethereum.Address,
	thirdPartyReserves []ethereum.Address) *Fetcher {
	return &Fetcher{
		storage:            storage,
		blockchain:         nil,
		runner:             runner,
		ethRate:            ethUSDRate,
		deployBlock:        deployBlock,
		reserveAddress:     reserve,
		thirdPartyReserves: thirdPartyReserves,
	}
}

func (self *Fetcher) Stop() error {
	return self.runner.Stop()
}

func (self *Fetcher) GetEthRate(timepoint uint64) float64 {
	rate := self.ethRate.GetUSDRate(timepoint)
	log.Printf("ETH-USD rate: %f", rate)
	return rate
}

func (self *Fetcher) SetBlockchain(blockchain Blockchain) {
	self.blockchain = blockchain
	self.FetchCurrentBlock()
}

func (self *Fetcher) Run() error {
	log.Printf("Fetcher runner is starting...")
	self.runner.Start()
	go self.RunBlockAndLogFetcher()
	go self.RunReserveRatesFetcher()
	log.Printf("Fetcher runner is running...")
	return nil
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
	for _, reserveAddr := range supportedReserves {
		wg.Add(1)
		go self.GetReserveRates(self.currentBlock, reserveAddr, tokens, &data, &wg)
	}
	wg.Wait()
	data.Range(func(key, value interface{}) bool {
		reserveAddr := key.(string)
		rates := value.(common.ReserveRates)
		self.storage.StoreReserveRates(reserveAddr, rates, common.GetTimepoint())
		return true
	})
}

func (self *Fetcher) RunBlockAndLogFetcher() {
	for {
		log.Printf("waiting for signal from block channel")
		t := <-self.runner.GetBlockTicker()
		log.Printf("got signal in block channel with timestamp %d", common.TimeToTimepoint(t))
		timepoint := common.TimeToTimepoint(t)
		self.FetchCurrentBlock()
		log.Printf("fetched block from blockchain")
		lastBlock, err := self.storage.LastBlock()
		if lastBlock == 0 {
			lastBlock = self.deployBlock
		}
		if err == nil {
			toBlock := lastBlock + 1 + 1440 // 1440 is considered as 6 hours
			if toBlock > self.currentBlock {
				// set toBlock to 0 so we will fetch to last block
				toBlock = 0
			}
			nextBlock := self.FetchLogs(lastBlock+1, toBlock, timepoint)
			if nextBlock == lastBlock && toBlock != 0 {
				// in case that we are querying old blocks (6 hours in the past)
				// and got no logs. we will still continue with next block
				// It is not the case if toBlock == 0, means we are querying
				// best window, we should keep querying it in order not to
				// miss any logs due to node inconsistency
				nextBlock = toBlock + 1
			}
			self.storage.UpdateLogBlock(nextBlock, timepoint)
			log.Printf("nextBlock: %d", nextBlock)
		} else {
			log.Printf("failed to get last fetched log block, err: %+v", err)
		}
	}
}

// return block number that we just fetched the logs
func (self *Fetcher) FetchLogs(fromBlock uint64, toBlock uint64, timepoint uint64) uint64 {
	log.Printf("fetching logs data from block %d", fromBlock)
	logs, err := self.blockchain.GetLogs(fromBlock, toBlock, self.GetEthRate(common.GetTimepoint()))
	if err != nil {
		log.Printf("fetching logs data from block %d failed, error: %v", fromBlock, err)
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
					log.Printf("blockno: %d - %d", l.BlockNumber, l.TransactionIndex)
					err = self.storage.StoreTradeLog(l, timepoint)
					if err != nil {
						log.Printf("storing trade log failed, abort storing process and return latest stored log block number, err: %+v", err)
						return l.BlockNumber
					} else {
						self.aggregateTradeLog(l)
					}
				} else if il.Type() == "SetCatLog" {
					l := il.(common.SetCatLog)
					log.Printf("blockno: %d", l.BlockNumber)
					log.Printf("log: %+v", l)
					err = self.storage.StoreCatLog(l)
					if err != nil {
						log.Printf("storing cat log failed, abort storing process and return latest stored log block number, err: %+v", err)
						return l.BlockNumber
					}
				}
			}
			return logs[len(logs)-1].BlockNo()
		} else {
			return fromBlock - 1
		}
	}
}

func (self *Fetcher) aggregateTradeLog(trade common.TradeLog) (err error) {
	srcAddr := common.AddrToString(trade.SrcAddress)
	dstAddr := common.AddrToString(trade.DestAddress)
	reserveAddr := common.AddrToString(trade.ReserveAddress)
	walletAddr := common.AddrToString(trade.WalletAddress)
	userAddr := common.AddrToString(trade.UserAddress)

	var srcAmount, destAmount, ethAmount, burnFee, walletFee float64
	var tokenAddr string
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

	eth := common.SupportedTokens["ETH"]
	if trade.BurnFee != nil {
		burnFee = common.BigToFloat(trade.BurnFee, eth.Decimal)
	}
	if trade.WalletFee != nil {
		walletFee = common.BigToFloat(trade.WalletFee, eth.Decimal)
	}

	updates := common.TradeStats{
		strings.Join([]string{"assets_volume", srcAddr}, "_"):              srcAmount,
		strings.Join([]string{"assets_volume", dstAddr}, "_"):              destAmount,
		strings.Join([]string{"assets_eth_amount", tokenAddr}, "_"):        ethAmount,
		strings.Join([]string{"assets_usd_amount", srcAddr}, "_"):          trade.FiatAmount,
		strings.Join([]string{"assets_usd_amount", dstAddr}, "_"):          trade.FiatAmount,
		strings.Join([]string{"burn_fee", reserveAddr}, "_"):               burnFee,
		strings.Join([]string{"wallet_fee", reserveAddr, walletAddr}, "_"): walletFee,
		strings.Join([]string{"user_volume", userAddr}, "_"):               trade.FiatAmount,
	}

	for _, freq := range []string{"M", "H", "D"} {
		err = self.storage.SetTradeStats(freq, trade.Timestamp, updates)
		if err != nil {
			return
		}
	}

	// total stats on trading
	updates = common.TradeStats{
		"eth_volume":  ethAmount,
		"usd_volume":  trade.FiatAmount,
		"burn_fee":    burnFee,
		"trade_count": 1,
	}
	err = self.storage.SetTradeStats("D", trade.Timestamp, updates)
	if err != nil {
		return
	}

	// stats on user
	user_stats, err := self.storage.SaveUserAddress(trade.Timestamp, userAddr)
	err = self.storage.SetTradeStats("D", trade.Timestamp, user_stats)
	if err != nil {
		return
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
