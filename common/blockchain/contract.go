package blockchain

import (
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ethereum "github.com/ethereum/go-ethereum/common"
)

type Contract struct {
	Address ethereum.Address
	ABI     abi.ABI
}

func NewContract(address ethereum.Address, rawABI string) *Contract {
	parsed, err := abi.JSON(strings.NewReader(rawABI))
	if err != nil {
		panic(err)
	}
	return &Contract{address, parsed}
}
