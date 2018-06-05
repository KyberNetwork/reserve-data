package http

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"reflect"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/core"
	"github.com/KyberNetwork/reserve-data/data"
	"github.com/KyberNetwork/reserve-data/data/storage"
	"github.com/KyberNetwork/reserve-data/http/httputil"
	"github.com/KyberNetwork/reserve-data/metric"
	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
)

type assertFn func(t *testing.T, resp *httptest.ResponseRecorder)

type testCase struct {
	msg      string
	endpoint string
	method   string
	data     string
	assert   assertFn
}

func newAssertGetEquation(expectedData []byte) assertFn {
	return func(t *testing.T, resp *httptest.ResponseRecorder) {
		t.Helper()

		var expected metric.PWIEquationRequestV2
		if resp.Code != http.StatusOK {
			t.Fatalf("wrong return code, expected: %d, got: %d", http.StatusOK, resp.Code)
		}

		type responseBody struct {
			Success bool
			Data    metric.PWIEquationRequestV2
		}

		decoded := responseBody{}
		if aErr := json.NewDecoder(resp.Body).Decode(&decoded); aErr != nil {
			t.Fatal(aErr)
		}

		if decoded.Success != true {
			t.Errorf("wrong success status, expected: %t, got: %t", true, decoded.Success)
		}

		t.Logf("returned pending PWI equation request: %v", decoded.Data)

		if len(decoded.Data) != 2 {
			t.Fatalf("wrong number of tokens, expected: %d, got %d", 2, len(decoded.Data))
		}

		if aErr := json.Unmarshal(expectedData, &expected); aErr != nil {
			t.Fatal(aErr)
		}

		if !reflect.DeepEqual(expected, decoded.Data) {
			t.Logf("expected data doesn't match: %v, decoded: %v", expected, decoded)
		}
	}
}

func testHTTPRequest(t *testing.T, tc testCase, handler http.Handler) {
	t.Helper()

	req, tErr := http.NewRequest(tc.method, tc.endpoint, nil)
	if tErr != nil {
		t.Fatal(tErr)
	}

	if tc.data != "" {
		form := url.Values{}
		form.Add("data", tc.data)
		req.PostForm = form
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	tc.assert(t, resp)
}

func TestHTTPServerPWIEquationV2(t *testing.T) {
	const (
		storePendingPWIEquationV2Endpoint = "/v2/set-pwis-equation"
		getPendingPWIEquationV2Endpoint   = "/v2/pending-pwis-equation"
		confirmPWIEquationV2              = "/v2/confirm-pwis-equation"
		rejectPWIEquationV2               = "/v2/reject-pwis-equation"
		getPWIEquationV2                  = "/v2/pwis-equation"
		testDataV1                        = `EOS_750_500_0.25|ETH_750_500_0.25`
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

	common.RegisterInternalActiveToken(common.Token{ID: "EOS"})
	common.RegisterInternalActiveToken(common.Token{ID: "ETH"})

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

	s := HTTPServer{
		app:         data.NewReserveData(st, nil, nil, nil, nil, nil),
		core:        core.NewReserveCore(nil, st, ethereum.Address{}),
		metric:      st,
		authEnabled: false,
		r:           gin.Default()}
	s.register()

	var tests = []testCase{
		{
			msg:      "invalid post form",
			endpoint: storePendingPWIEquationV2Endpoint,
			method:   http.MethodPost,
			data:     `{"invalid_key": "invalid_value"}`,
			assert:   httputil.ExpectFailure,
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
			msg:      "setting equation v1 pending",
			endpoint: "/set-pwis-equation",
			method:   http.MethodPost,
			data:     testDataV1,
			assert:   httputil.ExpectSuccess,
		},
		{
			msg:      "confirm equation v1 pending",
			endpoint: "/confirm-pwis-equation",
			method:   http.MethodPost,
			data:     testDataV1,
			assert:   httputil.ExpectSuccess,
		},
		{
			msg:      "getting fallback v1 equation",
			endpoint: getPWIEquationV2,
			method:   http.MethodGet,
			assert:   newAssertGetEquation([]byte(testData)),
		},
		{
			msg:      "unsupported token",
			endpoint: storePendingPWIEquationV2Endpoint,
			method:   http.MethodPost,
			data:     testDataUnsupported,
			assert:   httputil.ExpectFailure,
		},
		{
			msg:      "confirm when no pending equation request exists",
			endpoint: confirmPWIEquationV2,
			method:   http.MethodPost,
			data:     testData,
			assert:   httputil.ExpectFailure,
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
			data:     testData,
			assert:   httputil.ExpectSuccess,
		},
		{
			msg:      "setting when pending exists",
			endpoint: storePendingPWIEquationV2Endpoint,
			method:   http.MethodPost,
			data:     testData,
			assert:   httputil.ExpectFailure,
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
			data:     testDataWrongConfirmation,
			assert:   httputil.ExpectFailure,
		},
		{
			msg:      "confirm with correct data",
			endpoint: confirmPWIEquationV2,
			method:   http.MethodPost,
			data:     testDataConfirmation,
			assert:   httputil.ExpectSuccess,
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
			data:     testData,
			assert:   httputil.ExpectSuccess,
		},
		{
			msg:      "reject when there is pending equation",
			endpoint: rejectPWIEquationV2,
			method:   http.MethodPost,
			data:     "some random post form or this request will be unauthenticated",
			assert:   httputil.ExpectSuccess,
		},
	}

	for _, tc := range tests {
		t.Run(tc.msg, func(t *testing.T) { testHTTPRequest(t, tc, s.r) })
	}
}

func TestHTTPServerTargetQtyV2(t *testing.T) {
	const (
		storePendingTargetQtyV2 = "/v2/settargetqty"
		getPendingTargetQtyV2   = "/v2/pendingtargetqty"
		confirmTargetQtyV2      = "/v2/confirmtargetqty"
		rejectTargetQtyV2       = "/v2/canceltargetqty"
		getTargetQtyV2          = "/v2/targetqty"
		testDataV1              = `EOS_750_500_0.25_0.324|ETH_750_500_0.25_0.342`
		testData                = `{
  "EOS": {
      "TotalTarget": 750,
      "ReserveTarget": 500,
	  "RebalanceThreshold": 0.25,
	  "TransferThreshold": 0.343
    },
  },
  "ETH": {
      "TotalTarget": 750,
      "ReserveTarget": 500,
	  "RebalanceThreshold": 0.25,
	  "TransferThreshold": 0.343,
	  "MinimumAmount" : {
		"huobi" : 10,
		"binance" : 20
	  },
	  "ExchangeRatio":  {
		"huobi": 3,
		"bianace": 4
	}
  }
}
`
		testDataWrongConfirmation = `{
  "EOS": {
      "TotalTarget": 751,
      "ReserveTarget": 500,
	  "RebalanceThreshold": 0.25,
	  "TransferThreshold": 0.343
    },
  "ETH": {
      "TotalTarget": 750,
      "ReserveTarget": 500,
	  "RebalanceThreshold": 0.25,
	  "TransferThreshold": 0.343,
	  "MinimumAmount" : {
		"huobi" : 10,
		"binance" : 20
	  },
	  "ExchangeRatio":  {
		"huobi": 3,
		"bianace": 4
	}
  }
}
`
		testDataUnsupported = `{
  "OMG": {
      "TotalTarget": 750,
      "ReserveTarget": 500,
	  "RebalanceThreshold": 0.25,
	  "TransferThreshold": 0.343
  }
}
`
	)

	common.RegisterInternalActiveToken(common.Token{ID: "EOS"})
	common.RegisterInternalActiveToken(common.Token{ID: "ETH"})

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

	s := HTTPServer{
		app:         data.NewReserveData(st, nil, nil, nil, nil, nil),
		core:        core.NewReserveCore(nil, st, ethereum.Address{}),
		metric:      st,
		authEnabled: false,
		r:           gin.Default()}
	s.register()

	var tests = []testCase{
		{
			msg:      "invalid post form",
			endpoint: storePendingTargetQtyV2,
			method:   http.MethodPost,
			data:     `{"invalid_key": "invalid_value"}`,
			assert:   httputil.ExpectFailure,
		},
		{
			msg:      "getting non exists pending equation",
			endpoint: getPendingTargetQtyV2,
			method:   http.MethodGet,
			assert:   httputil.ExpectFailure,
		},
		{
			msg:      "getting non exists equation",
			endpoint: getTargetQtyV2,
			method:   http.MethodGet,
			assert:   httputil.ExpectFailure,
		},
		// {
		// 	msg:      "setting equation v1 pending",
		// 	endpoint: "/settargetqty",
		// 	method:   http.MethodPost,
		// 	data:     testDataV1,
		// 	assert:   httputil.ExpectSuccess,
		// },
		// {
		// 	msg:      "confirm equation v1 pending",
		// 	endpoint: "/confirmtargetqty",
		// 	method:   http.MethodPost,
		// 	data:     testDataV1,
		// 	assert:   httputil.ExpectSuccess,
		// },
		// {
		// 	msg:      "getting fallback v1 equation",
		// 	endpoint: getTargetQtyV2,
		// 	method:   http.MethodGet,
		// 	assert:   newAssertGetEquation([]byte(testData)),
		// },
		{
			msg:      "unsupported token",
			endpoint: storePendingTargetQtyV2,
			method:   http.MethodPost,
			data:     testDataUnsupported,
			assert:   httputil.ExpectFailure,
		},
		{
			msg:      "confirm when no pending equation request exists",
			endpoint: confirmTargetQtyV2,
			method:   http.MethodPost,
			data:     testData,
			assert:   httputil.ExpectFailure,
		},
		{
			msg:      "reject when no pending equation request exists",
			endpoint: rejectTargetQtyV2,
			method:   http.MethodPost,
			assert:   httputil.ExpectFailure,
		},
		{
			msg:      "valid post form",
			endpoint: storePendingTargetQtyV2,
			method:   http.MethodPost,
			data:     testData,
			assert:   httputil.ExpectSuccess,
		},
		{
			msg:      "setting when pending exists",
			endpoint: storePendingTargetQtyV2,
			method:   http.MethodPost,
			data:     testData,
			assert:   httputil.ExpectFailure,
		},
		{
			msg:      "getting existing pending equation",
			endpoint: getPendingTargetQtyV2,
			method:   http.MethodGet,
			assert:   newAssertGetEquation([]byte(testData)),
		},
		{
			msg:      "confirm with wrong data",
			endpoint: confirmTargetQtyV2,
			method:   http.MethodPost,
			data:     testDataWrongConfirmation,
			assert:   httputil.ExpectFailure,
		},
		{
			msg:      "confirm with correct data",
			endpoint: confirmTargetQtyV2,
			method:   http.MethodPost,
			data:     testData,
			assert:   httputil.ExpectSuccess,
		},
		{
			msg:      "getting exists equation",
			endpoint: getTargetQtyV2,
			method:   http.MethodGet,
			assert:   newAssertGetEquation([]byte(testData)),
		},
		{
			msg:      "valid post form",
			endpoint: storePendingTargetQtyV2,
			method:   http.MethodPost,
			data:     testData,
			assert:   httputil.ExpectSuccess,
		},
		{
			msg:      "reject when there is pending equation",
			endpoint: rejectTargetQtyV2,
			method:   http.MethodPost,
			data:     "some random post form or this request will be unauthenticated",
			assert:   httputil.ExpectSuccess,
		},
	}

	for _, tc := range tests {
		t.Run(tc.msg, func(t *testing.T) { testHTTPRequest(t, tc, s.r) })
	}
}
