package stat

import (
	"fmt"

	"github.com/KyberNetwork/reserve-data/common"
	ethereum "github.com/ethereum/go-ethereum/common"
)

type LogStorageTest struct {
	storage LogStorage
}

func NewLogStorageTest(storage LogStorage) *LogStorageTest {
	return &LogStorageTest{storage}
}

func (self *LogStorageTest) TestCatLog() error {
	var err error
	var catLog = common.SetCatLog{
		Timestamp:       111,
		BlockNumber:     222,
		TransactionHash: ethereum.HexToHash("TESTHASH"),
		Index:           1,
		Address:         ethereum.HexToAddress(testUserAddr),
		Category:        "test",
	}
	err = self.storage.StoreCatLog(catLog)
	if err != nil {
		return err
	}
	catLog = common.SetCatLog{
		Timestamp:       333,
		BlockNumber:     444,
		TransactionHash: ethereum.HexToHash("TESTHASH"),
		Index:           2,
		Address:         ethereum.HexToAddress(testUserAddr),
		Category:        "test",
	}
	err = self.storage.StoreCatLog(catLog)
	if err != nil {
		return err
	}
	result, err := self.storage.GetCatLogs(0, 8640000)
	if err != nil {
		return err
	}
	if len(result) != 2 {
		return fmt.Errorf("GetCatLogs return wrong number of records, expected 2, got %d", len(result))
	}
	record, err := self.storage.GetFirstCatLog()
	if err != nil {
		return err
	}
	if record.BlockNumber != 222 {
		return fmt.Errorf("GetFirstCatLog return wrong record, expect BlockNumber 222, got %d", record.BlockNumber)
	}
	record, err = self.storage.GetLastCatLog()
	if err != nil {
		return err
	}
	if record.BlockNumber != 444 {
		return fmt.Errorf("GetFirstCatLog return wrong record, expect BlockNumber 444, got %d", record.BlockNumber)
	}
	return err
}

func (self *LogStorageTest) TestTradeLog() error {
	var err error
	var tradeLog = common.TradeLog{
		Timestamp:       111,
		BlockNumber:     222,
		TransactionHash: ethereum.HexToHash("TESTHASH"),
		Index:           1,
	}
	err = self.storage.StoreTradeLog(tradeLog, 111)
	if err != nil {
		return err
	}
	tradeLog = common.TradeLog{
		Timestamp:       333,
		BlockNumber:     444,
		TransactionHash: ethereum.HexToHash("TESTHASH"),
		Index:           2,
	}
	err = self.storage.StoreTradeLog(tradeLog, 333)
	if err != nil {
		return err
	}
	result, err := self.storage.GetTradeLogs(0, 8640000)
	if err != nil {
		return err
	}
	if len(result) != 2 {
		return fmt.Errorf("GetCatLogs return wrong number of records, expected 2, got %d", len(result))
	}
	record, err := self.storage.GetFirstTradeLog()
	if err != nil {
		return err
	}
	if record.BlockNumber != 222 {
		return fmt.Errorf("GetFirstCatLog return wrong record, expect BlockNumber 222, got %d", record.BlockNumber)
	}
	record, err = self.storage.GetLastTradeLog()
	if err != nil {
		return err
	}
	if record.BlockNumber != 444 {
		return fmt.Errorf("GetFirstCatLog return wrong record, expect BlockNumber 444, got %d", record.BlockNumber)
	}
	return err
}

func (self *LogStorageTest) TestUtil() error {
	var err error
	err = self.storage.UpdateLogBlock(222, 111)
	if err != nil {
		return err
	}
	err = self.storage.UpdateLogBlock(333, 112)
	if err != nil {
		return err
	}
	lastBlock, err := self.storage.LastBlock()
	if lastBlock != 333 {
		return fmt.Errorf("LastBlock return wrong result, expect 333, got %d", lastBlock)
	}
	maxlogrange := self.storage.MaxRange()
	if maxlogrange <= 0 {
		return fmt.Errorf("Check maxrange return, got unexpected result %d", maxlogrange)
	}
	return err

}
