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
	globalDataTicker chan time.Time

	// unused tickers, keep for compatibility
	rsticker                chan time.Time
	lticker                 chan time.Time
	tradeLogProcessorTicker chan time.Time
	catLogProcessorTicker   chan time.Time

	server *HttpRunnerServer
}

// GetGlobalDataTicker returns the global data ticker.
func (hr *HttpRunner) GetGlobalDataTicker() <-chan time.Time {
	return hr.globalDataTicker
}

// GetTradeLogProcessorTicker returns the trade log processor ticker.
func (hr *HttpRunner) GetTradeLogProcessorTicker() <-chan time.Time {
	return hr.tradeLogProcessorTicker
}

// GetCatLogProcessorTicker returns the cat log processor ticker.
func (hr *HttpRunner) GetCatLogProcessorTicker() <-chan time.Time {
	return hr.catLogProcessorTicker
}

// GetLogTicker returns the log ticker.
func (hr *HttpRunner) GetLogTicker() <-chan time.Time {
	return hr.lticker
}

// GetBlockTicker returns the block ticker.
func (hr *HttpRunner) GetBlockTicker() <-chan time.Time {
	return hr.bticker
}

// GetOrderbookTicker returns the order book ticker.
func (hr *HttpRunner) GetOrderbookTicker() <-chan time.Time {
	return hr.oticker
}

// GetAuthDataTicker returns the auth data ticker.
func (hr *HttpRunner) GetAuthDataTicker() <-chan time.Time {
	return hr.aticker
}

// GetRateTicker returns the rate ticker.
func (hr *HttpRunner) GetRateTicker() <-chan time.Time {
	return hr.rticker
}

// GetReserveRatesTicker returns the reserve rates ticker.
func (hr *HttpRunner) GetReserveRatesTicker() <-chan time.Time {
	return hr.rsticker
}

// waitPingResponse waits until HTTP ticker server responses to request.
func (hr *HttpRunner) waitPingResponse() error {
	var (
		tickCh   = time.NewTicker(time.Second / 2).C
		expireCh = time.NewTicker(time.Second * 5).C
		client   = http.Client{Timeout: time.Second}
	)

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/%s", hr.port, "ping"), nil)
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
func (hr *HttpRunner) Start() error {
	if hr.server != nil {
		return errors.New("runner start already")
	} else {
		var addr string
		if hr.port != 0 {
			addr = fmt.Sprintf(":%d", hr.port)
		}
		hr.server = NewHttpRunnerServer(hr, addr)
		go func() {
			if err := hr.server.Start(); err != nil {
				log.Printf("Http server for runner couldn't start or get stopped. Error: %s", err)
			}
		}()

		// wait until the HTTP server is ready
		<-hr.server.notifyCh
		return hr.waitPingResponse()
	}
}

// Stop stops the HTTP server. It returns an error if the server is already stopped.
func (hr *HttpRunner) Stop() error {
	if hr.server != nil {
		err := hr.server.Stop()
		hr.server = nil
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
func NewHttpRunner(options ...HttpRunnerOption) (*HttpRunner, error) {
	ochan := make(chan time.Time)
	achan := make(chan time.Time)
	rchan := make(chan time.Time)
	bchan := make(chan time.Time)
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
	return runner, nil
}
