package fetcher

import (
	"sort"
	"sync"

	"github.com/KyberNetwork/reserve-data/common"
)

const (
	LIMIT_PRICE_NUMBER int = 50
)

type ConcurrentAllPriceData struct {
	mu      sync.RWMutex
	data    map[common.TokenPairID]common.OnePrice
	getable chan bool
}

func NewConcurrentAllPriceData() *ConcurrentAllPriceData {
	return &ConcurrentAllPriceData{
		mu:      sync.RWMutex{},
		data:    map[common.TokenPairID]common.OnePrice{},
		getable: make(chan bool),
	}
}

func (self *ConcurrentAllPriceData) UpdatePrice(
	oldPrice []common.PriceEntry,
	newPrice []common.PriceEntry) []common.PriceEntry {
	for _, price := range newPrice {
		if price.Quantity == 0 {
			// find the rate
			i := sort.Search(len(oldPrice), func(i int) bool { return oldPrice[i].Rate == price.Rate })
			// if exist, remove it
			if i < len(oldPrice) {
				oldPrice = append(oldPrice[:i], oldPrice[i+1:]...)
			}
		} else {
			// insert to the right place
			i := sort.Search(len(oldPrice), func(i int) bool { return oldPrice[i].Rate >= price.Rate })
			if i < len(oldPrice) {
				if oldPrice[i].Rate == price.Rate {
					oldPrice[i] = price
				} else if len(oldPrice) == LIMIT_PRICE_NUMBER {
					n := len(oldPrice)
					oldPrice = append(oldPrice[:n], oldPrice[n+1:]...)
					oldPrice = append(oldPrice[:i], append([]common.PriceEntry{price}, oldPrice[i:]...)...)
				}
			} else if len(oldPrice) < LIMIT_PRICE_NUMBER {
				oldPrice = append(oldPrice[:i], append([]common.PriceEntry{price}, oldPrice[i:]...)...)
			}
		}
	}
	return oldPrice
}

func (self *ConcurrentAllPriceData) SetOnePrice(
	exchange common.ExchangeID,
	pair common.TokenPairID,
	d common.ExchangePrice) {
	self.mu.Lock()
	defer self.mu.Unlock()
	_, exist := self.data.Data[pair]
	if !exist {
		self.data.Data[pair] = common.OnePrice{}
	}
	self.data.Data[pair][exchange] = d
}

func (self *ConcurrentAllPriceData) UpdateOnePrice(
	exchange common.ExchangeID,
	pair common.TokenPairID,
	d common.ExchangePrice) {
	self.mu.Lock()
	defer self.mu.Unlock()
	_, exist := self.data[pair]
	if !exist {
		self.data[pair] = common.OnePrice{}
	}
	exchangePrice := self.data[pair][exchange]
	exchangePrice.Valid = d.Valid
	exchangePrice.Timestamp = d.Timestamp
	exchangePrice.ReturnTime = d.ReturnTime
	exchangePrice.Error = d.Error
	exchangePrice.Bids = self.UpdatePrice(exchangePrice.Bids, d.Bids)
	exchangePrice.Asks = self.UpdatePrice(exchangePrice.Asks, d.Asks)
	self.data[pair][exchange] = exchangePrice
}

func (self *ConcurrentAllPriceData) UnifyData(dataSocket map[common.TokenPairID]common.OnePrice) {
	self.mu.Lock()
	defer self.mu.Unlock()
	for pair, socketData := range dataSocket {
		_, exist := self.data[pair]
		if !exist {
			self.data[pair] = socketData
		} else {
			// insert to data
			for exchangeID, exchangePrice := range socketData {
				self.data[pair][exchangeID] = exchangePrice
			}
		}
	}
}

func (self *ConcurrentAllPriceData) CheckNewSnapShot(exchanges []Exchange) {
	checkedList := map[common.ExchangeID]bool{}
	for _, exchange := range exchanges {
		exchangeID := exchange.ID()
		if _, exist := checkedList[exchangeID]; exist {
			continue
		}
		for _, val := range self.data {
			if _, ok := val[exchangeID]; ok {
				checkedList[exchangeID] = true
				break
			}
		}
	}
	if len(checkedList) == len(exchanges) {
		go func() {
			self.getable <- true
		}()
	}
}

func (self *ConcurrentAllPriceData) GetCurrentData() map[common.TokenPairID]common.OnePrice {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return self.data
}

func (self *ConcurrentAllPriceData) GetData() map[common.TokenPairID]common.OnePrice {
	self.mu.RLock()
	defer self.mu.RUnlock()
	<-self.getable
	return self.data
}
