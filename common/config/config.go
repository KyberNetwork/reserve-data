package config

import (
	"encoding/json"
	"io/ioutil"
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

// ExchangePoints ...
type ExchangePoints struct {
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

// AppConfig represnet for app configuration
type AppConfig struct {
	Authentication
	KeyStorePath        string `json:"keystore"`
	Passphrase          string `json:"passphrase"`
	KeyStoreDepositPath string `json:"deposit_keystore"`
	PassphraseDeposit   string `json:"passphrase_deposit"`

	ExchangePoints ExchangePoints `json:"exchange_points"`
	WorldEndpoints WorldEndpoints `json:"world_endpoints"`
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
