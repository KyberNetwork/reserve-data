package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/robfig/cron"
	"github.com/urfave/cli"
	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data/blockchain"
	"github.com/KyberNetwork/reserve-data/cmd/configuration"
	"github.com/KyberNetwork/reserve-data/cmd/deployment"
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/profiler"
	"github.com/KyberNetwork/reserve-data/exchange/binance"
	apphttp "github.com/KyberNetwork/reserve-data/http"
	"github.com/KyberNetwork/reserve-data/lib/app"
	authhttp "github.com/KyberNetwork/reserve-data/lib/auth-http"
	"github.com/KyberNetwork/reserve-data/lib/migration"
	"github.com/KyberNetwork/reserve-data/lib/rtypes"
	"github.com/KyberNetwork/reserve-data/reservesetting/storage/postgres"
)

func main() {
	app := cli.NewApp()
	app.Name = "Reserve Core"
	app.Usage = "Kyber Reserve core component that helps manage reserves of tokens"
	app.Version = "0.11.0"
	app.Action = run

	app.Flags = configuration.NewCliFlags()
	app.Flags = append(app.Flags, profiler.NewCliFlags()...)

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func initEthClient(rc common.RawConfig) (*common.EthClient, []*common.EthClient, error) {
	mainNode, err := common.NewEthClient(rc.Nodes.Main)
	if err != nil {
		return nil, nil, err
	}
	bks := make([]*common.EthClient, 0, len(rc.Nodes.Backup))
	for _, v := range rc.Nodes.Backup {
		bkNode, err := common.NewEthClient(v)
		if err != nil {
			return nil, nil, fmt.Errorf("connect backup node %s error %+v", v, err)
		}
		bks = append(bks, bkNode)
	}
	return mainNode, bks, nil
}
func toJSONString(i interface{}) string {
	data, _ := json.Marshal(i)
	return string(data)
}
func run(c *cli.Context) error {
	configuration.SetupLogging()

	dpl, err := deployment.NewDeploymentFromContext(c)
	if err != nil {
		return err
	}
	l, flusher, err := app.NewSugaredLogger(c)
	if err != nil {
		panic(err)
	}
	defer func() {
		flusher()
	}()
	zap.ReplaceGlobals(l.Desugar())

	configFile, secretConfigFile := configuration.NewConfigFilesFromContext(c)

	rcf := common.RawConfig{}
	if err := loadConfigFromFile(configFile, &rcf); err != nil {
		l.Errorw("error load config file", "error", err)
		return err
	}
	if err := loadConfigFromFile(secretConfigFile, &rcf); err != nil {
		l.Errorw("error load config file", "error", err)
		return err
	}
	l.Debugw("start with config",
		"world_endpoints", toJSONString(rcf.WorldEndpoints),
		"contract_address", toJSONString(rcf.ContractAddresses),
		"exchange_endpoints", toJSONString(rcf.ExchangeEndpoints),
		"nodes", toJSONString(rcf.Nodes),
		"fetch_delay", toJSONString(rcf.FetcherDelay),
		"gas_config", toJSONString(rcf.GasConfig))

	rcf.MigrationPath = migration.NewMigrationPathFromContext(c)
	httpClient := &http.Client{}

	mainNode, backupNodes, err := initEthClient(rcf)
	if err != nil {
		l.Panicw("failed to init eth client", "err", err)
	}
	kyberNetworkProxy, err := blockchain.NewNetworkProxy(rcf.ContractAddresses.Proxy,
		mainNode.Client)
	if err != nil {
		log.Panicf("cannot create network proxy client, err %+v", err)
	}

	store, err := createStorage(c, rcf.MigrationPath)
	if err != nil {
		l.Errorw("failed to create storage", "error", err)
		return err
	}
	if err = schedulePartition(store); err != nil {
		l.Errorw("failed to schedule partition", "error", err)
		return err
	}

	conf, err := configuration.NewConfigurationFromContext(c, rcf, store, mainNode, backupNodes)
	if err != nil {
		l.Errorw("failed to new configuration from context", "error", err)
		return err
	}

	bc, err := configuration.CreateBlockchain(conf)
	if err != nil {
		l.Errorw("Can not create blockchain", "err", err)
		return err
	}

	dryRun := configuration.NewDryRunFromContext(c)

	rData, rCore, gasInfo := configuration.CreateDataCore(conf, dpl, bc, kyberNetworkProxy, rcf, httpClient)
	if !dryRun {
		if err = rData.Run(); err != nil {
			l.Errorw("failed to run data service", "err", err)
			return err
		}
	}

	for _, ex := range conf.Exchanges {
		common.SupportedExchanges[ex.ID()] = ex
	}

	binanceMainClient := binance.NewBinanceEndpoint(binance.NewSigner("", ""),
		binance.NewRealInterface(rcf.ExchangeEndpoints.Binance.URL), deployment.Production, httpClient, rtypes.Binance,
		rcf.MarketDataBaseURL, rcf.AccountData.BaseURL, rcf.BinanceAccountMainID,
		authhttp.NewAuthHTTP(rcf.AccountData.AccessKey, rcf.AccountData.AccessSecret))

	host := rcf.HTTPAPIAddr
	server := apphttp.NewHTTPServer(
		rData, rCore,
		host,
		dpl,
		bc,
		conf.SettingStorage,
		gasInfo,
		binanceMainClient,
	)
	if profiler.IsEnableProfilerFromContext(c) {
		server.EnableProfiler()
	}

	if !dryRun {
		server.Run()
	} else {
		l.Infow("Dry run finished. All configs are corrected")
	}

	return err
}

func loadConfigFromFile(path string, rcf *common.RawConfig) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, rcf)
}

func createStorage(c *cli.Context, path string) (*postgres.Storage, error) {
	db, err := configuration.NewDBFromContext(c)
	if err != nil {
		return nil, err
	}
	dbName := configuration.DatabaseNameFromContext(c)

	if _, err := migration.RunMigrationUp(db.DB, path, dbName); err != nil {
		return nil, err
	}

	// as this is core connect to setting db, the core endpoint is not needed
	sr, err := postgres.NewStorage(db)
	if err != nil {
		return nil, err
	}
	return sr, nil
}

func schedulePartition(store *postgres.Storage) error {
	err := store.MakeFetchDataTablePartition() // run immediately one shot
	if err != nil {
		return err
	}
	if ex := store.PrepareOrderBookTablePartition(); ex != nil {
		zap.S().Fatalw("failed to prepare order book partition", "err", err)
		return err
	}
	mc := cron.New()
	err = mc.AddFunc("0 0 * * *", func() { // check everyday is good enough
		if ex := store.MakeFetchDataTablePartition(); ex != nil {
			zap.S().Fatalw("failed to prepare partition", "err", err)
		}
		if ex := store.PrepareOrderBookTablePartition(); ex != nil {
			zap.S().Fatalw("failed to prepare order book partition", "err", err)
		}
	})
	if err != nil {
		return err
	}
	mc.Start()
	return nil
}
