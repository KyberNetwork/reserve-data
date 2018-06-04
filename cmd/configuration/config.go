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
	"github.com/KyberNetwork/reserve-data/data/fetcher/http_runner"
	"github.com/KyberNetwork/reserve-data/data/storage"
	"github.com/KyberNetwork/reserve-data/exchange/binance"
	"github.com/KyberNetwork/reserve-data/exchange/bittrex"
	"github.com/KyberNetwork/reserve-data/exchange/huobi"
	"github.com/KyberNetwork/reserve-data/http"
	"github.com/KyberNetwork/reserve-data/metric"
	"github.com/KyberNetwork/reserve-data/settings"
	"github.com/KyberNetwork/reserve-data/stat"
	"github.com/KyberNetwork/reserve-data/stat/statpruner"
	statstorage "github.com/KyberNetwork/reserve-data/stat/storage"
	"github.com/KyberNetwork/reserve-data/world"
)

// SettingPaths contains path of all setting files.
type SettingPaths struct {
	settingPath           string
	feePath               string
	dataStoragePath       string
	analyticStoragePath   string
	statStoragePath       string
	logStoragePath        string
	rateStoragePath       string
	userStoragePath       string
	feeSetRateStoragePath string
	secretPath            string
	endPoint              string
	bkendpoints           []string
}

// NewSettingPaths creates new SettingPaths instance from given parameters.
func NewSettingPaths(
	settingPath, feePath, dataStoragePath, analyticStoragePath, statStoragePath,
	logStoragePath, rateStoragePath, userStoragePath, feeSetRateStoragePath, secretPath, endPoint string,
	bkendpoints []string) SettingPaths {
	cmdDir := common.CmdDirLocation()
	return SettingPaths{
		settingPath:           filepath.Join(cmdDir, settingPath),
		feePath:               filepath.Join(cmdDir, feePath),
		dataStoragePath:       filepath.Join(cmdDir, dataStoragePath),
		analyticStoragePath:   filepath.Join(cmdDir, analyticStoragePath),
		statStoragePath:       filepath.Join(cmdDir, statStoragePath),
		logStoragePath:        filepath.Join(cmdDir, logStoragePath),
		rateStoragePath:       filepath.Join(cmdDir, rateStoragePath),
		userStoragePath:       filepath.Join(cmdDir, userStoragePath),
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
	StatStorage          stat.StatStorage
	AnalyticStorage      stat.AnalyticStorage
	UserStorage          stat.UserStorage
	LogStorage           stat.LogStorage
	RateStorage          stat.RateStorage
	FeeSetRateStorage    stat.FeeSetRateStorage
	FetcherStorage       fetcher.Storage
	FetcherGlobalStorage fetcher.GlobalStorage
	MetricStorage        metric.MetricStorage
	Archive              archive.Archive

	World                *world.TheWorld
	FetcherRunner        fetcher.FetcherRunner
	DataControllerRunner datapruner.StorageControllerRunner
	StatFetcherRunner    stat.FetcherRunner
	StatControllerRunner statpruner.ControllerRunner
	FetcherExchanges     []fetcher.Exchange
	Exchanges            []common.Exchange
	BlockchainSigner     blockchain.Signer
	DepositSigner        blockchain.Signer
	//IntermediatorSigner blockchain.Signer

	EnableAuthentication bool
	AuthEngine           http.Authentication

	EthereumEndpoint        string
	BackupEthereumEndpoints []string
	Blockchain              *blockchain.BaseBlockchain

	// etherscan api key (optional)
	EtherscanApiKey string

	ChainType string
	Setting   *settings.Settings
}

// GetStatConfig: load config to run stat server only
func (self *Config) AddStatConfig(settingPath SettingPaths) {

	analyticStorage, err := statstorage.NewBoltAnalyticStorage(settingPath.analyticStoragePath)
	if err != nil {
		panic(err)
	}

	statStorage, err := statstorage.NewBoltStatStorage(settingPath.statStoragePath)
	if err != nil {
		panic(err)
	}

	logStorage, err := statstorage.NewBoltLogStorage(settingPath.logStoragePath)
	if err != nil {
		panic(err)
	}

	rateStorage, err := statstorage.NewBoltRateStorage(settingPath.rateStoragePath)
	if err != nil {
		panic(err)
	}

	userStorage, err := statstorage.NewBoltUserStorage(settingPath.userStoragePath)
	if err != nil {
		panic(err)
	}

	feeSetRateStorage, err := statstorage.NewBoltFeeSetRateStorage(settingPath.feeSetRateStoragePath)
	if err != nil {
		panic(err)
	}

	var statFetcherRunner stat.FetcherRunner
	var statControllerRunner statpruner.ControllerRunner
	if common.RunningMode() == common.SIMULATION_MODE {
		if statFetcherRunner, err = http_runner.NewHttpRunner(http_runner.WithHttpRunnerPort(8002)); err != nil {
			panic(err)
		}
	} else {
		statFetcherRunner = stat.NewTickerRunner(
			5*time.Second,  // block fetching interval
			7*time.Second,  // log fetching interval
			10*time.Second, // rate fetching interval
			2*time.Second,  // tradelog processing interval
			2*time.Second)  // catlog processing interval
		statControllerRunner = statpruner.NewControllerTickerRunner(24 * time.Hour)

	}

	apiKey := GetEtherscanAPIKey(settingPath.secretPath)

	self.StatStorage = statStorage
	self.AnalyticStorage = analyticStorage
	self.UserStorage = userStorage
	self.LogStorage = logStorage
	self.RateStorage = rateStorage
	self.StatControllerRunner = statControllerRunner
	self.FeeSetRateStorage = feeSetRateStorage
	self.StatFetcherRunner = statFetcherRunner
	self.EtherscanApiKey = apiKey
}

func (self *Config) AddCoreConfig(settingPath SettingPaths, kyberENV string) {
	dataStorage, err := storage.NewBoltStorage(settingPath.dataStoragePath)
	if err != nil {
		panic(err)
	}

	var fetcherRunner fetcher.FetcherRunner
	var dataControllerRunner datapruner.StorageControllerRunner
	if common.RunningMode() == common.SIMULATION_MODE {
		if fetcherRunner, err = http_runner.NewHttpRunner(http_runner.WithHttpRunnerPort(8001)); err != nil {
			log.Fatalf("failed to create HTTP runner: %s", err.Error())
		}
	} else {
		fetcherRunner = fetcher.NewTickerRunner(
			7*time.Second,  // orderbook fetching interval
			5*time.Second,  // authdata fetching interval
			3*time.Second,  // rate fetching interval
			5*time.Second,  // block fetching interval
			10*time.Minute, // tradeHistory fetching interval
			10*time.Second, // global data fetching interval
		)
		dataControllerRunner = datapruner.NewStorageControllerTickerRunner(24 * time.Hour)
	}

	pricingSigner := PricingSignerFromConfigFile(settingPath.secretPath)
	depositSigner := DepositSignerFromConfigFile(settingPath.secretPath)

	self.ActivityStorage = dataStorage
	self.DataStorage = dataStorage
	self.DataGlobalStorage = dataStorage
	self.FetcherStorage = dataStorage
	self.FetcherGlobalStorage = dataStorage
	self.MetricStorage = dataStorage
	self.FetcherRunner = fetcherRunner
	self.DataControllerRunner = dataControllerRunner
	self.BlockchainSigner = pricingSigner
	//self.IntermediatorSigner = huoBiintermediatorSigner
	self.DepositSigner = depositSigner

	//self.ExchangeStorage = exsStorage
	// var huobiConfig common.HuobiConfig
	// exchangesIDs := os.Getenv("KYBER_EXCHANGES")
	// if strings.Contains(exchangesIDs, "huobi") {
	// 	huobiConfig = *self.GetHuobiConfig(kyberENV, addressConfig.Intermediator, huobiIntermediatorSigner)
	// }

	// create Exchange pool
	exchangePool, err := NewExchangePool(
		settingPath,
		self.Blockchain,
		kyberENV,
		self.Setting,
	)
	if err != nil {
		log.Panicf("Can not create exchangePool: %s", err.Error())
	}
	self.FetcherExchanges = exchangePool.FetcherExchanges()
	self.Exchanges = exchangePool.CoreExchanges()
}

var ConfigPaths = map[string]SettingPaths{
	common.DEV_MODE: NewSettingPaths(
		"dev_setting.json",
		"fee.json",
		"dev.db",
		"dev_analytics.db",
		"dev_stats.db",
		"dev_logs.db",
		"dev_rates.db",
		"dev_users.db",
		"dev_fee_setrate.db",
		"config.json",
		"https://mainnet.infura.io",
		[]string{
			"https://api.myetherapi.com/eth",
		},
	),
	common.KOVAN_MODE: NewSettingPaths(
		"kovan_setting.json",
		"fee.json",
		"kovan.db",
		"kovan_analytics.db",
		"kovan_stats.db",
		"kovan_logs.db",
		"kovan_rates.db",
		"kovan_users.db",
		"kovan_fee_setrate.db",
		"config.json",
		"https://kovan.infura.io",
		[]string{},
	),
	common.PRODUCTION_MODE: NewSettingPaths(
		"mainnet_setting.json",
		"fee.json",
		"mainnet.db",
		"mainnet_analytics.db",
		"mainnet_stats.db",
		"mainnet_logs.db",
		"mainnet_rates.db",
		"mainnet_users.db",
		"mainnet_fee_setrate.db",
		"mainnet_config.json",
		"https://mainnet.infura.io",
		[]string{
			"https://semi-node.kyber.network",
			"https://api.mycryptoapi.com/eth",
			"https://api.myetherapi.com/eth",
			"https://mew.giveth.io/",
		},
	),
	common.MAINNET_MODE: NewSettingPaths(
		"mainnet_setting.json",
		"fee.json",
		"mainnet.db",
		"mainnet_analytics.db",
		"mainnet_stats.db",
		"mainnet_logs.db",
		"mainnet_rates.db",
		"mainnet_users.db",
		"mainnet_fee_setrate.db",
		"mainnet_config.json",
		"https://mainnet.infura.io",
		[]string{
			"https://mainnet.infura.io",
			"https://semi-node.kyber.network",
			"https://api.mycryptoapi.com/eth",
			"https://api.myetherapi.com/eth",
			"https://mew.giveth.io/",
		},
	),
	common.STAGING_MODE: NewSettingPaths(
		"staging_setting.json",
		"fee.json",
		"staging.db",
		"staging_analytics.db",
		"staging_stats.db",
		"staging_logs.db",
		"staging_rates.db",
		"staging_users.db",
		"staging_fee_setrate.db",
		"staging_config.json",
		"https://mainnet.infura.io",
		[]string{
			"https://mainnet.infura.io",
			"https://semi-node.kyber.network",
			"https://api.mycryptoapi.com/eth",
			"https://api.myetherapi.com/eth",
			"https://mew.giveth.io/",
		},
	),
	common.SIMULATION_MODE: NewSettingPaths(
		"shared/deployment_dev.json",
		"fee.json",
		"core.db",
		"core_analytics.db",
		"core_stats.db",
		"core_logs.db",
		"core_rates.db",
		"core_users.db",
		"core_fee_setrate.db",
		"config.json",
		"http://blockchain:8545",
		[]string{
			"http://blockchain:8545",
		},
	),
	common.ROPSTEN_MODE: NewSettingPaths(
		"ropsten_setting.json",
		"fee.json",
		"ropsten.db",
		"ropsten_analytics.db",
		"ropsten_stats.db",
		"ropsten_logs.db",
		"ropsten_rates.db",
		"ropsten_users.db",
		"ropsten_fee_setrate.db",
		"config.json",
		"https://ropsten.infura.io",
		[]string{
			"https://api.myetherapi.com/rop",
		},
	),
	common.ANALYTIC_DEV_MODE: NewSettingPaths(
		"shared/deployment_dev.json",
		"fee.json",
		"core.db",
		"core_analytics.db",
		"core_stats.db",
		"core_logs.db",
		"core_rates.db",
		"core_users.db",
		"core_fee_setrate.db",
		"config.json",
		"http://blockchain:8545",
		[]string{
			"http://blockchain:8545",
		},
	),
}

var Baseurl string = "http://127.0.0.1"

var BinanceInterfaces = make(map[string]binance.Interface)
var HuobiInterfaces = make(map[string]huobi.Interface)
var BittrexInterfaces = make(map[string]bittrex.Interface)

func SetInterface(base_url string) {
	BittrexInterfaces[common.DEV_MODE] = bittrex.NewDevInterface()
	BittrexInterfaces[common.KOVAN_MODE] = bittrex.NewKovanInterface(base_url)
	BittrexInterfaces[common.MAINNET_MODE] = bittrex.NewRealInterface()
	BittrexInterfaces[common.STAGING_MODE] = bittrex.NewRealInterface()
	BittrexInterfaces[common.SIMULATION_MODE] = bittrex.NewSimulatedInterface(base_url)
	BittrexInterfaces[common.ROPSTEN_MODE] = bittrex.NewRopstenInterface(base_url)
	BittrexInterfaces[common.ANALYTIC_DEV_MODE] = bittrex.NewRopstenInterface(base_url)

	HuobiInterfaces[common.DEV_MODE] = huobi.NewDevInterface()
	HuobiInterfaces[common.KOVAN_MODE] = huobi.NewKovanInterface(base_url)
	HuobiInterfaces[common.MAINNET_MODE] = huobi.NewRealInterface()
	HuobiInterfaces[common.STAGING_MODE] = huobi.NewRealInterface()
	HuobiInterfaces[common.SIMULATION_MODE] = huobi.NewSimulatedInterface(base_url)
	HuobiInterfaces[common.ROPSTEN_MODE] = huobi.NewRopstenInterface(base_url)
	HuobiInterfaces[common.ANALYTIC_DEV_MODE] = huobi.NewRopstenInterface(base_url)

	BinanceInterfaces[common.DEV_MODE] = binance.NewDevInterface()
	BinanceInterfaces[common.KOVAN_MODE] = binance.NewKovanInterface(base_url)
	BinanceInterfaces[common.MAINNET_MODE] = binance.NewRealInterface()
	BinanceInterfaces[common.STAGING_MODE] = binance.NewRealInterface()
	BinanceInterfaces[common.SIMULATION_MODE] = binance.NewSimulatedInterface(base_url)
	BinanceInterfaces[common.ROPSTEN_MODE] = binance.NewRopstenInterface(base_url)
	BinanceInterfaces[common.ANALYTIC_DEV_MODE] = binance.NewRopstenInterface(base_url)
}
