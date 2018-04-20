package stat

import (
	"time"
)

type ControllerRunner interface {
	GetAnalyticStorageControlTicker() <-chan time.Time
	Start() error
	Stop() error
}
