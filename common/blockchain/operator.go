package blockchain

import (
	ethereum "github.com/ethereum/go-ethereum/common"
)

type Operator struct {
	Address     ethereum.Address
	NonceCorpus NonceCorpus
	Signer      Signer
}

func NewOperator(signer Signer, nonce NonceCorpus) *Operator {
	return &Operator{
		Address:     nonce.GetAddress(),
		NonceCorpus: nonce,
		Signer:      signer,
	}
}
