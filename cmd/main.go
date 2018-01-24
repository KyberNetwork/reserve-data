package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"runtime"

	"github.com/KyberNetwork/reserve-data/blockchain"
	"github.com/KyberNetwork/reserve-data/blockchain/nonce"
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/core"
	"github.com/KyberNetwork/reserve-data/data"
	"github.com/KyberNetwork/reserve-data/data/fetcher"
	"github.com/KyberNetwork/reserve-data/http"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

func loadTimestamp(path string) []uint64 {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	timestamp := []uint64{}
	err = json.Unmarshal(raw, &timestamp)
	if err != nil {
		panic(err)
	}
	return timestamp
}

func main() {
	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU)

	var config *Config
	switch os.Getenv("KYBER_ENV") {
	case "mainnet", "production":
		log.Printf("Running in production mode")
		config = GetConfigForMainnet()
		break
	case "internal_mainnet":
		log.Printf("Running in internal mainnet mode")
		config = GetConfigForInternalMainnet()
		break
	case "simulation":
		log.Printf("Running in simulation mode")
		config = GetConfigForSimulation()
		break
	case "kovan":
		log.Printf("Running in kovan mode")
		config = GetConfigForKovan()
		break
	case "ropsten":
		log.Printf("Running in ropsten mode")
		config = GetConfigForRopsten()
		break
	case "dev":
		log.Printf("Running in dev mode")
		config = GetConfigForDev()
	default:
		log.Printf("Running in dev mode")
		config = GetConfigForDev()
	}

	addresses := config.GetAddress()
	logPath := "/go/src/github.com/KyberNetwork/reserve-data/cmd/log.log"
	f, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Couldn't open log file: %v", err)
	}
	mw := io.MultiWriter(os.Stdout, f)
	defer f.Close()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(mw)

	fetcher := fetcher.NewFetcher(
		config.FetcherStorage,
		config.FetcherRunner,
		config.ReserveAddress,
	)
	for _, ex := range config.Exchanges {
		common.SupportedExchanges[ex.ID()] = ex
	}
	for _, ex := range config.FetcherExchanges {
		fetcher.AddExchange(ex)
	}
	client, err := rpc.Dial(config.EthereumEndpoint)
	if err != nil {
		panic(err)
	}
	infura := ethclient.NewClient(client)

	// nonceCorpus := nonce.NewAutoIncreasing(infura, fileSigner)
	nonceCorpus := nonce.NewTimeWindow(infura, config.BlockchainSigner)

	bc, err := blockchain.NewBlockchain(
		client,
		infura,
		config.WrapperAddress,
		config.PricingAddress,
		config.FeeBurnerAddress,
		config.NetworkAddress,
		config.ReserveAddress,
		config.BlockchainSigner,
		nonceCorpus,
	)
	for _, token := range config.SupportedTokens {
		bc.AddToken(token)
	}
	err = bc.LoadAndSetTokenIndices()
	if err != nil {
		fmt.Printf("Can't load and set token indices: %s\n", err)
	} else {
		fetcher.SetBlockchain(bc)
		app := data.NewReserveData(
			config.DataStorage,
			fetcher,
		)
		app.Run()
		core := core.NewReserveCore(bc, config.ActivityStorage, config.ReserveAddress)
		server := http.NewHTTPServer(
			app, core,
			config.MetricStorage,
			":8000",
			config.EnableAuthentication,
			config.AuthEngine,
			addresses,
		)
		server.Run()
	}
}
