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
	"time"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/settings"
	ethereum "github.com/ethereum/go-ethereum/common"
)

const (
	binanceEpsilon float64 = 0.0000001 // 10e-7
	batchSize      int     = 4
)

type Binance struct {
	interf  BinanceInterface
	storage BinanceStorage
	setting Setting
}

func (self *Binance) TokenAddresses() (map[string]ethereum.Address, error) {
	addresses, err := self.setting.GetDepositAddresses(settings.Binance)
	if err != nil {
		return nil, err
	}
	return addresses.GetData(), nil
}

func (self *Binance) MarshalText() (text []byte, err error) {
	return []byte(self.ID()), nil
}

// Address returns the deposit address of a token on Binance.
// It will prioritize the live adress from Binance over the current address in storage
func (self *Binance) Address(token common.Token) (ethereum.Address, bool) {
	liveAddress, err := self.interf.GetDepositAddress(token.ID)
	if err != nil || liveAddress.Address == "" {
		log.Printf("WARNING: Get Binance live deposit address for token %s failed: err: (%v) or the address repplied is empty . Use the currently available address instead", token.ID, err)
		addrs, uErr := self.setting.GetDepositAddresses(settings.Binance)
		if uErr != nil {
			log.Printf("WARNING: get address of token %s in Binance exchange failed:(%s), it will be considered as not supported", token.ID, err.Error())
			return ethereum.Address{}, false
		}
		return addrs.Get(token.ID)
	}
	log.Printf("Got Binance live deposit address for token %s, attempt to update it to current setting", token.ID)
	addrs := common.NewExchangeAddresses()
	addrs.Update(token.ID, ethereum.HexToAddress(liveAddress.Address))
	if err = self.setting.UpdateDepositAddress(settings.Binance, *addrs); err != nil {
		log.Printf("WARNING: cannot update deposit address for token %s on Binance: (%s)", token.ID, err.Error())
	}
	return ethereum.HexToAddress(liveAddress.Address), true
}

func (self *Binance) UpdateDepositAddress(token common.Token, address string) error {
	liveAddress, err := self.interf.GetDepositAddress(token.ID)
	if err != nil || liveAddress.Address == "" {
		log.Printf("WARNING: Get Binance live deposit address for token %s failed: err: (%v) or the address repplied is empty . Use the currently available address instead", token.ID, err)
		addrs := common.NewExchangeAddresses()
		addrs.Update(token.ID, ethereum.HexToAddress(address))
		return self.setting.UpdateDepositAddress(settings.Binance, *addrs)
	}
	log.Printf("Got Binance live deposit address for token %s, attempt to update it to current setting", token.ID)
	addrs := common.NewExchangeAddresses()
	addrs.Update(token.ID, ethereum.HexToAddress(liveAddress.Address))
	return self.setting.UpdateDepositAddress(settings.Binance, *addrs)
}

func (self *Binance) precisionFromStepSize(stepSize string) int {
	re := regexp.MustCompile("0*$")
	parts := strings.Split(re.ReplaceAllString(stepSize, ""), ".")
	if len(parts) > 1 {
		return len(parts[1])
	}
	return 0
}

// GetLiveExchangeInfo queries the Exchange Endpoint for exchange precision and limit of a certain pair ID
// It return error if occurs.
func (self *Binance) GetLiveExchangeInfos(tokenPairIDs []common.TokenPairID) (common.ExchangeInfo, error) {
	result := make(common.ExchangeInfo)
	exchangeInfo, err := self.interf.GetExchangeInfo()
	if err != nil {
		return result, err
	}
	symbols := exchangeInfo.Symbols
	for _, pairID := range tokenPairIDs {
		exchangePrecisionLimit, ok := self.getPrecisionLimitFromSymbols(pairID, symbols)
		if !ok {
			return result, fmt.Errorf("Binance Exchange Info reply doesn't contain token pair %s", string(pairID))
		}
		result[pairID] = exchangePrecisionLimit
	}
	return result, nil
}

// getPrecisionLimitFromSymbols find the pairID amongs symbols from exchanges,
// return ExchangePrecisionLimit of that pair and true if the pairID exist amongs symbols, false if otherwise
func (self *Binance) getPrecisionLimitFromSymbols(pair common.TokenPairID, symbols []BinanceSymbol) (common.ExchangePrecisionLimit, bool) {
	var result common.ExchangePrecisionLimit
	pairName := strings.ToUpper(strings.Replace(string(pair), "-", "", 1))
	for _, symbol := range symbols {
		if strings.ToUpper(symbol.Symbol) == pairName {
			//update precision
			result.Precision.Amount = symbol.BaseAssetPrecision
			result.Precision.Price = symbol.QuotePrecision
			// update limit
			for _, filter := range symbol.Filters {
				if filter.FilterType == "LOT_SIZE" {
					// update amount min
					minQuantity, _ := strconv.ParseFloat(filter.MinQuantity, 64)
					result.AmountLimit.Min = minQuantity
					// update amount max
					maxQuantity, _ := strconv.ParseFloat(filter.MaxQuantity, 64)
					result.AmountLimit.Max = maxQuantity
					result.Precision.Amount = self.precisionFromStepSize(filter.StepSize)
				}

				if filter.FilterType == "PRICE_FILTER" {
					// update price min
					minPrice, _ := strconv.ParseFloat(filter.MinPrice, 64)
					result.PriceLimit.Min = minPrice
					// update price max
					maxPrice, _ := strconv.ParseFloat(filter.MaxPrice, 64)
					result.PriceLimit.Max = maxPrice
					result.Precision.Price = self.precisionFromStepSize(filter.TickSize)
				}

				if filter.FilterType == "MIN_NOTIONAL" {
					minNotional, _ := strconv.ParseFloat(filter.MinNotional, 64)
					result.MinNotional = minNotional
				}
			}
			return result, true
		}
	}
	return result, false
}

func (self *Binance) UpdatePairsPrecision() error {
	exchangeInfo, err := self.interf.GetExchangeInfo()
	if err != nil {
		return err
	}
	symbols := exchangeInfo.Symbols
	exInfo, err := self.GetInfo()
	if err != nil {
		return fmt.Errorf("Can't get Exchange Info for Binance from persistent storage. (%s)", err)
	}
	if exInfo == nil {
		return errors.New("Exchange info of Binance is nil")
	}
	for pair := range exInfo.GetData() {
		exchangePrecisionLimit, exist := self.getPrecisionLimitFromSymbols(pair, symbols)
		if !exist {
			return fmt.Errorf("Binance Exchange Info reply doesn't contain token pair %s", pair)
		}
		exInfo[pair] = exchangePrecisionLimit
	}
	return self.setting.UpdateExchangeInfo(settings.Binance, exInfo)
}

func (self *Binance) GetInfo() (common.ExchangeInfo, error) {
	return self.setting.GetExchangeInfo(settings.Binance)
}

func (self *Binance) GetExchangeInfo(pair common.TokenPairID) (common.ExchangePrecisionLimit, error) {
	exInfo, err := self.setting.GetExchangeInfo(settings.Binance)
	if err != nil {
		return common.ExchangePrecisionLimit{}, err
	}
	return exInfo.Get(pair)
}

func (self *Binance) GetFee() (common.ExchangeFees, error) {
	return self.setting.GetFee(settings.Binance)
}

func (self *Binance) GetMinDeposit() (common.ExchangesMinDeposit, error) {
	return self.setting.GetMinDeposit(settings.Binance)
}

// ID must return the exact string or else simulation will fail
func (self *Binance) ID() common.ExchangeID {
	return common.ExchangeID(settings.Binance.String())
}

func (self *Binance) TokenPairs() ([]common.TokenPair, error) {
	result := []common.TokenPair{}
	exInfo, err := self.setting.GetExchangeInfo(settings.Binance)
	if err != nil {
		return nil, err
	}
	for pair := range exInfo.GetData() {
		pairIDs := strings.Split(string(pair), "-")
		if len(pairIDs) != 2 {
			return result, fmt.Errorf("Binance PairID %s is malformed", string(pair))
		}
		tok1, uErr := self.setting.GetTokenByID(pairIDs[0])
		if uErr != nil {
			return result, fmt.Errorf("Binance cant get Token %s, %s", pairIDs[0], uErr)
		}
		tok2, uErr := self.setting.GetTokenByID(pairIDs[1])
		if uErr != nil {
			return result, fmt.Errorf("Binance cant get Token %s, %s", pairIDs[1], uErr)
		}
		tokPair := common.TokenPair{
			Base:  tok1,
			Quote: tok2,
		}
		result = append(result, tokPair)
	}
	return result, nil
}

func (self *Binance) Name() string {
	return "binance"
}

func (self *Binance) QueryOrder(symbol string, id uint64) (done float64, remaining float64, finished bool, err error) {
	result, err := self.interf.OrderStatus(symbol, id)
	if err != nil {
		return 0, 0, false, err
	} else {
		done, _ := strconv.ParseFloat(result.ExecutedQty, 64)
		total, _ := strconv.ParseFloat(result.OrigQty, 64)
		return done, total - done, total-done < binanceEpsilon, nil
	}
}

func (self *Binance) Trade(tradeType string, base common.Token, quote common.Token, rate float64, amount float64, timepoint uint64) (id string, done float64, remaining float64, finished bool, err error) {
	result, err := self.interf.Trade(tradeType, base, quote, rate, amount)

	if err != nil {
		return "", 0, 0, false, err
	} else {
		done, remaining, finished, err := self.QueryOrder(
			base.ID+quote.ID,
			result.OrderID,
		)
		id := strconv.FormatUint(result.OrderID, 10)
		return id, done, remaining, finished, err
	}
}

func (self *Binance) Withdraw(token common.Token, amount *big.Int, address ethereum.Address, timepoint uint64) (string, error) {
	tx, err := self.interf.Withdraw(token, amount, address)
	return tx, err
}

func (self *Binance) CancelOrder(id string, base, quote string) error {
	idNo, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return err
	}
	symbol := base + quote
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
	resp_data, err := self.interf.GetDepthOnePair(pair)
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
					common.NewPriceEntry(
						quantity,
						rate,
					),
				)
			}
			for _, sell := range resp_data.Asks {
				quantity, _ := strconv.ParseFloat(sell.Quantity, 64)
				rate, _ := strconv.ParseFloat(sell.Rate, 64)
				result.Asks = append(
					result.Asks,
					common.NewPriceEntry(
						quantity,
						rate,
					),
				)
			}
		}
	}
	data.Store(pair.PairID(), result)
}

func (self *Binance) FetchPriceData(timepoint uint64) (map[common.TokenPairID]common.ExchangePrice, error) {
	wait := sync.WaitGroup{}
	data := sync.Map{}
	pairs, err := self.TokenPairs()
	if err != nil {
		return nil, err
	}
	var i int = 0
	var x int = 0
	for i < len(pairs) {
		for x = i; x < len(pairs) && x < i+batchSize; x++ {
			wait.Add(1)
			pair := pairs[x]
			go self.FetchOnePairData(&wait, pair, &data, timepoint)
		}
		wait.Wait()
		i = x
	}
	result := map[common.TokenPairID]common.ExchangePrice{}
	data.Range(func(key, value interface{}) bool {
		//if there is conversion error, continue to next key,val
		tokenPairID, ok := key.(common.TokenPairID)
		if !ok {
			err = fmt.Errorf("Key (%v) cannot be asserted to TokenPairID", key)
			return false
		}
		exPrice, ok := value.(common.ExchangePrice)
		if !ok {
			err = fmt.Errorf("Value (%v) cannot be asserted to ExchangePrice", value)
			return false
		}
		result[tokenPairID] = exPrice
		return true
	})
	return result, err
}

func (self *Binance) OpenOrdersForOnePair(
	wg *sync.WaitGroup,
	pair common.TokenPair,
	data *sync.Map,
	timepoint uint64) {

	defer wg.Done()

	result, err := self.interf.OpenOrdersForOnePair(pair)

	if err == nil {
		orders := []common.Order{}
		for _, order := range result {
			price, _ := strconv.ParseFloat(order.Price, 64)
			orgQty, _ := strconv.ParseFloat(order.OrigQty, 64)
			executedQty, _ := strconv.ParseFloat(order.ExecutedQty, 64)
			orders = append(orders, common.Order{
				ID:          fmt.Sprintf("%d_%s%s", order.OrderId, strings.ToUpper(pair.Base.ID), strings.ToUpper(pair.Quote.ID)),
				Base:        strings.ToUpper(pair.Base.ID),
				Quote:       strings.ToUpper(pair.Quote.ID),
				OrderId:     fmt.Sprintf("%d", order.OrderId),
				Price:       price,
				OrigQty:     orgQty,
				ExecutedQty: executedQty,
				TimeInForce: order.TimeInForce,
				Type:        order.Type,
				Side:        order.Side,
				StopPrice:   order.StopPrice,
				IcebergQty:  order.IcebergQty,
				Time:        order.Time,
			})
		}
		data.Store(pair.PairID(), orders)
	} else {
		log.Printf("Unsuccessful response from Binance: %s", err)
	}
}

func (self *Binance) FetchEBalanceData(timepoint uint64) (common.EBalanceEntry, error) {
	result := common.EBalanceEntry{}
	result.Timestamp = common.Timestamp(fmt.Sprintf("%d", timepoint))
	result.Valid = true
	result.Error = ""
	resp_data, err := self.interf.GetInfo()
	result.ReturnTime = common.GetTimestamp()
	if err != nil {
		result.Valid = false
		result.Error = err.Error()
		result.Status = false
	} else {
		result.AvailableBalance = map[string]float64{}
		result.LockedBalance = map[string]float64{}
		result.DepositBalance = map[string]float64{}
		result.Status = true
		if resp_data.Code != 0 {
			result.Valid = false
			result.Error = fmt.Sprintf("Code: %d, Msg: %s", resp_data.Code, resp_data.Msg)
			result.Status = false
		} else {
			for _, b := range resp_data.Balances {
				tokenID := b.Asset
				_, err := self.setting.GetTokenByID(tokenID)
				if err == nil {
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

//FetchOnePairTradeHistory fetch trade history for one pair from exchange
func (self *Binance) FetchOnePairTradeHistory(
	wait *sync.WaitGroup,
	data *sync.Map,
	pair common.TokenPair) {

	defer wait.Done()
	result := []common.TradeHistory{}
	tokenPair := fmt.Sprintf("%s-%s", pair.Base.ID, pair.Quote.ID)
	fromID, err := self.storage.GetLastIDTradeHistory(tokenPair)
	if err != nil {
		log.Printf("Cannot get last ID trade history: %s", err.Error())
	}
	resp, err := self.interf.GetAccountTradeHistory(pair.Base, pair.Quote, fromID)
	if err != nil {
		log.Printf("Binance Cannot fetch data for pair %s%s: %s", pair.Base.ID, pair.Quote.ID, err.Error())
	}
	pairString := pair.PairID()
	for _, trade := range resp {
		price, _ := strconv.ParseFloat(trade.Price, 64)
		quantity, _ := strconv.ParseFloat(trade.Qty, 64)
		historyType := "sell"
		if trade.IsBuyer {
			historyType = "buy"
		}
		tradeHistory := common.NewTradeHistory(
			strconv.FormatUint(trade.ID, 10),
			price,
			quantity,
			historyType,
			trade.Time,
		)
		result = append(result, tradeHistory)
	}
	data.Store(pairString, result)
}

//FetchTradeHistory get all trade history for all tokens in the exchange
func (self *Binance) FetchTradeHistory() {
	t := time.NewTicker(10 * time.Minute)
	go func() {
		for {
			result := common.ExchangeTradeHistory{}
			data := sync.Map{}
			pairs, err := self.TokenPairs()
			if err != nil {
				log.Printf("Binance Get Token pairs setting failed (%s)", err.Error())
				continue
			}
			wait := sync.WaitGroup{}
			var i int
			var x int
			for i < len(pairs) {
				for x = i; x < len(pairs) && x < i+batchSize; x++ {
					wait.Add(1)
					pair := pairs[x]
					go self.FetchOnePairTradeHistory(&wait, &data, pair)
				}
				i = x
				wait.Wait()
			}
			var integrity bool = true
			data.Range(func(key, value interface{}) bool {
				tokenPairID, ok := key.(common.TokenPairID)
				//if there is conversion error, continue to next key,val
				if !ok {
					log.Printf("Key (%v) cannot be asserted to TokenPairID", key)
					integrity = false
					return false
				}
				tradeHistories, ok := value.([]common.TradeHistory)
				if !ok {
					log.Printf("Value (%v) cannot be asserted to []TradeHistory", value)
					integrity = false
					return false
				}
				result[tokenPairID] = tradeHistories
				return true
			})
			if !integrity {
				log.Print("Binance fetch trade history returns corrupted. Try again in 10 mins")
				continue
			}
			if err := self.storage.StoreTradeHistory(result); err != nil {
				log.Printf("Binance Store trade history error: %s", err.Error())
			}
			<-t.C
		}
	}()
}

func (self *Binance) GetTradeHistory(fromTime, toTime uint64) (common.ExchangeTradeHistory, error) {
	return self.storage.GetTradeHistory(fromTime, toTime)
}

func (self *Binance) DepositStatus(id common.ActivityID, txHash, currency string, amount float64, timepoint uint64) (string, error) {
	startTime := timepoint - 86400000
	endTime := timepoint
	deposits, err := self.interf.DepositHistory(startTime, endTime)
	if err != nil || !deposits.Success {
		return "", err
	} else {
		for _, deposit := range deposits.Deposits {
			if deposit.TxID == txHash {
				if deposit.Status == 1 {
					return "done", nil
				} else {
					return "", nil
				}
			}
		}
		log.Printf("Binance Deposit is not found in deposit list returned from Binance. This might cause by wrong start/end time, please check again.")
		return "", nil
	}
}

func (self *Binance) WithdrawStatus(id, currency string, amount float64, timepoint uint64) (string, string, error) {
	startTime := timepoint - 86400000
	endTime := timepoint
	withdraws, err := self.interf.WithdrawHistory(startTime, endTime)
	if err != nil || !withdraws.Success {
		return "", "", err
	} else {
		for _, withdraw := range withdraws.Withdrawals {
			if withdraw.ID == id {
				if withdraw.Status == 3 || withdraw.Status == 5 || withdraw.Status == 6 {
					return "done", withdraw.TxID, nil
				} else {
					return "", withdraw.TxID, nil
				}
			}
		}
		log.Printf("Binance Withdrawal doesn't exist. This shouldn't happen unless tx returned from withdrawal from binance and activity ID are not consistently designed")
		return "", "", nil
	}
}

func (self *Binance) OrderStatus(id string, base, quote string) (string, error) {
	orderID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		panic(err)
	}
	symbol := base + quote
	order, err := self.interf.OrderStatus(symbol, orderID)
	if err != nil {
		return "", err
	}
	if order.Status == "NEW" || order.Status == "PARTIALLY_FILLED" || order.Status == "PENDING_CANCEL" {
		return "", nil
	} else {
		return "done", nil
	}
}

func NewBinance(
	interf BinanceInterface,
	storage BinanceStorage,
	setting Setting) (*Binance, error) {
	binance := &Binance{
		interf,
		storage,
		setting,
	}
	binance.FetchTradeHistory()
	return binance, nil
}
