package http

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/KyberNetwork/reserve-data/core"
	"github.com/KyberNetwork/reserve-data/data"
	"github.com/KyberNetwork/reserve-data/data/storage"
	"github.com/KyberNetwork/reserve-data/http/httputil"
	"github.com/KyberNetwork/reserve-data/settings"
	settingsstorage "github.com/KyberNetwork/reserve-data/settings/storage"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	feeRequest = `{
		"Trading": {
			"maker": 0.001,
			"taker": 0.001
		},
		"Funding": {
			"Withdraw": {
			"ZEC": 0.005,
			"ZIL": 100,
			"ZRX": 5.8
			},
			"Deposit": {
			"ZEC": 0,
			"ZIL": 0,
			"ZRX": 2
			}
		}
  	}`
	minDepositRequest = `{
		"POWR": 0.1,
		"MANA": 0.2	 
  	}`
	depositAddressRequest = `{
		"POWR": "0x778599Dd7893C8166D313F0F9B5F6cbF7536c293",
		"MANA": "0x1233542DSAC333FCCc6565463525F6cbF7536c29"
	}`
	exchangeInfoRequest = `{
		"LINK-ETH": {
		  "precision": {
			"amount": 0,
			"price": 8
		  },
		  "amount_limit": {
			"min": 1,
			"max": 90000000
		  },
		  "price_limit": {
			"min": 1e-8,
			"max": 120000
		  },
		  "min_notional": 0.01
		}
	}`
)

func TestHTTPUpdateExchange(t *testing.T) {
	const (
		updateExchangeFeeEndpoint    = "/setting/update-exchange-fee"
		updateMinDepositEndpoint     = "/setting/update-exchange-mindeposit"
		updateDepositAddressEndpoint = "/setting/update-deposit-address"
		updateExchangeInfoEndpoint   = "/setting/update-exchange-info"
	)
	tmpDir, err := ioutil.TempDir("", "test_exchange_APIS")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if rErr := os.RemoveAll(tmpDir); rErr != nil {
			t.Error(rErr)
		}
	}()

	boltSettingStorage, err := settingsstorage.NewBoltSettingStorage(filepath.Join(tmpDir, "setting.db"))
	if err != nil {
		log.Fatal(err)
	}
	tokenSetting, err := settings.NewTokenSetting(boltSettingStorage)
	if err != nil {
		log.Fatal(err)
	}
	addressSetting := &settings.AddressSetting{}

	exchangeSetting, err := settings.NewExchangeSetting(boltSettingStorage)
	if err != nil {
		log.Fatal(err)
	}

	setting, err := settings.NewSetting(tokenSetting, addressSetting, exchangeSetting)
	if err != nil {
		log.Fatal(err)
	}

	testStorage, err := storage.NewBoltStorage(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}

	testServer := Server{
		app:         data.NewReserveData(nil, nil, nil, nil, nil, nil, setting),
		core:        core.NewReserveCore(nil, nil, setting, nil),
		metric:      testStorage,
		authEnabled: false,
		r:           gin.Default(),
		blockchain:  testHTTPBlockchain{},
		setting:     setting,
		l:           zap.S(),
	}
	testServer.register()

	var tests = []testCase{
		//invalid post formats
		{
			msg:      "invalid post form in UpdateFee",
			endpoint: updateExchangeFeeEndpoint,
			method:   http.MethodPost,
			data: map[string]string{
				"invalid_key": "invalid_val",
			},
			assert: httputil.ExpectFailure,
		},
		{
			msg:      "invalid post form in UpdateMinDeposit",
			endpoint: updateMinDepositEndpoint,
			method:   http.MethodPost,
			data: map[string]string{
				"invalid_key": "invalid_val",
			},
			assert: httputil.ExpectFailure,
		},
		{
			msg:      "invalid post form in UpdateDepositAddress",
			endpoint: updateDepositAddressEndpoint,
			method:   http.MethodPost,
			data: map[string]string{
				"invalid_key": "invalid_val",
			},
			assert: httputil.ExpectFailure,
		},
		{
			msg:      "invalid post form in UpdateExchangeInfo",
			endpoint: updateExchangeInfoEndpoint,
			method:   http.MethodPost,
			data: map[string]string{
				"invalid_key": "invalid_val",
			},
			assert: httputil.ExpectFailure,
		},
		//Exchangefee test.
		{
			msg:      "Update Exchange Fee on an unsupported exchange",
			endpoint: updateExchangeFeeEndpoint,
			method:   http.MethodPost,
			data: map[string]string{
				"name": "ex",
				"data": feeRequest,
			},
			assert: httputil.ExpectFailure,
		},
		{
			msg:      "Update Exchange Fee on a supported exchange with wrong data format",
			endpoint: updateExchangeFeeEndpoint,
			method:   http.MethodPost,
			data: map[string]string{
				"name": "binance",
				"data": minDepositRequest,
			},
			assert: httputil.ExpectFailure,
		},
		{
			msg:      "Update Exchange Fee on a supported exchange without record in setting DB",
			endpoint: updateExchangeFeeEndpoint,
			method:   http.MethodPost,
			data: map[string]string{
				"name": "binance",
				"data": feeRequest,
			},
			assert: httputil.ExpectSuccess,
		},
		{
			msg:      "Update Exchange Fee on a supported exchange with record in setting DB",
			endpoint: updateExchangeFeeEndpoint,
			method:   http.MethodPost,
			data: map[string]string{
				"name": "binance",
				"data": feeRequest,
			},
			assert: httputil.ExpectSuccess,
		},
		//ExchangeMinDeposit test
		{
			msg:      "Update Exchange Min Deposit on an unsupported exchange",
			endpoint: updateMinDepositEndpoint,
			method:   http.MethodPost,
			data: map[string]string{
				"name": "ex",
				"data": minDepositRequest,
			},
			assert: httputil.ExpectFailure,
		},
		{
			msg:      "Update Exchange Min Deposit on a supported exchange with wrong data format",
			endpoint: updateMinDepositEndpoint,
			method:   http.MethodPost,
			data: map[string]string{
				"name": "binance",
				"data": feeRequest,
			},
			assert: httputil.ExpectFailure,
		},
		{
			msg:      "Update Exchange Min Deposit on a supported exchange without record in setting DB",
			endpoint: updateMinDepositEndpoint,
			method:   http.MethodPost,
			data: map[string]string{
				"name": "binance",
				"data": minDepositRequest,
			},
			assert: httputil.ExpectSuccess,
		},
		{
			msg:      "Update Exchange Min Deposit on a supported exchange with record in setting DB",
			endpoint: updateMinDepositEndpoint,
			method:   http.MethodPost,
			data: map[string]string{
				"name": "binance",
				"data": minDepositRequest,
			},
			assert: httputil.ExpectSuccess,
		},
		//ExchangeDepositAddress test
		{
			msg:      "Update Exchange Deposit Address on an unsupported exchange",
			endpoint: updateDepositAddressEndpoint,
			method:   http.MethodPost,
			data: map[string]string{
				"name": "ex",
				"data": depositAddressRequest,
			},
			assert: httputil.ExpectFailure,
		},
		{
			msg:      "Update Exchange Deposit Address on a supported exchange with wrong data format",
			endpoint: updateDepositAddressEndpoint,
			method:   http.MethodPost,
			data: map[string]string{
				"name": "binance",
				"data": feeRequest,
			},
			assert: httputil.ExpectFailure,
		},
		{
			msg:      "Update Exchange Deposit Address on a supported exchange without record in setting DB",
			endpoint: updateDepositAddressEndpoint,
			method:   http.MethodPost,
			data: map[string]string{
				"name": "binance",
				"data": depositAddressRequest,
			},
			assert: httputil.ExpectSuccess,
		},
		{
			msg:      "Update Exchange Deposit Address on a supported exchange with record in setting DB",
			endpoint: updateDepositAddressEndpoint,
			method:   http.MethodPost,
			data: map[string]string{
				"name": "binance",
				"data": depositAddressRequest,
			},
			assert: httputil.ExpectSuccess,
		},
		//ExchangeInfo test
		{
			msg:      "Update Exchange Info  on an unsupported exchange",
			endpoint: updateExchangeInfoEndpoint,
			method:   http.MethodPost,
			data: map[string]string{
				"name": "ex",
				"data": exchangeInfoRequest,
			},
			assert: httputil.ExpectFailure,
		},
		{
			msg:      "Update Exchange Info  on a supported exchange with wrong data format",
			endpoint: updateExchangeInfoEndpoint,
			method:   http.MethodPost,
			data: map[string]string{
				"name": "binance",
				"data": feeRequest,
			},
			assert: httputil.ExpectFailure,
		},
		{
			msg:      "Update Exchange Info  on a supported exchange without record in setting DB",
			endpoint: updateExchangeInfoEndpoint,
			method:   http.MethodPost,
			data: map[string]string{
				"name": "binance",
				"data": exchangeInfoRequest,
			},
			assert: httputil.ExpectSuccess,
		},
		{
			msg:      "Update Exchange Info  on a supported exchange with record in setting DB",
			endpoint: updateExchangeInfoEndpoint,
			method:   http.MethodPost,
			data: map[string]string{
				"name": "binance",
				"data": exchangeInfoRequest,
			},
			assert: httputil.ExpectSuccess,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.msg, func(t *testing.T) { testHTTPRequest(t, tc, testServer.r) })
	}

}
