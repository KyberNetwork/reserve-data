package httputil

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

type responseBody struct {
	Success bool
	Reason  string `json:"reason"`
}

// expectStatus asserts that given response matches the expected status.
func expectStatus(t *testing.T, resp *httptest.ResponseRecorder, success bool, reason string) {
	t.Helper()

	if resp.Code != http.StatusOK {
		t.Fatalf("wrong return code, expected: %d, got: %d", http.StatusOK, resp.Code)
	}
	decoded := &responseBody{}
	if err := json.NewDecoder(resp.Body).Decode(decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Success != success {
		t.Errorf("wrong success status, expected: %t, got: %t %v", success, decoded.Success, decoded.Reason)
	}
	if len(reason) != 0 && reason != decoded.Reason {
		t.Errorf("wrong error msg, expected: %v, got: %v", reason, decoded.Reason)
	}
}

// ExpectSuccess asserts that given response is a success response.
func ExpectSuccess(t *testing.T, resp *httptest.ResponseRecorder) {
	t.Helper()

	expectStatus(t, resp, true, "")
}

// ExpectFailure asserts that given response is a failure response.
func ExpectFailure(t *testing.T, resp *httptest.ResponseRecorder) {
	t.Helper()

	expectStatus(t, resp, false, "")
}

// ExpectFailureWithReason func check response code and its reason
func ExpectFailureWithReason(reason string) func(t *testing.T, resp *httptest.ResponseRecorder) {
	return func(t *testing.T, resp *httptest.ResponseRecorder) {
		t.Helper()
		expectStatus(t, resp, false, reason)
	}
}

type assertFn func(t *testing.T, resp *httptest.ResponseRecorder)

// TestCase is a http test case
type TestCase struct {
	Msg         string
	Endpoint    string
	EndpointExp func() string
	Method      string
	Data        interface{}
	Assert      assertFn
}

// TestHTTPRequest run a test http handler
func TestHTTPRequest(t *testing.T, tc TestCase, handler http.Handler) {
	t.Helper()
	if tc.Endpoint == "" && tc.EndpointExp != nil {
		tc.Endpoint = tc.EndpointExp()
	}
	req, tErr := http.NewRequest(tc.Method, tc.Endpoint, nil)
	if tErr != nil {
		t.Fatal(tErr)
	}

	data, err := json.Marshal(tc.Data)
	if err != nil {
		t.Fatal(err)
	}

	if tc.Data != nil {
		req.Body = ioutil.NopCloser(bytes.NewReader(data))
		req.Header.Add("Content-Type", "application/json")
	}

	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	tc.Assert(t, resp)
}
