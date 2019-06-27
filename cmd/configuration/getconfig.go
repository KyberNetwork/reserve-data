package configuration

import (
	"log"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/KyberNetwork/reserve-data/cmd/deployment"
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/archive"
	"github.com/KyberNetwork/reserve-data/common/blockchain"
	"github.com/KyberNetwork/reserve-data/exchange/binance"
	"github.com/KyberNetwork/reserve-data/exchange/huobi"
	"github.com/KyberNetwork/reserve-data/http"
	"github.com/KyberNetwork/reserve-data/settings"
	settingstorage "github.com/KyberNetwork/reserve-data/settings/storage"
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

func GetSetting(
	dpl deployment.Deployment,
	contractAddressConf *common.ContractAddressConfiguration,
	settingDataFile string) (*settings.Settings, error) {
	boltSettingStorage, err := settingstorage.NewBoltSettingStorage(settingDataFile)
	if err != nil {
		return nil, err
	}
	tokenSetting, err := settings.NewTokenSetting(boltSettingStorage)
	if err != nil {
		return nil, err
	}
	exchangeSetting, err := settings.NewExchangeSetting(boltSettingStorage)
	if err != nil {
		return nil, err
	}

	setting, err := settings.NewSetting(
		tokenSetting,
		contractAddressConf,
		exchangeSetting,
		settings.WithHandleEmptyToken(mustGetTokenConfig(dpl)),
		settings.WithHandleEmptyFee(FeeConfigs),
		settings.WithHandleEmptyMinDeposit(ExchangesMinDepositConfig),
		settings.WithHandleEmptyDepositAddress(mustGetExchangeConfig(dpl)),
		settings.WithHandleEmptyExchangeInfo())
	return setting, err
}

func GetConfig(
	dpl deployment.Deployment,
	authEnbl bool,
	nodeConf *EthereumNodeConfiguration,
	bi binance.Interface,
	hi huobi.Interface,
	contractAddressConf *common.ContractAddressConfiguration,
	dataFile string,
	secretConfigFile string,
	settingDataFile string,
) (*Config, error) {
	theWorld, err := world.NewTheWorld(dpl, secretConfigFile)
	if err != nil {
		log.Printf("Can't init the world (which is used to get global data), err " + err.Error())
		return nil, err
	}

	hmac512auth := http.NewKNAuthenticationFromFile(secretConfigFile)

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

	if !authEnbl {
		log.Printf("\nWARNING: No authentication mode\n")
	}
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
		AuthEngine:              hmac512auth,
		EnableAuthentication:    authEnbl,
		Archive:                 s3archive,
		World:                   theWorld,
		ContractAddresses:       contractAddressConf,
	}

	log.Printf("configured endpoint: %s, backup: %v", config.EthereumEndpoint, config.BackupEthereumEndpoints)
	if err = config.AddCoreConfig(secretConfigFile, dpl, bi, hi, contractAddressConf, dataFile, settingDataFile); err != nil {
		return nil, err
	}
	return config, nil
}
