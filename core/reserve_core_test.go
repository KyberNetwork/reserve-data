package core

import (
	"io/ioutil"
	"log"
	"math/big"
	"path/filepath"
	"testing"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/settings"
	"github.com/KyberNetwork/reserve-data/settings/storage"
	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type testExchange struct {
}

func (self testExchange) ID() common.ExchangeID {
	return "bittrex"
}

func (self testExchange) Address(token common.Token) (address ethereum.Address, supported bool) {
	return ethereum.Address{}, true
}
func (self testExchange) Withdraw(token common.Token, amount *big.Int, address ethereum.Address, timepoint uint64) (string, error) {
	return "withdrawid", nil
}
func (self testExchange) Trade(tradeType string, base common.Token, quote common.Token, rate float64, amount float64, timepoint uint64) (id string, done float64, remaining float64, finished bool, err error) {
	return "tradeid", 10, 5, false, nil
}
func (self testExchange) CancelOrder(id string, base, quote string) error {
	return nil
}
func (self testExchange) MarshalText() (text []byte, err error) {
	return []byte("bittrex"), nil
}
func (self testExchange) GetExchangeInfo(pair common.TokenPairID) (common.ExchangePrecisionLimit, error) {
	return common.ExchangePrecisionLimit{}, nil
}
func (self testExchange) GetFee() (common.ExchangeFees, error) {
	return common.ExchangeFees{}, nil
}
func (self testExchange) GetMinDeposit() (common.ExchangesMinDeposit, error) {
	return common.ExchangesMinDeposit{}, nil
}
func (self testExchange) GetInfo() (common.ExchangeInfo, error) {
	return common.ExchangeInfo{}, nil
}
func (self testExchange) TokenAddresses() (map[string]ethereum.Address, error) {
	return map[string]ethereum.Address{}, nil
}
func (self testExchange) UpdateDepositAddress(token common.Token, address string) error {
	return nil
}

func (self testExchange) GetTradeHistory(fromTime, toTime uint64) (common.ExchangeTradeHistory, error) {
	return common.ExchangeTradeHistory{}, nil
}

func (self testExchange) GetLiveExchangeInfos(pairIDs []common.TokenPairID) (common.ExchangeInfo, error) {
	return common.ExchangeInfo{}, nil
}

type testBlockchain struct {
}

func (self testBlockchain) Send(
	token common.Token,
	amount *big.Int,
	address ethereum.Address) (*types.Transaction, error) {
	tx := types.NewTransaction(
		0,
		ethereum.Address{},
		big.NewInt(0),
		300000,
		big.NewInt(1000000000),
		[]byte{})
	return tx, nil
}

func (self testBlockchain) SetRates(
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

func (self testBlockchain) StandardGasPrice() float64 {
	return 0
}

func (self testBlockchain) SetRateMinedNonce() (uint64, error) {
	return 0, nil
}

func (self testBlockchain) GetAddresses() (*common.Addresses, error) {
	return &common.Addresses{}, nil
}

type testActivityStorage struct {
	PendingDeposit bool
}

func (self testActivityStorage) Record(
	action string,
	id common.ActivityID,
	destination string,
	params map[string]interface{},
	result map[string]interface{},
	estatus string,
	mstatus string,
	timepoint uint64) error {
	return nil
}

func (self testActivityStorage) GetActivity(id common.ActivityID) (common.ActivityRecord, error) {
	return common.ActivityRecord{}, nil
}

func (self testActivityStorage) PendingSetrate(minedNonce uint64) (*common.ActivityRecord, uint64, error) {
	return nil, 0, nil
}

func (self testActivityStorage) HasPendingDeposit(token common.Token, exchange common.Exchange) (bool, error) {
	if token.ID == "OMG" && exchange.ID() == "bittrex" {
		return self.PendingDeposit, nil
	} else {
		return false, nil
	}
}

func getTestCore(hasPendingDeposit bool) *ReserveCore {
	tmpDir, err := ioutil.TempDir("", "core_test")
	if err != nil {
		log.Fatal(err)
	}
	boltSettingStorage, err := storage.NewBoltSettingStorage(filepath.Join(tmpDir, "setting.db"))
	if err != nil {
		log.Fatal(err)
	}
	tokenSetting, err := settings.NewTokenSetting(boltSettingStorage)
	if err != nil {
		log.Fatal(err)
	}
	addressSetting, err := settings.NewAddressSetting(boltSettingStorage)
	if err != nil {
		log.Fatal(err)
	}
	exchangeSetting, err := settings.NewExchangeSetting(boltSettingStorage)
	if err != nil {
		log.Fatal(err)
	}

	setting, err := settings.NewSetting(tokenSetting, addressSetting, exchangeSetting)
	if err != nil {
		log.Fatal(err)
	}
	return NewReserveCore(
		testBlockchain{},
		testActivityStorage{hasPendingDeposit},
		setting,
	)
}

func TestNotAllowDeposit(t *testing.T) {
	core := getTestCore(true)
	_, err := core.Deposit(
		testExchange{},
		common.NewToken("OMG", "omise-go", "0x1111111111111111111111111111111111111111", 18, true, true, 0),
		big.NewInt(10),
		common.GetTimepoint(),
	)
	if err == nil {
		t.Fatalf("Expected to return an error protecting user from deposit when there is another pending deposit")
	}
	_, err = core.Deposit(
		testExchange{},
		common.NewToken("KNC", "Kyber-coin", "0x1111111111111111111111111111111111111111", 18, true, true, 0),
		big.NewInt(10),
		common.GetTimepoint(),
	)
	if err != nil {
		t.Fatalf("Expected to be able to deposit different token")
	}
}
