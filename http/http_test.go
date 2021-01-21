package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/KyberNetwork/reserve-data/cmd/deployment"
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/testutil"
	"github.com/KyberNetwork/reserve-data/core"
	"github.com/KyberNetwork/reserve-data/data"
	"github.com/KyberNetwork/reserve-data/data/storage"
	"github.com/KyberNetwork/reserve-data/http/httputil"
)

func initData(t *testing.T, s *storage.PostgresStorage) {
	err := s.Record("deposit", common.ActivityID{Timepoint: 1570199521602622015, EID: "USDC|19.2"}, "binance", common.ActivityParams{
		Asset:     3,
		Amount:    19.2,
		Exchange:  1,
		Timepoint: common.NowInMillis(),
	},
		common.ActivityResult{},
		"",
		"",
		common.NowInMillis(),
		true,
		common.NowInMillis()) // default pending status
	assert.NoError(t, err)
}

type activityResponse struct {
	Data    []common.ActivityRecord `json:"data"`
	Success bool
}

func TestGetActivities(t *testing.T) {
	var (
		fromTime = common.TimeToMillis(time.Now().Add(-1 * time.Hour))
		toTime   = common.TimeToMillis(time.Now().Add(1 * time.Hour))
	)

	db, tearDown := testutil.MustNewDevelopmentDB("../cmd/migrations")
	defer func() {
		assert.NoError(t, tearDown())
	}()
	s, err := storage.NewPostgresStorage(db)
	assert.NoError(t, err)
	initData(t, s)

	rData := data.NewReserveData(
		s,   // storage
		nil, // fetcher
		nil, // globalStorage
		nil, // exchanges
		nil, // settingStorage
	)

	rCore := core.NewReserveCore(
		nil,
		s,
		nil,
		nil,
	)

	sv := NewHTTPServer(
		rData,                  // reserve data
		rCore,                  // reserve core
		"",                     // host
		deployment.Development, // deployment mode
		nil,                    // blockchain
		nil,                    // storage
		nil,
		nil,
	)

	sv.register()

	var tests = []httputil.HTTPTestCase{
		{
			Msg:      "test get activities",
			Endpoint: fmt.Sprintf("/v3/activities?fromTime=%d&toTime=%d", fromTime, toTime),
			Method:   http.MethodGet,
			Assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assert.Equal(t, resp.Result().StatusCode, http.StatusOK)

				var activities activityResponse
				log.Printf("%s", resp.Body.Bytes())
				err := json.Unmarshal(resp.Body.Bytes(), &activities)
				assert.NoError(t, err)
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.Msg, func(t *testing.T) { httputil.RunHTTPTestCase(t, tc, sv.r) })
	}

}
