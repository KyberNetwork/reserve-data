package exchange

import (
	"math/big"
	"testing"

	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/KyberNetwork/reserve-data/common"
	commonv3 "github.com/KyberNetwork/reserve-data/reservesetting/common"
)

func TestBinance(t *testing.T) {

	binanceEndpoint := &binanceTestInterface{}
	binance, err := NewBinance(common.Binance, binanceEndpoint, nil, nil)
	require.NoError(t, err)
	t.Log(binance.ID())
}

// interface for testing
type binanceTestInterface struct {
}

func (bi *binanceTestInterface) GetDepthOnePair(baseID, quoteID string) (Binaresp, error) {
	panic("implement me")
}

func (bi *binanceTestInterface) OpenOrdersForOnePair(pair *commonv3.TradingPairSymbols) (Binaorders, error) {
	panic("implement me")
}

func (bi *binanceTestInterface) GetInfo() (Binainfo, error) {
	panic("implement me")
}

func (bi *binanceTestInterface) GetExchangeInfo() (BinanceExchangeInfo, error) {
	panic("implement me")
}

func (bi *binanceTestInterface) GetDepositAddress(tokenID string) (Binadepositaddress, error) {
	panic("implement me")
}

func (bi *binanceTestInterface) GetAccountTradeHistory(baseSymbol, quoteSymbol, fromID string) (BinaAccountTradeHistory, error) {
	panic("implement me")
}

func (bi *binanceTestInterface) Withdraw(
	asset commonv3.Asset,
	amount *big.Int,
	address ethereum.Address) (string, error) {
	panic("implement me")
}

func (bi *binanceTestInterface) Trade(
	tradeType string,
	pair commonv3.TradingPairSymbols,
	rate, amount float64) (Binatrade, error) {
	panic("implement me")
}

func (bi *binanceTestInterface) CancelOrder(symbol string, id uint64) (Binacancel, error) {
	panic("implement me")
}

func (bi *binanceTestInterface) DepositHistory(startTime, endTime uint64) (Binadeposits, error) {
	panic("implement me")
}

func (bi *binanceTestInterface) WithdrawHistory(startTime, endTime uint64) (Binawithdrawals, error) {
	panic("implement me")
}

func (bi *binanceTestInterface) OrderStatus(symbol string, id uint64) (Binaorder, error) {
	panic("implement me")
}
