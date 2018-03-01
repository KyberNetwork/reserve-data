package configuration

import (
	"github.com/KyberNetwork/reserve-data/blockchain"
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/core"
	"github.com/KyberNetwork/reserve-data/core/intermediator"
	"github.com/KyberNetwork/reserve-data/data"
	"github.com/KyberNetwork/reserve-data/data/fetcher"
	"github.com/KyberNetwork/reserve-data/exchange/binance"
	"github.com/KyberNetwork/reserve-data/exchange/bittrex"
	"github.com/KyberNetwork/reserve-data/exchange/huobi"
	"github.com/KyberNetwork/reserve-data/http"
	"github.com/KyberNetwork/reserve-data/metric"
	"github.com/KyberNetwork/reserve-data/stat"
	statfetcher "github.com/KyberNetwork/reserve-data/stat/fetcher"
	ethereum "github.com/ethereum/go-ethereum/common"
)

type SettingPaths struct {
	settingPath     string
	feePath         string
	dataStoragePath string
	statStoragePath string
	signerPath      string
	endPoint        string
	bkendpoints     []string
}

type Config struct {
	ActivityStorage    core.ActivityStorage
	DataStorage        data.Storage
	StatStorage        stat.Storage
	FetcherStorage     fetcher.Storage
	StatFetcherStorage statfetcher.Storage
	MetricStorage      metric.MetricStorage

	FetcherRunner     fetcher.FetcherRunner
	StatFetcherRunner statfetcher.FetcherRunner
	FetcherExchanges  []fetcher.Exchange
	Exchanges         []common.Exchange
	BlockchainSigner  blockchain.Signer
	DepositSigner     blockchain.Signer

	EnableAuthentication bool
	AuthEngine           http.Authentication

	EthereumEndpoint        string
	BackupEthereumEndpoints []string

	SupportedTokens []common.Token

	WrapperAddress   ethereum.Address
	PricingAddress   ethereum.Address
	ReserveAddress   ethereum.Address
	FeeBurnerAddress ethereum.Address
	NetworkAddress   ethereum.Address
	WhitelistAddress ethereum.Address
	ImtorAddress     ethereum.Address
	ImtorRunner      intermediator.IntermediatorRunner
}

func (self *Config) MapTokens() map[string]common.Token {
	result := map[string]common.Token{}
	for _, t := range self.SupportedTokens {
		result[t.ID] = t
	}
	return result
}

var ConfigPaths = map[string]SettingPaths{
	"dev": {
		"/go/src/github.com/KyberNetwork/reserve-data/cmd/staging_setting.json",
		"/go/src/github.com/KyberNetwork/reserve-data/cmd/fee.json",
		"/go/src/github.com/KyberNetwork/reserve-data/cmd/dev.db",
		"/go/src/github.com/KyberNetwork/reserve-data/cmd/dev_stats.db",
		"/go/src/github.com/KyberNetwork/reserve-data/cmd/config.json",
		"https://mainnet.infura.io",
		[]string{
			"https://mainnet.infura.io",
		},
	},
	"kovan": {
		"/go/src/github.com/KyberNetwork/reserve-data/cmd/kovan_setting.json",
		"/go/src/github.com/KyberNetwork/reserve-data/cmd/fee.json",
		"/go/src/github.com/KyberNetwork/reserve-data/cmd/kovan.db",
		"/go/src/github.com/KyberNetwork/reserve-data/cmd/kovan_stats.db",
		"/go/src/github.com/KyberNetwork/reserve-data/cmd/config.json",
		"https://kovan.infura.io",
		[]string{},
	},
	"production": {
		"/go/src/github.com/KyberNetwork/reserve-data/cmd/mainnet_setting.json",
		"/go/src/github.com/KyberNetwork/reserve-data/cmd/fee.json",
		"/go/src/github.com/KyberNetwork/reserve-data/cmd/mainnet.db",
		"/go/src/github.com/KyberNetwork/reserve-data/cmd/mainnet_stats.db",
		"/go/src/github.com/KyberNetwork/reserve-data/cmd/mainnet_config.json",
		"https://mainnet.infura.io",
		[]string{
			"https://node.kyber.network",
			"https://mainnet.infura.io",
			"https://api.mycryptoapi.com/eth",
			"https://api.myetherapi.com/eth",
			"https://mew.giveth.io/",
		},
	},
	"mainnet": {
		"/go/src/github.com/KyberNetwork/reserve-data/cmd/mainnet_setting.json",
		"/go/src/github.com/KyberNetwork/reserve-data/cmd/fee.json",
		"/go/src/github.com/KyberNetwork/reserve-data/cmd/mainnet.db",
		"/go/src/github.com/KyberNetwork/reserve-data/cmd/mainnet_stats.db",
		"/go/src/github.com/KyberNetwork/reserve-data/cmd/mainnet_config.json",
		"https://mainnet.infura.io",
		[]string{
			"https://node.kyber.network",
			"https://mainnet.infura.io",
			"https://api.mycryptoapi.com/eth",
			"https://api.myetherapi.com/eth",
			"https://mew.giveth.io/",
		},
	},
	"staging": {
		"/go/src/github.com/KyberNetwork/reserve-data/cmd/staging_setting.json",
		"/go/src/github.com/KyberNetwork/reserve-data/cmd/fee.json",
		"/go/src/github.com/KyberNetwork/reserve-data/cmd/staging.db",
		"/go/src/github.com/KyberNetwork/reserve-data/cmd/staging_stats.db",
		"/go/src/github.com/KyberNetwork/reserve-data/cmd/staging_config.json",
		"https://mainnet.infura.io",
		[]string{
			"https://node.kyber.network",
			"https://mainnet.infura.io",
			"https://api.mycryptoapi.com/eth",
			"https://api.myetherapi.com/eth",
			"https://mew.giveth.io/",
		},
	},
	"simulation": {
		"/go/src/github.com/KyberNetwork/reserve-data/cmd/shared/deployment_dev.json",
		"/go/src/github.com/KyberNetwork/reserve-data/cmd/fee.json",
		"/go/src/github.com/KyberNetwork/reserve-data/cmd/core.db",
		"/go/src/github.com/KyberNetwork/reserve-data/cmd/core_stats.db",
		"/go/src/github.com/KyberNetwork/reserve-data/cmd/config.json",
		"http://blockchain:8545",
		[]string{
			"http://blockchain:8545",
		},
	},
	"ropsten": {
		"/go/src/github.com/KyberNetwork/reserve-data/cmd/ropsten_setting.json",
		"/go/src/github.com/KyberNetwork/reserve-data/cmd/fee.json",
		"/go/src/github.com/KyberNetwork/reserve-data/cmd/ropsten.db",
		"/go/src/github.com/KyberNetwork/reserve-data/cmd/ropsten_stats.db",
		"/go/src/github.com/KyberNetwork/reserve-data/cmd/config.json",
		"https://ropsten.infura.io",
		[]string{
			"https://api.myetherapi.com/rop",
		},
	},
}

var Baseurl string = "http://127.0.0.1"

var BinanceInterfaces = make(map[string]binance.Interface)
var HuobiInterfaces = make(map[string]huobi.Interface)
var BittrexInterfaces = make(map[string]bittrex.Interface)

func SetInterface(base_url string) {

	BittrexInterfaces["dev"] = bittrex.NewDevInterface()
	BittrexInterfaces["kovan"] = bittrex.NewKovanInterface(base_url)
	BittrexInterfaces["mainnet"] = bittrex.NewRealInterface()
	BittrexInterfaces["staging"] = bittrex.NewRealInterface()
	BittrexInterfaces["simulation"] = bittrex.NewSimulatedInterface(base_url)
	BittrexInterfaces["ropsten"] = bittrex.NewRopstenInterface(base_url)

	HuobiInterfaces["dev"] = huobi.NewDevInterface()
	HuobiInterfaces["kovan"] = huobi.NewKovanInterface(base_url)
	HuobiInterfaces["mainnet"] = huobi.NewRealInterface()
	HuobiInterfaces["staging"] = huobi.NewRealInterface()
	HuobiInterfaces["simulation"] = huobi.NewSimulatedInterface(base_url)
	HuobiInterfaces["ropsten"] = huobi.NewRopstenInterface(base_url)

	BinanceInterfaces["dev"] = binance.NewDevInterface()
	BinanceInterfaces["kovan"] = binance.NewKovanInterface(base_url)
	BinanceInterfaces["mainnet"] = binance.NewRealInterface()
	BinanceInterfaces["staging"] = binance.NewRealInterface()
	BinanceInterfaces["simulation"] = binance.NewSimulatedInterface(base_url)
	BinanceInterfaces["ropsten"] = binance.NewRopstenInterface(base_url)
}

var HuobiAsync = map[string]bool{
	"dev":        false,
	"kovan":      true,
	"mainnet":    true,
	"staging":    true,
	"simulation": false,
	"ropsten":    true,
}
