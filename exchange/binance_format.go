package exchange

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
	OrigClientOrderID string `json:"origClientOrderId"`
	OrderID           uint64 `json:"orderId"`
	ClientOrderID     string `json:"clientOrderId"`
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

// Binawithdrawal object for withdraw from binance
type Binawithdrawal struct {
	ID        string  `json:"id"`
	Amount    float64 `json:"amount"`
	Address   string  `json:"address"`
	Asset     string  `json:"asset"`
	TxID      string  `json:"txId"`
	ApplyTime uint64  `json:"applyTime"`
	Fee       float64 `json:"transactionFee"`
	Status    int     `json:"status"`
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
