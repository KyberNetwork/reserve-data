package storage

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	ethereum "github.com/ethereum/go-ethereum/common"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/data/testutil"
	commonv3 "github.com/KyberNetwork/reserve-data/reservesetting/common"
)

func TestHasPendingDepositBoltStorage(t *testing.T) {
	boltFile := "test_bolt.db"
	tmpDir, err := ioutil.TempDir("", "pending_deposit")
	if err != nil {
		t.Fatal(err)
	}
	storage, err := NewBoltStorage(filepath.Join(tmpDir, boltFile))
	if err != nil {
		t.Fatalf("Couldn't init bolt storage %v", err)
	}
	exchange := common.TestExchange{}
	timepoint := common.NowInMillis()
	asset := commonv3.Asset{
		ID:                 1,
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
		AssetExchanges:     nil,
		Target:             nil,
		Created:            time.Now(),
		Updated:            time.Now(),
	}
	out, err := storage.HasPendingDeposit(asset, exchange)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if out != false {
		t.Fatalf("Expected ram storage to return true false there is no pending deposit for the same currency and exchange")
	}
	err = storage.Record(
		"deposit",
		common.NewActivityID(1, "1"),
		exchange.ID().String(),
		common.ActivityParams{
			Exchange:  exchange.ID(),
			Asset:     asset.ID,
			Amount:    1.0,
			Timepoint: timepoint,
		},
		common.ActivityResult{
			Tx:    "",
			Error: "",
		},
		"",
		"submitted",
		common.NowInMillis())
	if err != nil {
		t.Fatalf("Store activity error: %s", err.Error())
	}
	b, err := storage.HasPendingDeposit(commonv3.Asset{
		ID:                 1,
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
		AssetExchanges:     nil,
		Target:             nil,
		Created:            time.Now(),
		Updated:            time.Now(),
	}, exchange)
	if err != nil {
		t.Fatalf(err.Error())
	}
	out = b
	if out != true {
		t.Fatalf("Expected ram storage to return true when there is pending deposit")
	}

	if err = os.RemoveAll(tmpDir); err != nil {
		t.Error(err)
	}
}

func TestGlobalStorageBoltImplementation(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "test_bolt_storage")
	if err != nil {
		t.Fatal(err)
	}
	storage, err := NewBoltStorage(filepath.Join(tmpDir, "test_bolt.db"))
	if err != nil {
		t.Fatal(err)
	}
	testutil.NewGlobalStorageTestSuite(t, storage, storage).Run()

	if err = os.RemoveAll(tmpDir); err != nil {
		t.Error(err)
	}
}
