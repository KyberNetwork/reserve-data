package exchange

import (
	"fmt"
	"math/big"
	"strconv"
	"sync"

	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data/common"
	commonv3 "github.com/KyberNetwork/reserve-data/reservesetting/common"
	"github.com/KyberNetwork/reserve-data/reservesetting/storage"
)

const (
	binanceEpsilon float64 = 0.0000001 // 10e-7
	batchSize      int     = 4
)

// Binance instance for binance exchange
type Binance struct {
	interf  BinanceInterface
	storage BinanceStorage
	sr      storage.Interface
	l       *zap.SugaredLogger
	BinanceLive
	id common.ExchangeID
}

// TokenAddresses return deposit addresses of token
func (bn *Binance) TokenAddresses() (map[common.AssetID]ethereum.Address, error) {
	result, err := bn.sr.GetDepositAddresses(uint64(bn.id))
	if err != nil {
		return nil, err
	}
	return result, nil
}

// MarshalText Return exchange id by name
func (bn *Binance) MarshalText() (text []byte, err error) {
	return []byte(bn.ID().String()), nil
}

// Address returns the deposit address of given token.
func (bn *Binance) Address(asset commonv3.Asset) (ethereum.Address, bool) {
	var symbol string
	for _, exchange := range asset.Exchanges {
		if exchange.ExchangeID == uint64(bn.id) {
			symbol = exchange.Symbol
		}
	}
	liveAddress, err := bn.interf.GetDepositAddress(symbol)
	if err != nil || liveAddress.Address == "" {
		bn.l.Warnw("Get Binance live deposit address for token failed or the address replied is empty . Use the currently available address instead", "assetID", asset.ID, "err", err)
		addrs, uErr := bn.sr.GetDepositAddresses(uint64(bn.id))
		if uErr != nil {
			bn.l.Warnw("get address of token in Binance exchange failed, it will be considered as not supported", "assetID", asset.ID, "err", err)
			return ethereum.Address{}, false
		}
		depositAddr, ok := addrs[common.AssetID(asset.ID)]
		return depositAddr, ok && !commonv3.IsZeroAddress(depositAddr)
	}
	bn.l.Infof("Got Binance live deposit address for token %d, attempt to update it to current setting", asset.ID)
	if err = bn.sr.UpdateDepositAddress(
		asset.ID,
		uint64(bn.id),
		ethereum.HexToAddress(liveAddress.Address)); err != nil {
		bn.l.Warnw("failed to update deposit address", "err", err)
		return ethereum.Address{}, false

	}
	return ethereum.HexToAddress(liveAddress.Address), true
}

// ID must return the exact string or else simulation will fail
func (bn *Binance) ID() common.ExchangeID {
	return bn.id
}

// TokenPairs return token pairs supported by exchange
func (bn *Binance) TokenPairs() ([]commonv3.TradingPairSymbols, error) {
	pairs, err := bn.sr.GetTradingPairs(uint64(bn.id))
	if err != nil {
		return nil, err
	}
	return pairs, nil
}

// QueryOrder return current order status
func (bn *Binance) QueryOrder(symbol string, id uint64) (done float64, remaining float64, finished bool, err error) {
	result, err := bn.interf.OrderStatus(symbol, id)
	if err != nil {
		return 0, 0, false, err
	}
	done, _ = strconv.ParseFloat(result.ExecutedQty, 64)
	total, _ := strconv.ParseFloat(result.OrigQty, 64)
	return done, total - done, total-done < binanceEpsilon, nil
}

func (bn *Binance) Trade(tradeType string, pair commonv3.TradingPairSymbols, rate float64, amount float64) (id string, done float64, remaining float64, finished bool, err error) {
	result, err := bn.interf.Trade(tradeType, pair, rate, amount)
	if err != nil {
		return "", 0, 0, false, err
	}
	done, remaining, finished, err = bn.QueryOrder(
		pair.BaseSymbol+pair.QuoteSymbol,
		result.OrderID,
	)
	id = strconv.FormatUint(result.OrderID, 10)
	return id, done, remaining, finished, err
}

func (bn *Binance) Withdraw(asset commonv3.Asset, amount *big.Int, address ethereum.Address) (string, error) {
	tx, err := bn.interf.Withdraw(asset, amount, address)
	return tx, err
}

func (bn *Binance) CancelOrder(id string, base, quote string) error {
	idNo, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return err
	}
	symbol := base + quote
	_, err = bn.interf.CancelOrder(symbol, idNo)
	if err != nil {
		return err
	}
	return nil
}

func (bn *Binance) FetchOnePairData(
	wg *sync.WaitGroup,
	pair commonv3.TradingPairSymbols,
	data *sync.Map,
	timepoint uint64) {

	defer wg.Done()
	result := common.ExchangePrice{}

	timestamp := common.Timestamp(fmt.Sprintf("%d", timepoint))
	result.Timestamp = timestamp
	result.Valid = true
	respData, err := bn.interf.GetDepthOnePair(pair.BaseSymbol, pair.QuoteSymbol)
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
	data.Store(pair.ID, result)
}

func (bn *Binance) FetchPriceData(timepoint uint64) (map[uint64]common.ExchangePrice, error) {
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
			go bn.FetchOnePairData(&wait, pair, &data, timepoint)
		}
		wait.Wait()
		i = x
	}
	result := map[uint64]common.ExchangePrice{}
	data.Range(func(key, value interface{}) bool {
		//if there is conversion error, continue to next key,val
		tokenPairID, ok := key.(uint64)
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
		result.AvailableBalance = map[common.AssetID]float64{}
		result.LockedBalance = map[common.AssetID]float64{}
		result.DepositBalance = map[common.AssetID]float64{}
		result.Status = true
		if respData.Code != 0 {
			result.Valid = false
			result.Error = fmt.Sprintf("Code: %d, Msg: %s", respData.Code, respData.Msg)
			result.Status = false
		} else {
			assets, err := bn.sr.GetAssets()
			if err != nil {
				return common.EBalanceEntry{}, err
			}
			for _, b := range respData.Balances {
				tokenSymbol := b.Asset
				for _, asset := range assets {
					for _, exchg := range asset.Exchanges {
						if exchg.ExchangeID == uint64(bn.id) && exchg.Symbol == tokenSymbol {
							avai, _ := strconv.ParseFloat(b.Free, 64)
							locked, _ := strconv.ParseFloat(b.Locked, 64)
							result.AvailableBalance[common.AssetID(asset.ID)] = avai
							result.LockedBalance[common.AssetID(asset.ID)] = locked
							result.DepositBalance[common.AssetID(asset.ID)] = 0
						}
					}
				}
			}
		}
	}
	return result, nil
}

//FetchOnePairTradeHistory fetch trade history for one pair from exchange
func (bn *Binance) FetchOnePairTradeHistory(pair commonv3.TradingPairSymbols) ([]common.TradeHistory, error) {
	var result []common.TradeHistory
	fromID, err := bn.storage.GetLastIDTradeHistory(pair.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "Cannot get last ID trade history")
	}
	resp, err := bn.interf.GetAccountTradeHistory(pair.BaseSymbol, pair.QuoteSymbol, fromID)
	if err != nil {
		return nil, errors.Wrapf(err, "Binance Cannot fetch data for pair %s%s", pair.BaseSymbol, pair.QuoteSymbol)
	}
	for _, trade := range resp {
		price, err := strconv.ParseFloat(trade.Price, 64)
		if err != nil {
			return nil, errors.Wrapf(err, "Can not parse price: %v", price)
		}
		quantity, err := strconv.ParseFloat(trade.Qty, 64)
		if err != nil {
			return nil, errors.Wrapf(err, "Can not parse quantity: %v", trade.Qty)
		}
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
	return result, nil
}

//FetchTradeHistory get all trade history for all tokens in the exchange
func (bn *Binance) FetchTradeHistory() {
	pairs, err := bn.TokenPairs()
	if err != nil {
		bn.l.Warnw("Binance Get Token pairs setting failed", "err", err)
		return
	}
	var (
		result        = common.ExchangeTradeHistory{}
		guard         = &sync.Mutex{}
		wait          = &sync.WaitGroup{}
		batchStart, x int
	)

	for batchStart < len(pairs) {
		for x = batchStart; x < len(pairs) && x < batchStart+batchSize; x++ {
			wait.Add(1)
			go func(pair commonv3.TradingPairSymbols) {
				defer wait.Done()
				histories, err := bn.FetchOnePairTradeHistory(pair)
				if err != nil {
					bn.l.Warnw("Cannot fetch data for pair",
						"pair", fmt.Sprintf("%s%s", pair.BaseSymbol, pair.QuoteSymbol), "err", err)
					return
				}
				guard.Lock()
				result[pair.ID] = histories
				guard.Unlock()
			}(pairs[x])
		}
		batchStart = x
		wait.Wait()
	}

	if err := bn.storage.StoreTradeHistory(result); err != nil {
		bn.l.Warnw("Binance Store trade history error", "err", err)
	}
}

func (bn *Binance) GetTradeHistory(fromTime, toTime uint64) (common.ExchangeTradeHistory, error) {
	return bn.storage.GetTradeHistory(fromTime, toTime)
}

func (bn *Binance) DepositStatus(id common.ActivityID, txHash string, assetID uint64, amount float64, timepoint uint64) (string, error) {
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
	bn.l.Warnw("Binance Deposit is not found in deposit list returned from Binance. " +
		"This might cause by wrong start/end time, please check again.")
	return "", nil
}

func (bn *Binance) WithdrawStatus(id string, assetID uint64, amount float64, timepoint uint64) (string, string, error) {
	startTime := timepoint - 86400000
	endTime := timepoint
	withdraws, err := bn.interf.WithdrawHistory(startTime, endTime)
	if err != nil || !withdraws.Success {
		return "", "", err
	}
	for _, withdraw := range withdraws.Withdrawals {
		if withdraw.ID == id {
			if withdraw.Status == 3 || withdraw.Status == 5 || withdraw.Status == 6 {
				return common.ExchangeStatusDone, withdraw.TxID, nil
			}
			return "", withdraw.TxID, nil
		}
	}
	bn.l.Warnw("Binance Withdrawal doesn't exist. This shouldn't happen unless tx returned from withdrawal from binance and activity ID are not consistently designed",
		"id", id, "asset_id", assetID, "amount", amount, "timepoint", timepoint)
	return "", "", nil
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

// OpenOrders get open orders from binance
func (bn *Binance) OpenOrders(pair commonv3.TradingPairSymbols) ([]common.Order, error) {
	var (
		result []common.Order
		orders []Binaorder
		err    error
	)
	if pair.ID != 0 {
		orders, err = bn.interf.OpenOrdersForOnePair(&pair)
		if err != nil {
			return nil, err
		}
	} else {
		orders, err = bn.interf.OpenOrdersForOnePair(nil)
		if err != nil {
			return nil, err
		}
	}
	for _, order := range orders {
		originalQty, err := strconv.ParseFloat(order.OrigQty, 64)
		if err != nil {
			return nil, err
		}
		price, err := strconv.ParseFloat(order.Price, 64)
		if err != nil {
			return nil, err
		}
		result = append(result, common.Order{
			OrderID: strconv.FormatUint(order.OrderID, 10),
			Side:    order.Side,
			Type:    order.Type,
			OrigQty: originalQty,
			Price:   price,
			Base:    pair.BaseSymbol,
			Quote:   pair.QuoteSymbol,
		})
	}
	return result, nil
}

// NewBinance init new binance instance
func NewBinance(id common.ExchangeID, interf BinanceInterface, storage BinanceStorage, sr storage.Interface) (*Binance, error) {
	binance := &Binance{
		interf:  interf,
		storage: storage,
		sr:      sr,
		BinanceLive: BinanceLive{
			interf: interf,
		},
		id: id,
		l:  zap.S(),
	}
	return binance, nil
}
