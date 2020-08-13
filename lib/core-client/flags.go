package coreclient

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/urfave/cli"
)

const (
	coreEndpointFlag    = "core-endpoint"
	defaultCoreEndpoint = "http://localhost:8000" // suppose to change
)

// NewCoreFlag return flag for core endpoint
func NewCoreFlag() cli.Flag {
	return cli.StringFlag{
		Name:   coreEndpointFlag,
		Usage:  "core endpoint URL",
		EnvVar: "CORE_ENDPOINT",
		Value:  defaultCoreEndpoint,
	}
}

func checkCoreEndpoint(endpoint string) bool {
	if endpoint == "" {
		return true
	}

	resp, err := http.Get(fmt.Sprintf("%s/v3/timeserver", endpoint))
	if err != nil {
		log.Printf("Failed to check time server, error: %s", err.Error())
		return false
	}

	if resp.StatusCode != http.StatusOK {
		return false
	}

	return true
}

// NewCoreEndpointFromContext return core endpoint
func NewCoreEndpointFromContext(c *cli.Context) (string, error) {
	coreEndpoint := c.String(coreEndpointFlag)
	if !checkCoreEndpoint(coreEndpoint) {
		return "", errors.New("core endpoint is required")
	}
	return coreEndpoint, nil
}

// NewCoreClientFromContext return new core client instance
func NewCoreClientFromContext(c *cli.Context) (*Client, error) {
	coreEndpoint := c.String(coreEndpointFlag)
	if !checkCoreEndpoint(coreEndpoint) {
		return nil, errors.New("core endpoint is required")
	}
	if coreEndpoint == "" {
		return nil, nil
	}
	return NewCoreClient(coreEndpoint), nil
}
