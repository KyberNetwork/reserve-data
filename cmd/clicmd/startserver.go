package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/KyberNetwork/reserve-data"
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/http"
	"github.com/spf13/cobra"
)

const remoteLogPath string = "core-log"

// logDir is located at base of this repository.
var logDir = filepath.Join(filepath.Dir(filepath.Dir(common.CurrentDir())), "log")
var noAuthEnable bool
var servPort int = 8000
var endpointOW string
var base_url string
var enableStat bool
var noCore bool
var stdoutLog bool
var dryrun bool

func serverStart(_ *cobra.Command, _ []string) {
	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU)
	configLog(stdoutLog)
	//create temp folder for exported file
	err := os.MkdirAll("./exported", 0730)
	if err != nil {
		panic(err)
	}
	//get configuration from ENV variable
	kyberENV := common.RunningMode()
	InitInterface(kyberENV)
	config := GetConfigFromENV(kyberENV)
	backupLog(config.Archive)

	var rData reserve.ReserveData
	var rCore reserve.ReserveCore
	var rStat reserve.ReserveStats

	//set static field supportExchange from common...
	for _, ex := range config.Exchanges {
		common.SupportedExchanges[ex.ID()] = ex
	}

	//Create blockchain object
	bc, err := CreateBlockchain(config, kyberENV)
	if err != nil {
		log.Panicf("Can not create blockchain: (%s)", err)
	}

	//Create Data and Core, run if not in dry mode
	if !noCore {
		rData, rCore = CreateDataCore(config, kyberENV, bc)
		if !dryrun {
			if kyberENV != common.SIMULATION_MODE {
				if err := rData.RunStorageController(); err != nil {
					log.Panic(err)
				}
			}
			if err := rData.Run(); err != nil {
				log.Panic(err)
			}
		}
	}

	//Create Stat, run if not in dry mode
	if enableStat {
		rStat = CreateStat(config, kyberENV, bc)
		if !dryrun {
			if kyberENV != common.SIMULATION_MODE {
				if err := rStat.RunStorageController(); err != nil {
					log.Panic(err)
				}
			}
			if err := rStat.Run(); err != nil {
				log.Panic(err)
			}
		}
	}

	//Create Server
	servPortStr := fmt.Sprintf(":%d", servPort)
	server := http.NewHTTPServer(
		rData, rCore, rStat,
		config.MetricStorage,
		servPortStr,
		config.EnableAuthentication,
		config.AuthEngine,
		kyberENV,
	)

	if !dryrun {
		server.Run()
	} else {
		log.Printf("Dry run finished. All configs are corrected")
	}
}

var startServer = &cobra.Command{
	Use:   "server ",
	Short: "initiate the server with specific config",
	Long: `Start reserve-data core server with preset Environment and
Allow overwriting some parameter`,
	Example: "KYBER_ENV=dev KYBER_EXCHANGES=bittrex ./cmd server --noauth -p 8000",
	Run:     serverStart,
}

func init() {
	// start server flags.
	startServer.Flags().BoolVarP(&noAuthEnable, "noauth", "", false, "disable authentication")
	startServer.Flags().IntVarP(&servPort, "port", "p", 8000, "server port")
	startServer.Flags().StringVar(&endpointOW, "endpoint", "", "endpoint, default to configuration file")
	startServer.PersistentFlags().StringVar(&base_url, "base_url", "http://127.0.0.1", "base_url for authenticated enpoint")
	startServer.Flags().BoolVarP(&enableStat, "enable-stat", "", false, "enable stat related fetcher and api, event logs will not be fetched")
	startServer.Flags().BoolVarP(&noCore, "no-core", "", false, "disable core related fetcher and api, this should be used only when we want to run an independent stat server")
	startServer.Flags().BoolVarP(&stdoutLog, "log-to-stdout", "", false, "send log to both log file and stdout terminal")
	startServer.Flags().BoolVarP(&dryrun, "dryrun", "", false, "only test if all the configs are set correctly, will not actually run core")

	RootCmd.AddCommand(startServer)
}
