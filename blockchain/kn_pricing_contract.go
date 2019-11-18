package blockchain

import (
	"context"
	"math/big"
	"time"

	"github.com/KyberNetwork/reserve-data/common/blockchain"
	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func (bc *Blockchain) GeneratedSetBaseRate(opts blockchain.TxOpts, tokens []ethereum.Address, baseBuy []*big.Int, baseSell []*big.Int, buy [][14]byte, sell [][14]byte, blockNumber *big.Int, indices []*big.Int) (*types.Transaction, error) {
	timeout, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return bc.BuildTx(timeout, opts, bc.pricing, "setBaseRate", tokens, baseBuy, baseSell, buy, sell, blockNumber, indices)
}

func (bc *Blockchain) GeneratedSetCompactData(opts blockchain.TxOpts, buy [][14]byte, sell [][14]byte, blockNumber *big.Int, indices []*big.Int) (*types.Transaction, error) {
	timeout, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return bc.BuildTx(timeout, opts, bc.pricing, "setCompactData", buy, sell, blockNumber, indices)
}

func (bc *Blockchain) GeneratedSetImbalanceStepFunction(opts blockchain.TxOpts, token ethereum.Address, xBuy []*big.Int, yBuy []*big.Int, xSell []*big.Int, ySell []*big.Int) (*types.Transaction, error) {
	timeout, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return bc.BuildTx(timeout, opts, bc.pricing, "setImbalanceStepFunction", token, xBuy, yBuy, xSell, ySell)
}

func (bc *Blockchain) GeneratedSetQtyStepFunction(opts blockchain.TxOpts, token ethereum.Address, xBuy []*big.Int, yBuy []*big.Int, xSell []*big.Int, ySell []*big.Int) (*types.Transaction, error) {
	timeout, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return bc.BuildTx(timeout, opts, bc.pricing, "setQtyStepFunction", token, xBuy, yBuy, xSell, ySell)
}

func (bc *Blockchain) GeneratedGetRate(opts blockchain.CallOpts, token ethereum.Address, currentBlockNumber *big.Int, buy bool, qty *big.Int) (*big.Int, error) {
	timeOut := 2 * time.Second
	out := big.NewInt(0)
	err := bc.Call(timeOut, opts, bc.pricing, out, "getRate", token, currentBlockNumber, buy, qty)
	return out, err
}

//GeneratedGetStepFunctionData get step function data for an token
func (bc *Blockchain) GeneratedGetStepFunctionData(opts blockchain.CallOpts, token ethereum.Address, command *big.Int, param *big.Int) (*big.Int, error) {
	timeOut := 2 * time.Second
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := bc.Call(timeOut, opts, bc.pricing, out, "getStepFunctionData", token, command, param)
	return *ret0, err
}
