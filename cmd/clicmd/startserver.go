package cmd

import (
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data"
	"github.com/KyberNetwork/reserve-data/blockchain"
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/http"
)

const (
	remoteLogPath  = "core-log"
	defaultBaseURL = "http://127.0.0.1"

	defaultOrderBookFetchingInterval  = 7 * time.Second
	defaultAuthDataFetchingInterval   = 5 * time.Second
	defaultRateFetchingInterval       = 3 * time.Second
	defaultBlockFetchingInterval      = 5 * time.Second
	defaultGlobalDataFetchingInterval = 10 * time.Second
)

var (
	// logDir is located at base of this repository.
	logDir         = filepath.Join(filepath.Dir(filepath.Dir(common.CurrentDir())), "log")
	noAuthEnable   bool
	servPort       = 8000
	endpointOW     string
	baseURL        string
	stdoutLog      bool
	dryRun         bool
	profilerPrefix string

	sentryDSN   string
	sentryLevel string
	zapMode     string

	cliAddress   common.AddressConfig
	runnerConfig common.RunnerConfig
)

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

	//get configuration from ENV variable
	kyberENV := common.RunningMode()
	InitInterface()
	config := GetConfigFromENV(kyberENV)
	//backup other log daily
	backupLog(config.Archive, "@daily", "core.+\\.log")
	//backup core.log every 2 hour
	backupLog(config.Archive, "@every 2h", "core\\.log")

	var (
		rData reserve.Data
		rCore reserve.Core
		bc    *blockchain.Blockchain
	)

	bc, err = CreateBlockchain(config)
	if err != nil {
		log.Panicf("Can not create blockchain: (%s)", err)
	}

	rData, rCore = CreateDataCore(config, kyberENV, bc)
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

	//set static field supportExchange from common...
	for _, ex := range config.Exchanges {
		common.SupportedExchanges[ex.ID()] = ex
	}

	//Create Server
	servPortStr := fmt.Sprintf(":%d", servPort)
	server := http.NewHTTPServer(
		rData, rCore,
		config.MetricStorage,
		servPortStr,
		config.EnableAuthentication,
		profilerPrefix,
		config.AuthEngine,
		kyberENV,
		bc, config.Setting,
	)

	if !dryRun {
		server.Run()
	} else {
		s.Infof("Dry run finished. All configs are corrected")
	}
}
