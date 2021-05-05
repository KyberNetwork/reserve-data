package storage

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/testutil"
	"github.com/KyberNetwork/reserve-data/exchange"
	rtypes "github.com/KyberNetwork/reserve-data/lib/rtypes"
	"github.com/KyberNetwork/reserve-data/reservesetting/storage/postgres"
)

func Test_BinancePostgres(t *testing.T) {
	t.Skip() // skip as we do not have a way to update database yet
	var storage exchange.BinanceStorage
	var err error
	db, tearDown := testutil.MustNewDevelopmentDB("../../../cmd/migrations")
	defer func() {
		assert.NoError(t, tearDown())
	}()
	_, err = postgres.NewStorage(db)
	require.NoError(t, err)
	storage, err = NewPostgresStorage(db)
	require.NoError(t, err)

	exchangeTradeHistory := common.ExchangeTradeHistory{
		1: []common.TradeHistory{
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
	} // this store should not create a new record, so the next GetTradeHistory won't break

	// get trade history
	var tradeHistory common.ExchangeTradeHistory
	fromTime := uint64(1528934400000)
	toTime := uint64(1529020800000)
	tradeHistory, err = storage.GetTradeHistory(rtypes.Binance, fromTime, toTime)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(tradeHistory, exchangeTradeHistory) {
		t.Fatalf("Get wrong trade history %+v", tradeHistory)
	}

	// get last trade history id
	var lastHistoryID string
	lastHistoryID, err = storage.GetLastIDTradeHistory(1)
	if err != nil {
		t.Fatalf("Get last trade history id error: %s", err.Error())
	}
	if lastHistoryID != "12342" {
		t.Fatalf("Get last trade history wrong")
	}
}

func Test_BinanceIntermediateTx(t *testing.T) {
	var err error
	db, tearDown := testutil.MustNewDevelopmentDB("../../../cmd/migrations")
	defer func() {
		assert.NoError(t, tearDown())
	}()
	_, err = postgres.NewStorage(db)
	require.NoError(t, err)
	storage, err := NewPostgresStorage(db)
	require.NoError(t, err)

	id := common.ActivityID{
		Timepoint: 1620381149259,
		EID:       "0x9ab1774626d4fdb88ac375ca442f18c5c9edff754d0676c09c4cd8faa0664ca3|BNB|0.5",
	}
	activity := common.TXEntry{
		Hash:     "0x9ab1774626d4fdb88ac375ca442f18c5c9edff754d0676c09c4cd8faa0664ca3",
		Exchange: "binance",
	}
	err = storage.StoreIntermediateDeposit(id, activity, true)
	require.NoError(t, err)

	txEntry, err := storage.GetPendingIntermediateTx(id)
	require.NoError(t, err)
	assert.Equal(t, "0x9ab1774626d4fdb88ac375ca442f18c5c9edff754d0676c09c4cd8faa0664ca3", txEntry.Hash)

	err = storage.StoreIntermediateDeposit(id, activity, false) // update activity to be done
	require.NoError(t, err)

	// this should error no row
	_, err = storage.GetPendingIntermediateTx(id)
	assert.Equal(t, err, sql.ErrNoRows)
}
