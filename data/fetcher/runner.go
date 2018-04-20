package fetcher

import (
	"time"
)

// Runner to trigger fetcher
type FetcherRunner interface {
	GetOrderbookTicker() <-chan time.Time
	GetAuthDataTicker() <-chan time.Time
	GetRateTicker() <-chan time.Time
	GetBlockTicker() <-chan time.Time
	GetTradeHistoryTicker() <-chan time.Time
	// Start must be non-blocking and must only return after runner
	// gets to ready state before GetOrderbookTicker() and
	// GetAuthDataTicker() get called
	Start() error
	// Stop should only be invoked when the runner is already running
	Stop() error
}

type TickerRunner struct {
	oduration                 time.Duration
	aduration                 time.Duration
	rduration                 time.Duration
	bduration                 time.Duration
	tduration                 time.Duration
	rsduration                time.Duration
	lduration                 time.Duration
	tradeLogProcessorDuration time.Duration
	catLogProcessorDuration   time.Duration
	oclock                    *time.Ticker
	aclock                    *time.Ticker
	rclock                    *time.Ticker
	bclock                    *time.Ticker
	tclock                    *time.Ticker
	rsclock                   *time.Ticker
	lclock                    *time.Ticker
	tradeLogProcessorClock    *time.Ticker
	catLogProcessorClock      *time.Ticker
	signal                    chan bool
}

// func (self *TickerRunner) GetTradeLogProcessorTicker() <-chan time.Time {
// 	if self.tradeLogProcessorClock == nil {
// 		<-self.signal
// 	}
// 	return self.tradeLogProcessorClock.C
// }

// func (self *TickerRunner) GetCatLogProcessorTicker() <-chan time.Time {
// 	if self.catLogProcessorClock == nil {
// 		<-self.signal
// 	}
// 	return self.catLogProcessorClock.C
// }

// func (self *TickerRunner) GetLogTicker() <-chan time.Time {
// 	if self.lclock == nil {
// 		<-self.signal
// 	}
// 	return self.lclock.C
// }

func (self *TickerRunner) GetBlockTicker() <-chan time.Time {
	if self.bclock == nil {
		<-self.signal
	}
	return self.bclock.C
}
func (self *TickerRunner) GetOrderbookTicker() <-chan time.Time {
	if self.oclock == nil {
		<-self.signal
	}
	return self.oclock.C
}
func (self *TickerRunner) GetAuthDataTicker() <-chan time.Time {
	if self.aclock == nil {
		<-self.signal
	}
	return self.aclock.C
}
func (self *TickerRunner) GetRateTicker() <-chan time.Time {
	if self.rclock == nil {
		<-self.signal
	}
	return self.rclock.C
}
func (self *TickerRunner) GetTradeHistoryTicker() <-chan time.Time {
	if self.tclock == nil {
		<-self.signal
	}
	return self.tclock.C
}

// func (self *TickerRunner) GetReserveRatesTicker() <-chan time.Time {
// 	if self.rsclock == nil {
// 		<-self.signal
// 	}
// 	return self.rsclock.C
// }

func (self *TickerRunner) Start() error {
	self.oclock = time.NewTicker(self.oduration)
	self.signal <- true
	self.aclock = time.NewTicker(self.aduration)
	self.signal <- true
	self.rclock = time.NewTicker(self.rduration)
	self.signal <- true
	self.bclock = time.NewTicker(self.bduration)
	self.signal <- true
	self.tclock = time.NewTicker(self.tduration)
	self.signal <- true
	self.rsclock = time.NewTicker(self.rsduration)
	self.signal <- true
	self.lclock = time.NewTicker(self.lduration)
	self.signal <- true
	self.tradeLogProcessorClock = time.NewTicker(self.tradeLogProcessorDuration)
	self.signal <- true
	self.catLogProcessorClock = time.NewTicker(self.catLogProcessorDuration)
	self.signal <- true
	return nil
}

func (self *TickerRunner) Stop() error {
	self.oclock.Stop()
	self.aclock.Stop()
	self.rclock.Stop()
	self.bclock.Stop()
	self.tclock.Stop()
	self.rsclock.Stop()
	self.lclock.Stop()
	self.tradeLogProcessorClock.Stop()
	self.catLogProcessorClock.Stop()
	return nil
}

func NewTickerRunner(
	oduration, aduration, rduration,
	bduration, tduration, rsduration,
	lduration,
	tradeLogProcessorDuration,
	catLogProcessorDuration time.Duration) *TickerRunner {
	return &TickerRunner{
		oduration,
		aduration,
		rduration,
		bduration,
		tduration,
		rsduration,
		lduration,
		tradeLogProcessorDuration,
		catLogProcessorDuration,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		make(chan bool, 9),
	}
}
