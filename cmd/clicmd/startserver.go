package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"runtime"

	"github.com/KyberNetwork/reserve-data"
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/archive"
	"github.com/KyberNetwork/reserve-data/http"
	"github.com/robfig/cron"
	"github.com/spf13/cobra"
)

const (
	LOG_PATH        string = "/go/src/github.com/KyberNetwork/reserve-data/log/"
	REMOTE_LOG_PATH string = "core-log/"
)

var noAuthEnable bool
var servPort int = 8000
var endpointOW string
var base_url, auth_url string
var enableStat bool
var noCore bool
var stdoutLog bool
var dryrun bool

func backupLog(arch archive.Archive) {
	c := cron.New()
	c.AddFunc("@daily", func() {
		files, err := ioutil.ReadDir(LOG_PATH)
		if err != nil {
			log.Printf("ERROR: Log backup: Can not view log folder")
		}
		for _, file := range files {
			matched, err := regexp.MatchString("core.*\\.log", file.Name())
			if (!file.IsDir()) && (matched) && (err == nil) {
				log.Printf("File name is %s", file.Name())
				err := arch.UploadFile(arch.GetLogBucketName(), REMOTE_LOG_PATH, LOG_PATH+file.Name())
				if err != nil {
					log.Printf("ERROR: Log backup: Can not upload Log file %s", err)
				} else {
					var err error
					var ok bool
					if file.Name() != "core.log" {
						ok, err = arch.CheckFileIntergrity(arch.GetLogBucketName(), REMOTE_LOG_PATH, LOG_PATH+file.Name())
						if !ok || (err != nil) {
							log.Printf("ERROR: Log backup: File intergrity is corrupted")
						}
						err = os.Remove(LOG_PATH + file.Name())
					}
					if err != nil {
						log.Printf("ERROR: Log backup: Cannot remove local log file %s", err)
					} else {
						log.Printf("Log backup: backup file %s succesfully", file.Name())
					}
				}
			}
		}
		return
	})
	c.Start()
}

func serverStart(cmd *cobra.Command, args []string) {
	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU)
	configLog(stdoutLog)
	//create temp folder for exported file
	err := os.MkdirAll("./exported", 0730)
	if err != nil {
		panic(err)
	}
	//get configuration from ENV variable
	kyberENV := os.Getenv("KYBER_ENV")
	if kyberENV == "" {
		kyberENV = "dev"
	}
	InitInterface(kyberENV)
	config := GetConfigFromENV(kyberENV)
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
			if kyberENV != "simulation" {
				rData.RunStorageController()
			}
			rData.Run()
		}
	}

	//Create Stat, run if not in dry mode
	if enableStat {
		rStat = CreateStat(config, kyberENV, bc)
		if !dryrun {
			if kyberENV != "simulation" {
				rStat.RunStorageController()
			}
			rStat.Run()
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
