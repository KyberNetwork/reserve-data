package storage

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/testutil"
	"github.com/KyberNetwork/reserve-data/lib/rtypes"
	commonv3 "github.com/KyberNetwork/reserve-data/reservesetting/common"
)

const (
	migrationPath = "../../cmd/migrations"
)

func TestRate(t *testing.T) {
	db, teardown := testutil.MustNewDevelopmentDB(migrationPath)
	defer func() {
		require.NoError(t, teardown())
	}()

	ps, err := NewPostgresStorage(db)
	require.NoError(t, err)

	// test store data
	baseBuy, _ := big.NewInt(0).SetString("940916409070162411520", 10)
	baseSell, _ := big.NewInt(0).SetString("1051489536265074", 10)
	rateData := common.AllRateEntry{
		Data: map[rtypes.AssetID]common.RateEntry{
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
	assert.Equal(t, common.Version(timepoint), currentRateVersion)

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
	db, teardown := testutil.MustNewDevelopmentDB(migrationPath)
	defer func() {
		require.NoError(t, teardown())
	}()

	ps, err := NewPostgresStorage(db)
	require.NoError(t, err)

	// test store data
	priceData := common.AllPriceEntry{
		Data: map[rtypes.TradingPairID]common.OnePrice{
			2: {
				rtypes.Binance: common.ExchangePrice{
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
					Error:     "",
					Valid:     true,
					Timestamp: "1568358536753",
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
	assert.Equal(t, common.Version(timepoint), currentPriceVersion)

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
	db, teardown := testutil.MustNewDevelopmentDB(migrationPath)
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
		Params: &common.ActivityParams{
			Amount:    39811.443679,
			Exchange:  rtypes.Binance,
			Timepoint: uint64(1568622125860),
			Asset:     2, // KNC id
		},
		Result: &common.ActivityResult{
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
		*activityTest.Params, *activityTest.Result, activityTest.ExchangeStatus, activityTest.MiningStatus, 1568622125860, true, common.TimeToMillis(time.Now())) // true is default pending status
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
	activity, err := ps.GetActivity(rtypes.Binance, testID.EID)
	assert.NoError(t, err)
	assert.Equal(t, activityTest.ID, activity.ID)
}

func TestAuthData(t *testing.T) {
	db, teardown := testutil.MustNewDevelopmentDB(migrationPath)
	defer func() {
		require.NoError(t, teardown())
	}()

	ps, err := NewPostgresStorage(db)
	require.NoError(t, err)
	var (
		ETH = rtypes.AssetID(1)
		KNC = rtypes.AssetID(2)
	)
	authDataTest := common.AuthDataSnapshot{
		Valid:      true,
		Error:      "",
		Timestamp:  "1568705819377",
		ReturnTime: "1568705821452",
		ExchangeBalances: map[rtypes.ExchangeID]common.EBalanceEntry{
			rtypes.Binance: {
				Valid:      true,
				Error:      "",
				Timestamp:  "1568705819377",
				ReturnTime: "1568705819461",
				AvailableBalance: map[rtypes.AssetID]float64{
					ETH: 177.72330689,
					KNC: 3851.21689913,
				},
				LockedBalance: map[rtypes.AssetID]float64{
					ETH: 0,
					KNC: 0,
				},
				DepositBalance: map[rtypes.AssetID]float64{
					ETH: 0,
					KNC: 0,
				},
				Status: true,
			},
		},
		ReserveBalances: map[rtypes.AssetID]common.BalanceEntry{
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
	db, teardown := testutil.MustNewDevelopmentDB(migrationPath)
	defer func() {
		require.NoError(t, teardown())
	}()

	ps, err := NewPostgresStorage(db)
	require.NoError(t, err)

	goldTest := common.GoldData{
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
			Value:     "0",
			Message:   "API Key Not Valid. Please go to 1forge.com to get an API key. If you have any questions please email us at contact@1forge.com",
			Timestamp: 0,
		},
		OneForgeUSD: common.OneForgeGoldData{
			Text:      "",
			Error:     true,
			Value:     "0",
			Message:   "API Key Not Valid. Please go to 1forge.com to get an API key. If you have any questions please email us at contact@1forge.com",
			Timestamp: 0,
		},
	}
	err = ps.StoreGoldInfo(goldTest)
	assert.NoError(t, err)
}

func TestBTCData(t *testing.T) {
	db, teardown := testutil.MustNewDevelopmentDB(migrationPath)
	defer func() {
		require.NoError(t, teardown())
	}()

	ps, err := NewPostgresStorage(db)
	require.NoError(t, err)

	btcTest := common.BTCData{
		Timestamp: 1573788299000,
		Coinbase: common.FeedProviderResponse{
			Bid:   0.123,
			Ask:   0.234,
			Valid: true,
		},
		Binance: common.FeedProviderResponse{
			Bid:   0.345,
			Ask:   0.456,
			Valid: true,
		},
	}
	err = ps.StoreBTCInfo(btcTest)
	assert.NoError(t, err)
}

func TestUSDData(t *testing.T) {
	db, teardown := testutil.MustNewDevelopmentDB(migrationPath)
	defer func() {
		require.NoError(t, teardown())
	}()

	ps, err := NewPostgresStorage(db)
	require.NoError(t, err)

	usdTest := common.USDData{
		CoinbaseETHUSDDAI5000: common.FeedProviderResponse{
			Bid:   0.123,
			Ask:   0.234,
			Valid: true,
		},
		CurveDAIUSDC10000: common.FeedProviderResponse{
			Bid:   1.0123,
			Ask:   1.0234,
			Valid: true,
		},
	}
	err = ps.StoreUSDInfo(usdTest)
	assert.NoError(t, err)
}
