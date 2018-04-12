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
