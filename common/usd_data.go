package common

// USDData ...
type USDData struct {
	Timestamp             uint64
	CoinbaseETHUSD10000   FeedProviderResponse `json:"CoinbaseETHUSD10000"`
	GeminiETHUSD10000     FeedProviderResponse `json:"GeminiETHUSD10000"`
	CoinbaseETHUSDC10000  FeedProviderResponse `json:"CoinbaseETHUSDC10000"`
	BinanceETHUSDC10000   FeedProviderResponse `json:"BinanceETHUSDC10000"`
	CoinbaseETHUSDDAI5000 FeedProviderResponse `json:"CoinbaseETHUSDDAI5000"`
	BitfinexETHUSDT10000  FeedProviderResponse `json:"BitfinexETHUSDT10000"`
	BinanceETHUSDT10000   FeedProviderResponse `json:"BinanceETHUSDT10000"`
	BinanceETHPAX5000     FeedProviderResponse `json:"BinanceETHPAX5000"`
	BinanceETHBUSD10000   FeedProviderResponse `json:"BinanceETHBUSD10000"`
}
