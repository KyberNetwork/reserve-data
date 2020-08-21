package rtypes

// ExchangeID is the name of exchanges of which core will use to rebalance.
//go:generate stringer -type=ExchangeID -linecomment
type ExchangeID uint64

const (
	//Binance is the enumerated key for binance
	Binance ExchangeID = iota + 1 //binance
	//Huobi is the enumerated key for huobi
	Huobi //huobi
	// Binance2 is second binance exchange
	Binance2 // binance_2
)

type AssetID uint64
type TradingPairID uint64
type TradingByID uint64
type AssetExchangeID uint64
type AssetAddressID uint64
type SettingChangeID uint64
type FeedWeightID uint64
