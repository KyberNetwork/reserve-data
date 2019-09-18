package http

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/testutil"
	"github.com/KyberNetwork/reserve-data/core"
	"github.com/KyberNetwork/reserve-data/data"
	"github.com/KyberNetwork/reserve-data/data/storage"
	"github.com/KyberNetwork/reserve-data/http/httputil"
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
		// pricesEndpoint        = "/prices"
	)

	var tests = []httputil.TestCase{
		{
			Msg:      "get prices version failure",
			Endpoint: fmt.Sprintf("%s?timestamp=%s", pricesVersionEndpoint, ""),
			Method:   http.MethodGet,
			Assert:   httputil.ExpectFailure,
		},
		// TODO: test success case
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
					// TODO: check feed change
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
				Endpoint: getRatesEndpoint,
				Method:   http.MethodGet,
				Assert:   httputil.ExpectFailure,
			},
			{
				Msg:      "test get all rates success",
				Endpoint: getAllRatesEndpoint,
				Method:   http.MethodGet,
				Assert:   httputil.ExpectSuccess,
			},
			// TODO: test success case
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

	// TODO: insert data manual as fetcher does not run

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
