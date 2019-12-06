package config

import (
	"encoding/json"
	"io/ioutil"

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

	CoinbaseBTC SiteConfig `json:"coinbase_btc"`
	GeminiBTC   SiteConfig `json:"gemini_btc"`

	CoinbaseUSDC SiteConfig `json:"coinbase_usdc"`
	BinanceUSDC  SiteConfig `json:"binance_usdc"`
	CoinbaseUSD  SiteConfig `json:"coinbase_usd"`
	CoinbaseDAI  SiteConfig `json:"coinbase_dai"`
	HitDai       SiteConfig `json:"hit_dai"`

	BitFinexUSDT SiteConfig `json:"bit_finex_usdt"`
	BinanceUSDT  SiteConfig `json:"binance_usdt"`
	BinancePAX   SiteConfig `json:"binance_pax"`
	BinanceTUSD  SiteConfig `json:"binance_tusd"`
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

// TokenSet ..
type TokenSet map[string]Token

// DepositAddressesSet ..
type DepositAddressesSet map[string]DepositAddresses

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

// AppConfig represnet for app configuration
type AppConfig struct {
	Authentication      Authentication `json:"authentication"`
	AWSConfig           AWSConfig      `json:"aws_config"`
	KeyStorePath        string         `json:"keystore_path"`
	Passphrase          string         `json:"passphrase"`
	KeyStoreDepositPath string         `json:"keystore_deposit_path"`
	PassphraseDeposit   string         `json:"passphrase_deposit"`

	ExchangeEndpoints   ExchangeEndpoints   `json:"exchange_endpoints"`
	WorldEndpoints      WorldEndpoints      `json:"world_endpoints"`
	ContractAddresses   ContractAddresses   `json:"contract_addresses"`
	TokenSet            TokenSet            `json:"tokens"`
	SettingDB           string              `json:"setting_db"`
	DepositAddressesSet DepositAddressesSet `json:"deposit_addresses"`
	Node                Node                `json:"nodes"`
	HoubiKeystorePath   string              `json:"keystore_intermediator_path"`
	HuobiPassphrase     string              `json:"passphrase_intermediate_account"`

	BinanceKey    string `json:"binance_key"`
	BinanceSecret string `json:"binance_secret"`
	BinanceDB     string `json:"binance_db"`

	HuobiKey    string `json:"huobi_key"`
	HuobiSecret string `json:"huobi_secret"`
	HuobiDB     string `json:"huobi_db"`
}

// LoadConfig parse json config and return config object
func LoadConfig(file string) (AppConfig, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return AppConfig{}, err
	}
	var a AppConfig
	err = json.Unmarshal(data, &a)
	if err != nil {
		return AppConfig{}, err
	}
	return a, nil
}
