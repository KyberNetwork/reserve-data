package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var onePermission = []Permission{ReadOnly}
var twoPermission = []Permission{ReadOnly, ReadAndWrite}
var oneSecretKey = SecrectKeys{
	ReadOnly: "123",
}
var twoSecretKey = SecrectKeys{
	ReadOnly:     "123",
	ReadAndWrite: "345",
}

// ValidateNonceTrue interfaces always return True
type ValidateNonceTrue struct{}

// IsValid always return True
func (v ValidateNonceTrue) IsValid(nonce int64) bool {
	return true
}

func setupRouter(permissions []Permission, secrets SecrectKeys, v ValidateNonce) (*gin.Engine, error) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	m, err := Authenticated(permissions, secrets, v)
	if err != nil {
		return nil, err
	}
	r.Use(m)
	r.GET("/", testGET)
	r.POST("/", testPOST)
	return r, nil
}

func TestErrorSetupRouter(t *testing.T) {
	_, err := setupRouter(twoPermission, oneSecretKey, NewValidateNonceByTime())
	assert.Equal(t, err, errAuthenticatedPermission)
}

func TestNoNonce(t *testing.T) {
	r, err := setupRouter(onePermission, oneSecretKey, NewValidateNonceByTime())
	if err != nil {
		t.Error("Error while setup router", err.Error())
		return
	}
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestNonceNotInt(t *testing.T) {
	r, err := setupRouter(onePermission, oneSecretKey, NewValidateNonceByTime())
	if err != nil {
		t.Error("Error while setup router", err.Error())
		return
	}
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/?nonce=abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
func TestNonceNotInRange(t *testing.T) {
	r, err := setupRouter(onePermission, oneSecretKey, NewValidateNonceByTime())
	if err != nil {
		t.Error("Error while setup router", err.Error())
		return
	}
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/?nonce=1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestNoSignTokenInHeader(t *testing.T) {
	r, err := setupRouter(onePermission, oneSecretKey, NewValidateNonceByTime())
	if err != nil {
		t.Error("Error while setup router", err.Error())
		return
	}
	w := httptest.NewRecorder()
	nonce := strconv.FormatUint(getTimepoint(), 10)
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	form, _ := url.ParseQuery(req.URL.RawQuery)
	form.Add("nonce", nonce)
	req.URL.RawQuery = form.Encode()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
func TestSignNonceInQuery(t *testing.T) {
	r, err := setupRouter(onePermission, oneSecretKey, ValidateNonceTrue{})
	if err != nil {
		t.Error("Error while setup router", err.Error())
		return
	}
	w := httptest.NewRecorder()
	const (
		nonce  = "1234"
		signed = "9390b9d8030bb6d26d008d9480d9e268540a3fcaa5db3904323491c87eebe7e3772cec9ac11b364db17866b45c2ec9df1b4e9ac0e6096af55d8ddbe3214ec567"
	)
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("signed", signed)
	form, _ := url.ParseQuery(req.URL.RawQuery)
	form.Add("nonce", nonce)
	req.URL.RawQuery = form.Encode()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSignNonceBody(t *testing.T) {
	r, err := setupRouter(onePermission, oneSecretKey, ValidateNonceTrue{})
	if err != nil {
		t.Error("Error while setup router", err.Error())
		return
	}
	w := httptest.NewRecorder()
	const (
		nonce  = "1234"
		signed = "9390b9d8030bb6d26d008d9480d9e268540a3fcaa5db3904323491c87eebe7e3772cec9ac11b364db17866b45c2ec9df1b4e9ac0e6096af55d8ddbe3214ec567"
	)
	data := url.Values{}
	data.Set("nonce", nonce)
	req, _ := http.NewRequest("POST", "/", bytes.NewBufferString(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("signed", signed)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSignNonceBodyAndParams(t *testing.T) {
	r, err := setupRouter(onePermission, oneSecretKey, ValidateNonceTrue{})
	if err != nil {
		t.Error("Error while setup router", err.Error())
		return
	}
	w := httptest.NewRecorder()
	const (
		nonce = "1234"
		// a=a&b=b&nonce=1234
		signed = "8ab47f69ef58474de7e5b66c6e8e205ea628283874793c08be32e9577a3d67ce2f26d4eb30e2895f5b93d61c7f3f27ae1dc2f6562bb1a5cd48eee39b25df0147"
	)
	data := url.Values{}
	data.Set("nonce", nonce)
	data.Set("a", "a")
	req, _ := http.NewRequest("POST", "/", bytes.NewBufferString(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("signed", signed)
	urlRawQuery := url.Values{}
	urlRawQuery.Add("b", "b")
	req.URL.RawQuery = urlRawQuery.Encode()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func testGET(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

func testPOST(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}
