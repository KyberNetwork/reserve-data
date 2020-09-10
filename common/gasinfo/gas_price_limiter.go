package gasinfo

import (
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data/blockchain"
	"github.com/KyberNetwork/reserve-data/common"
)

type GasPriceLimiter interface {
	MaxGasPrice() (float64, error)
}

// ConstGasPriceLimiter use in test only
type ConstGasPriceLimiter struct {
}

func (c ConstGasPriceLimiter) MaxGasPrice() (float64, error) {
	return 100.0, nil
}

// KyberGasPriceLimiter manage gas price limit
type KyberGasPriceLimiter struct {
	p               *blockchain.NetworkProxy
	localCachedTime int64
	maxCacheSeconds int64
	cachedValue     *float64
	guard           sync.Mutex
}

// NewNetworkGasPriceLimiter create a new gas price limiter
func NewNetworkGasPriceLimiter(kb *blockchain.NetworkProxy, cacheSeconds int64) GasPriceLimiter {
	return &KyberGasPriceLimiter{p: kb, maxCacheSeconds: cacheSeconds}
}

// MaxGasPrice get max gas price, use cache if possible
func (c *KyberGasPriceLimiter) MaxGasPrice() (float64, error) {
	now := time.Now().Unix()
	c.guard.Lock()
	defer c.guard.Unlock()
	if c.maxCacheSeconds > 0 && now-c.localCachedTime <= c.maxCacheSeconds && c.cachedValue != nil {
		return *c.cachedValue, nil
	}
	v, err := c.p.MaxGasPrice(&bind.CallOpts{})
	if err == nil {
		c.localCachedTime = now
		fValue := common.BigToFloat(v, 9)
		c.cachedValue = &fValue
		zap.S().Infow("fetch new max_gas_price", "value", fValue)
		return fValue, nil
	}
	return 0, err
}
