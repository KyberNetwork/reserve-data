package http

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KyberNetwork/reserve-data/lib/rtypes"
)

const (
	binance           = rtypes.Binance
	huobi             = rtypes.Huobi
	settingChangePath = "/v3/setting-change-main"
)

type assertFn func(t *testing.T, resp *httptest.ResponseRecorder)

type testCase struct {
	msg         string
	endpoint    string
	endpointExp func() string
	method      string
	data        interface{}
	assert      assertFn
}

func testHTTPRequest(t *testing.T, tc testCase, handler http.Handler) {
	t.Helper()
	if tc.endpoint == "" && tc.endpointExp != nil {
		tc.endpoint = tc.endpointExp()
	}
	req, tErr := http.NewRequest(tc.method, tc.endpoint, nil)
	if tErr != nil {
		t.Fatal(tErr)
	}

	data, err := json.Marshal(tc.data)
	if err != nil {
		t.Fatal(err)
	}

	if tc.data != nil {
		req.Body = ioutil.NopCloser(bytes.NewReader(data))
		req.Header.Add("Content-Type", "application/json")
	}

	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	tc.assert(t, resp)
}
