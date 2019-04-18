package configuration

import (
	"log"
	"path/filepath"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/archive"
	"github.com/KyberNetwork/reserve-data/common/blockchain"
	"github.com/KyberNetwork/reserve-data/http"
	"github.com/KyberNetwork/reserve-data/settings"
	settingstorage "github.com/KyberNetwork/reserve-data/settings/storage"
	"github.com/KyberNetwork/reserve-data/world"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

const (
	byzantiumChainType = "byzantium"
	homesteadChainType = "homestead"
)

func GetSettingDBName(kyberENV string) string {
	switch kyberENV {
	case common.MainnetMode, common.ProductionMode:
		return "mainnet_setting.db"
	case common.DevMode:
		return "dev_setting.db"
	case common.KovanMode:
		return "kovan_setting.db"
	case common.StagingMode:
		return "staging_setting.db"
	case common.SimulationMode, common.AnalyticDevMode:
		return "sim_setting.db"
	case common.RopstenMode:
		return "ropsten_setting.db"
	default:
		return "dev_setting.db"
	}
}

//GetChainType return type of chain base on running env
func GetChainType(kyberENV string) string {
	switch kyberENV {
	case common.MainnetMode, common.ProductionMode:
		return byzantiumChainType
	case common.DevMode:
		return homesteadChainType
	case common.KovanMode:
		return homesteadChainType
	case common.StagingMode:
		return byzantiumChainType
	case common.SimulationMode, common.AnalyticDevMode:
		return homesteadChainType
	case common.RopstenMode:
		return byzantiumChainType
	default:
		return homesteadChainType
	}
}

//GetConfigPaths return paths for config files
func GetConfigPaths(kyberENV, configFile string) SettingPaths {
	// common.ProductionMode and common.MainnetMode are same thing.
	if kyberENV == common.ProductionMode {
		kyberENV = common.MainnetMode
	}

	if sp, ok := ConfigPaths[kyberENV]; ok {
		sp.secretPath = configFile
		return sp
	}
	log.Println("Environment setting paths is not found, using dev...")
	sp := ConfigPaths[common.DevMode]
	sp.secretPath = configFile
	return sp
}

//GetSetting return setting base on running environment
func GetSetting(setPath SettingPaths, kyberENV string, addressSetting *settings.AddressSetting) (*settings.Settings, error) {
	boltSettingStorage, err := settingstorage.NewBoltSettingStorage(filepath.Join(common.CmdDirLocation(), GetSettingDBName(kyberENV)))
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
		addressSetting,
		exchangeSetting,
		settings.WithHandleEmptyToken(setPath.settingPath),
		settings.WithHandleEmptyFee(setPath.feePath),
		settings.WithHandleEmptyMinDeposit(filepath.Join(common.CmdDirLocation(), "min_deposit.json")),
		settings.WithHandleEmptyDepositAddress(setPath.settingPath),
		settings.WithHandleEmptyExchangeInfo())
	return setting, err
}

func GetConfig(kyberENV string, authEnbl bool, endpointOW string) *Config {
	setPath := GetConfigPaths(kyberENV)

	theWorld, err := world.NewTheWorld(kyberENV, setPath.secretPath)
	if err != nil {
		panic("Can't init the world (which is used to get global data), err " + err.Error())
	}

	hmac512auth := http.NewKNAuthenticationFromFile(setPath.secretPath)
	addressSetting, err := settings.NewAddressSetting(setPath.settingPath)
	if err != nil {
		log.Panicf("cannot init address setting %s", err)
	}
	var endpoint string
	if endpointOW != "" {
		log.Printf("overwriting Endpoint with %s\n", endpointOW)
		endpoint = endpointOW
	} else {
		endpoint = setPath.endPoint
	}

	bkEndpoints := setPath.bkendpoints

	// appending secret node to backup endpoints, as the fallback contract won't use endpoint
	if endpointOW != "" {
		bkEndpoints = append([]string{endpointOW}, bkEndpoints...)
	}

	chainType := GetChainType(kyberENV)

	//set client & endpoint
	client, err := rpc.Dial(endpoint)
	if err != nil {
		panic(err)
	}

	mainClient := ethclient.NewClient(client)
	bkClients := map[string]*ethclient.Client{}

	var callClients []*ethclient.Client
	for _, ep := range bkEndpoints {
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
		blockchain.NewContractCaller(callClients, bkEndpoints),
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
		Blockchain:              bc,
		EthereumEndpoint:        endpoint,
		BackupEthereumEndpoints: bkEndpoints,
		ChainType:               chainType,
		AuthEngine:              hmac512auth,
		EnableAuthentication:    authEnbl,
		Archive:                 s3archive,
		World:                   theWorld,
		AddressSetting:          addressSetting,
	}

	log.Printf("configured endpoint: %s, backup: %v", config.EthereumEndpoint, config.BackupEthereumEndpoints)

	config.AddCoreConfig(setPath, kyberENV)
	return config
}
