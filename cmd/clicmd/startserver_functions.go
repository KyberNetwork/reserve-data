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
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	STARTING_BLOCK uint64 = 5069586
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
		config.WrapperAddress,
		config.PricingAddress,
		config.FeeBurnerAddress,
		config.NetworkAddress,
		config.ReserveAddress,
		config.WhitelistAddress,
	)
	if err != nil {
		panic(err)
	}

	// old contract addresses are used for events fetcher
	switch kyberENV {
	case common.PRODUCTION_MODE, common.MAINNET_MODE:
		bc.AddOldBurners(ethereum.HexToAddress("0x4E89bc8484B2c454f2F7B25b612b648c45e14A8e"))
		// TODO: add old contract v1 addresses
	case common.STAGING_MODE:
		// contract v1
		bc.AddOldNetwork(ethereum.HexToAddress("0xD2D21FdeF0D054D2864ce328cc56D1238d6b239e"))
		bc.AddOldBurners(ethereum.HexToAddress("0xB2cB365D803Ad914e63EA49c95eC663715c2F673"))
	}

	for _, token := range config.SupportedTokens {
		bc.AddToken(token)
	}
	err = bc.LoadAndSetTokenIndices()
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
		config.ReserveAddress,
		kyberENV == common.SIMULATION_MODE,
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
	)

	rCore := core.NewReserveCore(bc, config.ActivityStorage, config.ReserveAddress)
	return rData, rCore
}

func CreateStat(config *configuration.Config, kyberENV string, bc *blockchain.Blockchain) *stat.ReserveStats {
	var deployBlock uint64
	if kyberENV == common.MAINNET_MODE || kyberENV == common.PRODUCTION_MODE || kyberENV == common.DEV_MODE {
		deployBlock = STARTING_BLOCK
	}
	statFetcher := stat.NewFetcher(
		config.StatStorage,
		config.LogStorage,
		config.RateStorage,
		config.UserStorage,
		config.FeeSetRateStorage,
		config.StatFetcherRunner,
		deployBlock,
		config.ReserveAddress,
		config.PricingAddress,
		deployBlock,
		config.EtherscanApiKey,
		config.ThirdPartyReserves,
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
