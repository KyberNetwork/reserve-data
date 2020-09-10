package cmd

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data/blockchain"
	"github.com/KyberNetwork/reserve-data/cmd/configuration"
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/blockchain/nonce"
	"github.com/KyberNetwork/reserve-data/common/config"
	"github.com/KyberNetwork/reserve-data/common/gasinfo"
	"github.com/KyberNetwork/reserve-data/common/gaspricedata-client"
	"github.com/KyberNetwork/reserve-data/core"
	"github.com/KyberNetwork/reserve-data/data"
	"github.com/KyberNetwork/reserve-data/data/fetcher"
	apphttp "github.com/KyberNetwork/reserve-data/http"
)

const (
	remoteLogPath = "core-log"
)

var (
	// logDir is located at base of this repository.
	logDir         = filepath.Join(filepath.Dir(filepath.Dir(common.CurrentDir())), "log")
	noAuthEnable   bool
	stdoutLog      bool
	dryRun         bool
	profilerPrefix string

	sentryDSN   string
	sentryLevel string
	zapMode     string
	configFile  string
	secretFile  string
)

func initEthClient(ac config.AppConfig) (*common.EthClient, []*common.EthClient, error) {
	mainNode, err := common.NewEthClient(ac.Node.Main)
	if err != nil {
		return nil, nil, err
	}
	bks := make([]*common.EthClient, 0, len(ac.Node.Backup))
	for _, v := range ac.Node.Backup {
		bkNode, err := common.NewEthClient(v)
		if err != nil {
			return nil, nil, fmt.Errorf("connect backup node %s error %+v", v, err)
		}
		bks = append(bks, bkNode)
	}
	return mainNode, bks, nil
}

func serverStart(_ *cobra.Command, _ []string) {
	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU)

	w := configLog(stdoutLog)
	s, f, err := newSugaredLogger(w)
	if err != nil {
		panic(err)
	}
	defer f()
	zap.ReplaceGlobals(s.Desugar())
	kyberENV := common.RunningMode()
	var ac = config.DefaultAppConfig()
	if err = config.LoadConfig(configFile, &ac); err != nil {
		s.Panicw("failed to load config file", "err", err, "file", configFile)
	}
	if err = config.LoadConfig(secretFile, &ac); err != nil {
		s.Panicw("failed to load secret file", "err", err, "file", secretFile)
	}
	httpClient := &http.Client{}

	mainNode, backupNodes, err := initEthClient(ac)
	if err != nil {
		s.Panicw("failed to init eth client", "err", err)
	}
	kyberNetworkProxy, err := blockchain.NewNetworkProxy(ac.ContractAddresses.Proxy,
		mainNode.Client)
	if err != nil {
		log.Panicf("cannot create network proxy client, err %+v", err)
	}

	appState := configuration.InitAppState(!noAuthEnable, ac, mainNode, backupNodes)
	// backup other log daily
	backupLog(appState.Archive, "@daily", "core.+\\.log")
	// backup core.log every 2 hour
	backupLog(appState.Archive, "@every 2h", "core\\.log")

	bc, err := CreateBlockchain(appState)
	if err != nil {
		log.Panicf("Can not create blockchain: (%s)", err)
	}
	rData, gasInfo, rCore := initReserveComponents(appState, kyberENV, bc, kyberNetworkProxy, ac, httpClient)
	gasinfo.SetGlobal(gasInfo)
	if !dryRun {
		if kyberENV != common.SimulationMode {
			if err = rData.RunStorageController(); err != nil {
				log.Panic(err)
			}
		}
		if err = rData.Run(); err != nil {
			log.Panic(err)
		}
	}
	for _, ex := range appState.Exchanges {
		common.SupportedExchanges[ex.ID()] = ex
	}

	server := apphttp.NewHTTPServer(
		rData, rCore,
		appState.MetricStorage,
		ac.HTTPAPIAddr,
		appState.EnableAuthentication,
		profilerPrefix,
		appState.AuthEngine,
		kyberENV,
		bc, appState.Setting,
		gasInfo,
	)

	if !dryRun {
		server.Run()
	} else {
		s.Infof("Dry run finished. All configs are corrected")
	}
}

func initReserveComponents(appState *configuration.AppState, kyberENV string, bc *blockchain.Blockchain,
	kyberNetworkProxy *blockchain.NetworkProxy, ac config.AppConfig, httpClient *http.Client) (*data.ReserveData, *gasinfo.GasPriceInfo, *core.ReserveCore) {
	// get fetcher based on config and ENV == simulation.
	dataFetcher := fetcher.NewFetcher(
		appState.FetcherStorage,
		appState.FetcherGlobalStorage,
		appState.World,
		appState.FetcherRunner,
		kyberENV == common.SimulationMode,
		appState.Setting,
	)
	for _, ex := range appState.FetcherExchanges {
		dataFetcher.AddExchange(ex)
	}
	nonceCorpus := nonce.NewTimeWindow(appState.BlockchainSigner.GetAddress(), 2000)
	nonceDeposit := nonce.NewTimeWindow(appState.DepositSigner.GetAddress(), 10000)
	bc.RegisterPricingOperator(appState.BlockchainSigner, nonceCorpus)
	bc.RegisterDepositOperator(appState.DepositSigner, nonceDeposit)
	dataFetcher.SetBlockchain(bc)
	rData := data.NewReserveData(
		appState.DataStorage,
		dataFetcher,
		appState.DataControllerRunner,
		appState.Archive,
		appState.DataGlobalStorage,
		appState.Exchanges,
		appState.Setting,
	)

	gasPriceLimiter := gasinfo.NewNetworkGasPriceLimiter(kyberNetworkProxy, ac.GasConfig.FetchMaxGasCacheSeconds)
	gasInfo := gasinfo.NewGasPriceInfo(gasPriceLimiter, rData, gaspricedataclient.New(httpClient, ac.GasConfig.GasPriceURL))

	rCore := core.NewReserveCore(bc, appState.ActivityStorage, appState.Setting, gasInfo)
	return rData, gasInfo, rCore
}
