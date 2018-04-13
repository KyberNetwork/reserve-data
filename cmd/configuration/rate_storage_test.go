package configuration

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/KyberNetwork/reserve-data/stat"
	statstorage "github.com/KyberNetwork/reserve-data/stat/storage"
)

func SetUpBoltRateStorageTester(name string) (*stat.RateStorageTest, func() error, error) {
	tmpDir, err := ioutil.TempDir("", "test_stats")
	if err != nil {
		return nil, nil, err
	}
	storage, err := statstorage.NewBoltRateStorage(filepath.Join(tmpDir, name))
	if err != nil {
		return nil, nil, err
	}
	tearDownFn := func() error {
		return os.Remove(tmpDir)
	}
	return stat.NewRateStorageTest(storage), tearDownFn, nil
}

func doBoltRateTest(f func(tester *stat.RateStorageTest, t *testing.T), t *testing.T) {
	dbname := "boltratestoragetest.db"
	tester, tearDownFn, err := SetUpBoltRateStorageTester(dbname)
	if err != nil {
		t.Fatalf("Testing stat_bolt as a stat storage: init fialed (%s)", err)
	}
	defer tearDownFn()
	f(tester, t)
}

func TestBoltRateForStatStorage(t *testing.T) {
	doBoltRateTest(func(tester *stat.RateStorageTest, t *testing.T) {
		if err := tester.TestReserveRates(); err != nil {
			t.Fatalf("Testing stat_bolt as a stat storage: Test Cat Log failed (%s)", err)
		}
	}, t)

}
