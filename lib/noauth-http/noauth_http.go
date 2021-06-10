package noauthhttp

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

type Client struct {
	cli *http.Client
}

func New() *Client {
	return &Client{
		cli: &http.Client{},
	}
}

func (c *Client) DoReq(url, method string, data interface{}) ([]byte, error) {
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
	_ = rsp.Body.Close()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read body")
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
