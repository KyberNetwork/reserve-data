package storage

import (
	"testing"

	"github.com/KyberNetwork/reserve-data/common"
)

func TestHasPendingDepositRamStorage(t *testing.T) {
	storage := NewRamStorage()
	token := common.NewToken("OMG", "0x1111111111111111111111111111111111111111", 18)
	exchange := common.TestExchange{}
	timepoint := common.GetTimepoint()
	out := storage.HasPendingDeposit(token, exchange)
	if out != false {
		t.Fatalf("Expected ram storage to return true false there is no pending deposit for the same currency and exchange")
	}
	storage.Record(
		"deposit",
		common.NewActivityID(1, "1"),
		string(exchange.ID()),
		map[string]interface{}{
			"exchange":  exchange,
			"token":     token,
			"amount":    "1.0",
			"timepoint": timepoint,
		},
		map[string]interface{}{
			"tx":    "",
			"error": nil,
		},
		"",
		"submitted",
		common.GetTimepoint())
	out = storage.HasPendingDeposit(token, exchange)
	if out != true {
		t.Fatalf("Expected ram storage to return true when there is pending deposit")
	}
}
