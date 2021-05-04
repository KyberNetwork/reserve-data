package core

import (
	"math/big"

	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/KyberNetwork/reserve-data/common/blockchain"
	"github.com/KyberNetwork/reserve-data/reservesetting/common"
)

// Blockchain is the interface wraps around all core methods to interact
// with Ethereum blockchain.
type Blockchain interface {
	Send(
		asset common.Asset,
		amount *big.Int,
		address ethereum.Address,
		nonce *big.Int,
		gasPrice *big.Int) (*types.Transaction, error)
	TransferToSelf(op string, gasPrice *big.Int, nonce *big.Int) (*types.Transaction, error)
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
	GetPricingOPAddress() ethereum.Address
	SignAndBroadcast(tx *types.Transaction, from string) (*types.Transaction, error)
	SpeedupTx(tx ethereum.Hash, recommendGasPrice float64, maxGasPrice float64, opAccount string) (*types.Transaction, error)
	CurrentBlock() (uint64, error)
}
