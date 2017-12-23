package exchange

import (
	"math/big"
	"sync"

	"github.com/KyberNetwork/reserve-data/common"
	ethereum "github.com/ethereum/go-ethereum/common"
)

type BinanceInterface interface {
	FetchOnePairData(
		wg *sync.WaitGroup,
		pair common.TokenPair,
		data *sync.Map,
		timepoint uint64)

	StoreOrderBookData(
		wg *sync.WaitGroup,
		pair common.TokenPair,
		data *sync.Map,
		dataChannel chan Orderbook)

	SocketFetchOnePairData(
		wg *sync.WaitGroup,
		pair common.TokenPair,
		data *sync.Map,
		dataChannel chan Orderbook)

	SocketFetchAggTrade(
		pair common.TokenPair,
		dataChannel chan interface{})

	SocketGetUser(dataChannel chan interface{})

	OpenOrdersForOnePair(
		wg *sync.WaitGroup,
		pair common.TokenPair,
		data *sync.Map,
		timepoint uint64)

	GetInfo(timepoint uint64) (Binainfo, error)

	Withdraw(
		token common.Token,
		amount *big.Int,
		address ethereum.Address,
		timepoint uint64) (string, error)

	Trade(
		tradeType string,
		base, quote common.Token,
		rate, amount float64,
		timepoint uint64) (id string, done float64, remaining float64, finished bool, err error)

	CancelOrder(symbol string, id uint64) (Binacancel, error)

	DepositHistory(startTime, endTime uint64) (Binadeposits, error)

	WithdrawHistory(startTime, endTime uint64) (Binawithdrawals, error)

	OrderStatus(symbol string, id uint64, timepoint uint64) (Binaorder, error)
}
