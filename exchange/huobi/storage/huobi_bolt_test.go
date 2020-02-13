package storage

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/KyberNetwork/reserve-data/common"
)

// Deprecated
func TestHuobiStoreTradeHistory(t *testing.T) {
	boltFile := "test_huobi_storage.db"
	tmpDir, err := ioutil.TempDir("", "huobi_storage")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if rErr := os.RemoveAll(tmpDir); rErr != nil {
			t.Error(rErr)
		}
	}()
	storage, err := NewBoltStorage(filepath.Join(tmpDir, boltFile))
	if err != nil {
		t.Fatalf("Could not init huobi bolt storage: %s", err.Error())
	}
	//Mock exchange history
	exchangeTradeHistory := common.ExchangeTradeHistory{
		2: []common.TradeHistory{
			{
				ID:        "12342",
				Price:     0.132131,
				Qty:       12.3123,
				Type:      "buy",
				Timestamp: 1528949872,
			},
		},
	}

	// store trade history
	err = storage.StoreTradeHistory(exchangeTradeHistory)
	if err != nil {
		t.Fatal(err)
	}

	// get trade history
	var tradeHistory common.ExchangeTradeHistory
	fromTime := uint64(1528934400000)
	toTime := uint64(1529020800000)
	tradeHistory, err = storage.GetTradeHistory(fromTime, toTime)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(tradeHistory, exchangeTradeHistory) {
		t.Fatal("Huobi get wrong trade history")
	}

}

// Deprecated
func TestHuobiStoreDepositActivity(t *testing.T) {
	boltFile := "test_huobi_storage.db"
	tmpDir, err := ioutil.TempDir("", "huobi_storage")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if rErr := os.RemoveAll(tmpDir); rErr != nil {
			t.Error(rErr)
		}
	}()
	storage, err := NewBoltStorage(filepath.Join(tmpDir, boltFile))
	if err != nil {
		t.Fatalf("Could not init huobi bolt storage: %s", err.Error())
	}

	// Mock intermediate deposit transaction and transaction id
	txEntry := common.TXEntry{
		Hash:           "0x884767eb42edef14b1a00759cf4050ea49a612424dbf484f8672fd17ebb92752",
		Exchange:       "huobi",
		AssetID:        1,
		MiningStatus:   "",
		ExchangeStatus: "",
		Amount:         10,
		Timestamp:      "1529307214",
	}
	txID := common.ActivityID{
		Timepoint: 1529307214,
		EID:       "0x884767eb42edef14b1a00759cf4050ea49a612424dbf484f8672fd17ebb92752|KNC|10",
	}

	// Test get pending intermediate TXs empty
	pendingIntermediateTxs, err := storage.GetPendingIntermediateTXs()
	if err != nil {
		t.Fatalf("Huobi get pending intermediate txs failed: %s", err.Error())
	}

	if len(pendingIntermediateTxs) != 0 {
		t.Fatalf("Huobi get pending intermediate txs expected empty get %d pending activities.", len(pendingIntermediateTxs))
	}

	// Test store pending intermediate Tx
	err = storage.StorePendingIntermediateTx(txID, txEntry)
	if err != nil {
		t.Fatalf("Huobi store pending intermediate failed: %s", err.Error())
	}

	// get pending tx
	pendingIntermediateTxs, err = storage.GetPendingIntermediateTXs()
	if err != nil {
		t.Fatalf("failed to get pending intermediate tx: %s", err.Error())
	}
	if len(pendingIntermediateTxs) != 1 {
		t.Fatalf("Huobi get pending intermediate txs expected 1 pending tx got %d pending txs.", len(pendingIntermediateTxs))
	}

	pendingTx, exist := pendingIntermediateTxs[txID]
	if !exist {
		t.Fatalf("Huobi store pending intermediate wrong. Expected key: %+v not found.", txID)
	}

	// check if pending tx is correct
	if equal := reflect.DeepEqual(txEntry, pendingTx); !equal {
		t.Fatalf("Huobi store pending wrong pending tx. Expected %+v, got %+v", txEntry, pendingTx)
	}

	// stored intermediate tx
	err = storage.StoreIntermediateTx(txID, txEntry)
	if err != nil {
		t.Fatalf("Huobi store intermediate tx failed: %s", err.Error())
	}

	// get intermediate tx
	tx, err := storage.GetIntermedatorTx(txID)
	if err != nil {
		t.Fatalf("Huobi get intermediate tx failed: %s", err.Error())
	}

	if equal := reflect.DeepEqual(tx, txEntry); !equal {
		t.Fatal("Huobi get intermediate tx wrong")
	}

	// Test pending intermediate tx shoud be removed
	pendingIntermediateTxs, err = storage.GetPendingIntermediateTXs()
	if err != nil {
		t.Fatalf("Huobi get pending intermediate tx failed: %s", err.Error())
	}

	_, exist = pendingIntermediateTxs[txID]
	if exist {
		t.Fatal("Huobi remove pending intermediate failed. Pending intermediate should be removed.")
	}
}
