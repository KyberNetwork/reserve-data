package fetcher

import (
	"log"
	"os"
	"path"
	"sync"
	"testing"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/data/fetcher/http_runner"
	"github.com/KyberNetwork/reserve-data/data/storage"
	ethereum "github.com/ethereum/go-ethereum/common"
)

func TestUnchangedFunc(t *testing.T) {
	// test different len
	a1 := map[common.ActivityID]common.ActivityStatus{
		common.ActivityID{1, "1"}: common.ActivityStatus{
			"done", "0x123", 0, "mined", nil,
		},
	}
	b1 := map[common.ActivityID]common.ActivityStatus{
		common.ActivityID{1, "1"}: common.ActivityStatus{
			"done", "0x123", 0, "mined", nil,
		},
		common.ActivityID{2, "1"}: common.ActivityStatus{
			"done", "0x123", 0, "mined", nil,
		},
	}
	if unchanged(a1, b1) != false {
		t.Fatalf("Expected unchanged() to return false, got true")
	}
	// test different id
	a1 = map[common.ActivityID]common.ActivityStatus{
		common.ActivityID{1, "1"}: common.ActivityStatus{
			"done", "0x123", 0, "mined", nil,
		},
	}
	b1 = map[common.ActivityID]common.ActivityStatus{
		common.ActivityID{2, "1"}: common.ActivityStatus{
			"done", "0x123", 0, "mined", nil,
		},
	}
	if unchanged(a1, b1) != false {
		t.Fatalf("Expected unchanged() to return false, got true")
	}
	// test different exchange status
	a1 = map[common.ActivityID]common.ActivityStatus{
		common.ActivityID{1, "1"}: common.ActivityStatus{
			"", "0x123", 0, "mined", nil,
		},
	}
	b1 = map[common.ActivityID]common.ActivityStatus{
		common.ActivityID{1, "1"}: common.ActivityStatus{
			"done", "0x123", 0, "mined", nil,
		},
	}
	if unchanged(a1, b1) != false {
		t.Fatalf("Expected unchanged() to return false, got true")
	}
	// test different mining status
	a1 = map[common.ActivityID]common.ActivityStatus{
		common.ActivityID{1, "1"}: common.ActivityStatus{
			"done", "0x123", 0, "mined", nil,
		},
	}
	b1 = map[common.ActivityID]common.ActivityStatus{
		common.ActivityID{1, "1"}: common.ActivityStatus{
			"done", "0x123", 0, "", nil,
		},
	}
	if unchanged(a1, b1) != false {
		t.Fatalf("Expected unchanged() to return false, got true")
	}
	// test different tx
	a1 = map[common.ActivityID]common.ActivityStatus{
		common.ActivityID{1, "1"}: common.ActivityStatus{
			"done", "0x123", 0, "mined", nil,
		},
	}
	b1 = map[common.ActivityID]common.ActivityStatus{
		common.ActivityID{1, "1"}: common.ActivityStatus{
			"done", "0x124", 0, "mined", nil,
		},
	}
	if unchanged(a1, b1) != false {
		t.Fatalf("Expected unchanged() to return false, got true")
	}
	// test identical statuses
	a1 = map[common.ActivityID]common.ActivityStatus{
		common.ActivityID{1, "1"}: common.ActivityStatus{
			"done", "0x123", 0, "mined", nil,
		},
	}
	b1 = map[common.ActivityID]common.ActivityStatus{
		common.ActivityID{1, "1"}: common.ActivityStatus{
			"done", "0x123", 0, "mined", nil,
		},
	}
	if unchanged(a1, b1) != true {
		t.Fatalf("Expected unchanged() to return true, got false")
	}
}

func TestExchangeDown(t *testing.T) {
	// mock fetcher
	testFetcherStoragePath := path.Join(common.CurrentDir(), "test_fetcher.db")
	fstorage, err := storage.NewBoltStorage(testFetcherStoragePath)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer os.Remove(testFetcherStoragePath)
	runner := http_runner.NewHttpRunner(9000)
	fetcher := NewFetcher(fstorage, runner, ethereum.Address{}, true)

	// mock normal data
	var estatuses, bstatuses sync.Map
	ebalanceValue := common.EBalanceEntry{
		Valid:      true,
		Error:      "",
		Timestamp:  common.GetTimestamp(),
		ReturnTime: common.GetTimestamp(),
		AvailableBalance: map[string]float64{
			"KNC": 500,
		},
		LockedBalance:  map[string]float64{},
		DepositBalance: map[string]float64{},
		Status:         true,
	}
	ebalance := sync.Map{}
	ebalance.Store(common.ExchangeID("binance"), ebalanceValue)

	rawBalance := common.RawBalance{}
	tokenBalance := common.BalanceEntry{
		Valid:      true,
		Error:      "",
		Timestamp:  common.GetTimestamp(),
		ReturnTime: common.GetTimestamp(),
		Balance:    rawBalance,
	}

	bbalance := map[string]common.BalanceEntry{
		"KNC": tokenBalance,
	}

	// empty pending activities
	pendings := []common.ActivityRecord{}

	var snapshot common.AuthDataSnapshot
	timepoint := common.GetTimepoint()

	// Persist normal auth snapshot
	err = fetcher.PersistSnapshot(&ebalance, bbalance, &estatuses, &bstatuses, pendings, &snapshot, timepoint)
	if err != nil {
		t.Fatalf("Cannot persist snapshot: %s", err.Error())
	}

	// mock empty data as exchange down
	ebalanceValue = common.EBalanceEntry{
		Valid:            false,
		Error:            "Connection time out",
		Timestamp:        common.GetTimestamp(),
		ReturnTime:       common.GetTimestamp(),
		AvailableBalance: map[string]float64{},
		LockedBalance:    map[string]float64{},
		DepositBalance:   map[string]float64{},
		Status:           false, // exchange status false - down, true - up
	}
	ebalance.Store(common.ExchangeID("binance"), ebalanceValue)
	err = fetcher.PersistSnapshot(&ebalance, bbalance, &estatuses, &bstatuses, pendings, &snapshot, timepoint)
	if err != nil {
		t.Fatalf("Cannot persist snapshot: %s", err.Error())
	}
	// check if snapshot store latest data instead of empty
	version, err := fetcher.storage.CurrentAuthDataVersion(common.GetTimepoint())
	if err != nil {
		t.Fatalf("Snapshot did not saved: %s", err.Error())
	}
	authData, err := fetcher.storage.GetAuthData(version)
	if err != nil {
		t.Fatalf("Cannot get snapshot: %s", err.Error())
	}
	exchangeBalance := authData.ExchangeBalances[common.ExchangeID("binance")]
	if exchangeBalance.AvailableBalance["KNC"] != 500 {
		t.Fatalf("Snapshot did not get the latest auth data instead")
	}

	if exchangeBalance.Error != "Connection time out" {
		t.Fatalf("Snapshot did not save exchange error")
	}
}
