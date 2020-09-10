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

func TestHTTPServerPWIEquationV2(t *testing.T) {
	const (
		storePendingPWIEquationV2Endpoint = "/v2/set-pwis-equation"
		getPendingPWIEquationV2Endpoint   = "/v2/pending-pwis-equation"
		confirmPWIEquationV2              = "/v2/confirm-pwis-equation"
		rejectPWIEquationV2               = "/v2/reject-pwis-equation"
		getPWIEquationV2                  = "/v2/pwis-equation"
		testData                          = `{
  "EOS": {
    "bid": {
      "a": 750,
      "B": 500,
      "c": 0.25,
      "min_min_spread": 0,
      "price_multiply_factor": 0
    },
    "ask": {
      "a": 750,
      "B": 500,
      "c": 0.25,
      "min_min_spread": 0,
      "price_multiply_factor": 0
    }
  },
  "ETH": {
    "bid": {
      "a": 750,
      "B": 500,
      "c": 0.25,
      "min_min_spread": 0,
      "price_multiply_factor": 0
    },
    "ask": {
      "a": 750,
      "B": 500,
      "c": 0.25,
      "min_min_spread": 0,
      "price_multiply_factor": 0
    }
  }
}
`
		testDataWrongConfirmation = `{
  "EOS": {
    "bid": {
      "a": 751,
      "B": 500,
      "c": 0,
      "min_min_spread": 0,
      "price_multiply_factor": 0
    },
    "ask": {
      "a": 800,
      "B": 600,
      "c": 0,
      "min_min_spread": 0,
      "price_multiply_factor": 0
    }
  },
  "ETH": {
    "bid": {
      "a": 750,
      "B": 500,
      "c": 0,
      "min_min_spread": 0,
      "price_multiply_factor": 0
    },
    "ask": {
      "a": 800,
      "B": 600,
      "c": 0,
      "min_min_spread": 0,
      "price_multiply_factor": 0
    }
  }
}
`
		testDataUnsupported = `{
  "OMG": {
    "bid": {
      "a": 750,
      "B": 500,
      "c": 0,
      "min_min_spread": 0,
      "price_multiply_factor": 0
    },
    "ask": {
      "a": 800,
      "B": 600,
      "c": 0,
      "min_min_spread": 0,
      "price_multiply_factor": 0
    }
  }
`
		testDataConfirmation = `{
	"ETH": {
		"bid": {
		  "a": 750,
		  "B": 500,
		  "c": 0.25,
		  "min_min_spread": 0,
		  "price_multiply_factor": 0
		},
		"ask": {
		  "a": 750,
		  "B": 500,
		  "c": 0.25,
		  "min_min_spread": 0,
		  "price_multiply_factor": 0
		}
	  },
	"EOS": {
		"bid": {
		"a": 750,
		"B": 500,
		"c": 0.25,
		"min_min_spread": 0,
		"price_multiply_factor": 0
		},
		"ask": {
		"a": 750,
		"B": 500,
		"c": 0.25,
		"min_min_spread": 0,
		"price_multiply_factor": 0
		}
	}
}
	`
	)

	tmpDir, err := ioutil.TempDir("", "test_pwi_equation_v2")
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
		ID:       "EOS",
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
		setting:     setting}
	s.register()

	var tests = []testCase{
		{
			msg:      "invalid post form",
			endpoint: storePendingPWIEquationV2Endpoint,
			method:   http.MethodPost,
			data: map[string]string{
				"invalid_key": "invalid_value",
			},
			assert: httputil.ExpectFailure,
		},
		{
			msg:      "getting non exists pending equation",
			endpoint: getPendingPWIEquationV2Endpoint,
			method:   http.MethodGet,
			assert:   httputil.ExpectFailure,
		},
		{
			msg:      "getting non exists equation",
			endpoint: getPWIEquationV2,
			method:   http.MethodGet,
			assert:   httputil.ExpectFailure,
		},
		{
			msg:      "unsupported token",
			endpoint: storePendingPWIEquationV2Endpoint,
			method:   http.MethodPost,
			data: map[string]string{
				"data": testDataUnsupported,
			},
			assert: httputil.ExpectFailure,
		},
		{
			msg:      "confirm when no pending equation request exists",
			endpoint: confirmPWIEquationV2,
			method:   http.MethodPost,
			data: map[string]string{
				"data": testData,
			},
			assert: httputil.ExpectFailure,
		},
		{
			msg:      "reject when no pending equation request exists",
			endpoint: rejectPWIEquationV2,
			method:   http.MethodPost,
			assert:   httputil.ExpectFailure,
		},
		{
			msg:      "valid post form",
			endpoint: storePendingPWIEquationV2Endpoint,
			method:   http.MethodPost,
			data: map[string]string{
				"data": testData,
			},
			assert: httputil.ExpectSuccess,
		},
		{
			msg:      "setting when pending exists",
			endpoint: storePendingPWIEquationV2Endpoint,
			method:   http.MethodPost,
			data: map[string]string{
				"data": testData,
			},
			assert: httputil.ExpectFailure,
		},
		{
			msg:      "getting existing pending equation",
			endpoint: getPendingPWIEquationV2Endpoint,
			method:   http.MethodGet,
			assert:   newAssertGetEquation([]byte(testData)),
		},
		{
			msg:      "confirm with wrong data",
			endpoint: confirmPWIEquationV2,
			method:   http.MethodPost,
			data: map[string]string{
				"data": testDataWrongConfirmation,
			},
			assert: httputil.ExpectFailure,
		},
		{
			msg:      "confirm with correct data",
			endpoint: confirmPWIEquationV2,
			method:   http.MethodPost,
			data: map[string]string{
				"data": testDataConfirmation,
			},
			assert: httputil.ExpectSuccess,
		},
		{
			msg:      "getting exists equation",
			endpoint: getPWIEquationV2,
			method:   http.MethodGet,
			assert:   newAssertGetEquation([]byte(testData)),
		},
		{
			msg:      "valid post form",
			endpoint: storePendingPWIEquationV2Endpoint,
			method:   http.MethodPost,
			data: map[string]string{
				"data": testData,
			},
			assert: httputil.ExpectSuccess,
		},
		{
			msg:      "reject when there is pending equation",
			endpoint: rejectPWIEquationV2,
			method:   http.MethodPost,
			data: map[string]string{
				"data": "some random post form or this request will be unauthenticated",
			},
			assert: httputil.ExpectSuccess,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.msg, func(t *testing.T) { testHTTPRequest(t, tc, s.r) })
	}
}
