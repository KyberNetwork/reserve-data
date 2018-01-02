package exchange

import (
	"math/big"

	"sync"

	"github.com/KyberNetwork/reserve-data/common"
	ethereum "github.com/ethereum/go-ethereum/common"
)

type BittrexInterface interface {
	FetchOnePairData(wq *sync.WaitGroup, pair common.TokenPair, data *sync.Map, timepoint uint64)

	GetInfo(timepoint uint64) (Bittinfo, error)

	GetExchangeInfo() (BittExchangeInfo, error)

	Withdraw(
		token common.Token,
		amount *big.Int,
		address ethereum.Address,
		timepoint uint64) (Bittwithdraw, error)

	Trade(
		tradeType string,
		base, quote common.Token,
		rate, amount float64,
		timepoint uint64) (id string, done float64, remaining float64, finished bool, err error)

	CancelOrder(uuid string, timepoint uint64) (Bittcancelorder, error)

	DepositHistory(currency string, timepoint uint64) (Bittdeposithistory, error)

	WithdrawHistory(currency string, timepoint uint64) (Bittwithdrawhistory, error)

	OrderStatus(uuid string, timepoint uint64) (Bitttraderesult, error)
}
