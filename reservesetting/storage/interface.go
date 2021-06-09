package storage

import (
	ethereum "github.com/ethereum/go-ethereum/common"

	"github.com/KyberNetwork/reserve-data/lib/rtypes"
	v3 "github.com/KyberNetwork/reserve-data/reservesetting/common"
)

// Interface is the common persistent storage interface of V3 APIs.
type Interface interface {
	SettingReader
	ControlInfoInterface
	UpdateDepositAddress(assetID rtypes.AssetID, exchangeID rtypes.ExchangeID, address ethereum.Address) error
	UpdateTradingPair(id rtypes.TradingPairID, opts UpdateTradingPairOpts) error

	CreateSettingChange(v3.ChangeCatalog, v3.SettingChange, string) (rtypes.SettingChangeID, error)
	GetSettingChange(id rtypes.SettingChangeID) (v3.SettingChangeResponse, error)
	GetSettingChanges(catalog v3.ChangeCatalog, status v3.ChangeStatus) ([]v3.SettingChangeResponse, error)
	RejectSettingChange(id rtypes.SettingChangeID, keyID string) error
	ConfirmSettingChange(rtypes.SettingChangeID, bool) (*v3.AdditionalDataReturn, error)

	CreatePriceFactor(v3.PriceFactorAtTime) (uint64, error)
	GetPriceFactors(uint64, uint64) ([]v3.PriceFactorAtTime, error)

	UpdateExchange(id rtypes.ExchangeID, updateOpts UpdateExchangeOpts) error
	UpdateFeedStatus(name string, setRate v3.SetRate, enabled bool) error

	SetGeneralData(data v3.GeneralData) (uint64, error)

	UpdateAssetExchangeWithdrawFee(withdrawFee float64, assetExchangeID rtypes.AssetExchangeID) error

	SetPreferGasSource(v v3.PreferGasSource) error

	AddScheduleJob(data v3.ScheduleJobData) (uint64, error)
	RemoveScheduleJob(id uint64) error
}

// SettingReader is the common interface for reading exchanges, assets configuration.
type SettingReader interface {
	GetAsset(id rtypes.AssetID) (v3.Asset, error)
	GetAssetBySymbol(symbol string) (v3.Asset, error)
	GetAssetExchangeBySymbol(exchangeID rtypes.ExchangeID, symbol string) (v3.AssetExchange, error)
	GetAssetExchange(id rtypes.AssetExchangeID) (v3.AssetExchange, error)
	GetExchange(id rtypes.ExchangeID) (v3.Exchange, error)
	GetExchangeByName(name string) (v3.Exchange, error)
	GetExchanges() ([]v3.Exchange, error)
	GetTradingPair(id rtypes.TradingPairID, withDeleted bool) (v3.TradingPairSymbols, error)
	GetTradingPairs(exchangeID rtypes.ExchangeID) ([]v3.TradingPairSymbols, error)
	GetTradingBy(tradingByID rtypes.TradingByID) (v3.TradingBy, error)
	// TODO: check usages of this method to see if it should be replaced with GetDepositAddress(exchangeID, tokenID)
	GetDepositAddresses(exchangeID rtypes.ExchangeID) (map[rtypes.AssetID]ethereum.Address, error)
	GetAssets() ([]v3.Asset, error)
	// GetTransferableAssets returns all assets that the set rate strategy is not not_set.
	GetTransferableAssets() ([]v3.Asset, error)
	GetMinNotional(exchangeID rtypes.ExchangeID, baseID, quoteID rtypes.AssetID) (float64, error)
	GetStableTokenParams() (map[string]interface{}, error)
	// GetFeedConfigurations return all feed configuration
	GetFeedConfigurations() ([]v3.FeedConfiguration, error)
	GetFeedConfiguration(name string, setRate v3.SetRate) (v3.FeedConfiguration, error)

	GetGeneralData(key string) (v3.GeneralData, error)
	GetPreferGasSource() (v3.PreferGasSource, error)

	ApproveSettingChange(keyID string, settingChangeID uint64) error
	DisapproveSettingChange(keyID string, settingChangeID uint64) error

	GetListApprovalSettingChange(settingChangeID uint64) ([]v3.ApprovalSettingChangeInfo, error)
	GetScheduleSettingChange() ([]rtypes.SettingChangeID, error)

	GetAllScheduleJob() ([]v3.ScheduleJobData, error)
	GetScheduleJob(id uint64) (v3.ScheduleJobData, error)
	GetEligibleScheduleJob() ([]v3.ScheduleJobData, error)
}

// ControlInfoInterface ...
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

// UpdateTradingPairOpts ...
type UpdateTradingPairOpts struct {
	ID              rtypes.TradingPairID `json:"id"`
	PricePrecision  *uint64              `json:"price_precision"`
	AmountPrecision *uint64              `json:"amount_precision"`
	AmountLimitMin  *float64             `json:"amount_limit_min"`
	AmountLimitMax  *float64             `json:"amount_limit_max"`
	PriceLimitMin   *float64             `json:"price_limit_min"`
	PriceLimitMax   *float64             `json:"price_limit_max"`
	MinNotional     *float64             `json:"min_notional"`
	StaleThreshold  *float64             `json:"stale_threshold"`
}
