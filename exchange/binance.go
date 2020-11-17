package exchange

import (
	"database/sql"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"

	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/lib/caller"
	"github.com/KyberNetwork/reserve-data/lib/rtypes"
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
	id rtypes.ExchangeID
}

// TokenAddresses return deposit addresses of token
func (bn *Binance) TokenAddresses() (map[rtypes.AssetID]ethereum.Address, error) {
	result, err := bn.sr.GetDepositAddresses(bn.id)
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
		if exchange.ExchangeID == bn.id {
			symbol = exchange.Symbol
		}
	}
	liveAddress, err := bn.interf.GetDepositAddress(symbol)
	if err != nil || liveAddress.Address == "" {
		bn.l.Warnw("Get Binance live deposit address for token failed or the address replied is empty . Use the currently available address instead", "assetID", asset.ID, "err", err)
		addrs, uErr := bn.sr.GetDepositAddresses(bn.id)
		if uErr != nil {
			bn.l.Warnw("get address of token in Binance exchange failed, it will be considered as not supported", "assetID", asset.ID, "err", err)
			return ethereum.Address{}, false
		}
		depositAddr, ok := addrs[asset.ID]
		return depositAddr, ok && !commonv3.IsZeroAddress(depositAddr)
	}
	bn.l.Infof("Got Binance live deposit address for token %d, attempt to update it to current setting", asset.ID)
	if err = bn.sr.UpdateDepositAddress(
		asset.ID,
		bn.id,
		ethereum.HexToAddress(liveAddress.Address)); err != nil {
		bn.l.Warnw("failed to update deposit address", "err", err)
		return ethereum.Address{}, false

	}
	return ethereum.HexToAddress(liveAddress.Address), true
}

// ID must return the exact string or else simulation will fail
func (bn *Binance) ID() rtypes.ExchangeID {
	return bn.id
}

// TokenPairs return token pairs supported by exchange
func (bn *Binance) TokenPairs() ([]commonv3.TradingPairSymbols, error) {
	pairs, err := bn.sr.GetTradingPairs(bn.id)
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

// Trade create a new trade on binance
func (bn *Binance) Trade(tradeType string, pair commonv3.TradingPairSymbols, rate float64, amount float64) (id string, done float64, remaining float64, finished bool, err error) {
	result, err := bn.interf.Trade(tradeType, pair, rate, amount)
	if err != nil {
		return "", 0, 0, false, err
	}
	for i := 0; i < 5; i++ { // sometime binance get trouble when query order info right after it created, so we
		// add a retry to handle it here
		done, remaining, finished, err = bn.QueryOrder(
			pair.BaseSymbol+pair.QuoteSymbol,
			result.OrderID,
		)
		if err == nil {
			break
		}
		bn.l.Errorw("failed to query order info", "err", err, "i", i, "orderID",
			result.OrderID, "base", pair.BaseSymbol, "quote", pair.QuoteSymbol)
		if strings.Contains(err.Error(), "Order does not exist") { // only retry if got specified error
			time.Sleep(time.Second)
			continue
		}
		break
	}
	id = strconv.FormatUint(result.OrderID, 10)
	return id, done, remaining, finished, err
}

// Withdraw create a withdrawal from binance to our reserve
func (bn *Binance) Withdraw(asset commonv3.Asset, amount *big.Int, address ethereum.Address) (string, error) {
	tx, err := bn.interf.Withdraw(asset, amount, address)
	return tx, err
}

func (bn *Binance) Transfer(fromAccount string, toAccount string, asset commonv3.Asset, amount *big.Int) (string, error) {
	return bn.interf.Transfer(fromAccount, toAccount, asset, amount)
}

// CancelOrder cancel order on binance
func (bn *Binance) CancelOrder(id, symbol string) error {
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

// CancelAllOrders cancel all open orders of a symbol
func (bn *Binance) CancelAllOrders(symbol string) error {
	_, err := bn.interf.CancelAllOrders(symbol)
	return err
}

// FetchOnePairData fetch price data for one pair of token
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
				result.Bids = append(
					result.Bids,
					common.NewPriceEntry(
						buy.Quantity,
						buy.Rate,
					),
				)
			}
			for _, sell := range respData.Asks {
				result.Asks = append(
					result.Asks,
					common.NewPriceEntry(
						sell.Quantity,
						sell.Rate,
					),
				)
			}
		}
	}
	data.Store(pair.ID, result)
}

// FetchPriceData fetch price data for all token that we supported
func (bn *Binance) FetchPriceData(timepoint uint64) (map[rtypes.TradingPairID]common.ExchangePrice, error) {
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
	result := map[rtypes.TradingPairID]common.ExchangePrice{}
	data.Range(func(key, value interface{}) bool {
		//if there is conversion error, continue to next key,val
		tokenPairID, ok := key.(rtypes.TradingPairID)
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

// FetchEBalanceData fetch balance data from binance
func (bn *Binance) FetchEBalanceData(timepoint uint64) (common.EBalanceEntry, error) {
	var (
		logger = bn.l.With("func", caller.GetCurrentFunctionName())
	)
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
		result.AvailableBalance = map[rtypes.AssetID]float64{}
		result.LockedBalance = map[rtypes.AssetID]float64{}
		result.DepositBalance = map[rtypes.AssetID]float64{}
		result.Status = true
		if respData.Code != 0 {
			result.Valid = false
			result.Error = fmt.Sprintf("Code: %d, Msg: %s", respData.Code, respData.Msg)
			result.Status = false
		} else {
			assets, err := bn.sr.GetAssets()
			if err != nil {
				logger.Errorw("failed to get asset from storage", "error", err)
				return common.EBalanceEntry{}, err
			}
			for _, b := range respData.Balances {
				tokenSymbol := b.Asset
				for _, asset := range assets {
					for _, exchg := range asset.Exchanges {
						if exchg.ExchangeID == bn.id && exchg.Symbol == tokenSymbol {
							avai, _ := strconv.ParseFloat(b.Free, 64)
							locked, _ := strconv.ParseFloat(b.Locked, 64)
							result.AvailableBalance[asset.ID] = avai
							result.LockedBalance[asset.ID] = locked
							result.DepositBalance[asset.ID] = 0
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
	if err != nil && err != sql.ErrNoRows {
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

// GetTradeHistory get trade history from binance
func (bn *Binance) GetTradeHistory(fromTime, toTime uint64) (common.ExchangeTradeHistory, error) {
	return bn.storage.GetTradeHistory(bn.ID(), fromTime, toTime)
}

// DepositStatus return status of a deposit on binance
func (bn *Binance) DepositStatus(id common.ActivityID, txHash string, assetID rtypes.AssetID, amount float64, timepoint uint64) (string, error) {
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

const (
	binanceCancelled        = 1
	binanceAwaitingApproval = 2
	binanceRejected         = 3
	binanceProcessing       = 4
	binanceFailure          = 5
	binanceCompleted        = 6
)

// WithdrawStatus return status of a withdrawal on binance
// return
// 		 string withdraw_status
// 		 string withraw_tx
// 		 float64 withdraw_fee
//		 error
func (bn *Binance) WithdrawStatus(id string, assetID rtypes.AssetID, amount float64, timepoint uint64) (string, string, float64, error) {
	startTime := timepoint - 86400000
	endTime := timepoint
	withdraws, err := bn.interf.WithdrawHistory(startTime, endTime)
	if err != nil || !withdraws.Success {
		return "", "", 0, err
	}
	for _, withdraw := range withdraws.Withdrawals {
		if withdraw.ID == id {
			switch withdraw.Status {
			case binanceRejected, binanceFailure: // 3 = rejected, 5 = failed
				return common.ExchangeStatusFailed, withdraw.TxID, withdraw.Fee, nil
			case binanceCompleted: // 6 = success
				return common.ExchangeStatusDone, withdraw.TxID, withdraw.Fee, nil
			case binanceCancelled: // 1 = cancelled
				return common.ExchangeStatusCancelled, withdraw.TxID, withdraw.Fee, nil
			case binanceAwaitingApproval, binanceProcessing: // no action, just leave it as pending
			}
			return "", withdraw.TxID, withdraw.Fee, nil
		}
	}
	bn.l.Warnw("Binance Withdrawal doesn't exist. This shouldn't happen unless tx returned from withdrawal from binance and activity ID are not consistently designed",
		"id", id, "asset_id", assetID, "amount", amount, "timepoint", timepoint)
	return "", "", 0, nil
}

// OrderStatus return status of an order on binance
func (bn *Binance) OrderStatus(id, base, quote string) (string, float64, error) {
	orderID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("can not parse orderID (val %s) to uint", id)
	}
	symbol := base + quote
	order, err := bn.interf.OrderStatus(symbol, orderID)
	if err != nil {
		return "", 0, err
	}
	qtyLeft, err := remainingQty(order.OrigQty, order.ExecutedQty)
	if err != nil {
		bn.l.Errorw("failed to parse amount", "err", err, "order", order)
	}
	switch order.Status {
	case "CANCELED":
		return common.ExchangeStatusCancelled, qtyLeft, nil
	case "NEW", "PARTIALLY_FILLED", "PENDING_CANCEL":
		return "", qtyLeft, nil
	default:
		return common.ExchangeStatusDone, qtyLeft, nil
	}
}

// OpenOrders get open orders from binance
func (bn *Binance) OpenOrders(pair commonv3.TradingPairSymbols) ([]common.Order, error) {
	var (
		result       = []common.Order{}
		orders       []Binaorder
		err          error
		tradingPairs []commonv3.TradingPairSymbols
	)
	if pair.ID != 0 {
		orders, err = bn.interf.OpenOrdersForOnePair(&pair)
		if err != nil {
			return nil, err
		}
	} else { // pair.ID == 0
		orders, err = bn.interf.OpenOrdersForOnePair(nil)
		if err != nil {
			return nil, err
		}
		tradingPairs, err = bn.TokenPairs()
		if err != nil {
			return nil, err
		}
	}
	pairBK := pair
	for _, order := range orders {
		originalQty, err := strconv.ParseFloat(order.OrigQty, 64)
		if err != nil {
			return nil, err
		}
		price, err := strconv.ParseFloat(order.Price, 64)
		if err != nil {
			return nil, err
		}
		if pairBK.ID == 0 { // pair is not provided
			for _, opair := range tradingPairs {
				if strings.EqualFold(opair.BaseSymbol+opair.QuoteSymbol, order.Symbol) {
					pair = opair
					break
				}
			}
		}

		result = append(result, common.Order{
			OrderID:       strconv.FormatUint(order.OrderID, 10),
			Side:          order.Side,
			Type:          order.Type,
			OrigQty:       originalQty,
			Price:         price,
			Symbol:        order.Symbol,
			Quote:         pair.QuoteSymbol,
			Base:          pair.BaseSymbol,
			Time:          order.Time,
			TradingPairID: pair.ID,
		})
	}
	return result, nil
}

// NewBinance init new binance instance
func NewBinance(id rtypes.ExchangeID, interf BinanceInterface, storage BinanceStorage, sr storage.Interface) (*Binance, error) {
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
