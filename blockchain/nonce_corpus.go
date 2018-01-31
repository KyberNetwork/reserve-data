package blockchain

import (
	ethereum "github.com/ethereum/go-ethereum/common"
	"math/big"
)

type NonceCorpus interface {
	GetAddress(string) ethereum.Address
	GetNextNonce(string) (*big.Int, error)
}
