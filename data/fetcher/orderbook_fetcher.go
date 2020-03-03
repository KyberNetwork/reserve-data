package fetcher

import (
	"sync"

	"github.com/KyberNetwork/reserve-data/common"
)

// RunOrderbookFetcher run order book job
func (f *Fetcher) RunOrderbookFetcher() {
	for {
		f.l.Infof("waiting for signal from runner orderbook channel")
		t := <-f.runner.GetOrderbookTicker()
		f.l.Infof("got signal in orderbook channel with timestamp %d", common.TimeToMillis(t))
		f.FetchOrderbook(common.TimeToMillis(t))
		f.l.Info("fetched data from exchanges")
	}
}

// FetchOrderbook return order book from exchanges
func (f *Fetcher) FetchOrderbook(timepoint uint64) {
	data := NewConcurrentAllPriceData()
	// start fetching
	wait := sync.WaitGroup{}
	for _, exchange := range f.exchanges {
		wait.Add(1)
		go f.fetchPriceFromExchange(&wait, exchange, data, timepoint)
	}
	wait.Wait()
	data.SetBlockNumber(f.currentBlock)
	err := f.storage.StorePrice(data.GetData(), timepoint)
	if err != nil {
		f.l.Warnw("Storing data failed", "err", err)
	}
}

func (f *Fetcher) fetchPriceFromExchange(wg *sync.WaitGroup, exchange Exchange, data *ConcurrentAllPriceData, timepoint uint64) {
	defer wg.Done()
	exdata, err := exchange.FetchPriceData(timepoint)
	if err != nil {
		f.l.Warnw("Fetching data failed", "exchange", exchange.ID().String(), "err", err)
		return
	}
	for pair, exchangeData := range exdata {
		data.SetOnePrice(exchange.ID(), pair, exchangeData)
	}
}
