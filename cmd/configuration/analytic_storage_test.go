package configuration

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/KyberNetwork/reserve-data/stat"

	statstorage "github.com/KyberNetwork/reserve-data/stat/storage"
)

func SetupBoltAnalyticStorageTester(name string) (*stat.AnalyticStorageTest, func() error, error) {
	tmpDir, err := ioutil.TempDir("", "test_analyticStorage")
	if err != nil {
		return nil, nil, err
	}
	storage, err := statstorage.NewBoltAnalyticStorage(filepath.Join(tmpDir, name))
	if err != nil {
		return nil, nil, err
	}
	tearDownFn := func() error {
		return os.Remove(tmpDir)
	}
	return stat.NewAnalyticStorageTest(storage), tearDownFn, nil
}

func doAnalyticStorageTest(f func(tester *stat.AnalyticStorageTest, t *testing.T), t *testing.T) {
	dbname := "test1.db"
	tester, tearDownFn, err := SetupBoltAnalyticStorageTester(dbname)
	if err != nil {
		t.Fatalf("Testing bolt as a stat storage: init failed(%s)", err)
	}
	defer tearDownFn()
	f(tester, t)
}

func TestBoltAnalyticForStatStorage(t *testing.T) {
	doAnalyticStorageTest(func(tester *stat.AnalyticStorageTest, t *testing.T) {
		if err := tester.TestPriceAnalyticData(); err != nil {
			t.Fatalf("Testing bolt as a stat storage: test analytic storage write/read failed: %s", err)
		}
	}, t)

}
