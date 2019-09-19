package http

import (
	"fmt"
	"log"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/testutil"
	"github.com/KyberNetwork/reserve-data/core"
	"github.com/KyberNetwork/reserve-data/data"
	"github.com/KyberNetwork/reserve-data/data/storage"
	"github.com/KyberNetwork/reserve-data/http/httputil"
	commonv3 "github.com/KyberNetwork/reserve-data/reservesetting/common"
	"github.com/KyberNetwork/reserve-data/reservesetting/storage/postgres"
)

var (
	s              *Server
	st             *storage.PostgresStorage
	settingStorage *postgres.Storage
)

func TestPrices(t *testing.T) {
	const (
		pricesVersionEndpoint = "/v3/prices-version"
		pricesEndpoint        = "/v3/prices"
	)

	var tests = []httputil.TestCase{
		{
			Msg:      "get prices version failure",
			Endpoint: fmt.Sprintf("%s?timestamp=%s", pricesVersionEndpoint, "1568358536752"),
			Method:   http.MethodGet,
			Assert:   httputil.ExpectFailure,
		},
		{
			Msg:      "get prices version success",
			Endpoint: fmt.Sprintf("%s?timestamp=%s", pricesVersionEndpoint, "1568358536754"),
			Method:   http.MethodGet,
			Assert:   httputil.ExpectSuccess,
		},
		{
			Msg:      "get prices success",
			Endpoint: fmt.Sprintf("%s?timestamp=%s", pricesEndpoint, "1568358536754"),
			Method:   http.MethodGet,
			Assert:   httputil.ExpectSuccess,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.Msg, func(t *testing.T) { httputil.TestHTTPRequest(t, tc, s.r) })
	}
}

func TestFeedConfiguration(t *testing.T) {
	const (
		setFeedConfigurationEndpoint = "/v3/set-feed-configuration"
		getFeedConfigurationEndpoint = "/v3/get-feed-configuration"
	)

	var (
		tests = []httputil.TestCase{
			{
				Msg:      "get default configuration",
				Endpoint: getFeedConfigurationEndpoint,
				Method:   http.MethodGet,
				Assert:   httputil.ExpectSuccess,
			},
			{
				Msg:      "set configuration success",
				Endpoint: setFeedConfigurationEndpoint,
				Method:   http.MethodPost,
				Data: common.FeedConfigurationRequest{
					Data: struct {
						Name    string `json:"name"`
						Enabled bool   `json:"enabled"`
					}{
						Name:    "Kraken",
						Enabled: false,
					},
				},
				Assert: httputil.ExpectSuccess,
			},
			{
				Msg:      "get feed configuration after change a config",
				Endpoint: getFeedConfigurationEndpoint,
				Method:   http.MethodGet,
				Assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
					assert.Equal(t, http.StatusOK, resp.Code)
				},
			},
		}
	)

	for _, tc := range tests {
		tc := tc
		t.Run(tc.Msg, func(t *testing.T) { httputil.TestHTTPRequest(t, tc, s.r) })
	}
}

func TestRate(t *testing.T) {
	const (
		getRatesEndpoint    = "/v3/getrates"
		getAllRatesEndpoint = "/v3/get-all-rates"
	)

	var (
		tests = []httputil.TestCase{
			{
				Msg:      "test get rates failure",
				Endpoint: fmt.Sprintf("%s?timestamp=%s", getRatesEndpoint, "1568358532783"),
				Method:   http.MethodGet,
				Assert:   httputil.ExpectFailure,
			},
			{
				Msg:      "test get all rates success",
				Endpoint: getAllRatesEndpoint,
				Method:   http.MethodGet,
				Assert:   httputil.ExpectSuccess,
			},
			{
				Msg:      "test get rates success",
				Endpoint: fmt.Sprintf("%s?timestamp=%s", getRatesEndpoint, "1568358532785"),
				Method:   http.MethodGet,
				Assert:   httputil.ExpectSuccess,
			},
		}
	)

	for _, tc := range tests {
		tc := tc
		t.Run(tc.Msg, func(t *testing.T) { httputil.TestHTTPRequest(t, tc, s.r) })
	}
}

func TestAuthdata(t *testing.T) {
	const (
		getAuthdataVersionEndpoint         = "/v3/authdata-version"
		getAuthdataEndpoint                = "/v3/authdata"
		activitiesEndpoint                 = "/v3/activities"
		immediatePendingActivitiesEndpoint = "/v3/immediate-pending-activities"
	)
	var (
		tests = []httputil.TestCase{
			{
				Msg:      "test get authdata version failure",
				Endpoint: getAuthdataVersionEndpoint,
				Method:   http.MethodGet,
				Assert:   httputil.ExpectFailure,
			},
			{
				Msg:      "test get authdata failure",
				Endpoint: getAuthdataEndpoint,
				Method:   http.MethodGet,
				Assert:   httputil.ExpectFailure,
			},
			{
				Msg:      "test get activities success",
				Endpoint: activitiesEndpoint,
				Method:   http.MethodGet,
				Assert:   httputil.ExpectSuccess,
			},
			{
				Msg:      "test get immediate pending activities success",
				Endpoint: immediatePendingActivitiesEndpoint,
				Method:   http.MethodGet,
				Assert:   httputil.ExpectSuccess,
			},
		}
	)

	for _, tc := range tests {
		tc := tc
		t.Run(tc.Msg, func(t *testing.T) { httputil.TestHTTPRequest(t, tc, s.r) })
	}
}

func TestMain(m *testing.M) {
	var (
		err error
	)
	db, teardown := testutil.MustNewDevelopmentDB()

	settingStorage, err = postgres.NewStorage(db)
	if err != nil {
		log.Fatal(err)
	}

	st, err = storage.NewPostgresStorage(db)
	if err != nil {
		log.Fatal(err)
	}

	// init asset, asset exchange and pair
	_, err = settingStorage.CreateAssetExchange(uint64(common.Binance), 1, "ETH", ethereum.HexToAddress("0x00"), 10, 0.2, 5.0, 0.3, nil)
	if err != nil {
		log.Fatal(err)
	}

	// create asset
	pwi := &commonv3.AssetPWI{
		Ask: commonv3.PWIEquation{
			A:                   234,
			B:                   23,
			C:                   12,
			MinMinSpread:        234,
			PriceMultiplyFactor: 123,
		},
		Bid: commonv3.PWIEquation{
			A:                   23,
			B:                   234,
			C:                   234,
			MinMinSpread:        234,
			PriceMultiplyFactor: 234,
		},
	}
	target := &commonv3.AssetTarget{
		Total:              12,
		Reserve:            24,
		TransferThreshold:  34,
		RebalanceThreshold: 1,
	}
	exchanges := []commonv3.AssetExchange{
		{
			Symbol:      "KNC",
			ExchangeID:  uint64(common.Binance),
			MinDeposit:  34,
			WithdrawFee: 34,
			TargetRatio: 34,
			TradingPairs: []commonv3.TradingPair{
				{
					Quote: 1,
				},
			},
			DepositAddress:    ethereum.HexToAddress("0x3f105f78359ad80562b4c34296a87b8e66c584c5"),
			TargetRecommended: 234,
		},
	}
	rb := &commonv3.RebalanceQuadratic{
		A: 12,
		B: 34,
		C: 24,
	}
	_, err = settingStorage.CreateAsset("KNC", "KyberNetwork", ethereum.HexToAddress("0xdd974d5c2e2928dea5f71b9825b8b646686bd200"), uint64(18),
		true, commonv3.SetRateNotSet, true, true, pwi, rb, exchanges, target)

	if err != nil {
		log.Fatal(err)
	}

	// store price
	priceData := common.AllPriceEntry{
		Data: map[uint64]common.OnePrice{
			1: {
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
	err = st.StorePrice(priceData, timepoint)
	if err != nil {
		log.Fatal(err)
	}

	// store rate
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
	timepoint = uint64(1568358532784)
	err = st.StoreRate(rateData, timepoint)
	if err != nil {
		log.Fatal(err)
	}

	app := data.NewReserveData(
		st,  // storage
		nil, // fetcher
		nil, // storageControllerRunner
		nil, // archive
		st,  // globalStorage
		nil, // exchanges
		settingStorage)
	core := core.NewReserveCore(nil, st, nil)
	host := ""
	s = NewHTTPServer(app, core, host, nil, settingStorage)
	s.register()

	ret := m.Run()
	if err = teardown(); err != nil {
		log.Fatal(err)
	}

	os.Exit(ret)
}
