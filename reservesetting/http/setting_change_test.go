package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	eth "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1common "github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/feed"
	"github.com/KyberNetwork/reserve-data/common/testutil"
	"github.com/KyberNetwork/reserve-data/http/httputil"
	rtypes "github.com/KyberNetwork/reserve-data/lib/rtypes"
	"github.com/KyberNetwork/reserve-data/reservesetting/common"
	"github.com/KyberNetwork/reserve-data/reservesetting/storage"
	"github.com/KyberNetwork/reserve-data/reservesetting/storage/postgres"
)

func createSampleAsset(store *postgres.Storage) (rtypes.AssetID, error) {
	_, err := store.CreateAssetExchange(binance, 1, "ETH", eth.HexToAddress("0x00"), 10,
		0.2, 5.0, 0.3, nil)
	if err != nil {
		return 0, err
	}
	err = store.UpdateExchange(binance, storage.UpdateExchangeOpts{
		Disable:         common.BoolPointer(false),
		TradingFeeTaker: common.FloatPointer(0.1),
		TradingFeeMaker: common.FloatPointer(0.2),
	})
	if err != nil {
		return 0, err
	}

	id, err := store.CreateAsset("ABC", "ABC", eth.HexToAddress("0x00000000000000001"),
		18, true, common.ExchangeFeed, true, false, true, &common.AssetPWI{
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
			SizeA:  0,
			SizeB:  0,
			SizeC:  0,
			PriceA: 0,
			PriceB: 0,
			PriceC: 0,
		}, []common.AssetExchange{
			{
				Symbol:            "ABC",
				DepositAddress:    eth.HexToAddress("0x00001"),
				ExchangeID:        binance,
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
		}, nil, nil, 0.1, 0.2, 10000, 1, 2)
	if err != nil {
		return 0, err
	}
	return id, err
}

// TODO write more test cases
func TestServer_SettingChangeBasic(t *testing.T) {
	const (
		settingChangePath = "/v3/setting-change-main"

		expectedAskSpread = 34.1
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

	assetID, err := createSampleAsset(s)
	require.NoError(t, err)

	server := NewServer(s, "", supportedExchanges, "", nil, nil, nil)

	emptyFeedWeight := make(common.FeedWeight)
	btcFeed := common.BTCFeed

	var tests = []testCase{
		{
			msg:      "create asset exchange",
			endpoint: settingChangePath,
			method:   http.MethodPost,
			data: &common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeCreateAssetExchange,
						Data: common.CreateAssetExchangeEntry{
							AssetID:           assetID,
							ExchangeID:        huobi,
							Symbol:            "ETH",
							DepositAddress:    eth.HexToAddress("0x007"),
							MinDeposit:        10.0,
							WithdrawFee:       11.0,
							TargetRecommended: 12.0,
							TargetRatio:       13.0,
						},
					},
					{
						Type: common.ChangeTypeUpdateAssetExchange,
						Data: common.UpdateAssetExchangeEntry{
							ID:             2,
							DepositAddress: common.AddressPointer(eth.HexToAddress("0x0000000000001")),
							MinDeposit:     common.FloatPointer(6.0),
						},
					},
				},
				Message: "support huobi",
			},
			assert: httputil.ExpectSuccess,
		},
		{
			msg:      "confirm setting change",
			endpoint: fmt.Sprint(settingChangePath, "/", 1),
			method:   http.MethodPut,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				httputil.ExpectSuccess(t, resp)
				//check update asset exchange
				assetExchange, err := s.GetAssetExchange(2)
				require.NoError(t, err)
				require.Equal(t, 6.0, assetExchange.MinDeposit)
				//check create asset exchange
				asset, err := s.GetAsset(assetID)
				require.NoError(t, err)
				found := false
				for _, x := range asset.Exchanges {
					if x.ExchangeID == huobi {
						assetExchange = x
						found = true
						break
					}
				}
				require.Equal(t, true, found)
				require.Equal(t, "ETH", assetExchange.Symbol)
				require.Equal(t, huobi, assetExchange.ExchangeID)
			},
		},
		{
			msg:      "create asset",
			endpoint: settingChangePath,
			method:   http.MethodPost,
			data: &common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeCreateAsset,
						Data: common.CreateAssetEntry{
							Symbol:    "KNC",
							Name:      "Kyber",
							Address:   eth.HexToAddress("0x3f105f78359ad80562b4c34296a87b8e66c584c5"),
							Decimals:  18,
							SetRate:   common.SetRateNotSet,
							IsQuote:   true,
							Rebalance: true,
							PWI: &common.AssetPWI{
								Ask: common.PWIEquation{
									A:                   234,
									B:                   23,
									C:                   12,
									MinMinSpread:        234,
									PriceMultiplyFactor: 123,
								},
								Bid: common.PWIEquation{
									A:                   23,
									B:                   234,
									C:                   234,
									MinMinSpread:        234,
									PriceMultiplyFactor: 234,
								},
							},
							Target: &common.AssetTarget{
								Total:              12,
								Reserve:            24,
								TransferThreshold:  34,
								RebalanceThreshold: 1,
							},
							Exchanges: []common.AssetExchange{
								{
									Symbol:      "KNC",
									ExchangeID:  binance,
									MinDeposit:  34,
									WithdrawFee: 34,
									TargetRatio: 34,
									TradingPairs: []common.TradingPair{
										{
											Quote: 1,
										},
									},
									DepositAddress:    eth.HexToAddress("0x3f105f78359ad80562b4c34296a87b8e66c584c5"),
									TargetRecommended: 234,
								},
							},
							RebalanceQuadratic: &common.RebalanceQuadratic{
								SizeA:  12,
								SizeB:  34,
								SizeC:  24,
								PriceA: 9,
								PriceB: 15,
								PriceC: 21,
							},
							StableParam: &common.StableParam{
								AskSpread: expectedAskSpread,
							},
							NormalUpdatePerPeriod: 0.123,
							MaxImbalanceRatio:     0.456,
							OrderDurationMillis:   15000,
							PriceETHAmount:        1,
							ExchangeETHAmount:     2,
						},
					},
				},
				Message: "create a new asset",
			},
			assert: httputil.ExpectSuccess,
		},
		{
			msg:      "confirm setting change",
			endpoint: fmt.Sprint(settingChangePath, "/", 2),
			method:   http.MethodPut,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				httputil.ExpectSuccess(t, resp)
				asset, err := s.GetAssetBySymbol("KNC")
				require.NoError(t, err)
				asset, err = s.GetAsset(asset.ID)
				require.NoError(t, err)
				require.Equal(t, 0.0, asset.StableParam.PriceUpdateThreshold)
				assert.Equal(t, expectedAskSpread, asset.StableParam.AskSpread)
				require.Equal(t, 0.123, asset.NormalUpdatePerPeriod)
				require.Equal(t, 0.456, asset.MaxImbalanceRatio)
				require.Equal(t, uint64(15000), asset.OrderDurationMillis)
				require.Equal(t, float64(1), asset.PriceETHAmount)
				require.Equal(t, float64(2), asset.ExchangeETHAmount)
			},
		},
		{
			msg:      "update asset",
			endpoint: settingChangePath,
			method:   http.MethodPost,
			data: &common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeUpdateAsset,
						Data: common.UpdateAssetEntry{
							AssetID:               assetID,
							NormalUpdatePerPeriod: common.FloatPointer(0.111),
							MaxImbalanceRatio:     common.FloatPointer(0.222),
							OrderDurationMillis:   common.Uint64Pointer(15666),
							PriceETHAmount:        common.FloatPointer(1.1),
							ExchangeETHAmount:     common.FloatPointer(2.2),
						},
					},
				},
				Message: "update asset",
			},
			assert: httputil.ExpectSuccess,
		},
		{
			msg:      "confirm update setting change",
			endpoint: fmt.Sprint(settingChangePath, "/", 3),
			method:   http.MethodPut,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				httputil.ExpectSuccess(t, resp)
				asset, err := s.GetAsset(assetID)
				require.NoError(t, err)
				require.Equal(t, 0.111, asset.NormalUpdatePerPeriod)
				require.Equal(t, 0.222, asset.MaxImbalanceRatio)
				require.Equal(t, uint64(15666), asset.OrderDurationMillis)
				require.Equal(t, float64(1.1), asset.PriceETHAmount)
				require.Equal(t, float64(2.2), asset.ExchangeETHAmount)
			},
		},
		{
			msg:      "create asset with feed weight",
			endpoint: settingChangePath,
			method:   http.MethodPost,
			data: &common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeCreateAsset,
						Data: common.CreateAssetEntry{
							Symbol:    "OMG",
							Name:      "Omisego",
							Address:   eth.HexToAddress("0xd26114cd6ee289accf82350c8d8487fedb8a0c07"),
							Decimals:  18,
							SetRate:   common.BTCFeed,
							IsQuote:   true,
							Rebalance: true,
							PWI: &common.AssetPWI{
								Ask: common.PWIEquation{
									A:                   234,
									B:                   23,
									C:                   12,
									MinMinSpread:        234,
									PriceMultiplyFactor: 123,
								},
								Bid: common.PWIEquation{
									A:                   23,
									B:                   234,
									C:                   234,
									MinMinSpread:        234,
									PriceMultiplyFactor: 234,
								},
							},
							Target: &common.AssetTarget{
								Total:              12,
								Reserve:            24,
								TransferThreshold:  34,
								RebalanceThreshold: 1,
							},
							Exchanges: []common.AssetExchange{
								{
									Symbol:      "OMG",
									ExchangeID:  binance,
									MinDeposit:  34,
									WithdrawFee: 34,
									TargetRatio: 34,
									TradingPairs: []common.TradingPair{
										{
											Quote: 1,
										},
									},
									DepositAddress:    eth.HexToAddress("0x3f105f78359ad80562b4c34296a87b8e66c584c5"),
									TargetRecommended: 234,
								},
							},
							RebalanceQuadratic: &common.RebalanceQuadratic{
								SizeA:  12,
								SizeB:  34,
								SizeC:  24,
								PriceA: 9,
								PriceB: 15,
								PriceC: 21,
							},
							StableParam: &common.StableParam{
								AskSpread: expectedAskSpread,
							},
							FeedWeight: &common.FeedWeight{
								feed.CoinbaseETHBTC3.String(): 3.0,
								feed.BinanceETHBTC3.String():  1.2,
							},
							NormalUpdatePerPeriod: 0.123,
							MaxImbalanceRatio:     0.456,
							OrderDurationMillis:   10000,
							PriceETHAmount:        1.2,
							ExchangeETHAmount:     2.1,
						},
					},
				},
				Message: "create OMG",
			},
			assert: httputil.ExpectSuccess,
		},
		{
			msg:      "confirm setting change with feed weight",
			endpoint: fmt.Sprint(settingChangePath, "/", 4),
			method:   http.MethodPut,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				httputil.ExpectSuccess(t, resp)
				asset, err := s.GetAssetBySymbol("OMG")
				require.NoError(t, err)
				asset, err = s.GetAsset(asset.ID)
				require.NoError(t, err)
				require.Equal(t, 0.0, asset.StableParam.PriceUpdateThreshold)
				assert.Equal(t, expectedAskSpread, asset.StableParam.AskSpread)
				assert.Equal(t, 2, len(*asset.FeedWeight))
			},
		},
		{
			msg:      "update asset with feed weight",
			endpoint: settingChangePath,
			method:   http.MethodPost,
			data: &common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeUpdateAsset,
						Data: common.UpdateAssetEntry{
							SetRate: &btcFeed,
							AssetID: 6,
							FeedWeight: &common.FeedWeight{
								feed.BinanceETHBTC3.String(): 3.0,
							},
						},
					},
				},
				Message: "Update feed weight",
			},
			assert: httputil.ExpectSuccess,
		},
		{
			msg:      "confirm update change with feed weight",
			endpoint: fmt.Sprint(settingChangePath, "/", 5),
			method:   http.MethodPut,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				httputil.ExpectSuccess(t, resp)
				asset, err := s.GetAssetBySymbol("OMG")
				require.NoError(t, err)
				asset, err = s.GetAsset(asset.ID)
				require.NoError(t, err)
				require.Equal(t, 0.0, asset.StableParam.PriceUpdateThreshold)
				assert.Equal(t, expectedAskSpread, asset.StableParam.AskSpread)
				assert.Equal(t, 1, len(*asset.FeedWeight))
			},
		},
		{
			msg:      "update asset with ignoring feed weight",
			endpoint: settingChangePath,
			method:   http.MethodPost,
			data: &common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeUpdateAsset,
						Data: common.UpdateAssetEntry{
							AssetID: 6,
							IsQuote: common.BoolPointer(false),
						},
					},
				},
				Message: "Change quote type",
			},
			assert: httputil.ExpectSuccess,
		},
		{
			msg:      "confirm update change with ignoring feed weight",
			endpoint: fmt.Sprint(settingChangePath, "/", 6),
			method:   http.MethodPut,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				httputil.ExpectSuccess(t, resp)
				asset, err := s.GetAssetBySymbol("OMG")
				require.NoError(t, err)
				asset, err = s.GetAsset(asset.ID)
				require.NoError(t, err)
				require.Equal(t, 0.0, asset.StableParam.PriceUpdateThreshold)
				assert.Equal(t, expectedAskSpread, asset.StableParam.AskSpread)
				assert.Equal(t, 1, len(*asset.FeedWeight))
			},
		},
		{
			msg:      "update asset remove feed weight",
			endpoint: settingChangePath,
			method:   http.MethodPost,
			data: &common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeUpdateAsset,
						Data: common.UpdateAssetEntry{
							SetRate:    &btcFeed,
							AssetID:    6,
							FeedWeight: &emptyFeedWeight,
						},
					},
				},
				Message: "remove feed weight",
			},
			assert: httputil.ExpectSuccess,
		},
		{
			msg:      "confirm update change remove feed weight",
			endpoint: fmt.Sprint(settingChangePath, "/", 7),
			method:   http.MethodPut,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				httputil.ExpectSuccess(t, resp)
				asset, err := s.GetAssetBySymbol("OMG")
				require.NoError(t, err)
				asset, err = s.GetAsset(asset.ID)
				require.NoError(t, err)
				require.Equal(t, 0.0, asset.StableParam.PriceUpdateThreshold)
				assert.Equal(t, expectedAskSpread, asset.StableParam.AskSpread)
				assert.Equal(t, 0, len(*asset.FeedWeight))
			},
		},
		{
			msg:      "create asset with feed weight failed",
			endpoint: settingChangePath,
			method:   http.MethodPost,
			data: &common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeCreateAsset,
						Data: common.CreateAssetEntry{
							Symbol:    "SNT",
							Name:      "Status",
							Address:   eth.HexToAddress("0x744d70fdbe2ba4cf95131626614a1763df805b9e"),
							Decimals:  18,
							SetRate:   common.BTCFeed,
							IsQuote:   true,
							Rebalance: true,
							PWI: &common.AssetPWI{
								Ask: common.PWIEquation{
									A:                   234,
									B:                   23,
									C:                   12,
									MinMinSpread:        234,
									PriceMultiplyFactor: 123,
								},
								Bid: common.PWIEquation{
									A:                   23,
									B:                   234,
									C:                   234,
									MinMinSpread:        234,
									PriceMultiplyFactor: 234,
								},
							},
							Target: &common.AssetTarget{
								Total:              12,
								Reserve:            24,
								TransferThreshold:  34,
								RebalanceThreshold: 1,
							},
							Exchanges: []common.AssetExchange{
								{
									Symbol:      "SNT",
									ExchangeID:  binance,
									MinDeposit:  34,
									WithdrawFee: 34,
									TargetRatio: 34,
									TradingPairs: []common.TradingPair{
										{
											Quote: 1,
										},
									},
									DepositAddress:    eth.HexToAddress("0x3f105f78359ad80562b4c34296a87b8e66c584c5"),
									TargetRecommended: 234,
								},
							},
							RebalanceQuadratic: &common.RebalanceQuadratic{
								SizeA:  12,
								SizeB:  34,
								SizeC:  24,
								PriceA: 9,
								PriceB: 15,
								PriceC: 21,
							},
							StableParam: &common.StableParam{
								AskSpread: expectedAskSpread,
							},
							FeedWeight: &common.FeedWeight{
								"some_random_feed":           1.2,
								feed.BinanceETHBTC3.String(): 3.0,
							},
							NormalUpdatePerPeriod: 0.123,
							MaxImbalanceRatio:     0.456,
						},
					},
				},
				Message: "update feed weight",
			},
			assert: httputil.ExpectFailure,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.msg, func(t *testing.T) {
			testHTTPRequest(t, tc, server.r)
		})
	}

}

func TestHTTPServerAssetExchangeWithOptionalTradingPair(t *testing.T) {
	const (
		settingChangePath = "/v3/setting-change-main"
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

	assetID, err := createSampleAsset(s)
	require.NoError(t, err)

	_, err = s.CreateAssetExchange(rtypes.ExchangeID(2), rtypes.AssetID(1), "ETH",
		eth.HexToAddress("0x8cacc5bf46ea076a09febb2ae213dc4da6384e74"),
		0.001,
		0.001,
		0.001,
		0.001, nil,
	)
	require.NoError(t, err)

	asset, err := s.GetAsset(1)
	require.NoError(t, err)
	t.Log(asset)

	server := NewServer(s, "", supportedExchanges, "", nil, nil, nil)

	var tests = []testCase{
		{
			msg:      "create asset exchange with invalid trading pair",
			endpoint: settingChangePath,
			method:   http.MethodPost,
			assert:   httputil.ExpectFailure,
			data: &common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeCreateAssetExchange,
						Data: common.CreateAssetExchangeEntry{
							AssetID:           assetID,
							ExchangeID:        huobi,
							Symbol:            "ETH",
							DepositAddress:    eth.HexToAddress("0x001"),
							MinDeposit:        10.0,
							WithdrawFee:       11.0,
							TargetRecommended: 12.0,
							TargetRatio:       13.0,
							TradingPairs: []common.TradingPair{
								{
									Base:            1,
									Quote:           1,
									AmountPrecision: 10,
									AmountLimitMin:  1000,
									AmountLimitMax:  10000,
									PriceLimitMin:   0.1,
									PriceLimitMax:   10.10,
									MinNotional:     0.001,
								},
							},
						},
					},
				},
				Message: "Support huobi",
			},
		},
		{
			msg:      "create asset exchange with invalid trading pair - both base and quote are zero",
			endpoint: settingChangePath,
			method:   http.MethodPost,
			assert:   httputil.ExpectFailure,
			data: &common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeCreateAssetExchange,
						Data: common.CreateAssetExchangeEntry{
							AssetID:           assetID,
							ExchangeID:        huobi,
							Symbol:            "ETH",
							DepositAddress:    eth.HexToAddress("0x001"),
							MinDeposit:        10.0,
							WithdrawFee:       11.0,
							TargetRecommended: 12.0,
							TargetRatio:       13.0,
							TradingPairs: []common.TradingPair{
								{
									Base:            0,
									Quote:           0,
									AmountPrecision: 10,
									AmountLimitMin:  1000,
									AmountLimitMax:  10000,
									PriceLimitMin:   0.1,
									PriceLimitMax:   10.10,
									MinNotional:     0.001,
								},
							},
						},
					},
				},
				Message: "Create trading pair",
			},
		},
		{
			msg:      "create asset exchange with invalid trading pair - invalid quote token",
			endpoint: settingChangePath,
			method:   http.MethodPost,
			assert:   httputil.ExpectFailure,
			data: &common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeCreateAssetExchange,
						Data: common.CreateAssetExchangeEntry{
							AssetID:           assetID,
							ExchangeID:        huobi,
							Symbol:            "ETH",
							DepositAddress:    eth.HexToAddress("0x001"),
							MinDeposit:        10.0,
							WithdrawFee:       11.0,
							TargetRecommended: 12.0,
							TargetRatio:       13.0,
							TradingPairs: []common.TradingPair{
								{
									Base:            0,
									Quote:           1234,
									AmountPrecision: 10,
									AmountLimitMin:  1000,
									AmountLimitMax:  10000,
									PriceLimitMin:   0.1,
									PriceLimitMax:   10.10,
									MinNotional:     0.001,
								},
							},
						},
					},
				},
				Message: "create trading pair",
			},
		},
		{
			msg:      "create asset exchange with invalid trading pair - invalid base token",
			endpoint: settingChangePath,
			method:   http.MethodPost,
			assert:   httputil.ExpectFailure,
			data: &common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeCreateAssetExchange,
						Data: common.CreateAssetExchangeEntry{
							AssetID:           assetID,
							ExchangeID:        huobi,
							Symbol:            "ETH",
							DepositAddress:    eth.HexToAddress("0x001"),
							MinDeposit:        10.0,
							WithdrawFee:       11.0,
							TargetRecommended: 12.0,
							TargetRatio:       13.0,
							TradingPairs: []common.TradingPair{
								{
									Base:            1234,
									Quote:           0,
									AmountPrecision: 10,
									AmountLimitMin:  1000,
									AmountLimitMax:  10000,
									PriceLimitMin:   0.1,
									PriceLimitMax:   10.10,
									MinNotional:     0.001,
								},
							},
						},
					},
				},
				Message: "create trading pair",
			},
		},
		{
			msg:      "create asset exchange with duplicate trading pair",
			endpoint: settingChangePath,
			method:   http.MethodPost,
			assert:   httputil.ExpectFailureWithReason(`failed to create trading pair base=2 quote=1 exchange_id=2 err=duplicate key value violates unique constraint "trading_pairs_exchange_id_base_id_quote_id_key"`),
			data: &common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeCreateAssetExchange,
						Data: common.CreateAssetExchangeEntry{
							AssetID:           assetID,
							ExchangeID:        huobi,
							Symbol:            "ETH",
							DepositAddress:    eth.HexToAddress("0x001"),
							MinDeposit:        10.0,
							WithdrawFee:       11.0,
							TargetRecommended: 12.0,
							TargetRatio:       13.0,
							TradingPairs: []common.TradingPair{
								{
									Base:            0,
									Quote:           1,
									AmountPrecision: 10,
									AmountLimitMin:  1000,
									AmountLimitMax:  10000,
									PriceLimitMin:   0.1,
									PriceLimitMax:   10.10,
									MinNotional:     0.001,
								},
								{
									Base:            0,
									Quote:           1,
									AmountPrecision: 10,
									AmountLimitMin:  1000,
									AmountLimitMax:  10000,
									PriceLimitMin:   0.1,
									PriceLimitMax:   10.10,
									MinNotional:     0.001,
								},
							},
						},
					},
				},
				Message: "create new trading pair",
			},
		},
		{
			msg:      "create asset exchange with trading pair successfully",
			endpoint: settingChangePath,
			method:   http.MethodPost,
			assert:   httputil.ExpectSuccess,
			data: &common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeCreateAssetExchange,
						Data: common.CreateAssetExchangeEntry{
							AssetID:           assetID,
							ExchangeID:        huobi,
							Symbol:            "ETH",
							DepositAddress:    eth.HexToAddress("0x001"),
							MinDeposit:        10.0,
							WithdrawFee:       11.0,
							TargetRecommended: 12.0,
							TargetRatio:       13.0,
							TradingPairs: []common.TradingPair{
								{
									Base:            0,
									Quote:           1,
									AmountPrecision: 10,
									AmountLimitMin:  1000,
									AmountLimitMax:  10000,
									PriceLimitMin:   0.1,
									PriceLimitMax:   10.10,
									MinNotional:     0.001,
								},
							},
						},
					},
				},
				Message: "create new trading pair",
			},
		},
		{
			msg:      "confirm asset exchange with trading pair successfully",
			endpoint: settingChangePath + "/2",
			method:   http.MethodPut,
			assert:   httputil.ExpectSuccess,
			data:     nil,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Log(tc.msg)
		t.Run(tc.msg, func(t *testing.T) { testHTTPRequest(t, tc, server.r) })
	}
}

func TestHTTPServer_SettingChangeUpdateExchange(t *testing.T) {
	db, tearDown := testutil.MustNewDevelopmentDB(migrationPath)
	defer func() {
		assert.NoError(t, tearDown())
	}()
	s, err := postgres.NewStorage(db)
	require.NoError(t, err)

	exchangeID := rtypes.ExchangeID(1)
	// pre-insert exchange
	server := NewServer(s, "", nil, "", nil, nil, nil)
	const updateExchange = "/v3/setting-change-update-exchange"
	var updateExchID uint64
	var tests = []testCase{
		{
			msg:      "create update exchange",
			endpoint: updateExchange,
			method:   http.MethodPost,
			data: common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeUpdateExchange,
						Data: common.UpdateExchangeEntry{
							ExchangeID:      exchangeID,
							TradingFeeMaker: common.FloatPointer(0.4),
							TradingFeeTaker: common.FloatPointer(0.6),
							Disable:         common.BoolPointer(false),
						},
					},
				},
				Message: "Update exchange",
			},
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var idResponse struct {
					ID      uint64 `json:"id"`
					Success bool   `json:"success"`
				}
				require.Equal(t, http.StatusOK, resp.Code)
				err = json.Unmarshal(resp.Body.Bytes(), &idResponse)
				require.NoError(t, err)
				require.True(t, idResponse.Success)
				updateExchID = idResponse.ID
			},
		},
		{
			msg: "test get pending",
			endpointExp: func() string {
				return updateExchange + fmt.Sprintf("/%d", updateExchID)
			},
			method: http.MethodGet,
			assert: httputil.ExpectSuccess,
		},
		{
			msg:      "test get pending objs",
			endpoint: updateExchange,
			method:   http.MethodGet,
			assert:   httputil.ExpectSuccess,
		},
		{
			msg: "confirm update exchange",
			endpointExp: func() string {
				return updateExchange + fmt.Sprintf("/%d", updateExchID)
			},
			method: http.MethodPut,
			assert: httputil.ExpectSuccess,
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
				require.Equal(t, http.StatusOK, resp.Code)
				err = json.Unmarshal(resp.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.True(t, response.Success)
				assert.Equal(t, false, response.Data.Disable)
				assert.Equal(t, 0.4, response.Data.TradingFeeMaker)
				assert.Equal(t, 0.6, response.Data.TradingFeeTaker)
			},
		},
		{
			msg:      "create update exchange",
			endpoint: updateExchange,
			method:   http.MethodPost,
			data: common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeUpdateExchange,
						Data: common.UpdateExchangeEntry{
							ExchangeID:      10000,
							TradingFeeMaker: common.FloatPointer(0.4),
							TradingFeeTaker: common.FloatPointer(0.6),
							Disable:         common.BoolPointer(false),
						},
					},
				},
				Message: "update exchange fee",
			},
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				msg := resp.Body.Bytes()
				require.NoError(t, err)
				t.Logf("%+v", string(msg))
				var response struct {
					Success bool `json:"success"`
				}
				require.Equal(t, http.StatusOK, resp.Code)
				err = json.Unmarshal(msg, &response)
				require.NoError(t, err)
				assert.False(t, response.Success)
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.msg, func(t *testing.T) { testHTTPRequest(t, tc, server.r) })
	}
}

func TestHTTPServer_ChangeAssetAddress(t *testing.T) {
	db, tearDown := testutil.MustNewDevelopmentDB(migrationPath)
	defer func() {
		assert.NoError(t, tearDown())
	}()
	var (
		supportedExchanges = make(map[rtypes.ExchangeID]v1common.LiveExchange)
	)

	//create map of test exchange
	for _, exchangeID := range []rtypes.ExchangeID{rtypes.Binance, rtypes.Huobi} {
		exchange := v1common.TestExchange{}
		supportedExchanges[exchangeID] = exchange
	}

	s, err := postgres.NewStorage(db)
	require.NoError(t, err)
	t.Log(s)
	server := NewServer(s, "", supportedExchanges, "", nil, nil, nil)
	const changeAssetAddress = "/v3/setting-change-main"
	var changeID uint64
	var tests = []testCase{
		{
			msg:      "create change asset address",
			endpoint: changeAssetAddress,
			method:   http.MethodPost,
			data: common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeChangeAssetAddr,
						Data: common.ChangeAssetAddressEntry{
							ID:      1,
							Address: eth.HexToAddress("0x52bc44d5378309EE2abF1539BF71dE1b7d7bE3b5"),
						},
					},
				},
				Message: "update asset address",
			},
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var idResponse struct {
					ID      uint64 `json:"id"`
					Success bool   `json:"success"`
				}
				require.Equal(t, http.StatusOK, resp.Code)
				err = json.Unmarshal(resp.Body.Bytes(), &idResponse)
				require.NoError(t, err)
				require.True(t, idResponse.Success)
				changeID = idResponse.ID
			},
		},
		{
			msg: "test get pending",
			endpointExp: func() string {
				return changeAssetAddress + fmt.Sprintf("/%d", changeID)
			},
			method: http.MethodGet,
			assert: httputil.ExpectSuccess,
		},
		{
			msg:      "test get pending objs",
			endpoint: changeAssetAddress,
			method:   http.MethodGet,
			assert:   httputil.ExpectSuccess,
		},
		{
			msg: "confirm update exchange",
			endpointExp: func() string {
				return changeAssetAddress + fmt.Sprintf("/%d", changeID)
			},
			method: http.MethodPut,
			assert: httputil.ExpectSuccess,
		},
		{
			msg:      "get exchange",
			endpoint: fmt.Sprintf("/v3/asset/%d", 1),
			method:   http.MethodGet,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var response struct {
					Success bool         `json:"success"`
					Data    common.Asset `json:"data"`
				}
				require.Equal(t, http.StatusOK, resp.Code)
				err = json.Unmarshal(resp.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.True(t, response.Success)
				assert.Equal(t, eth.HexToAddress("0x52bc44d5378309EE2abF1539BF71dE1b7d7bE3b5"), response.Data.Address)
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.msg, func(t *testing.T) { testHTTPRequest(t, tc, server.r) })
	}
}

func TestHTTPServer_DeleteTradingPair(t *testing.T) {
	db, tearDown := testutil.MustNewDevelopmentDB(migrationPath)
	defer func() {
		assert.NoError(t, tearDown())
	}()
	var (
		supportedExchanges = make(map[rtypes.ExchangeID]v1common.LiveExchange)
	)

	//create map of test exchange
	for _, exchangeID := range []rtypes.ExchangeID{rtypes.Binance, rtypes.Huobi} {
		exchange := v1common.TestExchange{}
		supportedExchanges[exchangeID] = exchange
	}

	s, err := postgres.NewStorage(db)
	require.NoError(t, err)
	t.Log(s)
	server := NewServer(s, "", supportedExchanges, "", nil, nil, nil)
	_, err = createSampleAsset(s)
	require.NoError(t, err)

	const deleteTradingPair = "/v3/setting-change-main"
	var deleteTradingPairID uint64
	var tests = []testCase{
		{
			msg:      "create delete trading pair",
			endpoint: deleteTradingPair,
			method:   http.MethodPost,
			data: common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeDeleteTradingPair,
						Data: common.DeleteTradingPairEntry{
							TradingPairID: 1,
						},
					},
				},
				Message: "delete trading pair",
			},
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var idResponse struct {
					ID      uint64 `json:"id"`
					Success bool   `json:"success"`
					Reason  string `json:"reason"`
				}
				require.Equal(t, http.StatusOK, resp.Code)
				err = json.Unmarshal(resp.Body.Bytes(), &idResponse)
				require.NoError(t, err)
				require.True(t, idResponse.Success)
				deleteTradingPairID = idResponse.ID
			},
		},
		{
			msg: "test get pending",
			endpointExp: func() string {
				return deleteTradingPair + fmt.Sprintf("/%d", deleteTradingPairID)
			},
			method: http.MethodGet,
			assert: httputil.ExpectSuccess,
		},
		{
			msg: "confirm delete trading pair",
			endpointExp: func() string {
				return deleteTradingPair + fmt.Sprintf("/%d", deleteTradingPairID)
			},
			method: http.MethodPut,
			assert: httputil.ExpectSuccess,
		},
		{
			msg:      "get setting change accepted",
			endpoint: fmt.Sprintf("/v3/setting-change-main?status=%s", common.ChangeStatusAccepted.String()),
			method:   http.MethodGet,
			assert:   httputil.ExpectSuccess,
		},
		{
			msg:      "get trading pair",
			endpoint: fmt.Sprintf("/v3/trading-pair/%d", 1),
			method:   http.MethodGet,
			assert:   httputil.ExpectFailureWithReason(common.ErrNotFound.Error()),
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.msg, func(t *testing.T) { testHTTPRequest(t, tc, server.r) })
	}
}

func TestHTTPServer_DeleteAssetExchange(t *testing.T) {
	db, tearDown := testutil.MustNewDevelopmentDB(migrationPath)
	defer func() {
		assert.NoError(t, tearDown())
	}()

	var (
		supportedExchanges = make(map[rtypes.ExchangeID]v1common.LiveExchange)
	)

	//create map of test exchange
	for _, exchangeID := range []rtypes.ExchangeID{rtypes.Binance, rtypes.Huobi} {
		exchange := v1common.TestExchange{}
		supportedExchanges[exchangeID] = exchange
	}
	s, err := postgres.NewStorage(db)
	require.NoError(t, err)
	server := NewServer(s, "", supportedExchanges, "", nil, nil, nil)
	_, err = createSampleAsset(s)
	require.NoError(t, err)

	const deleteAssetExchange = "/v3/setting-change-main"
	const deleteTradingPair = "/v3/setting-change-main"
	var deleteAssetExchangeID uint64
	var deleteTradingPairID uint64
	var tests = []testCase{
		{
			msg:      "create delete trading pair",
			endpoint: deleteTradingPair,
			method:   http.MethodPost,
			data: common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeDeleteTradingPair,
						Data: common.DeleteTradingPairEntry{
							TradingPairID: 1,
						},
					},
				},
				Message: "delete trading pair",
			},
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var idResponse struct {
					ID      uint64 `json:"id"`
					Success bool   `json:"success"`
					Reason  string `json:"reason"`
				}
				require.Equal(t, http.StatusOK, resp.Code)
				err = json.Unmarshal(resp.Body.Bytes(), &idResponse)
				require.NoError(t, err)
				require.True(t, idResponse.Success)
				deleteTradingPairID = idResponse.ID
			},
		},
		{
			msg: "confirm delete trading pair",
			endpointExp: func() string {
				return deleteTradingPair + fmt.Sprintf("/%d", deleteTradingPairID)
			},
			method: http.MethodPut,
			assert: httputil.ExpectSuccess,
		},
		{
			msg:      "create delete asset exchange",
			endpoint: deleteAssetExchange,
			method:   http.MethodPost,
			data: common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeDeleteAssetExchange,
						Data: common.DeleteAssetExchangeEntry{
							AssetExchangeID: 2,
						},
					},
				},
				Message: "delete asset exchange",
			},
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var idResponse struct {
					ID      uint64 `json:"id"`
					Success bool   `json:"success"`
					Reason  string `json:"reason"`
				}
				require.Equal(t, http.StatusOK, resp.Code)
				err = json.Unmarshal(resp.Body.Bytes(), &idResponse)
				require.NoError(t, err)
				log.Printf("%v", idResponse)
				require.True(t, idResponse.Success)
				deleteAssetExchangeID = idResponse.ID
			},
		},
		{
			msg: "test get pending",
			endpointExp: func() string {
				return deleteAssetExchange + fmt.Sprintf("/%d", deleteAssetExchangeID)
			},
			method: http.MethodGet,
			assert: httputil.ExpectSuccess,
		},
		{
			msg: "confirm delete asset exchange",
			endpointExp: func() string {
				return deleteAssetExchange + fmt.Sprintf("/%d", deleteAssetExchangeID)
			},
			method: http.MethodPut,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var response struct {
					Success bool   `json:"success"`
					Reason  string `json:"reason"`
				}
				require.Equal(t, http.StatusOK, resp.Code)
				err = json.Unmarshal(resp.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Equal(t, true, response.Success)
				_, err := s.GetAssetExchange(2)
				require.Equal(t, common.ErrNotFound, err)
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.msg, func(t *testing.T) { testHTTPRequest(t, tc, server.r) })
	}
}

func TestCreateTradingPair(t *testing.T) {
	var (
		supportedExchanges = make(map[rtypes.ExchangeID]v1common.LiveExchange)
	)

	// create map of test exchange
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
	id, err := createSampleAsset(s)
	require.NoError(t, err)
	server := NewServer(s, "", supportedExchanges, "", nil, nil, nil)
	c := apiClient{s: server}
	quote := rtypes.AssetID(1) // ETH
	postRes, err := c.createSettingChange(common.SettingChange{
		ChangeList: []common.SettingChangeEntry{
			{
				Type: common.ChangeTypeCreateTradingPair,
				Data: common.CreateTradingPairEntry{
					TradingPair: common.TradingPair{
						Base:  id,
						Quote: quote, // ETH
					},
					AssetID:    id,
					ExchangeID: binance,
				},
			},
		},
		Message: "create trading pair",
	})
	require.NoError(t, err)
	_, err = c.confirmSettingChange(postRes.ID)
	require.NoError(t, err)
	assetResp, err := c.getAsset(id)
	require.NoError(t, err)
	found := false
	for _, ex := range assetResp.Asset.Exchanges {
		if ex.ExchangeID == binance {
			for _, tp := range ex.TradingPairs {
				if tp.Quote == quote && tp.Base == id {
					found = true
					break
				}
			}
		}
	}
	assert.Equal(t, true, found)
}

func TestSetFeedConfiguration(t *testing.T) {
	db, tearDown := testutil.MustNewDevelopmentDB(migrationPath)
	defer func() {
		assert.NoError(t, tearDown())
	}()
	s, err := postgres.NewStorage(db)
	require.NoError(t, err)
	supportedExchanges := make(map[rtypes.ExchangeID]v1common.LiveExchange)
	server := NewServer(s, "", supportedExchanges, "", nil, nil, nil)
	var (
		setFeedConfigurationEndpoint = "/v3/setting-change-feed-configuration"
		setFeedConfigurationID       uint64

		fname                 = feed.CoinbaseETHUSDDAI5000.String()
		setRate               = common.USDFeed
		fenabled              = false
		fbaseVolatilitySpread = 1.1
		fnormalSpread         = 1.2

		expectFC = common.FeedConfiguration{
			Name:                 fname,
			SetRate:              setRate,
			Enabled:              fenabled,
			BaseVolatilitySpread: fbaseVolatilitySpread,
			NormalSpread:         fnormalSpread,
		}
	)

	var tests = []testCase{
		{
			msg:      "test create set feed configuration",
			endpoint: setFeedConfigurationEndpoint,
			method:   http.MethodPost,
			data: common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeSetFeedConfiguration,
						Data: common.SetFeedConfigurationEntry{
							Name:                 fname,
							SetRate:              setRate,
							Enabled:              &fenabled,
							BaseVolatilitySpread: &fbaseVolatilitySpread,
							NormalSpread:         &fnormalSpread,
						},
					},
				},
				Message: "set feed configuration",
			},
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var idResponse struct {
					ID      uint64 `json:"id"`
					Success bool   `json:"success"`
					Reason  string `json:"reason"`
				}
				require.Equal(t, http.StatusOK, resp.Code)
				err = json.Unmarshal(resp.Body.Bytes(), &idResponse)
				require.NoError(t, err)
				require.True(t, idResponse.Success)
				setFeedConfigurationID = idResponse.ID
			},
		},
		{
			msg: "test get pending set feed configuration",
			endpointExp: func() string {
				return setFeedConfigurationEndpoint + fmt.Sprintf("/%d", setFeedConfigurationID)
			},
			method: http.MethodGet,
			assert: httputil.ExpectSuccess,
		},
		{
			msg: "confirm set feed configuration",
			endpointExp: func() string {
				return setFeedConfigurationEndpoint + fmt.Sprintf("/%d", setFeedConfigurationID)
			},
			method: http.MethodPut,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var response struct {
					Success bool   `json:"success"`
					Reason  string `json:"reason"`
				}
				require.Equal(t, http.StatusOK, resp.Code)
				err = json.Unmarshal(resp.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Equal(t, true, response.Success)
				newFC, err := s.GetFeedConfiguration(fname, setRate)
				require.NoError(t, err)
				require.Equal(t, expectFC, newFC)
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.msg, func(t *testing.T) { testHTTPRequest(t, tc, server.r) })
	}
}
