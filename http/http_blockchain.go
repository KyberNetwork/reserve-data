package http

import (
	ethereum "github.com/ethereum/go-ethereum/common"
)

// Blockchain is used in http server as the caller to blockchain for information.
// Currently it is used for smart contract token's indice query.
type Blockchain interface {
	LoadAndSetTokenIndices([]ethereum.Address) error
	CheckTokenIndices(ethereum.Address) error
	GetPricingOPAddress() ethereum.Address
	GetDepositOPAddress() ethereum.Address
	GetIntermediatorOPAddress() ethereum.Address
}
