package common

import (
	"encoding/json"
	"time"

	"github.com/KyberNetwork/reserve-data/common/ethutil"
	ethereum "github.com/ethereum/go-ethereum/common"

	"github.com/KyberNetwork/reserve-data/lib/rtypes"
)

// Exchange represents a centralized exchange in database.
type Exchange struct {
	ID              rtypes.ExchangeID `json:"id"`
	Name            string            `json:"name"`
	TradingFeeMaker float64           `json:"trading_fee_maker"`
	TradingFeeTaker float64           `json:"trading_fee_taker"`
	Disable         bool              `json:"disable"`
}

// SetRate ..
//go:generate enumer -type=SetRate -linecomment -json=true -sql=true
type SetRate int

const (
	// SetRateNotSet indicates that set rate is not enabled for the asset.
	SetRateNotSet SetRate = iota // not_set
	// ExchangeFeed is used when asset rate is set from fetching historical prices from exchanges.
	ExchangeFeed // exchange_feed
	// GoldFeed is used when asset rate is set from fetching gold prices.
	GoldFeed // gold_feed
	// BTCFeed is used when asset rate is set from fetching Bitcoin prices.
	BTCFeed // btc_feed
	// USDFeed is used when asset rate is set from fetching usd prices.
	USDFeed // usd_feed
)

// TradingPair is a trading in an exchange.
type TradingPair struct {
	ID              rtypes.TradingPairID `json:"id"`
	Base            rtypes.AssetID       `json:"base"`
	Quote           rtypes.AssetID       `json:"quote"`
	PricePrecision  uint64               `json:"price_precision"`
	AmountPrecision uint64               `json:"amount_precision"`
	AmountLimitMin  float64              `json:"amount_limit_min"`
	AmountLimitMax  float64              `json:"amount_limit_max"`
	PriceLimitMin   float64              `json:"price_limit_min"`
	PriceLimitMax   float64              `json:"price_limit_max"`
	MinNotional     float64              `json:"min_notional"`
	ExchangeID      rtypes.ExchangeID    `json:"-"`
	BaseSymbol      string               `json:"base_symbol"`
	QuoteSymbol     string               `json:"quote_symbol"`
	StaleThreshold  float64              `json:"stale_threshold"`
}

// TradingBy is a struct hold trading pair and its asset
type TradingBy struct {
	TradingPairID rtypes.TradingPairID `json:"trading_pair_id"`
	AssetID       rtypes.AssetID       `json:"asset_id"`
}

// TradingPairSymbols is a pair of token trading
type TradingPairSymbols struct {
	TradingPair
	BaseSymbol  string `json:"base_symbol"`
	QuoteSymbol string `json:"quote_symbol"`
}

// AssetExchange is the configuration of an asset for a specific exchange.
type AssetExchange struct {
	ID                rtypes.AssetExchangeID `json:"id"`
	AssetID           rtypes.AssetID         `json:"asset_id"`
	ExchangeID        rtypes.ExchangeID      `json:"exchange_id"`
	Symbol            string                 `json:"symbol" binding:"required"`
	DepositAddress    ethereum.Address       `json:"deposit_address"`
	MinDeposit        float64                `json:"min_deposit"`
	WithdrawFee       float64                `json:"withdraw_fee"`
	TargetRecommended float64                `json:"target_recommended"`
	TargetRatio       float64                `json:"target_ratio"`
	TradingPairs      []TradingPair          `json:"trading_pairs,omitempty" binding:"dive"`
}

// AssetTarget is the target setting of an asset.
type AssetTarget struct {
	Total                float64 `json:"total"`
	Reserve              float64 `json:"reserve"`
	RebalanceThreshold   float64 `json:"rebalance_threshold"`
	TransferThreshold    float64 `json:"transfer_threshold"`
	MinWithdrawThreshold float64 `json:"min_withdraw_threshold"`
}

// PWIEquation is a PWI equation. An asset will have 2 PWI equation: ask and bid.
type PWIEquation struct {
	A                   float64 `json:"a"`
	B                   float64 `json:"b"`
	C                   float64 `json:"c"`
	MinMinSpread        float64 `json:"min_min_spread"`
	PriceMultiplyFactor float64 `json:"price_multiply_factor"`
}

// AssetPWI is the PWI configuration of an asset.
type AssetPWI struct {
	Ask PWIEquation `json:"ask"`
	Bid PWIEquation `json:"bid"`
}

// RebalanceQuadratic is params of quadratic equation
type RebalanceQuadratic struct {
	SizeA       float64 `json:"size_a"`
	SizeB       float64 `json:"size_b"`
	SizeC       float64 `json:"size_c"`
	PriceA      float64 `json:"price_a"`
	PriceB      float64 `json:"price_b"`
	PriceC      float64 `json:"price_c"`
	PriceOffset float64 `json:"price_offset"`
}

// StableParam is params of stablize action
type StableParam struct {
	PriceUpdateThreshold float64 `json:"price_update_threshold"`
	AskSpread            float64 `json:"ask_spread"`
	BidSpread            float64 `json:"bid_spread"`
	SingleFeedMaxSpread  float64 `json:"single_feed_max_spread"`
	MultipleFeedsMaxDiff float64 `json:"multiple_feeds_max_diff"`
}

// UpdateStableParam is the placeholder for updating stable params
type UpdateStableParam struct {
	PriceUpdateThreshold *float64 `json:"price_update_threshold"`
	AskSpread            *float64 `json:"ask_spread"`
	BidSpread            *float64 `json:"bid_spread"`
	SingleFeedMaxSpread  *float64 `json:"single_feed_max_spread"`
	MultipleFeedsMaxDiff *float64 `json:"multiple_feeds_max_diff"`
}

// FeedWeight is feed weight configuration for stable token feed like
type FeedWeight map[string]float64

// Asset represents an asset in centralized exchange, eg: ETH, KNC, Bitcoin...
type Asset struct {
	ID                    rtypes.AssetID      `json:"id"`
	Symbol                string              `json:"symbol" binding:"required"`
	Name                  string              `json:"name"`
	Address               ethereum.Address    `json:"address"`
	OldAddresses          []ethereum.Address  `json:"old_addresses,omitempty"`
	Decimals              uint64              `json:"decimals"`
	Transferable          bool                `json:"transferable"`
	SetRate               SetRate             `json:"set_rate"`
	Rebalance             bool                `json:"rebalance"`
	IsQuote               bool                `json:"is_quote"`
	PWI                   *AssetPWI           `json:"pwi,omitempty"`
	RebalanceQuadratic    *RebalanceQuadratic `json:"rebalance_quadratic,omitempty"`
	Exchanges             []AssetExchange     `json:"exchanges,omitempty" binding:"dive"`
	Target                *AssetTarget        `json:"target,omitempty"`
	StableParam           StableParam         `json:"stable_param"`
	Created               time.Time           `json:"created"`
	Updated               time.Time           `json:"updated"`
	FeedWeight            *FeedWeight         `json:"feed_weight,omitempty"`
	NormalUpdatePerPeriod float64             `json:"normal_update_per_period"`
	MaxImbalanceRatio     float64             `json:"max_imbalance_ratio"`
	OrderDurationMillis   uint64              `json:"order_duration_millis"`
	PriceETHAmount        float64             `json:"price_eth_amount"`
	ExchangeETHAmount     float64             `json:"exchange_eth_amount"`
	SanityInfo            SanityInfo          `json:"sanity_info"`
}

// IsMainQuoteAsset check if token is main coin
func (a Asset) IsMainQuoteAsset() bool {
	return a.ID == 1
}

// IsNetworkAsset check if token is main coin
func (a Asset) IsNetworkAsset() bool {
	return ethutil.IsEthereumAddress(a.Address)
}

// TODO: write custom marshal json for created/updated fields

// CreateAssetExchangeEntry is the configuration of an asset for a specific exchange.
type CreateAssetExchangeEntry struct {
	settingChangeMarker
	AssetID           rtypes.AssetID    `json:"asset_id"`
	ExchangeID        rtypes.ExchangeID `json:"exchange_id"`
	Symbol            string            `json:"symbol" binding:"required"`
	DepositAddress    ethereum.Address  `json:"deposit_address"`
	MinDeposit        float64           `json:"min_deposit"`
	WithdrawFee       float64           `json:"withdraw_fee"`
	TargetRecommended float64           `json:"target_recommended"`
	TargetRatio       float64           `json:"target_ratio"`
	TradingPairs      []TradingPair     `json:"trading_pairs"`
}

// UpdateAssetExchangeEntry is the configuration of an asset for a specific exchange to be update
type UpdateAssetExchangeEntry struct {
	settingChangeMarker
	ID                rtypes.AssetExchangeID `json:"id"`
	Symbol            *string                `json:"symbol"`
	DepositAddress    *ethereum.Address      `json:"deposit_address"`
	MinDeposit        *float64               `json:"min_deposit"`
	TargetRecommended *float64               `json:"target_recommended"`
	TargetRatio       *float64               `json:"target_ratio"`
}

// CreateAssetEntry represents an asset in centralized exchange, eg: ETH, KNC, Bitcoin...
type CreateAssetEntry struct {
	settingChangeMarker
	Symbol                string              `json:"symbol" binding:"required"`
	Name                  string              `json:"name" binding:"required"`
	Address               ethereum.Address    `json:"address"`
	OldAddresses          []ethereum.Address  `json:"old_addresses"`
	Decimals              uint64              `json:"decimals"`
	Transferable          bool                `json:"transferable"`
	SetRate               SetRate             `json:"set_rate"`
	Rebalance             bool                `json:"rebalance"`
	IsQuote               bool                `json:"is_quote"`
	IsEnabled             bool                `json:"is_enabled"`
	PWI                   *AssetPWI           `json:"pwi"`
	RebalanceQuadratic    *RebalanceQuadratic `json:"rebalance_quadratic"`
	Exchanges             []AssetExchange     `json:"exchanges" binding:"dive"`
	Target                *AssetTarget        `json:"target"`
	StableParam           *StableParam        `json:"stable_param"`
	FeedWeight            *FeedWeight         `json:"feed_weight,omitempty"`
	NormalUpdatePerPeriod float64             `json:"normal_update_per_period"`
	MaxImbalanceRatio     float64             `json:"max_imbalance_ratio"`
	OrderDurationMillis   uint64              `json:"order_duration_millis"`
	PriceETHAmount        float64             `json:"price_eth_amount"`
	ExchangeETHAmount     float64             `json:"exchange_eth_amount"`
	SanityInfo            SanityInfo          `json:"sanity_info"`
}

type SanityInfo struct {
	Provider  string   `json:"provider"`
	Threshold float64  `json:"threshold"`
	Path      []string `json:"path"`
}

// UpdateAssetEntry entry object for update asset
type UpdateAssetEntry struct {
	settingChangeMarker
	AssetID               rtypes.AssetID      `json:"asset_id" binding:"required"`
	Symbol                *string             `json:"symbol"`
	Name                  *string             `json:"name"`
	Address               *ethereum.Address   `json:"address"`
	Decimals              *uint64             `json:"decimals"`
	Transferable          *bool               `json:"transferable"`
	SetRate               *SetRate            `json:"set_rate"`
	Rebalance             *bool               `json:"rebalance"`
	IsQuote               *bool               `json:"is_quote"`
	IsEnabled             *bool               `json:"is_enabled"`
	PWI                   *AssetPWI           `json:"pwi"`
	RebalanceQuadratic    *RebalanceQuadratic `json:"rebalance_quadratic"`
	Target                *AssetTarget        `json:"target"`
	StableParam           *UpdateStableParam  `json:"stable_param"`
	FeedWeight            *FeedWeight         `json:"feed_weight,omitempty"`
	NormalUpdatePerPeriod *float64            `json:"normal_update_per_period"`
	MaxImbalanceRatio     *float64            `json:"max_imbalance_ratio"`
	OrderDurationMillis   *uint64             `json:"order_duration_millis"`
	PriceETHAmount        *float64            `json:"price_eth_amount"`
	ExchangeETHAmount     *float64            `json:"exchange_eth_amount"`
	SanityInfo            SanityInfo          `json:"sanity_info"`
}

type UpdateExchangeEntry struct {
	settingChangeMarker
	ExchangeID      rtypes.ExchangeID `json:"exchange_id"`
	TradingFeeMaker *float64          `json:"trading_fee_maker"`
	TradingFeeTaker *float64          `json:"trading_fee_taker"`
	Disable         *bool             `json:"disable"`
}

// CreateTradingPairEntry represents an trading pair in central exchange.
// this is use when create new trading pair in separate step(not when define Asset), so ExchangeID is required.
type CreateTradingPairEntry struct {
	settingChangeMarker
	TradingPair
	AssetID    rtypes.AssetID    `json:"asset_id"`
	ExchangeID rtypes.ExchangeID `json:"exchange_id"`
}

type UpdateTradingPairEntry struct {
	settingChangeMarker
	TradingPairID  rtypes.TradingPairID `json:"trading_pair_id"`
	StaleThreshold *float64             `json:"stale_threshold"` // currently, this is the only field we allow to update using API
	// other fields of trading pair is update from exchange automatically
}

// ChangeAssetAddressEntry present data to create a change asset address
type ChangeAssetAddressEntry struct {
	settingChangeMarker
	ID      rtypes.AssetID   `json:"id" binding:"required"`
	Address ethereum.Address `json:"address" binding:"required"`
}

// SetFeedConfigurationEntry data to set feed config
type SetFeedConfigurationEntry struct {
	settingChangeMarker
	Name                 string   `json:"name" binding:"required"`
	SetRate              SetRate  `json:"set_rate"`
	Enabled              *bool    `json:"enabled"`
	BaseVolatilitySpread *float64 `json:"base_volatility_spread"`
	NormalSpread         *float64 `json:"normal_spread"`
}

// ChangeCatalog represent catalog the change list belong to, each catalog keep track pending change independent
//go:generate enumer -type=ChangeCatalog -linecomment -json=true
type ChangeCatalog int

const (
	ChangeCatalogSetTarget          ChangeCatalog = iota // set_target
	ChangeCatalogSetPWIS                                 // set_pwis
	ChangeCatalogStableToken                             // set_stable_token
	ChangeCatalogRebalanceQuadratic                      // set_rebalance_quadratic
	ChangeCatalogUpdateExchange                          // update_exchange
	ChangeCatalogMain                                    // main
	ChangeCatalogFeedConfiguration                       // set_feed_configuration
	ChangeCatalogTradingPair                             // update_tpair
)

// ChangeType represent type of change type entry in list change
//go:generate enumer -type=ChangeType -linecomment -json=true
type ChangeType int

const (
	// ChangeTypeCreateAsset is used when create an asset
	ChangeTypeCreateAsset ChangeType = iota // create_asset
	// ChangeTypeUpdateAsset is used when update an asset
	ChangeTypeUpdateAsset // update_asset
	// ChangeTypeCreateAssetExchange is used when create an asset exchange
	ChangeTypeCreateAssetExchange // create_asset_exchange
	// ChangeTypeUpdateAssetExchange is used when update an asset exchange
	ChangeTypeUpdateAssetExchange // update_asset_exchange
	// ChangeTypeCreateTradingPair is used when create a trading pair
	ChangeTypeCreateTradingPair // create_trading_pair
	// ChangeTypeUpdateExchange is used when update exchange
	ChangeTypeUpdateExchange // update_exchange
	// ChangeTypeChangeAssetAddr is used when update address of an asset
	ChangeTypeChangeAssetAddr // change_asset_addr
	// ChangeTypeDeleteTradingPair is used to present delete trading pair
	ChangeTypeDeleteTradingPair // delete_trading_pair
	// ChangeTypeDeleteAssetExchange is used in present delete asset exchange object.
	ChangeTypeDeleteAssetExchange // delete_asset_exchange
	// ChangeTypeUpdateStableTokenParams is used in present update stable token params
	ChangeTypeUpdateStableTokenParams // update_stable_token_params
	// ChangeTypeSetFeedConfiguration is used when set feed configuration
	ChangeTypeSetFeedConfiguration // set_feed_configuration
	// ChangeTypeUpdateTradingPair is use when update trading pair
	ChangeTypeUpdateTradingPair // update_trading_pair
)

// ChangeStatus represent status of change
//go:generate enumer -type=ChangeStatus -linecomment -json=true
type ChangeStatus int

const (
	ChangeStatusPending  ChangeStatus = iota // pending
	ChangeStatusAccepted                     // accepted
	ChangeStatusRejected                     // rejected
)

// SettingChangeType interface just make sure that only some of selected type can be put into SettingChange list
type SettingChangeType interface {
	nope()
}
type settingChangeMarker struct {
}

func (s settingChangeMarker) nope() {

}

// SettingChangeEntry present a an entry of change
type SettingChangeEntry struct {
	Type ChangeType        `json:"type"`
	Data SettingChangeType `json:"data"`
}

// SettingChange present for setting change request
type SettingChange struct {
	ChangeList    []SettingChangeEntry `json:"change_list"`
	ScheduledTime uint64               `json:"scheduled_time"`
	Message       string               `json:"message"`
}

// SettingChangeResponse setting change response
type SettingChangeResponse struct {
	ID            rtypes.SettingChangeID      `json:"id"`
	Created       time.Time                   `json:"created"`
	ChangeList    []SettingChangeEntry        `json:"change_list"`
	ScheduledTime uint64                      `json:"scheduled_time,omitempty"`
	Proposer      string                      `json:"proposer,omitempty"`
	Rejector      string                      `json:"rejector,omitempty"`
	ListApproval  []ApprovalSettingChangeInfo `json:"list_approval,omitempty"`
}

// DeleteTradingPairEntry hold data to delete a trading pair entry
type DeleteTradingPairEntry struct {
	settingChangeMarker
	TradingPairID rtypes.TradingPairID `json:"id"`
}

// DeleteAssetExchangeEntry presents data to delete an asset exchange
type DeleteAssetExchangeEntry struct {
	settingChangeMarker
	AssetExchangeID rtypes.AssetExchangeID `json:"id"`
}

// UpdateStableTokenParamsEntry ...
type UpdateStableTokenParamsEntry struct {
	settingChangeMarker
	Params map[string]interface{} `json:"params"`
}

// FeedConfiguration feed configuration
type FeedConfiguration struct {
	Name                 string  `json:"name" db:"name"`
	SetRate              SetRate `json:"set_rate" db:"set_rate"`
	Enabled              bool    `json:"enabled" db:"enabled"`
	BaseVolatilitySpread float64 `json:"base_volatility_spread" db:"base_volatility_spread"`
	NormalSpread         float64 `json:"normal_spread" db:"normal_spread"`
}

// GeneralData ...
type GeneralData struct {
	Key   string `db:"key"`
	Value string `db:"value"`
}

// PreferGasSource keep prefer gas source name
type PreferGasSource struct {
	Name string `json:"name" binding:"required"`
}

// AdditionalDataReturn ...
type AdditionalDataReturn struct {
	AddedTradingPairs []rtypes.TradingPairID
}

// ApprovalSettingChangeInfo ...
type ApprovalSettingChangeInfo struct {
	KeyID     string    `json:"key_id" db:"key_id"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
}

// ScheduledJobData ...
type ScheduledJobData struct {
	ID            uint64          `json:"id"`
	Endpoint      string          `json:"endpoint"`
	HTTPMethod    string          `json:"http_method"`
	Data          json.RawMessage `json:"data"`
	ScheduledTime time.Time       `json:"scheduled_time"`
}
