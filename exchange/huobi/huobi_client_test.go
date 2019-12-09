package huobi

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDepositAddress(t *testing.T) {
	t.Skip()     // skip as external test
	key := ""    // enter only once for test
	secret := "" // enter only once for test
	signer, err := NewSigner(key, secret)
	assert.NoError(t, err)
	interf := NewRealInterface("https://api.huobi.pro")
	ep := NewHuobiClient(*signer, interf)

	listTokens := []string{"LBA", "ZIL", "DTA", "EKO", "POLY", "CVC", "BIX"}
	for _, token := range listTokens {
		depositAddress, err := ep.GetDepositAddress(token)
		assert.NoError(t, err)

		log.Printf("deposit address %s: %+v", token, depositAddress.Data)
	}
}
