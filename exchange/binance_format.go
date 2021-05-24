package exchange

import (
	"github.com/KyberNetwork/reserve-data/common"
	rtypes "github.com/KyberNetwork/reserve-data/lib/rtypes"
)

// Binaprice is binance order book
type Binaprice struct {
	Quantity float64 `json:"size"`
	Rate     float64 `json:"price"`
}

// Binaresp response from binance
type Binaresp struct {
	LastUpdatedID int64       `json:"lastUpdateId"`
	Code          int         `json:"code"`
	Msg           string      `json:"msg"`
	Bids          []Binaprice `json:"bids"`
	Asks          []Binaprice `json:"asks"`
	Timestamp     uint64      `json:"timestamp"`
}

// Binainfo binance account info
type Binainfo struct {
	Code             int       `json:"code"`
	Msg              string    `json:"msg"`
	MakerCommission  int64     `json:"makerCommission"`
	TakerCommission  int64     `json:"takerCommission"`
	BuyerCommission  int64     `json:"buyerCommission"`
	SellerCommission int64     `json:"sellerCommission"`
	CanTrade         bool      `json:"canTrade"`
	CanWithdraw      bool      `json:"canWithdraw"`
	CanDeposit       bool      `json:"canDeposit"`
	UpdateTime       uint64    `json:"updateTime"`
	AccountType      string    `json:"accountType"`
	Balances         []Balance `json:"balances"`
	Permissions      []string  `json:"permissions"`
}

// Balance of account
type Balance struct {
	Asset  string `json:"asset"`
	Free   string `json:"free"`
	Locked string `json:"locked"`
}

type BinanceMainAccountBalance struct {
	AssetID rtypes.AssetID `json:"asset_id"`
	Symbol  string         `json:"symbol"`
	Free    string         `json:"free"`
	Locked  string         `json:"locked"`
}

// CrossMarginAccountDetails ...
type CrossMarginAccountDetails struct {
	BorrowEnabled       bool                           `json:"borrowEnabled"`
	MarginLevel         string                         `json:"marginLevel"`
	TotalAssetOfBtc     string                         `json:"totalAssetOfBtc"`
	TotalLiabilityOfBtc string                         `json:"totalLiabilityOfBtc"`
	TotalNetAssetOfBtc  string                         `json:"totalNetAssetOfBtc"`
	TradeEnabled        bool                           `json:"tradeEnabled"`
	TransferEnabled     bool                           `json:"transferEnabled"`
	UserAssets          []common.RawAssetMarginBalance `json:"userAssets"`
}

type FilterLimit struct {
	FilterType  string `json:"filterType"`
	MinPrice    string `json:"minPrice"`
	MaxPrice    string `json:"maxPrice"`
	MinQuantity string `json:"minQty"`
	MaxQuantity string `json:"maxQty"`
	StepSize    string `json:"stepSize"`
	TickSize    string `json:"tickSize"`
	MinNotional string `json:"minNotional"`
}

type BinanceSymbol struct {
	Symbol              string        `json:"symbol"`
	BaseAssetPrecision  int           `json:"baseAssetPrecision"`
	QuoteAssetPrecision int           `json:"quoteAssetPrecision"`
	Filters             []FilterLimit `json:"filters"`
}

type BinanceExchangeInfo struct {
	Symbols []BinanceSymbol
}

type Binatrade struct {
	Symbol        string `json:"symbol"`
	OrderID       uint64 `json:"orderId"`
	ClientOrderID string `json:"clientOrderId"`
	TransactTime  uint64 `json:"transactTime"`
}

type Binawithdraw struct {
	Success bool   `json:"success"`
	Msg     string `json:"msg"`
	ID      string `json:"id"`
}

type Binaorder struct {
	Code          int    `json:"code"`
	Msg           string `json:"msg"`
	Symbol        string `json:"symbol"`
	OrderID       uint64 `json:"orderId"`
	ClientOrderID string `json:"clientOrderId"`
	Price         string `json:"price"`
	OrigQty       string `json:"origQty"`
	ExecutedQty   string `json:"executedQty"`
	Status        string `json:"status"`
	TimeInForce   string `json:"timeInForce"`
	Type          string `json:"type"`
	Side          string `json:"side"`
	StopPrice     string `json:"stopPrice"`
	IcebergQty    string `json:"icebergQty"`
	Time          uint64 `json:"time"`
}

type Binaorders []Binaorder

type CoinDepositAddress struct {
	Address string `json:"address"`
	Coin    string `json:"coin"`
	Tag     string `json:"tag"`
	URL     string `json:"url"`
}

type Binacancel struct {
	Code              int    `json:"code"`
	Msg               string `json:"msg"`
	Symbol            string `json:"symbol"`
	OrigClientOrderID string `json:"origClientOrderId"`
	OrderID           uint64 `json:"orderId"`
	ClientOrderID     string `json:"clientOrderId"`
}

// Binadeposit ...
type Binadeposits []Binadeposit

type Binadeposit struct {
	InsertTime   uint64 `json:"insertTime"`
	Amount       string `json:"amount"`
	Coin         string `json:"coin"`
	Address      string `json:"address"`
	TxID         string `json:"txId"`
	Status       int    `json:"status"`
	Network      string `json:"network"`
	TransferType int    `json:"transferType"` // 1 for internal, 0 for external
	ConfirmTimes string `json:"confirmTimes"` // exp: "12/12"
}

// Binawithdrawals ...
type Binawithdrawals []Binawithdrawal

// Binawithdrawal object for withdraw from binance
type Binawithdrawal struct {
	ID              string `json:"id"`
	WithdrawOrderID string `json:"withdrawOrderId"`
	Coin            string `json:"coin"`
	Amount          string `json:"amount"`
	Address         string `json:"address"`
	TxID            string `json:"txId"`
	ApplyTime       string `json:"applyTime"`
	Status          int    `json:"status"`
	Network         string `json:"network"`
}

type BinaServerTime struct {
	ServerTime uint64 `json:"serverTime"`
}

// BinanceTradeHistory object for recent trade on binance
type BinanceTradeHistory []struct {
	ID           uint64 `json:"id"`
	Price        string `json:"price"`
	Qty          string `json:"qty"`
	Time         uint64 `json:"time"`
	IsBuyerMaker bool   `json:"isBuyerMaker"`
	IsBestMatch  bool   `json:"isBestMatch"`
}

// BinaAccountTradeHistory object for binance account trade history
type BinaAccountTradeHistory []struct {
	Symbol          string `json:"symbol"`
	ID              uint64 `json:"id"`
	OrderID         uint64 `json:"orderId"`
	Price           string `json:"price"`
	Qty             string `json:"qty"`
	QuoteQty        string `json:"quoteQty"`
	Commission      string `json:"commission"`
	CommissionAsset string `json:"commissionAsset"`
	Time            uint64 `json:"time"`
	IsBuyer         bool   `json:"isBuyer"`
	IsMaker         bool   `json:"isMaker"`
	IsBestMatch     bool   `json:"isBestMatch"`
}

// BinanceAssetDetail ...
type BinanceAssetDetail struct {
	MinWithdrawAmount float64 `json:"minWithdrawAmount"`
	DepositStatus     bool    `json:"depositStatus"`
	WithdrawFee       float64 `json:"withdrawFee"`
	WithdrawStatus    bool    `json:"withdrawStatus"`
	DepositTip        string  `json:"depositTip"` // reason if deposit status is false
}
