package storage

import (
	ethereum "github.com/ethereum/go-ethereum/common"

	v3 "github.com/KyberNetwork/reserve-data/v3/common"
)

// Interface is the common persistent storage interface of V3 APIs.
type Interface interface {
	SettingReader

	GetExchanges() ([]v3.Exchange, error)
	UpdateExchange(id uint64, opts UpdateExchangeOpts) error

	CreateAssetExchange(exchangeID, assetID uint64, symbol string, depositAddress ethereum.Address,
		minDeposit, withdrawFee, targetRecommended, targetRatio float64) (uint64, error)
	UpdateAssetExchange(id uint64, opts UpdateAssetExchangeOpts) error

	ConfirmCreateAssetExchange(id uint64) error

	ConfirmUpdateAssetExchange(id uint64) error

	CreateAsset(
		symbol, name string,
		address ethereum.Address,
		decimals uint64,
		transferable bool,
		setRate v3.SetRate,
		rebalance bool,
		isQuote bool,
		pwi *v3.AssetPWI,
		rb *v3.RebalanceQuadratic,
		exchanges []v3.AssetExchange,
		target *v3.AssetTarget,
	) (uint64, error)
	UpdateAsset(id uint64, opts UpdateAssetOpts) error
	// ChangeAssetAddress make the current address address of asset old address and set new address as current.
	ChangeAssetAddress(id uint64, address ethereum.Address) error
	UpdateDepositAddress(assetID, exchangeID uint64, address ethereum.Address) error

	UpdateTradingPair(id uint64, opts UpdateTradingPairOpts) error

	// TODO method for batch update PWI
	// TODO method for batch update rebalance quadratic
	// TODO method for batch update exchange configuration
	// TODO meethod for batch update target

	ConfirmCreateAsset(id uint64) error

	ConfirmUpdateAsset(id uint64) error

	ConfirmUpdateExchange(id uint64) error

	ConfirmCreateTradingPair(id uint64) error

	ConfirmUpdateTradingPair(id uint64) error

	ConfirmCreateTradingBy(id uint64) error

	ConfirmChangeAssetAddress(id uint64) error

	CreatePendingObject(interface{}, v3.PendingObjectType) (uint64, error)
	GetPendingObject(uint64, v3.PendingObjectType) (v3.PendingObject, error)
	GetPendingObjects(v3.PendingObjectType) ([]v3.PendingObject, error)
	RejectPendingObject(uint64, v3.PendingObjectType) error
}

// SettingReader is the common interface for reading exchanges, assets configuration.
type SettingReader interface {
	GetAsset(id uint64) (v3.Asset, error)
	GetAssetBySymbol(symbol string) (v3.Asset, error)
	GetAssetExchangeBySymbol(exchangeID uint64, symbol string) (v3.Asset, error)
	GetAssetExchange(id uint64) (v3.AssetExchange, error)
	GetExchange(id uint64) (v3.Exchange, error)
	GetExchangeByName(name string) (v3.Exchange, error)
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

//UpdateExchangeOpts options
type UpdateExchangeOpts struct {
	TradingFeeMaker *float64 `json:"trading_fee_maker"`
	TradingFeeTaker *float64 `json:"trading_fee_taker"`
	Disable         *bool    `json:"disable"`
}

//UpdateAssetOpts update asset options
type UpdateAssetOpts struct {
	Symbol             *string                `json:"symbol"`
	Name               *string                `json:"name"`
	Address            *ethereum.Address      `json:"address"`
	Decimals           *uint64                `json:"decimals"`
	Transferable       *bool                  `json:"transferable"`
	SetRate            *v3.SetRate            `json:"set_rate"`
	Rebalance          *bool                  `json:"rebalance"`
	IsQuote            *bool                  `json:"is_quote"`
	PWI                *v3.AssetPWI           `json:"pwi"`
	RebalanceQuadratic *v3.RebalanceQuadratic `json:"rebalance_quadratic"`
	Target             *v3.AssetTarget        `json:"target"`
}

// UpdateTradingPairOpts ...
type UpdateTradingPairOpts = v3.UpdateTradingPairEntry
