package ethutil

import (
	ethereum "github.com/ethereum/go-ethereum/common"
)

// IsEthereumAddress returns true if the given address is ethereum.
func IsEthereumAddress(addr ethereum.Address) bool {
	return addr == ethereum.HexToAddress("0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee")
}

// IsZeroAddress check if address is zero
func IsZeroAddress(a ethereum.Address) bool {
	return a.Hash().Big().BitLen() == 0
}
