package http_runner

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
)

// HttpRunner is an implementation of FetcherRunner
// that run a HTTP server and tick when it receives request to a certain endpoints.
type HttpRunner struct {
	port int

	oticker          chan time.Time
	aticker          chan time.Time
	rticker          chan time.Time
	bticker          chan time.Time
	tticker          chan time.Time
	globalDataTicker chan time.Time

	// unused tickers, keep for compatibility
	rsticker                chan time.Time
	lticker                 chan time.Time
	tradeLogProcessorTicker chan time.Time
	catLogProcessorTicker   chan time.Time

	server *HttpRunnerServer
}

// GetGlobalDataTicker returns the global data ticker.
func (self *HttpRunner) GetGlobalDataTicker() <-chan time.Time {
	return self.globalDataTicker
}

// GetTradeLogProcessorTicker returns the trade log processor ticker.
func (self *HttpRunner) GetTradeLogProcessorTicker() <-chan time.Time {
	return self.tradeLogProcessorTicker
}

// GetCatLogProcessorTicker returns the cat log processor ticker.
func (self *HttpRunner) GetCatLogProcessorTicker() <-chan time.Time {
	return self.catLogProcessorTicker
}

// GetLogTicker returns the log ticker.
func (self *HttpRunner) GetLogTicker() <-chan time.Time {
	return self.lticker
}

// GetBlockTicker returns the block ticker.
func (self *HttpRunner) GetBlockTicker() <-chan time.Time {
	return self.bticker
}

// GetOrderbookTicker returns the order book ticker.
func (self *HttpRunner) GetOrderbookTicker() <-chan time.Time {
	return self.oticker
}

// GetAuthDataTicker returns the auth data ticker.
func (self *HttpRunner) GetAuthDataTicker() <-chan time.Time {
	return self.aticker
}

// GetRateTicker returns the rate ticker.
func (self *HttpRunner) GetRateTicker() <-chan time.Time {
	return self.rticker
}

// GetTradeHistoryTicker returns the trade history ticker.
func (self *HttpRunner) GetTradeHistoryTicker() <-chan time.Time {
	return self.tticker
}

// GetReserveRatesTicker returns the reserve rates ticker.
func (self *HttpRunner) GetReserveRatesTicker() <-chan time.Time {
	return self.rsticker
}

// waitPingResponse waits until HTTP ticker server responses to request.
func (self *HttpRunner) waitPingResponse() error {
	var (
		tickCh   = time.NewTicker(time.Second / 2).C
		expireCh = time.NewTicker(time.Second * 5).C
		client   = http.Client{Timeout: time.Second}
	)

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/%s", self.port, "ping"), nil)
	if err != nil {
		return err
	}

	for {
		select {
		case <-expireCh:
			return errors.New("HTTP ticker does not response to ping request")
		case <-tickCh:
			rsp, dErr := client.Do(req)
			if dErr != nil {
				log.Printf("HTTP server is returning an error: %s, retrying", dErr.Error())
				break
			}
			if rsp.StatusCode == http.StatusOK {
				log.Print("HTTP ticker server is ready")
				return nil
			}
		}
	}
}

// Start initializes and starts the ticker HTTP server.
// It returns an error if the server is started already.
// It is guaranteed that the HTTP server is ready to serve request after
// this method is returned.
// The HTTP server is listened on all network interfaces.
func (self *HttpRunner) Start() error {
	if self.server != nil {
		return errors.New("runner start already")
	} else {
		var addr string
		if self.port != 0 {
			addr = fmt.Sprintf(":%d", self.port)
		}
		self.server = NewHttpRunnerServer(self, addr)
		go func() {
			if err := self.server.Start(); err != nil {
				log.Printf("Http server for runner couldn't start or get stopped. Error: %s", err)
			}
		}()

		// wait until the HTTP server is ready
		<-self.server.notifyCh
		return self.waitPingResponse()
	}
}

// Stop stops the HTTP server. It returns an error if the server is already stopped.
func (self *HttpRunner) Stop() error {
	if self.server != nil {
		err := self.server.Stop()
		self.server = nil
		return err
	} else {
		return errors.New("runner stop already")
	}
}

// HttpRunnerOption is the option to setup the HttpRunner on creation.
type HttpRunnerOption func(hr *HttpRunner)

// WithHttpRunnerPort setups the HttpRunner instance with the given port.
// Without this option, NewHttpRunner will use a random port.
func WithHttpRunnerPort(port int) HttpRunnerOption {
	return func(hr *HttpRunner) {
		hr.port = port
	}
}

// NewHttpRunner creates a new instance of HttpRunner.
// The HTTP server is also started after creation.
func NewHttpRunner(options ...HttpRunnerOption) (*HttpRunner, error) {
	ochan := make(chan time.Time)
	achan := make(chan time.Time)
	rchan := make(chan time.Time)
	bchan := make(chan time.Time)
	tchan := make(chan time.Time)
	rschan := make(chan time.Time)
	lchan := make(chan time.Time)
	tradeLogProcessorChan := make(chan time.Time)
	catLogProcessorChan := make(chan time.Time)
	globalDataChan := make(chan time.Time)

	runner := &HttpRunner{
		oticker:                 ochan,
		aticker:                 achan,
		rticker:                 rchan,
		bticker:                 bchan,
		tticker:                 tchan,
		rsticker:                rschan,
		lticker:                 lchan,
		tradeLogProcessorTicker: tradeLogProcessorChan,
		catLogProcessorTicker:   catLogProcessorChan,
		globalDataTicker:        globalDataChan,
		server:                  nil,
	}

	for _, option := range options {
		option(runner)
	}

	if err := runner.Start(); err != nil {
		return nil, err
	}
	return runner, nil
}
