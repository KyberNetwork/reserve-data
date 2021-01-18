package configuration

import (
	"context"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/urfave/cli"
	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/blockchain"
	"github.com/KyberNetwork/reserve-data/exchange/binance"
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
	contractAddressConf *common.ContractAddressConfiguration,
	settingStorage storage.Interface,
	rcf common.RawConfig,
	mainNode *common.EthClient, backupNodes []*common.EthClient,
) (*Config, error) {
	l := zap.S()

	bkClients := map[string]*ethclient.Client{}
	callClients := make([]*common.EthClient, 0, len(backupNodes)+1)

	// add main node to callclients
	callClients = append(callClients, mainNode)
	bkClients[mainNode.URL] = mainNode.Client
	for _, bn := range backupNodes {
		callClients = append(callClients, bn)
		bkClients[bn.URL] = bn.Client
	}

	bc := blockchain.NewBaseBlockchain(
		mainNode.RPCClient, mainNode.Client, map[string]*blockchain.Operator{},
		blockchain.NewBroadcaster(bkClients),
		blockchain.NewContractCaller(callClients),
	)

	theWorld := world.NewTheWorld(rcf.WorldEndpoints)

	config := &Config{
		Blockchain:              bc,
		EthereumEndpoint:        nodeConf.Main,
		BackupEthereumEndpoints: nodeConf.Backup,
		World:                   theWorld,
		ContractAddresses:       contractAddressConf,
		SettingStorage:          settingStorage,
	}
	chainID, err := mainNode.ChainID(context.Background())
	if err != nil {
		l.Errorw("get error get chainID", "error", err)
		return nil, err
	}
	l.Infow("configured endpoint", "endpoint", config.EthereumEndpoint, "backup", config.BackupEthereumEndpoints)
	if err := config.AddCoreConfig(cliCtx, rcf, bi, hi, settingStorage, chainID); err != nil {
		l.Errorw("failed to add core config", "error", err)
		return nil, err
	}
	return config, nil
}
