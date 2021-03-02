package coreclient

import (
	"fmt"
	"net/http"
	"time"

	ethereum "github.com/ethereum/go-ethereum/common"

	"github.com/KyberNetwork/reserve-data/reservesetting/common"
)

const (
	defaultTimeout = 5 * time.Second
)

// Client is client for core
type Client struct {
	endpoint string
}

// NewCoreClient return new core client instance
func NewCoreClient(endpoint string) *Client {
	return &Client{
		endpoint: endpoint,
	}
}

// CheckTokenIndice check token indice in core
func (c *Client) CheckTokenIndice(tokenAddress ethereum.Address) error {
	// check token indice in core
	endpoint := fmt.Sprintf("%s/v3/check-token-indice?address=%s", c.endpoint, tokenAddress.Hex())
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create new check token indice request: %s", err)
	}
	client := http.Client{
		Timeout: defaultTimeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return common.ErrAssetAddressIsNotIndexInContract
	}
	if resp.StatusCode != http.StatusOK {
		return common.ErrAssetAddressIsNotIndexInContract
	}
	return nil
}

// UpdateTokenIndice call to core to update token indice
func (c *Client) UpdateTokenIndice() error {
	// update token indice in core
	endpoint := fmt.Sprintf("%s/v3/update-token-indice", c.endpoint)
	req, err := http.NewRequest(http.MethodPut, endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create new update token indices request: %s", err)
	}
	client := http.Client{
		Timeout: defaultTimeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to update token indice: %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("update token endpoint failed, status code: %d", resp.StatusCode)
	}
	return nil
}
