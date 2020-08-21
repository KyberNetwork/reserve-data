package configuration

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/lib/rtypes"
)

const (
	exchangesFlag = "exchanges"
)

// NewExchangeCliFlag creates new cli flag for enable/disable exchanges.
func NewExchangeCliFlag() cli.Flag {
	return cli.StringSliceFlag{
		Name:   exchangesFlag,
		Usage:  "Enable an exchange for fetching order books, rebalancing purpose. By default all exchanges are disabled.",
		EnvVar: "KYBER_EXCHANGES",
	}
}

// NewExchangesFromContext returns configured exchanges from cli context.
func NewExchangesFromContext(c *cli.Context) ([]rtypes.ExchangeID, error) {
	var exchanges []rtypes.ExchangeID

	for _, exchangeName := range c.GlobalStringSlice(exchangesFlag) {
		exchange, ok := common.ValidExchangeNames[exchangeName]
		if !ok {
			return nil, fmt.Errorf("invalid exchange %v", exchangeName)
		}
		exchanges = append(exchanges, exchange)
	}

	return exchanges, nil
}
