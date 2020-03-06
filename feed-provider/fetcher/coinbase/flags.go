package coinbase

import (
	"net/http"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/urfave/cli"
	"go.uber.org/zap"
)

const (
	coinbaseDAIETHFlag      = "coibase-daieth-url"
	defaultCoinbaseDAIETH   = "https://api.pro.coinbase.com/products/eth-dai/book?level=2"
	daiRequiredDepthFlag    = "coinbase-dai-required-depth"
	defaultDaiRequiredDepth = 10000
)

func NewCliFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:   coinbaseDAIETHFlag,
			Usage:  "endpoint to get dai-eth on coinbase",
			EnvVar: "COINBASE_DAIETH_URL",
			Value:  defaultCoinbaseDAIETH,
		},
		cli.Float64Flag{
			Name:   daiRequiredDepthFlag,
			Usage:  "required amount of dai to calculate on each size of orderbook",
			EnvVar: "COINBASE_REQUIRE_DAI",
			Value:  defaultDaiRequiredDepth,
		},
	}
}

func NewFetcherFromCli(c *cli.Context, sugar *zap.SugaredLogger) (*Fetcher, error) {
	url := c.String(coinbaseDAIETHFlag)
	err := validation.Validate(url, validation.Required, is.URL)
	requireAmount := c.Float64(daiRequiredDepthFlag)
	if err != nil {
		return nil, err
	}
	return &Fetcher{
		sugar:         sugar,
		endpoint:      url,
		requireAmount: requireAmount,
		client: &http.Client{
			Timeout: time.Minute,
		},
	}, nil
}
