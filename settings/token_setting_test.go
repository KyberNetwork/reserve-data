package settings_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/settings"
	settingsstorage "github.com/KyberNetwork/reserve-data/settings/storage"
	ethereum "github.com/ethereum/go-ethereum/common"
)

func newTestSetting(t *testing.T, tmpDir string) *settings.Settings {
	t.Helper()
	boltSettingStorage, err := settingsstorage.NewBoltSettingStorage(filepath.Join(tmpDir, "setting.db"))
	if err != nil {
		t.Fatal(err)
	}
	tokenSetting, err := settings.NewTokenSetting(boltSettingStorage)
	if err != nil {
		t.Fatal(err)
	}
	addressSetting := &settings.AddressSetting{}

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

func testPositiveGetInternalToken(t *testing.T, setting *settings.Settings, testToken common.Token) {
	t.Helper()
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
	// CreationTime shall be ignore since it's not from UserInput
	token.CreationTime, testToken.CreationTime = 0, 0
	if !reflect.DeepEqual(token, testToken) {
		t.Fatalf("token returned was different from the input")
	}
	token, err = setting.GetInternalTokenByID(testToken.ID)
	if err != nil {
		t.Fatalf("cannot get internal token by ID %s", err.Error())
	}
	// CreationTime shall be ignore since it's not from UserInput
	token.CreationTime, testToken.CreationTime = 0, 0
	if !reflect.DeepEqual(token, testToken) {
		t.Fatalf("token returned was different from the input")
	}
}

func testNegativeGetInternalToken(t *testing.T, setting *settings.Settings, testToken common.Token) {
	t.Helper()
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
	t.Helper()
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
	// CreationTime shall be ignore since it's not from UserInput
	token.CreationTime, testToken.CreationTime = 0, 0
	if !reflect.DeepEqual(token, testToken) {
		t.Fatalf("token returned was different from the input")
	}
	token, err = setting.GetActiveTokenByAddress(testAddress)
	if err != nil {
		t.Fatalf("cannot get active token by Address %s", err.Error())
	}
	// CreationTime shall be ignore since it's not from UserInput
	token.CreationTime, testToken.CreationTime = 0, 0
	if !reflect.DeepEqual(token, testToken) {
		t.Fatalf("token returned was different from the input")
	}
}

func testGetToken(t *testing.T, setting *settings.Settings, testToken common.Token) {
	t.Helper()
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
	// CreationTime shall be ignore since it's not from UserInput
	token.CreationTime, testToken.CreationTime = 0, 0
	if !reflect.DeepEqual(token, testToken) {
		t.Fatalf("token returned was different from the input")
	}
	token, err = setting.GetTokenByAddress(testAddress)
	if err != nil {
		t.Fatalf("cannot get token by ID %s", err.Error())
	}
	// CreationTime shall be ignore since it's not from UserInput
	token.CreationTime, testToken.CreationTime = 0, 0
	if !reflect.DeepEqual(token, testToken) {
		t.Fatalf("token returned was different from the input")
	}
}

func TestInternaTokenSetting(t *testing.T) {
	testInternalToken := common.NewToken("OMG", "omise-go", "0x1111111111111111111111111111111111111111", 18, true, true, 0)
	tmpDir, err := ioutil.TempDir("", "test_setting")
	if err != nil {
		t.Fatal(err)
	}
	setting := newTestSetting(t, tmpDir)
	defer func() {
		if rErr := os.RemoveAll(tmpDir); rErr != nil {
			t.Error(rErr)
		}
	}()
	if err := setting.UpdateToken(testInternalToken, 0); err != nil {
		t.Fatal(err)
	}
	testPositiveGetInternalToken(t, setting, testInternalToken)
	testGetToken(t, setting, testInternalToken)
	testGetActiveToken(setting, testInternalToken, t)
}

func TestExternalTokenSetting(t *testing.T) {
	testExternalToken := common.NewToken("KNC", "Kyber-coin", "0x2222222222222222222222222222222222222222", 18, true, false, 0)
	tmpDir, err := ioutil.TempDir("", "test_setting")
	if err != nil {
		t.Fatal(err)
	}
	setting := newTestSetting(t, tmpDir)
	defer func() {
		if rErr := os.RemoveAll(tmpDir); rErr != nil {
			t.Error(rErr)
		}
	}()
	if err := setting.UpdateToken(testExternalToken, 0); err != nil {
		t.Fatal(err)
	}
	testGetToken(t, setting, testExternalToken)
	testGetActiveToken(setting, testExternalToken, t)
	testNegativeGetInternalToken(t, setting, testExternalToken)
}
