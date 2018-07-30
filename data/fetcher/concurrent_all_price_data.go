package fetcher

import (
	"sync"

	"github.com/KyberNetwork/reserve-data/common"
)

type ConcurrentAllPriceData struct {
	mu   sync.RWMutex
	data common.AllPriceEntry
}

func NewConcurrentAllPriceData() *ConcurrentAllPriceData {
	return &ConcurrentAllPriceData{
		mu: sync.RWMutex{},
		data: common.AllPriceEntry{
			Data:  map[common.TokenPairID]common.OnePrice{},
			Block: 0,
		},
	}
}

func (cap *ConcurrentAllPriceData) SetBlockNumber(block uint64) {
	cap.mu.Lock()
	defer cap.mu.Unlock()
	cap.data.Block = block
}

func (cap *ConcurrentAllPriceData) SetOnePrice(
	exchange common.ExchangeID,
	pair common.TokenPairID,
	d common.ExchangePrice) {
	cap.mu.Lock()
	defer cap.mu.Unlock()
	_, exist := cap.data.Data[pair]
	if !exist {
		cap.data.Data[pair] = common.OnePrice{}
	}
	cap.data.Data[pair][exchange] = d
}

func (cap *ConcurrentAllPriceData) GetData() common.AllPriceEntry {
	cap.mu.RLock()
	defer cap.mu.RUnlock()
	return cap.data
}
