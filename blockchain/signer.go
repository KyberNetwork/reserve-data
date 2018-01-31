package blockchain

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type Signer interface {
	GetAddress(string) ethereum.Address
	Sign(ethereum.Address, *types.Transaction, string) (*types.Transaction, error)
	GetTransactOpts(string) *bind.TransactOpts
}
