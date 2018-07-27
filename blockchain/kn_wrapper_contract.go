package blockchain

import (
	"log"
	"math/big"
	"time"

	"github.com/KyberNetwork/reserve-data/common/blockchain"
	ethereum "github.com/ethereum/go-ethereum/common"
)

func (b *Blockchain) GeneratedGetBalances(opts blockchain.CallOpts, reserve ethereum.Address, tokens []ethereum.Address) ([]*big.Int, error) {
	out := new([]*big.Int)
	timeOut := 2 * time.Second
	err := b.Call(timeOut, opts, b.wrapper, out, "getBalances", reserve, tokens)
	return *out, err
}

func (b *Blockchain) GeneratedGetTokenIndicies(opts blockchain.CallOpts, ratesContract ethereum.Address, tokenList []ethereum.Address) ([]*big.Int, []*big.Int, error) {
	var (
		ret0 = new([]*big.Int)
		ret1 = new([]*big.Int)
	)
	out := &[]interface{}{
		ret0,
		ret1,
	}
	timeOut := 2 * time.Second
	err := b.Call(timeOut, opts, b.wrapper, out, "getTokenIndicies", ratesContract, tokenList)
	return *ret0, *ret1, err
}

func (b *Blockchain) GeneratedGetTokenRates(
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
	err := b.Call(timeOut, opts, b.wrapper, out, "getTokenRates", ratesContract, tokenList)
	return *ret0, *ret1, *ret2, *ret3, *ret4, err
}

func (b *Blockchain) GeneratedGetReserveRates(
	opts blockchain.CallOpts,
	reserveAddress ethereum.Address,
	srcAddresses []ethereum.Address,
	destAddresses []ethereum.Address) ([]*big.Int, []*big.Int, error) {
	var (
		ret0 = new([]*big.Int)
		ret1 = new([]*big.Int)
	)
	out := &[]interface{}{
		ret0,
		ret1,
	}
	timeOut := 2 * time.Second
	err := b.Call(timeOut, opts, b.wrapper, out, "getReserveRate", reserveAddress, srcAddresses, destAddresses)
	if err != nil {
		log.Println("cannot get reserve rates: ", err.Error())
	}
	return *ret0, *ret1, err
}
