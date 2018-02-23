package exchange

import (
	"errors"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"
	"sync"

	"github.com/KyberNetwork/reserve-data/common"
	ethereum "github.com/ethereum/go-ethereum/common"
)

const HUOBI_EPSILON float64 = 0.0000000001 // 10e-10

type Huobi struct {
	interf       HuobiInterface
	pairs        []common.TokenPair
	addresses    *common.ExchangeAddresses
	exchangeInfo *common.ExchangeInfo
	fees         common.ExchangeFees
}

func (self *Huobi) MarshalText() (text []byte, err error) {
	return []byte(self.ID()), nil
}

func (self *Huobi) TokenAddresses() map[string]ethereum.Address {
	return self.addresses.GetData()
}

func (self *Huobi) Address(token common.Token) (ethereum.Address, bool) {
	addr, supported := self.addresses.Get(token.ID)
	return addr, supported
}

func (self *Huobi) UpdateAllDepositAddresses(address string) {
	data := self.addresses.GetData()
	for k, _ := range data {
		self.addresses.Update(k, ethereum.HexToAddress(address))
	}
}

func (self *Huobi) UpdateDepositAddress(token common.Token, address string) {
	liveAddress, _ := self.interf.GetDepositAddress(strings.ToLower(token.ID))
	if liveAddress.Address != "" {
		self.addresses.Update(token.ID, ethereum.HexToAddress(liveAddress.Address))
	} else {
		self.addresses.Update(token.ID, ethereum.HexToAddress(address))
	}
}

func (self *Huobi) UpdatePrecisionLimit(pair common.TokenPair, symbols HuobiExchangeInfo) {
	pairName := strings.ToLower(pair.Base.ID) + strings.ToLower(pair.Quote.ID)
	for _, symbol := range symbols.Data {
		if symbol.Base+symbol.Quote == pairName {
			exchangePrecisionLimit := common.ExchangePrecisionLimit{}
			exchangePrecisionLimit.Precision.Amount = symbol.AmountPrecision
			exchangePrecisionLimit.Precision.Price = symbol.PricePrecision
			self.exchangeInfo.Update(pair.PairID(), exchangePrecisionLimit)
			break
		}
	}
}

func (self *Huobi) UpdatePairsPrecision() {
	exchangeInfo, err := self.interf.GetExchangeInfo()
	if err != nil {
		log.Printf("Get exchange info failed: %s\n", err)
	} else {
		for _, pair := range self.pairs {
			self.UpdatePrecisionLimit(pair, exchangeInfo)
		}
	}
}

func (self *Huobi) GetInfo() (common.ExchangeInfo, error) {
	return *self.exchangeInfo, nil
}

func (self *Huobi) GetExchangeInfo(pair common.TokenPairID) (common.ExchangePrecisionLimit, error) {
	data, err := self.exchangeInfo.Get(pair)
	return data, err
}

func (self *Huobi) GetFee() common.ExchangeFees {
	return self.fees
}

func (self *Huobi) ID() common.ExchangeID {
	return common.ExchangeID("huobi")
}

func (self *Huobi) TokenPairs() []common.TokenPair {
	return self.pairs
}

func (self *Huobi) Name() string {
	return "huobi"
}

func (self *Huobi) QueryOrder(symbol string, id uint64, timepoint uint64) (done float64, remaining float64, finished bool, err error) {
	result, err := self.interf.OrderStatus(symbol, id, timepoint)
	if err != nil {
		return 0, 0, false, err
	} else {
		done, _ := strconv.ParseFloat(result.Data.ExecutedQty, 64)
		total, _ := strconv.ParseFloat(result.Data.OrigQty, 64)
		return done, total - done, total-done < HUOBI_EPSILON, nil
	}
}

func (self *Huobi) Trade(tradeType string, base common.Token, quote common.Token, rate float64, amount float64, timepoint uint64) (id string, done float64, remaining float64, finished bool, err error) {
	result, err := self.interf.Trade(tradeType, base, quote, rate, amount, timepoint)
	symbol := base.ID + quote.ID

	if err != nil {
		return "", 0, 0, false, err
	} else {
		orderID, _ := strconv.ParseUint(result.OrderID, 10, 64)
		done, remaining, finished, err := self.QueryOrder(
			base.ID+quote.ID,
			orderID,
			timepoint+20,
		)
		id := fmt.Sprintf("%s_%s", result.OrderID, symbol)
		return id, done, remaining, finished, err
	}
}

func (self *Huobi) Withdraw(token common.Token, amount *big.Int, address ethereum.Address, timepoint uint64) (string, error) {
	withdrawID, err := self.interf.Withdraw(token, amount, address, timepoint)
	if err != nil {
		return "", err
	}
	// this magical logic base on inspection on huobi website
	result := withdrawID + "01"
	return result, err
}

func (self *Huobi) CancelOrder(id common.ActivityID) error {
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
	if result.Status != "ok" {
		return errors.New("Couldn't cancel order id " + id.EID)
	}
	return nil
}

func (self *Huobi) FetchOnePairData(
	wg *sync.WaitGroup,
	pair common.TokenPair,
	data *sync.Map,
	timepoint uint64) {

	defer wg.Done()
	result := common.ExchangePrice{}

	timestamp := common.Timestamp(fmt.Sprintf("%d", timepoint))
	result.Timestamp = timestamp
	result.Valid = true
	resp_data, err := self.interf.GetDepthOnePair(pair, timepoint)
	returnTime := common.GetTimestamp()
	result.ReturnTime = returnTime
	if err != nil {
		result.Valid = false
		result.Error = err.Error()
	} else {
		if resp_data.Status != "ok" {
			result.Valid = false
		} else {
			for _, buy := range resp_data.Tick.Bids {
				quantity := buy[1]
				rate := buy[0]
				result.Bids = append(
					result.Bids,
					common.PriceEntry{
						quantity,
						rate,
					},
				)
			}
			for _, sell := range resp_data.Tick.Asks {
				quantity := sell[1]
				rate := sell[0]
				result.Asks = append(
					result.Asks,
					common.PriceEntry{
						quantity,
						rate,
					},
				)
			}
		}
	}
	data.Store(pair.PairID(), result)
}

func (self *Huobi) FetchPriceData(timepoint uint64) (map[common.TokenPairID]common.ExchangePrice, error) {
	wait := sync.WaitGroup{}
	data := sync.Map{}
	pairs := self.pairs
	for _, pair := range pairs {
		wait.Add(1)
		go self.FetchOnePairData(&wait, pair, &data, timepoint)
	}
	wait.Wait()
	result := map[common.TokenPairID]common.ExchangePrice{}
	data.Range(func(key, value interface{}) bool {
		result[key.(common.TokenPairID)] = value.(common.ExchangePrice)
		return true
	})
	return result, nil
}

func (self *Huobi) OpenOrdersForOnePair(
	wg *sync.WaitGroup,
	pair common.TokenPair,
	data *sync.Map,
	timepoint uint64) {

	// defer wg.Done()

	// result, err := self.interf.OpenOrdersForOnePair(pair, timepoint)

	//TODO: complete open orders for one pair
}

func (self *Huobi) FetchOrderData(timepoint uint64) (common.OrderEntry, error) {
	result := common.OrderEntry{}
	result.Timestamp = common.Timestamp(fmt.Sprintf("%d", timepoint))
	result.Valid = true
	result.Data = []common.Order{}

	wait := sync.WaitGroup{}
	data := sync.Map{}
	pairs := self.pairs
	for _, pair := range pairs {
		wait.Add(1)
		go self.OpenOrdersForOnePair(&wait, pair, &data, timepoint)
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

func (self *Huobi) FetchEBalanceData(timepoint uint64) (common.EBalanceEntry, error) {
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
		if resp_data.Status != "ok" {
			result.Valid = false
			result.Error = fmt.Sprintf("Cannot fetch ebalance")
		} else {
			balances := resp_data.Data.List
			for _, b := range balances {
				tokenID := strings.ToUpper(b.Currency)
				_, exist := common.SupportedTokens[tokenID]
				if exist {
					balance, _ := strconv.ParseFloat(b.Balance, 64)
					if b.Type == "trade" {
						result.AvailableBalance[tokenID] = balance
					} else {
						result.LockedBalance[tokenID] = balance
					}
					result.DepositBalance[tokenID] = 0
				}
			}
			return result, nil
		}
	}
	return result, nil
}

func (self *Huobi) FetchOnePairTradeHistory(
	wait *sync.WaitGroup,
	data *sync.Map,
	pair common.TokenPair,
	timepoint uint64) {

	defer wait.Done()
	result := []common.TradeHistory{}
	resp, err := self.interf.GetAccountTradeHistory(pair.Base, pair.Quote, timepoint)
	if err != nil {
		log.Printf("Cannot fetch data for pair %s%s: %s", pair.Base.ID, pair.Quote.ID, err.Error())
	}
	pairString := pair.PairID()
	for _, trade := range resp.Data {
		price, _ := strconv.ParseFloat(trade.Price, 64)
		quantity, _ := strconv.ParseFloat(trade.Amount, 64)
		historyType := "sell"
		if trade.Type == "buy-limit" {
			historyType = "buy"
		}
		tradeHistory := common.TradeHistory{
			strconv.FormatUint(trade.ID, 10),
			price,
			quantity,
			historyType,
			trade.Timestamp,
		}
		result = append(result, tradeHistory)
	}
	data.Store(pairString, result)
}

func (self *Huobi) FetchTradeHistory(timepoint uint64) (map[common.TokenPairID][]common.TradeHistory, error) {
	result := map[common.TokenPairID][]common.TradeHistory{}
	data := sync.Map{}
	pairs := self.pairs
	wait := sync.WaitGroup{}
	for _, pair := range pairs {
		wait.Add(1)
		go self.FetchOnePairTradeHistory(&wait, &data, pair, timepoint)
	}
	wait.Wait()
	data.Range(func(key, value interface{}) bool {
		result[key.(common.TokenPairID)] = value.([]common.TradeHistory)
		return true
	})
	return result, nil
}

func (self *Huobi) DepositStatus(id common.ActivityID, timepoint uint64) (string, error) {
	idParts := strings.Split(id.EID, "|")
	txID := idParts[0]
	deposits, err := self.interf.DepositHistory()
	if err != nil && deposits.Status != "ok" {
		return "", err
	}
	for _, deposit := range deposits.Data {
		if deposit.TxHash == txID {
			if deposit.State == "safe" {
				return "done", nil
			}
			return "", nil
		}
	}
	return "", errors.New("Deposit doesn't exist. This shouldn't happen unless tx returned from huobi and activity ID are not consistently designed")
}

func (self *Huobi) WithdrawStatus(id common.ActivityID, timepoint uint64) (string, string, error) {
	withdrawID, _ := strconv.ParseUint(id.EID, 10, 64)
	withdraws, err := self.interf.WithdrawHistory()
	if err != nil {
		return "", "", nil
	}
	log.Printf("Withdrawal id: %s", withdrawID)
	for _, withdraw := range withdraws.Data {
		if withdraw.TxID == withdrawID {
			if withdraw.State == "confirmed" {
				return "done", withdraw.TxHash, nil
			}
			return "", withdraw.TxHash, nil
		}
	}
	return "", "", errors.New("Withdrawal doesn't exist. This shouldn't happen unless tx returned from withdrawal from huobi and activity ID are not consistently designed")
}

func (self *Huobi) OrderStatus(id common.ActivityID, timepoint uint64) (string, error) {
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
	if order.Data.State == "pre-submitted" || order.Data.State == "submitting" || order.Data.State == "submitted" || order.Data.State == "partial-filled" || order.Data.State == "partial-canceled" {
		return "", nil
	} else {
		return "done", nil
	}
}

func NewHuobi(addressConfig map[string]string, feeConfig common.ExchangeFees, interf HuobiInterface) *Huobi {
	pairs, fees := getExchangePairsAndFeesFromConfig(addressConfig, feeConfig, "huobi")
	return &Huobi{
		interf,
		pairs,
		common.NewExchangeAddresses(),
		common.NewExchangeInfo(),
		fees,
	}
}
