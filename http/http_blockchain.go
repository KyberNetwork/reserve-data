package http

import (
	ethereum "github.com/ethereum/go-ethereum/common"
)

// Blockchain is used in http server as the caller to blockchain for information.
// Currently it is used for smart contract token's indice query.
type Blockchain interface {
	LoadAndSetTokenIndices() error
	CheckTokenIndices(ethereum.Address) error
	GetPricingOPAddress() ethereum.Address
	GetDepositOPAddress() ethereum.Address
	GetIntermediatorOPAddress(opName string) ethereum.Address
	GetWrapperAddress() ethereum.Address
	GetProxyAddress() ethereum.Address
	GetReserveAddress() ethereum.Address
	GetRateQueryHelperAddress() ethereum.Address
	ListedTokens() []ethereum.Address
}
