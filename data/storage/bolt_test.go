package storage

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/metric"
)

func TestHasPendingDepositBoltStorage(t *testing.T) {
	boltFile := "test_bolt.db"
	os.Remove(boltFile)
	storage, err := NewBoltStorage(boltFile)
	if err != nil {
		t.Fatalf("Couldn't init bolt storage %v", err)
	}
	token := common.NewToken("OMG", "0x1111111111111111111111111111111111111111", 18)
	exchange := common.TestExchange{}
	timepoint := common.GetTimepoint()
	out, err := storage.HasPendingDeposit(token, exchange)
	if err != nil {
		t.Fatalf(err.Error())
	}
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
	out, err = storage.HasPendingDeposit(token, exchange)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if out != true {
		t.Fatalf("Expected ram storage to return true when there is pending deposit")
	}
}

func TestConvertTargetQtyV1toV2(t *testing.T) {
	// Mock v1 data, just data, other fields does not affected
	v1Data := metric.TokenTargetQty{
		Data: "KNC_0.12314_0.427482_0.42348_0.423489|EOS_0.124_0.42342_0.42342_0.4246",
	}
	v2Data := convertTargetQtyV1toV2(v1Data)
	if len(v2Data) != 2 {
		t.Fatalf("Expected convert from v1 to v2 data with length 2, got length %d", len(v2Data))
	}
	for _, data := range v2Data {
		bytes, err := json.Marshal(data)
		if err != nil {
			t.Fatalf("Expected data is a interface of metric TargetQtyStruct")
		}
		v2Struct := metric.TargetQtyStruct{}
		if err := json.Unmarshal(bytes, &v2Struct); err != nil {
			t.Fatalf("Expected data is a inteface of metric")
		}
	}
}
