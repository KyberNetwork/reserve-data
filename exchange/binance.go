package exchange

import (
	"errors"
	"fmt"
	"log"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"math/big"

	"github.com/KyberNetwork/reserve-data/common"
	ethereum "github.com/ethereum/go-ethereum/common"
)

const BINANCE_EPSILON float64 = 0.0000001 // 10e-7

type Binance struct {
	interf       BinanceInterface
	pairs        []common.TokenPair
	addresses    *common.ExchangeAddresses
	exchangeInfo *common.ExchangeInfo
	fees         common.ExchangeFees
}

func (self *Binance) TokenAddresses() map[string]ethereum.Address {
	return self.addresses.GetData()
}

func (self *Binance) MarshalText() (text []byte, err error) {
	return []byte(self.ID()), nil
}

func (self *Binance) Address(token common.Token) (ethereum.Address, bool) {
	addr, supported := self.addresses.Get(token.ID)
	return addr, supported
}

func (self *Binance) UpdateAllDepositAddresses(address string) {
	data := self.addresses.GetData()
	for k, _ := range data {
		self.addresses.Update(k, ethereum.HexToAddress(address))
	}
}

func (self *Binance) UpdateDepositAddress(token common.Token, address string) {
	liveAddress, _ := self.interf.GetDepositAddress(strings.ToLower(token.ID))
	if liveAddress.Address != "" {
		self.addresses.Update(token.ID, ethereum.HexToAddress(liveAddress.Address))
	} else {
		self.addresses.Update(token.ID, ethereum.HexToAddress(address))
	}
}

func (self *Binance) precisionFromStepSize(stepSize string) int {
	re := regexp.MustCompile("0*$")
	parts := strings.Split(re.ReplaceAllString(stepSize, ""), ".")
	if len(parts) > 1 {
		return len(parts[1])
	}
	return 0
}

func (self *Binance) UpdatePrecisionLimit(pair common.TokenPair, symbols []BinanceSymbol) {
	pairName := strings.ToUpper(pair.Base.ID) + strings.ToUpper(pair.Quote.ID)
	for _, symbol := range symbols {
		if symbol.Symbol == strings.ToUpper(pairName) {
			//update precision
			exchangePrecisionLimit := common.ExchangePrecisionLimit{}
			exchangePrecisionLimit.Precision.Amount = symbol.BaseAssetPrecision
			exchangePrecisionLimit.Precision.Price = symbol.QuotePrecision
			// update limit
			for _, filter := range symbol.Filters {
				if filter.FilterType == "LOT_SIZE" {
					// update amount min
					minQuantity, _ := strconv.ParseFloat(filter.MinQuantity, 64)
					exchangePrecisionLimit.AmountLimit.Min = minQuantity
					// update amount max
					maxQuantity, _ := strconv.ParseFloat(filter.MaxQuantity, 64)
					exchangePrecisionLimit.AmountLimit.Max = maxQuantity
					exchangePrecisionLimit.Precision.Amount = self.precisionFromStepSize(filter.StepSize)
				}

				if filter.FilterType == "PRICE_FILTER" {
					// update price min
					minPrice, _ := strconv.ParseFloat(filter.MinPrice, 64)
					exchangePrecisionLimit.PriceLimit.Min = minPrice
					// update price max
					maxPrice, _ := strconv.ParseFloat(filter.MaxPrice, 64)
					exchangePrecisionLimit.PriceLimit.Max = maxPrice
					exchangePrecisionLimit.Precision.Price = self.precisionFromStepSize(filter.TickSize)
				}

				if filter.FilterType == "MIN_NOTIONAL" {
					minNotional, _ := strconv.ParseFloat(filter.MinNotional, 64)
					exchangePrecisionLimit.MinNotional = minNotional
				}
			}
			self.exchangeInfo.Update(pair.PairID(), exchangePrecisionLimit)
			break
		}
	}
}

func (self *Binance) UpdatePairsPrecision() {
	exchangeInfo, err := self.interf.GetExchangeInfo()
	if err != nil {
		log.Printf("Get exchange info failed: %s\n", err)
	} else {
		symbols := exchangeInfo.Symbols
		for _, pair := range self.pairs {
			self.UpdatePrecisionLimit(pair, symbols)
		}
	}
}

func (self *Binance) GetInfo() (common.ExchangeInfo, error) {
	return *self.exchangeInfo, nil
}

func (self *Binance) GetExchangeInfo(pair common.TokenPairID) (common.ExchangePrecisionLimit, error) {
	data, err := self.exchangeInfo.Get(pair)
	return data, err
}

func (self *Binance) GetFee() common.ExchangeFees {
	return self.fees
}

func (self *Binance) ID() common.ExchangeID {
	return common.ExchangeID("binance")
}

func (self *Binance) DatabusType() string {
	return self.databusType
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

func (self *Binance) QueryOrder(symbol string, id uint64, timepoint uint64) (done float64, remaining float64, finished bool, err error) {
	result, err := self.interf.OrderStatus(symbol, id, timepoint)
	if err != nil {
		return 0, 0, false, err
	} else {
		done, _ := strconv.ParseFloat(result.ExecutedQty, 64)
		total, _ := strconv.ParseFloat(result.OrigQty, 64)
		return done, total - done, total-done < BINANCE_EPSILON, nil
	}
}

func (self *Binance) Trade(tradeType string, base common.Token, quote common.Token, rate float64, amount float64, timepoint uint64) (id string, done float64, remaining float64, finished bool, err error) {
	result, err := self.interf.Trade(tradeType, base, quote, rate, amount, timepoint)
	symbol := base.ID + quote.ID

	if err != nil {
		return "", 0, 0, false, err
	} else {
		done, remaining, finished, err := self.QueryOrder(
			base.ID+quote.ID,
			result.OrderID,
			timepoint+20,
		)
		id := fmt.Sprintf("%s_%s", strconv.FormatUint(result.OrderID, 10), symbol)
		return id, done, remaining, finished, err
	}
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
	_, err = self.interf.CancelOrder(symbol, idNo)
	if err != nil {
		return err
	}
	return nil
}

func (self *Binance) FetchOnePairData(
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
		if resp_data.Code != 0 || resp_data.Msg != "" {
			result.Valid = false
			result.Error = fmt.Sprintf("Code: %d, Msg: %s", resp_data.Code, resp_data.Msg)
		} else {
			for _, buy := range resp_data.Bids {
				quantity, _ := strconv.ParseFloat(buy.Quantity, 64)
				rate, _ := strconv.ParseFloat(buy.Rate, 64)
				result.Bids = append(
					result.Bids,
					common.PriceEntry{
						quantity,
						rate,
					},
				)
			}
			for _, sell := range resp_data.Asks {
				quantity, _ := strconv.ParseFloat(sell.Quantity, 64)
				rate, _ := strconv.ParseFloat(sell.Rate, 64)
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

func (self *Binance) FetchPriceData(timepoint uint64) (map[common.TokenPairID]common.ExchangePrice, error) {
	wait := sync.WaitGroup{}
	data := sync.Map{}
	pairs := self.pairs
	result := map[common.TokenPairID]common.ExchangePrice{}

	for _, pair := range pairs {
		wait.Add(1)
		go self.FetchOnePairData(&wait, pair, &data, timepoint)
	}
	wait.Wait()
	data.Range(func(key, value interface{}) bool {
		result[key.(common.TokenPairID)] = value.(common.ExchangePrice)
		return true
	})
	return result, nil
}

func (self Binance) FetchPriceDataUsingSocket(exchangePriceChan chan *sync.Map) {
	data := sync.Map{}
	pairs := self.pairs

	for _, pair := range pairs {
		go self.interf.SocketFetchOnePairData(pair, &data, exchangePriceChan)
	}
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

func (self *Binance) FetchOnePairTradeHistory(
	wait *sync.WaitGroup,
	data *sync.Map,
	pair common.TokenPair,
	timepoint uint64) {

	defer wait.Done()
	result := []common.TradeHistory{}
	resp, err := self.interf.GetAccountTradeHistory(pair.Base, pair.Quote, 0, timepoint)
	if err != nil {
		log.Printf("Cannot fetch data for pair %s%s: %s", pair.Base.ID, pair.Quote.ID, err.Error())
	}
	pairString := pair.PairID()
	for _, trade := range resp {
		price, _ := strconv.ParseFloat(trade.Price, 64)
		quantity, _ := strconv.ParseFloat(trade.Qty, 64)
		historyType := "sell"
		if trade.IsBuyer {
			historyType = "buy"
		}
		tradeHistory := common.TradeHistory{
			strconv.FormatUint(trade.ID, 10),
			price,
			quantity,
			historyType,
			trade.Time,
		}
		result = append(result, tradeHistory)
	}
	data.Store(pairString, result)
}

func (self *Binance) FetchTradeHistory(timepoint uint64) (map[common.TokenPairID][]common.TradeHistory, error) {
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

func (self *Binance) DepositStatus(id common.ActivityID, timepoint uint64) (string, error) {
	idParts := strings.Split(id.EID, "|")
	if len(idParts) != 3 {
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

func (self *Binance) WithdrawStatus(id common.ActivityID, timepoint uint64) (string, string, error) {
	withdrawID := id.EID
	startTime := timepoint - 86400000
	endTime := timepoint
	withdraws, err := self.interf.WithdrawHistory(startTime, endTime)
	if err != nil || !withdraws.Success {
		return "", "", err
	} else {
		for _, withdraw := range withdraws.Withdrawals {
			if withdraw.ID == withdrawID {
				if withdraw.Status == 3 || withdraw.Status == 5 || withdraw.Status == 6 {
					return "done", withdraw.TxID, nil
				} else {
					return "", withdraw.TxID, nil
				}
			}
		}
		return "", "", errors.New("Withdrawal doesn't exist. This shouldn't happen unless tx returned from withdrawal from binance and activity ID are not consistently designed")
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

func NewBinance(addressConfig map[string]string, feeConfig common.ExchangeFees, interf BinanceInterface) *Binance {
	pairs, fees := getExchangePairsAndFeesFromConfig(addressConfig, feeConfig, "binance")
	return &Binance{
		interf,
		pairs,
		common.NewExchangeAddresses(),
		common.NewExchangeInfo(),
		fees,
	}
}
