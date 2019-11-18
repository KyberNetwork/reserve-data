package blockchain

import (
	"context"
	"errors"
	"math/big"
	"time"

	ether "github.com/ethereum/go-ethereum"
	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data/common"
)

type ContractCaller struct {
	clients []*common.EthClient
	l       *zap.SugaredLogger
}

func NewContractCaller(clients []*common.EthClient) *ContractCaller {
	return &ContractCaller{
		clients: clients,
		l:       zap.S(),
	}
}

func (c ContractCaller) CallContract(msg ether.CallMsg, blockNo *big.Int, timeOut time.Duration) ([]byte, error) {
	for _, client := range c.clients {
		ctx, cancel := context.WithTimeout(context.Background(), timeOut)
		output, err := client.CallContract(ctx, msg, blockNo)
		cancel()
		if err != nil {
			c.l.Infow("contract call failed, fallback to next node",
				"current_client", client.URL, "err", err)
			continue
		}
		return output, nil
	}
	return nil, errors.New("failed to call contract, all clients failed")
}
