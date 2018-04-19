package configuration

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/KyberNetwork/reserve-data/stat"
	statstorage "github.com/KyberNetwork/reserve-data/stat/storage"
)

func SetupBoltLogStorageTester(name string) (*stat.LogStorageTest, func() error, error) {
	tmpDir, err := ioutil.TempDir("", "test_stats")
	if err != nil {
		return nil, nil, err
	}
	storage, err := statstorage.NewBoltLogStorage(filepath.Join(tmpDir, name))
	if err != nil {
		return nil, nil, err
	}
	tearDownFn := func() error {
		return os.Remove(tmpDir)
	}
	return stat.NewLogStorageTest(storage), tearDownFn, nil
}

func doBoltLogTest(f func(tester *stat.LogStorageTest, t *testing.T), t *testing.T) {
	dbname := "boltlogstoragetest.db"
	tester, tearDownFn, err := SetupBoltLogStorageTester(dbname)
	if err != nil {
		t.Fatalf("Testing stat_bolt as a stat storage: init fialed (%s)", err)
	}
	defer tearDownFn()
	f(tester, t)
}

func TestBoltLogForStatStorage(t *testing.T) {
	doBoltLogTest(func(tester *stat.LogStorageTest, t *testing.T) {
		if err := tester.TestCatLog(); err != nil {
			t.Fatalf("Testing stat_bolt as a stat storage: Test Cat Log failed (%s)", err)
		}
	}, t)
	doBoltLogTest(func(tester *stat.LogStorageTest, t *testing.T) {
		if err := tester.TestTradeLog(); err != nil {
			t.Fatalf("Testing stat_bolt as a stat storage: Test Trade Log failed (%s)", err)
		}
	}, t)
	doBoltLogTest(func(tester *stat.LogStorageTest, t *testing.T) {
		if err := tester.TestUtil(); err != nil {
			t.Fatalf("Testing stat_bolt as a stat storage: Test Trade Log failed (%s)", err)
		}
	}, t)
}
