package configuration

import (
	"fmt"
	"log"

	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli"

	"github.com/KyberNetwork/reserve-data/blockchain"
	"github.com/KyberNetwork/reserve-data/cmd/deployment"
	"github.com/KyberNetwork/reserve-data/cmd/mode"
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/blockchain/nonce"
	"github.com/KyberNetwork/reserve-data/core"
	"github.com/KyberNetwork/reserve-data/data"
	"github.com/KyberNetwork/reserve-data/data/fetcher"
	"github.com/KyberNetwork/reserve-data/exchange/binance"
	"github.com/KyberNetwork/reserve-data/exchange/huobi"
	"github.com/KyberNetwork/reserve-data/v3/storage/postgres"
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

// NewBinanceCliFlags returns the Binance endpoints configuration from cli context.
func NewBinanceInterfaceFromContext(c *cli.Context) binance.Interface {
	return binance.NewRealInterface(
		c.GlobalString(binancePublicEndpointFlag),
		c.GlobalString(binanceAuthenticatedEndpointFlag),
	)
}

// NewhuobiCliFlags returns new configuration flags for huobi.
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

// NewhuobiCliFlags returns the huobi endpoints configuration from cli context.
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

	return flags
}

func CreateBlockchain(config *Config) (*blockchain.Blockchain, error) {
	var (
		bc  *blockchain.Blockchain
		err error
	)
	bc, err = blockchain.NewBlockchain(
		config.Blockchain,
		config.ContractAddresses,
		config.SettingStorage,
	)
	if err != nil {
		log.Printf("failed to create block chain err=%s", err.Error())
		return nil, err
	}

	// old contract addresses are used for events fetcher

	assets, err := config.SettingStorage.GetTransferableAssets()
	if err != nil {
		log.Printf("Can't get the list of Internal Tokens for indices: %s", err.Error())
		return nil, err
	}

	var assetAddrs []ethereum.Address
	for _, asset := range assets {
		if !common.IsEthereumAddress(asset.Address) {
			assetAddrs = append(assetAddrs, asset.Address)
		}
	}

	err = bc.LoadAndSetTokenIndices(assetAddrs)
	if err != nil {
		log.Printf("Can't load and set token indices: %s", err.Error())
		return nil, err
	}

	return bc, nil
}

// CreateDataCore create reserve data component
func CreateDataCore(config *Config, dpl deployment.Deployment, bc *blockchain.Blockchain) (*data.ReserveData, *core.ReserveCore) {
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
func NewConfigurationFromContext(c *cli.Context) (*Config, error) {
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

	enabledExchanges, err := NewExchangesFromContext(c)
	if err != nil {
		return nil, err
	}

	config, err := GetConfig(
		dpl,
		ethereumNodeConf,
		bi,
		hi,
		contractAddressConf,
		dataFile,
		secretConfigFile,
		enabledExchanges,
		sr,
	)
	if err != nil {
		return nil, err
	}

	return config, nil
}
