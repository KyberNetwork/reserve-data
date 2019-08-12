package configuration

import (
	"log"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data/cmd/deployment"
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/archive"
	"github.com/KyberNetwork/reserve-data/common/blockchain"
	"github.com/KyberNetwork/reserve-data/exchange/binance"
	"github.com/KyberNetwork/reserve-data/exchange/huobi"
	"github.com/KyberNetwork/reserve-data/v3/storage"
	"github.com/KyberNetwork/reserve-data/world"
)

const (
	byzantiumChainType = "byzantium"
	homesteadChainType = "homestead"
)

func GetChainType(dpl deployment.Deployment) string {
	switch dpl {
	case deployment.Production:
		return byzantiumChainType
	case deployment.Development:
		return homesteadChainType
	case deployment.Kovan:
		return homesteadChainType
	case deployment.Staging:
		return byzantiumChainType
	case deployment.Simulation, deployment.Analytic:
		return homesteadChainType
	case deployment.Ropsten:
		return byzantiumChainType
	default:
		return homesteadChainType
	}
}

func GetConfig(
	logger *zap.SugaredLogger,
	dpl deployment.Deployment,
	nodeConf *EthereumNodeConfiguration,
	bi binance.Interface,
	hi huobi.Interface,
	contractAddressConf *common.ContractAddressConfiguration,
	dataFile string,
	secretConfigFile string,
	enabledExchanges []common.ExchangeName,
	settingStorage storage.Interface,
) (*Config, error) {
	theWorld, err := world.NewTheWorld(logger, dpl, secretConfigFile)
	if err != nil {
		log.Printf("Can't init the world (which is used to get global data), err " + err.Error())
		return nil, err
	}

	chainType := GetChainType(dpl)

	//set client & endpoint
	client, err := rpc.Dial(nodeConf.Main)
	if err != nil {
		return nil, err
	}

	mainClient := ethclient.NewClient(client)
	bkClients := map[string]*ethclient.Client{}

	var callClients []*ethclient.Client
	for _, ep := range nodeConf.Backup {
		var bkClient *ethclient.Client
		bkClient, err = ethclient.Dial(ep)
		if err != nil {
			log.Printf("Cannot connect to %s, err %s. Ignore it.", ep, err)
		} else {
			bkClients[ep] = bkClient
			callClients = append(callClients, bkClient)
		}
	}

	bc := blockchain.NewBaseBlockchain(
		client, mainClient, map[string]*blockchain.Operator{},
		blockchain.NewBroadcaster(bkClients),
		chainType,
		blockchain.NewContractCaller(callClients, nodeConf.Backup),
	)

	awsConf, err := archive.GetAWSconfigFromFile(secretConfigFile)
	if err != nil {
		log.Printf("failed to load AWS config from file %s", secretConfigFile)
		return nil, err
	}
	s3archive := archive.NewS3Archive(awsConf)
	config := &Config{
		Blockchain:              bc,
		EthereumEndpoint:        nodeConf.Main,
		BackupEthereumEndpoints: nodeConf.Backup,
		Archive:                 s3archive,
		World:                   theWorld,
		ContractAddresses:       contractAddressConf,
		SettingStorage:          settingStorage,
	}

	log.Printf("configured endpoint: %s, backup: %v", config.EthereumEndpoint, config.BackupEthereumEndpoints)
	if err = config.AddCoreConfig(secretConfigFile, dpl, bi, hi, contractAddressConf, dataFile, enabledExchanges, settingStorage); err != nil {
		return nil, err
	}
	return config, nil
}
