package configuration

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/KyberNetwork/reserve-data/stat"
	statstorage "github.com/KyberNetwork/reserve-data/stat/storage"
)

// SetupBoltUserStorageTester returns a UserStorageTest instance
// and its tearDown function to cleanup after test.
func SetupBoltUserStorageTester(name string) (*stat.UserStorageTest, func() error, error) {
	tmpDir, err := ioutil.TempDir("", "test_fetcher")
	if err != nil {
		return nil, nil, err
	}
	storage, err := statstorage.NewBoltUserStorage(filepath.Join(tmpDir, name))
	if err != nil {
		return nil, nil, err
	}
	tearDownFn := func() error {
		return os.Remove(tmpDir)
	}
	return stat.NewUserStorageTest(storage), tearDownFn, nil
}

func doOneTest(f func(tester *stat.UserStorageTest, t *testing.T), t *testing.T) {
	dbname := "test1.db"
	tester, tearDownFn, err := SetupBoltUserStorageTester(dbname)
	if err != nil {
		t.Fatalf("Testing bolt as a stat storage: init failed(%s)", err)
	}
	defer tearDownFn()
	f(tester, t)
}

func TestBoltUserForStatStorage(t *testing.T) {
	doOneTest(func(tester *stat.UserStorageTest, t *testing.T) {
		if err := tester.TestUpdateAddressCategory(); err != nil {
			t.Fatalf("Testing bolt as a stat storage: test update address category failed(%s)", err)
		}
	}, t)
	doOneTest(func(tester *stat.UserStorageTest, t *testing.T) {
		if err := tester.TestUpdateUserAddressesThenUpdateAddressCategory(); err != nil {
			t.Fatalf("Testing bolt as a stat storage: test update user addresses and then update address category failed(%s)", err)
		}
	}, t)
	doOneTest(func(tester *stat.UserStorageTest, t *testing.T) {
		if err := tester.TestUpdateAddressCategoryThenUpdateUserAddresses(); err != nil {
			t.Fatalf("Testing bolt as a stat storage: test update address category and then update user addresses failed(%s)", err)
		}
	}, t)
}
