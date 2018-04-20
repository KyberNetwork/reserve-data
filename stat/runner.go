package stat

import (
	"time"
)

type FetcherRunner interface {
	GetBlockTicker() <-chan time.Time
	GetLogTicker() <-chan time.Time
	GetReserveRatesTicker() <-chan time.Time
	GetTradeLogProcessorTicker() <-chan time.Time
	GetCatLogProcessorTicker() <-chan time.Time
	Start() error
	Stop() error
}

type TickerRunner struct {
	blockDuration       time.Duration
	logDuration         time.Duration
	rateDuration        time.Duration
	tlogProcessDuration time.Duration
	clogProcessDuration time.Duration
	ascduration         time.Duration
	blockClock          *time.Ticker
	logClock            *time.Ticker
	rateClock           *time.Ticker
	tlogProcessClock    *time.Ticker
	clogProcessClock    *time.Ticker
	ascclock            *time.Ticker
	signal              chan bool
}

func (self *TickerRunner) GetBlockTicker() <-chan time.Time {
	if self.blockClock == nil {
		<-self.signal
	}
	return self.blockClock.C
}

func (self *TickerRunner) GetLogTicker() <-chan time.Time {
	if self.logClock == nil {
		<-self.signal
	}
	return self.logClock.C
}

func (self *TickerRunner) GetReserveRatesTicker() <-chan time.Time {
	if self.rateClock == nil {
		<-self.signal
	}
	return self.rateClock.C
}

func (self *TickerRunner) GetTradeLogProcessorTicker() <-chan time.Time {
	if self.tlogProcessClock == nil {
		<-self.signal
	}
	return self.tlogProcessClock.C
}

func (self *TickerRunner) GetCatLogProcessorTicker() <-chan time.Time {
	if self.clogProcessClock == nil {
		<-self.signal
	}
	return self.clogProcessClock.C
}

func (self *TickerRunner) GetAnalyticStorageControlTicker() <-chan time.Time {
	if self.ascclock == nil {
		<-self.signal
	}
	return self.ascclock.C
}

func (self *TickerRunner) Start() error {
	self.blockClock = time.NewTicker(self.blockDuration)
	self.signal <- true
	self.logClock = time.NewTicker(self.logDuration)
	self.signal <- true
	self.rateClock = time.NewTicker(self.rateDuration)
	self.signal <- true
	self.tlogProcessClock = time.NewTicker(self.tlogProcessDuration)
	self.signal <- true
	self.clogProcessClock = time.NewTicker(self.clogProcessDuration)
	self.signal <- true
	self.ascclock = time.NewTicker(self.ascduration)
	self.signal <- true
	return nil
}

func (self *TickerRunner) Stop() error {
	self.blockClock.Stop()
	self.logClock.Stop()
	self.rateClock.Stop()
	self.tlogProcessClock.Stop()
	self.clogProcessClock.Stop()
	self.ascclock.Stop()
	return nil
}

func NewTickerRunner(
	blockDuration, logDuration, rateDuration, tlogProcessDuration, clogProcessDuration, ascduration time.Duration) *TickerRunner {
	return &TickerRunner{
		blockDuration,
		logDuration,
		rateDuration,
		tlogProcessDuration,
		clogProcessDuration,
		ascduration,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		make(chan bool, 6),
	}
}
