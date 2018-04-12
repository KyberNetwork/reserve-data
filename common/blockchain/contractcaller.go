package blockchain

import (
	"context"
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

func (self ContractCaller) CallContract(msg ether.CallMsg, blockNo *big.Int, timeOut time.Duration) (output []byte, err error) {
	for i, client := range self.clients {
		urlstring := self.urls[i]
		ctx, cancel := context.WithTimeout(context.Background(), timeOut)

		defer cancel()
		output, err = client.CallContract(ctx, msg, blockNo)
		if err != nil {
			log.Printf("FALLBACK: Ether client %s done, getting err %v, trying next one...", urlstring, err)
		} else {
			log.Printf("FALLBACK: Ether client %s done, returnning result...", urlstring)
			return
		}
	}
	return
}
