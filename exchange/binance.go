package exchange

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"math/big"

	"github.com/KyberNetwork/reserve-data/common"
	ethereum "github.com/ethereum/go-ethereum/common"
)

type Binance struct {
	interf      BinanceInterface
	pairs       []common.TokenPair
	databusType string
	addresses   map[string]ethereum.Address
}

func (self *Binance) MarshalText() (text []byte, err error) {
	return []byte(self.ID()), nil
}

func (self *Binance) Address(token common.Token) (ethereum.Address, bool) {
	addr, supported := self.addresses[token.ID]
	return addr, supported
}

func (self *Binance) UpdateAllDepositAddresses(address string) {
	for k, _ := range self.addresses {
		self.addresses[k] = ethereum.HexToAddress(address)
	}
}

func (self *Binance) UpdateDepositAddress(token common.Token, address string) {
	self.addresses[token.ID] = ethereum.HexToAddress(address)
}

func (self *Binance) ID() common.ExchangeID {
	return common.ExchangeID("binance")
}

func (self *Binance) UpdateFetcherDatabusType(databusType string) {
	self.databusType = databusType
}

func (self *Binance) TokenPairs() []common.TokenPair {
	return self.pairs
}

func (self *Binance) Name() string {
	return "binance"
}

func (self *Binance) Trade(tradeType string, base common.Token, quote common.Token, rate float64, amount float64, timepoint uint64) (id string, done float64, remaining float64, finished bool, err error) {
	return self.interf.Trade(tradeType, base, quote, rate, amount, timepoint)
}

func (self *Binance) Withdraw(token common.Token, amount *big.Int, address ethereum.Address, timepoint uint64) (string, error) {
	tx, err := self.interf.Withdraw(token, amount, address, timepoint)
	return tx, err
}

func (self *Binance) CancelOrder(id common.ActivityID) error {
	idParts := strings.Split(id.EID, "_")
	idNo, err := strconv.ParseUint(idParts[0], 10, 64)
	if err != nil {
		return err
	}
	symbol := idParts[1]
	result, err := self.interf.CancelOrder(symbol, idNo)
	if err != nil {
		return err
	}
	if result.Code != 0 {
		return errors.New("Couldn't cancel order id " + id.EID + " err: " + result.Msg)
	}
	return nil
}

func (self Binance) FetchPriceData(timepoint uint64) (map[common.TokenPairID]common.ExchangePrice, error) {
	wait := sync.WaitGroup{}
	data := sync.Map{}
	pairs := self.pairs
	dataChannel := make(chan Orderbook)
	for _, pair := range pairs {
		wait.Add(1)
		if self.databusType == "socket" {
			go self.interf.SocketFetchOnePairData(&wait, pair, &data, dataChannel)
		} else {
			go self.interf.FetchOnePairData(&wait, pair, &data, timepoint)
		}
	}
	wait.Wait()
	result := map[common.TokenPairID]common.ExchangePrice{}
	data.Range(func(key, value interface{}) bool {
		result[key.(common.TokenPairID)] = value.(common.ExchangePrice)
		return true
	})
	return result, nil
}

func (self *Binance) FetchOrderData(timepoint uint64) (common.OrderEntry, error) {
	result := common.OrderEntry{}
	result.Timestamp = common.Timestamp(fmt.Sprintf("%d", timepoint))
	result.Valid = true
	result.Data = []common.Order{}

	wait := sync.WaitGroup{}
	data := sync.Map{}
	pairs := self.pairs
	for _, pair := range pairs {
		wait.Add(1)
		go self.interf.OpenOrdersForOnePair(&wait, pair, &data, timepoint)
	}
	wait.Wait()

	result.ReturnTime = common.GetTimestamp()

	data.Range(func(key, value interface{}) bool {
		orders := value.([]common.Order)
		result.Data = append(result.Data, orders...)
		return true
	})
	return result, nil
}

func (self *Binance) FetchEBalanceData(timepoint uint64) (common.EBalanceEntry, error) {
	result := common.EBalanceEntry{}
	result.Timestamp = common.Timestamp(fmt.Sprintf("%d", timepoint))
	result.Valid = true
	resp_data, err := self.interf.GetInfo(timepoint)
	result.ReturnTime = common.GetTimestamp()
	if err != nil {
		result.Valid = false
		result.Error = err.Error()
	} else {
		result.AvailableBalance = map[string]float64{}
		result.LockedBalance = map[string]float64{}
		result.DepositBalance = map[string]float64{}
		if resp_data.Code != 0 {
			result.Valid = false
			result.Error = fmt.Sprintf("Code: %s, Msg: %s", resp_data.Code, resp_data.Msg)
		} else {
			for _, b := range resp_data.Balances {
				tokenID := b.Asset
				_, exist := common.SupportedTokens[tokenID]
				if exist {
					avai, _ := strconv.ParseFloat(b.Free, 64)
					locked, _ := strconv.ParseFloat(b.Locked, 64)
					result.AvailableBalance[tokenID] = avai
					result.LockedBalance[tokenID] = locked
					result.DepositBalance[tokenID] = 0
				}
			}
		}
	}
	return result, nil
}

func (self *Binance) DepositStatus(id common.ActivityID, timepoint uint64) (string, error) {
	idParts := strings.Split(id.EID, "|")
	if len(idParts) != 2 {
		// here, the exchange id part in id is malformed
		// 1. because analytic didn't pass original ID
		// 2. id is not constructed correctly in a form of uuid + "|" + token
		return "", errors.New("Invalid deposit id")
	}
	txID := idParts[0]
	startTime := timepoint - 86400000
	endTime := timepoint
	deposits, err := self.interf.DepositHistory(startTime, endTime)
	if err != nil || !deposits.Success {
		return "", err
	} else {
		for _, deposit := range deposits.Deposits {
			if deposit.TxID == txID {
				if deposit.Status == 1 {
					return "done", nil
				} else {
					return "", nil
				}
			}
		}
		return "", errors.New("Deposit doesn't exist. This shouldn't happen unless tx returned from binance and activity ID are not consistently designed")
	}
}

func (self *Binance) WithdrawStatus(id common.ActivityID, timepoint uint64) (string, error) {
	withdrawID := id.EID
	startTime := timepoint - 86400000
	endTime := timepoint
	withdraws, err := self.interf.WithdrawHistory(startTime, endTime)
	if err != nil || !withdraws.Success {
		return "", err
	} else {
		for _, withdraw := range withdraws.Withdrawals {
			if withdraw.ID == withdrawID {
				if withdraw.Status == 3 || withdraw.Status == 5 || withdraw.Status == 6 {
					return "done", nil
				} else {
					return "", nil
				}
			}
		}
		return "", errors.New("Withdrawal doesn't exist. This shouldn't happen unless tx returned from withdrawal from binance and activity ID are not consistently designed")
	}
}

func (self *Binance) OrderStatus(id common.ActivityID, timepoint uint64) (string, error) {
	tradeID := id.EID
	parts := strings.Split(tradeID, "_")
	orderID, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		// if this crashes, it means core put malformed activity ID
		panic(err)
	}
	symbol := parts[1]
	order, err := self.interf.OrderStatus(symbol, orderID, timepoint)
	if err != nil {
		return "", err
	}
	if order.Status == "NEW" || order.Status == "PARTIALLY_FILLED" || order.Status == "PENDING_CANCEL" {
		return "", nil
	} else {
		return "done", nil
	}
}

func NewBinance(interf BinanceInterface) *Binance {
	return &Binance{
		interf,
		[]common.TokenPair{
			common.MustCreateTokenPair("FUN", "ETH"),
			common.MustCreateTokenPair("MCO", "ETH"),
			common.MustCreateTokenPair("OMG", "ETH"),
			common.MustCreateTokenPair("EOS", "ETH"),
			common.MustCreateTokenPair("KNC", "ETH"),
			common.MustCreateTokenPair("LINK", "ETH"),
		},
		"http",
		map[string]ethereum.Address{},
	}
}
