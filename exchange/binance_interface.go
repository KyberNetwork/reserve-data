package exchange

import (
	"math/big"

	"github.com/KyberNetwork/reserve-data/common"
	ethereum "github.com/ethereum/go-ethereum/common"
)

type BinanceInterface interface {
	GetDepthOnePair(pair common.TokenPair) (Binaresp, error)

	OpenOrdersForOnePair(pair common.TokenPair) (Binaorders, error)

	GetInfo() (Binainfo, error)

	GetExchangeInfo() (BinanceExchangeInfo, error)

	GetDepositAddress(tokenID string) (Binadepositaddress, error)

	GetAccountTradeHistory(base, quote common.Token, fromID string) (BinaAccountTradeHistory, error)

	Withdraw(
		token common.Token,
		amount *big.Int,
		address ethereum.Address) (string, error)

	Trade(
		tradeType string,
		base, quote common.Token,
		rate, amount float64) (Binatrade, error)

	CancelOrder(symbol string, id uint64) (Binacancel, error)

	DepositHistory(startTime, endTime uint64) (Binadeposits, error)

	WithdrawHistory(
		startTime, endTime uint64) (Binawithdrawals, error)

	OrderStatus(
		symbol string, id uint64) (Binaorder, error)

	GetServerTime() (uint64, error)
}
