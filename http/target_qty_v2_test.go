package http

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/core"
	"github.com/KyberNetwork/reserve-data/data"
	"github.com/KyberNetwork/reserve-data/data/storage"
	"github.com/KyberNetwork/reserve-data/http/httputil"
	"github.com/KyberNetwork/reserve-data/settings"
	settingsstorage "github.com/KyberNetwork/reserve-data/settings/storage"
)

//TestHTTPServerTargetQtyV2 check if api v2 work correctly
func TestHTTPServerTargetQtyV2(t *testing.T) {
	const (
		storePendingTargetQtyV2 = "/v2/settargetqty"
		getPendingTargetQtyV2   = "/v2/pendingtargetqty"
		confirmTargetQtyV2      = "/v2/confirmtargetqty"
		rejectTargetQtyV2       = "/v2/canceltargetqty"
		getTargetQtyV2          = "/v2/targetqty"
		testData                = `{
  "OMG": {
	  "set_target": {
	      "total_target": 750,
	      "reserve_target": 500,
		  "rebalance_threshold": 0.25,
		  "transfer_threshold": 0.343
	  }
  },
  "ETH": {
	  "set_target": {
	      "total_target": 750,
	      "reserve_target": 500,
		  "rebalance_target": 0.25,
		  "transfer_threshold": 0.343
	  },
	  "recommend_balance" : {
		"huobi" : 10,
		"binance" : 20
	  }, "exchange_ratio":  {
		"huobi": 3,
		"bianace": 4
	}
  }
}
`

		testDataTokenNotSupported = `{
  "SNT": {
	  "set_target": {
 	     "total_target": 750,
 	     "reserve_target": 500,
		  "rebalance_threshold": 0.25,
		  "transfer_threshold": 0.343
	  }
  },
  "ETH": {
	  "set_target": {
	      "total_target": 750,
	      "reserve_target": 500,
		  "rebalance_target": 0.25,
		  "transfer_threshold": 0.343
	  },
	  "recommend_balance" : {
		"huobi" : 10,
		"binance" : 20
	  }, "exchange_ratio":  {
		"huobi": 3,
		"bianace": 4
	}
  }
		}`

		testDataWrongConfirmation = `{
  "OMG": {
	  "set_target": {
	      "total_target": 751,
	      "reserve_target": 500,
		  "rebalance_threshold": 0.25,
		  "transfer_threshold": 0.343
	  }
    },
  "ETH": {
	  "set_target": {
	      "total_target": 750,
	      "reserve_target": 500,
		  "rebalance_target": 0.25,
		  "transfer_threshold": 0.343
	  },
	  "recommend_balance" : {
		"huobi" : 10,
		"binance" : 20
	  },
	  "exchange_ratio":  {
		"huobi": 3,
		"bianace": 4
	}
  }
}
`
	)

	tmpDir, err := ioutil.TempDir("", "test_target_qty_v2")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if rErr := os.RemoveAll(tmpDir); rErr != nil {
			t.Error(rErr)
		}
	}()

	st, err := storage.NewBoltStorage(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
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
	err = setting.UpdateToken(common.Token{
		ID:       "OMG",
		Address:  "xxx",
		Internal: true,
		Active:   true,
	}, 0)
	if err != nil {
		t.Fatal(err)
	}
	err = setting.UpdateToken(common.Token{
		ID:       "ETH",
		Address:  "xxx",
		Internal: true,
		Active:   true,
	}, 0)
	if err != nil {
		t.Fatal(err)
	}
	s := Server{
		app:         data.NewReserveData(st, nil, nil, nil, nil, nil, setting),
		core:        core.NewReserveCore(nil, st, setting, nil),
		metric:      st,
		authEnabled: false,
		r:           gin.Default(),
		setting:     setting,
	}
	s.register()

	var tests = []testCase{
		{
			msg:      "getting non exists pending target quantity",
			endpoint: getPendingTargetQtyV2,
			method:   http.MethodGet,
			assert:   httputil.ExpectFailure,
		},
		{
			msg:      "getting non exists target",
			endpoint: getTargetQtyV2,
			method:   http.MethodGet,
			assert:   httputil.ExpectFailure,
		},
		{
			msg:      "confirm when no pending target quantity request exists",
			endpoint: confirmTargetQtyV2,
			method:   http.MethodPost,
			data: map[string]string{
				"value": testData,
			},
			assert: httputil.ExpectFailure,
		},
		{
			msg:      "reject when no pending target quantity request exists",
			endpoint: rejectTargetQtyV2,
			method:   http.MethodPost,
			assert:   httputil.ExpectFailure,
		},
		{
			msg:      "token not supported",
			endpoint: storePendingTargetQtyV2,
			method:   http.MethodPost,
			data: map[string]string{
				"value": testDataTokenNotSupported,
			},
			assert: httputil.ExpectFailure,
		},
		{
			msg:      "valid post form",
			endpoint: storePendingTargetQtyV2,
			method:   http.MethodPost,
			data: map[string]string{
				"value": testData,
			},
			assert: httputil.ExpectSuccess,
		},
		{
			msg:      "setting when pending exists",
			endpoint: storePendingTargetQtyV2,
			method:   http.MethodPost,
			data: map[string]string{
				"value": testData,
			},
			assert: httputil.ExpectFailure,
		},
		{
			msg:      "confirm with wrong data",
			endpoint: confirmTargetQtyV2,
			method:   http.MethodPost,
			data: map[string]string{
				"value": testDataWrongConfirmation,
			},
			assert: httputil.ExpectFailure,
		},
		{
			msg:      "confirm with correct data",
			endpoint: confirmTargetQtyV2,
			method:   http.MethodPost,
			data: map[string]string{
				"value": testData,
			},
			assert: httputil.ExpectSuccess,
		},
		{
			msg:      "post a valid form to test reject",
			endpoint: storePendingTargetQtyV2,
			method:   http.MethodPost,
			data: map[string]string{
				"value": testData,
			},
			assert: httputil.ExpectSuccess,
		},
		{
			msg:      "reject when there is pending target quantity",
			endpoint: rejectTargetQtyV2,
			method:   http.MethodPost,
			data: map[string]string{
				"value": "some random post form or this request will be unauthenticated",
			},
			assert: httputil.ExpectSuccess,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.msg, func(t *testing.T) { testHTTPRequest(t, tc, s.r) })
	}
}
