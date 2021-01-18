package feed

type Feed int

// Feed ..
//go:generate stringer -type=Feed -linecomment
const (
	OneForgeXAUETH Feed = iota + 1 // OneForgeXAUETH
	OneForgeXAUUSD                 // OneForgeXAUUSD
	GDAXETHUSD                     // GDAXETHUSD
	KrakenETHUSD                   // KrakenETHUSD
	GeminiETHUSD                   // GeminiETHUSD

	CoinbaseETHBTC3 // CoinbaseETHBTC3
	BinanceETHBTC3  // BinanceETHBTC3

	CoinbaseETHUSDDAI5000 // CoinbaseETHUSDDAI5000
	CurveDAIUSDC10000     // CurveDAIUSDC10000
	BinanceETHUSDC10000   // BinanceETHUSDC10000
)

var (
	dummyStruct = struct{}{}
	// usdFeeds list of supported usd feeds
	usdFeeds = map[string]struct{}{
		CoinbaseETHUSDDAI5000.String(): dummyStruct,
		CurveDAIUSDC10000.String():     dummyStruct,
		BinanceETHUSDC10000.String():   dummyStruct,
	}

	// btcFeeds list of supported btc feeds
	btcFeeds = map[string]struct{}{
		BinanceETHBTC3.String():  dummyStruct,
		CoinbaseETHBTC3.String(): dummyStruct,
	}

	goldFeeds = map[string]struct{}{
		OneForgeXAUETH.String(): dummyStruct,
		OneForgeXAUUSD.String(): dummyStruct,
		GDAXETHUSD.String():     dummyStruct,
		KrakenETHUSD.String():   dummyStruct,
		GeminiETHUSD.String():   dummyStruct,
	}
)

// SupportedFeed ...
type SupportedFeed struct {
	Gold map[string]struct{}
	BTC  map[string]struct{}
	USD  map[string]struct{}
}

// AllFeeds returns all configured feed sources.
func AllFeeds() SupportedFeed {
	return SupportedFeed{
		Gold: goldFeeds,
		BTC:  btcFeeds,
		USD:  usdFeeds,
	}
}
