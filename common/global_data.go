package common

type GoldRate struct {
	Symbol string  `json:"symbol"`
	Price  float64 `json:"price"`
	Time   uint64  `json:"time"`
}

type DGXGoldData struct {
	Valid     bool
	Timestamp uint64
	Status    string     `json:"success"`
	Data      []GoldRate `json:"data"`
	Error     string
}

type OneForgeGoldData struct {
	Value     float64
	Text      string
	Timestamp uint64
	Error     bool
	Message   string
}

type GDAXGoldData struct {
	Valid   bool
	Error   string
	TradeID uint64 `json:"trade_id"`
	Price   string `json:"price"`
	Size    string `json:"size"`
	Bid     string `json:"bid"`
	Ask     string `json:"ask"`
	Volume  string `json:"volume"`
	Time    string `json:"time"`
}

type KrakenGoldData struct {
	Valid           bool
	Error           string        `json:"network_error"`
	ErrorFromKraken []interface{} `json:"error"`
	Result          map[string]struct {
		A []string `json:"a"`
		B []string `json:"b"`
		C []string `json:"c"`
		V []string `json:"v"`
		P []string `json:"p"`
		T []uint64 `json:"t"`
		L []string `json:"l"`
		H []string `json:"h"`
		O string   `json:"o"`
	} `json:"result"`
}

type GeminiGoldData struct {
	Valid  bool
	Error  string
	Bid    string `json:"bid"`
	Ask    string `json:"ask"`
	Volume struct {
		ETH       string `json:"ETH"`
		USD       string `json:"USD"`
		Timestamp uint64 `json:"timestamp"`
	} `json:"volume"`
	Last string `json:"last"`
}

type GoldData struct {
	Timestamp   uint64
	DGX         DGXGoldData
	OneForgeETH OneForgeGoldData
	OneForgeUSD OneForgeGoldData
	GDAX        GDAXGoldData
	Kraken      KrakenGoldData
	Gemini      GeminiGoldData
}
