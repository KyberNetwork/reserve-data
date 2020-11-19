package http

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1common "github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/testutil"
	"github.com/KyberNetwork/reserve-data/http/httputil"
	rtypes "github.com/KyberNetwork/reserve-data/lib/rtypes"
	"github.com/KyberNetwork/reserve-data/reservesetting/common"
	"github.com/KyberNetwork/reserve-data/reservesetting/storage/postgres"
)

func TestServer_StableTokenParams(t *testing.T) {

	const (
		settingChangePath = "/v3/setting-change-stable"
		getParamsPath     = "/v3/stable-token-params"
	)
	var (
		supportedExchanges = make(map[rtypes.ExchangeID]v1common.LiveExchange)
	)

	//create map of test exchange
	for _, exchangeID := range []rtypes.ExchangeID{rtypes.Binance, rtypes.Huobi} {
		exchange := v1common.TestExchange{}
		supportedExchanges[exchangeID] = exchange
	}

	db, tearDown := testutil.MustNewDevelopmentDB(migrationPath)
	defer func() {
		assert.NoError(t, tearDown())
	}()

	s, err := postgres.NewStorage(db)
	require.NoError(t, err)

	server := NewServer(s, "", supportedExchanges, "", nil, nil, nil, 0)

	require.NoError(t, err)
	var tests = []testCase{
		{
			msg:      "create request",
			endpoint: settingChangePath,
			method:   http.MethodPost,
			data: common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeUpdateStableTokenParams,
						Data: common.UpdateStableTokenParamsEntry{
							Params: map[string]interface{}{
								"a": "def",
								"b": 2,
								"e": []int{1, 2, 3},
							},
						},
					},
				},
			},
			assert: httputil.ExpectSuccess,
		},
		{
			msg:      "confirm request",
			endpoint: settingChangePath + "/1",
			method:   http.MethodPut,
			assert:   httputil.ExpectSuccess,
		},
		{
			msg:      "get stable token params request",
			endpoint: getParamsPath,
			method:   http.MethodGet,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				require.Equal(t, 200, resp.Code)
				msg, err := ioutil.ReadAll(resp.Body)
				require.NoError(t, err)
				t.Log(string(msg))
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Log(tc.msg)
		t.Run(tc.msg, func(t *testing.T) { testHTTPRequest(t, tc, server.r) })
	}

}
