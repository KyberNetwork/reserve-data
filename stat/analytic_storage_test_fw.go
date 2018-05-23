package stat

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type AnalyticStorageTest struct {
	storage AnalyticStorage
}

func NewAnalyticStorageTest(storage AnalyticStorage) *AnalyticStorageTest {
	return &AnalyticStorageTest{storage}
}
func (self *AnalyticStorageTest) TestPriceAnalyticData() error {
	var err error
	//test UpdatePriceAnalytic
	testDict := map[string]interface{}{
		"token":             "OMG",
		"ask_price":         "xxxx",
		"bid_price":         "xxx",
		"mid_afp_price":     0.6555,
		"mid afp_old_price": 0.34555,
		"min_spread":        0.233,
		"trigger_update":    true,
	}
	data, _ := json.Marshal(testDict)
	err = self.storage.UpdatePriceAnalyticData(5678, data)
	if err != nil {
		return err
	}

	//test GetPriceAnalytic
	resp, err := self.storage.GetPriceAnalyticData(0, 86400000)
	if (resp == nil) || (len(resp) < 1) {
		return fmt.Errorf("GetPriceAnalyticData returns empty")
	}
	if resp[0].Timestamp != 5678 {
		return fmt.Errorf("Wrong timestamp return, expect 5678, got %d", resp[0].Timestamp)
	}
	afp, ok := resp[0].Data["mid_afp_price"]
	if !ok {
		return fmt.Errorf("corrupted result returned")
	}
	afpFloat, ok := afp.(float64)
	if !ok {
		return fmt.Errorf("result returns wrong type")
	}
	log.Printf("afp is %v", afpFloat)
	if afpFloat != 0.6555 {
		return fmt.Errorf("Expect mid afp price to be 0.6555, got %v", afpFloat)
	}

	fileName := "testFile"
	defer os.Remove(fileName)
	//test ExportExpiredPriceAnalyticData
	nRecord, err := self.storage.ExportExpiredPriceAnalyticData(31*86400000+1, fileName)
	if err != nil {
		return err
	}
	if nRecord != 1 {
		return fmt.Errorf("Expect pruned 1 record, got %d", nRecord)
	}
	//test PruneExpiredPriceAnalyticData
	nRecord, err = self.storage.PruneExpiredPriceAnalyticData(31*86400000 + 1)
	if err != nil {
		return err
	}
	if nRecord != 1 {
		return fmt.Errorf("Expect pruned 1 record, got %d", nRecord)
	}
	return err
}
