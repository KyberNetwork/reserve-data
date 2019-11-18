package common

import (
	"github.com/ethereum/go-ethereum/ethclient"
)

// EthClient just a wrap for ethclient.Client with url connect to
type EthClient struct {
	*ethclient.Client
	URL string
}

// NewEthClient create a new ethclient
func NewEthClient(url string) (*EthClient, error) {
	c, err := ethclient.Dial(url)
	if err != nil {
		return nil, err
	}
	return &EthClient{
		Client: c,
		URL:    url,
	}, nil
}
