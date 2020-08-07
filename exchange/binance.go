package exchange

import (
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	ethereum "github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/settings"
)

const (
	binanceEpsilon float64 = 0.0000001 // 10e-7
	batchSize      int     = 4
)

type Binance struct {
	interf  BinanceInterface
	storage BinanceStorage
	setting Setting
	l       *zap.SugaredLogger
}

func (bn *Binance) TokenAddresses() (map[string]ethereum.Address, error) {
	addresses, err := bn.setting.GetDepositAddresses(settings.Binance)
	if err != nil {
		return nil, err
	}
	return addresses.GetData(), nil
}

func (bn *Binance) MarshalText() (text []byte, err error) {
	return []byte(bn.ID()), nil
}

// Address returns the deposit address of a token on Binance.
// It will prioritize the live adress from Binance over the current address in storage
func (bn *Binance) Address(token common.Token) (ethereum.Address, bool) {
	liveAddress, err := bn.interf.GetDepositAddress(token.ID)
	if err != nil || liveAddress.Address == "" {
		if err != nil {
			bn.l.Warnw("Get Binance live deposit address for token failed. Use the currently available address instead", "tokenID", token.ID, "err", err)
		} else {
			bn.l.Warnw("Get Binance live deposit address for token failed: The address replied is empty. Use the currently available address instead", "tokenID", token.ID)
		}
		addrs, uErr := bn.setting.GetDepositAddresses(settings.Binance)
		if uErr != nil {
			bn.l.Warnw("get address of token in Binance exchange failed, it will be considered as not supported", "tokenID", token.ID, "err", err)
			return ethereum.Address{}, false
		}
		return addrs.Get(token.ID)
	}
	bn.l.Infof("Got Binance live deposit address for token %s, attempt to update it to current setting", token.ID)
	addrs := common.NewExchangeAddresses()
	addrs.Update(token.ID, ethereum.HexToAddress(liveAddress.Address))
	if err = bn.setting.UpdateDepositAddress(settings.Binance, *addrs, common.GetTimepoint()); err != nil {
		bn.l.Warnw("cannot update deposit address for token on Binance", "tokenID", token.ID, "err", err)
	}
	return ethereum.HexToAddress(liveAddress.Address), true
}

// UpdateDepositAddress update deposit address from binance api
func (bn *Binance) UpdateDepositAddress(token common.Token, address string) error {
	liveAddress, err := bn.interf.GetDepositAddress(token.ID)
	if err != nil || liveAddress.Address == "" {
		if err != nil {
			bn.l.Warnw("Get Binance live deposit address for token failed. Use the currently available address instead", "tokenID", token.ID, "err", err)
		} else {
			bn.l.Warnw("Get Binance live deposit address for token failed. The address replied is empty. Use the currently available address instead", "tokenID", token.ID)
		}
		addrs := common.NewExchangeAddresses()
		addrs.Update(token.ID, ethereum.HexToAddress(address))
		return bn.setting.UpdateDepositAddress(settings.Binance, *addrs, common.GetTimepoint())
	}
	bn.l.Infof("Got Binance live deposit address for token %s, attempt to update it to current setting", token.ID)
	addrs := common.NewExchangeAddresses()
	addrs.Update(token.ID, ethereum.HexToAddress(liveAddress.Address))
	return bn.setting.UpdateDepositAddress(settings.Binance, *addrs, common.GetTimepoint())
}

func (bn *Binance) precisionFromStepSize(stepSize string) int {
	re := regexp.MustCompile("0*$")
	parts := strings.Split(re.ReplaceAllString(stepSize, ""), ".")
	if len(parts) > 1 {
		return len(parts[1])
	}
	return 0
}

// GetLiveExchangeInfo queries the Exchange Endpoint for exchange precision and limit of a certain pair ID
// It return error if occurs.
func (bn *Binance) GetLiveExchangeInfos(tokenPairIDs []common.TokenPairID) (common.ExchangeInfo, error) {
	result := make(common.ExchangeInfo)
	exchangeInfo, err := bn.interf.GetExchangeInfo()
	if err != nil {
		return result, err
	}
	symbols := exchangeInfo.Symbols
	for _, pairID := range tokenPairIDs {
		exchangePrecisionLimit, ok := bn.getPrecisionLimitFromSymbols(pairID, symbols)
		if !ok {
			return result, fmt.Errorf("binance Exchange Info reply doesn't contain token pair %s", string(pairID))
		}
		result[pairID] = exchangePrecisionLimit
	}
	return result, nil
}

// getPrecisionLimitFromSymbols find the pairID amongs symbols from exchanges,
// return ExchangePrecisionLimit of that pair and true if the pairID exist amongs symbols, false if otherwise
func (bn *Binance) getPrecisionLimitFromSymbols(pair common.TokenPairID, symbols []BinanceSymbol) (common.ExchangePrecisionLimit, bool) {
	var result common.ExchangePrecisionLimit
	pairName := strings.ToUpper(strings.Replace(string(pair), "-", "", 1))
	for _, symbol := range symbols {
		if strings.ToUpper(symbol.Symbol) == pairName {
			//update precision
			result.Precision.Amount = symbol.BaseAssetPrecision
			result.Precision.Price = symbol.QuoteAssetPrecision
			// update limit
			for _, filter := range symbol.Filters {
				if filter.FilterType == "LOT_SIZE" {
					// update amount min
					minQuantity, _ := strconv.ParseFloat(filter.MinQuantity, 64)
					result.AmountLimit.Min = minQuantity
					// update amount max
					maxQuantity, _ := strconv.ParseFloat(filter.MaxQuantity, 64)
					result.AmountLimit.Max = maxQuantity
					result.Precision.Amount = bn.precisionFromStepSize(filter.StepSize)
				}

				if filter.FilterType == "PRICE_FILTER" {
					// update price min
					minPrice, _ := strconv.ParseFloat(filter.MinPrice, 64)
					result.PriceLimit.Min = minPrice
					// update price max
					maxPrice, _ := strconv.ParseFloat(filter.MaxPrice, 64)
					result.PriceLimit.Max = maxPrice
					result.Precision.Price = bn.precisionFromStepSize(filter.TickSize)
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

func (bn *Binance) UpdatePairsPrecision() error {
	exchangeInfo, err := bn.interf.GetExchangeInfo()
	if err != nil {
		return err
	}
	symbols := exchangeInfo.Symbols
	exInfo, err := bn.GetInfo()
	if err != nil {
		return fmt.Errorf("can't get Exchange Info for Binance from persistent storage. (%s)", err)
	}
	if exInfo == nil {
		return errors.New("exchange info of Binance is nil")
	}
	for pair := range exInfo.GetData() {
		exchangePrecisionLimit, exist := bn.getPrecisionLimitFromSymbols(pair, symbols)
		if !exist {
			return fmt.Errorf("binance Exchange Info reply doesn't contain token pair %s", pair)
		}
		exInfo[pair] = exchangePrecisionLimit
	}
	return bn.setting.UpdateExchangeInfo(settings.Binance, exInfo, common.GetTimepoint())
}

func (bn *Binance) GetInfo() (common.ExchangeInfo, error) {
	return bn.setting.GetExchangeInfo(settings.Binance)
}

func (bn *Binance) GetExchangeInfo(pair common.TokenPairID) (common.ExchangePrecisionLimit, error) {
	exInfo, err := bn.setting.GetExchangeInfo(settings.Binance)
	if err != nil {
		return common.ExchangePrecisionLimit{}, err
	}
	return exInfo.Get(pair)
}

func (bn *Binance) GetFee() (common.ExchangeFees, error) {
	return bn.setting.GetFee(settings.Binance)
}

func (bn *Binance) GetMinDeposit() (common.ExchangesMinDeposit, error) {
	return bn.setting.GetMinDeposit(settings.Binance)
}

// ID must return the exact string or else simulation will fail
func (bn *Binance) ID() common.ExchangeID {
	return common.ExchangeID(settings.Binance.String())
}

func (bn *Binance) TokenPairs() ([]common.TokenPair, error) {
	result := []common.TokenPair{}
	exInfo, err := bn.setting.GetExchangeInfo(settings.Binance)
	if err != nil {
		return nil, err
	}
	for pair := range exInfo.GetData() {
		pairIDs := strings.Split(string(pair), "-")
		if len(pairIDs) != 2 {
			return result, fmt.Errorf("binance PairID %s is malformed", string(pair))
		}
		tok1, uErr := bn.setting.GetTokenByID(pairIDs[0])
		if uErr != nil {
			return result, fmt.Errorf("binance cant get Token %s, %s", pairIDs[0], uErr)
		}
		tok2, uErr := bn.setting.GetTokenByID(pairIDs[1])
		if uErr != nil {
			return result, fmt.Errorf("binance cant get Token %s, %s", pairIDs[1], uErr)
		}
		tokPair := common.TokenPair{
			Base:  tok1,
			Quote: tok2,
		}
		result = append(result, tokPair)
	}
	return result, nil
}

func (bn *Binance) Name() string {
	return "binance"
}

func (bn *Binance) QueryOrder(symbol string, id uint64) (done float64, remaining float64, finished bool, err error) {
	result, err := bn.interf.OrderStatus(symbol, id)
	if err != nil {
		return 0, 0, false, err
	}
	done, _ = strconv.ParseFloat(result.ExecutedQty, 64)
	total, _ := strconv.ParseFloat(result.OrigQty, 64)
	return done, total - done, total-done < binanceEpsilon, nil
}

func (bn *Binance) Trade(tradeType string, base common.Token, quote common.Token, rate float64, amount float64, timepoint uint64) (id string, done float64, remaining float64, finished bool, err error) {
	result, err := bn.interf.Trade(tradeType, base, quote, rate, amount)
	if err != nil {
		return "", 0, 0, false, err
	}
	for i := 0; i < 5; i++ { // sometime binance get trouble when query order info right after it created, so we
		// add a retry to handle it here
		done, remaining, finished, err = bn.QueryOrder(
			base.ID+quote.ID,
			result.OrderID,
		)
		if err == nil {
			break
		}
		bn.l.Errorw("failed to query order info", "err", err, "i", i, "orderID", result.OrderID, "base", base.ID, "quote", quote.ID)
		if strings.Contains(err.Error(), "Order does not exist") { // only retry if got specified error
			time.Sleep(time.Second)
			continue
		}
		break
	}
	id = strconv.FormatUint(result.OrderID, 10)
	return id, done, remaining, finished, err
}

func (bn *Binance) Withdraw(token common.Token, amount *big.Int, address ethereum.Address, timepoint uint64) (string, error) {
	tx, err := bn.interf.Withdraw(token, amount, address)
	return tx, err
}

func (bn *Binance) CancelOrder(id string, symbol string) error {
	idNo, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return err
	}
	_, err = bn.interf.CancelOrder(symbol, idNo)
	if err != nil {
		return err
	}
	return nil
}

//CancelAllOrders cancel all open orders on a symbol
func (bn *Binance) CancelAllOrders(symbol string) error {
	_, err := bn.interf.CancelAllOrders(symbol)
	if err != nil {
		return err
	}
	return nil
}

func (bn *Binance) FetchOnePairData(
	wg *sync.WaitGroup,
	baseID, quoteID string,
	data *sync.Map,
	timepoint uint64) {

	defer wg.Done()
	result := common.ExchangePrice{}

	timestamp := common.Timestamp(fmt.Sprintf("%d", timepoint))
	result.Timestamp = timestamp
	result.Valid = true
	respData, err := bn.interf.GetDepthOnePair(baseID, quoteID)
	returnTime := common.GetTimestamp()
	result.ReturnTime = returnTime
	if err != nil {
		result.Valid = false
		result.Error = err.Error()
	} else {
		if respData.Code != 0 || respData.Msg != "" {
			result.Valid = false
			result.Error = fmt.Sprintf("Code: %d, Msg: %s", respData.Code, respData.Msg)
		} else {
			for _, buy := range respData.Bids {
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
			for _, sell := range respData.Asks {
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
	data.Store(common.NewTokenPairID(baseID, quoteID), result)
}

func (bn *Binance) FetchPriceData(timepoint uint64, fetchBTCPrice bool) (map[common.TokenPairID]common.ExchangePrice, error) {
	wait := sync.WaitGroup{}
	data := sync.Map{}
	pairs, err := bn.TokenPairs()
	if err != nil {
		return nil, err
	}
	var (
		i int
		x int
	)
	for i < len(pairs) {
		for x = i; x < len(pairs) && x < i+batchSize; x++ {
			wait.Add(1)
			pair := pairs[x]
			baseID, quoteID := pair.GetBaseQuoteID()
			go bn.FetchOnePairData(&wait, baseID, quoteID, &data, timepoint)
			if fetchBTCPrice {
				wait.Add(1)
				go bn.FetchOnePairData(&wait, baseID, BtcID, &data, timepoint)
			}
		}
		if fetchBTCPrice {
			wait.Add(1)
			go bn.FetchOnePairData(&wait, bn.setting.ETHToken().ID, BtcID, &data, timepoint)
		}
		wait.Wait()
		i = x
	}
	result := map[common.TokenPairID]common.ExchangePrice{}
	data.Range(func(key, value interface{}) bool {
		//if there is conversion error, continue to next key,val
		tokenPairID, ok := key.(common.TokenPairID)
		if !ok {
			err = fmt.Errorf("key (%v) cannot be asserted to TokenPairID", key)
			return false
		}
		exPrice, ok := value.(common.ExchangePrice)
		if !ok {
			err = fmt.Errorf("value (%v) cannot be asserted to ExchangePrice", value)
			return false
		}
		result[tokenPairID] = exPrice
		return true
	})
	return result, err
}

// OpenOrders return open orders from binance
func (bn *Binance) OpenOrders() ([]common.Order, error) {
	var (
		orders = make([]common.Order, 0)
	)
	result, err := bn.interf.OpenOrders()
	if err != nil {
		return nil, err
	}
	for _, order := range result {
		price, _ := strconv.ParseFloat(order.Price, 64)
		orgQty, _ := strconv.ParseFloat(order.OrigQty, 64)
		executedQty, _ := strconv.ParseFloat(order.ExecutedQty, 64)
		orders = append(orders, common.Order{
			ID:          fmt.Sprintf("%d_%s", order.OrderID, strings.ToUpper(order.Symbol)),
			OrderID:     fmt.Sprintf("%d", order.OrderID),
			Price:       price,
			OrigQty:     orgQty,
			ExecutedQty: executedQty,
			TimeInForce: order.TimeInForce,
			Type:        order.Type,
			Side:        order.Side,
			StopPrice:   order.StopPrice,
			IcebergQty:  order.IcebergQty,
			Time:        order.Time,
			Symbol:      order.Symbol,
			Quote:       "",
		})
	}
	return orders, nil
}

func (bn *Binance) FetchEBalanceData(timepoint uint64) (common.EBalanceEntry, error) {
	result := common.EBalanceEntry{}
	result.Timestamp = common.Timestamp(fmt.Sprintf("%d", timepoint))
	result.Valid = true
	result.Error = ""
	respData, err := bn.interf.GetInfo()
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
		if respData.Code != 0 {
			result.Valid = false
			result.Error = fmt.Sprintf("Code: %d, Msg: %s", respData.Code, respData.Msg)
			result.Status = false
		} else {
			for _, b := range respData.Balances {
				tokenID := b.Asset
				_, err := bn.setting.GetTokenByID(tokenID)
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
func (bn *Binance) FetchOnePairTradeHistory(
	wait *sync.WaitGroup,
	data *sync.Map,
	pair common.TokenPair) {

	defer wait.Done()
	result := []common.TradeHistory{}
	tokenPair := fmt.Sprintf("%s-%s", pair.Base.ID, pair.Quote.ID)
	fromID, err := bn.storage.GetLastIDTradeHistory(tokenPair)
	if err != nil {
		bn.l.Warnw("Cannot get last ID trade history", "err", err, "tokenPair", tokenPair)
		return
	}
	resp, err := bn.interf.GetAccountTradeHistory(pair.Base, pair.Quote, fromID)
	if err != nil {
		bn.l.Warnw("Binance Cannot fetch data for pair",
			"base", pair.Base.ID, "quote", pair.Quote.ID, "err", err)
		return
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
func (bn *Binance) FetchTradeHistory() {
	t := time.NewTicker(10 * time.Minute)
	go func() {
		for {
			result := common.ExchangeTradeHistory{}
			data := sync.Map{}
			pairs, err := bn.TokenPairs()
			if err != nil {
				bn.l.Warnw("Binance Get Token pairs setting failed", "err", err)
				continue
			}
			wait := sync.WaitGroup{}
			var i int
			var x int
			for i < len(pairs) {
				for x = i; x < len(pairs) && x < i+batchSize; x++ {
					wait.Add(1)
					pair := pairs[x]
					go bn.FetchOnePairTradeHistory(&wait, &data, pair)
				}
				i = x
				wait.Wait()
			}
			var integrity = true
			data.Range(func(key, value interface{}) bool {
				tokenPairID, ok := key.(common.TokenPairID)
				//if there is conversion error, continue to next key,val
				if !ok {
					bn.l.Infof("Key (%v) cannot be asserted to TokenPairID", key)
					integrity = false
					return false
				}
				tradeHistories, ok := value.([]common.TradeHistory)
				if !ok {
					bn.l.Infof("Value (%v) cannot be asserted to []TradeHistory", value)
					integrity = false
					return false
				}
				result[tokenPairID] = tradeHistories
				return true
			})
			if !integrity {
				bn.l.Warnw("Binance fetch trade history returns corrupted. Try again in 10 mins")
				continue
			}
			if err := bn.storage.StoreTradeHistory(result); err != nil {
				bn.l.Warnw("Binance Store trade history", "err", err)
			}
			<-t.C
		}
	}()
}

func (bn *Binance) GetTradeHistory(fromTime, toTime uint64) (common.ExchangeTradeHistory, error) {
	return bn.storage.GetTradeHistory(fromTime, toTime)
}

func (bn *Binance) DepositStatus(id common.ActivityID, txHash, currency string, amount float64, timepoint uint64) (string, error) {
	startTime := timepoint - 86400000
	endTime := timepoint
	deposits, err := bn.interf.DepositHistory(startTime, endTime)
	if err != nil || !deposits.Success {
		return "", err
	}
	for _, deposit := range deposits.Deposits {
		if deposit.TxID == txHash {
			if deposit.Status == 1 {
				return common.ExchangeStatusDone, nil
			}
			return "", nil
		}
	}
	bn.l.Infof("Binance Deposit is not found in deposit list returned from Binance. " +
		"This might cause by wrong start/end time, please check again.")
	return "", nil
}

// WithdrawStatus return withdraw status from binance
func (bn *Binance) WithdrawStatus(id, currency string, amount float64, timepoint uint64) (string, string, float64, error) {
	startTime := timepoint - 86400000
	endTime := timepoint
	withdraws, err := bn.interf.WithdrawHistory(startTime, endTime)
	if err != nil || !withdraws.Success {
		return "", "", 0, err
	}
	for _, withdraw := range withdraws.Withdrawals {
		if withdraw.ID == id {
			if withdraw.Status == 3 || withdraw.Status == 5 {
				return common.ExchangeStatusFailed, "", withdraw.Fee, nil
			}
			if withdraw.Status == 6 {
				return common.ExchangeStatusDone, withdraw.TxID, withdraw.Fee, nil
			}
			return "", withdraw.TxID, withdraw.Fee, nil
		}
	}
	bn.l.Infof("Binance Withdrawal doesn't exist. This shouldn't happen unless tx returned from withdrawal from binance and activity ID are not consistently designed")
	return "", "", 0, nil
}

func (bn *Binance) OrderStatus(id string, base, quote string) (string, error) {
	orderID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return "", fmt.Errorf("can not parse orderID (val %s) to uint", id)
	}
	symbol := base + quote
	order, err := bn.interf.OrderStatus(symbol, orderID)
	if err != nil {
		return "", err
	}
	if order.Status == "NEW" || order.Status == "PARTIALLY_FILLED" || order.Status == "PENDING_CANCEL" {
		return "", nil
	}
	return common.ExchangeStatusDone, nil
}

func NewBinance(
	interf BinanceInterface,
	storage BinanceStorage,
	setting Setting) (*Binance, error) {
	binance := &Binance{
		interf:  interf,
		storage: storage,
		setting: setting,
		l:       zap.S(),
	}
	binance.FetchTradeHistory()
	return binance, nil
}
