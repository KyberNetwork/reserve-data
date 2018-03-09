package stat

import (
	"log"
	"strings"
	// "sync"

	"github.com/KyberNetwork/reserve-data/common"
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
}

func NewFetcher(
	storage Storage,
	ethUSDRate EthUSDRate,
	runner FetcherRunner,
	deployBlock uint64) *Fetcher {
	return &Fetcher{
		storage:     storage,
		blockchain:  nil,
		runner:      runner,
		ethRate:     ethUSDRate,
		deployBlock: deployBlock,
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
	log.Printf("Fetcher runner is running...")
	return nil
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

	walletFeeKey := strings.Join([]string{reserveAddr, walletAddr}, "_")

	var srcAmount, destAmount, burnFee, walletFee float64
	for _, token := range common.SupportedTokens {
		if strings.ToLower(token.Address) == srcAddr {
			srcAmount = common.BigToFloat(trade.SrcAmount, token.Decimal)
		}

		if strings.ToLower(token.Address) == dstAddr {
			destAmount = common.BigToFloat(trade.DestAmount, token.Decimal)
		}
	}

	eth := common.SupportedTokens["ETH"]
	if trade.BurnFee != nil {
		burnFee = common.BigToFloat(trade.BurnFee, eth.Decimal)
	}
	if trade.WalletFee != nil {
		walletFee = common.BigToFloat(trade.WalletFee, eth.Decimal)
	}

	updates := []struct {
		metric     string
		tradeStats common.TradeStats
	}{
		{
			"assets_volume",
			common.TradeStats{
				srcAddr: srcAmount,
				dstAddr: destAmount,
			},
		},
		{
			"burn_fee",
			common.TradeStats{
				reserveAddr: burnFee,
			},
		},
		{
			"wallet_fee",
			common.TradeStats{
				walletFeeKey: walletFee,
			},
		},
		{
			"user_volume",
			common.TradeStats{
				userAddr: trade.FiatAmount,
			},
		},
	}
	for _, update := range updates {
		for _, freq := range []string{"M", "H", "D"} {
			err = self.storage.SetTradeStats(update.metric, freq, trade.Timestamp, update.tradeStats)
			if err != nil {
				return
			}
		}
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
