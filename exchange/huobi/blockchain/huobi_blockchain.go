package blockchain

import (
	"math/big"

	"github.com/KyberNetwork/reserve-data/common/blockchain"
	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const HUOBI_OP string = "huobi_op"

type Blockchain struct {
	*blockchain.BaseBlockchain
}

func (self *Blockchain) SendTokenFromAccountToExchange(amount *big.Int, exchangeAddress ethereum.Address, tokenAddress ethereum.Address) (*types.Transaction, error) {
	opts, err := self.GetTxOpts(HUOBI_OP, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	tx, err := self.BuildSendERC20Tx(opts, amount, exchangeAddress, tokenAddress)
	if err != nil {
		return nil, err
	} else {
		return self.SignAndBroadcast(tx, HUOBI_OP)
	}
}

func (self *Blockchain) SendETHFromAccountToExchange(amount *big.Int, exchangeAddress ethereum.Address) (*types.Transaction, error) {
	opts, err := self.GetTxOpts(HUOBI_OP, nil, nil, amount)
	if err != nil {
		return nil, err
	}
	tx, err := self.BuildSendETHTx(opts, exchangeAddress)
	if err != nil {
		return nil, err
	} else {
		return self.SignAndBroadcast(tx, HUOBI_OP)
	}
}

func NewBlockchain(
	base *blockchain.BaseBlockchain,
	signer blockchain.Signer, nonce blockchain.NonceCorpus) (*Blockchain, error) {

	base.RegisterOperator(HUOBI_OP, blockchain.NewOperator(signer, nonce))

	return &Blockchain{
		BaseBlockchain: base,
	}, nil
}

// func (self *Blockchain) CheckBalance(token common.Token) *big.Int {
// 	addr := self.intermediateSigner.GetAddress()
// 	balance, err := self.FetchBalanceData(addr, token)
// 	if err != nil || !balance.Valid {
// 		return big.NewInt(0)
// 	}

// 	balanceFloat := balance.Balance.ToFloat(token.Decimal)
// 	return (getBigIntFromFloat(balanceFloat, token.Decimal))

// }

// func (self *Blockchain) FetchBalanceData(reserve ethereum.Address, token common.Token) (common.BalanceEntry, error) {
// 	result := common.BalanceEntry{}
// 	tokens := []ethereum.Address{}
// 	tokens = append(tokens, ethereum.HexToAddress(token.Address))

// 	timestamp := common.GetTimestamp()
// 	balances, err := self.wrapper.GetBalances(nil, nil, reserve, tokens)
// 	returnTime := common.GetTimestamp()
// 	log.Printf("Fetcher ------> balances: %v, err: %s", balances, err)
// 	if err != nil {
// 		result = common.BalanceEntry{
// 			Valid:      false,
// 			Error:      err.Error(),
// 			Timestamp:  timestamp,
// 			ReturnTime: returnTime,
// 		}
// 	} else {
// 		if balances[0].Cmp(big.NewInt(0)) == 0 || balances[0].Cmp(big.NewInt(10).Exp(big.NewInt(10), big.NewInt(33), nil)) > 0 {
// 			log.Printf("Fetcher ------> balances of token %s is invalid", token.ID)
// 			result = common.BalanceEntry{
// 				Valid:      false,
// 				Error:      "Got strange balances from node. It equals to 0 or is bigger than 10^33",
// 				Timestamp:  timestamp,
// 				ReturnTime: returnTime,
// 				Balance:    common.RawBalance(*balances[0]),
// 			}
// 		} else {
// 			result = common.BalanceEntry{
// 				Valid:      true,
// 				Timestamp:  timestamp,
// 				ReturnTime: returnTime,
// 				Balance:    common.RawBalance(*balances[0]),
// 			}
// 		}
// 	}

// 	return result, nil
// }
