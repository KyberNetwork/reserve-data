package core

import (
	"errors"
	"math/big"
	"testing"
	"time"

	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/blockchain"
	"github.com/KyberNetwork/reserve-data/common/gasinfo"
	gaspricedataclient "github.com/KyberNetwork/reserve-data/common/gaspricedata-client"
	"github.com/KyberNetwork/reserve-data/lib/rtypes"
	commonv3 "github.com/KyberNetwork/reserve-data/reservesetting/common"
)

type testExchange struct {
}

func (te testExchange) Transfer(fromAccount string, toAccount string, asset commonv3.Asset, amount *big.Int) (string, error) {
	return "", nil
}
func (te testExchange) ID() rtypes.ExchangeID {
	return rtypes.Binance
}

func (te testExchange) Address(_ commonv3.Asset) (address ethereum.Address, supported bool) {
	return ethereum.Address{}, true
}
func (te testExchange) Withdraw(token commonv3.Asset, amount *big.Int, address ethereum.Address) (string, error) {
	return "withdrawid", nil
}
func (te testExchange) Trade(tradeType string, pair commonv3.TradingPairSymbols, rate float64, amount float64) (id string, done float64, remaining float64, finished bool, err error) {
	return "tradeid", 10, 5, false, nil
}
func (te testExchange) CancelOrder(id, symbol string) error {
	return nil
}
func (te testExchange) CancelAllOrders(symbol string) error {
	return nil
}
func (te testExchange) MarshalText() (text []byte, err error) {
	return []byte("bittrex"), nil
}

func (te testExchange) GetTradeHistory(fromTime, toTime uint64) (common.ExchangeTradeHistory, error) {
	return common.ExchangeTradeHistory{}, nil
}

func (te testExchange) GetLiveExchangeInfos(pairs []commonv3.TradingPairSymbols) (common.ExchangeInfo, error) {
	return common.ExchangeInfo{}, nil
}

func (te testExchange) OpenOrders(pair commonv3.TradingPairSymbols) ([]common.Order, error) {
	return nil, nil
}

// GetLiveWithdrawFee ...
func (te testExchange) GetLiveWithdrawFee(asset string) (float64, error) {
	return 0.1, nil
}

type testBlockchain struct {
}

func (tbc testBlockchain) TransferToSelf(op string, gasPrice *big.Int, nonce *big.Int) (*types.Transaction, error) {
	panic("implement me")
}

func (tbc testBlockchain) SpeedupDeposit(tx ethereum.Hash, recommendGasPrice float64, maxGasPrice float64) (*types.Transaction, error) {
	panic("implement me")
}

func (tbc testBlockchain) BuildSendETHTx(opts blockchain.TxOpts, to ethereum.Address) (*types.Transaction, error) {
	return nil, errors.New("not supported")
}

func (tbc testBlockchain) GetDepositOPAddress() ethereum.Address {
	return ethereum.Address{}
}

func (tbc testBlockchain) CurrentBlock() (uint64, error) {
	return 0, nil
}

func (tbc testBlockchain) SignAndBroadcast(tx *types.Transaction, from string) (*types.Transaction, error) {
	return nil, errors.New("not supported")
}

func (tbc testBlockchain) Send(
	asset commonv3.Asset,
	amount *big.Int,
	address ethereum.Address,
	nonce *big.Int,
	gasPrice *big.Int) (*types.Transaction, error) {
	tx := types.NewTransaction(
		0,
		ethereum.Address{},
		big.NewInt(0),
		300000,
		big.NewInt(1000000000),
		[]byte{})
	return tx, nil
}

func (tbc testBlockchain) SetRates(
	tokens []ethereum.Address,
	buys []*big.Int,
	sells []*big.Int,
	block *big.Int,
	nonce *big.Int,
	gasPrice *big.Int) (*types.Transaction, error) {
	tx := types.NewTransaction(
		0,
		ethereum.Address{},
		big.NewInt(0),
		300000,
		big.NewInt(1000000000),
		[]byte{})
	return tx, nil
}

func (tbc testBlockchain) StandardGasPrice() float64 {
	return 0
}

func (tbc testBlockchain) GetMinedNonceWithOP(string) (uint64, error) {
	return 0, nil
}

type testActivityStorage struct {
	PendingDeposit bool
}

func (tas testActivityStorage) GetActivityForOverride(action string, minedNonce uint64) (*common.ActivityRecord, error) {
	panic("implement me")
}

func (tas testActivityStorage) GetPendingSetRate(action string, minedNonce uint64) (*common.ActivityRecord, error) {
	panic("implement me")
}

func (tas testActivityStorage) MaxPendingNonce(action string) (int64, error) {
	return 0, nil
}

func (tas testActivityStorage) Record(
	action string,
	id common.ActivityID,
	destination string,
	params common.ActivityParams,
	result common.ActivityResult,
	estatus string,
	mstatus string,
	timepoint uint64,
	isPending bool,
	orgTime uint64,
) error {
	return nil
}

func (tas testActivityStorage) GetActivity(exchangeID rtypes.ExchangeID, id string) (common.ActivityRecord, error) {
	return common.ActivityRecord{}, nil
}

func (tas testActivityStorage) PendingActivityForAction(minedNonce uint64, activityType string) (*common.ActivityRecord, uint64, error) {
	return nil, 0, nil
}

func (tas testActivityStorage) HasPendingDeposit(token commonv3.Asset, exchange common.Exchange) (bool, error) {
	if token.Symbol == "OMG" && exchange.ID().String() == "binance" {
		return tas.PendingDeposit, nil
	}
	return false, nil
}

func getTestCore(hasPendingDeposit bool) *ReserveCore {
	addressSetting := &common.ContractAddressConfiguration{}

	return NewReserveCore(testBlockchain{}, testActivityStorage{hasPendingDeposit}, addressSetting,
		gasinfo.NewGasPriceInfo(gasinfo.NewConstGasPriceLimiter(100), &ExampleGasConfig{}, &ExampleGasClient{}))
}

type ExampleGasConfig struct {
}

func (e ExampleGasConfig) SetPreferGasSource(v commonv3.PreferGasSource) error {
	panic("implement me")
}

func (e ExampleGasConfig) GetPreferGasSource() (commonv3.PreferGasSource, error) {
	return commonv3.PreferGasSource{Name: "ethgasstation"}, nil
}

type ExampleGasClient struct {
}

func (e ExampleGasClient) GetGas() (gaspricedataclient.GasResult, error) {
	return gaspricedataclient.GasResult{
		"ethgasstation": gaspricedataclient.SourceData{
			Value: gaspricedataclient.Data{
				Fast:     100,
				Standard: 80,
				Slow:     50,
			},
			Timestamp: time.Now().Unix() * 1000,
		},
	}, nil
}

func TestNotAllowDeposit(t *testing.T) {
	core := getTestCore(true)
	_, err := core.Deposit(
		testExchange{},
		commonv3.Asset{
			ID:                 0,
			Symbol:             "OMG",
			Name:               "omise-go",
			Address:            ethereum.HexToAddress("0x1111111111111111111111111111111111111111"),
			OldAddresses:       nil,
			Decimals:           12,
			Transferable:       true,
			SetRate:            commonv3.SetRateNotSet,
			Rebalance:          false,
			IsQuote:            false,
			PWI:                nil,
			RebalanceQuadratic: nil,
			Exchanges:          nil,
			Target:             nil,
			Created:            time.Now(),
			Updated:            time.Now(),
		},
		big.NewInt(10),
		common.NowInMillis(),
	)
	if err == nil {
		t.Fatalf("Expected to return an error protecting user from deposit when there is another pending deposit")
	}
	_, err = core.Deposit(
		testExchange{},
		commonv3.Asset{
			ID:                 0,
			Symbol:             "KNC",
			Name:               "Kyber Network Crystal",
			Address:            ethereum.HexToAddress("0x1111111111111111111111111111111111111111"),
			OldAddresses:       nil,
			Decimals:           12,
			Transferable:       true,
			SetRate:            commonv3.SetRateNotSet,
			Rebalance:          false,
			IsQuote:            false,
			PWI:                nil,
			RebalanceQuadratic: nil,
			Exchanges:          nil,
			Target:             nil,
			Created:            time.Now(),
			Updated:            time.Now(),
		},
		big.NewInt(10),
		common.NowInMillis(),
	)
	if err != nil {
		t.Fatalf("Expected to be able to deposit different token")
	}
}

func TestCalculateNewGasPrice(t *testing.T) {
	initPrice := common.GweiToWei(1)
	newPrice := calculateNewGasPrice(initPrice, 0, 100.0)
	if newPrice.Cmp(newPrice) != 0 {
		t.Errorf("new price is not equal to initial price with count == 0")
	}

	prevPrice := initPrice
	for count := uint64(1); count < 10; count++ {
		newPrice = calculateNewGasPrice(initPrice, count, 100.0)
		if newPrice.Cmp(prevPrice) != 1 {
			t.Errorf("new price %s is not higher than previous price %s",
				newPrice.String(),
				prevPrice.String())
		}
		t.Logf("new price: %s", newPrice.String())
		prevPrice = newPrice
	}
}
