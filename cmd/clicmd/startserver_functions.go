package cmd

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/KyberNetwork/reserve-data/blockchain"
	"github.com/KyberNetwork/reserve-data/cmd/configuration"
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/archive"
	"github.com/KyberNetwork/reserve-data/common/blockchain/nonce"
	"github.com/KyberNetwork/reserve-data/core"
	"github.com/KyberNetwork/reserve-data/data"
	"github.com/KyberNetwork/reserve-data/data/fetcher"
	"github.com/robfig/cron"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

const (
	logFileName = "core.log"
)

func backupLog(arch archive.Archive, cronTimeExpression string, fileNameRegrexPattern string) {
	c := cron.New()
	err := c.AddFunc(cronTimeExpression, func() {
		files, rErr := ioutil.ReadDir(logDir)
		if rErr != nil {
			log.Printf("BACKUPLOG ERROR: Can not view log folder - %s", rErr.Error())
		}
		for _, file := range files {
			matched, err := regexp.MatchString(fileNameRegrexPattern, file.Name())
			if (!file.IsDir()) && (matched) && (err == nil) {
				err := arch.UploadFile(arch.GetLogBucketName(), remoteLogPath, filepath.Join(logDir, file.Name()))
				if err != nil {
					log.Printf("BACKUPLOG ERROR: Can not upload Log file %s", err)
				} else {
					var err error
					var ok bool
					if file.Name() != logFileName {
						ok, err = arch.CheckFileIntergrity(arch.GetLogBucketName(), remoteLogPath, filepath.Join(logDir, file.Name()))
						if !ok || (err != nil) {
							log.Printf("BACKUPLOG ERROR: File intergrity is corrupted")
						}
						err = os.Remove(filepath.Join(logDir, file.Name()))
					}
					if err != nil {
						log.Printf("BACKUPLOG ERROR: Cannot remove local log file %s", err)
					} else {
						log.Printf("BACKUPLOG Log backup: backup file %s succesfully", file.Name())
					}
				}
			}
		}
	})
	if err != nil {
		log.Printf("BACKUPLOG Cannot rotate log: %s", err.Error())
	}
	c.Start()
}

//set config log: Write log into a predefined file, and rotate log daily
//if stdoutLog is set, the log is also printed on stdout.
func configLog(stdoutLog bool) {
	logger := &lumberjack.Logger{
		Filename: filepath.Join(logDir, logFileName),
		// MaxSize:  1, // megabytes
		MaxBackups: 0,
		MaxAge:     0, //days
		// Compress:   true, // disabled by default
	}

	if stdoutLog {
		mw := io.MultiWriter(os.Stdout, logger)
		log.SetOutput(mw)
	} else {
		log.SetOutput(logger)
	}
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)

	c := cron.New()
	err := c.AddFunc("@daily", func() {
		if lErr := logger.Rotate(); lErr != nil {
			log.Printf("Error rotate log: %s", lErr.Error())
		}
	})
	if err != nil {
		log.Printf("Error add log cron daily: %s", err.Error())
	}
	c.Start()
}

//InitInterface initialize core interface
func InitInterface() {
	if baseURL != defaultBaseURL {
		log.Printf("Overwriting base URL with %s \n", baseURL)
	}
	configuration.SetInterface(baseURL)
}

// GetConfigFromENV From ENV variable and overwriting instruction, build the config
func GetConfigFromENV(kyberENV, configFile string) *configuration.Config {
	log.Printf("Running in %s mode \n", kyberENV)
	config := configuration.GetConfig(kyberENV,
		!noAuthEnable,
		endpointOW)
	return config
}

func CreateBlockchain(config *configuration.Config) (bc *blockchain.Blockchain, err error) {
	bc, err = blockchain.NewBlockchain(
		config.Blockchain,
		config.Setting,
	)

	if err != nil {
		panic(err)
	}

	// old contract addresses are used for events fetcher

	tokens, err := config.Setting.GetInternalTokens()
	if err != nil {
		log.Panicf("Can't get the list of Internal Tokens for indices: %s", err)
	}
	err = bc.LoadAndSetTokenIndices(common.GetTokenAddressesList(tokens))
	if err != nil {
		log.Panicf("Can't load and set token indices: %s", err)
	}
	return
}

func CreateDataCore(config *configuration.Config, kyberENV string, bc *blockchain.Blockchain) (*data.ReserveData, *core.ReserveCore) {
	//get fetcher based on config and ENV == simulation.
	dataFetcher := fetcher.NewFetcher(
		config.FetcherStorage,
		config.FetcherGlobalStorage,
		config.World,
		config.FetcherRunner,
		kyberENV == common.SimulationMode,
		config.Setting,
	)
	for _, ex := range config.FetcherExchanges {
		dataFetcher.AddExchange(ex)
	}
	nonceCorpus := nonce.NewTimeWindow(config.BlockchainSigner.GetAddress(), 2000)
	nonceDeposit := nonce.NewTimeWindow(config.DepositSigner.GetAddress(), 10000)
	bc.RegisterPricingOperator(config.BlockchainSigner, nonceCorpus)
	bc.RegisterDepositOperator(config.DepositSigner, nonceDeposit)
	dataFetcher.SetBlockchain(bc)
	rData := data.NewReserveData(
		config.DataStorage,
		dataFetcher,
		config.DataControllerRunner,
		config.Archive,
		config.DataGlobalStorage,
		config.Exchanges,
		config.Setting,
	)

	rCore := core.NewReserveCore(bc, config.ActivityStorage, config.Setting)
	return rData, rCore
}
