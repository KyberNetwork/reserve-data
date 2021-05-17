package exchange

import (
	"math/big"

	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// HuobiBlockchain contains methods to interact with blockchain from Huobi address.
type HuobiBlockchain interface {
	SendETHFromAccountToExchange(amount *big.Int, exchangeAddress ethereum.Address, gasPrice *big.Int, opName string, nonce *big.Int) (*types.Transaction, error)
	SendTokenFromAccountToExchange(amount *big.Int, exchangeAddress ethereum.Address, tokenAddress ethereum.Address, gasPrice *big.Int, opName string) (*types.Transaction, error)
	TxStatus(hash ethereum.Hash) (string, uint64, error)
	GetIntermediatorAddr(opName string) ethereum.Address
	RecommendedGasPriceFromNode() (*big.Int, error)
}
