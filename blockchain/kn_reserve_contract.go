package blockchain

import (
	"context"
	"fmt"
	"math/big"
	"time"

	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/KyberNetwork/reserve-data/common/blockchain"
)

func (bc *Blockchain) GeneratedWithdraw(opts blockchain.TxOpts, token ethereum.Address, amount *big.Int, destination ethereum.Address) (*types.Transaction, error) {
	timeout, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	tx, err := bc.BuildTx(timeout, opts, bc.reserve, "withdraw", token, amount, destination)
	if err == blockchain.ErrEstimateGasFailed {
		balance, bErr := bc.BalanceOf(token, bc.reserve.Address)
		if bErr == nil {
			bc.l.Errorw("estimate gas failed for a deposit", "token", token.String(),
				"reserve_balance", balance, "amount", amount)
			if balance.Cmp(amount) < 0 {
				return tx, fmt.Errorf("estimate gas for reserve withdraw failed, possible balance not enough, ask %v, has %v", amount, balance)
			}
		}
	}
	return tx, err
}
