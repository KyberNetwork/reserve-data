package main

import (
	"github.com/KyberNetwork/reserve-data/cmd/deployment"
	"github.com/KyberNetwork/reserve-data/cmd/mode"
	"github.com/KyberNetwork/reserve-data/feed-provider/collector"
	"github.com/KyberNetwork/reserve-data/feed-provider/fetcher/coinbase"
	"github.com/KyberNetwork/reserve-data/lib/app"
	"github.com/KyberNetwork/reserve-data/lib/httputil"
	"github.com/urfave/cli"
)

func NewCliFlags() []cli.Flag {
	var flags []cli.Flag
	flags = append(flags, mode.NewCliFlag(), deployment.NewCliFlag())
	flags = append(flags, app.NewSentryFlags()...)
	flags = append(flags, coinbase.NewCliFlags()...)
	flags = append(flags, collector.NewCliFlags()...)
	flags = append(flags, httputil.NewHTTPCliFlags(httputil.FeedProviderPort)...)
	return flags
}
