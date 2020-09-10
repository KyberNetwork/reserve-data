package exchange

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"

	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	pe "github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/blockchain"
	"github.com/KyberNetwork/reserve-data/common/gasinfo"
	huobiblockchain "github.com/KyberNetwork/reserve-data/exchange/huobi/blockchain"
	huobihttp "github.com/KyberNetwork/reserve-data/exchange/huobi/http"
	"github.com/KyberNetwork/reserve-data/settings"
)

const (
	huobiEpsilon float64 = 0.0000000001 // 10e-10
)

type Huobi struct {
	interf     HuobiInterface
	blockchain HuobiBlockchain
	storage    HuobiStorage
	setting    Setting
	l          *zap.SugaredLogger
}

func (h *Huobi) TokenAddresses() (map[string]ethereum.Address, error) {
	addrs, err := h.setting.GetDepositAddresses(settings.Huobi)
	if err != nil {
		return nil, err
	}
	return addrs.GetData(), nil
}

func (h *Huobi) MarshalText() (text []byte, err error) {
	return []byte(h.ID()), nil
}

// RealDepositAddress return the actual Huobi deposit address of a token
// It should only be used to send 2nd transaction.
func (h *Huobi) RealDepositAddress(tokenID string) (ethereum.Address, error) {
	liveAddress, err := h.interf.GetDepositAddress(tokenID)
	if err != nil || len(liveAddress.Data) == 0 || liveAddress.Data[0].Address == "" {
		if err != nil {
			h.l.Warnw("WARNING: Get Huobi live deposit address failed. Check the currently available address instead", "tokenID", tokenID, "err", err)
		} else {
			h.l.Warnw("WARNING: Get Huobi live deposit address failed, the replied address is empty. Check the currently available address instead", "tokenID", tokenID)
		}
		addrs, uErr := h.setting.GetDepositAddresses(settings.Huobi)
		if uErr != nil {
			return ethereum.Address{}, uErr
		}
		result, supported := addrs.Get(tokenID)
		if !supported {
			return result, fmt.Errorf("real deposit address of token %s is not available", tokenID)
		}
		return result, nil
	}
	return ethereum.HexToAddress(liveAddress.Data[0].Address), nil
}

// Address return the deposit address of a token in Huobi exchange.
// Due to the logic of Huobi exchange, every token if supported will be
// deposited to an Intermediator address instead.
func (h *Huobi) Address(token common.Token) (ethereum.Address, bool) {
	result := h.blockchain.GetIntermediatorAddr()
	_, err := h.RealDepositAddress(token.ID)
	//if the realDepositAddress can not be querried, that mean the token isn't supported on Huobi
	if err != nil {
		return result, false
	}
	return result, true
}

// UpdateDepositAddress update the deposit address of a token in Huobi
// It will prioritize the live address over the input address
func (h *Huobi) UpdateDepositAddress(token common.Token, address string) error {
	liveAddress, err := h.interf.GetDepositAddress(token.ID)
	if err != nil || len(liveAddress.Data) == 0 || liveAddress.Data[0].Address == "" {
		if err != nil {
			h.l.Warnw("Get Huobi live deposit address failed. Check the currently available address instead", "tokenID", token.ID, "err", err)
		} else {
			h.l.Warnw("Get Huobi live deposit address failed: the replied address is empty. Check the currently available address instead", "tokenID", token.ID)
		}
		addrs := common.NewExchangeAddresses()
		addrs.Update(token.ID, ethereum.HexToAddress(address))
		return h.setting.UpdateDepositAddress(settings.Huobi, *addrs, common.GetTimepoint())
	}
	h.l.Infof("Got Huobi live deposit address for token %s, attempt to update it to current setting", token.ID)
	addrs := common.NewExchangeAddresses()
	addrs.Update(token.ID, ethereum.HexToAddress(liveAddress.Data[0].Address))
	return h.setting.UpdateDepositAddress(settings.Huobi, *addrs, common.GetTimepoint())
}

// GetLiveExchangeInfos querry the Exchange Endpoint for exchange precision and limit of a list of tokenPairIDs
// It return error if occurs.
func (h *Huobi) GetLiveExchangeInfos(tokenPairIDs []common.TokenPairID) (common.ExchangeInfo, error) {
	result := make(common.ExchangeInfo)
	exchangeInfo, err := h.interf.GetExchangeInfo()
	if err != nil {
		return result, err
	}
	for _, pairID := range tokenPairIDs {
		exchangePrecisionLimit, ok := h.getPrecisionLimitFromSymbols(pairID, exchangeInfo)
		if !ok {
			return result, fmt.Errorf("huobi Exchange Info reply doesn't contain token pair %s", string(pairID))
		}
		result[pairID] = exchangePrecisionLimit
	}
	return result, nil
}

// getPrecisionLimitFromSymbols find the pairID amongs symbols from exchanges,
// return ExchangePrecisionLimit of that pair and true if the pairID exist amongs symbols, false if otherwise
func (h *Huobi) getPrecisionLimitFromSymbols(pair common.TokenPairID, symbols HuobiExchangeInfo) (common.ExchangePrecisionLimit, bool) {
	var result common.ExchangePrecisionLimit
	pairName := strings.ToUpper(strings.Replace(string(pair), "-", "", 1))
	for _, symbol := range symbols.Data {
		symbolName := strings.ToUpper(symbol.Base + symbol.Quote)
		if symbolName == pairName {
			result.Precision.Amount = symbol.AmountPrecision
			result.Precision.Price = symbol.PricePrecision
			result.MinNotional = 0.02
			return result, true
		}
	}
	return result, false
}

func (h *Huobi) UpdatePairsPrecision() error {
	exchangeInfo, err := h.interf.GetExchangeInfo()
	if err != nil {
		return err
	}
	exInfo, err := h.GetInfo()
	if err != nil {
		return fmt.Errorf("INFO: Can't get Exchange Info for Huobi from persistent storage (%s)", err)
	}
	if exInfo == nil {
		return errors.New("exchange info of Huobi is nil")
	}
	for pair := range exInfo.GetData() {
		exchangePrecisionLimit, exist := h.getPrecisionLimitFromSymbols(pair, exchangeInfo)
		if !exist {
			return fmt.Errorf("huobi Exchange Info reply doesn't contain token pair %s", pair)
		}
		exInfo[pair] = exchangePrecisionLimit
	}
	return h.setting.UpdateExchangeInfo(settings.Huobi, exInfo, common.GetTimepoint())
}

func (h *Huobi) GetInfo() (common.ExchangeInfo, error) {
	return h.setting.GetExchangeInfo(settings.Huobi)
}

func (h *Huobi) GetExchangeInfo(pair common.TokenPairID) (common.ExchangePrecisionLimit, error) {
	exInfo, err := h.setting.GetExchangeInfo(settings.Huobi)
	if err != nil {
		return common.ExchangePrecisionLimit{}, err
	}
	data, err := exInfo.Get(pair)
	return data, err
}

func (h *Huobi) GetFee() (common.ExchangeFees, error) {
	return h.setting.GetFee(settings.Huobi)
}

func (h *Huobi) GetMinDeposit() (common.ExchangesMinDeposit, error) {
	return h.setting.GetMinDeposit(settings.Huobi)
}

// ID must return the exact string or else simulation will fail
func (h *Huobi) ID() common.ExchangeID {
	return common.ExchangeID(settings.Huobi.String())
}

func (h *Huobi) TokenPairs() ([]common.TokenPair, error) {
	result := []common.TokenPair{}
	exInfo, err := h.setting.GetExchangeInfo(settings.Huobi)
	if err != nil {
		return nil, err
	}
	for pair := range exInfo.GetData() {
		pairIDs := strings.Split(string(pair), "-")
		if len(pairIDs) != 2 {
			return result, fmt.Errorf("huobi PairID %s is malformed", string(pair))
		}
		tok1, uErr := h.setting.GetTokenByID(pairIDs[0])
		if uErr != nil {
			return result, fmt.Errorf("huobi cant get Token %s, %s", pairIDs[0], uErr)
		}
		tok2, uErr := h.setting.GetTokenByID(pairIDs[1])
		if uErr != nil {
			return result, fmt.Errorf("huobi cant get Token %s, %s", pairIDs[1], uErr)
		}
		tokPair := common.TokenPair{
			Base:  tok1,
			Quote: tok2,
		}
		result = append(result, tokPair)
	}
	return result, nil
}

func (h *Huobi) Name() string {
	return "huobi"
}

func (h *Huobi) QueryOrder(symbol string, id uint64) (done float64, remaining float64, finished bool, err error) {
	result, err := h.interf.OrderStatus(symbol, id)
	if err != nil {
		return 0, 0, false, err
	}
	if result.Data.ExecutedQty != "" {
		done, err = strconv.ParseFloat(result.Data.ExecutedQty, 64)
		if err != nil {
			return 0, 0, false, err
		}
	}
	var total float64
	if result.Data.OrigQty != "" {
		total, err = strconv.ParseFloat(result.Data.OrigQty, 64)
		if err != nil {
			return 0, 0, false, err
		}
	}
	return done, total - done, total-done < huobiEpsilon, nil
}

func (h *Huobi) Trade(tradeType string, base common.Token, quote common.Token, rate float64, amount float64, timepoint uint64) (id string, done float64, remaining float64, finished bool, err error) {
	result, err := h.interf.Trade(tradeType, base, quote, rate, amount, timepoint)

	if err != nil {
		return "", 0, 0, false, err
	}
	var orderID uint64
	if result.OrderID != "" {
		orderID, err = strconv.ParseUint(result.OrderID, 10, 64)
		if err != nil {
			return "", 0, 0, false, err
		}
	}
	done, remaining, finished, err = h.QueryOrder(
		base.ID+quote.ID,
		orderID,
	)
	if err != nil {
		h.l.Warnw("Huobi Query order", "err", err)
	}
	return result.OrderID, done, remaining, finished, err
}

//Withdraw return withdraw id from huobi
func (h *Huobi) Withdraw(token common.Token, amount *big.Int, address ethereum.Address, timepoint uint64) (string, error) {
	withdrawID, err := h.interf.Withdraw(token, amount, address)
	if err != nil {
		return "", err
	}
	return withdrawID, err
}

// CancelOrder cancel an order from huobi
func (h *Huobi) CancelOrder(id, symbol string) error {
	idNo, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return err
	}
	result, err := h.interf.CancelOrder(symbol, idNo)
	if err != nil {
		return err
	}
	if result.Status != "ok" {
		return errors.New("Huobi Couldn't cancel order id " + id)
	}
	return nil
}

// CancelAllOrders cancel all open orders of a symbol
func (h *Huobi) CancelAllOrders(symbol string) error {
	return errors.New("huobi does not support this kind of api yet")
}

func (h *Huobi) FetchOnePairData(
	wg *sync.WaitGroup,
	baseID, quoteID string,
	data *sync.Map,
	timepoint uint64) {

	defer wg.Done()
	result := common.ExchangePrice{}

	timestamp := common.Timestamp(fmt.Sprintf("%d", timepoint))
	result.Timestamp = timestamp
	result.Valid = true
	respData, err := h.interf.GetDepthOnePair(baseID, quoteID)
	returnTime := common.GetTimestamp()
	result.ReturnTime = returnTime
	if err != nil {
		result.Valid = false
		result.Error = err.Error()
	} else {
		if respData.Status != "ok" {
			result.Valid = false
		} else {
			for _, buy := range respData.Tick.Bids {
				quantity := buy[1]
				rate := buy[0]
				result.Bids = append(
					result.Bids,
					common.NewPriceEntry(
						quantity,
						rate,
					),
				)
			}
			for _, sell := range respData.Tick.Asks {
				quantity := sell[1]
				rate := sell[0]
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

func (h *Huobi) FetchPriceData(timepoint uint64, fetchPriceData bool) (map[common.TokenPairID]common.ExchangePrice, error) {
	wait := sync.WaitGroup{}
	data := sync.Map{}
	pairs, err := h.TokenPairs()
	if err != nil {
		return nil, err
	}
	for _, pair := range pairs {
		wait.Add(1)
		baseID, quoteID := pair.GetBaseQuoteID()
		go h.FetchOnePairData(&wait, baseID, quoteID, &data, timepoint)
		if fetchPriceData {
			wait.Add(1)
			go h.FetchOnePairData(&wait, baseID, BtcID, &data, timepoint)
		}
	}
	if fetchPriceData {
		wait.Add(1)
		go h.FetchOnePairData(&wait, h.setting.ETHToken().ID, BtcID, &data, timepoint)
	}
	wait.Wait()
	result := map[common.TokenPairID]common.ExchangePrice{}
	data.Range(func(key, value interface{}) bool {
		tokenPairID, ok := key.(common.TokenPairID)
		//if there is conversion error, continue to next key,val
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

// OpenOrders return current open orders from huobi
func (h *Huobi) OpenOrders() ([]common.Order, error) {
	var (
		result   = make([]common.Order, 0)
		errGroup errgroup.Group
		mu       sync.Mutex
	)
	pairs, err := h.TokenPairs()
	if err != nil {
		return nil, err
	}
	for _, pair := range pairs {
		errGroup.Go(
			func(pair common.TokenPair) func() error {
				return func() error {
					orders, err := h.interf.OpenOrders(&pair)
					if err != nil {
						return err
					}
					for _, order := range orders.Data {
						originQty, err := strconv.ParseFloat(order.OrigQty, 64)
						if err != nil {
							return err
						}
						price, err := strconv.ParseFloat(order.Price, 64)
						if err != nil {
							return err
						}
						mu.Lock()
						result = append(result, common.Order{
							OrderID: strconv.FormatUint(order.OrderID, 10),
							OrigQty: originQty,
							Base:    strings.ToUpper(pair.Base.ID),
							Quote:   strings.ToUpper(pair.Quote.ID),
							Symbol:  order.Symbol,
							Price:   price,
						})
						mu.Unlock()
					}
					return nil
				}
			}(pair),
		)
	}
	if err := errGroup.Wait(); err != nil {
		return result, err
	}
	return result, nil
}

// FetchEBalanceData return balance from huobi
func (h *Huobi) FetchEBalanceData(timepoint uint64) (common.EBalanceEntry, error) {
	result := common.EBalanceEntry{}
	result.Timestamp = common.Timestamp(fmt.Sprintf("%d", timepoint))
	result.Valid = true
	result.Error = ""
	respData, err := h.interf.GetInfo()
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
		if respData.Status != "ok" {
			result.Valid = false
			result.Error = "Cannot fetch ebalance"
			result.Status = false
		} else {
			balances := respData.Data.List
			for _, b := range balances {
				tokenID := strings.ToUpper(b.Currency)
				_, err := h.setting.GetInternalTokenByID(tokenID)
				if err == nil {
					balance, _ := strconv.ParseFloat(b.Balance, 64)
					if b.Type == "trade" {
						result.AvailableBalance[tokenID] = balance
					} else {
						result.LockedBalance[tokenID] = balance
					}
					result.DepositBalance[tokenID] = 0
				}
			}
		}
	}
	return result, nil
}

func (h *Huobi) FetchOnePairTradeHistory(
	wait *sync.WaitGroup,
	data *sync.Map,
	pair common.TokenPair) {

	defer wait.Done()
	var result []common.TradeHistory
	resp, err := h.interf.GetAccountTradeHistory(pair.Base, pair.Quote)
	if err != nil {
		h.l.Warnw("Cannot fetch data for pair",
			"base", pair.Base.ID, "quote", pair.Quote.ID, "err", err)
		return
	}
	pairString := pair.PairID()
	for _, trade := range resp.Data {
		price, _ := strconv.ParseFloat(trade.Price, 64)
		quantity, _ := strconv.ParseFloat(trade.Amount, 64)
		historyType := tradeTypeSell
		if trade.Type == "buy-limit" {
			historyType = tradeTypeBuy
		}
		tradeHistory := common.NewTradeHistory(
			strconv.FormatUint(trade.ID, 10),
			price,
			quantity,
			historyType,
			trade.Timestamp,
		)
		result = append(result, tradeHistory)
	}
	data.Store(pairString, result)
}

//FetchTradeHistory get all trade history for all pairs from huobi exchange
func (h *Huobi) FetchTradeHistory() {
	t := time.NewTicker(10 * time.Minute)
	go func() {
		for {
			result := map[common.TokenPairID][]common.TradeHistory{}
			data := sync.Map{}
			pairs, err := h.TokenPairs()
			if err != nil {
				h.l.Warnw("Huobi fetch trade history failed. This might due to pairs setting hasn't been init yet", "err", err)
				continue
			}
			wait := sync.WaitGroup{}
			for _, pair := range pairs {
				wait.Add(1)
				go h.FetchOnePairTradeHistory(&wait, &data, pair)
			}
			wait.Wait()
			var integrity = true
			data.Range(func(key, value interface{}) bool {
				tokenPairID, ok := key.(common.TokenPairID)
				//if there is conversion error, continue to next key,val
				if !ok {
					h.l.Warnw("key cannot be asserted to TokenPairID", "key", key)
					integrity = false
					return false
				}
				tradeHistories, ok := value.([]common.TradeHistory)
				if !ok {
					h.l.Warnw("Value cannot be asserted to []TradeHistory", "value", value)
					integrity = false
					return false
				}
				result[tokenPairID] = tradeHistories
				return true
			})
			if !integrity {
				h.l.Warnw("Huobi fetch trade history returns corrupted. Try again in 10 mins")
				continue
			}
			if err := h.storage.StoreTradeHistory(result); err != nil {
				h.l.Warnw("Store trade history", "err", err)
			}
			<-t.C
		}
	}()
}

func (h *Huobi) GetTradeHistory(fromTime, toTime uint64) (common.ExchangeTradeHistory, error) {
	return h.storage.GetTradeHistory(fromTime, toTime)
}

func (h *Huobi) Send2ndTransaction(amount float64, token common.Token, exchangeAddress ethereum.Address) (*types.Transaction, error) {
	IAmount := common.FloatToBigInt(amount, token.Decimals)
	// Check balance, removed from huobi's blockchain object.
	// currBalance := h.blockchain.CheckBalance(token)
	// log.Printf("current balance of token %s is %d", token.ID, currBalance)
	// //h.blockchain.
	// if currBalance.Cmp(IAmount) < 0 {
	// 	log.Printf("balance is not enough, wait till next check")
	// 	return nil, errors.New("balance is not enough")
	// }
	var tx *types.Transaction
	gasInfo := gasinfo.GetGlobal()
	if gasInfo == nil {
		h.l.Errorw("gasInfo not setup, retry later")
		return nil, fmt.Errorf("gasInfo not setup, retry later")
	}
	recommendedPrice, err := gasInfo.GetCurrentGas()
	if err != nil {
		h.l.Errorw("failed to get gas price, use default", "err", err)
	}
	var gasPrice *big.Int
	highBoundGasPrice, err := gasInfo.MaxGas()
	if err != nil {
		h.l.Errorw("failed to receive high bound gas, use default", "err", err)
		highBoundGasPrice = 100.0
	}
	if recommendedPrice == 0 || recommendedPrice > highBoundGasPrice {
		gasPrice = common.GweiToWei(10)
	} else {
		gasPrice = common.GweiToWei(recommendedPrice)
	}
	h.l.Infof("Send2ndTransaction, gas price: %s", gasPrice.String())
	if token.ID == "ETH" {
		tx, err = h.blockchain.SendETHFromAccountToExchange(IAmount, exchangeAddress, gasPrice)
	} else {
		tx, err = h.blockchain.SendTokenFromAccountToExchange(IAmount, exchangeAddress, ethereum.HexToAddress(token.Address), gasPrice)
	}
	if err != nil {
		h.l.Warnw("Can not send transaction to exchange", "err", err, "token_id", token.ID)
		return nil, err
	}
	h.l.Infof("Transaction submitted. Tx is: %v", tx)
	return tx, nil

}

func (h *Huobi) PendingIntermediateTxs() (map[common.ActivityID]common.TXEntry, error) {
	result, err := h.storage.GetPendingIntermediateTXs()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (h *Huobi) FindTx2InPending(id common.ActivityID) (common.TXEntry, bool) {
	pendings, err := h.storage.GetPendingIntermediateTXs()
	if err != nil {
		h.l.Warnw("can't get pendings tx2 records", "err", err)
		return common.TXEntry{}, false
	}
	for actID, txentry := range pendings {
		if actID == id {
			return txentry, true
		}
	}
	return common.TXEntry{}, false
}

//FindTx2 : find Tx2 Record associates with activity ID, return
func (h *Huobi) FindTx2(id common.ActivityID) (tx2 common.TXEntry, found bool) {
	found = true
	//first look it up in permanent bucket
	tx2, err := h.storage.GetIntermedatorTx(id)
	if err != nil {
		//couldn't look for it in permanent bucket, look for it in pending bucket
		tx2, found = h.FindTx2InPending(id)
	}
	return tx2, found
}

func (h *Huobi) exchangeDepositStatus(id common.ActivityID, tx2Entry common.TXEntry, currency string, sentAmount float64) (string, error) {
	tokens, err := h.setting.GetAllTokens()
	if err != nil {
		h.l.Warnw("Huobi ERROR: Can not get list of tokens from setting", "err", err)
		return "", err
	}
	deposits, err := h.interf.DepositHistory(tokens)
	if err != nil || deposits.Status != "ok" {
		h.l.Warnw("Huobi Getting deposit history from huobi failed", "err", err, "deposits", deposits)
		return "", nil
	}
	//check tx2 deposit status from Huobi
	for _, deposit := range deposits.Data {
		h.l.Infof("deposit tx is %s, with token %s", deposit.TxHash, deposit.Currency)
		if deposit.TxHash[0:2] != "0x" {
			deposit.TxHash = "0x" + deposit.TxHash
		}
		if deposit.TxHash == tx2Entry.Hash {
			if deposit.State == "safe" || deposit.State == "confirmed" {
				data := common.NewTXEntry(tx2Entry.Hash,
					h.Name(),
					currency,
					common.MiningStatusMined,
					exchangeStatusDone,
					sentAmount,
					common.GetTimestamp(),
				)
				if err = h.storage.StoreIntermediateTx(id, data); err != nil {
					h.l.Warnw("Huobi Trying to store intermediate tx to huobi storage. Ignore it and try later", "err", err)
					return "", nil
				}
				return exchangeStatusDone, nil
			}
			//TODO : handle other states following https://github.com/huobiapi/API_Docs_en/wiki/REST_Reference#deposit-states
			h.l.Infof("Huobi Tx %s is found but the status was not safe but %s", deposit.TxHash, deposit.State)
			return "", nil
		}
	}
	h.l.Infof("Huobi Deposit doesn't exist. Huobi hasn't recognized the deposit yet or in theory, you have more than %d deposits at the same time.", len(tokens)*2)
	return "", nil
}

func (h *Huobi) process1stTx(id common.ActivityID, tx1Hash, currency string, sentAmount float64) (string, error) {
	status, blockno, err := h.blockchain.TxStatus(ethereum.HexToHash(tx1Hash))
	if err != nil {
		h.l.Warnw("Huobi Can not get TX status", "err", err)
		return "", nil
	}
	h.l.Infof("Huobi Status for Tx1 was %s at block %d ", status, blockno)
	if status == common.MiningStatusMined {
		//if it is mined, send 2nd tx.
		h.l.Infof("Found a new deposit status, which deposit %f %s. Procceed to send it to Huobi", sentAmount, currency)
		//check if the token is supported, the token can be active or inactivee
		token, err := h.setting.GetTokenByID(currency)
		if err != nil {
			return "", err
		}
		exchangeAddress, err := h.RealDepositAddress(currency)
		if err != nil {
			return "", err
		}
		tx2, err := h.Send2ndTransaction(sentAmount, token, exchangeAddress)
		if err != nil {
			h.l.Warnw("Huobi Trying to send 2nd tx failed. Will retry next time", "err", err)
			return "", nil
		}
		//store tx2 to pendingIntermediateTx
		data := common.NewTXEntry(
			tx2.Hash().Hex(),
			h.Name(),
			currency,
			common.MiningStatusSubmitted,
			"",
			sentAmount,
			common.GetTimestamp(),
		)
		if err = h.storage.StorePendingIntermediateTx(id, data); err != nil {
			h.l.Warnw("Trying to store 2nd tx to pending tx storage failed. It will be ignored and can make us to send to huobi again and the deposit will be marked as failed because the fund is not efficient", "err", err)
		}
		return "", nil
	}
	//No need to handle other blockchain status of TX1 here, since Fetcher will handle it from blockchain Status.
	return "", nil
}

func (h *Huobi) DepositStatus(id common.ActivityID, tx1Hash, currency string, sentAmount float64, timepoint uint64) (string, error) {
	var data common.TXEntry
	tx2Entry, found := h.FindTx2(id)
	//if not found, meaning there is no tx2 yet, process 1st Tx and send 2nd Tx.
	if !found {
		return h.process1stTx(id, tx1Hash, currency, sentAmount)
	}
	// if there is tx2Entry, check it blockchain status and handle the status accordingly:
	miningStatus, _, err := h.blockchain.TxStatus(ethereum.HexToHash(tx2Entry.Hash))
	if err != nil {
		return "", err
	}
	switch miningStatus {
	case common.MiningStatusMined:
		h.l.Infof("Huobi 2nd Transaction is mined. Processed to store it and check the Huobi Deposit history")
		data = common.NewTXEntry(
			tx2Entry.Hash,
			h.Name(),
			currency,
			common.MiningStatusMined,
			"",
			sentAmount,
			common.GetTimestamp())
		if uErr := h.storage.StorePendingIntermediateTx(id, data); uErr != nil {
			h.l.Warnw("Huobi Trying to store intermediate tx to huobi storage. Ignore it and try later", "err", uErr)
			return "", nil
		}
		return h.exchangeDepositStatus(id, tx2Entry, currency, sentAmount)
	case common.MiningStatusFailed:
		data = common.NewTXEntry(
			tx2Entry.Hash,
			h.Name(),
			currency,
			common.MiningStatusFailed,
			common.ExchangeStatusFailed,
			sentAmount,
			common.GetTimestamp(),
		)
		if err = h.storage.StoreIntermediateTx(id, data); err != nil {
			h.l.Warnw("Huobi Trying to store intermediate tx failed. Ignore it and treat it like it is still pending", "err", err)
			return "", nil
		}
		return common.ExchangeStatusFailed, nil
	case common.MiningStatusLost:
		elapsed := common.GetTimepoint() - tx2Entry.Timestamp.MustToUint64()
		if elapsed > uint64(15*time.Minute/time.Millisecond) {
			data = common.NewTXEntry(
				tx2Entry.Hash,
				h.Name(),
				currency,
				common.MiningStatusLost,
				common.ExchangeStatusLost,
				sentAmount,
				common.GetTimestamp(),
			)
			if err = h.storage.StoreIntermediateTx(id, data); err != nil {
				h.l.Infof("Huobi Trying to store intermediate tx failed, error: %+v. Ignore it and treat it like it is still pending", err)
				return "", nil
			}
			h.l.Infof("Huobi The tx is not found for over 15mins, it is considered as lost and the deposit failed")
			return common.ExchangeStatusFailed, nil
		}
		return "", nil
	}
	return "", nil
}

//WithdrawStatus return withdraw status from huobi
func (h *Huobi) WithdrawStatus(
	id, currency string, amount float64, timepoint uint64) (string, string, error) {
	withdrawID, _ := strconv.ParseUint(id, 10, 64)
	tokens, err := h.setting.GetAllTokens()
	if err != nil {
		return "", "", pe.Wrap(err, "huobi Can't get list of token from setting")
	}
	withdraws, err := h.interf.WithdrawHistory(tokens)
	if err != nil {
		return "", "", pe.Wrap(err, "can't get withdraw history from huobi")
	}
	h.l.Infof("Huobi Withdrawal id: %d", withdrawID)
	for _, withdraw := range withdraws.Data {
		if withdraw.ID != withdrawID {
			continue
		}
		switch withdraw.State {
		case "confirmed":
			if withdraw.TxHash[0:2] != "0x" {
				withdraw.TxHash = "0x" + withdraw.TxHash
			}
			return common.ExchangeStatusDone, withdraw.TxHash, nil
		case "reject", "wallet-reject", "confirm-error":
			return common.ExchangeStatusFailed, "", nil
		}
		return "", withdraw.TxHash, nil
	}
	return "", "", errors.New("huobi Withdrawal doesn't exist. This shouldn't happen unless tx returned from withdrawal from huobi and activity ID are not consistently designed")
}

func (h *Huobi) OrderStatus(id string, base, quote string) (string, error) {
	orderID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return "", err
	}
	symbol := base + quote
	order, err := h.interf.OrderStatus(symbol, orderID)
	if err != nil {
		return "", err
	}
	if order.Data.State == "pre-submitted" || order.Data.State == "submitting" || order.Data.State == "submitted" || order.Data.State == "partial-filled" || order.Data.State == "partial-canceled" {
		return "", nil
	}
	return common.ExchangeStatusDone, nil
}

//NewHuobi creates new Huobi exchange instance
func NewHuobi(interf HuobiInterface, blockchain *blockchain.BaseBlockchain, signer blockchain.Signer,
	nonce blockchain.NonceCorpus, storage HuobiStorage, setting Setting) (*Huobi, error) {

	bc, err := huobiblockchain.NewBlockchain(blockchain, signer, nonce)
	if err != nil {
		return nil, err
	}

	huobiObj := Huobi{
		interf:     interf,
		blockchain: bc,
		storage:    storage,
		setting:    setting,
		l:          zap.S(),
	}
	huobiObj.FetchTradeHistory()
	huobiServer := huobihttp.NewHuobiHTTPServer(&huobiObj)
	go huobiServer.Run()
	return &huobiObj, nil
}
