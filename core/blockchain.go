package core

import (
	"math/big"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/blockchain"

	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// Blockchain is the interface wraps around all core methods to interact
// with Ethereum blockchain.
type Blockchain interface {
	Send(token common.Token, amount *big.Int, address ethereum.Address, nonce *big.Int, gasPrice *big.Int) (*types.Transaction, error)
	SetRates(
		tokens []ethereum.Address,
		buys []*big.Int,
		sells []*big.Int,
		block *big.Int,
		nonce *big.Int,
		gasPrice *big.Int) (*types.Transaction, error)
	blockchain.MinedNoncePicker

	BuildSendETHTx(opts blockchain.TxOpts, to ethereum.Address) (*types.Transaction, error)
	GetDepositOPAddress() ethereum.Address
	SignAndBroadcast(tx *types.Transaction, from string) (*types.Transaction, error)
}
