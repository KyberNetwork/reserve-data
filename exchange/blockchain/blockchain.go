package blockchain

import (
	"math/big"

	"github.com/KyberNetwork/reserve-data/common/blockchain"
	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type Blockchain struct {
	*blockchain.BaseBlockchain
}

func (b *Blockchain) GetIntermediatorAddr(opName string) ethereum.Address {
	return b.OperatorAddresses()[opName]
}

func (b *Blockchain) SendTokenFromAccountToExchange(amount *big.Int, exchangeAddress ethereum.Address, tokenAddress ethereum.Address, gasPrice *big.Int, opName string) (*types.Transaction, error) {
	opts, err := b.GetTxOpts(opName, nil, gasPrice, nil)
	if err != nil {
		return nil, err
	}
	tx, err := b.BuildSendERC20Tx(opts, amount, exchangeAddress, tokenAddress)
	if err != nil {
		return nil, err
	}
	return b.SignAndBroadcast(tx, opName)
}

func (b *Blockchain) SendETHFromAccountToExchange(amount *big.Int, exchangeAddress ethereum.Address, gasPrice *big.Int, opName string) (*types.Transaction, error) {
	opts, err := b.GetTxOpts(opName, nil, gasPrice, amount)
	if err != nil {
		return nil, err
	}
	tx, err := b.BuildSendETHTx(opts, exchangeAddress)
	if err != nil {
		return nil, err
	}
	return b.SignAndBroadcast(tx, opName)
}

func NewBlockchain(
	base *blockchain.BaseBlockchain,
	signer blockchain.Signer,
	nonce blockchain.NonceCorpus,
	opName string) (*Blockchain, error) {

	base.MustRegisterOperator(opName, blockchain.NewOperator(signer, nonce))

	return &Blockchain{
		BaseBlockchain: base,
	}, nil
}
