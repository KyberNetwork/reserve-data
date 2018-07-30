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
}

// TickerRunner is an implementation of FetcherRunner that use simple time ticker.
type TickerRunner struct {
	oduration          time.Duration
	aduration          time.Duration
	rduration          time.Duration
	bduration          time.Duration
	globalDataDuration time.Duration

	oclock          *time.Ticker
	aclock          *time.Ticker
	rclock          *time.Ticker
	bclock          *time.Ticker
	globalDataClock *time.Ticker
}

func (tr *TickerRunner) GetGlobalDataTicker() <-chan time.Time {
	return tr.globalDataClock.C
}

func (tr *TickerRunner) GetBlockTicker() <-chan time.Time {
	return tr.bclock.C
}
func (tr *TickerRunner) GetOrderbookTicker() <-chan time.Time {
	return tr.oclock.C
}
func (tr *TickerRunner) GetAuthDataTicker() <-chan time.Time {
	return tr.aclock.C
}
func (tr *TickerRunner) GetRateTicker() <-chan time.Time {
	return tr.rclock.C
}

func (tr *TickerRunner) Start() error {
	tr.oclock = time.NewTicker(tr.oduration)
	tr.aclock = time.NewTicker(tr.aduration)
	tr.rclock = time.NewTicker(tr.rduration)
	tr.bclock = time.NewTicker(tr.bduration)
	tr.globalDataClock = time.NewTicker(tr.globalDataDuration)
	return nil
}

func (tr *TickerRunner) Stop() error {
	tr.oclock.Stop()
	tr.aclock.Stop()
	tr.rclock.Stop()
	tr.bclock.Stop()
	tr.globalDataClock.Stop()
	return nil
}

// NewTickerRunner creates a new instance of TickerRunner with given time durations in parameters.
func NewTickerRunner(
	oduration, aduration, rduration,
	bduration, globalDataDuration time.Duration) *TickerRunner {
	return &TickerRunner{
		oduration:          oduration,
		aduration:          aduration,
		rduration:          rduration,
		bduration:          bduration,
		globalDataDuration: globalDataDuration,
	}
}
