package blockchain

import (
	"context"
	"math/big"
	"time"

	"github.com/KyberNetwork/reserve-data/common/blockchain"
	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func (bc *Blockchain) GeneratedWithdraw(opts blockchain.TxOpts, token ethereum.Address, amount *big.Int, destination ethereum.Address) (*types.Transaction, error) {
	timeout, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return bc.BuildTx(timeout, opts, bc.reserve, "withdraw", token, amount, destination)
}
