package blockchain

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Broadcaster takes a signed tx and try to broadcast it to all
// nodes that it manages as fast as possible. It returns a map of
// failures and a bool indicating that the tx is broadcasted to
// at least 1 node
type Broadcaster struct {
	clients map[string]*ethclient.Client
}

func (bc Broadcaster) broadcast(
	ctx context.Context,
	id string, client *ethclient.Client, tx *types.Transaction,
	wg *sync.WaitGroup, failures *sync.Map) {
	defer wg.Done()
	err := client.SendTransaction(ctx, tx)
	if err != nil {
		failures.Store(id, err)
	}
}

//Broadcast broadcast a transaction and return list of error if have
func (bc Broadcaster) Broadcast(tx *types.Transaction) (map[string]error, bool) {
	failures := sync.Map{}
	wg := sync.WaitGroup{}
	for id, client := range bc.clients {
		wg.Add(1)
		timeout, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		bc.broadcast(timeout, id, client, tx, &wg, &failures)
		defer cancel()
	}
	wg.Wait()
	result := map[string]error{}
	failures.Range(func(key, value interface{}) bool {
		k, ok := key.(string)
		if !ok {
			log.Printf("Broadcast: key (%v) cannot be asserted to string", key)
			return true
		}
		err, ok := value.(error)
		if !ok {
			log.Printf("Broadcast: value (%v) cannot be asserted to error", value)
			return true
		}
		result[k] = err
		return true
	})
	return result, len(result) != len(bc.clients) && len(bc.clients) > 0
}

//NewBroadcaster return a new Broadcaster instance
func NewBroadcaster(clients map[string]*ethclient.Client) *Broadcaster {
	return &Broadcaster{
		clients: clients,
	}
}
