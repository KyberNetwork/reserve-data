package cmd

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/KyberNetwork/reserve-data/blockchain"
	"github.com/KyberNetwork/reserve-data/cmd/configuration"
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
	STARTING_BLOCK uint64 = 5069586
)

func loadTimestamp(path string) []uint64 {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	timestamp := []uint64{}
	err = json.Unmarshal(raw, &timestamp)
	if err != nil {
		panic(err)
	}
	return timestamp
}

//set config log: Write log into a predefined file, and rotate log daily
//if stdoutLog is set, the log is also printed on stdout.
func configLog(stdoutLog bool) {
	logger := &lumberjack.Logger{
		Filename: "/go/src/github.com/KyberNetwork/reserve-data/log/core.log",
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
	c.AddFunc("@daily", func() { logger.Rotate() })
	c.Start()
}

func InitInterface(kyberENV string) {
	if base_url != configuration.Baseurl {
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
	// we need to implicitly add old contract addresses to production
	if kyberENV == "production" || kyberENV == "mainnet" {
		// bc.AddOldNetwork(...)
		bc.AddOldBurners(ethereum.HexToAddress("0x4E89bc8484B2c454f2F7B25b612b648c45e14A8e"))
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
		kyberENV == "simulation",
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
	)

	rCore := core.NewReserveCore(bc, config.ActivityStorage, config.ReserveAddress)
	return rData, rCore
}

func CreateStat(config *configuration.Config, kyberENV string, bc *blockchain.Blockchain) *stat.ReserveStats {
	var deployBlock uint64
	if kyberENV == "mainnet" || kyberENV == "production" || kyberENV == "dev" {
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
