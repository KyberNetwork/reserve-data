package collector

import (
	"time"

	"github.com/KyberNetwork/reserve-data/feed-provider/fetcher"
	"github.com/KyberNetwork/reserve-data/feed-provider/storage"
	"github.com/urfave/cli"
	"go.uber.org/zap"
)

const (
	intervalFlag = "collector-interval"
)

func NewCliFlags() []cli.Flag {
	return []cli.Flag{
		cli.DurationFlag{
			Name:   intervalFlag,
			Usage:  "time between data collection",
			EnvVar: "COLLECTOR_INTERVAL",
			Value:  5 * time.Second,
		},
	}
}

func NewCollectorFromCli(c *cli.Context, sugar *zap.SugaredLogger, s storage.Storage, fetchers map[string]fetcher.Fetcher) *Collector {
	duration := c.Duration(intervalFlag)
	return &Collector{
		sugar:    sugar,
		s:        s,
		fetchers: fetchers,
		duration: duration,
	}
}
