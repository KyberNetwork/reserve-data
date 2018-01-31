package nonce

import (
	"context"
	"math/big"
	"sync"

	"github.com/KyberNetwork/reserve-data/blockchain"
	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type AutoIncreasing struct {
	ethclient   *ethclient.Client
	signer      blockchain.Signer
	mu          sync.Mutex
	manualNonce *big.Int
}

func NewAutoIncreasing(
	ethclient *ethclient.Client,
	signer blockchain.Signer) *AutoIncreasing {
	return &AutoIncreasing{
		ethclient,
		signer,
		sync.Mutex{},
		big.NewInt(0),
	}
}

func (self *AutoIncreasing) GetAddress(transaction string) ethereum.Address {
	return self.signer.GetAddress(transaction)
}

func (self *AutoIncreasing) getNonceFromNode(transaction string) (*big.Int, error) {
	option := context.Background()
	nonce, err := self.ethclient.PendingNonceAt(option, self.signer.GetAddress(transaction))
	return big.NewInt(int64(nonce)), err
}

func (self *AutoIncreasing) GetNextNonce(transaction string) (*big.Int, error) {
	self.mu.Lock()
	defer self.mu.Unlock()
	nodeNonce, err := self.getNonceFromNode(transaction)
	if err != nil {
		return nodeNonce, err
	} else {
		if nodeNonce.Cmp(self.manualNonce) == 1 {
			self.manualNonce = big.NewInt(0).Add(nodeNonce, ethereum.Big1)
			return nodeNonce, nil
		} else {
			nextNonce := self.manualNonce
			self.manualNonce = big.NewInt(0).Add(nextNonce, ethereum.Big1)
			return nextNonce, nil
		}
	}
}
