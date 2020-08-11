package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// SiteConfig contain config for a remote api access
type SiteConfig struct {
	URL string `json:"url"`
}

// WorldEndpoints hold detail information to fetch feed(url,header, api key...)
type WorldEndpoints struct {
	GoldData        SiteConfig `json:"gold_data"`
	OneForgeGoldETH SiteConfig `json:"one_forge_gold_eth"`
	OneForgeGoldUSD SiteConfig `json:"one_forge_gold_usd"`
	GDAXData        SiteConfig `json:"gdax_data"`
	KrakenData      SiteConfig `json:"kraken_data"`
	GeminiData      SiteConfig `json:"gemini_data"`

	CoinbaseETHBTC3 SiteConfig `json:"coinbase_eth_btc_3"`
	BinanceETHBTC3  SiteConfig `json:"binance_eth_btc_3"`

	CoinbaseETHUSDC10000  SiteConfig `json:"coinbase_eth_usdc_10000"`
	BinanceETHUSDC10000   SiteConfig `json:"binance_eth_usdc_10000"`
	CoinbaseETHUSD10000   SiteConfig `json:"coinbase_eth_usd_10000"`
	CoinbaseETHUSDDAI5000 SiteConfig `json:"coinbase_eth_usd_dai_5000"`
	BitfinexETHUSDT10000  SiteConfig `json:"bitfinex_eth_usdt_10000"`
	BinanceETHUSDT10000   SiteConfig `json:"binance_eth_usdt_10000"`
	BinanceETHPAX5000     SiteConfig `json:"binance_eth_pax_5000"`
	BinanceETHBUSD10000   SiteConfig `json:"binance_busd_10000"`
	GeminiETHUSD10000     SiteConfig `json:"gemini_eth_usd_10000"`
}

// ExchangeEndpoints ...
type ExchangeEndpoints struct {
	Binance SiteConfig `json:"binance"`
	Houbi   SiteConfig `json:"houbi"`
}

// Authentication config
type Authentication struct {
	KNSecret               string `json:"kn_secret"`
	KNReadOnly             string `json:"kn_readonly"`
	KNConfiguration        string `json:"kn_configuration"`
	KNConfirmConfiguration string `json:"kn_confirm_configuration"`
}

// ContractAddresses ...
type ContractAddresses struct {
	Proxy   common.Address `json:"proxy"`
	Reserve common.Address `json:"reserve"`
	Wrapper common.Address `json:"wrapper"`
	Pricing common.Address `json:"pricing"`
}

type Node struct {
	Main   string   `json:"main"`
	Backup []string `json:"backup"`
}

// Token present for a token
type Token struct {
	Address  string `json:"address"`
	Name     string `json:"name"`
	Decimals int64  `json:"decimals"`
	Internal bool   `json:"internal use"`
	Active   bool   `json:"listed"`
}

type HumanDuration time.Duration

func (d *HumanDuration) UnmarshalJSON(text []byte) error {
	if len(text) < 2 || (text[0] != '"' || text[len(text)-1] != '"') {
		return fmt.Errorf("expect value in double quote")
	}
	v, err := time.ParseDuration(string(text[1 : len(text)-1]))
	if err != nil {
		return err
	}
	*d = HumanDuration(v)
	return nil
}

// TokenSet ..
type TokenSet map[string]Token

// ExchangesTokensDepositAddresses ..
type ExchangesTokensDepositAddresses map[string]DepositAddresses

// DepositAddresses ..
type DepositAddresses map[string]common.Address

// AWSConfig ...
type AWSConfig struct {
	Region                       string `json:"aws_region"`
	AccessKeyID                  string `json:"aws_access_key_id"`
	SecretKey                    string `json:"aws_secret_access_key"`
	Token                        string `json:"aws_token"`
	ExpiredStatDataBucketName    string `json:"aws_expired_stat_data_bucket_name"`
	ExpiredReserveDataBucketName string `json:"aws_expired_reserve_data_bucket_name"`
	LogBucketName                string `json:"aws_log_bucket_name"`
}

// FetcherDelay ...
type FetcherDelay struct {
	OrderBook     HumanDuration `json:"order_book"`
	AuthData      HumanDuration `json:"auth_data"`
	RateFetching  HumanDuration `json:"rate_fetching"`
	BlockFetching HumanDuration `json:"block_fetching"`
	GlobalData    HumanDuration `json:"global_data"`
}

// GasConfig ...
type GasConfig struct {
	PreferUseGasStation     bool   `json:"prefer_use_gas_station"`
	FetchMaxGasCacheSeconds int64  `json:"fetch_max_gas_cache_seconds"`
	GasStationAPIKey        string `json:"gas_station_api_key"`
}

// AppConfig represnet for app configuration
type AppConfig struct {
	Authentication       Authentication `json:"authentication"`
	AWSConfig            AWSConfig      `json:"aws_config"`
	KeyStorePath         string         `json:"keystore_path"`
	Passphrase           string         `json:"passphrase"`
	KeyStoreDepositPath  string         `json:"keystore_deposit_path"`
	PassphraseDeposit    string         `json:"passphrase_deposit"`
	HTTPAPIAddr          string         `json:"http_api_addr"`
	SimulationRunnerAddr string         `json:"http_simulation_runner_addr"`
	GasConfig            GasConfig      `json:"gas_config"`

	ExchangeEndpoints   ExchangeEndpoints               `json:"exchange_endpoints"`
	WorldEndpoints      WorldEndpoints                  `json:"world_endpoints"`
	ContractAddresses   ContractAddresses               `json:"contract_addresses"`
	TokenSet            TokenSet                        `json:"tokens"`
	SettingDB           string                          `json:"setting_db"`
	DataDB              string                          `json:"data_db"`
	DepositAddressesSet ExchangesTokensDepositAddresses `json:"deposit_addresses"`
	Node                Node                            `json:"nodes"`
	HoubiKeystorePath   string                          `json:"keystore_intermediator_path"`
	HuobiPassphrase     string                          `json:"passphrase_intermediate_account"`

	BinanceKey    string `json:"binance_key"`
	BinanceSecret string `json:"binance_secret"`
	BinanceDB     string `json:"binance_db"`

	HuobiKey    string `json:"huobi_key"`
	HuobiSecret string `json:"huobi_secret"`
	HuobiDB     string `json:"huobi_db"`

	FetcherDelay FetcherDelay `json:"fetcher_delay"`
}

// LoadConfig parse json config and return config object
func LoadConfig(file string, ac *AppConfig) error {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, ac)
	if err != nil {
		return err
	}
	return nil
}

// DefaultAppConfig ... set default value, currently only delay fetcher set, other should be explicit set
func DefaultAppConfig() AppConfig {
	return AppConfig{
		FetcherDelay: FetcherDelay{
			OrderBook:     HumanDuration(time.Second * 7),
			AuthData:      HumanDuration(time.Second * 5),
			RateFetching:  HumanDuration(time.Second * 3),
			BlockFetching: HumanDuration(time.Second * 5),
			GlobalData:    HumanDuration(time.Second * 10),
		},
	}
}
