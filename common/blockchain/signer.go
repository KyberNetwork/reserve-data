package blockchain

import (
	"io"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type Signer interface {
	GetAddress() ethereum.Address
	Sign(*types.Transaction) (*types.Transaction, error)
}

type EthereumSigner struct {
	opts *bind.TransactOpts
}

func (self EthereumSigner) GetAddress() ethereum.Address {
	return self.opts.From
}

func (self EthereumSigner) Sign(tx *types.Transaction) (*types.Transaction, error) {
	return self.opts.Signer(types.HomesteadSigner{}, self.GetAddress(), tx)
}

func NewEthereumSigner(keyPath string, passphrase string) *EthereumSigner {
	var (
		err error
		key io.Reader
	)

	if key, err = os.Open(keyPath); err != nil {
		// try to auto generate the key for development
		if key, err = GenerateDevelopmentKeystoreIfNotExists(err, keyPath, passphrase); err != nil {
			panic(err)
		}
	}

	auth, err := bind.NewTransactor(key, passphrase)
	if err != nil {
		panic(err)
	}
	return &EthereumSigner{opts: auth}
}
