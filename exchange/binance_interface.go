package exchange

import (
	"math/big"

	ethereum "github.com/ethereum/go-ethereum/common"

	commonv3 "github.com/KyberNetwork/reserve-data/reservesetting/common"
)

// BinanceInterface contains the methods to interact with Binance centralized exchange.
type BinanceInterface interface {
	GetDepthOnePair(baseID, quoteID string) (Binaresp, error)

	OpenOrdersForOnePair(pair commonv3.TradingPairSymbols) (Binaorders, error)

	GetInfo() (Binainfo, error)

	GetExchangeInfo() (BinanceExchangeInfo, error)

	GetDepositAddress(exchangeSymbol string) (Binadepositaddress, error)

	GetAccountTradeHistory(baseSymbol, quoteSymbol, fromID string) (BinaAccountTradeHistory, error)

	Withdraw(
		asset commonv3.Asset,
		amount *big.Int,
		address ethereum.Address) (string, error)

	Trade(
		tradeType string,
		pair commonv3.TradingPairSymbols,
		rate, amount float64) (Binatrade, error)

    OpenOrders(symbol string) ([]Binaorder, error)

	CancelOrder(symbol string, id uint64) (Binacancel, error)

	DepositHistory(startTime, endTime uint64) (Binadeposits, error)

	WithdrawHistory(startTime, endTime uint64) (Binawithdrawals, error)

	OrderStatus(symbol string, id uint64) (Binaorder, error)
}
