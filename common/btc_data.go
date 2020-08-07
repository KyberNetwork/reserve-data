package common

// BTCData is the data returned by /btc-feed API.
type BTCData struct {
	Timestamp       uint64
	CoinbaseETHBTC3 FeedProviderResponse `json:"CoinbaseETHBTC3"`
	BinanceETHBTC3  FeedProviderResponse `json:"BinanceETHBTC3"`
}
