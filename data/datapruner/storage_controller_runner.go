package datapruner

import (
	"time"
)

type StorageControllerRunner interface {
	GetAuthBucketTicker() <-chan time.Time
	Start() error
	Stop() error
}

type ControllerTickerRunner struct {
	authDuration time.Duration
	authClock    *time.Ticker
	signal       chan bool
}

func (self *ControllerTickerRunner) GetAuthBucketTicker() <-chan time.Time {
	if self.authClock == nil {
		<-self.signal
	}
	return self.authClock.C
}

func (self *ControllerTickerRunner) Start() error {
	self.authClock = time.NewTicker(self.authDuration)
	self.signal <- true
	return nil
}

func (self *ControllerTickerRunner) Stop() error {
	self.authClock.Stop()
	return nil
}

func NewStorageControllerTickerRunner(
	authDuration time.Duration) *ControllerTickerRunner {
	return &ControllerTickerRunner{
		authDuration,
		nil,
		make(chan bool, 1),
	}
}
