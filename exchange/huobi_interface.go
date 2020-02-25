package exchange

import (
	"math/big"

	ethereum "github.com/ethereum/go-ethereum/common"

	commonv3 "github.com/KyberNetwork/reserve-data/reservesetting/common"
)

// HuobiInterface contains the methods to interact with Huobi centralized exchange.
type HuobiInterface interface {
	GetDepthOnePair(baseID, quoteID string) (HuobiDepth, error)

	GetInfo() (HuobiInfo, error)

	GetExchangeInfo() (HuobiExchangeInfo, error)

	GetDepositAddress(token string) (HuobiDepositAddress, error)

	GetAccountTradeHistory(baseSymbol, quoteSymbol string) (HuobiTradeHistory, error)

	Withdraw(
		asset commonv3.Asset,
		amount *big.Int,
		address ethereum.Address) (string, error)

	Trade(
		tradeType string,
		pair commonv3.TradingPairSymbols,
		rate, amount float64) (HuobiTrade, error)

	CancelOrder(symbol string, id uint64) (HuobiCancel, error)

	DepositHistory(size int) (HuobiDeposits, error)

	WithdrawHistory(size int) (HuobiWithdraws, error)

	OrderStatus(symbol string, id uint64) (HuobiOrder, error)

	OpenOrdersForOnePair(pair commonv3.TradingPairSymbols) (HuobiOpenOrders, error)
}
