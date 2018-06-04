package configuration

import (
	"log"
	"path/filepath"

	"github.com/KyberNetwork/reserve-data/settings"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/archive"
	"github.com/KyberNetwork/reserve-data/common/blockchain"
	"github.com/KyberNetwork/reserve-data/http"
	settingstorage "github.com/KyberNetwork/reserve-data/settings/storage"
	"github.com/KyberNetwork/reserve-data/world"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

const (
	settingDBFileName string = "setting.db"
)

func GetChainType(kyberENV string) string {
	switch kyberENV {
	case common.MAINNET_MODE, common.PRODUCTION_MODE:
		return "byzantium"
	case common.DEV_MODE:
		return "homestead"
	case common.KOVAN_MODE:
		return "homestead"
	case common.STAGING_MODE:
		return "byzantium"
	case common.SIMULATION_MODE, common.ANALYTIC_DEV_MODE:
		return "homestead"
	case common.ROPSTEN_MODE:
		return "byzantium"
	default:
		return "homestead"
	}
}

func GetConfigPaths(kyberENV string) SettingPaths {
	// common.PRODUCTION_MODE and common.MAINNET_MODE are same thing.
	if kyberENV == common.PRODUCTION_MODE {
		kyberENV = common.MAINNET_MODE
	}

	if sp, ok := ConfigPaths[kyberENV]; ok {
		return sp
	}
	log.Println("Environment setting paths is not found, using dev...")
	return ConfigPaths[common.DEV_MODE]
}

func GetSetting(setPath SettingPaths) (*settings.Settings, error) {
	boltSettingStorage, err := settingstorage.NewBoltSettingStorage(filepath.Join(common.CmdDirLocation(), settingDBFileName))
	if err != nil {
		return nil, err
	}
	tokenSetting, err := settings.NewTokenSetting(boltSettingStorage)
	if err != nil {
		return nil, err
	}
	addressSetting, err := settings.NewAddressSetting(boltSettingStorage)
	if err != nil {
		return nil, err
	}
	exchangeSetting, err := settings.NewExchangeSetting(boltSettingStorage)
	if err != nil {
		return nil, err
	}
	setting, err := settings.NewSetting(
		tokenSetting,
		addressSetting,
		exchangeSetting,
		settings.WithHandleEmptyToken(setPath.settingPath),
		settings.WithHandleEmptyAddress(setPath.settingPath),
		settings.WithHandleEmptyFee(setPath.feePath),
		settings.WithHandleEmptyMinDeposit(filepath.Join(common.CmdDirLocation(), "min_deposit.json")),
		settings.WithHandleEmptyDepositAddress(setPath.settingPath),
		settings.WithHandleEmptyTokenPairs())
	return setting, err
}

func GetConfig(kyberENV string, authEnbl bool, endpointOW string, noCore, enableStat bool) *Config {
	setPath := GetConfigPaths(kyberENV)
	world, err := world.NewTheWorld(kyberENV, setPath.secretPath)
	if err != nil {
		panic("Can't init the world (which is used to get global data), err " + err.Error())
	}

	hmac512auth := http.NewKNAuthenticationFromFile(setPath.secretPath)
	setting, err := GetSetting(setPath)
	if err != nil {
		log.Panicf("Failed to create setting: %s", err.Error())
	}
	var endpoint string
	if endpointOW != "" {
		log.Printf("overwriting Endpoint with %s\n", endpointOW)
		endpoint = endpointOW
	} else {
		endpoint = setPath.endPoint
	}

	bkendpoints := setPath.bkendpoints
	chainType := GetChainType(kyberENV)

	//set client & endpoint
	client, err := rpc.Dial(endpoint)
	if err != nil {
		panic(err)
	}
	infura := ethclient.NewClient(client)
	bkclients := map[string]*ethclient.Client{}
	var callClients []*ethclient.Client
	for _, ep := range bkendpoints {
		bkclient, err := ethclient.Dial(ep)
		if err != nil {
			log.Printf("Cannot connect to %s, err %s. Ignore it.", ep, err)
		} else {
			bkclients[ep] = bkclient
			callClients = append(callClients, bkclient)
		}
	}

	blockchain := blockchain.NewBaseBlockchain(
		client, infura, map[string]*blockchain.Operator{},
		blockchain.NewBroadcaster(bkclients),
		blockchain.NewCMCEthUSDRate(),
		chainType,
		blockchain.NewContractCaller(callClients, setPath.bkendpoints),
	)

	if !authEnbl {
		log.Printf("\nWARNING: No authentication mode\n")
	}
	awsConf, err := archive.GetAWSconfigFromFile(setPath.secretPath)
	if err != nil {
		panic(err)
	}
	s3archive := archive.NewS3Archive(awsConf)

	config := &Config{
		Blockchain:              blockchain,
		EthereumEndpoint:        endpoint,
		BackupEthereumEndpoints: bkendpoints,
		ChainType:               chainType,
		AuthEngine:              hmac512auth,
		EnableAuthentication:    authEnbl,
		Archive:                 s3archive,
		World:                   world,
		Setting:                 setting,
	}

	if enableStat {
		config.AddStatConfig(setPath)
	}
	if !noCore {
		config.AddCoreConfig(setPath, kyberENV)
	}
	return config
}
