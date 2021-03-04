package blockchain

import (
	"bytes"
	"io/ioutil"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// Signer contains method to sign a Ethereum transaction.
type Signer interface {
	GetAddress() ethereum.Address
	Sign(*types.Transaction) (*types.Transaction, error)
}

type EthereumSigner struct {
	opts    *bind.TransactOpts
	chainID *big.Int
}

func (es EthereumSigner) GetAddress() ethereum.Address {
	return es.opts.From
}

func (es EthereumSigner) Sign(tx *types.Transaction) (*types.Transaction, error) {
	return es.opts.Signer(es.GetAddress(), tx)
}

func NewEthereumSigner(keyPath string, passphrase string, chainID *big.Int) *EthereumSigner {
	data, err := ioutil.ReadFile(keyPath)
	if err != nil {
		panic(err)
	}
	auth, err := bind.NewTransactorWithChainID(bytes.NewReader(data), passphrase, chainID)
	if err != nil {
		panic(err)
	}
	return &EthereumSigner{opts: auth, chainID: chainID}
}
