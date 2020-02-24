package storage

import (
	ethereum "github.com/ethereum/go-ethereum/common"

	common2 "github.com/KyberNetwork/reserve-data/common"
	v3 "github.com/KyberNetwork/reserve-data/reservesetting/common"
)

// Interface is the common persistent storage interface of V3 APIs.
type Interface interface {
	SettingReader
	ControlInfoInterface
	UpdateDepositAddress(assetID, exchangeID uint64, address ethereum.Address) error
	UpdateTradingPair(id uint64, opts UpdateTradingPairOpts) error

	CreateSettingChange(v3.ChangeCatalog, v3.SettingChange) (uint64, error)
	GetSettingChange(uint64) (v3.SettingChangeResponse, error)
	GetSettingChanges(catalog v3.ChangeCatalog) ([]v3.SettingChangeResponse, error)
	RejectSettingChange(uint64) error
	ConfirmSettingChange(uint64, bool) error

	CreatePriceFactor(v3.PriceFactorAtTime) (uint64, error)
	GetPriceFactors(uint64, uint64) ([]v3.PriceFactorAtTime, error)

	UpdateExchange(id uint64, updateOpts UpdateExchangeOpts) error
	UpdateFeedStatus(name string, enabled bool) error
}

// SettingReader is the common interface for reading exchanges, assets configuration.
type SettingReader interface {
	GetAsset(id uint64) (v3.Asset, error)
	GetAssetBySymbol(symbol string) (v3.Asset, error)
	GetAssetExchangeBySymbol(exchangeID uint64, symbol string) (v3.Asset, error)
	GetAssetExchange(id uint64) (v3.AssetExchange, error)
	GetExchange(id uint64) (v3.Exchange, error)
	GetExchangeByName(name string) (v3.Exchange, error)
	GetExchanges() ([]v3.Exchange, error)
	GetTradingPair(id uint64) (v3.TradingPairSymbols, error)
	GetTradingPairs(exchangeID uint64) ([]v3.TradingPairSymbols, error)
	GetTradingBy(tradingByID uint64) (v3.TradingBy, error)
	// TODO: check usages of this method to see if it should be replaced with GetDepositAddress(exchangeID, tokenID)
	GetDepositAddresses(exchangeID uint64) (map[common2.AssetID]ethereum.Address, error)
	GetAssets() ([]v3.Asset, error)
	// GetTransferableAssets returns all assets that the set rate strategy is not not_set.
	GetTransferableAssets() ([]v3.Asset, error)
	GetMinNotional(exchangeID, baseID, quoteID uint64) (float64, error)
	GetStableTokenParams() (map[string]interface{}, error)
	// GetFeedConfigurations return all feed configuration
	GetFeedConfigurations() ([]v3.FeedConfiguration, error)
	GetFeedConfiguration(name string) (v3.FeedConfiguration, error)
}

type ControlInfoInterface interface {
	GetSetRateStatus() (bool, error)
	SetSetRateStatus(status bool) error

	GetRebalanceStatus() (bool, error)
	SetRebalanceStatus(status bool) error
}

// UpdateAssetExchangeOpts these type match user type define in common package so we just need to make an alias here
// in case they did not later, we need to redefine the structure here, or review this again.
type UpdateAssetExchangeOpts = v3.UpdateAssetExchangeEntry

// UpdateExchangeOpts options
type UpdateExchangeOpts = v3.UpdateExchangeEntry

// UpdateAssetOpts update asset options
type UpdateAssetOpts = v3.UpdateAssetEntry

// UpdateTradingPairOpts
type UpdateTradingPairOpts struct {
	ID              uint64   `json:"id"`
	PricePrecision  *uint64  `json:"price_precision"`
	AmountPrecision *uint64  `json:"amount_precision"`
	AmountLimitMin  *float64 `json:"amount_limit_min"`
	AmountLimitMax  *float64 `json:"amount_limit_max"`
	PriceLimitMin   *float64 `json:"price_limit_min"`
	PriceLimitMax   *float64 `json:"price_limit_max"`
	MinNotional     *float64 `json:"min_notional"`
}
