package statpruner

import (
	"time"
)

// ControllerRunner contains the tickers that control the pruner services.
type ControllerRunner interface {
	GetAnalyticStorageControlTicker() <-chan time.Time
	GetRateStorageControlTicker() <-chan time.Time
	Start() error
	Stop() error
}

type ControllerTickerRunner struct {
	ascDuration  time.Duration
	rateDuration time.Duration
	ascClock     *time.Ticker
	rateClock    *time.Ticker
	signal       chan bool
}

func (self *ControllerTickerRunner) GetAnalyticStorageControlTicker() <-chan time.Time {
	if self.ascClock == nil {
		<-self.signal
	}
	return self.ascClock.C
}

func (self *ControllerTickerRunner) GetRateStorageControlTicker() <-chan time.Time {
	if self.rateClock == nil {
		<-self.signal
	}
	return self.rateClock.C
}

func (self *ControllerTickerRunner) Start() error {
	self.ascClock = time.NewTicker(self.ascDuration)
	self.signal <- true
	self.rateClock = time.NewTicker(self.rateDuration)
	self.signal <- true
	return nil
}

func (self *ControllerTickerRunner) Stop() error {
	self.rateClock.Stop()
	self.ascClock.Stop()
	return nil
}

func NewControllerTickerRunner(
	ascDuration time.Duration,
	rateDuration time.Duration) *ControllerTickerRunner {
	return &ControllerTickerRunner{
		ascDuration:  ascDuration,
		rateDuration: rateDuration,
		ascClock:     nil,
		rateClock:    nil,
		signal:       make(chan bool, 2),
	}
}
