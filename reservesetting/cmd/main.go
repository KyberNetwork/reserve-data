package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/urfave/cli"
	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data/cmd/configuration"
	"github.com/KyberNetwork/reserve-data/cmd/deployment"
	"github.com/KyberNetwork/reserve-data/cmd/mode"
	v1common "github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/gasstation"
	"github.com/KyberNetwork/reserve-data/common/profiler"
	"github.com/KyberNetwork/reserve-data/exchange"
	"github.com/KyberNetwork/reserve-data/exchange/binance"
	"github.com/KyberNetwork/reserve-data/exchange/huobi"
	libapp "github.com/KyberNetwork/reserve-data/lib/app"
	coreclient "github.com/KyberNetwork/reserve-data/lib/core-client"
	"github.com/KyberNetwork/reserve-data/lib/httputil"
	"github.com/KyberNetwork/reserve-data/lib/migration"
	"github.com/KyberNetwork/reserve-data/lib/rtypes"
	settinghttp "github.com/KyberNetwork/reserve-data/reservesetting/http"
	"github.com/KyberNetwork/reserve-data/reservesetting/storage"
	"github.com/KyberNetwork/reserve-data/reservesetting/storage/postgres"
)

const (
	defaultDB = "reserve_data"

	binanceAPIKeyFlag    = "binance-api-key"
	binanceSecretKeyFlag = "binance-secret-key"
	gasStationKeyFlag    = "gas-station-key"
	etherscanKeyFlag     = "etherscan-key"

	intervalUpdateWithdrawFeeDBFlag      = "interval-update-withdraw-fee-db"
	defaultIntervalUpdateWithdrawFeeDB   = 10 * time.Minute
	intervalUpdateWithdrawFeeLiveFlag    = "interval-update-withdraw-fee-live"
	defaultIntervalUpdateWithdrawFeeLive = 5 * time.Minute
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
	app.Flags = append(app.Flags, libapp.NewSentryFlags()...)
	app.Flags = append(app.Flags, coreclient.NewCoreFlag())
	app.Flags = append(app.Flags, migration.NewMigrationFolderPathFlag())

	app.Flags = append(app.Flags,
		cli.StringFlag{
			Name:   binanceAPIKeyFlag,
			Usage:  "binance api key",
			EnvVar: "BINANCE_API_KEY",
		},
		cli.StringFlag{
			Name:   binanceSecretKeyFlag,
			Usage:  "binance secret key",
			EnvVar: "BINANCE_SECRET_KEY",
		},
		cli.DurationFlag{
			Name:   intervalUpdateWithdrawFeeDBFlag,
			Usage:  "interval update withdraw fee on db",
			Value:  defaultIntervalUpdateWithdrawFeeDB,
			EnvVar: "INTERVAL_UPDATE_WITHDRAW_FEE_DB",
		},
		cli.DurationFlag{
			Name:   intervalUpdateWithdrawFeeLiveFlag,
			Usage:  "interval update withdraw fee live",
			Value:  defaultIntervalUpdateWithdrawFeeLive,
			EnvVar: "INTERVAL_UPDATE_WITHDRAW_FEE_LIVE",
		},
		cli.StringFlag{
			Name:   gasStationKeyFlag,
			Usage:  "gas station api key",
			EnvVar: "GAS_STATION_KEY",
		},
		cli.StringFlag{
			Name:   etherscanKeyFlag,
			Usage:  "etherescan api key",
			EnvVar: "ETHERSCAN_KEY",
		},
	)

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

	if _, err := migration.RunMigrationUp(db.DB, migration.NewMigrationPathFromContext(c), configuration.DatabaseNameFromContext(c)); err != nil {
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

	// QUESTION: should we keep using flag for this config or move it to config file?
	bi := configuration.NewBinanceInterfaceFromContext(c)

	binanceSigner := binance.NewSigner(c.String(binanceAPIKeyFlag), c.String(binanceSecretKeyFlag))
	httpClient := &http.Client{Timeout: time.Second * 30}
	binanceEndpoint := binance.NewBinanceEndpoint(binanceSigner, bi, dpl, httpClient, rtypes.Binance, "", "", "")
	hi := configuration.NewhuobiInterfaceFromContext(c)

	// dummy signer as live infos does not need to sign
	huobiSigner := huobi.NewSigner("", "")
	huobiEndpoint := huobi.NewHuobiEndpoint(huobiSigner, hi, httpClient, "")
	// TODO: binance exchange here must be use public function only.
	liveExchanges, err := getLiveExchanges(enableExchanges, binanceEndpoint, huobiEndpoint, c.Duration(intervalUpdateWithdrawFeeLiveFlag))
	if err != nil {
		return fmt.Errorf("failed to initiate live exchanges: %s", err)
	}

	coreClient, err := coreclient.NewCoreClientFromContext(c)
	if err != nil {
		sugar.Error("core endpoint is not provided, if you create new asset, you cannot update token indice, please provide.")
		return err
	}
	sr, err := postgres.NewStorage(db)
	if err != nil {
		return err
	}

	// run interval update
	if len(enableExchanges) > 0 {
		go func() {
			t := time.NewTicker(c.Duration(intervalUpdateWithdrawFeeDBFlag))
			for range t.C {
				runUpdateWithdrawFee(sugar, liveExchanges, sr)
			}
		}()
	}
	gasStation := gasstation.New(httpClient, c.String(gasStationKeyFlag), c.String(etherscanKeyFlag), sugar)

	sentryDSN := libapp.SentryDSNFromFlag(c)
	server := settinghttp.NewServer(sr, host, liveExchanges, sentryDSN, coreClient, gasStation)
	if profiler.IsEnableProfilerFromContext(c) {
		server.EnableProfiler()
	}
	server.Run()
	return nil
}

func getLiveExchanges(enabledExchanges []rtypes.ExchangeID,
	bi exchange.BinanceInterface,
	hi exchange.HuobiInterface,
	intervalUpdateWithdrawFee time.Duration) (map[rtypes.ExchangeID]v1common.LiveExchange, error) {
	var (
		liveExchanges = make(map[rtypes.ExchangeID]v1common.LiveExchange)
	)
	for _, exchangeID := range enabledExchanges {
		switch exchangeID {
		case rtypes.Binance, rtypes.Binance2:
			binanceLive := exchange.NewBinanceLive(bi)
			go binanceLive.RunUpdateAssetDetails(intervalUpdateWithdrawFee)
			liveExchanges[exchangeID] = binanceLive
		case rtypes.Huobi:
			huobiLive := exchange.NewHuobiLive(hi)
			go huobiLive.RunUpdateAssetDetails(intervalUpdateWithdrawFee)
			liveExchanges[exchangeID] = huobiLive
		}
	}
	return liveExchanges, nil
}

func runUpdateWithdrawFee(sugar *zap.SugaredLogger, exchanges map[rtypes.ExchangeID]v1common.LiveExchange, s storage.Interface) {
	assets, err := s.GetAssets()
	if err != nil {
		sugar.Errorw("cannot get assets", "err", err)
		return
	}
	for _, asset := range assets {
		for _, ae := range asset.Exchanges {
			if le, ok := exchanges[ae.ExchangeID]; ok {
				withdrawFee, err := le.GetLiveWithdrawFee(ae.Symbol)
				if err != nil {
					sugar.Warnw("cannot get live withdraw fee", "err", err, "exchange symbol", ae.Symbol, "token symbol", asset.Symbol)
					continue
				}
				if err := s.UpdateAssetExchangeWithdrawFee(withdrawFee, ae.ID); err != nil {
					sugar.Errorw("cannot update asset exchange", "err", err)
					return
				}
			}
		}
	}
}
