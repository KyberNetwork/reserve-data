package exchange

import (
	"encoding/json"
)

type Orderbook interface {
	GetBids() [][]string
	GetAsks() [][]string

type Binaprice struct {
	Quantity string
	Rate     string
}

func (self *Binaprice) UnmarshalJSON(text []byte) error {
	temp := []interface{}{}
	err := json.Unmarshal(text, &temp)
	self.Quantity = temp[1].(string)
	self.Rate = temp[0].(string)
	return err
}

type Binaresp struct {
	LastUpdatedId int64       `json:"lastUpdateId"`
	Code          int         `json:"code"`
	Msg           string      `json:"msg"`
	Bids          []Binaprice `json:"bids"`
	Asks          []Binaprice `json:"asks"`
}

type Binasocketresp struct {
	EventType     string     `json:"e"`
	EventTime     int64      `json:"E"`
	Symbol        string     `json:"s"`
	LastUpdatedID int64      `json:"u"`
	Bids          [][]string `json:"b"`
	Asks          [][]string `json:"a"`
}

func (self Binaresp) GetBids() [][]string {
	return self.Bids
}

func (self Binaresp) GetAsks() [][]string {
	return self.Asks
}

func (self Binasocketresp) GetBids() [][]string {
	return self.Bids
}

func (self Binasocketresp) GetAsks() [][]string {
	return self.Asks
}

type Binasocketuser struct {
	EventType        string `json:"e"`
	EventTime        int64  `json:"E"`
	MakerCommission  int64  `json:"m"`
	TakerCommission  int64  `json:"t"`
	BuyerCommission  int64  `json:"b"`
	SellerCommission int64  `json:"s"`
	CanTrade         bool   `json:"T"`
	CanWithdraw      bool   `json:"W"`
	CanDeposit       bool   `json:"D"`
	Balances         []struct {
		Asset  string `json:"a"`
		Free   string `json:"f"`
		Locked string `json:"l"`
	}
}

type Binasocketaggtrade struct {
	EventType        string `json:"e"`
	EventTime        int64  `json:"E"`
	Symbol           string `json:"s"`
	AggregatedID     string `json:"a"`
	Price            string `json:"p"`
	Quantity         string `json:"q"`
	FirstBreakdownID int64  `json:"f"`
	LastBreakdownID  int64  `json:"l"`
	TradeTime        int64  `json:"T"`
	Maker            bool   `json:"m"`
}

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
		Locked string `json:"locked`
	} `json:"balances`
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
	Symbol             string        `json:"symbol"`
	BaseAssetPrecision int           `json:"baseAssetPrecision"`
	QuotePrecision     int           `json:"quotePrecision"`
	Filters            []FilterLimit `json:"filters"`
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
	OrderId       uint64 `json:"orderId"`
	ClientOrderId string `json:"clientOrderId"`
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

type Binasocketorder struct {
	Symbol        string `json:"s"`
	OrderID       uint64 `json:"i"`
	ClientOrderID string `json:"c"`
	Price         string `json:"p"`
	OrigQty       string `json:"q"`
	Status        string `json:"X"`
	TimeInForce   string `json:"f"`
	Type          string `json:"o"`
	Side          string `json:"S"`
	Time          uint64 `json:"T"`
}

type Binaorders []Binaorder

type Binadepositaddress struct {
	Success    bool   `json:"success"`
	Msg        string `json:"msg"`
	Address    string `json:"address"`
	AddressTag string `json:"addressTag"`
	Asset      string `json:"asset"`
}

type Binacancel struct {
	Code              int    `json:"code"`
	Msg               string `json:"msg"`
	Symbol            string `json:"symbol"`
	OrigClientOrderId string `json:"origClientOrderId"`
	OrderId           uint64 `json:"orderId"`
	ClientOrderId     string `json:"clientOrderId"`
}

// {
// 	"depositList": [
// 		{
// 			"insertTime": 1508198532000,
// 			"amount": 0.04670582,
// 			"asset": "ETH",
// 			"address": "0x6915f16f8791d0a1cc2bf47c13a6b2a92000504b",
// 			"txId": "0xdf33b22bdb2b28b1f75ccd201a4a4m6e7g83jy5fc5d5a9d1340961598cfcb0a1",
// 			"status": 1
// 		},
// 		{
// 			"insertTime": 1508298532000,
// 			"amount": 1000,
// 			"asset": "XMR",
// 			"address": "463tWEBn5XZJSxLU34r6g7h8jtxuNcDbjLSjkn3XAXHCbLrTTErJrBWYgHJQyrCwkNgYvyV3z8zctJLPCZy24jvb3NiTcTJ",
// 			"addressTag": "342341222",
// 			"txId": "b3c6219639c8ae3f9cf010cdc24fw7f7yt8j1e063f9b4bd1a05cb44c4b6e2509",
// 			"status": 1
// 		}
// 	],
// 	"success": true
// }
type Binadeposits struct {
	Success  bool          `json:"success"`
	Msg      string        `json:"msg"`
	Deposits []Binadeposit `json:"depositList"`
}

type Binadeposit struct {
	InsertTime uint64  `json:"insertTime"`
	Amount     float64 `json:"amount"`
	Asset      string  `json:"asset"`
	Address    string  `json:"address"`
	TxID       string  `json:"txId"`
	Status     int     `json:"status"`
}

// {
// 	"withdrawList": [
// 		{
// 			"id":"7213fea8e94b4a5593d507237e5a555b"
// 			"amount": 1,
// 			"address": "0x6915f16f8791d0a1cc2bf47c13a6b2a92000504b",
// 			"asset": "ETH",
// 			"txId": "0xdf33b22bdb2b28b1f75ccd201a4a4m6e7g83jy5fc5d5a9d1340961598cfcb0a1",
// 			"applyTime": 1508198532000
// 			"status": 4
// 		},
// 		{
// 			"id":"7213fea8e94b4a5534ggsd237e5a555b"
// 			"amount": 1000,
// 			"address": "463tWEBn5XZJSxLU34r6g7h8jtxuNcDbjLSjkn3XAXHCbLrTTErJrBWYgHJQyrCwkNgYvyV3z8zctJLPCZy24jvb3NiTcTJ",
// 			"addressTag": "342341222",
// 			"txId": "b3c6219639c8ae3f9cf010cdc24fw7f7yt8j1e063f9b4bd1a05cb44c4b6e2509",
// 			"asset": "XMR",
// 			"applyTime": 1508198532000,
// 			"status": 4
// 		}
// 	],
// 	"success": true
// }
type Binawithdrawals struct {
	Success     bool             `json:"success"`
	Msg         string           `json:"msg"`
	Withdrawals []Binawithdrawal `json:"withdrawList"`
}

type Binawithdrawal struct {
	ID        string  `json:"id"`
	Amount    float64 `json:"amount"`
	Address   string  `json:"address"`
	Asset     string  `json:"asset"`
	TxID      string  `json:"txId"`
	ApplyTime uint64  `json:"applyTime"`
	Status    int     `json:"status"`
}

type BinaServerTime struct {
	ServerTime uint64 `json:"serverTime"`
}

type BinanceTradeHistory []struct {
	ID          uint64 `json:"id"`
	Price       string `json:"price"`
	Qty         string `json:"qty"`
	Time        uint64 `json:"time"`
	IsBuyer     bool   `json:"isBuyer"`
	IsMaker     bool   `json:"isMaker"`
	IsBestMatch bool   `json:"isBestMatch"`
}

type BinaAccountTradeHistory []struct {
	ID          uint64 `json:"id"`
	OrderID     uint64 `json:"orderId"`
	Price       string `json:"price"`
	Qty         string `json:"qty"`
	Time        uint64 `json:"time"`
	IsBuyer     bool   `json:"isBuyer"`
	IsMaker     bool   `json:"isMaker"`
	isBestMatch bool   `json:"isBestMatch"`
}
