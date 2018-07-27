package blockchain

import (
	"context"
	"errors"
	"log"
	"math/big"
	"time"

	ether "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/ethclient"
)

type ContractCaller struct {
	clients []*ethclient.Client
	urls    []string
}

func NewContractCaller(clients []*ethclient.Client, urls []string) *ContractCaller {
	return &ContractCaller{
		clients: clients,
		urls:    urls,
	}
}

func (cc ContractCaller) CallContract(msg ether.CallMsg, blockNo *big.Int, timeOut time.Duration) ([]byte, error) {
	for i, client := range cc.clients {
		url := cc.urls[i]

		output, err := func() ([]byte, error) {
			ctx, cancel := context.WithTimeout(context.Background(), timeOut)
			defer cancel()
			return client.CallContract(ctx, msg, blockNo)
		}()
		if err != nil {
			log.Printf("FALLBACK: Ether client %s done, getting err %v, trying next one...", url, err)
			continue
		}

		log.Printf("FALLBACK: Ether client %s done, returnning result...", url)
		return output, nil
	}
	return nil, errors.New("failed to call contract, all clients failed")
}
