package stat

import (
	"fmt"
	"strings"

	"github.com/KyberNetwork/reserve-data/common"
)

type RateStorageTest struct {
	storage RateStorage
}

func NewRateStorageTest(storage RateStorage) *RateStorageTest {
	return &RateStorageTest{storage}
}

func (self *RateStorageTest) TestReserveRates() error {
	var err error
	rate := common.ReserveRates{
		Timestamp:     111,
		ReturnTime:    222,
		BlockNumber:   333,
		ToBlockNumber: 444,
	}
	err = self.storage.StoreReserveRates(TESTASSETADDR, rate, 111)
	if err != nil {
		return err
	}
	result, err := self.storage.GetReserveRates(0, 8640000, strings.ToLower(TESTASSETADDR))
	if result == nil || len(result) < 1 {
		return fmt.Errorf("GetReserverRates return empty result (LOWER CASE ADDR)")
	}
	rate = result[0]
	if rate.BlockNumber != 333 {
		return fmt.Errorf("Get ReserverRates return wrong result, expect blockNumber 333, got %d (LOWER CASE ADDR)", rate.BlockNumber)
	}
	result, err = self.storage.GetReserveRates(0, 8640000, strings.ToUpper(TESTASSETADDR))
	if result == nil || len(result) < 1 {
		return fmt.Errorf("GetReserverRates return empty result (UPPER CASE ADDR)")
	}
	rate = result[0]
	if rate.BlockNumber != 333 {
		return fmt.Errorf("Get ReserverRates return wrong result, expect blockNumber 333, got %d (UPPER CASE ADDR)", rate.BlockNumber)
	}
	return err
}
