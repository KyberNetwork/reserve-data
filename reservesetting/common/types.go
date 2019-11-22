package common

import (
	"fmt"
	"time"

	ethereum "github.com/ethereum/go-ethereum/common"
)

// Exchange represents a centralized exchange in database.
type Exchange struct {
	ID              uint64  `json:"id"`
	Name            string  `json:"name"`
	TradingFeeMaker float64 `json:"trading_fee_maker"`
	TradingFeeTaker float64 `json:"trading_fee_taker"`
	Disable         bool    `json:"disable"`
}

// SetRate ..
//go:generate stringer -type=SetRate -linecomment
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

var validSetRateTypes = map[string]SetRate{
	SetRateNotSet.String(): SetRateNotSet,
	ExchangeFeed.String():  ExchangeFeed,
	GoldFeed.String():      GoldFeed,
	BTCFeed.String():       BTCFeed,
	USDFeed.String():       USDFeed,
}

// SetRateFromString returns the SetRate value from its string presentation, if exists.
func SetRateFromString(s string) (SetRate, bool) {
	sr, ok := validSetRateTypes[s]
	return sr, ok
}

func (i SetRate) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, i.String())), nil
}
func isString(input []byte) bool {
	return len(input) >= 2 && input[0] == '"' && input[len(input)-1] == '"'
}
func (i *SetRate) UnmarshalJSON(input []byte) error {
	if !isString(input) {
		return fmt.Errorf("not is string")
	}
	r, ok := SetRateFromString(string(input[1 : len(input)-1]))
	if !ok {
		return fmt.Errorf("%s is not a valid SetRate", input)
	}
	*i = r
	return nil
}

// TradingPair is a trading in an exchange.
type TradingPair struct {
	ID              uint64  `json:"id"`
	Base            uint64  `json:"base"`
	Quote           uint64  `json:"quote"`
	PricePrecision  uint64  `json:"price_precision"`
	AmountPrecision uint64  `json:"amount_precision"`
	AmountLimitMin  float64 `json:"amount_limit_min"`
	AmountLimitMax  float64 `json:"amount_limit_max"`
	PriceLimitMin   float64 `json:"price_limit_min"`
	PriceLimitMax   float64 `json:"price_limit_max"`
	MinNotional     float64 `json:"min_notional"`
	ExchangeID      uint64  `json:"-"`
}

// TradingBy is a struct hold trading pair and its asset
type TradingBy struct {
	TradingPairID uint64 `json:"trading_pair_id"`
	AssetID       uint64 `json:"asset_id"`
}

// TradingPairSymbols is a pair of token trading
type TradingPairSymbols struct {
	TradingPair
	BaseSymbol  string `json:"base_symbol"`
	QuoteSymbol string `json:"quote_symbol"`
}

// AssetExchange is the configuration of an asset for a specific exchange.
type AssetExchange struct {
	ID                uint64           `json:"id"`
	AssetID           uint64           `json:"asset_id"`
	ExchangeID        uint64           `json:"exchange_id"`
	Symbol            string           `json:"symbol" binding:"required"`
	DepositAddress    ethereum.Address `json:"deposit_address"`
	MinDeposit        float64          `json:"min_deposit"`
	WithdrawFee       float64          `json:"withdraw_fee"`
	TargetRecommended float64          `json:"target_recommended"`
	TargetRatio       float64          `json:"target_ratio"`
	TradingPairs      []TradingPair    `json:"trading_pairs,omitempty" binding:"dive"`
}

// AssetTarget is the target setting of an asset.
type AssetTarget struct {
	Total              float64 `json:"total"`
	Reserve            float64 `json:"reserve"`
	RebalanceThreshold float64 `json:"rebalance_threshold"`
	TransferThreshold  float64 `json:"transfer_threshold"`
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
	A float64 `json:"a"`
	B float64 `json:"b"`
	C float64 `json:"c"`
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
	ID                 uint64              `json:"id"`
	Symbol             string              `json:"symbol" binding:"required"`
	Name               string              `json:"name"`
	Address            ethereum.Address    `json:"address"`
	OldAddresses       []ethereum.Address  `json:"old_addresses,omitempty"`
	Decimals           uint64              `json:"decimals"`
	Transferable       bool                `json:"transferable"`
	SetRate            SetRate             `json:"set_rate"`
	Rebalance          bool                `json:"rebalance"`
	IsQuote            bool                `json:"is_quote"`
	PWI                *AssetPWI           `json:"pwi,omitempty"`
	RebalanceQuadratic *RebalanceQuadratic `json:"rebalance_quadratic,omitempty"`
	Exchanges          []AssetExchange     `json:"exchanges,omitempty" binding:"dive"`
	Target             *AssetTarget        `json:"target,omitempty"`
	StableParam        StableParam         `json:"stable_param"`
	Created            time.Time           `json:"created"`
	Updated            time.Time           `json:"updated"`
	FeedWeight         *FeedWeight         `json:"feed_weight,omitempty"`
}

// TODO: write custom marshal json for created/updated fields

// CreateAssetExchangeEntry is the configuration of an asset for a specific exchange.
type CreateAssetExchangeEntry struct {
	settingChangeMarker
	AssetID           uint64           `json:"asset_id"`
	ExchangeID        uint64           `json:"exchange_id"`
	Symbol            string           `json:"symbol" binding:"required"`
	DepositAddress    ethereum.Address `json:"deposit_address"`
	MinDeposit        float64          `json:"min_deposit"`
	WithdrawFee       float64          `json:"withdraw_fee"`
	TargetRecommended float64          `json:"target_recommended"`
	TargetRatio       float64          `json:"target_ratio"`
	TradingPairs      []TradingPair    `json:"trading_pairs"`
}

// UpdateAssetExchangeEntry is the configuration of an asset for a specific exchange to be update
type UpdateAssetExchangeEntry struct {
	settingChangeMarker
	ID                uint64            `json:"id"`
	Symbol            *string           `json:"symbol"`
	DepositAddress    *ethereum.Address `json:"deposit_address"`
	MinDeposit        *float64          `json:"min_deposit"`
	WithdrawFee       *float64          `json:"withdraw_fee"`
	TargetRecommended *float64          `json:"target_recommended"`
	TargetRatio       *float64          `json:"target_ratio"`
}

// CreateAssetEntry represents an asset in centralized exchange, eg: ETH, KNC, Bitcoin...
type CreateAssetEntry struct {
	settingChangeMarker
	Symbol             string              `json:"symbol" binding:"required"`
	Name               string              `json:"name" binding:"required"`
	Address            ethereum.Address    `json:"address"`
	OldAddresses       []ethereum.Address  `json:"old_addresses"`
	Decimals           uint64              `json:"decimals"`
	Transferable       bool                `json:"transferable"`
	SetRate            SetRate             `json:"set_rate"`
	Rebalance          bool                `json:"rebalance"`
	IsQuote            bool                `json:"is_quote"`
	PWI                *AssetPWI           `json:"pwi"`
	RebalanceQuadratic *RebalanceQuadratic `json:"rebalance_quadratic"`
	Exchanges          []AssetExchange     `json:"exchanges" binding:"dive"`
	Target             *AssetTarget        `json:"target"`
	StableParam        *StableParam        `json:"stable_param"`
	FeedWeight         *FeedWeight         `json:"feed_weight,omitempty"`
}

// UpdateAssetEntry entry object for update asset
type UpdateAssetEntry struct {
	settingChangeMarker
	AssetID            uint64              `json:"asset_id" binding:"required"`
	Symbol             *string             `json:"symbol"`
	Name               *string             `json:"name"`
	Address            *ethereum.Address   `json:"address"`
	Decimals           *uint64             `json:"decimals"`
	Transferable       *bool               `json:"transferable"`
	SetRate            *SetRate            `json:"set_rate"`
	Rebalance          *bool               `json:"rebalance"`
	IsQuote            *bool               `json:"is_quote"`
	PWI                *AssetPWI           `json:"pwi"`
	RebalanceQuadratic *RebalanceQuadratic `json:"rebalance_quadratic"`
	Target             *AssetTarget        `json:"target"`
	StableParam        *UpdateStableParam  `json:"stable_param"`
	FeedWeight         *FeedWeight         `json:"feed_weight,omitempty"`
}

type UpdateExchangeEntry struct {
	settingChangeMarker
	ExchangeID      uint64   `json:"exchange_id"`
	TradingFeeMaker *float64 `json:"trading_fee_maker"`
	TradingFeeTaker *float64 `json:"trading_fee_taker"`
	Disable         *bool    `json:"disable"`
}

// CreateTradingPairEntry represents an trading pair in central exchange.
// this is use when create new trading pair in separate step(not when define Asset), so ExchangeID is required.
type CreateTradingPairEntry struct {
	settingChangeMarker
	TradingPair
	AssetID    uint64 `json:"asset_id"`
	ExchangeID uint64 `json:"exchange_id"`
}

// ChangeAssetAddressEntry present data to create a change asset address
type ChangeAssetAddressEntry struct {
	settingChangeMarker
	ID      uint64           `json:"id" binding:"required"`
	Address ethereum.Address `json:"address" binding:"required"`
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
	ChangeList []SettingChangeEntry `json:"change_list"`
}

// SettingChangeResponse setting change response
type SettingChangeResponse struct {
	ID         uint64               `json:"id"`
	Created    time.Time            `json:"created"`
	ChangeList []SettingChangeEntry `json:"change_list"`
}

// DeleteTradingPairEntry hold data to delete a trading pair entry
type DeleteTradingPairEntry struct {
	settingChangeMarker
	TradingPairID uint64 `json:"id"`
}

// DeleteAssetExchangeEntry presents data to delete an asset exchange
type DeleteAssetExchangeEntry struct {
	settingChangeMarker
	AssetExchangeID uint64 `json:"id"`
}

type UpdateStableTokenParamsEntry struct {
	settingChangeMarker
	Params map[string]interface{} `json:"params"`
}
