package metric

import (
	"errors"
	"log"
	"strconv"
	"sync"

	"github.com/KyberNetwork/reserve-data/common"
)

const MAX_CAPACITY int = 1000

type RamMetricStorage struct {
	mu               sync.RWMutex
	data             []*MetricEntry
	pendingTargetQty TokenTargetQty
	tokenTargetQty   TokenTargetQty
}

func NewRamMetricStorage() *RamMetricStorage {
	return &RamMetricStorage{
		mu:               sync.RWMutex{},
		data:             []*MetricEntry{},
		pendingTargetQty: TokenTargetQty{},
		tokenTargetQty:   TokenTargetQty{},
	}
}

func (self *RamMetricStorage) StoreMetric(data *MetricEntry, timepoint uint64) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.data = append(self.data, data)
	first := len(self.data) - MAX_CAPACITY
	if first > 0 {
		for i := 0; i < first; i++ {
			self.data[i] = nil
		}
		self.data = self.data[first:]
	}
	return nil
}

func (self *RamMetricStorage) GetMetric(tokens []common.Token, fromTime, toTime uint64) (map[string]MetricList, error) {
	self.mu.RLock()
	defer self.mu.RUnlock()
	imResult := map[string]*MetricList{}
	for _, tok := range tokens {
		imResult[tok.ID] = &MetricList{}
	}
	for i := len(self.data) - 1; i >= 0; i-- {
		data := self.data[i]
		if fromTime <= data.Timestamp && data.Timestamp <= toTime {
			log.Printf("iterate over %d", data.Timestamp)
			for tok, metric := range data.Data {
				metricList, found := imResult[tok]
				if found {
					*metricList = append(*metricList, TokenMetricResponse{
						Timestamp: data.Timestamp,
						AfpMid:    metric.AfpMid,
						Spread:    metric.Spread,
					})
					log.Printf("token: %s, metricList: %+v", tok, metricList)
				}
			}
			log.Printf("result: %+v", imResult)
		} else if data.Timestamp <= fromTime {
			break
		}
	}
	result := map[string]MetricList{}
	for k, v := range imResult {
		result[k] = *v
	}
	return result, nil
}

func (self *RamMetricStorage) StorePendingTargetQty(data, dataType string) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	if self.pendingTargetQty.ID != 0 {
		return errors.New("There is one pending target quantity, please confirm or cancel it before adding a new one")
	}
	self.pendingTargetQty.Type, _ = strconv.ParseInt(dataType, 10, 64)
	self.pendingTargetQty.Data = data
	self.pendingTargetQty.Status = "unconfirmed"
	self.pendingTargetQty.ID = common.GetTimepoint()
	return nil
}

func (self *RamMetricStorage) RemovePendingTargetQty() error {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.pendingTargetQty = TokenTargetQty{}
	return nil
}

func (self *RamMetricStorage) GetPendingTargetQty() (TokenTargetQty, error) {
	self.mu.RLock()
	defer self.mu.RUnlock()
	result := self.pendingTargetQty
	return result, nil
}

func (self *RamMetricStorage) StoreTokenTargetQty(id, data string) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	if self.pendingTargetQty.ID == 0 {
		return errors.New("There is not pending data. Please set before confirm")
	}
	self.tokenTargetQty = self.pendingTargetQty
	self.tokenTargetQty.Status = "confirmed"
	self.RemovePendingTargetQty()
	return nil
}

func (self *RamMetricStorage) GetTokenTargetQty() (TokenTargetQty, error) {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return self.tokenTargetQty, nil
}

func (self *RamMetricStorage) GetRebalanceControl() (RebalanceControl, error) {
	// TODO: update this
	result := RebalanceControl{}
	return result, nil
}

func (self *RamMetricStorage) StoreRebalanceControl(status bool) error {
	// TODO: update this
	return nil
}

func (self *RamMetricStorage) GetSetrateControl() (SetrateControl, error) {
	// TODO: update this
	result := SetrateControl{}
	return result, nil
}

func (self *RamMetricStorage) StoreSetrateControl(status bool) error {
	// TODO: update this
	return nil
}

func (self *RamMetricStorage) StorePendingPWIEquation(data string) error {
	return nil
}

func (self *RamMetricStorage) GetPendingPWIEquation() (PWIEquation, error) {
	return PWIEquation{}, nil
}

func (self *RamMetricStorage) StorePWIEquation(data string) error {
	return nil
}

func (self *RamMetricStorage) GetPWIEquation() (PWIEquation, error) {
	return PWIEquation{}, nil
}

func (self *RamMetricStorage) RemovePendingPWIEquation() error {
	return nil
}
