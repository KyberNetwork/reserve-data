package blockchain

import (
	"context"
	"math/big"
	"time"

	"github.com/KyberNetwork/reserve-data/common/blockchain"
	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// GeneratedSetBaseRate build tx for set base rate function
func (b *Blockchain) GeneratedSetBaseRate(opts blockchain.TxOpts, tokens []ethereum.Address, baseBuy []*big.Int, baseSell []*big.Int, buy [][14]byte, sell [][14]byte, blockNumber *big.Int, indices []*big.Int) (*types.Transaction, error) {
	timeout, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return b.BuildTx(timeout, opts, b.pricing, "setBaseRate", tokens, baseBuy, baseSell, buy, sell, blockNumber, indices)
}

// GeneratedSetCompactData build tx for set compact data function
func (b *Blockchain) GeneratedSetCompactData(opts blockchain.TxOpts, buy [][14]byte, sell [][14]byte, blockNumber *big.Int, indices []*big.Int) (*types.Transaction, error) {
	timeout, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return b.BuildTx(timeout, opts, b.pricing, "setCompactData", buy, sell, blockNumber, indices)
}

// GeneratedGetRate call get rate function from pricing contract
func (b *Blockchain) GeneratedGetRate(opts blockchain.CallOpts, token ethereum.Address, currentBlockNumber *big.Int, buy bool, qty *big.Int) (*big.Int, error) {
	timeOut := 2 * time.Second
	out := big.NewInt(0)
	err := b.Call(timeOut, opts, b.pricing, out, "getRate", token, currentBlockNumber, buy, qty)
	return out, err
}

// GeneratedGetListedTokens call getListedToken function in pricing contract
func (b *Blockchain) GeneratedGetListedTokens(opts blockchain.CallOpts) ([]ethereum.Address, error) {
	timeOut := 2 * time.Second
	var (
		ret0 = new([]ethereum.Address)
	)
	out := ret0
	err := b.Call(timeOut, opts, b.pricing, out, "getListedTokens")
	return *ret0, err
}
