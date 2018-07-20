package stat

import (
	"fmt"

	ethereum "github.com/ethereum/go-ethereum/common"

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
	testAsset := ethereum.HexToAddress(testAssetAddr)
	err = self.storage.StoreReserveRates(testAsset, rate, 111)
	if err != nil {
		return err
	}
	result, err := self.storage.GetReserveRates(0, 8640000, testAsset)
	if result == nil || len(result) < 1 {
		return fmt.Errorf("GetReserverRates return empty result ")
	}
	rate = result[0]
	if rate.BlockNumber != 333 {
		return fmt.Errorf("Get ReserverRates return wrong result, expect blockNumber 333, got %d ", rate.BlockNumber)
	}
	return err
}
