package nonce

import (
	"context"
	"math/big"
	"sync"
	"time"

	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

//AutoIncreasing is used for manually increase transaction nonce
type AutoIncreasing struct {
	address     ethereum.Address
	mu          sync.Mutex
	manualNonce *big.Int
}

//NewAutoIncreasing return a new AutoIncreasing instance
func NewAutoIncreasing(
	address ethereum.Address) *AutoIncreasing {
	return &AutoIncreasing{
		address,
		sync.Mutex{},
		big.NewInt(0),
	}
}

//GetAddress return a AutoIncreasing addresss
func (ai *AutoIncreasing) GetAddress() ethereum.Address {
	return ai.GetAddress()
}

func (ai *AutoIncreasing) getNonceFromNode(ethclient *ethclient.Client) (*big.Int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	nonce, err := ethclient.PendingNonceAt(ctx, ai.GetAddress())
	return big.NewInt(int64(nonce)), err
}

//MinedNonce return nonce which is mined
func (ai *AutoIncreasing) MinedNonce(ethclient *ethclient.Client) (*big.Int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	nonce, err := ethclient.NonceAt(ctx, ai.GetAddress(), nil)
	return big.NewInt(int64(nonce)), err
}

//GetNextNonce return next transaction nonce
func (ai *AutoIncreasing) GetNextNonce(ethclient *ethclient.Client) (*big.Int, error) {
	ai.mu.Lock()
	defer ai.mu.Unlock()
	nodeNonce, err := ai.getNonceFromNode(ethclient)
	if err != nil {
		return nodeNonce, err
	} else {
		if nodeNonce.Cmp(ai.manualNonce) == 1 {
			ai.manualNonce = big.NewInt(0).Add(nodeNonce, ethereum.Big1)
			return nodeNonce, nil
		} else {
			nextNonce := ai.manualNonce
			ai.manualNonce = big.NewInt(0).Add(nextNonce, ethereum.Big1)
			return nextNonce, nil
		}
	}
}
