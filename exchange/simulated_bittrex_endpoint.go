package exchange

import (
	"os"
)

type SimulatedBittrexEndpoint struct {
	*RealBittrexEndpoint
}

func NewSimulatedBittrexEndpoint() *SimulatedBittrexEndpoint {
	realone := NewRealBittrexEndpoint()
	fakeMarketEndpoint := "http://127.0.0.1:8000/market"
	fakeAccountEndpoint := "http://127.0.0.1:8000/account"
	if len(os.Args) > 2 {
		fakeMarketEndpoint = os.Args[1]
		fakeAccountEndpoint = os.Args[2]
	}
	realone.MarketEndpoint = fakeMarketEndpoint
	realone.AccountEndpoint = fakeAccountEndpoint
	return &SimulatedBittrexEndpoint{realone}
}
