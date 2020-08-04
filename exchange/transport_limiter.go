package exchange

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"
)

var (
	binance418 = regexp.MustCompile(`banned\suntil\s(\d+)`)
)

// TransportRateLimiter a workaround rate limiter, currently for binance
type TransportRateLimiter struct {
	blockUntil time.Time
	lock       sync.RWMutex
	c          *http.Client
	l          *zap.SugaredLogger
}

// NewTransportRateLimiter create a new rate limiter
func NewTransportRateLimiter(c *http.Client) *TransportRateLimiter {
	return &TransportRateLimiter{
		c: c,
		l: zap.S(),
	}
}
func (b *TransportRateLimiter) roundTripBinance(request *http.Request) (*http.Response, error) {
	now := time.Now()
	b.lock.RLock()
	blockUntil := b.blockUntil
	b.lock.RUnlock()
	if now.Before(blockUntil) {
		return &http.Response{StatusCode: http.StatusTeapot, Body: ioutil.NopCloser(bytes.NewBuffer(nil))}, fmt.Errorf("rate limit guard, until %s", blockUntil.String())
	}
	resp, err := b.c.Do(request)
	if err != nil {
		return resp, err
	}
	if resp.StatusCode != http.StatusTeapot {
		return resp, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	if err != nil {
		return resp, err
	}

	// binance return Retry-After in second, and a timestamp in message,
	// I want to use timestamp in the message as it easier to calculate, but it may change in future
	// so we handle both, use the greater value.

	nextReq := now
	after := resp.Header.Get("Retry-After")
	if after != "" {
		sec, _ := strconv.ParseInt(after, 10, 0)
		nextReq = nextReq.Add(time.Second * time.Duration(sec))
	}
	allMatch := binance418.FindAllStringSubmatch(string(body), -1)
	if len(allMatch) == 0 {
		b.l.Errorw("found binance 418 status code, but can't match 'block until'", "body", string(body))
	} else {
		until, _ := strconv.ParseInt(allMatch[0][1], 10, 0)
		sec := until / 1000
		ms := until - sec*1000
		untilTime := time.Unix(sec, ms*1000000)
		if untilTime.After(nextReq) {
			nextReq = untilTime
		}
	}
	b.l.Infow("set binance block until", "to", nextReq.String())
	b.lock.Lock()
	b.blockUntil = nextReq
	b.lock.Unlock()

	return resp, fmt.Errorf("rate limit guard, until %s", nextReq.String())
}
func (b *TransportRateLimiter) RoundTrip(request *http.Request) (*http.Response, error) {
	if request.URL.Host != "api.binance.com" {
		return b.c.Do(request)
	}
	return b.roundTripBinance(request)
}