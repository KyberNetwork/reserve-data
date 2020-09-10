package gasinfo

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data"
	gaspricedata "github.com/KyberNetwork/reserve-data/common/gaspricedata-client"
)

type GasPriceInfo struct {
	limiter   GasPriceLimiter
	storage   reserve.GasConfig
	gasClient gaspricedata.Client
	l         *zap.SugaredLogger
}

var (
	global *GasPriceInfo
	lock   sync.Mutex
)

// GetGlobal this global instance is using in huobi withdraw, where exchange pool init very soon and gasPrice instance
// not exist at that time.
func GetGlobal() *GasPriceInfo {
	lock.Lock()
	defer lock.Unlock()
	return global
}

// SetGlobal set global instance, will be set right after we have it.
func SetGlobal(instance *GasPriceInfo) {
	lock.Lock()
	defer lock.Unlock()
	global = instance
}

// NewGasPriceInfo create new gasPrice selector
func NewGasPriceInfo(limiter GasPriceLimiter, store reserve.GasConfig, gasClient gaspricedata.Client) *GasPriceInfo {
	return &GasPriceInfo{
		limiter:   limiter,
		storage:   store,
		gasClient: gasClient,
		l:         zap.S(),
	}
}

// MaxGas return max gas in gwei
func (g *GasPriceInfo) MaxGas() (float64, error) {
	return g.limiter.MaxGasPrice()
}

// GetCurrentGas return gas price from preferred source
func (g *GasPriceInfo) GetCurrentGas() (float64, error) {
	selectedSource := "ethgasstation"
	selected, err := g.storage.GetPreferGasSource()
	if err == nil {
		selectedSource = selected.Name
	} else {
		g.l.Errorw("failed to receive selected source, use default", "default", selectedSource)
	}
	allGas, err := g.gasClient.GetGas()
	if err != nil {
		return 0, err
	}
	gas, ok := allGas[selectedSource]
	if !ok {
		g.l.Errorw("selected source not found in result", "source", selectedSource)
		return 0, fmt.Errorf("selected gas source not found in result, source=%s", selectedSource)
	}
	g.l.Infow("got gas price", "source", selectedSource, "value", gas.Value)
	gTime := time.Unix(gas.Timestamp/1000, 0)
	gasUpdateSecondsRequire := 600.0
	seconds := time.Since(gTime).Seconds()
	if seconds > gasUpdateSecondsRequire {
		return 0, fmt.Errorf("gas price need to be update, current_version %s which is %f seconds ago", gTime, seconds)
	}
	return gas.Value.Fast, nil
}

// AllSourceGas return all supported source gas price
func (g *GasPriceInfo) AllSourceGas() (gaspricedata.GasResult, error) {
	return g.gasClient.GetGas()
}
