package common

import (
	"encoding/json"
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
)

// PendingObjectType represent type of pending obj in database
//go:generate stringer -type=PendingObjectType -linecomment
type PendingObjectType int

const (
	PendingTypeUnknown PendingObjectType = iota // unknown
	// PendingTypeCreateAsset is used when create an asset
	PendingTypeCreateAsset //create_asset
	// PendingTypeUpdateAsset is used when update an asset
	PendingTypeUpdateAsset // update_asset
	// PendingTypeCreateAssetExchange is used when create an asset exchange
	PendingTypeCreateAssetExchange // create_asset_exchange
	// PendingTypeUpdateAssetExchange is used when update an asset exchange
	PendingTypeUpdateAssetExchange // update_asset_exchange
	// PendingTypeCreateTradingPair is used when create a trading pair
	PendingTypeCreateTradingPair // create_trading_pair
	// PendingTypeUpdateTradingPair is used when update a trading pair
	PendingTypeUpdateTradingPair // update_trading_pair
	// PendingTypeCreateTradingBy is used when create a trading by
	PendingTypeCreateTradingBy // create_trading_by
	// PendingTypeUpdateExchange is used when update exchange
	PendingTypeUpdateExchange // update_exchange
	// PendingTypeChangeAssetAddr is used when update address of an asset
	PendingTypeChangeAssetAddr // change_asset_addr
)

var validSetRateTypes = map[string]SetRate{
	SetRateNotSet.String(): SetRateNotSet,
	ExchangeFeed.String():  ExchangeFeed,
	GoldFeed.String():      GoldFeed,
	BTCFeed.String():       BTCFeed,
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
	Symbol            string           `json:"symbol"`
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

// Asset represents an asset in centralized exchange, eg: ETH, KNC, Bitcoin...
type Asset struct {
	ID                 uint64              `json:"id"`
	Symbol             string              `json:"symbol"`
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
	Created            time.Time           `json:"created"`
	Updated            time.Time           `json:"updated"`
}

// TODO: write custom marshal json for created/updated fields

// PendingObject holds data of pending obj waiting to be confirmed
type PendingObject struct {
	ID      uint64          `json:"id"`
	Created time.Time       `json:"created"`
	Data    json.RawMessage `json:"data"`
}

// CreateAssetExchangeEntry is the configuration of an asset for a specific exchange.
type CreateAssetExchangeEntry struct {
	AssetID           uint64           `json:"asset_id"`
	ExchangeID        uint64           `json:"exchange_id"`
	Symbol            string           `json:"symbol"`
	DepositAddress    ethereum.Address `json:"deposit_address"`
	MinDeposit        float64          `json:"min_deposit"`
	WithdrawFee       float64          `json:"withdraw_fee"`
	TargetRecommended float64          `json:"target_recommended"`
	TargetRatio       float64          `json:"target_ratio"`
}

func (o CreateAssetExchangeEntry) Data() []byte {
	return nil
}

type CreateCreateAssetExchange struct {
	AssetExchanges []CreateAssetExchangeEntry `json:"asset_exchanges" binding:"required,dive"`
}

// UpdateAssetExchangeEntry is the configuration of an asset for a specific exchange to be update
type UpdateAssetExchangeEntry struct {
	ID                uint64            `json:"id"`
	Symbol            *string           `json:"symbol"`
	DepositAddress    *ethereum.Address `json:"deposit_address"`
	MinDeposit        *float64          `json:"min_deposit"`
	WithdrawFee       *float64          `json:"withdraw_fee"`
	TargetRecommended *float64          `json:"target_recommended"`
	TargetRatio       *float64          `json:"target_ratio"`
}

func (o UpdateAssetExchangeEntry) Data() []byte {
	return nil
}

// CreateUpdateAssetExchange present for a UpdateAssetExchange(pending) request
type CreateUpdateAssetExchange struct {
	AssetExchanges []UpdateAssetExchangeEntry `json:"asset_exchanges" binding:"required,dive"`
}

// CreateAssetEntry represents an asset in centralized exchange, eg: ETH, KNC, Bitcoin...
type CreateAssetEntry struct {
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
}

func (o CreateAssetEntry) Data() []byte {
	return nil
}

// CreateCreateAsset present for a CreateAsset(pending) request
type CreateCreateAsset struct {
	AssetInputs []CreateAssetEntry `json:"assets" binding:"required,dive"`
}

// CreateUpdateAsset present for an CreateUpdateAsset request
type CreateUpdateAsset struct {
	Assets []UpdateAssetEntry `json:"assets" binding:"required,dive"`
}

// UpdateAssetEntry
type UpdateAssetEntry struct {
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
}

func (o UpdateAssetEntry) Data() []byte {
	return nil
}

type UpdateExchangeEntry struct {
	ExchangeID      uint64   `json:"exchange_id"`
	TradingFeeMaker *float64 `json:"trading_fee_maker"`
	TradingFeeTaker *float64 `json:"trading_fee_taker"`
	Disable         *bool    `json:"disable"`
}

func (o UpdateExchangeEntry) Data() []byte {
	return nil
}

type CreateUpdateExchange struct {
	Exchanges []UpdateExchangeEntry `json:"exchanges"`
}

// CreateTradingPairEntry represents an trading pair in central exchange.
// this is use when create new trading pair in separate step(not when define Asset), so ExchangeID is required.
type CreateTradingPairEntry struct {
	TradingPair
	ExchangeID uint64 `json:"exchange_id"`
}

func (o CreateTradingPairEntry) Data() []byte {
	return nil
}

// CreateCreateTradingPair present for a CreateTradingPair(pending) request
type CreateCreateTradingPair struct {
	TradingPairs []CreateTradingPairEntry `json:"trading_pairs" binding:"required"`
}

// UpdateTradingPairOpts
type UpdateTradingPairEntry struct {
	ID              uint64   `json:"id"`
	PricePrecision  *uint64  `json:"price_precision"`
	AmountPrecision *uint64  `json:"amount_precision"`
	AmountLimitMin  *float64 `json:"amount_limit_min"`
	AmountLimitMax  *float64 `json:"amount_limit_max"`
	PriceLimitMin   *float64 `json:"price_limit_min"`
	PriceLimitMax   *float64 `json:"price_limit_max"`
	MinNotional     *float64 `json:"min_notional"`
}

func (o UpdateTradingPairEntry) Data() []byte {
	return nil
}

// CreateUpdateTradingPair present for a UpdateTradingPair(pending) request
type CreateUpdateTradingPair struct {
	TradingPairs []UpdateTradingPairEntry `json:"trading_pairs" binding:"required,dive"`
}

// CreateCreateTradingBy present for a CreateTradingBy(pending) request
type CreateCreateTradingBy struct {
	TradingBys []CreateTradingByEntry
}

// CreateTradingByEntry present the information to create a trading by
type CreateTradingByEntry struct {
	AssetID       uint64 `json:"asset_id"`
	TradingPairID uint64 `json:"trading_pair_id"`
}

func (o CreateTradingByEntry) Data() []byte {
	return nil
}

// ChangeAssetAddressEntry present data to create a change asset address
type ChangeAssetAddressEntry struct {
	ID      uint64           `json:"id" binding:"required"`
	Address ethereum.Address `json:"address" binding:"required,isAddress"`
}

func (o ChangeAssetAddressEntry) Data() []byte {
	return nil
}

// CreateChangeAssetAddress present data to create a change asset address
type CreateChangeAssetAddress struct {
	Assets []ChangeAssetAddressEntry `json:"assets" binding:"dive"`
}

// ChangeType represent type of change type entry in list change
//go:generate enumer -type=ChangeType -linecomment -json=true
type ChangeType int

const (
	ChangeTypeUnknown ChangeType = iota // unknown
	// ChangeTypeCreateAsset is used when create an asset
	ChangeTypeCreateAsset // create_asset
	// ChangeTypeUpdateAsset is used when update an asset
	ChangeTypeUpdateAsset // update_asset
	// ChangeTypeCreateAssetExchange is used when create an asset exchange
	ChangeTypeCreateAssetExchange // create_asset_exchange
	// ChangeTypeUpdateAssetExchange is used when update an asset exchange
	ChangeTypeUpdateAssetExchange // update_asset_exchange
	// ChangeTypeCreateTradingPair is used when create a trading pair
	ChangeTypeCreateTradingPair // create_trading_pair
	// ChangeTypeCreateTradingBy is used when create a trading by
	ChangeTypeCreateTradingBy // create_trading_by
	// ChangeTypeUpdateExchange is used when update exchange
	ChangeTypeUpdateExchange // update_exchange
	// ChangeTypeChangeAssetAddr is used when update address of an asset
	ChangeTypeChangeAssetAddr // change_asset_addr
	// ChangeTypeDeleteTradingPair is used to present delete trading pair
	ChangeTypeDeleteTradingPair
	// ChangeTypeDeleteAssetExchange is used in present delete asset exchange object.
	ChangeTypeDeleteAssetExchange
	// ChangeTypeDeleteTradingBy is used in present delete trading_by object.
	ChangeTypeDeleteTradingBy
)

// SettingChangeType interface just make sure that only some of selected type can be put into SettingChange list
type SettingChangeType interface {
	Data() []byte
}

type SettingChangeEntry struct {
	Type ChangeType        `json:"type"`
	Data SettingChangeType `json:"data"`
}

type SettingChange struct {
	ChangeList []SettingChangeEntry `json:"change_list"`
}

type SettingChangeResponse struct {
	ID         uint64               `json:"id"`
	Created    time.Time            `json:"created"`
	ChangeList []SettingChangeEntry `json:"change_list"`
}
