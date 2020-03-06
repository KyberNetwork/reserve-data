package main

import (
	"log"
	"os"

	"github.com/urfave/cli"
	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data/feed-provider/collector"
	"github.com/KyberNetwork/reserve-data/feed-provider/fetcher"
	"github.com/KyberNetwork/reserve-data/feed-provider/fetcher/coinbase"
	"github.com/KyberNetwork/reserve-data/feed-provider/httpserver"
	"github.com/KyberNetwork/reserve-data/feed-provider/storage"
	"github.com/KyberNetwork/reserve-data/lib/app"
	"github.com/KyberNetwork/reserve-data/lib/httputil"
)

func main() {
	app := cli.NewApp()
	app.Name = "Feed provider"
	app.Usage = "Calculate and convert feed data to keep the same format"
	app.Version = "0.0.1"
	app.Action = run

	app.Flags = NewCliFlags()

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(c *cli.Context) error {
	l, flusher, err := app.NewSugaredLogger(c)
	if err != nil {
		panic(err)
	}
	defer func() {
		flusher()
	}()
	zap.ReplaceGlobals(l.Desugar())
	coinbaseETHDAIfetcher, err := coinbase.NewFetcherFromCli(c, l)
	if err != nil {
		return err
	}
	fetchers := map[string]fetcher.Fetcher{
		"CoinbaseETHDAI": coinbaseETHDAIfetcher,
	}
	s := storage.NewRAMStorage()
	host := httputil.NewHTTPAddressFromContext(c)
	server, err := httpserver.NewHTTPServer(l, s, host)
	if err != nil {
		return err
	}
	collector := collector.NewCollectorFromCli(c, l, s, fetchers)
	go collector.Run()
	return server.Run()
}
