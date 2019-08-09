package http

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KyberNetwork/reserve-data/common/testutil"
	"github.com/KyberNetwork/reserve-data/v3/common"
	"github.com/KyberNetwork/reserve-data/v3/storage/postgres"
)

func TestUpdateExchange(t *testing.T) {
	db, tearDown := testutil.MustNewDevelopmentDB()
	defer func() {
		assert.NoError(t, tearDown())
	}()
	s, err := postgres.NewStorage(db)
	require.NoError(t, err)

	exchangeID := uint64(1)
	// pre-insert exchange

	server := NewServer(s, nil)
	const updateExchange = "/v3/update-exchange"
	var updateExchID uint64
	var tests = []testCase{
		{
			msg:      "create update exchange",
			endpoint: updateExchange,
			method:   http.MethodPost,
			data: &common.CreateUpdateExchange{
				Exchanges: []common.UpdateExchangeEntry{
					{
						ExchangeID:      exchangeID,
						TradingFeeMaker: common.FloatPointer(0.4),
						TradingFeeTaker: common.FloatPointer(0.6),
						Disable:         common.BoolPointer(false),
					},
				},
			},
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var idResponse struct {
					ID      uint64 `json:"id"`
					Success bool   `json:"success"`
				}
				err = readResponse(resp, &idResponse)
				require.NoError(t, err)
				updateExchID = idResponse.ID
			},
		},
		{
			msg: "confirm update exchange",
			endpointExp: func() string {
				return updateExchange + fmt.Sprintf("/%d", updateExchID)
			},
			method: http.MethodPut,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var idResponse struct {
					Success bool `json:"success"`
				}
				err = readResponse(resp, &idResponse)
				require.NoError(t, err)
				assert.True(t, idResponse.Success)
			},
		},
		{
			msg:      "get exchange",
			endpoint: fmt.Sprintf("/v3/exchange/%d", exchangeID),
			method:   http.MethodGet,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var response struct {
					Success bool            `json:"success"`
					Data    common.Exchange `json:"data"`
				}
				err = readResponse(resp, &response)
				require.NoError(t, err)
				assert.True(t, response.Success)
				assert.Equal(t, false, response.Data.Disable)
				assert.Equal(t, 0.4, response.Data.TradingFeeMaker)
				assert.Equal(t, 0.6, response.Data.TradingFeeTaker)
			},
		},
		{
			msg: "confirm a not exists update_exchange",
			endpointExp: func() string {
				return updateExchange + fmt.Sprintf("/%d", updateExchID)
			},
			method: http.MethodPut,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var idResponse struct {
					Success bool `json:"success"`
				}
				err = readResponse(resp, &idResponse)
				require.NoError(t, err)
				assert.True(t, idResponse.Success == false)
			},
		},
		{
			msg:      "create update exchange",
			endpoint: updateExchange,
			method:   http.MethodPost,
			data: &common.CreateUpdateExchange{
				Exchanges: []common.UpdateExchangeEntry{
					{
						ExchangeID:      exchangeID,
						TradingFeeMaker: common.FloatPointer(0.4),
						TradingFeeTaker: common.FloatPointer(0.6),
						Disable:         common.BoolPointer(false),
					},
				},
			},
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var idResponse struct {
					ID      uint64 `json:"id"`
					Success bool   `json:"success"`
				}
				err = readResponse(resp, &idResponse)
				require.NoError(t, err)
				updateExchID = idResponse.ID
			},
		},
		{
			msg:      "verify update exchange created",
			endpoint: updateExchange,
			method:   http.MethodGet,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var idResponse struct {
					Data    []common.PendingObject `json:"data"`
					Success bool                   `json:"success"`
				}
				err = readResponse(resp, &idResponse)
				require.NoError(t, err)
				assert.Len(t, idResponse.Data, 1)
			},
		},
		{
			msg: "reject update exchange",
			endpointExp: func() string {
				return updateExchange + fmt.Sprintf("/%d", updateExchID)
			},
			method: http.MethodPut,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var idResponse struct {
					Success bool `json:"success"`
				}
				err = readResponse(resp, &idResponse)
				require.NoError(t, err)
				assert.True(t, idResponse.Success)
			},
		},
		{
			msg:      "verify all update exchange removed",
			endpoint: updateExchange,
			method:   http.MethodGet,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var idResponse struct {
					Data    []common.PendingObject `json:"data"`
					Success bool                   `json:"success"`
				}
				err = readResponse(resp, &idResponse)
				require.NoError(t, err)
				assert.Len(t, idResponse.Data, 0)
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.msg, func(t *testing.T) { testHTTPRequest(t, tc, server.r) })
	}
}
