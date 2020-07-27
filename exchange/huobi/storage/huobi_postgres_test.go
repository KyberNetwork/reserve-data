package storage

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/testutil"
	"github.com/KyberNetwork/reserve-data/exchange"
)

const (
	migrationPath = "../../../cmd/migrations"
)

func TestPostgresStorage_TradeHistory(t *testing.T) {
	var storage exchange.HuobiStorage
	var err error

	db, tearDown := testutil.MustNewDevelopmentDB(migrationPath)
	defer func() {
		assert.NoError(t, tearDown())
	}()
	storage, err = NewPostgresStorage(db)
	require.NoError(t, err)
	//Mock exchange history
	exchangeTradeHistory := common.ExchangeTradeHistory{
		2: []common.TradeHistory{
			{
				ID:        "12342",
				Price:     0.132131,
				Qty:       12.3123,
				Type:      "buy",
				Timestamp: 1528949872000,
			},
		},
	}

	// store trade history
	err = storage.StoreTradeHistory(exchangeTradeHistory)
	if err != nil {
		t.Fatal(err)
	}
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

func TestPostgresStorage_DepositActivity(t *testing.T) {
	var storage exchange.HuobiStorage
	var err error

	db, tearDown := testutil.MustNewDevelopmentDB(migrationPath)
	defer func() {
		assert.NoError(t, tearDown())
	}()
	storage, err = NewPostgresStorage(db)
	require.NoError(t, err)

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
		t.Fatalf("Huobi store intermedate tx failed: %s", err.Error())
	}

	// get intermedate tx
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
