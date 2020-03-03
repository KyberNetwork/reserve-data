package fetcher

import (
	"github.com/KyberNetwork/reserve-data/common"
)

// RunRateFetcher start rate fetcher job
func (f *Fetcher) RunRateFetcher() {
	for {
		f.l.Infof("waiting for signal from runner rate channel")
		t := <-f.runner.GetRateTicker()
		f.l.Infof("got signal in rate channel with timestamp %d", common.TimeToMillis(t))
		f.FetchRate(common.TimeToMillis(t))
		f.l.Infof("fetched rates from blockchain")
	}
}

// FetchRate return rate for fetcher
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
