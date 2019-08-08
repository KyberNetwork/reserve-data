package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KyberNetwork/reserve-data/common/testutil"
	"github.com/KyberNetwork/reserve-data/http/httputil"
	"github.com/KyberNetwork/reserve-data/v3/common"
	"github.com/KyberNetwork/reserve-data/v3/storage/postgres"
)

func TestServer_TradingBy(t *testing.T) {
	const (
		assetBase       = "/v3/asset"
		createAssetBase = "/v3/create-asset"
		getTradingPair  = "/v3/trading-pair"
		createTradingBy = "/v3/create-trading-by"
	)

	db, tearDown := testutil.MustNewDevelopmentDB()
	defer func() {
		assert.NoError(t, tearDown())
	}()

	s, err := postgres.NewStorage(db)
	// asset = 1 for ETH is pre-insert in DB.
	_, err = s.CreateAssetExchange(0, 1, "ETH", ethereum.HexToAddress("0x00"), 10,
		0.2, 5.0, 0.3)
	require.NoError(t, err)
	require.NoError(t, err)
	server := NewServer(s, nil)

	var tests = []testCase{
		{
			msg:      "create pending asset",
			endpoint: createAssetBase,
			method:   http.MethodPost,
			assert:   httputil.ExpectSuccess,
			data:     createPEA,
		}, {
			msg:      "confirm pending asset",
			endpoint: createAssetBase + "/1",
			method:   http.MethodPut,
			data:     nil,
			assert:   httputil.ExpectSuccess,
		}, {
			msg:      "receive asset",
			endpoint: assetBase + "/2",
			method:   http.MethodGet,
			assert:   httputil.ExpectSuccess,
		}, {
			msg:      "trading pair is created in db",
			endpoint: getTradingPair + "/1",
			method:   http.MethodGet,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assert.Equal(t, resp.Code, http.StatusOK)
				var result struct {
					Success bool                      `json:"success"`
					Data    common.TradingPairSymbols `json:"data"`
				}
				err = json.NewDecoder(resp.Body).Decode(&result)
				require.NoError(t, err)
			},
		}, {
			msg:      "test create pending trading by fail by referring trading pair not found",
			endpoint: createTradingBy,
			method:   http.MethodPost,
			data: common.CreateCreateTradingBy{
				TradingBys: []common.CreateTradingByEntry{
					{
						TradingPairID: 2,
						AssetID:       1,
					},
				},
			},
			assert: httputil.ExpectFailureWithReason(common.ErrNotFound.Error()),
		},
		{
			msg:      "test create pending trading by fail by referring asset id invalid",
			endpoint: createTradingBy,
			method:   http.MethodPost,
			data: common.CreateCreateTradingBy{
				TradingBys: []common.CreateTradingByEntry{
					{
						TradingPairID: 1,
						AssetID:       123,
					},
				},
			},
			assert: httputil.ExpectFailureWithReason(common.ErrTradingByAssetIDInvalid.Error()),
		}, {
			msg:      "test create pending trading by",
			endpoint: createTradingBy,
			method:   http.MethodPost,
			data: common.CreateCreateTradingBy{
				TradingBys: []common.CreateTradingByEntry{
					{
						TradingPairID: 1,
						AssetID:       1,
					},
				},
			},
			assert: httputil.ExpectSuccess,
		}, {
			msg:      "confirm create trading by",
			endpoint: createTradingBy + "/2",
			method:   http.MethodPut,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				httputil.ExpectSuccess(t, resp)
				// get data in db to make sure
				tradingBy, err := server.storage.GetTradingBy(2)
				require.NoError(t, err)
				require.Equal(t, uint64(1), tradingBy.AssetID)
				require.Equal(t, uint64(1), tradingBy.TradingPairID)
			},
		}, {
			msg:      "failed to confirm create trading by (id invalid)",
			endpoint: createTradingBy + "/3",
			method:   http.MethodPut,
			assert:   httputil.ExpectFailureWithReason(common.ErrNotFound.Error()),
		}, {
			msg:      "test create pending trading by",
			endpoint: createTradingBy,
			method:   http.MethodPost,
			data: common.CreateCreateTradingBy{
				TradingBys: []common.CreateTradingByEntry{
					{
						TradingPairID: 1,
						AssetID:       2,
					},
				},
			},
			assert: httputil.ExpectSuccess,
		}, {
			msg:      "test delete pending trading by",
			endpoint: createTradingBy + "/3",
			method:   http.MethodDelete,
			assert:   httputil.ExpectSuccess,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.msg, func(t *testing.T) { testHTTPRequest(t, tc, server.r) })
	}
}
