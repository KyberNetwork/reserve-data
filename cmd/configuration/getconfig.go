package configuration

import (
	"log"
	"os"
	"time"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/core/intermediator"
	"github.com/KyberNetwork/reserve-data/data/fetcher"
	"github.com/KyberNetwork/reserve-data/data/fetcher/http_runner"
	"github.com/KyberNetwork/reserve-data/data/storage"
	"github.com/KyberNetwork/reserve-data/http"
	"github.com/KyberNetwork/reserve-data/signer"
	"github.com/KyberNetwork/reserve-data/stat"
	statstorage "github.com/KyberNetwork/reserve-data/stat/storage"
	ethereum "github.com/ethereum/go-ethereum/common"
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
	case "mainnet", "production":
		return "byzantium"
	case "dev":
		return "homestead"
	case "kovan":
		return "homestead"
	case "staging":
		return "byzantium"
	case "simulation":
		return "homestead"
	case "ropsten":
		return "byzantium"
	default:
		return "homestead"
	}
}

func GetConfigPaths(kyberENV string) SettingPaths {
	switch kyberENV {
	case "mainnet", "production":
		return (ConfigPaths["mainnet"])
	case "dev":
		return (ConfigPaths["dev"])
	case "kovan":
		return (ConfigPaths["kovan"])
	case "staging":
		return (ConfigPaths["staging"])
	case "simulation":
		return (ConfigPaths["simulation"])
	case "ropsten":
		return (ConfigPaths["ropsten"])
	default:
		log.Println("Environment setting paths is not found, using dev...")
		return (ConfigPaths["dev"])
	}
}

// GetConfig: load and set all config with preset params and customize param depends on env
// This is to generalized all the getconfig function.
func GetConfig(kyberENV string, authEnbl bool, endpointOW string) *Config {
	setPath := GetConfigPaths(kyberENV)
	// settingPath := "/go/src/github.com/KyberNetwork/reserve-data/cmd/dev_setting.json"
	addressConfig := GetAddressConfig(setPath.settingPath)

	feeConfig, err := common.GetFeeFromFile(setPath.feePath)
	if err != nil {
		log.Fatalf("Fees file %s cannot found at: %s", setPath.feePath, err)
	}
	wrapperAddr := ethereum.HexToAddress(addressConfig.Wrapper)
	reserveAddr := ethereum.HexToAddress(addressConfig.Reserve)
	pricingAddr := ethereum.HexToAddress(addressConfig.Pricing)
	burnerAddr := ethereum.HexToAddress(addressConfig.FeeBurner)
	networkAddr := ethereum.HexToAddress(addressConfig.Network)
	whitelistAddr := ethereum.HexToAddress(addressConfig.Whitelist)
	imtorAddr := ethereum.HexToAddress(addressConfig.Imtor)

	common.SupportedTokens = map[string]common.Token{}
	tokens := []common.Token{}
	for id, t := range addressConfig.Tokens {
		tok := common.Token{
			id, t.Address, t.Decimals,
		}
		common.SupportedTokens[id] = tok
		tokens = append(tokens, tok)
	}

	dataStorage, err := storage.NewBoltStorage(setPath.dataStoragePath)
	if err != nil {
		panic(err)
	}
	statStorage, err := statstorage.NewBoltStorage(setPath.statStoragePath)
	if err != nil {
		panic(err)
	}
	//fetcherRunner := http_runner.NewHttpRunner(8001)
	var fetcherRunner fetcher.FetcherRunner
	var statFetcherRunner stat.FetcherRunner
	var imrunner intermediator.IntermediatorRunner
	if os.Getenv("KYBER_ENV") == "simulation" {
		fetcherRunner = http_runner.NewHttpRunner(8001)
		statFetcherRunner = http_runner.NewHttpRunner(8002)
		imrunner = http_runner.NewHttpRunner(8003)
	} else {
		fetcherRunner = fetcher.NewTickerRunner(3*time.Second, 2*time.Second, 3*time.Second, 5*time.Second, 5*time.Second)
		statFetcherRunner = fetcher.NewTickerRunner(3*time.Second, 2*time.Second, 3*time.Second, 5*time.Second, 5*time.Second)
		imrunner = intermediator.NewTickerRunner(5 * time.Second)
	}
	fileSigner, depositSigner, intermediatorSigner := signer.NewFileSigner(setPath.signerPath)

	exchangePool := NewExchangePool(feeConfig, addressConfig, fileSigner, dataStorage, kyberENV)
	//exchangePool := exchangePoolFunc(feeConfig, addressConfig, fileSigner, storage)

	// endpoint := "https://ropsten.infura.io"
	// endpoint := "http://blockchain:8545"
	// endpoint := "https://kovan.infura.io"
	var endpoint string
	if endpointOW != "" {
		log.Printf("overwriting Endpoint with %s\n", endpointOW)
		endpoint = endpointOW
	} else {
		endpoint = setPath.endPoint
	}

	bkendpoints := setPath.bkendpoints
	var hmac512auth http.KNAuthentication

	hmac512auth = http.KNAuthentication{
		fileSigner.KNSecret,
		fileSigner.KNReadOnly,
		fileSigner.KNConfiguration,
		fileSigner.KNConfirmConf,
	}

	if !authEnbl {
		log.Printf("\nWARNING: No authentication mode\n")
	}

	chainType := GetChainType(kyberENV)

	return &Config{
		ActivityStorage:         dataStorage,
		DataStorage:             dataStorage,
		StatStorage:             statStorage,
		FetcherStorage:          dataStorage,
		StatFetcherStorage:      statStorage,
		MetricStorage:           dataStorage,
		FetcherRunner:           fetcherRunner,
		StatFetcherRunner:       statFetcherRunner,
		FetcherExchanges:        exchangePool.FetcherExchanges(),
		Exchanges:               exchangePool.CoreExchanges(),
		BlockchainSigner:        fileSigner,
		EnableAuthentication:    authEnbl,
		DepositSigner:           depositSigner,
		IntermediateSigner:      intermediatorSigner,
		AuthEngine:              hmac512auth,
		EthereumEndpoint:        endpoint,
		BackupEthereumEndpoints: bkendpoints,
		SupportedTokens:         tokens,
		WrapperAddress:          wrapperAddr,
		PricingAddress:          pricingAddr,
		ReserveAddress:          reserveAddr,
		FeeBurnerAddress:        burnerAddr,
		NetworkAddress:          networkAddr,
		WhitelistAddress:        whitelistAddr,
		ImtorAddress:            imtorAddr,
		ImtorRunner:             imrunner,
		ChainType:               chainType,
	}
}
