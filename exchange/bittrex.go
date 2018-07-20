package exchange

import (
	"errors"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/settings"
	ethereum "github.com/ethereum/go-ethereum/common"
)

const bittrexEpsilon float64 = 0.000001

type Bittrex struct {
	interf  BittrexInterface
	storage BittrexStorage
	setting Setting
}

func (self *Bittrex) TokenAddresses() (map[string]ethereum.Address, error) {
	addrs, err := self.setting.GetDepositAddresses(settings.Bittrex)
	if err != nil {
		return nil, err
	}
	return addrs.GetData(), nil
}

func (self *Bittrex) MarshalText() (text []byte, err error) {
	return []byte(self.ID()), nil
}

// Address return the deposit address of a token and return true if token is supported in the exchange.
// Otherwise return false. This function will prioritize live address from exchange above the current stored address.
func (self *Bittrex) Address(token common.Token) (ethereum.Address, bool) {
	liveAddress, err := self.interf.GetDepositAddress(token.ID)
	if err != nil || liveAddress.Result.Address == "" {
		log.Printf("WARNING: Get Bittrex live deposit address for token %s failed: err: (%v) or the address repplied is empty . Use the currently available address instead", token.ID, err)
		addrs, uErr := self.setting.GetDepositAddresses(settings.Bittrex)
		if uErr != nil {
			log.Printf("WARNING: get address of token %s in Bittrex exchange failed:(%s), it will be considered as not supported", token.ID, err.Error())
			return ethereum.Address{}, false
		}
		return addrs.Get(token.ID)
	}
	log.Printf("Got Bittrex live deposit address for token %s, attempt to update it to current setting", token.ID)
	addrs := common.NewExchangeAddresses()
	addrs.Update(token.ID, ethereum.HexToAddress(liveAddress.Result.Address))
	if err = self.setting.UpdateDepositAddress(settings.Bittrex, *addrs); err != nil {
		log.Printf("WARNING: can not update deposit address for token %s on Bittrex: (%s)", token.ID, err.Error())
	}
	return ethereum.HexToAddress(liveAddress.Result.Address), true
}

func (self *Bittrex) GetFee() (common.ExchangeFees, error) {
	return self.setting.GetFee(settings.Bittrex)
}

func (self *Bittrex) GetMinDeposit() (common.ExchangesMinDeposit, error) {
	return self.setting.GetMinDeposit(settings.Bittrex)
}

func (self *Bittrex) UpdateDepositAddress(token common.Token, address string) error {
	liveAddress, err := self.interf.GetDepositAddress(token.ID)
	if err != nil || liveAddress.Result.Address == "" {
		log.Printf("WARNING: Get Bittrex live deposit address for token %s failed: err: (%v) or the address repplied is empty . Use the currently available address instead", token.ID, err)
		addrs := common.NewExchangeAddresses()
		addrs.Update(token.ID, ethereum.HexToAddress(address))
		return self.setting.UpdateDepositAddress(settings.Bittrex, *addrs)
	}
	log.Printf("Got Bittrex live deposit address for token %s, attempt to update it to current setting", token.ID)
	addrs := common.NewExchangeAddresses()
	addrs.Update(token.ID, ethereum.HexToAddress(liveAddress.Result.Address))
	return self.setting.UpdateDepositAddress(settings.Bittrex, *addrs)
}

// GetLiveExchangeInfos querry the Exchange Endpoint for exchange precision and limit of a list of tokenPairIDs
// It return error if occurs.
func (self *Bittrex) GetLiveExchangeInfos(tokenPairIDs []common.TokenPairID) (common.ExchangeInfo, error) {
	result := make(common.ExchangeInfo)
	exchangeInfo, err := self.interf.GetExchangeInfo()
	if err != nil {
		return result, err
	}
	symbols := exchangeInfo.Pairs
	for _, pairID := range tokenPairIDs {
		exchangePrecisionLimit, ok := self.getPrecisionLimitFromSymbols(pairID, symbols)
		if !ok {
			return result, fmt.Errorf("Bittrex Exchange Info reply doesn't contain token pair %s", string(pairID))
		}
		result[pairID] = exchangePrecisionLimit
	}
	return result, nil
}

// getPrecisionLimitFromSymbols find the pairID amongs symbols from exchanges,
// return ExchangePrecisionLimit of that pair and true if the pairID exist amongs symbols, false if otherwise
func (self *Bittrex) getPrecisionLimitFromSymbols(pair common.TokenPairID, symbols []BittPairInfo) (common.ExchangePrecisionLimit, bool) {
	var result common.ExchangePrecisionLimit
	pairName := strings.ToUpper(strings.Replace(string(pair), "-", "", 1))
	for _, symbol := range symbols {
		symbolName := strings.ToUpper(symbol.Base + symbol.Quote)
		if symbolName == pairName {
			//update precision
			result.Precision.Amount = 8
			result.Precision.Price = 8
			// update limit
			result.AmountLimit.Min = symbol.MinAmount
			result.MinNotional = 0.02
			return result, true
		}
	}
	return result, false
}

func (self *Bittrex) UpdatePairsPrecision() error {
	exchangeInfo, err := self.interf.GetExchangeInfo()
	if err != nil {
		return err
	}
	symbols := exchangeInfo.Pairs
	exInfo, err := self.GetInfo()
	if err != nil {
		return fmt.Errorf("Can't get Exchange Info for Bittrex from persistent storage. (%s)", err)
	}
	if exInfo == nil {
		return errors.New("Exchange info of Bittrex is nil")
	}
	for pair := range exInfo.GetData() {
		exchangePrecisionLimit, exist := self.getPrecisionLimitFromSymbols(pair, symbols)
		if !exist {
			return fmt.Errorf("Bittrex Exchange Info reply doesn't contain token pair %s", pair)
		}
		exInfo[pair] = exchangePrecisionLimit
	}
	return self.setting.UpdateExchangeInfo(settings.Bittrex, exInfo)
}

func (self *Bittrex) GetExchangeInfo(pair common.TokenPairID) (common.ExchangePrecisionLimit, error) {
	exInfo, err := self.setting.GetExchangeInfo(settings.Bittrex)
	if err != nil {
		return common.ExchangePrecisionLimit{}, err
	}
	return exInfo.Get(pair)
}

func (self *Bittrex) GetInfo() (common.ExchangeInfo, error) {
	return self.setting.GetExchangeInfo(settings.Bittrex)
}

// ID must return the exact string or else simulation will fail
func (self *Bittrex) ID() common.ExchangeID {
	return common.ExchangeID(settings.Bittrex.String())
}

func (self *Bittrex) TokenPairs() ([]common.TokenPair, error) {
	result := []common.TokenPair{}
	exInfo, err := self.setting.GetExchangeInfo(settings.Bittrex)
	if err != nil {
		return nil, err
	}
	for pair := range exInfo.GetData() {
		pairIDs := strings.Split(string(pair), "-")
		if len(pairIDs) != 2 {
			return result, fmt.Errorf("Bittrex PairID %s is malformed", string(pair))
		}
		tok1, uErr := self.setting.GetTokenByID(pairIDs[0])
		if uErr != nil {
			return result, fmt.Errorf("Bittrex cant get Token %s, %s", pairIDs[0], uErr)
		}
		tok2, uErr := self.setting.GetTokenByID(pairIDs[1])
		if uErr != nil {
			return result, fmt.Errorf("Bittrex cant get Token %s, %s", pairIDs[1], uErr)
		}
		tokPair := common.TokenPair{
			Base:  tok1,
			Quote: tok2,
		}
		result = append(result, tokPair)
	}
	return result, nil
}

func (self *Bittrex) Name() string {
	return "bittrex"
}

func (self *Bittrex) QueryOrder(uuid string, timepoint uint64) (float64, float64, bool, error) {
	result, err := self.interf.OrderStatus(uuid)
	if err != nil {
		return 0, 0, false, err
	}
	remaining := result.Result.QuantityRemaining
	done := result.Result.Quantity - remaining
	return done, remaining, remaining < bittrexEpsilon, nil
}

func (self *Bittrex) Trade(tradeType string, base common.Token, quote common.Token, rate float64, amount float64, timepoint uint64) (string, float64, float64, bool, error) {
	result, err := self.interf.Trade(tradeType, base, quote, rate, amount)

	if err != nil {
		return "", 0, 0, false, errors.New("Trade rejected by Bittrex")
	}
	if result.Success {
		uuid := result.Result["uuid"]
		done, remaining, finished, err := self.QueryOrder(
			uuid, timepoint+20)
		return uuid, done, remaining, finished, err
	}
	return "", 0, 0, false, errors.New(result.Error)
}

func (self *Bittrex) Withdraw(token common.Token, amount *big.Int, address ethereum.Address, timepoint uint64) (string, error) {
	resp, err := self.interf.Withdraw(token, amount, address)
	if err != nil {
		return "", err
	}
	if resp.Success {
		return resp.Result["uuid"], nil
	}
	return "", errors.New(resp.Error)
}

func bitttimestampToUint64(input string) uint64 {
	var t time.Time
	var err error
	len := len(input)
	if len == 23 {
		t, err = time.Parse("2006-01-02T15:04:05.000", input)
	} else if len == 22 {
		t, err = time.Parse("2006-01-02T15:04:05.00", input)
	} else if len == 21 {
		t, err = time.Parse("2006-01-02T15:04:05.0", input)
	}
	if err != nil {
		panic(err)
	}
	return uint64(t.UnixNano() / int64(time.Millisecond))
}

func (self *Bittrex) DepositStatus(
	id common.ActivityID, txHash, currency string, amount float64, timepoint uint64) (string, error) {
	timestamp := id.Timepoint
	idParts := strings.Split(id.EID, "|")
	if len(idParts) != 3 {
		// here, the exchange id part in id is malformed
		// 1. because analytic didn't pass original ID
		// 2. id is not constructed correctly in a form of uuid + "|" + token + "|" + amount
		return "", errors.New("Invalid deposit id")
	}
	amount, err := strconv.ParseFloat(idParts[2], 64)
	if err != nil {
		panic(err)
	}
	histories, err := self.interf.DepositHistory(currency)
	if err != nil {
		return "", err
	}
	for _, deposit := range histories.Result {
		log.Printf("Bittrex deposit history check: %v %v %v %v",
			deposit.Currency == currency,
			deposit.Amount-amount < bittrexEpsilon,
			bitttimestampToUint64(deposit.LastUpdated) > timestamp/uint64(time.Millisecond),
			self.storage.IsNewBittrexDeposit(deposit.Id, id),
		)
		log.Printf("deposit.Currency: %s", deposit.Currency)
		log.Printf("currency: %s", currency)
		log.Printf("deposit.Amount: %f", deposit.Amount)
		log.Printf("amount: %f", amount)
		log.Printf("deposit.LastUpdated: %d", bitttimestampToUint64(deposit.LastUpdated))
		log.Printf("timestamp: %d", timestamp/uint64(time.Millisecond))
		log.Printf("is new deposit: %t", self.storage.IsNewBittrexDeposit(deposit.Id, id))
		if deposit.Currency == currency &&
			deposit.Amount-amount < bittrexEpsilon &&
			bitttimestampToUint64(deposit.LastUpdated) > timestamp/uint64(time.Millisecond) &&
			self.storage.IsNewBittrexDeposit(deposit.Id, id) {
			if err := self.storage.RegisterBittrexDeposit(deposit.Id, id); err != nil {
				log.Printf("Register bittrex deposit error: %s", err.Error())
			}
			return "done", nil
		}
	}
	return "", nil
}

func (self *Bittrex) CancelOrder(id, base, quote string) error {
	resp, err := self.interf.CancelOrder(id)
	if err != nil {
		return err
	}
	if resp.Success {
		return nil
	}
	return errors.New(resp.Error)
}

func (self *Bittrex) WithdrawStatus(id, currency string, amount float64, timepoint uint64) (string, string, error) {
	histories, err := self.interf.WithdrawHistory(currency)
	if err != nil {
		return "", "", err
	}
	for _, withdraw := range histories.Result {
		if withdraw.PaymentUuid == id {
			if withdraw.PendingPayment {
				return "", withdraw.TxId, nil
			}
			return "done", withdraw.TxId, nil
		}
	}
	log.Printf("Withdraw with uuid " + id + " of currency " + currency + " is not found on bittrex")
	return "", "", nil
}

func (self *Bittrex) OrderStatus(uuid string, base, quote string) (string, error) {
	resp_data, err := self.interf.OrderStatus(uuid)
	if err != nil {
		return "", err
	}
	if resp_data.Result.IsOpen {
		return "", nil
	}
	return "done", nil
}

func (self *Bittrex) FetchOnePairData(wq *sync.WaitGroup, pair common.TokenPair, data *sync.Map, timepoint uint64) {
	defer wq.Done()
	result := common.ExchangePrice{}
	result.Timestamp = common.Timestamp(fmt.Sprintf("%d", timepoint))
	result.Valid = true
	onePairData, err := self.interf.FetchOnePairData(pair)
	returnTime := common.GetTimestamp()
	result.ReturnTime = returnTime
	if err != nil {
		result.Valid = false
		result.Error = err.Error()
	} else {
		if !onePairData.Success {
			result.Valid = false
			result.Error = onePairData.Msg
		} else {
			for _, buy := range onePairData.Result["buy"] {
				result.Bids = append(
					result.Bids,
					common.NewPriceEntry(
						buy["Quantity"],
						buy["Rate"],
					),
				)
			}
			for _, sell := range onePairData.Result["sell"] {
				result.Asks = append(
					result.Asks,
					common.NewPriceEntry(
						sell["Quantity"],
						sell["Rate"],
					),
				)
			}
		}
	}
	data.Store(pair.PairID(), result)
}

func (self *Bittrex) FetchPriceData(timepoint uint64) (map[common.TokenPairID]common.ExchangePrice, error) {
	wait := sync.WaitGroup{}
	data := sync.Map{}
	pairs, err := self.TokenPairs()
	if err != nil {
		return nil, err
	}
	for _, pair := range pairs {
		wait.Add(1)
		go self.FetchOnePairData(&wait, pair, &data, timepoint)
	}
	wait.Wait()
	result := map[common.TokenPairID]common.ExchangePrice{}
	data.Range(func(key, value interface{}) bool {
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

func (self *Bittrex) FetchEBalanceData(timepoint uint64) (common.EBalanceEntry, error) {
	result := common.EBalanceEntry{}
	result.Timestamp = common.Timestamp(fmt.Sprintf("%d", timepoint))
	result.Valid = true
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
		if resp_data.Success {
			for _, b := range resp_data.Result {
				tokenID := b.Currency
				_, err := self.setting.GetTokenByID(tokenID)
				if err == nil {
					result.AvailableBalance[tokenID] = b.Available
					result.DepositBalance[tokenID] = b.Pending
					result.LockedBalance[tokenID] = 0
				}
			}
			// check if bittrex returned balance for all of the
			// supported token.
			// If it didn't, it is considered invalid
			depositAddresses, err := self.setting.GetDepositAddresses(settings.Bittrex)
			if err != nil {
				return result, fmt.Errorf("Can't Get deposit addresses of Bittrex for validation (%s)", err)
			}
			if len(result.AvailableBalance) != len(depositAddresses) {
				result.Valid = false
				result.Error = "Bittrex didn't return balance for all supported tokens"
			}
		} else {
			result.Valid = false
			result.Error = resp_data.Error
		}
	}
	return result, nil
}

func (self *Bittrex) FetchOnePairTradeHistory(
	wait *sync.WaitGroup,
	data *sync.Map,
	pair common.TokenPair,
	timepoint uint64) {

	defer wait.Done()
	result := []common.TradeHistory{}
	resp, err := self.interf.GetAccountTradeHistory(pair.Base, pair.Quote)
	if err != nil {
		log.Printf("Cannot fetch data for pair %s%s: %s", pair.Base.ID, pair.Quote.ID, err.Error())
	}
	for _, trade := range resp.Result {
		t, _ := time.Parse("2014-07-09T04:01:00.667", trade.TimeStamp)
		historyType := "sell"
		if trade.OrderType == "LIMIT_BUY" {
			historyType = "buy"
		}
		tradeHistory := common.NewTradeHistory(
			trade.OrderUuid,
			trade.Price,
			trade.Quantity,
			historyType,
			common.TimeToTimepoint(t),
		)
		result = append(result, tradeHistory)
	}
	pairString := pair.PairID()
	data.Store(pairString, result)
}

//FetchTradeHistory get all trade history for all pairs from bittrex exchange
func (self *Bittrex) FetchTradeHistory() {
	t := time.NewTicker(10 * time.Minute)
	go func() {
		for {
			result := map[common.TokenPairID][]common.TradeHistory{}
			timepoint := common.GetTimepoint()
			data := sync.Map{}
			pairs, err := self.TokenPairs()
			if err != nil {
				log.Printf("Bittrex fetch trade history failed (%s). This might due to pairs setting hasn't been init yet", err.Error())
				continue
			}
			wait := sync.WaitGroup{}
			for _, pair := range pairs {
				wait.Add(1)
				go self.FetchOnePairTradeHistory(&wait, &data, pair, timepoint)
			}
			wait.Wait()
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
				log.Print("Bittrex fetch trade history returns corrupted. Try again in 10 mins")
				continue
			}
			if err := self.storage.StoreTradeHistory(result); err != nil {
				log.Printf("Bittrex store trade history error: %s", err.Error())
			}
			<-t.C
		}
	}()
}

func (self *Bittrex) GetTradeHistory(fromTime, toTime uint64) (common.ExchangeTradeHistory, error) {
	return self.storage.GetTradeHistory(fromTime, toTime)
}

func NewBittrex(
	interf BittrexInterface,
	storage BittrexStorage,
	setting Setting) (*Bittrex, error) {
	bittrex := &Bittrex{
		interf,
		storage,
		setting,
	}
	bittrex.FetchTradeHistory()
	return bittrex, nil
}
