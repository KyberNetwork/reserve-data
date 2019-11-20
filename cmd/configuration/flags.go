package configuration

import (
	"fmt"
	"time"

	"github.com/urfave/cli"
	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data/blockchain"
	"github.com/KyberNetwork/reserve-data/cmd/deployment"
	"github.com/KyberNetwork/reserve-data/cmd/mode"
	"github.com/KyberNetwork/reserve-data/common/blockchain/nonce"
	"github.com/KyberNetwork/reserve-data/core"
	"github.com/KyberNetwork/reserve-data/data"
	"github.com/KyberNetwork/reserve-data/data/fetcher"
	"github.com/KyberNetwork/reserve-data/exchange/binance"
	"github.com/KyberNetwork/reserve-data/exchange/huobi"
	"github.com/KyberNetwork/reserve-data/lib/app"
	"github.com/KyberNetwork/reserve-data/reservesetting/storage/postgres"
)

const (
	httpAddressFlag  = "http-address"
	portDefaultValue = 8000

	dryRunFlag = "dry-run"

	binancePublicEndpointFlag         = "binance-public-endpoint"
	binancePublicEndpointValue        = "https://api.binance.com"
	binanceAuthenticatedEndpointFlag  = "binance-authenticated-endpoint"
	binanceAuthenticatedEndpointValue = "https://api.binance.com"

	huobiPublicEndpointFlag         = "huobi-public-endpoint"
	huobiPublicEndpointValue        = "https://api.huobi.pro"
	huobiAuthenticatedEndpointFlag  = "huobi-authenticated-endpoint"
	huobiAuthenticatedEndpointValue = "https://api.huobi.pro"

	defaultDB = "reserve_data"

	orderBookFetchingIntervalFlag    = "order-book-fetching-interval"
	authDataFetchingIntervalFlag     = "auth-data-fetching-interval"
	rateFetchingIntervalFlag         = "rate-fetching-interval"
	blockFetchingIntervalFlag        = "block-fetching-interval"
	globalDataFetchingIntervalFlag   = "global-data-fetching-interval"
	tradeHistoryFetchingIntervalFlag = "trade-history-fetching-interval"

	defaultOrderBookFetchingInterval    = 7 * time.Second
	defaultAuthDataFetchingInterval     = 5 * time.Second
	defaultRateFetchingInterval         = 3 * time.Second
	defaultBlockFetchingInterval        = 5 * time.Second
	defaultGlobalDataFetchingInterval   = 10 * time.Second
	defaultTradeHistoryFetchingInterval = 10 * time.Minute
)

// NewBinanceCliFlags returns new configuration flags for Binance.
func NewBinanceCliFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:   binancePublicEndpointFlag,
			Usage:  "Binance public API endpoint",
			EnvVar: "BINANCE_PUBLIC_ENDPOINT",
			Value:  binancePublicEndpointValue,
		},
		cli.StringFlag{
			Name:   binanceAuthenticatedEndpointFlag,
			Usage:  "Binance authenticated API endpoint",
			EnvVar: "BINANCE_AUTHENTICATED_ENDPOINT",
			Value:  binanceAuthenticatedEndpointValue,
		},
	}
}

// NewBinanceInterfaceFromContext returns the Binance endpoints configuration from cli context.
func NewBinanceInterfaceFromContext(c *cli.Context) binance.Interface {
	return binance.NewRealInterface(
		c.GlobalString(binancePublicEndpointFlag),
		c.GlobalString(binanceAuthenticatedEndpointFlag),
	)
}

// NewHuobiCliFlags returns new configuration flags for huobi.
func NewHuobiCliFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:   huobiPublicEndpointFlag,
			Usage:  "huobi public API endpoint",
			EnvVar: "huobi_PUBLIC_ENDPOINT",
			Value:  huobiPublicEndpointValue,
		},
		cli.StringFlag{
			Name:   huobiAuthenticatedEndpointFlag,
			Usage:  "huobi authenticated API endpoint",
			EnvVar: "huobi_AUTHENTICATED_ENDPOINT",
			Value:  huobiAuthenticatedEndpointValue,
		},
	}
}

// NewhuobiInterfaceFromContext returns the huobi endpoints configuration from cli context.
func NewhuobiInterfaceFromContext(c *cli.Context) huobi.Interface {
	return huobi.NewRealInterface(
		c.GlobalString(huobiPublicEndpointFlag),
		c.GlobalString(huobiAuthenticatedEndpointFlag),
	)
}

// NewDryRunCliFlag returns cli flag for dry run configuration.
func NewDryRunCliFlag() cli.Flag {
	return cli.BoolFlag{
		Name:   dryRunFlag,
		Usage:  "only test if all the configs are set correctly, will not actually run core",
		EnvVar: "DRY_RUN",
	}
}

// NewDryRunFromContext returns whether the to run reserve core in dry run mode.
func NewDryRunFromContext(c *cli.Context) bool {
	return c.GlobalBool(dryRunFlag)
}

//NewHTTPAddressFlag new flag for http address
func NewHTTPAddressFlag() cli.Flag {
	return cli.StringFlag{
		Name:   httpAddressFlag,
		Usage:  "bind address for http interface",
		EnvVar: "HTTP_ADDRESS",
		Value:  fmt.Sprintf("127.0.0.1:%d", portDefaultValue),
	}
}

//NewHTTPAddressFromContext return http listen address from context
func NewHTTPAddressFromContext(c *cli.Context) string {
	return c.GlobalString(httpAddressFlag)
}

// NewRunnerFlags return cli flag for runner configuration.
func NewRunnerFlags() []cli.Flag {
	return []cli.Flag{
		cli.DurationFlag{
			Name:   orderBookFetchingIntervalFlag,
			Usage:  "time interval fetch order book",
			EnvVar: "ORDER_BOOK_FETCHING_INTERVAL",
			Value:  defaultOrderBookFetchingInterval,
		},
		cli.DurationFlag{
			Name:   authDataFetchingIntervalFlag,
			Usage:  "time interval fetch auth data",
			EnvVar: "AUTH_DATA_FETCHING_INTERVAL",
			Value:  defaultAuthDataFetchingInterval,
		},
		cli.DurationFlag{
			Name:   rateFetchingIntervalFlag,
			Usage:  "time interval fetch rate",
			EnvVar: "RATE_FETCHING_INTERVAL",
			Value:  defaultRateFetchingInterval,
		},
		cli.DurationFlag{
			Name:   blockFetchingIntervalFlag,
			Usage:  "time interval fetch block",
			EnvVar: "BLOCK_FETCHING_INTERVAL",
			Value:  defaultBlockFetchingInterval,
		},
		cli.DurationFlag{
			Name:   globalDataFetchingIntervalFlag,
			Usage:  "time interval fetch global data",
			EnvVar: "GLOBAL_DATA_FETCHING_INTERVAL",
			Value:  defaultGlobalDataFetchingInterval,
		},
		cli.DurationFlag{
			Name:   tradeHistoryFetchingIntervalFlag,
			Usage:  "time interval fetch trade history",
			EnvVar: "TRADE_HISTORY_FETCHING_INTERVAL",
			Value:  defaultTradeHistoryFetchingInterval,
		},
	}
}

// NewTickerRunnerFromContext return TickerRunner from cli configs
func NewTickerRunnerFromContext(c *cli.Context) *fetcher.TickerRunner {
	return fetcher.NewTickerRunner(
		c.Duration(orderBookFetchingIntervalFlag),
		c.Duration(authDataFetchingIntervalFlag),
		c.Duration(rateFetchingIntervalFlag),
		c.Duration(blockFetchingIntervalFlag),
		c.Duration(globalDataFetchingIntervalFlag),
		c.Duration(tradeHistoryFetchingIntervalFlag),
	)
}

// NewCliFlags returns all cli flags of reserve core service.
func NewCliFlags() []cli.Flag {
	var flags []cli.Flag

	flags = append(flags, mode.NewCliFlag(), deployment.NewCliFlag())
	flags = append(flags, NewBinanceCliFlags()...)
	flags = append(flags, NewHuobiCliFlags()...)
	flags = append(flags, NewDryRunCliFlag())
	flags = append(flags, NewContractAddressCliFlags()...)
	flags = append(flags, NewEthereumNodesCliFlags()...)
	flags = append(flags, NewDataFileCliFlags()...)
	flags = append(flags, NewSecretConfigCliFlag())
	flags = append(flags, NewExchangeCliFlag())
	flags = append(flags, NewPostgreSQLFlags(defaultDB)...)
	flags = append(flags, NewHTTPAddressFlag())
	flags = append(flags, NewRunnerFlags()...)
	flags = append(flags, app.NewSentryFlags()...)

	return flags
}

// CreateBlockchain create new blockchain object
func CreateBlockchain(config *Config) (*blockchain.Blockchain, error) {
	var (
		bc  *blockchain.Blockchain
		err error
		l   = zap.S()
	)
	bc, err = blockchain.NewBlockchain(
		config.Blockchain,
		config.ContractAddresses,
		config.SettingStorage,
	)
	if err != nil {
		l.Errorw("failed to create block chain", "err", err)
		return nil, err
	}

	err = bc.LoadAndSetTokenIndices()
	if err != nil {
		l.Errorw("Can't load and set token indices", "err", err)
		return nil, err
	}

	return bc, nil
}

// CreateDataCore create reserve data component
func CreateDataCore(config *Config, dpl deployment.Deployment, bc *blockchain.Blockchain, l *zap.SugaredLogger) (*data.ReserveData, *core.ReserveCore) {
	//get fetcher based on config and ENV == simulation.
	dataFetcher := fetcher.NewFetcher(
		config.FetcherStorage,
		config.FetcherGlobalStorage,
		config.World,
		config.FetcherRunner,
		dpl == deployment.Simulation,
		config.ContractAddresses,
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
		config.SettingStorage,
	)

	rCore := core.NewReserveCore(bc, config.ActivityStorage, config.ContractAddresses)
	return rData, rCore
}

// NewConfigurationFromContext returns the Configuration object from cli context.
func NewConfigurationFromContext(c *cli.Context, s *zap.SugaredLogger) (*Config, error) {
	dpl, err := deployment.NewDeploymentFromContext(c)
	if err != nil {
		return nil, err
	}

	bi := NewBinanceInterfaceFromContext(c)
	hi := NewhuobiInterfaceFromContext(c)

	contractAddressConf, err := NewContractAddressConfigurationFromContext(c)
	if err != nil {
		return nil, err
	}

	ethereumNodeConf, err := NewEthereumNodeConfigurationFromContext(c)
	if err != nil {
		return nil, err
	}

	dataFile, err := NewDataFileFromContext(c)
	if err != nil {
		return nil, err
	}

	db, err := NewDBFromContext(c)
	if err != nil {
		return nil, err
	}

	sr, err := postgres.NewStorage(db)
	if err != nil {
		return nil, err
	}

	secretConfigFile := NewSecretConfigFileFromContext(c)

	config, err := GetConfig(
		c,
		dpl,
		ethereumNodeConf,
		bi,
		hi,
		contractAddressConf,
		dataFile,
		secretConfigFile,
		sr,
	)
	if err != nil {
		return nil, err
	}

	return config, nil
}
