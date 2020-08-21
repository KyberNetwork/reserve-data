package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/KyberNetwork/reserve-data/lib/rtypes"
	"github.com/KyberNetwork/reserve-data/reservesetting/common"
)

type apiClient struct {
	s *Server
}

type commonResponse struct {
	Success bool   `json:"success"`
	Reason  string `json:"reason"`
}

type postResponse struct {
	commonResponse
	ID uint64 `json:"id"`
}

type assetResponse struct {
	commonResponse
	Asset common.Asset `json:"data"`
}

type assetsResponse struct {
	commonResponse
	Assets []common.Asset `json:"data"`
}

type exchangeResponse struct {
	commonResponse
	Exchange common.Exchange `json:"data"`
}

type exchangesResponse struct {
	commonResponse
	Exchanges []common.Exchange `json:"data"`
}

func (c *apiClient) getAsset(id rtypes.AssetID) (assetResponse, error) {
	req, err := createRequest(http.MethodGet, fmt.Sprintf("/v3/asset/%d", id), nil)
	if err != nil {
		return assetResponse{}, err
	}
	httpResp := httptest.NewRecorder()
	c.s.r.ServeHTTP(httpResp, req)
	if httpResp.Code != http.StatusOK {
		return assetResponse{}, fmt.Errorf("server return %d - %s", httpResp.Code, httpResp.Body.String())
	}
	asset := assetResponse{}
	err = readResponse(httpResp.Body, &asset)
	if err != nil {
		return assetResponse{}, err
	}
	return asset, nil
}

func (c *apiClient) getAssets() (assetsResponse, error) {
	req, err := createRequest(http.MethodGet, "/v3/asset", nil)
	if err != nil {
		return assetsResponse{}, err
	}
	httpResp := httptest.NewRecorder()
	c.s.r.ServeHTTP(httpResp, req)
	if httpResp.Code != http.StatusOK {
		return assetsResponse{}, fmt.Errorf("server return %d - %s", httpResp.Code, httpResp.Body.String())
	}
	var assets assetsResponse
	err = readResponse(httpResp.Body, &assets)
	if err != nil {
		return assetsResponse{}, err
	}
	return assets, nil
}

func (c *apiClient) createSettingChange(change common.SettingChange) (postResponse, error) {
	req, err := createRequest(http.MethodPost, settingChangePath, change)
	if err != nil {
		return postResponse{}, err
	}
	resp := httptest.NewRecorder()
	c.s.r.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		return postResponse{}, fmt.Errorf("server return %d - %s", resp.Code, resp.Body.String())
	}
	var s postResponse
	err = readResponse(resp.Body, &s)
	return s, err
}

func (c *apiClient) confirmSettingChange(id uint64) (commonResponse, error) {
	req, err := createRequest(http.MethodPut, fmt.Sprintf("%s/%d", settingChangePath, id), nil)
	if err != nil {
		return commonResponse{}, err
	}
	resp := httptest.NewRecorder()
	c.s.r.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		return commonResponse{}, fmt.Errorf("server return %d - %s", resp.Code, resp.Body.String())
	}
	var status commonResponse
	err = readResponse(resp.Body, &status)
	return status, err
}

func (c *apiClient) getExchange(id rtypes.ExchangeID) (exchangeResponse, error) {
	req, err := createRequest(http.MethodGet, fmt.Sprintf("/v3/exchange/%d", id), nil)
	if err != nil {
		return exchangeResponse{}, err
	}
	resp := httptest.NewRecorder()
	c.s.r.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		return exchangeResponse{}, fmt.Errorf("server return %d - %s", resp.Code, resp.Body.String())
	}
	var status exchangeResponse
	err = readResponse(resp.Body, &status)
	return status, err
}

func (c *apiClient) getExchanges() (exchangesResponse, error) {
	req, err := createRequest(http.MethodGet, "/v3/exchange", nil)
	if err != nil {
		return exchangesResponse{}, err
	}
	resp := httptest.NewRecorder()
	c.s.r.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		return exchangesResponse{}, fmt.Errorf("server return %d - %s", resp.Code, resp.Body.String())
	}
	var status exchangesResponse
	err = readResponse(resp.Body, &status)
	return status, err
}

// ReadResponse retry to parse response into object
func readResponse(data io.Reader, dataField interface{}) error {
	return json.NewDecoder(data).Decode(dataField)
}

func createRequest(method, url string, data interface{}) (*http.Request, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return http.NewRequest(method, url, bytes.NewBuffer(body))
}

func (c *apiClient) updateExchangeStatus(id rtypes.ExchangeID, exs exchangeEnabledEntry) (commonResponse, error) {
	req, err := createRequest(http.MethodPut, fmt.Sprintf("/v3/set-exchange-enabled/%d", id), exs)
	if err != nil {
		return commonResponse{}, err
	}
	resp := httptest.NewRecorder()
	c.s.r.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		return commonResponse{}, fmt.Errorf("server return %d - %s", resp.Code, resp.Body.String())
	}
	var status commonResponse
	err = readResponse(resp.Body, &status)
	return status, err
}
