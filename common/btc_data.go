package common

// BTCData is the data returned by /btc-feed API.
type BTCData struct {
	Timestamp uint64
	Coinbase  CoinbaseData `json:"CoinbaseBTC"`
	Binance   BinanceData  `json:"BinanceBTC"`
}

// CoinBaseData is the response of Coinbase ETH/BTC ticker request.
// Example: https://api.pro.coinbase.com/products/eth-btc/ticker
// Response
// {
//  "trade_id": 7188449,
//  "price": "0.02707000",
//  "size": "3.43340266",
//  "time": "2019-05-13T04:56:14.777Z",
//  "bid": "0.02707",
//  "ask": "0.02708",
//  "volume": "31528.93120442"
//}
type CoinbaseData struct {
	Valid   bool
	Error   string
	TradeID uint64 `json:"trade_id"`
	Price   string `json:"price"`
	Size    string `json:"size"`
	Time    string `json:"time"`
	Bid     string `json:"bid"`
	Ask     string `json:"ask"`
	Volume  string `json:"volume"`
}

type GeminiVolumeETHBTC struct {
	ETH       string `json:"ETH"`
	BTC       string `json:"BTC"`
	Timestamp uint64 `json:"timestamp"`
}

// GeminiETHBTCData is the data returns by Gemini ETH/BTC ticker.
// Example: https://api.gemini.com/v1/pubticker/ethbtc
// Response
// {
//  "bid": "0.02686",
//  "ask": "0.02694",
//  "volume": {
//    "ETH": "5595.131076",
//    "BTC": "149.60020830467",
//    "timestamp": 1557728700000
//  },
//  "last": "0.02694"
// }
type GeminiETHBTCData struct {
	Valid  bool
	Error  string
	Bid    string             `json:"bid"`
	Ask    string             `json:"ask"`
	Volume GeminiVolumeETHBTC `json:"volume"`
	Last   string             `json:"last"`
}

type BitFinexData struct {
	Valid           bool
	Error           string
	Bid             float64 `json:"bid"`
	BidSize         float64 `json:"bid_size"`
	Ask             float64 `json:"ask"`
	AskSize         float64 `json:"ask_size"`
	DailyChange     float64 `json:"daily_change"`
	DailyChangePerc float64 `json:"daily_change_perc"`
	LastPrice       float64 `json:"last_price"`
	Volume          float64 `json:"volume"`
	High            float64 `json:"high"`
	Low             float64 `json:"low"`
}
