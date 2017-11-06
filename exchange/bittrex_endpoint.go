package exchange

import (
	"math/big"

	"github.com/KyberNetwork/reserve-data/common"
	ethereum "github.com/ethereum/go-ethereum/common"
	"sync"
)

type BittrexEndpoint interface {
	GetInfo(key string, signer Signer) (bittrexInfo, error)
	Withdraw(string, common.Token, *big.Int, ethereum.Address, Signer) error
	FetchOnePairData(wq *sync.WaitGroup, pair common.TokenPair, data *sync.Map)
	Trade(key string, tradeType string, base, quote common.Token, rate, amount float64, signer Signer) (done float64, remaining float64, finished bool, err error)
}
