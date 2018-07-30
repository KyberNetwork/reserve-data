package datapruner

import (
	"time"
)

// StorageControllerRunner is the controller interface of data pruner jobs.
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

func (ctr *ControllerTickerRunner) GetAuthBucketTicker() <-chan time.Time {
	if ctr.authClock == nil {
		<-ctr.signal
	}
	return ctr.authClock.C
}

func (ctr *ControllerTickerRunner) Start() error {
	ctr.authClock = time.NewTicker(ctr.authDuration)
	ctr.signal <- true
	return nil
}

func (ctr *ControllerTickerRunner) Stop() error {
	ctr.authClock.Stop()
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
