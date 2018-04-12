package http_runner

import (
	"errors"
	"fmt"
	"log"
	"time"
)

type HttpRunner struct {
	port                    int
	oticker                 chan time.Time
	aticker                 chan time.Time
	rticker                 chan time.Time
	bticker                 chan time.Time
	tticker                 chan time.Time
	rsticker                chan time.Time
	lticker                 chan time.Time
	tradeLogProcessorTicker chan time.Time
	catLogProcessorTicker   chan time.Time
	server                  *HttpRunnerServer
}

func (self *HttpRunner) GetTradeLogProcessorTicker() <-chan time.Time {
	return self.tradeLogProcessorTicker
}

func (self *HttpRunner) GetCatLogProcessorTicker() <-chan time.Time {
	return self.catLogProcessorTicker
}

func (self *HttpRunner) GetLogTicker() <-chan time.Time {
	return self.lticker
}

func (self *HttpRunner) GetBlockTicker() <-chan time.Time {
	return self.bticker
}

func (self *HttpRunner) GetOrderbookTicker() <-chan time.Time {
	return self.oticker
}

func (self *HttpRunner) GetAuthDataTicker() <-chan time.Time {
	return self.aticker
}

func (self *HttpRunner) GetRateTicker() <-chan time.Time {
	return self.rticker
}
func (self *HttpRunner) GetTradeHistoryTicker() <-chan time.Time {
	return self.tticker
}

func (self *HttpRunner) GetReserveRatesTicker() <-chan time.Time {
	return self.rsticker
}

func (self *HttpRunner) Start() error {
	if self.server != nil {
		return errors.New("runner start already")
	} else {
		self.server = NewHttpRunnerServer(self, fmt.Sprintf(":%d", self.port))
		go func() {
			err := self.server.Start()
			if err != nil {
				log.Printf("Http server for runner couldn't start or get stopped. Error: %s", err)
			}
		}()
		return nil
	}
}

func (self *HttpRunner) Stop() error {
	if self.server != nil {
		err := self.server.Stop()
		self.server = nil
		return err
	} else {
		return errors.New("runner stop already")
	}
}

func NewHttpRunner(port int) *HttpRunner {
	ochan := make(chan time.Time)
	achan := make(chan time.Time)
	rchan := make(chan time.Time)
	bchan := make(chan time.Time)
	tchan := make(chan time.Time)
	rschan := make(chan time.Time)
	lchan := make(chan time.Time)
	tradeLogProcessorChan := make(chan time.Time)
	catLogProcessorChan := make(chan time.Time)
	runner := HttpRunner{
		port,
		ochan,
		achan,
		rchan,
		bchan,
		tchan,
		rschan,
		lchan,
		tradeLogProcessorChan,
		catLogProcessorChan,
		nil,
	}
	runner.Start()
	return &runner
}
