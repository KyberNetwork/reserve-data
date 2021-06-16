package marketdata

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	nh "github.com/KyberNetwork/reserve-data/lib/noauth-http"
)

// Client ...
type Client struct {
	l   *zap.SugaredLogger
	cli *nh.Client
	url string
}

// NewClient return market data client
func NewClient(url string) *Client {
	return &Client{
		l:   zap.S(),
		url: url,
		cli: nh.New(),
	}
}

// AddFeed ...
func (c *Client) AddFeed(exchange, sourceSymbol, publicSymbol, extID string) error {
	url := fmt.Sprintf("%s/feed/%s", c.url, exchange)
	resp, err := c.cli.DoReq(url, http.MethodPost, struct {
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
	resp, err := c.cli.DoReq(url, http.MethodGet, nil)
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
