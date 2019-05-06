package core

import (
	"fmt"
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
	"github.com/stretchr/testify/assert"
)

type testExchange struct {
}

func (te testExchange) ID() common.ExchangeID {
	return "bittrex"
}

func (te testExchange) Address(token common.Token) (address ethereum.Address, supported bool) {
	return ethereum.Address{}, true
}
func (te testExchange) Withdraw(token common.Token, amount *big.Int, address ethereum.Address, timepoint uint64) (string, error) {
	return "withdrawid", nil
}
func (te testExchange) Trade(tradeType string, base common.Token, quote common.Token, rate float64, amount float64, timepoint uint64) (id string, done float64, remaining float64, finished bool, err error) {
	return "tradeid", 10, 5, false, nil
}
func (te testExchange) CancelOrder(id string, base, quote string) error {
	return nil
}
func (te testExchange) MarshalText() (text []byte, err error) {
	return []byte("bittrex"), nil
}
func (te testExchange) GetExchangeInfo(pair common.TokenPairID) (common.ExchangePrecisionLimit, error) {
	return common.ExchangePrecisionLimit{}, nil
}
func (te testExchange) GetFee() (common.ExchangeFees, error) {
	return common.ExchangeFees{}, nil
}
func (te testExchange) GetMinDeposit() (common.ExchangesMinDeposit, error) {
	return common.ExchangesMinDeposit{}, nil
}
func (te testExchange) GetInfo() (common.ExchangeInfo, error) {
	return common.ExchangeInfo{}, nil
}
func (te testExchange) UpdateDepositAddress(token common.Token, address string) error {
	return nil
}

func (te testExchange) GetTradeHistory(fromTime, toTime uint64) (common.ExchangeTradeHistory, error) {
	return common.ExchangeTradeHistory{}, nil
}

func (te testExchange) GetLiveExchangeInfos(pairIDs []common.TokenPairID) (common.ExchangeInfo, error) {
	return common.ExchangeInfo{}, nil
}

type testBlockchain struct {
}

func (tbc testBlockchain) Send(
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

func (tbc testBlockchain) SetRateMinedNonce() (uint64, error) {
	return 0, nil
}

type testActivityStorage struct {
	PendingDeposit  bool
	PendingDeposits map[string]bool
}

func (tas testActivityStorage) Record(
	action string,
	id common.ActivityID,
	destination string,
	params map[string]interface{},
	result map[string]interface{},
	estatus string,
	mstatus string,
	timepoint uint64) error {
	if !tas.PendingDeposit {
		return nil
	}
	if action == common.ActionDeposit {
		token := params["token"].(common.Token)
		exchange := params["exchange"].(common.Exchange)
		tokenExchange := fmt.Sprintf("%s_%s", token.ID, exchange.ID())
		log.Printf("Logging deposit for %s", tokenExchange)
		tas.PendingDeposits[tokenExchange] = true
	}

	return nil
}

func (tas testActivityStorage) GetActivity(id common.ActivityID) (common.ActivityRecord, error) {
	return common.ActivityRecord{}, nil
}

func (tas testActivityStorage) PendingSetRate(minedNonce uint64) (*common.ActivityRecord, uint64, error) {
	return nil, 0, nil
}

func (tas testActivityStorage) HasPendingDeposit(token common.Token, exchange common.Exchange) (bool, error) {
	tokenExchange := fmt.Sprintf("%s_%s", token.ID, exchange.ID())
	log.Printf("Checking deposit for %s", tokenExchange)
	if _, avail := tas.PendingDeposits[tokenExchange]; avail {
		return tas.PendingDeposit, nil

	}
	return false, nil
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
	addressSetting := &settings.AddressSetting{}
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
		testActivityStorage{hasPendingDeposit, make(map[string]bool)},
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
	if err != nil {
		t.Fatalf("Expected deposit succesfully the first time it is called")
	}

	_, err = core.Deposit(
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

func TestCalculateNewGasPrice(t *testing.T) {
	initPrice := common.GweiToWei(1)
	newPrice := calculateNewGasPrice(initPrice, 0)
	if newPrice.Cmp(newPrice) != 0 {
		t.Errorf("new price is not equal to initial price with count == 0")
	}

	prevPrice := initPrice
	for count := uint64(1); count < 10; count++ {
		newPrice = calculateNewGasPrice(initPrice, count)
		if newPrice.Cmp(prevPrice) != 1 {
			t.Errorf("new price %s is not higher than previous price %s",
				newPrice.String(),
				prevPrice.String())
		}
		t.Logf("new price: %s", newPrice.String())
		prevPrice = newPrice
	}
}

func TestRaceConditionWithMultipleDeposit(t *testing.T) {
	var (
		nThread = 10
		errChn  = make(chan error, nThread)
		nErr    = 0
	)
	core := getTestCore(true)
	for i := 0; i < nThread; i++ {
		go func(i int) {
			_, err := core.Deposit(
				testExchange{},
				common.NewToken("OMG", "omise-go", "0x1111111111111111111111111111111111111111", 18, true, true, 0),
				big.NewInt(10),
				common.GetTimepoint(),
			)
			errChn <- err

		}(i)
	}
	for i := 0; i < nThread; i++ {
		select {
		case err := <-errChn:
			if err != nil {
				nErr++
			}
		}
	}
	assert.Equal(t, nThread-1, nErr)
}
