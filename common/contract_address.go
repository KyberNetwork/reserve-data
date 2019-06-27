package common

import (
	ethereum "github.com/ethereum/go-ethereum/common"
)

// ContractAddressConfiguration contains the smart contract addresses.
type ContractAddressConfiguration struct {
	Reserve ethereum.Address
	Wrapper ethereum.Address
	Pricing ethereum.Address
}
