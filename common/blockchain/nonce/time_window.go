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

//TimeWindow is used for get transaction nonce base on time window
type TimeWindow struct {
	address     ethereum.Address
	mu          sync.Mutex
	manualNonce *big.Int
	time        uint64 // last time nonce was requested
	window      uint64 // window time in millisecond
}

// NewTimeWindow return a new TimeWindow instance
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

//GetAddress return TimeWindow address
func (tw *TimeWindow) GetAddress() ethereum.Address {
	return tw.address
}

func (tw *TimeWindow) getNonceFromNode(ethclient *ethclient.Client) (*big.Int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()
	nonce, err := ethclient.PendingNonceAt(ctx, tw.GetAddress())
	return big.NewInt(int64(nonce)), err
}

//MinedNonce return a TimeWindow mined nonce
func (tw *TimeWindow) MinedNonce(ethclient *ethclient.Client) (*big.Int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()
	nonce, err := ethclient.NonceAt(ctx, tw.GetAddress(), nil)
	return big.NewInt(int64(nonce)), err
}

//GetNextNonce return a time window next nonce
func (tw *TimeWindow) GetNextNonce(ethclient *ethclient.Client) (*big.Int, error) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	t := common.GetTimepoint()
	if t-tw.time < tw.window {
		tw.time = t
		tw.manualNonce.Add(tw.manualNonce, ethereum.Big1)
		return tw.manualNonce, nil
	}
	nonce, err := tw.getNonceFromNode(ethclient)
	if err != nil {
		return big.NewInt(0), err
	}
	tw.time = t
	tw.manualNonce = nonce
	return tw.manualNonce, nil
}
