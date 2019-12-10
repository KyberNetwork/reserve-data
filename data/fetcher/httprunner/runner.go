package httprunner

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// HTTPRunner is an implementation of FetcherRunner
// that run a HTTP server and tick when it receives request to a certain endpoints.
type HTTPRunner struct {
	bindAddr string

	oticker          chan time.Time
	aticker          chan time.Time
	rticker          chan time.Time
	bticker          chan time.Time
	globalDataTicker chan time.Time

	server *Server
	l      *zap.SugaredLogger
}

// GetGlobalDataTicker returns the global data ticker.
func (h *HTTPRunner) GetGlobalDataTicker() <-chan time.Time {
	return h.globalDataTicker
}

// GetBlockTicker returns the block ticker.
func (h *HTTPRunner) GetBlockTicker() <-chan time.Time {
	return h.bticker
}

// GetOrderbookTicker returns the order book ticker.
func (h *HTTPRunner) GetOrderbookTicker() <-chan time.Time {
	return h.oticker
}

// GetAuthDataTicker returns the auth data ticker.
func (h *HTTPRunner) GetAuthDataTicker() <-chan time.Time {
	return h.aticker
}

// GetRateTicker returns the rate ticker.
func (h *HTTPRunner) GetRateTicker() <-chan time.Time {
	return h.rticker
}

// waitPingResponse waits until HTTP ticker server responses to request.
func (h *HTTPRunner) waitPingResponse() error {
	var (
		tickCh   = time.NewTicker(time.Second / 2).C
		expireCh = time.NewTicker(time.Second * 5).C
		client   = http.Client{Timeout: time.Second}
	)

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/%s", h.bindAddr, "ping"), nil)
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
				h.l.Warnw("HTTP server is returning an error, retrying", "err", dErr)
				break
			}
			if rsp.StatusCode == http.StatusOK {
				h.l.Infof("HTTP ticker server is ready")
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
func (h *HTTPRunner) Start() error {
	if h.server != nil {
		return errors.New("runner start already")
	}
	h.server = NewServer(h, h.bindAddr)

	go func() {
		if err := h.server.Start(); err != nil {
			h.l.Fatalw("Http server for runner couldn't start or get stopped.", "err", err)
		}
	}()

	// wait until the HTTP server is ready
	<-h.server.notifyCh
	return h.waitPingResponse()
}

// Stop stops the HTTP server. It returns an error if the server is already stopped.
func (h *HTTPRunner) Stop() error {
	if h.server != nil {
		err := h.server.Stop()
		h.server = nil
		return err
	}
	return errors.New("runner stop already")
}

// Option is the option to setup the HTTPRunner on creation.
type Option func(hr *HTTPRunner)

// WithBindAddr setups the HTTPRunner instance with the given bindAddr.
// Without this option, NewHTTPRunner will use a random bindAddr.
func WithBindAddr(bindAddr string) Option {
	return func(hr *HTTPRunner) {
		hr.bindAddr = bindAddr
	}
}

// NewHTTPRunner creates a new instance of HTTPRunner.
func NewHTTPRunner(options ...Option) (*HTTPRunner, error) {
	ochan := make(chan time.Time)
	achan := make(chan time.Time)
	rchan := make(chan time.Time)
	bchan := make(chan time.Time)
	globalDataChan := make(chan time.Time)

	runner := &HTTPRunner{
		oticker:          ochan,
		aticker:          achan,
		rticker:          rchan,
		bticker:          bchan,
		globalDataTicker: globalDataChan,
		server:           nil,
		l:                zap.S(),
	}

	for _, option := range options {
		option(runner)
	}
	return runner, nil
}
