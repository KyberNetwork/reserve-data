package configuration

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/KyberNetwork/reserve-data/stat"
	statstorage "github.com/KyberNetwork/reserve-data/stat/storage"
)

func SetupBoltStatStorageTester(name string) (*stat.StatStorageTest, func() error, error) {
	tmpDir, err := ioutil.TempDir("", "test_stats")
	if err != nil {
		return nil, nil, err
	}
	storage, err := statstorage.NewBoltStatStorage(filepath.Join(tmpDir, name))
	if err != nil {
		return nil, nil, err
	}
	tearDownFn := func() error {
		return os.Remove(tmpDir)
	}
	return stat.NewStatStorageTest(storage), tearDownFn, nil
}

func doStatBoltTest(f func(tester *stat.StatStorageTest, t *testing.T), t *testing.T) {
	dbname := "boltstattest1.db"
	tester, tearDownFn, err := SetupBoltStatStorageTester(dbname)
	if err != nil {
		t.Fatalf("Testing stat_bolt as a stat storage: init failed (%s)", err)
	}
	defer tearDownFn()
	f(tester, t)
}

func TestBoltStatForStatStorage(t *testing.T) {
	doStatBoltTest(func(tester *stat.StatStorageTest, t *testing.T) {
		if err := tester.TestTradeStatsSummary(); err != nil {
			t.Fatalf("Testing stat_bolt as a stat storage: Test Trade Stats Summary failed (%s)", err)

		}
	}, t)
	doStatBoltTest(func(tester *stat.StatStorageTest, t *testing.T) {
		if err := tester.TestWalletAddress(); err != nil {
			t.Fatalf("Testing stat_bolt as a stat storage: Test Wallet Address  failed (%s)", err)

		}
	}, t)
	doStatBoltTest(func(tester *stat.StatStorageTest, t *testing.T) {
		if err := tester.TestWalletStats(); err != nil {
			t.Fatalf("Testing stat_bolt as a stat storage: Test Wallet Stats failed (%s)", err)

		}
	}, t)
	doStatBoltTest(func(tester *stat.StatStorageTest, t *testing.T) {
		if err := tester.TestCountryStats(); err != nil {
			t.Fatalf("Testing stat_bolt as a stat storage: Test Country Stats failed (%s)", err)

		}
	}, t)
	doStatBoltTest(func(tester *stat.StatStorageTest, t *testing.T) {
		if err := tester.TestVolumeStats(); err != nil {
			t.Fatalf("Testing stat_bolt as a stat storage: Test Volume Stats failed (%s)", err)

		}
	}, t)
	doStatBoltTest(func(tester *stat.StatStorageTest, t *testing.T) {
		if err := tester.TestBurnFee(); err != nil {
			t.Fatalf("Testing stat_bolt as a stat storage: Test Burn Fee failed (%s)", err)

		}
	}, t)
	doStatBoltTest(func(tester *stat.StatStorageTest, t *testing.T) {
		if err := tester.TestCountries(); err != nil {
			t.Fatalf("Testing stat_bolt as a stat storage: Test countries failed (%s)", err)

		}
	}, t)
	doStatBoltTest(func(tester *stat.StatStorageTest, t *testing.T) {
		if err := tester.TestFirstTradeEver(); err != nil {
			t.Fatalf("Testing stat_bolt as a stat storage: Test first trade ever failed (%s)", err)

		}
	}, t)
	doStatBoltTest(func(tester *stat.StatStorageTest, t *testing.T) {
		if err := tester.TestFirstTradeInDay(); err != nil {
			t.Fatalf("Testing stat_bolt as a stat storage: Test first trade ever failed (%s)", err)

		}
	}, t)
}
