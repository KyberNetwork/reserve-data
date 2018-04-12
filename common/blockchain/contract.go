package blockchain

import (
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ethereum "github.com/ethereum/go-ethereum/common"
)

type Contract struct {
	Address ethereum.Address
	ABI     abi.ABI
}

func NewContract(address ethereum.Address, abipath string) *Contract {
	file, err := os.Open(abipath)
	if err != nil {
		panic(err)
	}
	parsed, err := abi.JSON(file)
	if err != nil {
		panic(err)
	}
	return &Contract{address, parsed}
}
