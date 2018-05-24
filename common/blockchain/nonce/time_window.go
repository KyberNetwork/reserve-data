package nonce

import (
	"context"
	"math/big"
	"sync"
	"time"

	"github.com/KyberNetwork/reserve-data/common"
	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type TimeWindow struct {
	address     ethereum.Address
	mu          sync.Mutex
	manualNonce *big.Int
	time        uint64 // last time nonce was requested
	window      uint64 // window time in millisecond
}

// be very very careful to set `window` param, if we set it to high value, it can lead to nonce lost making the whole pricing operation stuck
func NewTimeWindow(address ethereum.Address, window uint64) *TimeWindow {
	return &TimeWindow{
		address,
		sync.Mutex{},
		big.NewInt(0),
		0,
		window,
	}
}

func (self *TimeWindow) GetAddress() ethereum.Address {
	return self.address
}

func (self *TimeWindow) getNonceFromNode(ethclient *ethclient.Client) (*big.Int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	nonce, err := ethclient.PendingNonceAt(ctx, self.GetAddress())
	return big.NewInt(int64(nonce)), err
}

func (self *TimeWindow) MinedNonce(ethclient *ethclient.Client) (*big.Int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	nonce, err := ethclient.NonceAt(ctx, self.GetAddress(), nil)
	return big.NewInt(int64(nonce)), err
}

func (self *TimeWindow) GetNextNonce(ethclient *ethclient.Client) (*big.Int, error) {
	self.mu.Lock()
	defer self.mu.Unlock()
	t := common.GetTimepoint()
	if t-self.time < self.window {
		self.time = t
		self.manualNonce.Add(self.manualNonce, ethereum.Big1)
		return self.manualNonce, nil
	} else {
		nonce, err := self.getNonceFromNode(ethclient)
		if err != nil {
			return big.NewInt(0), err
		}
		self.time = t
		self.manualNonce = nonce
		return self.manualNonce, nil
	}
}
