package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	eth "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1common "github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/testutil"
	"github.com/KyberNetwork/reserve-data/http/httputil"
	"github.com/KyberNetwork/reserve-data/v3/common"
	"github.com/KyberNetwork/reserve-data/v3/storage/postgres"
)

func TestHTTPServerTradingPair(t *testing.T) {

	const (
		assetBase           = "/v3/asset"
		createAssetBase     = "/v3/create-asset"
		createTradingPair   = "/v3/create-trading-pair"
		updateTradingPair   = "/v3/update-trading-pair"
		createAssetExchange = "/v3/create-asset-exchange"
		getTradingPair      = "/v3/trading-pair"
	)

	db, tearDown := testutil.MustNewDevelopmentDB()
	defer func() {
		assert.NoError(t, tearDown())
	}()

	//create map of test exchange
	for _, exchangeID := range []v1common.ExchangeName{v1common.Binance, v1common.Huobi, v1common.StableExchange} {
		exhID := v1common.ExchangeID(exchangeID.String())
		exchange := v1common.TestExchange{}
		v1common.SupportedExchanges[exhID] = exchange
	}

	s, err := postgres.NewStorage(db)
	require.NoError(t, err)
	// asset = 1 for ETH is pre-insert in DB.
	_, err = s.CreateAssetExchange(0, 1, "ETH", eth.HexToAddress("0x00"), 10,
		0.2, 5.0, 0.3)
	require.NoError(t, err)
	server := NewServer(s, nil)
	huobiID := uint64(1)

	var createPEAWithQuoteFalse = getCreatePEAWithQuoteFalse()

	var tests = []testCase{

		{
			msg:      "create pending asset",
			endpoint: createAssetBase,
			method:   http.MethodPost,
			assert:   httputil.ExpectSuccess,
			data:     createPEA,
		},
		{
			msg:      "confirm pending asset",
			endpoint: createAssetBase + "/1",
			method:   http.MethodPut,
			data:     nil,
			assert:   httputil.ExpectSuccess,
		},
		{
			msg:      "create pending asset",
			endpoint: createAssetBase,
			method:   http.MethodPost,
			assert:   httputil.ExpectSuccess,
			data:     createPEAWithQuoteFalse,
		},
		{
			msg:      "confirm pending asset",
			endpoint: createAssetBase + "/2",
			method:   http.MethodPut,
			data:     nil,
			assert:   httputil.ExpectSuccess,
		},
		{
			msg:      "receive asset",
			endpoint: assetBase + "/2",
			method:   http.MethodGet,
			assert:   httputil.ExpectSuccess,
		},
		{
			msg:      "receive asset",
			endpoint: assetBase + "/3",
			method:   http.MethodGet,
			assert:   httputil.ExpectSuccess,
		},
		{
			msg:      "create asset exchange",
			endpoint: createAssetExchange,
			method:   http.MethodPost,
			data: common.CreateCreateAssetExchange{
				AssetExchanges: []common.CreateAssetExchangeEntry{
					{
						AssetID:           1,
						ExchangeID:        huobiID,
						Symbol:            "ETH",
						DepositAddress:    eth.HexToAddress("0x001"),
						MinDeposit:        100.0,
						WithdrawFee:       100.0,
						TargetRecommended: 100.0,
						TargetRatio:       100.0,
					},
					{
						AssetID:           2,
						ExchangeID:        huobiID,
						Symbol:            "ABC",
						DepositAddress:    eth.HexToAddress("0x001"),
						MinDeposit:        100.0,
						WithdrawFee:       100.0,
						TargetRecommended: 100.0,
						TargetRatio:       100.0,
					},
					{
						AssetID:           3,
						ExchangeID:        huobiID,
						Symbol:            "ABC",
						DepositAddress:    eth.HexToAddress("0x001"),
						MinDeposit:        100.0,
						WithdrawFee:       100.0,
						TargetRecommended: 100.0,
						TargetRatio:       100.0,
					},
				},
			},
			assert: httputil.ExpectSuccess,
		},
		{
			msg:      "confirm asset exchange",
			endpoint: createAssetExchange + "/3",
			method:   http.MethodPut,
			assert:   httputil.ExpectSuccess,
		},
		{
			msg:      "create trading pair",
			endpoint: createTradingPair,
			method:   http.MethodPost,
			data: common.CreateCreateTradingPair{
				TradingPairs: []common.CreateTradingPairEntry{
					{
						TradingPair: common.TradingPair{
							Base:            2,
							Quote:           1,
							PricePrecision:  1.0,
							AmountPrecision: 1.0,
							AmountLimitMin:  100.0,
							AmountLimitMax:  1000.0,
							PriceLimitMin:   100.0,
							PriceLimitMax:   1000.0,
							MinNotional:     100.0,
						},
						ExchangeID: huobiID,
					},
				},
			},
			assert: httputil.ExpectSuccess,
		},
		{
			msg:      "confirm create trading pair",
			endpoint: createTradingPair + "/4",
			method:   http.MethodPut,
			assert:   httputil.ExpectSuccess, // will fail because we did not config assets on exchange id=1
		},
		{
			msg:      "create update trading pair",
			endpoint: updateTradingPair,
			method:   http.MethodPost,
			data: common.CreateUpdateTradingPair{
				TradingPairs: []common.UpdateTradingPairEntry{
					{
						ID:              2,
						PricePrecision:  common.Uint64Pointer(10),
						AmountPrecision: common.Uint64Pointer(11),
						AmountLimitMin:  common.FloatPointer(12.0),
						AmountLimitMax:  common.FloatPointer(13.0),
						PriceLimitMin:   common.FloatPointer(14.0),
						PriceLimitMax:   common.FloatPointer(15.0),
						MinNotional:     common.FloatPointer(16.0),
					},
				},
			},
			assert: httputil.ExpectSuccess,
		},
		{
			msg:      "confirm update trading pair",
			endpoint: updateTradingPair + "/5",
			method:   http.MethodPut,
			assert:   httputil.ExpectSuccess,
		},
		{
			msg:      "verify updated trading pair",
			endpoint: getTradingPair + "/2",
			method:   http.MethodGet,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assert.Equal(t, resp.Code, http.StatusOK)
				var result struct {
					Success bool                      `json:"success"`
					Data    common.TradingPairSymbols `json:"data"`
				}
				err = json.NewDecoder(resp.Body).Decode(&result)
				require.NoError(t, err)
				res := result.Data
				assert.Equal(t, uint64(10), res.PricePrecision)
				assert.Equal(t, uint64(11), res.AmountPrecision)
				assert.Equal(t, 12.0, res.AmountLimitMin)
				assert.Equal(t, 13.0, res.AmountLimitMax)
				assert.Equal(t, 14.0, res.PriceLimitMin)
				assert.Equal(t, 15.0, res.PriceLimitMax)
				assert.Equal(t, 16.0, res.MinNotional)
			},
		},
		{
			msg:      "create trading pair with invalid quote",
			endpoint: createTradingPair,
			method:   http.MethodPost,
			data: common.CreateCreateTradingPair{
				TradingPairs: []common.CreateTradingPairEntry{
					{
						TradingPair: common.TradingPair{
							Base:            2,
							Quote:           3,
							PricePrecision:  1.0,
							AmountPrecision: 1.0,
							AmountLimitMin:  100.0,
							AmountLimitMax:  1000.0,
							PriceLimitMin:   100.0,
							PriceLimitMax:   1000.0,
							MinNotional:     100.0,
						},
						ExchangeID: huobiID,
					},
				},
			},
			assert: httputil.ExpectFailureWithReason("quote asset should have is_quote=true: quote asset is invalid"),
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.msg, func(t *testing.T) { testHTTPRequest(t, tc, server.r) })
	}
}
