package core

import (
	"math/big"

	"github.com/KyberNetwork/reserve-data/common"
	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type Blockchain interface {
	StandardGasPrice() float64
	SafeLowGasPrice() float64
	FastGasPrice() float64
	Send(
		token common.Token,
		amount *big.Int,
		address ethereum.Address) (*types.Transaction, error)
	SetRates(
		tokens []ethereum.Address,
		buys []*big.Int,
		sells []*big.Int,
		block *big.Int,
		nonce *big.Int,
		gasPrice *big.Int) (*types.Transaction, error)
	SetRateMinedNonce() (uint64, error)
	GetAddresses() *common.Addresses
}
