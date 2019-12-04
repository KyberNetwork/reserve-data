package configuration

import (
	"path/filepath"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/archive"
	"github.com/KyberNetwork/reserve-data/common/blockchain"
	"github.com/KyberNetwork/reserve-data/http"
	"github.com/KyberNetwork/reserve-data/settings"
	settingstorage "github.com/KyberNetwork/reserve-data/settings/storage"
	"github.com/KyberNetwork/reserve-data/world"
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

func GetConfigPaths(kyberENV string) SettingPaths {
	// common.ProductionMode and common.MainnetMode are same thing.
	if kyberENV == common.ProductionMode {
		kyberENV = common.MainnetMode
	}

	if sp, ok := ConfigPaths[kyberENV]; ok {
		return sp
	}
	zap.S().Infof("Environment setting paths is not found, using dev")
	return ConfigPaths[common.DevMode]
}

func GetSetting(kyberENV string, addressSetting *settings.AddressSetting) (*settings.Settings, error) {
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
		settings.WithHandleEmptyToken(mustGetTokenConfig(kyberENV)),
		settings.WithHandleEmptyFee(FeeConfigs),
		settings.WithHandleEmptyMinDeposit(ExchangesMinDepositConfig),
		settings.WithHandleEmptyDepositAddress(mustGetExchangeConfig(kyberENV)),
		settings.WithHandleEmptyExchangeInfo())
	return setting, err
}

func newTheWorld(env string, sp SettingPaths) (*world.TheWorld, error) {
	switch env {
	case common.DevMode, common.KovanMode, common.MainnetMode, common.ProductionMode, common.StagingMode, common.RopstenMode, common.AnalyticDevMode:
		endpoint, err := world.NewRealEndpointFromFile(sp.secretPath)
		if err != nil {
			return nil, err
		}
		return world.NewTheWorld(endpoint, zap.S()), nil
	case common.SimulationMode:
		endpoint, err := world.NewSimulationEndpointFromFile(sp.worldEndpointFile)
		if err != nil {
			return nil, err
		}
		return world.NewTheWorld(endpoint, zap.S()), nil
	}
	panic("unsupported environment")
}

func GetConfig(kyberENV string, authEnbl bool, endpointOW string, cliAddress common.AddressConfig, runnerConfig common.RunnerConfig) *Config {
	l := zap.S()
	setPath := GetConfigPaths(kyberENV)

	theWorld, err := newTheWorld(kyberENV, setPath)
	if err != nil {
		l.Panicf("Can't init the world (which is used to get global data), err=%+v", err)
	}

	hmac512auth := http.NewKNAuthenticationFromFile(setPath.secretPath)
	addressSetting, err := settings.NewAddressSetting(mustGetAddressesConfig(kyberENV, cliAddress))
	if err != nil {
		l.Panicf("cannot init address setting %s", err)
	}
	var endpoint string
	if endpointOW != "" {
		l.Infof("overwriting Endpoint with %s\n", endpointOW)
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

	var callClients []*common.EthClient
	for _, ep := range bkEndpoints {
		client, err := common.NewEthClient(ep)
		if err != nil {
			l.Warnw("Cannot connect to RPC,ignore it.", "endpoint", ep, "err", err)
			continue
		}
		callClients = append(callClients, client)
		bkClients[ep] = client.Client
	}
	if len(callClients) == 0 {
		l.Warn("no backup client available")
	}

	bc := blockchain.NewBaseBlockchain(
		client, mainClient, map[string]*blockchain.Operator{},
		blockchain.NewBroadcaster(bkClients),
		chainType,
		blockchain.NewContractCaller(callClients),
	)

	if !authEnbl {
		l.Warnw("WARNING: No authentication mode")
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

	l.Infof("configured endpoint: %s, backup: %v", config.EthereumEndpoint, config.BackupEthereumEndpoints)

	config.AddCoreConfig(setPath, kyberENV, runnerConfig)
	return config
}
