package stat

import (
	"time"
)

type ControllerRunner interface {
	GetAnalyticStorageControlTicker() <-chan time.Time
	Start() error
	Stop() error
}

type ControllerTickerRunner struct {
	ascduration time.Duration
	ascclock    *time.Ticker
	signal      chan bool
}

func (self *ControllerTickerRunner) GetAnalyticStorageControlTicker() <-chan time.Time {
	if self.ascclock == nil {
		<-self.signal
	}
	return self.ascclock.C
}

func (self *ControllerTickerRunner) Start() error {
	self.ascclock = time.NewTicker(self.ascduration)
	self.signal <- true
	return nil
}

func (self *ControllerTickerRunner) Stop() error {
	self.ascclock.Stop()
	return nil
}

func NewControllerTickerRunner(
	ascduration time.Duration) *ControllerTickerRunner {
	return &ControllerTickerRunner{
		ascduration,
		nil,
		make(chan bool, 1),
	}
}
