package storage

import (
	ethereum "github.com/ethereum/go-ethereum/common"

	v3 "github.com/KyberNetwork/reserve-data/v3/common"
)

// Interface is the common persistent storage interface of V3 APIs.
type Interface interface {
	SettingReader
	UpdateDepositAddress(assetID, exchangeID uint64, address ethereum.Address) error
	UpdateTradingPair(id uint64, opts UpdateTradingPairOpts) error

	// TODO method for batch update PWI
	// TODO method for batch update rebalance quadratic
	// TODO method for batch update exchange configuration
	// TODO meethod for batch update target

	CreatePendingObject(interface{}, v3.PendingObjectType) (uint64, error)
	GetPendingObject(uint64, v3.PendingObjectType) (v3.PendingObject, error)
	GetPendingObjects(v3.PendingObjectType) ([]v3.PendingObject, error)
	RejectPendingObject(uint64, v3.PendingObjectType) error
	ConfirmPendingObject(uint64, v3.PendingObjectType) error

	CreateSettingChange(v3.SettingChange) (uint64, error)
	GetSettingChange(uint64) (v3.SettingChangeResponse, error)
	GetSettingChanges() ([]v3.SettingChangeResponse, error)
	RejectSettingChange(uint64) error
	ConfirmSettingChange(uint64, bool) error
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
	GetDepositAddresses(exchangeID uint64) (map[string]ethereum.Address, error)
	GetAssets() ([]v3.Asset, error)
	// GetTransferableAssets returns all assets that the set rate strategy is not not_set.
	GetTransferableAssets() ([]v3.Asset, error)
	GetMinNotional(exchangeID, baseID, quoteID uint64) (float64, error)
}

// UpdateAssetExchangeOpts these type match user type define in common package so we just need to make an alias here
// in case they did not later, we need to redefine the structure here, or review this again.
type UpdateAssetExchangeOpts = v3.UpdateAssetExchangeEntry

// UpdateExchangeOpts options
type UpdateExchangeOpts = v3.UpdateExchangeEntry

// UpdateAssetOpts update asset options
type UpdateAssetOpts = v3.UpdateAssetEntry

// UpdateTradingPairOpts ...
type UpdateTradingPairOpts = v3.UpdateTradingPairEntry
