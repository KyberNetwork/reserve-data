package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/urfave/cli"
	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data/blockchain"
	"github.com/KyberNetwork/reserve-data/cmd/configuration"
	"github.com/KyberNetwork/reserve-data/cmd/deployment"
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/gasstation"
	"github.com/KyberNetwork/reserve-data/common/profiler"
	"github.com/KyberNetwork/reserve-data/core"
	apphttp "github.com/KyberNetwork/reserve-data/http"
	"github.com/KyberNetwork/reserve-data/lib/app"
	"github.com/KyberNetwork/reserve-data/lib/migration"
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
		return err
	}
	if err := loadConfigFromFile(secretConfigFile, &rcf); err != nil {
		return err
	}

	rcf.MigrationPath = migration.NewMigrationPathFromContext(c)

	mainNode, backupNodes, err := initEthClient(rcf)
	if err != nil {
		l.Panicw("failed to init eth client", "err", err)
	}
	kyberNetworkProxy, err := blockchain.NewNetworkProxy(rcf.ContractAddresses.Proxy,
		mainNode.Client)
	if err != nil {
		log.Panicf("cannot create network proxy client, err %+v", err)
	}
	gasPriceLimiter := core.NewNetworkGasPriceLimiter(kyberNetworkProxy, rcf.GasConfig.FetchMaxGasCacheSeconds)
	gasstationClient := gasstation.New(&http.Client{}, rcf.GasConfig.GasStationAPIKey, rcf.EtherscanAPIKey, l)

	conf, err := configuration.NewConfigurationFromContext(c, rcf, l, mainNode, backupNodes, gasstationClient, gasPriceLimiter)
	if err != nil {
		return err
	}

	bc, err := configuration.CreateBlockchain(conf)
	if err != nil {
		l.Errorw("Can not create blockchain", "err", err)
		return err
	}

	dryRun := configuration.NewDryRunFromContext(c)

	rData, rCore := configuration.CreateDataCore(conf, dpl, bc, l, gasPriceLimiter)
	if !dryRun {
		if dpl != deployment.Simulation {
			if err = rData.RunStorageController(); err != nil {
				l.Errorw("failed to run storage controller", "err", err)
				return err
			}
		}
		if err = rData.Run(); err != nil {
			l.Errorw("failed to run data service", "err", err)
			return err
		}
	}

	for _, ex := range conf.Exchanges {
		common.SupportedExchanges[ex.ID()] = ex
	}

	host := rcf.HTTPAPIAddr
	server := apphttp.NewHTTPServer(
		rData, rCore,
		host,
		dpl,
		bc,
		conf.SettingStorage,
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
