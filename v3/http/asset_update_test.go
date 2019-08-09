package http

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	eth "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KyberNetwork/reserve-data/common/testutil"
	"github.com/KyberNetwork/reserve-data/v3/common"
	"github.com/KyberNetwork/reserve-data/v3/storage"
	"github.com/KyberNetwork/reserve-data/v3/storage/postgres"
)

func createSampleAsset(store storage.Interface) (uint64, error) {
	_, err := store.CreateAssetExchange(0, 1, "ETH", eth.HexToAddress("0x00"), 10,
		0.2, 5.0, 0.3)
	if err != nil {
		return 0, err
	}
	err = store.UpdateExchange(0, storage.UpdateExchangeOpts{
		Disable:         common.BoolPointer(false),
		TradingFeeTaker: common.FloatPointer(0.1),
		TradingFeeMaker: common.FloatPointer(0.2),
	})
	if err != nil {
		return 0, err
	}

	id, err := store.CreateAsset("ABC", "ABC", eth.HexToAddress("0x00000000000000001"),
		18, true, common.ExchangeFeed, true, false, &common.AssetPWI{
			Bid: common.PWIEquation{
				A:                   0,
				B:                   0,
				C:                   0,
				MinMinSpread:        0,
				PriceMultiplyFactor: 0,
			},
			Ask: common.PWIEquation{
				A:                   0,
				B:                   0,
				C:                   0,
				MinMinSpread:        0,
				PriceMultiplyFactor: 0,
			},
		}, &common.RebalanceQuadratic{
			A: 0,
			B: 0,
			C: 0,
		}, []common.AssetExchange{
			{
				Symbol:            "ABC",
				DepositAddress:    eth.HexToAddress("0x00001"),
				ExchangeID:        0,
				TargetRatio:       0.1,
				TargetRecommended: 1000.0,
				WithdrawFee:       0.5,
				MinDeposit:        100.0,
				TradingPairs: []common.TradingPair{
					{
						Quote:           1,
						Base:            0,
						AmountLimitMax:  1.0,
						AmountLimitMin:  1.0,
						MinNotional:     1.0,
						AmountPrecision: 1.0,
						PriceLimitMax:   1.0,
						PriceLimitMin:   1.0,
						PricePrecision:  1.0,
					},
				},
			},
		}, &common.AssetTarget{
			TransferThreshold:  1.0,
			RebalanceThreshold: 1.0,
			Reserve:            1.0,
			Total:              100.0,
		})
	if err != nil {
		return 0, err
	}
	return id, err
}

func createEmptySampleAsset(store storage.Interface) (uint64, error) {
	_, err := store.CreateAssetExchange(2, 1, "KNC", eth.HexToAddress("0x00"), 10,
		0.2, 5.0, 0.3)
	if err != nil {
		return 0, err
	}
	err = store.UpdateExchange(0, storage.UpdateExchangeOpts{
		Disable:         common.BoolPointer(false),
		TradingFeeTaker: common.FloatPointer(0.1),
		TradingFeeMaker: common.FloatPointer(0.2),
	})
	if err != nil {
		return 0, err
	}

	id, err := store.CreateAsset("DEF", "DEF", eth.HexToAddress("0x00000000000000002"),
		18, false, common.SetRateNotSet, false, false, nil, nil, []common.AssetExchange{
			{
				Symbol:            "KNC",
				DepositAddress:    eth.HexToAddress("0x00002"),
				ExchangeID:        2,
				TargetRatio:       0.1,
				TargetRecommended: 1000.0,
				WithdrawFee:       0.5,
				MinDeposit:        100.0,
				TradingPairs: []common.TradingPair{
					{
						Quote:           1,
						Base:            0,
						AmountLimitMax:  1.0,
						AmountLimitMin:  1.0,
						MinNotional:     1.0,
						AmountPrecision: 1.0,
						PriceLimitMax:   1.0,
						PriceLimitMin:   1.0,
						PricePrecision:  1.0,
					},
				},
			},
		}, nil)
	if err != nil {
		return 0, err
	}
	return id, err
}

func readResponse(resp *httptest.ResponseRecorder, out interface{}) error {
	if resp.Code != http.StatusOK {
		return fmt.Errorf("execute failed with http code %d", resp.Code)
	}
	data, _ := ioutil.ReadAll(resp.Body)
	log.Println("resp", string(data))
	return json.Unmarshal(data, out)
}

func TestCreateUpdateAsset(t *testing.T) {

	db, tearDown := testutil.MustNewDevelopmentDB()
	defer func() {
		assert.NoError(t, tearDown())
	}()
	s, err := postgres.NewStorage(db)
	require.NoError(t, err)

	assetID, err := createSampleAsset(s)
	require.NoError(t, err)

	server := NewServer(s, nil)
	const updateAsset = "/v3/update-asset"
	var updateAssetID uint64
	var tests = []testCase{
		{
			msg:      "create update asset",
			endpoint: updateAsset,
			method:   http.MethodPost,
			data: &common.CreateUpdateAsset{
				Assets: []common.UpdateAssetEntry{
					{
						AssetID:      assetID,
						Symbol:       common.StringPointer("XYZ"),
						Name:         common.StringPointer("ZXC"),
						Address:      common.AddressPointer(eth.HexToAddress("0x02")),
						Decimals:     common.Uint64Pointer(19),
						Transferable: common.BoolPointer(false),
						SetRate:      common.SetRatePointer(common.BTCFeed),
						Rebalance:    common.BoolPointer(true),
						IsQuote:      common.BoolPointer(false),
						Target: &common.AssetTarget{
							Total:              5.0,
							Reserve:            5.0,
							RebalanceThreshold: 5.0,
							TransferThreshold:  5.0,
						},
						RebalanceQuadratic: &common.RebalanceQuadratic{
							A: 5.0,
							B: 5.0,
							C: 5.0,
						},
						PWI: &common.AssetPWI{
							Ask: common.PWIEquation{
								A:                   5.0,
								B:                   5.0,
								C:                   5.0,
								MinMinSpread:        5.0,
								PriceMultiplyFactor: 5.0,
							},
							Bid: common.PWIEquation{
								A:                   5.0,
								B:                   5.0,
								C:                   5.0,
								MinMinSpread:        5.0,
								PriceMultiplyFactor: 5.0,
							},
						},
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
				updateAssetID = idResponse.ID
			},
		},
		{
			msg: "confirm update asset",
			endpointExp: func() string {
				return updateAsset + fmt.Sprintf("/%d", updateAssetID)
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
			msg:      "get asset",
			endpoint: fmt.Sprintf("/v3/asset/%d", assetID),
			method:   http.MethodGet,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var response struct {
					Success bool         `json:"success"`
					Data    common.Asset `json:"data"`
				}
				err = readResponse(resp, &response)
				require.NoError(t, err)
				assert.True(t, response.Success)
				assert.Equal(t, "XYZ", response.Data.Symbol)
				assert.Equal(t, "ZXC", response.Data.Name)
				assert.Equal(t, eth.HexToAddress("0x02"), response.Data.Address)
				assert.Equal(t, uint64(19), response.Data.Decimals)
				assert.Equal(t, false, response.Data.Transferable)
				assert.Equal(t, common.BTCFeed, response.Data.SetRate)
				assert.Equal(t, true, response.Data.Rebalance)
				assert.Equal(t, false, response.Data.IsQuote)
				assert.Equal(t, &common.AssetTarget{
					Total:              5.0,
					Reserve:            5.0,
					RebalanceThreshold: 5.0,
					TransferThreshold:  5.0,
				}, response.Data.Target)
				assert.Equal(t, &common.RebalanceQuadratic{
					A: 5.0,
					B: 5.0,
					C: 5.0,
				}, response.Data.RebalanceQuadratic)
				assert.Equal(t, &common.AssetPWI{
					Ask: common.PWIEquation{
						A:                   5.0,
						B:                   5.0,
						C:                   5.0,
						MinMinSpread:        5.0,
						PriceMultiplyFactor: 5.0,
					},
					Bid: common.PWIEquation{
						A:                   5.0,
						B:                   5.0,
						C:                   5.0,
						MinMinSpread:        5.0,
						PriceMultiplyFactor: 5.0,
					},
				}, response.Data.PWI)
			},
		},
		{
			msg: "confirm a not exists update_asset",
			endpointExp: func() string {
				return updateAsset + fmt.Sprintf("/%d", updateAssetID)
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
			msg:      "create update asset",
			endpoint: updateAsset,
			method:   http.MethodPost,
			data: &common.CreateUpdateAsset{
				Assets: []common.UpdateAssetEntry{
					{
						AssetID: assetID,
						Symbol:  common.StringPointer("ETC"),
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
				updateAssetID = idResponse.ID
			},
		},
		{
			msg:      "verify update asset created",
			endpoint: updateAsset,
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
			msg: "reject update asset",
			endpointExp: func() string {
				return updateAsset + fmt.Sprintf("/%d", updateAssetID)
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
			msg:      "verify all update asset removed",
			endpoint: updateAsset,
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

func TestCheckUpdateAssetParams(t *testing.T) {
	db, tearDown := testutil.MustNewDevelopmentDB()
	defer func() {
		assert.NoError(t, tearDown())
	}()
	s, err := postgres.NewStorage(db)
	require.NoError(t, err)

	assetID, err := createSampleAsset(s)
	emptyAssetID, err := createEmptySampleAsset(s)
	require.NoError(t, err)

	server := NewServer(s, nil)

	const updateAsset = "/v3/update-asset"
	var tests = []testCase{
		{ // pwi in server is not nil
			data: common.CreateUpdateAsset{
				Assets: []common.UpdateAssetEntry{
					{
						AssetID: assetID,
						SetRate: common.SetRatePointer(common.GoldFeed),
					},
				},
			},
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var idResponse struct {
					Success bool `json:"success"`
				}
				err = readResponse(resp, &idResponse)
				require.NoError(t, err)
				assert.True(t, idResponse.Success)
			},
			endpoint: updateAsset,
			method:   http.MethodPost,
		}, { // pwi in server is  nil and pwi param is nil
			data: common.CreateUpdateAsset{
				Assets: []common.UpdateAssetEntry{
					{
						AssetID: emptyAssetID,
						SetRate: common.SetRatePointer(common.GoldFeed),
					},
				},
			},
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var idResponse struct {
					Success bool   `json:"success"`
					Reason  string `json:"reason"`
				}
				err = readResponse(resp, &idResponse)
				require.NoError(t, err)
				assert.False(t, idResponse.Success)
				assert.True(t, strings.HasPrefix(idResponse.Reason, common.ErrPWIMissing.Error()))
			},
			endpoint: updateAsset,
			method:   http.MethodPost,
		}, { // pwi in server is  nil and pwi param is not nil
			data: common.CreateUpdateAsset{
				Assets: []common.UpdateAssetEntry{
					{
						AssetID: emptyAssetID,
						SetRate: common.SetRatePointer(common.GoldFeed),
						PWI: &common.AssetPWI{
							Ask: common.PWIEquation{
								A:                   5.0,
								B:                   5.0,
								C:                   5.0,
								MinMinSpread:        5.0,
								PriceMultiplyFactor: 5.0,
							},
							Bid: common.PWIEquation{
								A:                   5.0,
								B:                   5.0,
								C:                   5.0,
								MinMinSpread:        5.0,
								PriceMultiplyFactor: 5.0,
							},
						},
					},
				},
			},
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var idResponse struct {
					Success bool `json:"success"`
				}
				err = readResponse(resp, &idResponse)
				require.NoError(t, err)
				assert.True(t, idResponse.Success)
			},
			endpoint: updateAsset,
			method:   http.MethodPost,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.msg, func(t *testing.T) { testHTTPRequest(t, tc, server.r) })
	}
}
