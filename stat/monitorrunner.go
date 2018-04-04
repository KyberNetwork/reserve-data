package stat

import (
	"time"
)

type MonitorRunner interface {
	GetAnalyticStorageControlTicker() <-chan time.Time
	Start() error
	Stop() error
}

type TickerRunner struct {
	ascduration time.Duration
	ascclock    *time.Ticker
	signal      chan bool
}

func (self *TickerRunner) GetAnalyticStorageControlTicker() <-chan time.Time {
	if self.ascclock == nil {
		<-self.signal
	}
	return self.ascclock.C
}

func (self *TickerRunner) Start() error {
	self.ascclock = time.NewTicker(self.ascduration)
	self.signal <- true
	return nil
}

func (self *TickerRunner) Stop() error {
	self.ascclock.Stop()
	return nil
}

func NewTickerRunner(
	ascduration time.Duration) *TickerRunner {
	return &TickerRunner{
		ascduration,
		nil,
		make(chan bool, 1),
	}
}
