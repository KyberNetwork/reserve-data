package storage

import (
	"fmt"
	"sync"

	"github.com/KyberNetwork/reserve-data/feed-provider/common"
)

type RAMStorage struct {
	m    *sync.RWMutex
	data map[string]common.Feed
}

func NewRAMStorage() *RAMStorage {
	return &RAMStorage{
		m:    &sync.RWMutex{},
		data: make(map[string]common.Feed),
	}
}

func (r *RAMStorage) Save(feed string, data common.Feed) error {
	r.m.Lock()
	defer r.m.Unlock()
	r.data[feed] = data
	return nil
}

func (r *RAMStorage) Load(feed string) common.Feed {
	r.m.RLock()
	defer r.m.RUnlock()
	data, ok := r.data[feed]
	if !ok {
		return common.Feed{
			Error: fmt.Errorf("there is no record for %s", feed),
		}
	}
	return data
}
