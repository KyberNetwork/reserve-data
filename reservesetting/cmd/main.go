package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli"
	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data/cmd/configuration"
	"github.com/KyberNetwork/reserve-data/cmd/deployment"
	"github.com/KyberNetwork/reserve-data/cmd/mode"
	v1common "github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/profiler"
	"github.com/KyberNetwork/reserve-data/exchange"
	"github.com/KyberNetwork/reserve-data/exchange/binance"
	"github.com/KyberNetwork/reserve-data/exchange/huobi"
	libapp "github.com/KyberNetwork/reserve-data/lib/app"
	"github.com/KyberNetwork/reserve-data/lib/httputil"
	"github.com/KyberNetwork/reserve-data/reservesetting/blockchain"
	"github.com/KyberNetwork/reserve-data/reservesetting/http"
	"github.com/KyberNetwork/reserve-data/reservesetting/storage/postgres"
)

const (
	defaultDB        = "reserve_data"
	coreEndpointFlag = "core-endpoint"
)

func main() {
	app := cli.NewApp()
	app.Name = "HTTP gateway for reserve core"
	app.Action = run
	app.Flags = append(app.Flags, mode.NewCliFlag())
	app.Flags = append(app.Flags, deployment.NewCliFlag())
	app.Flags = append(app.Flags, configuration.NewBinanceCliFlags()...)
	app.Flags = append(app.Flags, configuration.NewHuobiCliFlags()...)
	app.Flags = append(app.Flags, configuration.NewPostgreSQLFlags(defaultDB)...)
	app.Flags = append(app.Flags, httputil.NewHTTPCliFlags(httputil.V3ServicePort)...)
	app.Flags = append(app.Flags, configuration.NewExchangeCliFlag())
	app.Flags = append(app.Flags, profiler.NewCliFlags()...)
	app.Flags = append(app.Flags, blockchain.NewWrapperAddressFlag()...)
	app.Flags = append(app.Flags, blockchain.NewEthereumNodeFlags())
	app.Flags = append(app.Flags, libapp.NewSentryFlags()...)
	app.Flags = append(app.Flags, cli.StringFlag{
		Name:   coreEndpointFlag,
		Usage:  "core endpoint URL",
		EnvVar: "CORE_ENDPOINT",
	})

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(c *cli.Context) error {
	sugar, flusher, err := libapp.NewSugaredLogger(c)
	if err != nil {
		return err
	}
	defer func() {
		flusher()
	}()
	zap.ReplaceGlobals(sugar.Desugar())

	host := httputil.NewHTTPAddressFromContext(c)
	db, err := configuration.NewDBFromContext(c)
	if err != nil {
		return err
	}

	dpl, err := deployment.NewDeploymentFromContext(c)
	if err != nil {
		return err
	}

	enableExchanges, err := configuration.NewExchangesFromContext(c)
	if err != nil {
		return fmt.Errorf("failed to get enabled exchanges: %s", err)
	}

	bi := configuration.NewBinanceInterfaceFromContext(c)
	// dummy signer as live infos does not need to sign
	binanceSigner := binance.NewSigner("", "")
	binanceEndpoint := binance.NewBinanceEndpoint(binanceSigner, bi, dpl)
	hi := configuration.NewhuobiInterfaceFromContext(c)

	// dummy signer as live infos does not need to sign
	huobiSigner := huobi.NewSigner("", "")
	huobiEndpoint := huobi.NewHuobiEndpoint(huobiSigner, hi)

	liveExchanges, err := getLiveExchanges(enableExchanges, binanceEndpoint, huobiEndpoint)
	if err != nil {
		return fmt.Errorf("failed to initiate live exchanges: %s", err)
	}

	coreEndpoint := c.String(coreEndpointFlag)
	if coreEndpoint == "" {
		sugar.Error("core endpoint is not provided, if you create new asset, you cannot update token indice, please provide.")
		return errors.New("core endpoint is required")
	}
	sr, err := postgres.NewStorage(db, coreEndpoint)
	if err != nil {
		return err
	}

	wrapperAddress, err := blockchain.NewWrapperAddressFromContext(c)
	if err != nil {
		return err
	}

	rateAddress, err := blockchain.NewRateAddressFromContext(c)
	if err != nil {
		return err
	}

	ethClient, err := blockchain.NewEthereumClientFromFlag(c)
	if err != nil {
		return err
	}

	newBlockchain, err := blockchain.NewBlockchain(wrapperAddress, rateAddress, ethClient)
	if err != nil {
		return err
	}
	sentryDSN := libapp.SentryDSNFromFlag(c)
	server := http.NewServer(sr, host, liveExchanges, newBlockchain, sentryDSN)
	if profiler.IsEnableProfilerFromContext(c) {
		server.EnableProfiler()
	}
	server.Run()
	return nil
}

func getLiveExchanges(enabledExchanges []v1common.ExchangeID, bi exchange.BinanceInterface, hi exchange.HuobiInterface) (map[v1common.ExchangeID]v1common.LiveExchange, error) {
	var (
		liveExchanges = make(map[v1common.ExchangeID]v1common.LiveExchange)
	)
	for _, exchangeID := range enabledExchanges {
		switch exchangeID {
		case v1common.Binance:
			binanceLive := exchange.NewBinanceLive(bi)
			liveExchanges[v1common.Binance] = binanceLive
		case v1common.Huobi:
			huobiLive := exchange.NewHuobiLive(hi)
			liveExchanges[v1common.Huobi] = huobiLive
		}
	}
	return liveExchanges, nil
}
