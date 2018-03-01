package intermediator

import (
	"time"
)

// Runner to trigger fetcher
type IntermediatorRunner interface {
	GetStatusTicker() <-chan time.Time
	// Start must be non-blocking and must only return after runner
	// gets to ready state before GetOrderbookTicker() and
	// GetAuthDataTicker() get called
	Start() error
	// Stop should only be invoked when the runner is already running
	Stop() error
}

type TickerRunner struct {
	sduration time.Duration
	sclock    *time.Ticker
	signal    chan bool
}

func (self *TickerRunner) GetStatusTicker() <-chan time.Time {
	if self.sclock == nil {
		<-self.signal
	}
	return self.sclock.C
}

func (self *TickerRunner) Start() error {
	self.sclock = time.NewTicker(self.sduration)
	self.signal <- true
	return nil
}

func (self *TickerRunner) Stop() error {
	self.sclock.Stop()
	return nil
}

func NewTickerRunner(sduration time.Duration) *TickerRunner {
	return &TickerRunner{
		sduration,
		nil,
		make(chan bool, 5),
	}
}
