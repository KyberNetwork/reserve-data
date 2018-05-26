package httputil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestResponse(t *testing.T) {
	var testValue = "this_is_a_test_value"

	type responseBody struct {
		TestKey string `json:"test_key"`
	}

	testHandler := func(c *gin.Context) {
		ResponseSuccess(c, WithMultipleFields(gin.H{
			"test_key": testValue,
		}))
	}

	r := gin.Default()
	r.GET("/", testHandler)

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("wrong return code, expected: %d, got: %d", http.StatusOK, resp.Code)
	}

	body := &responseBody{}
	if err = json.NewDecoder(resp.Body).Decode(body); err != nil {
		t.Fatal(err)
	}

	if body.TestKey != testValue {
		t.Errorf("wrong value returned, expected: %s, got: %s", testValue, body.TestKey)
	}
}
