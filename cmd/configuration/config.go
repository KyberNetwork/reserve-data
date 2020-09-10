package configuration

import (
	"log"
	"time"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/archive"
	"github.com/KyberNetwork/reserve-data/common/blockchain"
	"github.com/KyberNetwork/reserve-data/common/config"
	"github.com/KyberNetwork/reserve-data/core"
	"github.com/KyberNetwork/reserve-data/data"
	"github.com/KyberNetwork/reserve-data/data/datapruner"
	"github.com/KyberNetwork/reserve-data/data/fetcher"
	"github.com/KyberNetwork/reserve-data/data/fetcher/httprunner"
	"github.com/KyberNetwork/reserve-data/data/storage"
	"github.com/KyberNetwork/reserve-data/http"
	"github.com/KyberNetwork/reserve-data/metric"
	"github.com/KyberNetwork/reserve-data/settings"
	"github.com/KyberNetwork/reserve-data/world"
)

type AppState struct {
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

	Setting        *settings.Settings
	AddressSetting *settings.AddressSetting
	AppConfig      config.AppConfig
}

func (c *AppState) AddCoreConfig(appc config.AppConfig) {
	setting, err := GetSetting(appc, c.AddressSetting)
	if err != nil {
		log.Panicf("Failed to create setting: %+v", err)
	}
	c.Setting = setting
	dataStorage, err := storage.NewBoltStorage(appc.DataDB)
	if err != nil {
		panic(err)
	}

	var fetcherRunner fetcher.Runner
	var dataControllerRunner datapruner.StorageControllerRunner
	if common.RunningMode() == common.SimulationMode {
		if fetcherRunner, err = httprunner.NewHTTPRunner(httprunner.WithBindAddr(appc.SimulationRunnerAddr)); err != nil {
			log.Fatalf("failed to create HTTP runner: %+v", err)
		}
	} else {
		fetcherRunner = fetcher.NewTickerRunner(
			time.Duration(appc.FetcherDelay.OrderBook),
			time.Duration(appc.FetcherDelay.AuthData),
			time.Duration(appc.FetcherDelay.RateFetching),
			time.Duration(appc.FetcherDelay.BlockFetching),
			time.Duration(appc.FetcherDelay.GlobalData),
		)
		dataControllerRunner = datapruner.NewStorageControllerTickerRunner(24 * time.Hour)
	}

	pricingSigner := blockchain.NewEthereumSigner(appc.KeyStorePath, appc.Passphrase)
	depositSigner := blockchain.NewEthereumSigner(appc.KeyStoreDepositPath, appc.PassphraseDeposit)

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
	exchangePool, err := NewExchangePool(appc,
		c.Blockchain,
		c.Setting,
	)
	if err != nil {
		log.Panicf("Can not create exchangePool: %+v", err)
	}
	fetcherExchanges, err := exchangePool.FetcherExchanges()
	if err != nil {
		log.Panicf("cannot Create fetcher exchanges : (%+v)", err)
	}
	c.FetcherExchanges = fetcherExchanges
	coreExchanges, err := exchangePool.CoreExchanges()
	if err != nil {
		log.Panicf("cannot Create core exchanges : (%+v)", err)
	}
	c.Exchanges = coreExchanges
}
