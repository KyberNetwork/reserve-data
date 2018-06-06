package configuration

import (
	"log"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/archive"
	"github.com/KyberNetwork/reserve-data/common/blockchain"
	"github.com/KyberNetwork/reserve-data/http"
	"github.com/KyberNetwork/reserve-data/world"
	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

func GetAddressConfig(filePath string) common.AddressConfig {
	addressConfig, err := common.GetAddressConfigFromFile(filePath)
	if err != nil {
		log.Fatalf("Config file %s is not found. Check that KYBER_ENV is set correctly. Error: %s", filePath, err)
	}
	return addressConfig
}

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

func GetConfig(kyberENV string, authEnbl bool, endpointOW string, noCore, enableStat bool) *Config {
	setPath := GetConfigPaths(kyberENV)

	world, err := world.NewTheWorld(kyberENV, setPath.secretPath)
	if err != nil {
		panic("Can't init the world (which is used to get global data), err " + err.Error())
	}

	addressConfig := GetAddressConfig(setPath.settingPath)
	hmac512auth := http.NewKNAuthenticationFromFile(setPath.secretPath)
	wrapperAddr := ethereum.HexToAddress(addressConfig.Wrapper)
	pricingAddr := ethereum.HexToAddress(addressConfig.Pricing)
	reserveAddr := ethereum.HexToAddress(addressConfig.Reserve)
	var endpoint string
	if endpointOW != "" {
		log.Printf("overwriting Endpoint with %s\n", endpointOW)
		endpoint = endpointOW
	} else {
		endpoint = setPath.endPoint
	}

	for id, t := range addressConfig.Tokens {
		tok := common.NewToken(id, t.Address, t.Decimals)
		if t.Active {
			if t.KNReserveSupport {
				common.RegisterInternalActiveToken(tok)
			} else {
				common.RegisterExternalActiveToken(tok)
			}
		} else {
			common.RegisterInactiveToken(tok)
		}
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
		bkclient, vErr := ethclient.Dial(ep)
		if vErr != nil {
			log.Printf("Cannot connect to %s, err %s. Ignore it.", ep, vErr)
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
		SupportedTokens:         common.InternalTokens(),
		WrapperAddress:          wrapperAddr,
		PricingAddress:          pricingAddr,
		ReserveAddress:          reserveAddr,
		ChainType:               chainType,
		AuthEngine:              hmac512auth,
		EnableAuthentication:    authEnbl,
		Archive:                 s3archive,
		World:                   world,
	}

	if enableStat {
		config.AddStatConfig(setPath, addressConfig)
	}

	if !noCore {
		config.AddCoreConfig(setPath, addressConfig, kyberENV)
	}
	return config
}
