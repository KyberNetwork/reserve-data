package configuration

import (
	"time"

	"github.com/urfave/cli"
	"go.uber.org/zap"

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
	"github.com/KyberNetwork/reserve-data/exchange/coinbase"
	"github.com/KyberNetwork/reserve-data/exchange/huobi"
	storagev3 "github.com/KyberNetwork/reserve-data/reservesetting/storage"
	"github.com/KyberNetwork/reserve-data/world"
)

// Config config for core
type Config struct {
	ActivityStorage      core.ActivityStorage
	DataStorage          data.Storage
	DataGlobalStorage    data.GlobalStorage
	FetcherStorage       fetcher.Storage
	FetcherGlobalStorage fetcher.GlobalStorage
	Archive              archive.Archive

	World                *world.TheWorld
	FetcherRunner        fetcher.Runner
	DataControllerRunner datapruner.StorageControllerRunner
	FetcherExchanges     []fetcher.Exchange
	Exchanges            []common.Exchange
	BlockchainSigner     blockchain.Signer
	DepositSigner        blockchain.Signer

	EthereumEndpoint        string
	BackupEthereumEndpoints []string
	Blockchain              *blockchain.BaseBlockchain

	SettingStorage    storagev3.Interface
	ContractAddresses *common.ContractAddressConfiguration
}

// AddCoreConfig add config for core
func (c *Config) AddCoreConfig(cliCtx *cli.Context, rcf common.RawConfig, bi binance.Interface,
	hi huobi.Interface, cb coinbase.Interface, settingStore storagev3.Interface) error {
	l := zap.S()
	db, err := NewDBFromContext(cliCtx)
	if err != nil {
		return err
	}
	dpl, err := deployment.NewDeploymentFromContext(cliCtx)
	if err != nil {
		return err
	}
	dataStorage, err := storage.NewPostgresStorage(db)
	if err != nil {
		l.Errorw("failed to create new data storage database", "err", err)
		return err
	}

	var fetcherRunner fetcher.Runner
	var dataControllerRunner datapruner.StorageControllerRunner
	if dpl == deployment.Simulation {
		if fetcherRunner, err = httprunner.NewHTTPRunner(httprunner.WithPort(8001)); err != nil {
			l.Errorw("failed to create HTTP runner", "err", err.Error())
			return err
		}
	} else {
		fetcherRunner = fetcher.NewTickerRunner(
			time.Duration(rcf.FetcherDelay.OrderBook),
			time.Duration(rcf.FetcherDelay.AuthData),
			time.Duration(rcf.FetcherDelay.RateFetching),
			time.Duration(rcf.FetcherDelay.BlockFetching),
			time.Duration(rcf.FetcherDelay.GlobalData),
			time.Duration(rcf.FetcherDelay.TradeHistory),
		)
		dataControllerRunner = datapruner.NewStorageControllerTickerRunner(24 * time.Hour)
	}

	c.ActivityStorage = dataStorage
	c.DataStorage = dataStorage
	c.DataGlobalStorage = dataStorage
	c.FetcherStorage = dataStorage
	c.FetcherGlobalStorage = dataStorage
	c.FetcherRunner = fetcherRunner
	c.DataControllerRunner = dataControllerRunner
	c.BlockchainSigner = blockchain.NewEthereumSigner(rcf.PricingKeystore, rcf.PricingPassphrase)
	c.DepositSigner = blockchain.NewEthereumSigner(rcf.DepositKeystore, rcf.DepositPassphrase)

	// create Exchange pool
	exchangePool, err := NewExchangePool(
		cliCtx,
		rcf,
		c.Blockchain,
		dpl,
		bi,
		hi,
		cb,
		settingStore,
	)
	if err != nil {
		l.Errorw("Can not create exchangePool", "err", err)
		return err
	}
	fetcherExchanges, err := exchangePool.FetcherExchanges()
	if err != nil {
		l.Errorw("cannot create fetcher exchanges", "err", err)
		return err
	}
	c.FetcherExchanges = fetcherExchanges
	coreExchanges, err := exchangePool.CoreExchanges()
	if err != nil {
		l.Errorw("cannot create core exchanges", "err", err)
		return err
	}
	c.Exchanges = coreExchanges
	return nil
}
