package settings_test

import (
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/settings"
	settingsstorage "github.com/KyberNetwork/reserve-data/settings/storage"
	ethereum "github.com/ethereum/go-ethereum/common"
)

var (
	testInternalToken = common.NewToken("OMG", "omise-go", "0x1111111111111111111111111111111111111111", 18, true, true, 0)
	testExternalToken = common.NewToken("KNC", "Kyber-coin", "0x2222222222222222222222222222222222222222", 18, true, false, 0)
)

func getTestSetting(t *testing.T) *settings.Settings {
	tmpDir, err := ioutil.TempDir("", "test_setting.db")
	if err != nil {
		t.Fatal(err)
	}
	boltSettingStorage, err := settingsstorage.NewBoltSettingStorage(filepath.Join(tmpDir, "setting.db"))
	if err != nil {
		t.Fatal(err)
	}
	tokenSetting, err := settings.NewTokenSetting(boltSettingStorage)
	if err != nil {
		t.Fatal(err)
	}
	addressSetting, err := settings.NewAddressSetting(boltSettingStorage)
	if err != nil {
		t.Fatal(err)
	}
	exchangeSetting, err := settings.NewExchangeSetting(boltSettingStorage)
	if err != nil {
		t.Fatal(err)
	}
	setting, err := settings.NewSetting(tokenSetting, addressSetting, exchangeSetting)
	if err != nil {
		t.Fatal(err)
	}
	return setting
}

func testPositiveGetInternalToken(setting *settings.Settings, testToken common.Token, t *testing.T) {
	tokens, err := setting.GetInternalTokens()
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 1 {
		t.Fatalf("expect 1 token, got %d from getInternalToken", len(tokens))
	}
	testAddress := ethereum.HexToAddress(testToken.Address)
	token, err := setting.GetInternalTokenByAddress(testAddress)
	if err != nil {
		t.Fatalf("cannot get internal token by address %s", err.Error())
	}
	if !reflect.DeepEqual(token, testToken) {
		t.Fatalf("token returned was different from the input")
	}
	token, err = setting.GetInternalTokenByID(testToken.ID)
	if err != nil {
		t.Fatalf("cannot get internal token by ID %s", err.Error())
	}
	if !reflect.DeepEqual(token, testToken) {
		t.Fatalf("token returned was different from the input")
	}
}

func testNegativeGetInternalToken(setting *settings.Settings, testToken common.Token, t *testing.T) {
	tokens, err := setting.GetInternalTokens()
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 0 {
		t.Fatalf("expect 0 token, got %d from getInternalToken", len(tokens))
	}
	testAddress := ethereum.HexToAddress(testToken.Address)
	_, err = setting.GetInternalTokenByAddress(testAddress)
	if err != settings.ErrTokenNotFound {
		t.Fatal("expect there is no token, but the result was different")
	}
	_, err = setting.GetInternalTokenByID(testToken.ID)
	if err != settings.ErrTokenNotFound {
		t.Fatal("expect there is no token, but the result was different")
	}
}

func testGetActiveToken(setting *settings.Settings, testToken common.Token, t *testing.T) {
	tokens, err := setting.GetActiveTokens()
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 1 {
		t.Fatalf("expect 1 token, got %d from getAllTokens", len(tokens))
	}

	testAddress := ethereum.HexToAddress(testToken.Address)
	token, err := setting.GetActiveTokenByID(testToken.ID)
	if err != nil {
		t.Fatalf("cannot get active token by ID %s", err.Error())
	}
	if !reflect.DeepEqual(token, testToken) {
		t.Fatalf("token returned was different from the input")
	}
	token, err = setting.GetActiveTokenByAddress(testAddress)
	if err != nil {
		t.Fatalf("cannot get active token by Address %s", err.Error())
	}
	if !reflect.DeepEqual(token, testToken) {
		t.Fatalf("token returned was different from the input")
	}
}

func testGetToken(setting *settings.Settings, testToken common.Token, t *testing.T) {
	tokens, err := setting.GetAllTokens()
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 1 {
		t.Fatalf("expect 1 token, got %d from getAllTokens", len(tokens))
	}
	testAddress := ethereum.HexToAddress(testToken.Address)
	token, err := setting.GetTokenByID(testToken.ID)
	if err != nil {
		t.Fatalf("cannot  get token by ID %s", err.Error())
	}
	if !reflect.DeepEqual(token, testToken) {
		t.Fatalf("token returned was different from the input")
	}
	token, err = setting.GetTokenByAddress(testAddress)
	if err != nil {
		t.Fatalf("cannot get token by ID %s", err.Error())
	}
	if !reflect.DeepEqual(token, testToken) {
		t.Fatalf("token returned was different from the input")
	}
}

func TestInternaTokenSetting(t *testing.T) {
	setting := getTestSetting(t)
	if err := setting.UpdateToken(testInternalToken); err != nil {
		t.Fatal(err)
	}
	testPositiveGetInternalToken(setting, testInternalToken, t)
	testGetToken(setting, testInternalToken, t)
	testGetActiveToken(setting, testInternalToken, t)
}

func TestExternalTokenSetting(t *testing.T) {
	setting := getTestSetting(t)
	if err := setting.UpdateToken(testExternalToken); err != nil {
		t.Fatal(err)
	}
	testGetToken(setting, testExternalToken, t)
	testGetActiveToken(setting, testExternalToken, t)
	testNegativeGetInternalToken(setting, testExternalToken, t)
}
