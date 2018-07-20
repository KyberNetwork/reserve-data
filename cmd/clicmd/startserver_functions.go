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
	"github.com/KyberNetwork/reserve-data/stat"
	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/robfig/cron"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

const (
	// startingBlockProduction is the block the first version of
	// production network contract is created.
	startingBlockProduction = 5069586
	// startingBlockStaging is the block the first version of
	// staging network contract is created.
	startingBlockStaging = 5042909
)

func backupLog(arch archive.Archive) {
	c := cron.New()
	err := c.AddFunc("@daily", func() {
		files, rErr := ioutil.ReadDir(logDir)
		if rErr != nil {
			log.Printf("ERROR: Log backup: Can not view log folder - %s", rErr.Error())
		}
		for _, file := range files {
			matched, err := regexp.MatchString("core.*\\.log", file.Name())
			if (!file.IsDir()) && (matched) && (err == nil) {
				log.Printf("File name is %s", file.Name())
				err := arch.UploadFile(arch.GetLogBucketName(), remoteLogPath, logDir+file.Name())
				if err != nil {
					log.Printf("ERROR: Log backup: Can not upload Log file %s", err)
				} else {
					var err error
					var ok bool
					if file.Name() != "core.log" {
						ok, err = arch.CheckFileIntergrity(arch.GetLogBucketName(), remoteLogPath, logDir+file.Name())
						if !ok || (err != nil) {
							log.Printf("ERROR: Log backup: File intergrity is corrupted")
						}
						err = os.Remove(logDir + file.Name())
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
	if err != nil {
		log.Printf("Cannot rotate log: %s", err.Error())
	}
	c.Start()
}

//set config log: Write log into a predefined file, and rotate log daily
//if stdoutLog is set, the log is also printed on stdout.
func configLog(stdoutLog bool) {
	logger := &lumberjack.Logger{
		Filename: filepath.Join(logDir, "core.log"),
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

func InitInterface() {
	if base_url != defaultBaseURL {
		log.Printf("Overwriting base URL with %s \n", base_url)
	}
	configuration.SetInterface(base_url)
}

// GetConfigFromENV: From ENV variable and overwriting instruction, build the config
func GetConfigFromENV(kyberENV string) *configuration.Config {
	log.Printf("Running in %s mode \n", kyberENV)
	var config *configuration.Config
	config = configuration.GetConfig(kyberENV,
		!noAuthEnable,
		endpointOW,
		noCore,
		enableStat)
	return config
}

func CreateBlockchain(config *configuration.Config, kyberENV string) (bc *blockchain.Blockchain, err error) {
	bc, err = blockchain.NewBlockchain(
		config.Blockchain,
		config.Setting,
	)
	if err != nil {
		panic(err)
	}

	// old contract addresses are used for events fetcher
	switch kyberENV {
	case common.ProductionMode, common.MainnetMode, common.DevMode:
		if uErr := bc.AddOldBurners(ethereum.HexToAddress("0x4E89bc8484B2c454f2F7B25b612b648c45e14A8e")); uErr != nil {
			log.Panic("ERROR: cannot add old burner")
		}
		// contract v1
		if uErr := bc.AddOldNetwork(ethereum.HexToAddress("0x964F35fAe36d75B1e72770e244F6595B68508CF5")); uErr != nil {
			log.Panic("ERROR: cannot add old burner")
		}
		if uErr := bc.AddOldBurners(ethereum.HexToAddress("0x07f6e905f2a1559cd9fd43cb92f8a1062a3ca706")); uErr != nil {
			log.Panic("ERROR: cannot add old burner")
		}
	case common.StagingMode:
		// contract v1
		if uErr := bc.AddOldNetwork(ethereum.HexToAddress("0xD2D21FdeF0D054D2864ce328cc56D1238d6b239e")); uErr != nil {
			log.Panic("ERROR: cannot add old burner")
		}
		if uErr := bc.AddOldBurners(ethereum.HexToAddress("0xB2cB365D803Ad914e63EA49c95eC663715c2F673")); uErr != nil {
			log.Panic("ERROR: cannot add old burner")
		}
	}
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

func CreateStat(config *configuration.Config, kyberENV string, bc *blockchain.Blockchain) *stat.ReserveStats {
	var deployBlock uint64

	switch kyberENV {
	case common.MainnetMode, common.ProductionMode, common.DevMode:
		deployBlock = startingBlockProduction
	case common.StagingMode:
		deployBlock = startingBlockStaging
	}

	statFetcher := stat.NewFetcher(
		config.StatStorage,
		config.LogStorage,
		config.RateStorage,
		config.UserStorage,
		config.FeeSetRateStorage,
		config.StatFetcherRunner,
		deployBlock,
		deployBlock,
		config.EtherscanApiKey,
		config.Setting,
	)
	statFetcher.SetBlockchain(bc)
	rStat := stat.NewReserveStats(
		config.AnalyticStorage,
		config.StatStorage,
		config.LogStorage,
		config.RateStorage,
		config.UserStorage,
		config.FeeSetRateStorage,
		config.StatControllerRunner,
		statFetcher,
		config.Archive,
	)
	return rStat
}
