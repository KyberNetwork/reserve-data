package cmd

import (
	"log"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data"
	"github.com/KyberNetwork/reserve-data/blockchain"
	"github.com/KyberNetwork/reserve-data/cmd/configuration"
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/config"
	"github.com/KyberNetwork/reserve-data/http"
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
	kyberENV := common.RunningMode()
	ac, err := config.LoadConfig(configFile)
	if err != nil {
		s.Panicw("failed to load config file", "err", err)
	}
	appState := configuration.InitAppState(!noAuthEnable, ac)
	//backup other log daily
	backupLog(appState.Archive, "@daily", "core.+\\.log")
	//backup core.log every 2 hour
	backupLog(appState.Archive, "@every 2h", "core\\.log")

	var (
		rData reserve.Data
		rCore reserve.Core
		bc    *blockchain.Blockchain
	)

	bc, err = CreateBlockchain(appState)
	if err != nil {
		log.Panicf("Can not create blockchain: (%s)", err)
	}

	rData, rCore = CreateDataCore(appState, kyberENV, bc)
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

	server := http.NewHTTPServer(
		rData, rCore,
		appState.MetricStorage,
		ac.HTTPAPIAddr,
		appState.EnableAuthentication,
		profilerPrefix,
		appState.AuthEngine,
		kyberENV,
		bc, appState.Setting,
	)

	if !dryRun {
		server.Run()
	} else {
		s.Infof("Dry run finished. All configs are corrected")
	}
}
