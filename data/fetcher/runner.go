package fetcher

import (
	"time"
)

// FetcherRunner is the common interface of runners that will periodically trigger fetcher jobs.
type FetcherRunner interface {
	// Start initializes all tickers. It must be called before runner is usable.
	Start() error
	// Stop stops all tickers and free usage resources.
	// It must only be called after runner is started.
	Stop() error

	// All following methods should only becalled after Start() is executed
	GetGlobalDataTicker() <-chan time.Time
	GetOrderbookTicker() <-chan time.Time
	GetAuthDataTicker() <-chan time.Time
	GetRateTicker() <-chan time.Time
	GetBlockTicker() <-chan time.Time
	GetTradeHistoryTicker() <-chan time.Time
}

type TickerRunner struct {
	oduration          time.Duration
	aduration          time.Duration
	rduration          time.Duration
	bduration          time.Duration
	tduration          time.Duration
	globalDataDuration time.Duration
	oclock             *time.Ticker
	aclock             *time.Ticker
	rclock             *time.Ticker
	bclock             *time.Ticker
	tclock             *time.Ticker
	globalDataClock    *time.Ticker
}

func (self *TickerRunner) GetGlobalDataTicker() <-chan time.Time {
	return self.globalDataClock.C
}

func (self *TickerRunner) GetBlockTicker() <-chan time.Time {
	return self.bclock.C
}
func (self *TickerRunner) GetOrderbookTicker() <-chan time.Time {
	return self.oclock.C
}
func (self *TickerRunner) GetAuthDataTicker() <-chan time.Time {
	return self.aclock.C
}
func (self *TickerRunner) GetRateTicker() <-chan time.Time {
	return self.rclock.C
}
func (self *TickerRunner) GetTradeHistoryTicker() <-chan time.Time {
	return self.tclock.C
}

func (self *TickerRunner) Start() error {
	self.oclock = time.NewTicker(self.oduration)
	self.aclock = time.NewTicker(self.aduration)
	self.rclock = time.NewTicker(self.rduration)
	self.bclock = time.NewTicker(self.bduration)
	self.tclock = time.NewTicker(self.tduration)
	self.globalDataClock = time.NewTicker(self.globalDataDuration)
	return nil
}

func (self *TickerRunner) Stop() error {
	self.oclock.Stop()
	self.aclock.Stop()
	self.rclock.Stop()
	self.bclock.Stop()
	self.tclock.Stop()
	self.globalDataClock.Stop()
	return nil
}

func NewTickerRunner(
	oduration, aduration, rduration,
	bduration, tduration, globalDataDuration time.Duration) *TickerRunner {
	return &TickerRunner{
		oduration,
		aduration,
		rduration,
		bduration,
		tduration,
		globalDataDuration,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	}
}
