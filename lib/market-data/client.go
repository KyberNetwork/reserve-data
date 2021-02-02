package marketdata

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// Client ...
type Client struct {
	l   *zap.SugaredLogger
	cli *http.Client
	url string
}

// NewClient return market data client
func NewClient(url string) *Client {
	return &Client{
		l:   zap.S(),
		url: url,
		cli: &http.Client{},
	}
}

func (c *Client) doReq(url, method string, data interface{}) ([]byte, error) {
	var (
		httpMethod = strings.ToUpper(method)
		body       io.Reader
	)
	if httpMethod != http.MethodGet && data != nil {
		dataBody, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(dataBody)
	}
	req, err := http.NewRequest(httpMethod, url, body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create get request")
	}
	switch httpMethod {
	case http.MethodPost, http.MethodPut, http.MethodDelete:
		req.Header.Add("Content-Type", "application/json")
	case http.MethodGet:
	default:
		return nil, errors.Errorf("invalid method %s", httpMethod)
	}
	rsp, err := c.cli.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to do get req")
	}
	rspBody, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read body")
	}
	if err := rsp.Body.Close(); err != nil {
		return nil, errors.Wrap(err, "failed to close body")
	}
	if rsp.StatusCode != 200 {
		var rspData interface{}
		if err := json.Unmarshal(rspBody, &rspData); err != nil {
			return nil, errors.Wrap(err, "cannot unmarshal response data")
		}
		return nil, errors.Errorf("receive unexpected code, actual code: %d, data: %+v", rsp.StatusCode, rspData)
	}
	return rspBody, nil
}

// AddFeed ...
func (c *Client) AddFeed(exchange, sourceSymbol, publicSymbol, extID string) error {
	url := fmt.Sprintf("%s/feed/%s", c.url, exchange)
	resp, err := c.doReq(url, http.MethodPost, struct {
		SourceSymbolName string `json:"source_symbol_name"`
		PublicSymbolName string `json:"public_symbol_name"`
		ExtID            string `json:"ext_id"`
	}{
		SourceSymbolName: sourceSymbol,
		PublicSymbolName: publicSymbol,
		ExtID:            extID,
	})
	if err != nil {
		return err
	}
	var respData interface{}
	if err := json.Unmarshal(resp, &respData); err != nil {
		return err
	}
	c.l.Infow("response data", "data", respData)
	return nil
}

// IsValidSymbol ...
func (c *Client) IsValidSymbol(exchange, symbol string) (bool, error) {
	url := fmt.Sprintf("%s/is-valid-symbol?source=%s&symbol=%s", c.url, exchange, symbol)
	resp, err := c.doReq(url, http.MethodGet, nil)
	if err != nil {
		return false, err
	}
	var respData struct {
		IsValid bool `json:"is_valid"`
	}
	if err := json.Unmarshal(resp, &respData); err != nil {
		return false, err
	}
	return respData.IsValid, nil
}
