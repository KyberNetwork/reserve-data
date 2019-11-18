package blockchain

import (
	"math/big"
	"time"

	"github.com/KyberNetwork/reserve-data/common/blockchain"
	ethereum "github.com/ethereum/go-ethereum/common"
)

func (bc *Blockchain) GeneratedGetBalances(opts blockchain.CallOpts, reserve ethereum.Address, tokens []ethereum.Address) ([]*big.Int, error) {
	out := new([]*big.Int)
	timeOut := 2 * time.Second
	err := bc.Call(timeOut, opts, bc.wrapper, out, "getBalances", reserve, tokens)
	return *out, err
}

func (bc *Blockchain) GeneratedGetTokenIndicies(opts blockchain.CallOpts, ratesContract ethereum.Address, tokenList []ethereum.Address) ([]*big.Int, []*big.Int, error) {
	var (
		ret0 = new([]*big.Int)
		ret1 = new([]*big.Int)
	)
	out := &[]interface{}{
		ret0,
		ret1,
	}
	timeOut := 2 * time.Second
	err := bc.Call(timeOut, opts, bc.wrapper, out, "getTokenIndicies", ratesContract, tokenList)
	return *ret0, *ret1, err
}

func (bc *Blockchain) GeneratedGetTokenRates(
	opts blockchain.CallOpts,
	ratesContract ethereum.Address,
	tokenList []ethereum.Address) ([]*big.Int, []*big.Int, []int8, []int8, []*big.Int, error) {
	var (
		ret0 = new([]*big.Int)
		ret1 = new([]*big.Int)
		ret2 = new([]int8)
		ret3 = new([]int8)
		ret4 = new([]*big.Int)
	)
	out := &[]interface{}{
		ret0,
		ret1,
		ret2,
		ret3,
		ret4,
	}
	timeOut := 2 * time.Second
	err := bc.Call(timeOut, opts, bc.wrapper, out, "getTokenRates", ratesContract, tokenList)
	return *ret0, *ret1, *ret2, *ret3, *ret4, err
}
