package http

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/core"
	"github.com/KyberNetwork/reserve-data/data"
	"github.com/KyberNetwork/reserve-data/data/storage"
	"github.com/KyberNetwork/reserve-data/http/httputil"
	"github.com/KyberNetwork/reserve-data/settings"
	settingstorage "github.com/KyberNetwork/reserve-data/settings/storage"

	"github.com/gin-gonic/gin"
)

func TestHTTPServerRebalanceQuadratic(t *testing.T) {
	const (
		storePendingRebalanceQuadratic = "/set-rebalance-quadratic"
		getPendingRebalanceQuadratic   = "/pending-rebalance-quadratic"
		confirmReblanceQuadratic       = "/confirm-rebalance-quadratic"
		rejectReblanceQuadratic        = "/reject-rebalance-quadratic"
		getRebalanceQuadratic          = "/rebalance-quadratic"
		testData                       = `{
			"KNC": {
				"rebalance_quadratic": {
					"a": 0.7,
					"b": 1.2,
					"c": 1.1
				}
			},
			"ETH": {
				"rebalance_quadratic": {
					"a": 0,
					"b": 1.1,
					"c": 2.2
				}
			}
		}`
		testWrongDataConfirmation = `{
			"KNC": {
				"rebalance_quadratic": {
					"a": 0.8,
					"b": 1.2,
					"c": 1.1
				}
			},
			"ETH": {
				"rebalance_quadratic": {
					"a": 0,
					"b": 1.1,
					"c": 2.2
				}
			}	
		}`
		testDataUnsupported = `{
			"OMG": {
				"rebalance_quadratic": {
					"a": 0.8,
					"b": 1.2,
					"c": 1.1
				}
			},
			"ETH": {
				"rebalance_quadratic": {
					"a": 0,
					"b": 1.1,
					"c": 2.2
				}
			}	
		}`
	)

	tmpDir, err := ioutil.TempDir("", "test_rebalance_quadratic")
	if err != nil {
		t.Fatal(err)
	}
	boltSettingStorage, err := settingstorage.NewBoltSettingStorage(filepath.Join(tmpDir, "setting.db"))
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
	defer func() {
		if rErr := os.RemoveAll(tmpDir); rErr != nil {
			t.Error(rErr)
		}
	}()

	if uErr := setting.UpdateToken(common.NewToken("KNC", "KyberNetwork", "xxx", 18, true, true, 0)); err != nil {
		t.Fatal(uErr)
	}
	if uErr := setting.UpdateToken(common.NewToken("ETH", "Etherium", "xxx", 18, true, true, 0)); err != nil {
		t.Error(uErr)
	}
	rqStorage, err := storage.NewBoltStorage(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}

	s := HTTPServer{
		app:         data.NewReserveData(rqStorage, nil, nil, nil, nil, nil, setting),
		core:        core.NewReserveCore(nil, rqStorage, setting),
		metric:      rqStorage,
		authEnabled: false,
		r:           gin.Default(),
		setting:     setting}
	s.register()

	var tests = []testCase{
		{
			msg:      "invalid post form",
			endpoint: storePendingRebalanceQuadratic,
			method:   http.MethodPost,
			data: map[string]string{
				"invalid_key": "invalid_value",
			},
			assert: httputil.ExpectFailure,
		},
		{
			msg:      "getting non exists pending rebalance quadratic",
			endpoint: getPendingRebalanceQuadratic,
			method:   http.MethodGet,
			assert:   httputil.ExpectFailure,
		},
		{
			msg:      "getting non exists equation",
			endpoint: getRebalanceQuadratic,
			method:   http.MethodGet,
			assert:   httputil.ExpectFailure,
		},
		{
			msg:      "unsupported token",
			endpoint: storePendingRebalanceQuadratic,
			method:   http.MethodPost,
			data: map[string]string{
				"value": testDataUnsupported,
			},
			assert: httputil.ExpectFailure,
		},
		{
			msg:      "confirm when no pending rebalance quadratic equation request exists",
			endpoint: confirmReblanceQuadratic,
			method:   http.MethodPost,
			data: map[string]string{
				"value": testData,
			},
			assert: httputil.ExpectFailure,
		},
		{
			msg:      "reject when no pending rebalance quadratic equation request exists",
			endpoint: rejectReblanceQuadratic,
			method:   http.MethodPost,
			assert:   httputil.ExpectFailure,
		},
		{
			msg:      "valid post form",
			endpoint: storePendingRebalanceQuadratic,
			method:   http.MethodPost,
			data: map[string]string{
				"value": testData,
			},
			assert: httputil.ExpectSuccess,
		},
		{
			msg:      "setting when pending exists",
			endpoint: storePendingRebalanceQuadratic,
			method:   http.MethodPost,
			data: map[string]string{
				"value": testData,
			},
			assert: httputil.ExpectFailure,
		},
		{
			msg:      "getting existing pending equation",
			endpoint: getPendingRebalanceQuadratic,
			method:   http.MethodGet,
			assert:   newAssertGetEquation([]byte(testData)),
		},
		{
			msg:      "confirm with wrong data",
			endpoint: confirmReblanceQuadratic,
			method:   http.MethodPost,
			data: map[string]string{
				"value": testWrongDataConfirmation,
			},
			assert: httputil.ExpectFailure,
		},
		{
			msg:      "confirm with correct data",
			endpoint: confirmReblanceQuadratic,
			method:   http.MethodPost,
			data: map[string]string{
				"value": testData,
			},
			assert: httputil.ExpectSuccess,
		},
		{
			msg:      "getting exists reabalance quadratic equation",
			endpoint: getRebalanceQuadratic,
			method:   http.MethodGet,
			assert:   newAssertGetEquation([]byte(testData)),
		},
		{
			msg:      "valid post form",
			endpoint: storePendingRebalanceQuadratic,
			method:   http.MethodPost,
			data: map[string]string{
				"value": testData,
			},
			assert: httputil.ExpectSuccess,
		},
		{
			msg:      "reject when there is pending equation",
			endpoint: rejectReblanceQuadratic,
			method:   http.MethodPost,
			data: map[string]string{
				"value": "some random post form or this request will be unauthenticated",
			},
			assert: httputil.ExpectSuccess,
		},
	}

	for _, tc := range tests {
		t.Run(tc.msg, func(t *testing.T) { testHTTPRequest(t, tc, s.r) })
	}
}
