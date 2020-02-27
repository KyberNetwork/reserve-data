package storage

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/testutil"
	commonv3 "github.com/KyberNetwork/reserve-data/reservesetting/common"
)

func TestRate(t *testing.T) {
	db, teardown := testutil.MustNewDevelopmentDB()
	defer func() {
		require.NoError(t, teardown())
	}()

	ps, err := NewPostgresStorage(db)
	require.NoError(t, err)

	// test store data
	baseBuy, _ := big.NewInt(0).SetString("940916409070162411520", 10)
	baseSell, _ := big.NewInt(0).SetString("1051489536265074", 10)
	rateData := common.AllRateEntry{
		Data: map[uint64]common.RateEntry{
			2: {
				Block:       8539892,
				BaseBuy:     baseBuy,
				BaseSell:    baseSell,
				CompactBuy:  -11,
				CompactSell: 7,
			},
		},
		Timestamp:   "1568358532784",
		ReturnTime:  "1568358532956",
		BlockNumber: 8539899,
	}
	timepoint := uint64(1568358532784)

	// test store rate
	err = ps.StoreRate(rateData, timepoint)
	require.NoError(t, err)

	// test get current version
	timepointTest := uint64(1568358532785)
	currentRateVersion, err := ps.CurrentRateVersion(timepointTest)
	require.NoError(t, err)
	assert.Equal(t, common.Version(1), currentRateVersion)

	// test there is no version
	timepointTest = uint64(1568358532783)
	_, err = ps.CurrentRateVersion(timepointTest)
	assert.NotNil(t, err)

	// Test get rate
	rate, err := ps.GetRate(currentRateVersion)
	require.NoError(t, err)
	assert.Equal(t, rateData, rate)
}

func TestPrice(t *testing.T) {
	db, teardown := testutil.MustNewDevelopmentDB()
	defer func() {
		require.NoError(t, teardown())
	}()

	ps, err := NewPostgresStorage(db)
	require.NoError(t, err)

	// test store data
	priceData := common.AllPriceEntry{
		Data: map[uint64]common.OnePrice{
			2: {
				common.ExchangeID(1): common.ExchangePrice{
					Asks: []common.PriceEntry{
						{
							Rate:     0.001062,
							Quantity: 6,
						},
						{
							Rate:     0.0010677,
							Quantity: 376,
						},
					},
					Bids: []common.PriceEntry{
						{
							Rate:     0.0010603,
							Quantity: 46,
						},
						{
							Rate:     0.0010593,
							Quantity: 46,
						},
					},
					Error:      "",
					Valid:      true,
					Timestamp:  "1568358536753",
					ReturnTime: "1568358536834",
				},
			},
		},
		Block: 8539900,
	}

	timepoint := uint64(1568358536753)

	// test store price
	err = ps.StorePrice(priceData, timepoint)
	require.NoError(t, err)

	// test get current version
	timepointTest := uint64(1568358536753)
	currentPriceVersion, err := ps.CurrentPriceVersion(timepointTest)
	require.NoError(t, err)
	assert.Equal(t, common.Version(1), currentPriceVersion)

	// test there is no version
	timepointTest = uint64(1568358532783)
	_, err = ps.CurrentPriceVersion(timepointTest)
	assert.NotNil(t, err)

	// Test get rate
	prices, err := ps.GetAllPrices(currentPriceVersion)
	require.NoError(t, err)
	assert.Equal(t, priceData, prices)
}

func TestActivity(t *testing.T) {
	db, teardown := testutil.MustNewDevelopmentDB()
	defer func() {
		require.NoError(t, teardown())
	}()

	ps, err := NewPostgresStorage(db)
	require.NoError(t, err)

	activityTest := common.ActivityRecord{
		Action: "deposit",
		ID: common.ActivityID{
			Timepoint: 1568622132671609009,
			EID:       "0x7437e2ac582a7cdef75a6c8355d03167a8ab7670a178197d81f14cea76684d74|BQX|39811.443679",
		},
		Destination: "binance",
		Params: common.ActivityParams{
			Amount:    39811.443679,
			Exchange:  common.Binance,
			Timepoint: uint64(1568622125860),
			Asset:     2, // KNC id
		},
		Result: common.ActivityResult{
			BlockNumber: 8559409,
			Error:       "",
			GasPrice:    "50100000000",
			Nonce:       11039,
			StatusError: "",
			Tx:          "0x7437e2ac582a7cdef75a6c8355d03167a8ab7670a178197d81f14cea76684d74",
		},
		ExchangeStatus: "",
		MiningStatus:   "mined",
		Timestamp:      "1568622125860",
	}
	err = ps.Record(activityTest.Action, activityTest.ID, activityTest.Destination,
		activityTest.Params, activityTest.Result, activityTest.ExchangeStatus, activityTest.MiningStatus, 1568622125860)
	assert.NoError(t, err)

	hasPending, err := ps.HasPendingDeposit(commonv3.Asset{ID: 2}, common.TestExchange{})
	assert.NoError(t, err)
	assert.True(t, hasPending)

	// test update activity
	testID := common.ActivityID{
		Timepoint: 1568622132671609009,
		EID:       "0x7437e2ac582a7cdef75a6c8355d03167a8ab7670a178197d81f14cea76684d74|BQX|39811.443679",
	}

	activityTest.ExchangeStatus = common.ExchangeStatusDone
	err = ps.UpdateActivity(testID, activityTest)
	assert.NoError(t, err)

	hasPending, err = ps.HasPendingDeposit(commonv3.Asset{ID: 2}, common.TestExchange{})
	assert.NoError(t, err)
	assert.False(t, hasPending)

	// test get activity
	activity, err := ps.GetActivity(common.Binance, testID.EID)
	assert.NoError(t, err)
	assert.Equal(t, activityTest, activity)
}

func TestAuthData(t *testing.T) {
	db, teardown := testutil.MustNewDevelopmentDB()
	defer func() {
		require.NoError(t, teardown())
	}()

	ps, err := NewPostgresStorage(db)
	require.NoError(t, err)
	var (
		ETH = common.AssetID(1)
		KNC = common.AssetID(2)
	)
	authDataTest := common.AuthDataSnapshot{
		Valid:      true,
		Error:      "",
		Timestamp:  "1568705819377",
		ReturnTime: "1568705821452",
		ExchangeBalances: map[common.ExchangeID]common.EBalanceEntry{
			common.Binance: {
				Valid:      true,
				Error:      "",
				Timestamp:  "1568705819377",
				ReturnTime: "1568705819461",
				AvailableBalance: map[common.AssetID]float64{
					ETH: 177.72330689,
					KNC: 3851.21689913,
				},
				LockedBalance: map[common.AssetID]float64{
					ETH: 0,
					KNC: 0,
				},
				DepositBalance: map[common.AssetID]float64{
					ETH: 0,
					KNC: 0,
				},
				Status: true,
			},
		},
		ReserveBalances: map[common.AssetID]common.BalanceEntry{
			ETH: {
				Valid:      true,
				Error:      "",
				Timestamp:  "1568705820671",
				ReturnTime: "1568705820937",
				Balance:    common.RawBalance(*big.NewInt(432048208)),
			},
			KNC: {
				Valid:      true,
				Error:      "",
				Timestamp:  "1568705820671",
				ReturnTime: "1568705820937",
				Balance:    common.RawBalance(*big.NewInt(3194712941)),
			},
		},
		Block: 8565634,
	}

	timepoint := uint64(1560842137000)
	err = ps.StoreAuthSnapshot(&authDataTest, timepoint)
	assert.NoError(t, err)

	// get authdata
	version, err := ps.CurrentAuthDataVersion(1560842137001)
	assert.NoError(t, err)

	getAuthData, err := ps.GetAuthData(version)
	assert.NoError(t, err)
	assert.Equal(t, authDataTest, getAuthData)

	// prune outdated data
	timepoint = common.NowInMillis()
	deleted, err := ps.PruneExpiredAuthData(timepoint)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), deleted)
}

func TestGoldData(t *testing.T) {
	db, teardown := testutil.MustNewDevelopmentDB()
	defer func() {
		require.NoError(t, teardown())
	}()

	ps, err := NewPostgresStorage(db)
	require.NoError(t, err)

	goldTest := common.GoldData{
		DGX: common.DGXGoldData{
			Data: []common.GoldRate{
				{
					Time:   1568356189,
					Price:  0.27008084,
					Symbol: "DGXETH",
				},
			},
			Error:     "",
			Valid:     true,
			Status:    "",
			Timestamp: 0,
		},
		GDAX: common.GDAXGoldData{
			Ask:     "179.52",
			Bid:     "179.51",
			Size:    "3.00000000",
			Time:    "2019-09-13T06:29:23.037Z",
			Error:   "",
			Valid:   true,
			Price:   "179.53000000",
			Volume:  "34118.55527070",
			TradeID: 51291927,
		},
		Gemini: common.GeminiGoldData{
			Ask:   "179.56",
			Bid:   "179.52",
			Last:  "179.53",
			Error: "",
			Valid: true,
			Volume: struct {
				ETH       string `json:"ETH"`
				USD       string `json:"USD"`
				Timestamp uint64 `json:"timestamp"`
			}{
				ETH:       "2587.61414071",
				USD:       "464667.9095133173",
				Timestamp: 1568355900000,
			},
		},
		Kraken: common.KrakenGoldData{
			Valid:           true,
			ErrorFromKraken: nil,
			Result: map[string]struct {
				A []string `json:"a"`
				B []string `json:"b"`
				C []string `json:"c"`
				V []string `json:"v"`
				P []string `json:"p"`
				T []uint64 `json:"t"`
				L []string `json:"l"`
				H []string `json:"h"`
				O string   `json:"o"`
			}{
				"XETHZUSD": {
					A: []string{
						"179.55000",
						"120",
						"120.000",
					},
					B: []string{
						"179.49000",
						"33",
						"33.000",
					},
					C: []string{
						"179.54000",
						"0.05000000",
					},
					H: []string{
						"181.87000",
						"182.70000",
					},
					L: []string{
						"179.00000",
						"176.51000",
					},
					O: "180.99000",
					P: []string{
						"180.18664",
						"179.75713",
					},
					T: []uint64{
						606,
						3526,
					},
					V: []string{
						"3894.72837992",
						"17861.39347398",
					},
				},
			},
			Error: "",
		},
		Timestamp: 1568356191628,
		OneForgeETH: common.OneForgeGoldData{
			Text:      "",
			Error:     true,
			Value:     0,
			Message:   "API Key Not Valid. Please go to 1forge.com to get an API key. If you have any questions please email us at contact@1forge.com",
			Timestamp: 0,
		},
		OneForgeUSD: common.OneForgeGoldData{
			Text:      "",
			Error:     true,
			Value:     0,
			Message:   "API Key Not Valid. Please go to 1forge.com to get an API key. If you have any questions please email us at contact@1forge.com",
			Timestamp: 0,
		},
	}
	err = ps.StoreGoldInfo(goldTest)
	assert.NoError(t, err)
}

func TestBTCData(t *testing.T) {
	db, teardown := testutil.MustNewDevelopmentDB()
	defer func() {
		require.NoError(t, teardown())
	}()

	ps, err := NewPostgresStorage(db)
	require.NoError(t, err)

	btcTest := common.BTCData{
		Timestamp: 1573788299000,
		Coinbase: common.CoinbaseData{
			TradeID: 8544259,
			Price:   "0.02129",
			Size:    "2",
			Time:    "2019-11-15T03:22:51.126Z",
			Bid:     "0.02129",
			Ask:     "0.0213",
			Volume:  "9018.21928763",
		},
		Gemini: common.GeminiData{
			Bid: "0.02129",
			Ask: "0.0213",
			Volume: common.GeminiDataVolume{
				ETH:       "904.59371255",
				BTC:       "19.3047430451109",
				Timestamp: 1573788300000,
			},
			Last: "0.02124",
		},
	}
	err = ps.StoreBTCInfo(btcTest)
	assert.NoError(t, err)
}

func TestUSDData(t *testing.T) {
	db, teardown := testutil.MustNewDevelopmentDB()
	defer func() {
		require.NoError(t, teardown())
	}()

	ps, err := NewPostgresStorage(db)
	require.NoError(t, err)

	usdTest := common.USDData{
		CoinbaseUSD: common.CoinbaseData{
			TradeID: 309230,
			Price:   "182.31",
			Size:    "0.10697894",
			Time:    "2019-11-15T07:20:18.433Z",
			Bid:     "182.31",
			Ask:     "182.39",
			Volume:  "3382.10669068",
		},
		GeminiUSD: common.GeminiGoldData{
			Bid: "182.35",
			Ask: "182.36",
			Volume: struct {
				ETH       string `json:"ETH"`
				USD       string `json:"USD"`
				Timestamp uint64 `json:"timestamp"`
			}{
				ETH:       "7633.16020108",
				USD:       "1402216.1627546342",
				Timestamp: 1573802400000,
			},
			Last: "182.40",
		}, // USD and Gold use the same url
		CoinbaseUSDC: common.CoinbaseData{
			TradeID: 309230,
			Price:   "182.31",
			Size:    "0.10697894",
			Time:    "2019-11-15T07:20:18.433Z",
			Bid:     "182.31",
			Ask:     "182.39",
			Volume:  "3382.10669068",
		},
		CoinbaseDAI: common.CoinbaseData{
			TradeID: 309230,
			Price:   "182.31",
			Size:    "0.10697894",
			Time:    "2019-11-15T07:20:18.433Z",
			Bid:     "182.31",
			Ask:     "182.39",
			Volume:  "3382.10669068",
		},
		BinanceUSDC: common.BinanceData{
			Symbol:   "ETHUSDC",
			BidPrice: "182.43000000",
			BidQty:   "27.00000000",
			AskPrice: "182.54000000",
			AskQty:   "4.00000000",
		},
		BinanceUSDT: common.BinanceData{
			Symbol:   "ETHUSDT",
			BidPrice: "182.98000000",
			BidQty:   "0.27320000",
			AskPrice: "183.00000000",
			AskQty:   "27.22601000",
		},
		BinancePAX: common.BinanceData{
			Symbol:   "ETHPAX",
			BidPrice: "182.42000000",
			BidQty:   "4.75001000",
			AskPrice: "182.63000000",
			AskQty:   "5.00000000",
		},
		BinanceTUSD: common.BinanceData{
			Symbol:   "ETHTUSD",
			BidPrice: "182.40000000",
			BidQty:   "0.98270000",
			AskPrice: "182.68000000",
			AskQty:   "4.22871000",
		},
		HitDAI: common.HitData{
			Ask:         "182.195",
			Bid:         "181.638",
			Last:        "182.042",
			Open:        "184.691",
			Low:         "179.778",
			High:        "185.474",
			Volume:      "5239.2112",
			VolumeQuote: "960500.2438935",
			Timestamp:   "2019-11-15T07:34:27.024Z",
			Symbol:      "ETHDAI",
		},
		BitFinex: common.BitFinexData{
			Bid:             182.88,
			BidSize:         949.2862453,
			Ask:             182.91,
			AskSize:         1129.0389414699998,
			DailyChange:     -3.06,
			DailyChangePerc: -0.0164,
			LastPrice:       183,
			Volume:          33814.09021745,
			High:            186.64,
			Low:             180.80213703,
		},
	}
	err = ps.StoreUSDInfo(usdTest)
	assert.NoError(t, err)
}
