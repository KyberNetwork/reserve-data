package blockchain

import (
	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
)

type NonceCorpus interface {
	GetAddress() ethereum.Address
	GetNextNonce(ethclient *ethclient.Client) (*big.Int, error)
	MinedNonce(ethclient *ethclient.Client) (*big.Int, error)
}
