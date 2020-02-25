package configuration

import (
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/urfave/cli"
	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/archive"
	"github.com/KyberNetwork/reserve-data/common/blockchain"
	"github.com/KyberNetwork/reserve-data/exchange/binance"
	"github.com/KyberNetwork/reserve-data/exchange/coinbase"
	"github.com/KyberNetwork/reserve-data/exchange/huobi"
	"github.com/KyberNetwork/reserve-data/reservesetting/storage"
	"github.com/KyberNetwork/reserve-data/world"
)

// GetConfig return config for core
func GetConfig(
	cliCtx *cli.Context,
	nodeConf *EthereumNodeConfiguration,
	bi binance.Interface,
	hi huobi.Interface,
	cb coinbase.Interface,
	contractAddressConf *common.ContractAddressConfiguration,
	settingStorage storage.Interface,
	rcf common.RawConfig,
) (*Config, error) {
	l := zap.S()

	//set client & endpoint
	client, err := rpc.Dial(nodeConf.Main)
	if err != nil {
		return nil, err
	}

	mainClient := ethclient.NewClient(client)
	bkClients := map[string]*ethclient.Client{}

	var callClients []*common.EthClient

	// add main node to callclients
	callClients = append(callClients, &common.EthClient{
		Client: mainClient,
		URL:    nodeConf.Main,
	})
	for _, ep := range nodeConf.Backup {
		var bkClient *ethclient.Client
		bkClient, err = ethclient.Dial(ep)
		if err != nil {
			l.Warnw("Cannot connect to rpc endpoint", "endpoint", ep, "err", err)
		} else {
			bkClients[ep] = bkClient
			callClients = append(callClients, &common.EthClient{
				Client: bkClient,
				URL:    ep,
			})
		}
	}

	bc := blockchain.NewBaseBlockchain(
		client, mainClient, map[string]*blockchain.Operator{},
		blockchain.NewBroadcaster(bkClients),
		blockchain.NewContractCaller(callClients),
	)

	s3archive := archive.NewS3Archive(rcf.AWSConfig)
	theWorld := world.NewTheWorld(rcf.WorldEndpoints)

	config := &Config{
		Blockchain:              bc,
		EthereumEndpoint:        nodeConf.Main,
		BackupEthereumEndpoints: nodeConf.Backup,
		Archive:                 s3archive,
		World:                   theWorld,
		ContractAddresses:       contractAddressConf,
		SettingStorage:          settingStorage,
	}

	l.Infow("configured endpoint", "endpoint", config.EthereumEndpoint, "backup", config.BackupEthereumEndpoints)
	if err = config.AddCoreConfig(cliCtx, rcf, bi, hi, cb, settingStorage); err != nil {
		return nil, err
	}
	return config, nil
}
