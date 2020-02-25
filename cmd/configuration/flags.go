package configuration

import (
	"github.com/urfave/cli"
	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data/blockchain"
	"github.com/KyberNetwork/reserve-data/cmd/deployment"
	"github.com/KyberNetwork/reserve-data/cmd/mode"
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/blockchain/nonce"
	"github.com/KyberNetwork/reserve-data/core"
	"github.com/KyberNetwork/reserve-data/data"
	"github.com/KyberNetwork/reserve-data/data/fetcher"
	"github.com/KyberNetwork/reserve-data/exchange/binance"
	"github.com/KyberNetwork/reserve-data/exchange/coinbase"
	"github.com/KyberNetwork/reserve-data/exchange/huobi"
	"github.com/KyberNetwork/reserve-data/lib/app"
	"github.com/KyberNetwork/reserve-data/reservesetting/storage/postgres"
)

const (
	dryRunFlag = "dry-run"

	binancePublicEndpointFlag  = "binance-public-endpoint"
	binancePublicEndpointValue = "https://api.binance.com"

	huobiPublicEndpointFlag  = "huobi-public-endpoint"
	huobiPublicEndpointValue = "https://api.huobi.pro"

	defaultDB = "reserve_data"
)

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

// NewBinanceCliFlags returns new configuration flags for Binance.
func NewBinanceCliFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:   binancePublicEndpointFlag,
			Usage:  "Binance public API endpoint",
			EnvVar: "BINANCE_PUBLIC_ENDPOINT",
			Value:  binancePublicEndpointValue,
		},
	}
}

// NewBinanceInterfaceFromContext returns the Binance endpoints configuration from cli context.
func NewBinanceInterfaceFromContext(c *cli.Context) binance.Interface {
	return binance.NewRealInterface(
		c.GlobalString(binancePublicEndpointFlag),
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
	}
}

// NewhuobiInterfaceFromContext returns the huobi endpoints configuration from cli context.
func NewhuobiInterfaceFromContext(c *cli.Context) huobi.Interface {
	return huobi.NewRealInterface(
		c.GlobalString(huobiPublicEndpointFlag),
	)
}

// NewCliFlags returns all cli flags of reserve core service.
func NewCliFlags() []cli.Flag {
	var flags []cli.Flag

	flags = append(flags, mode.NewCliFlag(), deployment.NewCliFlag())
	flags = append(flags, NewDryRunCliFlag())
	flags = append(flags, NewSecretConfigCliFlag()...)
	flags = append(flags, NewExchangeCliFlag())
	flags = append(flags, NewPostgreSQLFlags(defaultDB)...)
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
func NewConfigurationFromContext(c *cli.Context, rcf common.RawConfig, s *zap.SugaredLogger) (*Config, error) {

	bi := binance.NewRealInterface(rcf.ExchangeEndpoints.Binance.URL)
	hi := huobi.NewRealInterface(rcf.ExchangeEndpoints.Houbi.URL)
	coinbaseEndpoint := coinbase.NewRealInterface(rcf.ExchangeEndpoints.Coinbase.URL)

	contractAddressConf := &common.ContractAddressConfiguration{
		Reserve: rcf.ContractAddresses.Reserve,
		Proxy:   rcf.ContractAddresses.Proxy,
		Wrapper: rcf.ContractAddresses.Wrapper,
		Pricing: rcf.ContractAddresses.Pricing,
	}

	ethereumNodeConf := NewEthereumNodeConfiguration(rcf.Nodes.Main, rcf.Nodes.Backup)
	db, err := NewDBFromContext(c)
	if err != nil {
		return nil, err
	}

	// as this is core connect to setting db, the core endpoint is not needed
	sr, err := postgres.NewStorage(db)
	if err != nil {
		return nil, err
	}

	config, err := GetConfig(
		c,
		ethereumNodeConf,
		bi,
		hi,
		coinbaseEndpoint,
		contractAddressConf,
		sr,
		rcf,
	)
	if err != nil {
		return nil, err
	}

	return config, nil
}
