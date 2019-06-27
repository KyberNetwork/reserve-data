package configuration

import (
	"log"
	"time"

	"github.com/KyberNetwork/reserve-data/cmd/deployment"
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

	Setting           *settings.Settings
	ContractAddresses *common.ContractAddressConfiguration
}

func (c *Config) AddCoreConfig(
	secretConfigFile string,
	dpl deployment.Deployment,
	bi binance.Interface,
	hi huobi.Interface,
	contractAddressConf *common.ContractAddressConfiguration,
	dataFile string,
	settingDataFile string,
) error {
	setting, err := GetSetting(dpl, contractAddressConf, settingDataFile)
	if err != nil {
		log.Printf("Failed to create setting: %s", err.Error())
		return err
	}
	c.Setting = setting
	dataStorage, err := storage.NewBoltStorage(dataFile)
	if err != nil {
		log.Printf("failed create new data storage database")
		return err
	}

	var fetcherRunner fetcher.Runner
	var dataControllerRunner datapruner.StorageControllerRunner
	if dpl == deployment.Simulation {
		if fetcherRunner, err = httprunner.NewHTTPRunner(httprunner.WithPort(8001)); err != nil {
			log.Printf("failed to create HTTP runner: %s", err.Error())
			return err
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

	pricingSigner, err := PricingSignerFromConfigFile(secretConfigFile)
	if err != nil {
		log.Printf("failed to get pricing signeeer from config file err=%s", err.Error())
		return err
	}
	depositSigner, err := DepositSignerFromConfigFile(secretConfigFile)
	if err != nil {
		log.Printf("failed to get deposit signer from config file err=%s", err.Error())
		return err
	}

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
		secretConfigFile,
		c.Blockchain,
		dpl,
		c.Setting,
		bi,
		hi,
	)
	if err != nil {
		log.Printf("Can not create exchangePool: %s", err.Error())
		return err
	}
	fetcherExchanges, err := exchangePool.FetcherExchanges()
	if err != nil {
		log.Printf("cannot Create fetcher exchanges : (%s)", err.Error())
		return err
	}
	c.FetcherExchanges = fetcherExchanges
	coreExchanges, err := exchangePool.CoreExchanges()
	if err != nil {
		log.Printf("cannot Create core exchanges : (%s)", err.Error())
		return err
	}
	c.Exchanges = coreExchanges
	return nil
}
