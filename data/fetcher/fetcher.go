package fetcher

import (
	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data/common"
)

// maxActivityLifeTime is the longest time of an activity. If the
// activity is pending for more than MAX_ACVITY_LIFE_TIME, it will be
// considered as failed.
const maxActivityLifeTime uint64 = 6 // activity max life time in hour

// Fetcher is job represent the fetcher
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

// NewFetcher return the new fetcher instance
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

// SetBlockchain setup blockchain for fetcher
func (f *Fetcher) SetBlockchain(blockchain Blockchain) {
	f.blockchain = blockchain
	f.FetchCurrentBlock(common.NowInMillis())
}

// AddExchange add exchange to fetcher
func (f *Fetcher) AddExchange(exchange Exchange) {
	f.exchanges = append(f.exchanges, exchange)
}

// Stop stop the fetcher
func (f *Fetcher) Stop() error {
	return f.runner.Stop()
}

// Run start the fetcher
func (f *Fetcher) Run() error {
	f.l.Info("Fetcher runner is starting...")
	if err := f.runner.Start(); err != nil {
		return err
	}
	go f.RunOrderbookFetcher()
	go f.RunAuthDataFetcher()
	go f.RunRateFetcher()
	go f.RunBlockFetcher()
	go f.runGlobalDataFetcher()
	go f.RunFetchExchangeHistory()
	f.l.Infof("Fetcher runner is running...")
	return nil
}

// RunBlockFetcher run block fetcher job which run interval by ticker
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

// FetchCurrentBlock update block time
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
