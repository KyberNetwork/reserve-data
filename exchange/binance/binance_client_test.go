package binance

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	binanceKey    = "" // insert once for test
	binanceSecret = "" // insert once for test
)

func TestGetDepositAddress(t *testing.T) {
	t.Skip() // skip as external test

	signer, err := NewSigner(binanceKey, binanceSecret)
	assert.NoError(t, err)
	interf := NewEndpoints("https://api.binance.com")
	binanceEndpoint := NewBinanceClient(*signer, interf)

	address, err := binanceEndpoint.GetDepositAddress("CHAT")
	assert.NoError(t, err)

	log.Printf("deposit address: %v", address)
}
