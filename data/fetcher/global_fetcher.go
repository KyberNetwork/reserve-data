package fetcher

import (
	"github.com/KyberNetwork/reserve-data/common"
)

func (f *Fetcher) runGlobalDataFetcher() {
	for {
		f.l.Info("waiting for signal from global data channel")
		t := <-f.runner.GetGlobalDataTicker()
		f.l.Infof("got signal in global data channel with timestamp %d", common.TimeToMillis(t))
		timepoint := common.TimeToMillis(t)
		f.fetchGlobalData(timepoint)
		f.l.Info("fetched block from blockchain")
	}
}

func (f *Fetcher) fetchGlobalData(timepoint uint64) {
	goldData, err := f.theworld.GetGoldInfo()
	if err != nil {
		f.l.Infof("failed to fetch Gold Info: %s", err.Error())
		return
	}
	goldData.Timestamp = common.NowInMillis()

	if err = f.globalStorage.StoreGoldInfo(goldData); err != nil {
		f.l.Infof("Storing gold info failed: %s", err.Error())
	}

	btcData, err := f.theworld.GetBTCInfo()
	if err != nil {
		f.l.Infof("failed to fetch BTC Info: %s", err.Error())
		return
	}
	btcData.Timestamp = common.NowInMillis()
	if err = f.globalStorage.StoreBTCInfo(btcData); err != nil {
		f.l.Infof("Storing BTC info failed: %s", err.Error())
	}

	usdData, err := f.theworld.GetUSDInfo()
	if err != nil {
		f.l.Warnw("failed to fetch USD info", "err", err)
		return
	}
	usdData.Timestamp = common.NowInMillis()
	if err = f.globalStorage.StoreUSDInfo(usdData); err != nil {
		f.l.Warnw("Store USD info failed", "err", err)
	}
}
