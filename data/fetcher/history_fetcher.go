package fetcher

import (
	"sync"
)

// RunFetchExchangeHistory starts a fetcher to get exchange trade history
func (f *Fetcher) RunFetchExchangeHistory() {
	for ; ; <-f.runner.GetExchangeHistoryTicker() {
		f.l.Info("got signal in orderbook channel with exchange-history")
		f.fetchExchangeTradeHistory()
		f.l.Info("fetched data from exchanges")
	}
}

func (f *Fetcher) fetchExchangeTradeHistory() {
	wait := sync.WaitGroup{}
	for _, exchange := range f.exchanges {
		wait.Add(1)
		go func(exchange Exchange) {
			defer wait.Done()
			exchange.FetchTradeHistory()
		}(exchange)
	}
	wait.Wait()
}
