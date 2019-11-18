package configuration

import (
	"log"
	"path/filepath"
	"time"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/archive"
	"github.com/KyberNetwork/reserve-data/common/blockchain"
	"github.com/KyberNetwork/reserve-data/core"
	"github.com/KyberNetwork/reserve-data/data"
	"github.com/KyberNetwork/reserve-data/data/datapruner"
	"github.com/KyberNetwork/reserve-data/data/fetcher"
	"github.com/KyberNetwork/reserve-data/data/fetcher/httprunner"
	"github.com/KyberNetwork/reserve-data/data/storage"
	"github.com/KyberNetwork/reserve-data/exchange/binance"
	"github.com/KyberNetwork/reserve-data/exchange/huobi"
	"github.com/KyberNetwork/reserve-data/http"
	"github.com/KyberNetwork/reserve-data/metric"
	"github.com/KyberNetwork/reserve-data/settings"
	"github.com/KyberNetwork/reserve-data/world"
)

const (
	infuraProjectID = "/v3/59d9e06a1abe487e8e74664c06b337f9"

	alchemyapiMainnetEndpoint = "https://eth-mainnet.alchemyapi.io/jsonrpc/V1GjKybGLx6rzSu517KSWpSrTSIIXmV7"
	infuraMainnetEndpoint     = "https://mainnet.infura.io" + infuraProjectID
	infuraKovanEndpoint       = "https://kovan.infura.io" + infuraProjectID
	infuraRopstenEndpoint     = "https://ropsten.infura.io" + infuraProjectID
	myEtherAPIMainnetEndpoint = "https://api.myetherwallet.com/eth"
	myEtherAPIRopstenEndpoint = "https://api.myetherwallet.com/rop"
	semidNodeKyberEndpoint    = "https://semi-node.kyber.network"
	myCryptoAPIEndpoint       = "https://api.mycryptoapi.com/eth"
	mewGivethAPIEndpoint      = "https://mew.giveth.io/"

	localDevChainEndpoint = "http://blockchain:8545"
)

// SettingPaths contains path of all setting files.
type SettingPaths struct {
	settingPath           string
	feePath               string
	dataStoragePath       string
	analyticStoragePath   string
	feeSetRateStoragePath string
	secretPath            string
	endPoint              string
	bkendpoints           []string
}

// NewSettingPaths creates new SettingPaths instance from given parameters.
func NewSettingPaths(
	settingPath, feePath, dataStoragePath, analyticStoragePath,
	feeSetRateStoragePath, secretPath, endPoint string,
	bkendpoints []string) SettingPaths {
	cmdDir := common.CmdDirLocation()
	return SettingPaths{
		settingPath:           filepath.Join(cmdDir, settingPath),
		feePath:               filepath.Join(cmdDir, feePath),
		dataStoragePath:       filepath.Join(cmdDir, dataStoragePath),
		analyticStoragePath:   filepath.Join(cmdDir, analyticStoragePath),
		feeSetRateStoragePath: filepath.Join(cmdDir, feeSetRateStoragePath),
		secretPath:            filepath.Join(cmdDir, secretPath),
		endPoint:              endPoint,
		bkendpoints:           bkendpoints,
	}
}

type Config struct {
	ActivityStorage      core.ActivityStorage
	DataStorage          data.Storage
	DataGlobalStorage    data.GlobalStorage
	FetcherStorage       fetcher.Storage
	FetcherGlobalStorage fetcher.GlobalStorage
	MetricStorage        metric.Storage
	Archive              archive.Archive

	World                *world.TheWorld
	FetcherRunner        fetcher.Runner
	DataControllerRunner datapruner.StorageControllerRunner
	FetcherExchanges     []fetcher.Exchange
	Exchanges            []common.Exchange
	BlockchainSigner     blockchain.Signer
	DepositSigner        blockchain.Signer

	EnableAuthentication bool
	AuthEngine           http.Authentication

	EthereumEndpoint        string
	BackupEthereumEndpoints []string
	Blockchain              *blockchain.BaseBlockchain

	ChainType      string
	Setting        *settings.Settings
	AddressSetting *settings.AddressSetting
}

func (c *Config) AddCoreConfig(settingPath SettingPaths, kyberENV string) {
	setting, err := GetSetting(kyberENV, c.AddressSetting)
	if err != nil {
		log.Panicf("Failed to create setting: %s", err.Error())
	}
	c.Setting = setting
	dataStorage, err := storage.NewBoltStorage(settingPath.dataStoragePath)
	if err != nil {
		panic(err)
	}

	var fetcherRunner fetcher.Runner
	var dataControllerRunner datapruner.StorageControllerRunner
	if common.RunningMode() == common.SimulationMode {
		if fetcherRunner, err = httprunner.NewHTTPRunner(httprunner.WithPort(8001)); err != nil {
			log.Fatalf("failed to create HTTP runner: %s", err.Error())
		}
	} else {
		fetcherRunner = fetcher.NewTickerRunner(
			7*time.Second,  // orderbook fetching interval
			5*time.Second,  // authdata fetching interval
			3*time.Second,  // rate fetching interval
			5*time.Second,  // block fetching interval
			10*time.Second, // global data fetching interval
		)
		dataControllerRunner = datapruner.NewStorageControllerTickerRunner(24 * time.Hour)
	}

	pricingSigner := PricingSignerFromConfigFile(settingPath.secretPath)
	depositSigner := DepositSignerFromConfigFile(settingPath.secretPath)

	c.ActivityStorage = dataStorage
	c.DataStorage = dataStorage
	c.DataGlobalStorage = dataStorage
	c.FetcherStorage = dataStorage
	c.FetcherGlobalStorage = dataStorage
	c.MetricStorage = dataStorage
	c.FetcherRunner = fetcherRunner
	c.DataControllerRunner = dataControllerRunner
	c.BlockchainSigner = pricingSigner
	c.DepositSigner = depositSigner

	// create Exchange pool
	exchangePool, err := NewExchangePool(
		settingPath,
		c.Blockchain,
		kyberENV,
		c.Setting,
	)
	if err != nil {
		log.Panicf("Can not create exchangePool: %s", err.Error())
	}
	fetcherExchanges, err := exchangePool.FetcherExchanges()
	if err != nil {
		log.Panicf("cannot Create fetcher exchanges : (%s)", err.Error())
	}
	c.FetcherExchanges = fetcherExchanges
	coreExchanges, err := exchangePool.CoreExchanges()
	if err != nil {
		log.Panicf("cannot Create core exchanges : (%s)", err.Error())
	}
	c.Exchanges = coreExchanges
}

var ConfigPaths = map[string]SettingPaths{
	common.DevMode: NewSettingPaths(
		"dev_setting.json",
		"fee.json",
		"dev.db",
		"dev_analytics.db",
		"dev_fee_setrate.db",
		"config.json",
		infuraMainnetEndpoint,
		[]string{
			semidNodeKyberEndpoint,
			infuraMainnetEndpoint,
			myCryptoAPIEndpoint,
			myEtherAPIMainnetEndpoint,
		},
	),
	common.KovanMode: NewSettingPaths(
		"kovan_setting.json",
		"fee.json",
		"kovan.db",
		"kovan_analytics.db",
		"kovan_fee_setrate.db",
		"config.json",
		infuraKovanEndpoint,
		[]string{},
	),
	common.ProductionMode: NewSettingPaths(
		"mainnet_setting.json",
		"fee.json",
		"mainnet.db",
		"mainnet_analytics.db",
		"mainnet_fee_setrate.db",
		"mainnet_config.json",
		alchemyapiMainnetEndpoint,
		[]string{
			alchemyapiMainnetEndpoint,
			infuraMainnetEndpoint,
			semidNodeKyberEndpoint,
			myCryptoAPIEndpoint,
			myEtherAPIMainnetEndpoint,
			mewGivethAPIEndpoint,
		},
	),
	common.MainnetMode: NewSettingPaths(
		"mainnet_setting.json",
		"fee.json",
		"mainnet.db",
		"mainnet_analytics.db",
		"mainnet_fee_setrate.db",
		"mainnet_config.json",
		alchemyapiMainnetEndpoint,
		[]string{
			alchemyapiMainnetEndpoint,
			infuraMainnetEndpoint,
			semidNodeKyberEndpoint,
			myCryptoAPIEndpoint,
			myEtherAPIMainnetEndpoint,
			mewGivethAPIEndpoint,
		},
	),
	common.StagingMode: NewSettingPaths(
		"staging_setting.json",
		"fee.json",
		"staging.db",
		"staging_analytics.db",
		"staging_fee_setrate.db",
		"staging_config.json",
		alchemyapiMainnetEndpoint,
		[]string{
			alchemyapiMainnetEndpoint,
			infuraMainnetEndpoint,
			semidNodeKyberEndpoint,
			myCryptoAPIEndpoint,
			myEtherAPIMainnetEndpoint,
			mewGivethAPIEndpoint,
		},
	),
	common.SimulationMode: NewSettingPaths(
		"shared/deployment_dev.json",
		"fee.json",
		"core.db",
		"core_analytics.db",
		"core_fee_setrate.db",
		"config.json",
		localDevChainEndpoint,
		[]string{
			localDevChainEndpoint,
		},
	),
	common.RopstenMode: NewSettingPaths(
		"ropsten_setting.json",
		"fee.json",
		"ropsten.db",
		"ropsten_analytics.db",
		"ropsten_fee_setrate.db",
		"config.json",
		infuraRopstenEndpoint,
		[]string{
			myEtherAPIRopstenEndpoint,
		},
	),
	common.AnalyticDevMode: NewSettingPaths(
		"shared/deployment_dev.json",
		"fee.json",
		"core.db",
		"core_analytics.db",
		"core_fee_setrate.db",
		"config.json",
		localDevChainEndpoint,
		[]string{
			localDevChainEndpoint,
		},
	),
}

var BinanceInterfaces = make(map[string]binance.Interface)
var HuobiInterfaces = make(map[string]huobi.Interface)

func SetInterface(baseURL string) {
	HuobiInterfaces[common.DevMode] = huobi.NewDevInterface()
	HuobiInterfaces[common.KovanMode] = huobi.NewKovanInterface(baseURL)
	HuobiInterfaces[common.MainnetMode] = huobi.NewRealInterface()
	HuobiInterfaces[common.StagingMode] = huobi.NewRealInterface()
	HuobiInterfaces[common.SimulationMode] = huobi.NewSimulatedInterface(baseURL)
	HuobiInterfaces[common.RopstenMode] = huobi.NewRopstenInterface(baseURL)
	HuobiInterfaces[common.AnalyticDevMode] = huobi.NewRopstenInterface(baseURL)

	BinanceInterfaces[common.DevMode] = binance.NewDevInterface()
	BinanceInterfaces[common.KovanMode] = binance.NewKovanInterface(baseURL)
	BinanceInterfaces[common.MainnetMode] = binance.NewRealInterface()
	BinanceInterfaces[common.StagingMode] = binance.NewRealInterface()
	BinanceInterfaces[common.SimulationMode] = binance.NewSimulatedInterface(baseURL)
	BinanceInterfaces[common.RopstenMode] = binance.NewRopstenInterface(baseURL)
	BinanceInterfaces[common.AnalyticDevMode] = binance.NewRopstenInterface(baseURL)
}
