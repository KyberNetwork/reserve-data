package exchange

import (
	"encoding/json"
	"fmt"
	"math/big"
	"testing"

	"github.com/KyberNetwork/reserve-data/common"
	ethereum "github.com/ethereum/go-ethereum/common"
)

type testBittrexInterface struct {
	DepositHistoryMock string
}

func (self testBittrexInterface) FetchOnePairData(pair common.TokenPair) (Bittresp, error) {
	return Bittresp{}, nil
}
func (self testBittrexInterface) GetExchangeInfo() (BittExchangeInfo, error) {
	return BittExchangeInfo{}, nil
}
func (self testBittrexInterface) GetInfo() (Bittinfo, error) {
	return Bittinfo{}, nil
}
func (self testBittrexInterface) Withdraw(
	token common.Token,
	amount *big.Int,
	address ethereum.Address) (Bittwithdraw, error) {
	return Bittwithdraw{}, nil
}
func (self testBittrexInterface) Trade(
	tradeType string,
	base, quote common.Token,
	rate, amount float64) (Bitttrade, error) {
	return Bitttrade{}, nil
}
func (self testBittrexInterface) CancelOrder(uuid string) (Bittcancelorder, error) {
	return Bittcancelorder{}, nil
}
func (self testBittrexInterface) DepositHistory(currency string) (Bittdeposithistory, error) {
	res := Bittdeposithistory{}
	err := json.Unmarshal([]byte(self.DepositHistoryMock), &res)
	fmt.Printf("%v\n", err)
	fmt.Printf("%v\n", res)
	return res, err
}
func (self testBittrexInterface) WithdrawHistory(currency string) (Bittwithdrawhistory, error) {
	return Bittwithdrawhistory{}, nil
}
func (self testBittrexInterface) OrderStatus(uuid string) (Bitttraderesult, error) {
	return Bitttraderesult{}, nil
}

func (self testBittrexInterface) GetAccountTradeHistory(base, quote common.Token) (BittTradeHistory, error) {
	return BittTradeHistory{}, nil
}

func (self testBittrexInterface) GetDepositAddress(currency string) (BittrexDepositAddress, error) {
	return BittrexDepositAddress{}, nil
}

type testBittrexStorage struct {
	IsNew bool
}

func (self *testBittrexStorage) IsNewBittrexDeposit(id uint64, actID common.ActivityID) bool {
	return self.IsNew
}

func (self *testBittrexStorage) RegisterBittrexDeposit(id uint64, actID common.ActivityID) error {
	self.IsNew = false
	return nil
}

func getTestBittrex(depositHistory string, registered bool) *Bittrex {
	return &Bittrex{
		testBittrexInterface{depositHistory},
		[]common.TokenPair{},
		[]common.Token{},
		common.NewExchangeAddresses(),
		&testBittrexStorage{registered},
		&common.ExchangeInfo{},
		common.ExchangeFees{},
	}
}

func TestDepositStatusNotFound(t *testing.T) {
	activityID := common.ActivityID{
		1513328774800747341,
		"0x4e3c6c5e5c56ef2f65867e0dac874f3e2ec2f66e05793b7a52281549c02e68d9|OMG|5",
	}
	bitt := getTestBittrex(
		`{"success":true,"message":"","result":[{"Id":46291182,"Amount":15.89963814,"Currency":"OMG","Confirmations":39,"LastUpdated":"2017-12-14T18:51:55.32","TxId":"0xb5273548bb8d3d33ac685c5797cdeb11490178bda9d8f7c9b6d2740eca18771f","CryptoAddress":"0x9db6e8d2d133448dbcf755f19d540253da4ba043"},{"Id":46191533,"Amount":25.00000000,"Currency":"OMG","Confirmations":53,"LastUpdated":"2017-12-14T10:14:53.19","TxId":"0x32fae94e542b36a409c0d602e342743f8bcda3d1e1e1e26022abe050cfaf80a6","CryptoAddress":"0x9db6e8d2d133448dbcf755f19d540253da4ba043"},{"Id":46150485,"Amount":0.31000000,"Currency":"OMG","Confirmations":42,"LastUpdated":"2017-12-14T06:28:08.83","TxId":"0x8a345f58910b99843e3ccd852f15cbb2601d455f39038de2dc08589e7c39e0a8","CryptoAddress":"0x9db6e8d2d133448dbcf755f19d540253da4ba043"}]}`,
		true,
	)
	out, err := bitt.DepositStatus(activityID, "0x4e3c6c5e5c56ef2f65867e0dac874f3e2ec2f66e05793b7a52281549c02e68d9", "OMG", 5, common.GetTimepoint())
	if err != nil {
		t.Fatalf("Expected convert successfully but got error: %v", err)
	} else {
		if out != "" {
			t.Fatalf("Expected \"\", got %v", out)
		}
	}
}

func TestDepositStatus(t *testing.T) {
	activityID := common.ActivityID{
		1513328774800747341,
		"0x4e3c6c5e5c56ef2f65867e0dac874f3e2ec2f66e05793b7a52281549c02e68d9|OMG|5",
	}
	bitt := getTestBittrex(
		`{"success":true,"message":"","result":[{"Id":46452872,"Amount":5.00000000,"Currency":"OMG","Confirmations":42,"LastUpdated":"2017-12-15T09:27:09.597","TxId":"0x15ccaab008f161efeee0febc3e32242846cea1fc93995e5abc6fb88d94ae7d21","CryptoAddress":"0x9db6e8d2d133448dbcf755f19d540253da4ba043"}]}`,
		true,
	)
	out, err := bitt.DepositStatus(activityID, "0x4e3c6c5e5c56ef2f65867e0dac874f3e2ec2f66e05793b7a52281549c02e68d9", "OMG", 5, common.GetTimepoint())
	if err != nil {
		t.Fatalf("Expected convert successfully but got error: %v", err)
	} else {
		if out != "done" {
			t.Fatalf("Expected done, got %v", out)
		}
	}
}

func TestDepositStatusIgnoreRegisteredHistory(t *testing.T) {
	activityID := common.ActivityID{
		1513328774800747341,
		"0x4e3c6c5e5c56ef2f65867e0dac874f3e2ec2f66e05793b7a52281549c02e68d9|OMG|5",
	}
	bitt := getTestBittrex(
		`{"success":true,"message":"","result":[{"Id":46452872,"Amount":5.00000000,"Currency":"OMG","Confirmations":42,"LastUpdated":"2017-12-15T09:27:09.597","TxId":"0x15ccaab008f161efeee0febc3e32242846cea1fc93995e5abc6fb88d94ae7d21","CryptoAddress":"0x9db6e8d2d133448dbcf755f19d540253da4ba043"}]}`,
		true,
	)
	out, err := bitt.DepositStatus(activityID, "0x4e3c6c5e5c56ef2f65867e0dac874f3e2ec2f66e05793b7a52281549c02e68d9", "OMG", 5, common.GetTimepoint())
	if err != nil {
		t.Fatalf("Expected convert successfully but got error: %v", err)
	} else {
		if out != "done" {
			t.Fatalf("Expected done, got %v", out)
		}
	}
	out, err = bitt.DepositStatus(activityID, "0x4e3c6c5e5c56ef2f65867e0dac874f3e2ec2f66e05793b7a52281549c02e68d9", "OMG", 5, common.GetTimepoint())
	if err != nil {
		t.Fatalf("Expected convert successfully but got error: %v", err)
	} else {
		if out != "" {
			t.Fatalf("Expected \"\", got %v", out)
		}
	}
}
