package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"runtime"

	"github.com/KyberNetwork/reserve-data"
	"github.com/KyberNetwork/reserve-data/blockchain"
	"github.com/KyberNetwork/reserve-data/blockchain/nonce"
	"github.com/KyberNetwork/reserve-data/cmd/configuration"
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/core"
	"github.com/KyberNetwork/reserve-data/core/intermediator"
	"github.com/KyberNetwork/reserve-data/data"
	"github.com/KyberNetwork/reserve-data/data/fetcher"
	"github.com/KyberNetwork/reserve-data/http"
	"github.com/KyberNetwork/reserve-data/stat"
	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/robfig/cron"
	"github.com/spf13/cobra"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

var noAuthEnable bool
var servPort int = 8000
var endpointOW string
var base_url, auth_url string
var enableStat bool
var noCore bool

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

// GetConfigFromENV: From ENV variable and overwriting instruction, build the config
func GetConfigFromENV(kyberENV string) *configuration.Config {
	log.Printf("Running in %s mode \n", kyberENV)
	var config *configuration.Config
	config = configuration.GetConfig(kyberENV,
		!noAuthEnable,
		endpointOW)
	return config
}

//set config log
func configLog() {
	logger := &lumberjack.Logger{
		Filename: "/go/src/github.com/KyberNetwork/reserve-data/log/core.log",
		// MaxSize:  1, // megabytes
		MaxBackups: 0,
		MaxAge:     0, //days
		// Compress:   true, // disabled by default
	}

	mw := io.MultiWriter(os.Stdout, logger)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
	log.SetOutput(mw)

	c := cron.New()
	c.AddFunc("@daily", func() { logger.Rotate() })
	c.Start()
}

func initInterface(kyberENV string) {
	if base_url != configuration.Baseurl {
		log.Printf("Overwriting base URL with %s \n", base_url)
	}
	configuration.SetInterface(base_url)
}

func serverStart(cmd *cobra.Command, args []string) {
	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU)
	configLog()

	//get configuration from ENV variable
	kyberENV := os.Getenv("KYBER_ENV")
	if kyberENV == "" {
		kyberENV = "dev"
	}
	initInterface(kyberENV)
	config := GetConfigFromENV(kyberENV)

	var dataFetcher *fetcher.Fetcher
	var statFetcher *stat.Fetcher
	var rData reserve.ReserveData
	var rCore reserve.ReserveCore
	var rStat reserve.ReserveStats

	//set static field supportExchange from common...
	for _, ex := range config.Exchanges {
		common.SupportedExchanges[ex.ID()] = ex
	}

	if !noCore {
		//get fetcher based on config and ENV == simulation.
		dataFetcher = fetcher.NewFetcher(
			config.FetcherStorage,
			config.FetcherRunner,
			config.ReserveAddress,
			kyberENV == "simulation",
		)
		for _, ex := range config.FetcherExchanges {
			dataFetcher.AddExchange(ex)
		}
	}

	if enableStat {
		var deployBlock uint64
		if kyberENV == "mainnet" || kyberENV == "production" {
			deployBlock = 5069586
		}
		statFetcher = stat.NewFetcher(
			config.StatFetcherStorage,
			stat.NewCMCEthUSDRate(),
			config.StatFetcherRunner,
			deployBlock,
		)
	}

	//set client & endpoint
	client, err := rpc.Dial(config.EthereumEndpoint)
	if err != nil {
		panic(err)
	}
	infura := ethclient.NewClient(client)
	bkclients := map[string]*ethclient.Client{}
	for _, ep := range config.BackupEthereumEndpoints {
		bkclient, err := ethclient.Dial(ep)
		if err != nil {
			log.Printf("Cannot connect to %s, err %s. Ignore it.", ep, err)
		} else {
			bkclients[ep] = bkclient
		}
	}

	//nonceCorpus := nonce.NewAutoIncreasing(infura, fileSigner)
	nonceCorpus := nonce.NewTimeWindow(infura, config.BlockchainSigner)
	nonceDeposit := nonce.NewTimeWindow(infura, config.DepositSigner)
	nonceIntemediate := nonce.NewTimeWindow(infura, config.IntermediateSigner)
	//set block chain
	bc, err := blockchain.NewBlockchain(
		client,
		infura,
		bkclients,
		config.WrapperAddress,
		config.PricingAddress,
		config.FeeBurnerAddress,
		config.NetworkAddress,
		config.ReserveAddress,
		config.WhitelistAddress,
		config.BlockchainSigner,
		config.DepositSigner,
		config.IntermediateSigner,
		nonceCorpus,
		nonceDeposit,
		nonceIntemediate,
		config.ChainType,
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
		fmt.Printf("Can't load and set token indices: %s\n", err)
	} else {
		if !noCore {
			dataFetcher.SetBlockchain(bc)
			rData = data.NewReserveData(
				config.DataStorage,
				dataFetcher,
			)

			imtor := intermediator.NewIntermediator(config.FetcherStorage, config.ImtorRunner, config.ImtorAddress, bc)
			imtor.Run()

			rData.Run()
			rCore = core.NewReserveCore(bc, config.ActivityStorage, config.ReserveAddress, config.ImtorAddress)
		}
		if enableStat {
			statFetcher.SetBlockchain(bc)
			rStat = stat.NewReserveStats(
				config.StatStorage,
				statFetcher,
			)
			rStat.Run()
		}

		servPortStr := fmt.Sprintf(":%d", servPort)
		server := http.NewHTTPServer(
			rData, rCore, rStat,
			config.MetricStorage,
			servPortStr,
			config.EnableAuthentication,
			config.AuthEngine,
			kyberENV,
		)

		server.Run()
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
	RootCmd.AddCommand(startServer)
}
