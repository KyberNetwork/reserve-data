package blockchain

import (
	"math/big"

	"github.com/KyberNetwork/reserve-data/common/blockchain"
	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const HuobiOP string = "huobi_op"

type Blockchain struct {
	*blockchain.BaseBlockchain
}

func (b *Blockchain) GetIntermediatorAddr() ethereum.Address {
	return b.OperatorAddresses()[HuobiOP]
}

func (b *Blockchain) SendTokenFromAccountToExchange(amount *big.Int, exchangeAddress ethereum.Address, tokenAddress ethereum.Address) (*types.Transaction, error) {
	opts, err := b.GetTxOpts(HuobiOP, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	tx, err := b.BuildSendERC20Tx(opts, amount, exchangeAddress, tokenAddress)
	if err != nil {
		return nil, err
	}
	return b.SignAndBroadcast(tx, HuobiOP)
}

func (b *Blockchain) SendETHFromAccountToExchange(amount *big.Int, exchangeAddress ethereum.Address) (*types.Transaction, error) {
	opts, err := b.GetTxOpts(HuobiOP, nil, nil, amount)
	if err != nil {
		return nil, err
	}
	tx, err := b.BuildSendETHTx(opts, exchangeAddress)
	if err != nil {
		return nil, err
	}
	return b.SignAndBroadcast(tx, HuobiOP)
}

func NewBlockchain(
	base *blockchain.BaseBlockchain,
	signer blockchain.Signer, nonce blockchain.NonceCorpus) (*Blockchain, error) {

	base.MustRegisterOperator(HuobiOP, blockchain.NewOperator(signer, nonce))

	return &Blockchain{
		BaseBlockchain: base,
	}, nil
}
