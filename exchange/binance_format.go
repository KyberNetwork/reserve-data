package exchange

import (
	"encoding/json"
	"fmt"
)

// Binaprice binance order book item
type Binaprice struct {
	Quantity string
	Rate     string
}

// UnmarshalJSON custom unmarshal reponse from binance api to Binaprice
func (bp *Binaprice) UnmarshalJSON(text []byte) error {
	temp := []interface{}{}
	if err := json.Unmarshal(text, &temp); err != nil {
		return err
	}
	qty, ok := temp[1].(string)
	if !ok {
		return fmt.Errorf("unmarshal err: interface %v can't be converted to string", temp[1])
	}
	bp.Quantity = qty
	rate, ok := temp[0].(string)
	if !ok {
		return fmt.Errorf("unmarshal err: interface %v can't be converted to string", temp[0])
	}
	bp.Rate = rate
	return nil
}

// Binaresp is response for binance orderbook
type Binaresp struct {
	LastUpdatedID int64       `json:"lastUpdateId"`
	Code          int         `json:"code"`
	Msg           string      `json:"msg"`
	Bids          []Binaprice `json:"bids"`
	Asks          []Binaprice `json:"asks"`
}

// Binainfo is user account info including balance
type Binainfo struct {
	Code             int    `json:"code"`
	Msg              string `json:"msg"`
	MakerCommission  int64  `json:"makerCommission"`
	TakerCommission  int64  `json:"takerCommission"`
	BuyerCommission  int64  `json:"buyerCommission"`
	SellerCommission int64  `json:"sellerCommission"`
	CanTrade         bool   `json:"canTrade"`
	CanWithdraw      bool   `json:"canWithdraw"`
	CanDeposit       bool   `json:"canDeposit"`
	Balances         []struct {
		Asset  string `json:"asset"`
		Free   string `json:"free"`
		Locked string `json:"locked"`
	} `json:"balances"`
}

// FilterLimit limit info from binance
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

// BinanceSymbol is precision from binance
type BinanceSymbol struct {
	Symbol             string        `json:"symbol"`
	BaseAssetPrecision int           `json:"baseAssetPrecision"`
	QuotePrecision     int           `json:"quotePrecision"`
	Filters            []FilterLimit `json:"filters"`
}

// BinanceExchangeInfo list of token info
type BinanceExchangeInfo struct {
	Symbols []BinanceSymbol
}

// Binatrade request
type Binatrade struct {
	Symbol        string `json:"symbol"`
	OrderID       uint64 `json:"orderId"`
	ClientOrderID string `json:"clientOrderId"`
	TransactTime  uint64 `json:"transactTime"`
}

// Binawithdraw response for withdrawal
type Binawithdraw struct {
	Success bool   `json:"success"`
	Msg     string `json:"msg"`
	ID      string `json:"id"`
}

// Binaorder - binance order status
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

// Binaorders list of order on binance
type Binaorders []Binaorder

// Binadepositaddress - deposit address from binance
type Binadepositaddress struct {
	Success    bool   `json:"success"`
	Msg        string `json:"msg"`
	Address    string `json:"address"`
	AddressTag string `json:"addressTag"`
	Asset      string `json:"asset"`
}

// Binacancel cancel order from binance
type Binacancel struct {
	Code              int    `json:"code"`
	Msg               string `json:"msg"`
	Symbol            string `json:"symbol"`
	OrigClientOrderID string `json:"origClientOrderId"`
	OrderID           uint64 `json:"orderId"`
	ClientOrderID     string `json:"clientOrderId"`
}

// Binadeposits list of deposit info
type Binadeposits struct {
	Success  bool          `json:"success"`
	Msg      string        `json:"msg"`
	Deposits []Binadeposit `json:"depositList"`
}

// Binadeposit deposit record
type Binadeposit struct {
	InsertTime uint64  `json:"insertTime"`
	Amount     float64 `json:"amount"`
	Asset      string  `json:"asset"`
	Address    string  `json:"address"`
	TxID       string  `json:"txId"`
	Status     int     `json:"status"`
}

// Binawithdrawals list of withdrawals
type Binawithdrawals struct {
	Success     bool             `json:"success"`
	Msg         string           `json:"msg"`
	Withdrawals []Binawithdrawal `json:"withdrawList"`
}

// Binawithdrawal withdrawal record from binance
type Binawithdrawal struct {
	ID        string  `json:"id"`
	Amount    float64 `json:"amount"`
	Address   string  `json:"address"`
	Asset     string  `json:"asset"`
	TxID      string  `json:"txId"`
	ApplyTime uint64  `json:"applyTime"`
	Status    int     `json:"status"`
}

// BinaServerTime time from binance server
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
