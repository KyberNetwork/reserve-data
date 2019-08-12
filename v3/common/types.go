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

type pendingObject struct {
	ID      uint64          `json:"id"`
	Created time.Time       `json:"created"`
	Data    json.RawMessage `json:"data"`
}

// CreateAsset hold state of being create Asset and waiting for confirm to be Asset.
type CreateAsset pendingObject

// CreateAssetExchange holds state of being create AssetExchange and waiting for confirm to be AssetExchange
type CreateAssetExchange pendingObject

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

type CreateCreateAssetExchange struct {
	AssetExchanges []CreateAssetExchangeEntry `json:"asset_exchanges" binding:"required,dive"`
}

// UpdateAssetExchange holds state of being update AssetExchanges and waiting for confirm.
type UpdateAssetExchange pendingObject

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

// CreateCreateAsset present for a CreateAsset(pending) request
type CreateCreateAsset struct {
	AssetInputs []CreateAssetEntry `json:"assets" binding:"required,dive"`
}

// UpdateAsset hold state of being update Asset and waiting for confirm to apply.
type UpdateAsset pendingObject

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

type UpdateExchangeEntry struct {
	ExchangeID      uint64   `json:"exchange_id"`
	TradingFeeMaker *float64 `json:"trading_fee_maker"`
	TradingFeeTaker *float64 `json:"trading_fee_taker"`
	Disable         *bool    `json:"disable"`
}

type CreateUpdateExchange struct {
	Exchanges []UpdateExchangeEntry `json:"exchanges"`
}

// UpdateExchange hold state of being update Exchange and waiting for confirm to apply.
type UpdateExchange pendingObject

// CreateTradingPair hold state of being create trading pair and waiting for confirm to apply, hold origin json content.
type CreateTradingPair pendingObject

// CreateTradingPairEntry represents an trading pair in central exchange.
// this is use when create new trading pair in separate step(not when define Asset), so ExchangeID is required.
type CreateTradingPairEntry struct {
	TradingPair
	ExchangeID uint64 `json:"exchange_id"`
}

// CreateCreateTradingPair present for a CreateTradingPair(pending) request
type CreateCreateTradingPair struct {
	TradingPairs []CreateTradingPairEntry `json:"trading_pairs" binding:"required"`
}

// CreateTradingPair hold state of being create trading pair and waiting for confirm to apply, hold origin json content.
type UpdateTradingPair pendingObject

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

// CreateUpdateTradingPair present for a UpdateTradingPair(pending) request
type CreateUpdateTradingPair struct {
	TradingPairs []UpdateTradingPairEntry `json:"trading_pairs" binding:"required,dive"`
}

// CreateTradingBy hold state of being create trading by and waiting for confirm to apply, hold origin json content.
type CreateTradingBy pendingObject

// CreateCreateTradingBy present for a CreateTradingBy(pending) request
type CreateCreateTradingBy struct {
	TradingBys []CreateTradingByEntry
}

// CreateTradingByEntry present the information to create a trading by
type CreateTradingByEntry struct {
	AssetID       uint64 `json:"asset_id"`
	TradingPairID uint64 `json:"trading_pair_id"`
}

// ChangeAssetAddress hold state of being changed address of asset and waiting for confirm to apply.
type ChangeAssetAddress pendingObject

// ChangeAssetAddressEntry present data to create a change asset address
type ChangeAssetAddressEntry struct {
	ID      uint64 `json:"id" binding:"required"`
	Address string `json:"address" binding:"required,isAddress"`
}

// CreateChangeAssetAddress present data to create a change asset address
type CreateChangeAssetAddress struct {
	Assets []ChangeAssetAddressEntry `json:"assets" binding:"dive"`
}
