package gaspricedataclient

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Data struct {
	Fast     float64 `json:"fast"`
	Standard float64 `json:"standard"`
	Slow     float64 `json:"slow"`
}

type SourceData struct {
	Value     Data  `json:"value"`
	Timestamp int64 `json:"timestamp"`
}

type GasResult map[string]SourceData

type Client interface {
	GetGas() (GasResult, error)
}

type RPCClient struct {
	c   *http.Client
	url string
}

// New create new instance of Client
func New(c *http.Client, url string) Client {
	return &RPCClient{
		c:   c,
		url: url,
	}
}

func (c *RPCClient) GetGas() (GasResult, error) {
	res, err := c.c.Get(c.url)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	_ = res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code %d, %s", res.StatusCode, string(body))
	}
	r := make(GasResult)
	err = json.Unmarshal(body, &r)
	return r, err
}
