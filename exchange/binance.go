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
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/bcnetwork"
	"github.com/KyberNetwork/reserve-data/common/blockchain"
	"github.com/KyberNetwork/reserve-data/common/gasinfo"
	binblockchain "github.com/KyberNetwork/reserve-data/exchange/blockchain"
	"github.com/KyberNetwork/reserve-data/lib/caller"
	"github.com/KyberNetwork/reserve-data/lib/rtypes"
	commonv3 "github.com/KyberNetwork/reserve-data/reservesetting/common"
	"github.com/KyberNetwork/reserve-data/reservesetting/storage"
)

const (
	binanceEpsilon float64 = 0.0000001 // 10e-7
	BSCNetwork             = "BSC"
)

// Binance instance for binance exchange
type Binance struct {
	interf  BinanceInterface
	storage BinanceStorage
	sr      storage.Interface
	l       *zap.SugaredLogger
	BinanceLive
	id            rtypes.ExchangeID
	binBlockchain *binblockchain.Blockchain
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
	network := bcnetwork.GetPreConfig().Network
	// Special case for BNB on bsc
	if network == BSCNetwork && asset.IsNetworkAsset() {
		return bn.binBlockchain.GetIntermediatorAddr(blockchain.BinanceOP), false
	}
	return bn.GetLiveDepositAddress(asset, symbol, network)
}

func (bn *Binance) GetLiveDepositAddress(asset commonv3.Asset, symbol, network string) (ethereum.Address, bool) {
	liveAddress, err := bn.interf.GetDepositAddress(symbol, network)
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
	bn.l.Infow("attempt to update binance deposit address to current setting",
		"asset", asset.ID, "deposit_address", liveAddress.Address)
	if err = bn.sr.UpdateDepositAddress(asset.ID, bn.id, ethereum.HexToAddress(liveAddress.Address)); err != nil {
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
		return "", 0, 0, true, err
	}
	id = strconv.FormatUint(result.OrderID, 10)
	return id, 0, amount, false, err
}

// Withdraw create a withdrawal from binance to our reserve
func (bn *Binance) Withdraw(asset commonv3.Asset, amount *big.Int, address ethereum.Address) (string, error) {
	tx, err := bn.interf.Withdraw(asset, amount, address)
	return tx, err
}

func (bn *Binance) Transfer(fromAccount string, toAccount string, asset commonv3.Asset, amount *big.Int, runAsync bool, referenceID string) (string, error) {
	return bn.interf.Transfer(fromAccount, toAccount, asset, amount, runAsync, referenceID)
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
func (bn *Binance) FetchOnePairData(pair commonv3.TradingPairSymbols, timepoint uint64) common.ExchangePrice {
	result := common.ExchangePrice{}

	result.Valid = true
	respData, err := bn.interf.GetDepthOnePair(pair.BaseSymbol, pair.QuoteSymbol)
	if err != nil {
		result.Valid = false
		result.Error = err.Error()
		return result
	}
	if respData.Code != 0 || respData.Msg != "" {
		result.Valid = false
		result.Error = fmt.Sprintf("Code: %d, Msg: %s", respData.Code, respData.Msg)
		return result
	}
	result.Timestamp = common.TimestampFromMillis(respData.Timestamp)
	for _, buy := range respData.Bids {
		result.Bids = append(result.Bids, common.NewPriceEntry(buy.Quantity, buy.Rate))
	}
	for _, sell := range respData.Asks {
		result.Asks = append(result.Asks, common.NewPriceEntry(sell.Quantity, sell.Rate))
	}
	return result
}

// FetchPriceData fetch price data for all token that we supported
func (bn *Binance) FetchPriceData(timepoint uint64) (map[rtypes.TradingPairID]common.ExchangePrice, error) {
	wait := sync.WaitGroup{}
	result := map[rtypes.TradingPairID]common.ExchangePrice{}
	var lock sync.Mutex
	pairs, err := bn.TokenPairs()
	if err != nil {
		return nil, err
	}
	jobs := make(chan commonv3.TradingPairSymbols, len(pairs))
	for _, p := range pairs {
		jobs <- p
	}
	close(jobs)
	const concurrentFactor = 8
	wait.Add(concurrentFactor)
	for i := 0; i < concurrentFactor; i++ {
		go func() {
			defer wait.Done()
			for pair := range jobs {
				res := bn.FetchOnePairData(pair, timepoint)
				lock.Lock()
				result[pair.ID] = res
				lock.Unlock()
			}
		}()
	}
	wait.Wait()
	return result, err
}

// FetchEBalanceData fetch balance data from binance
func (bn *Binance) FetchEBalanceData(timepoint uint64) (common.EBalanceEntry, error) {
	var (
		logger = bn.l.With("func", caller.GetCurrentFunctionName())
	)
	result := common.EBalanceEntry{}
	result.Timestamp = common.TimestampFromMillis(timepoint)
	result.Valid = true
	result.Error = ""
	respData, err := bn.interf.GetInfo()
	marginRespData, errM := bn.interf.GetMarginAccountInfo()
	result.ReturnTime = common.GetTimestamp()
	if err != nil {
		result.Valid = false
		result.Error = err.Error()
		result.Status = false
		return result, nil
	}
	if errM != nil {
		result.Valid = false
		result.Error = errM.Error()
		result.Status = false
		return result, nil
	}
	result.MarginBalance = map[rtypes.AssetID]common.AssetMarginBalance{}
	result.AvailableBalance = map[rtypes.AssetID]float64{}
	result.LockedBalance = map[rtypes.AssetID]float64{}
	result.DepositBalance = map[rtypes.AssetID]float64{}
	result.Status = true
	assets, err := bn.sr.GetAssets()
	if err != nil {
		logger.Errorw("failed to get asset from storage", "error", err)
		return common.EBalanceEntry{}, err
	}
	for _, ua := range marginRespData.UserAssets {
		tokenSymbol := ua.Asset
		for _, asset := range assets {
			for _, exchg := range asset.Exchanges {
				if exchg.ExchangeID == bn.id && exchg.Symbol == tokenSymbol {
					amb, err := ua.ToAssetMarginBalance()
					if err != nil {
						logger.Errorw("failed to convert raw data to asset margin balance", "error", err)
						return common.EBalanceEntry{}, err
					}
					result.MarginBalance[asset.ID] = amb
				}
			}
		}
	}
	if respData.Code != 0 {
		result.Valid = false
		result.Error = fmt.Sprintf("Code: %d, Msg: %s", respData.Code, respData.Msg)
		result.Status = false
	} else {
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
		result = common.ExchangeTradeHistory{}
		lock   = &sync.Mutex{}
		wg     sync.WaitGroup
	)
	jobs := make(chan commonv3.TradingPairSymbols, len(pairs))
	for _, p := range pairs {
		jobs <- p
	}
	close(jobs)
	const concurrentFactor = 4
	wg.Add(concurrentFactor)
	for i := 0; i < concurrentFactor; i++ {
		go func() {
			defer wg.Done()
			for pair := range jobs { // process until jobs empty
				histories, err := bn.FetchOnePairTradeHistory(pair)
				if err != nil {
					bn.l.Warnw("Cannot fetch data for pair",
						"pair", fmt.Sprintf("%s%s", pair.BaseSymbol, pair.QuoteSymbol), "err", err)
					return
				}
				lock.Lock()
				result[pair.ID] = histories
				lock.Unlock()
			}
		}()
	}
	wg.Wait()

	if err := bn.storage.StoreTradeHistory(result); err != nil {
		bn.l.Warnw("Binance Store trade history error", "err", err)
	}
}

// GetTradeHistory get trade history from binance
func (bn *Binance) GetTradeHistory(fromTime, toTime uint64) (common.ExchangeTradeHistory, error) {
	return bn.storage.GetTradeHistory(bn.ID(), fromTime, toTime)
}

func (bn *Binance) SendBNBToExchange(bc *blockchain.BaseBlockchain, amount *big.Int, exchangeAddress ethereum.Address, gasPrice *big.Int) (*types.Transaction, error) {
	opts, err := bc.GetTxOpts(blockchain.BinanceOP, nil, gasPrice, amount)
	if err != nil {
		return nil, err
	}
	tx, err := bc.BuildSendETHTx(opts, exchangeAddress)
	if err != nil {
		return nil, err
	}
	return bc.SignAndBroadcast(tx, blockchain.BinanceOP)
}

// Send2ndTransaction send the second transaction
func (bn *Binance) Send2ndTransaction(amount float64, asset commonv3.Asset, exchangeAddress ethereum.Address) (*types.Transaction, error) {
	IAmount := common.FloatToBigInt(amount, int64(asset.Decimals))
	var tx *types.Transaction
	gasInfo := gasinfo.GetGlobal()
	if gasInfo == nil {
		bn.l.Errorw("gasInfo not setup, retry later")
		return nil, fmt.Errorf("gasInfo not setup, retry later")
	}
	recommendedPrice, err := gasInfo.GetCurrentGas()
	if err != nil {
		bn.l.Errorw("failed to get gas price", "err", err)
		return nil, fmt.Errorf("can not get gas, retry later")
	}
	var gasPrice *big.Int
	highBoundGasPrice, err := gasInfo.MaxGas()
	if err != nil {
		bn.l.Errorw("failed to receive high bound gas, use default", "err", err)
		highBoundGasPrice = common.HighBoundGasPrice
	}
	if recommendedPrice == 0 || recommendedPrice > highBoundGasPrice {
		bn.l.Errorw("gas price invalid", "gas", recommendedPrice, "highBound", highBoundGasPrice)
		return nil, fmt.Errorf("gas price invalid (gas=%v, highBound=%v), retry later", recommendedPrice, highBoundGasPrice)
	}
	gasPrice = common.GweiToWei(recommendedPrice)
	bn.l.Infof("Send2ndTransaction, gas price: %s", gasPrice.String())
	tx, err = bn.binBlockchain.SendETHFromAccountToExchange(IAmount, exchangeAddress, gasPrice, blockchain.BinanceOP) // we only use this we to deposit BNB on BSC
	if err != nil {
		bn.l.Warnw("ERROR: Can not send transaction to exchange", "err", err)
		return nil, err
	}
	bn.l.Infof("Transaction submitted. Tx is: %v", tx)
	return tx, nil

}

func (bn *Binance) checkAndsend2ndTx(id common.ActivityID, asset commonv3.Asset, tx1Hash, network string, amount float64) error {
	miningStatus, _, err := bn.binBlockchain.TxStatus(ethereum.HexToHash(tx1Hash))
	if err != nil {
		return err
	}
	var (
		data common.TXEntry
	)
	if miningStatus == common.MiningStatusMined {
		// as auth data will call DepositStatus more than 1 time, store pending tx2 can be call multiple times,
		// so we handle it as on conflict update ...
		// if uErr := bn.storage.StoreIntermediateDeposit(id, data, false); uErr != nil {
		// 	return common.ExchangeStatusNA, nil
		// }
		depositAddress, ok := bn.GetLiveDepositAddress(asset, asset.Symbol, network)
		if !ok {
			bn.l.Errorw("cannot get live deposit address", "asset", asset.ID, "symbol", asset.Symbol)
			return fmt.Errorf("cannot get deposit address for %s", asset.Symbol)
		}
		tx2, err := bn.Send2ndTransaction(amount, asset, depositAddress)
		if err != nil {
			bn.l.Errorw("binance deposit failed to send 2nd transaction", "error", err)
			return err
		}
		data = common.NewTXEntry(
			tx2.Hash().Hex(),
			bn.ID().String(),
			asset.ID,
			common.MiningStatusFailed,
			common.ExchangeStatusFailed,
			amount,
			common.GetTimestamp(),
		)
		// update tx2 to activity, set blockchain status to pending
		return bn.storage.StoreIntermediateDeposit(id, data)
	}
	return nil
}

// DepositStatus return status of a deposit on binance
func (bn *Binance) DepositStatus(id common.ActivityID, txHash string, assetID rtypes.AssetID, amount float64, timepoint uint64) (string, error) {
	network := bcnetwork.GetPreConfig().Network
	asset, err := bn.sr.GetAsset(assetID)
	if err != nil {
		return common.ExchangeStatusNA, err
	}
	if isOPAvailable(bn.binBlockchain, blockchain.BinanceOP) && network == BSCNetwork && asset.IsNetworkAsset() {
		// this only apply for deposit BNB on BSC, when intermediate operator configured
		// deposit usually has trouble with NetworkAsset(BNB/ETH) sent from contract, as it did not emit event
		// that cex can track.
		// find second tx
		tx2Entry, err := bn.storage.GetPendingIntermediateTx(id)
		// if there is no second tx then process the first 1
		if err == sql.ErrNoRows {
			err := bn.checkAndsend2ndTx(id, asset, txHash, network, amount)
			return common.ExchangeStatusNA, err
		}

		if err != nil {
			return "", err
		}

		// if there is tx2Entry, check it blockchain status and handle the status accordingly:
		txHash = tx2Entry.Hash
		miningStatus, _, err := bn.binBlockchain.TxStatus(ethereum.HexToHash(txHash))
		if err != nil {
			return common.ExchangeStatusNA, err
		}
		switch miningStatus {
		case common.MiningStatusMined:
			data := common.NewTXEntry(
				txHash,
				bn.ID().String(),
				assetID,
				common.MiningStatusMined,
				common.ExchangeStatusNA,
				amount,
				common.GetTimestamp())
			if uErr := bn.storage.StoreIntermediateDeposit(id, data); uErr != nil { // update tx2 status
				return common.ExchangeStatusNA, nil
			} // 2nd tx mined, check exchange status as normal
		case common.MiningStatusFailed:
			data := common.NewTXEntry(
				txHash,
				bn.ID().String(),
				assetID,
				common.MiningStatusFailed,
				common.ExchangeStatusFailed,
				amount,
				common.GetTimestamp(),
			)
			if err = bn.storage.StoreIntermediateDeposit(id, data); err != nil {
				return common.ExchangeStatusNA, nil
			}
			return common.ExchangeStatusFailed, nil
		case common.MiningStatusLost:
			elapsed := common.NowInMillis() - tx2Entry.Timestamp.Millis()
			if elapsed > uint64(15*time.Minute/time.Millisecond) {
				data := common.NewTXEntry(
					txHash,
					bn.ID().String(),
					assetID,
					common.MiningStatusLost,
					common.ExchangeStatusLost,
					amount,
					common.GetTimestamp(),
				)
				if err = bn.storage.StoreIntermediateDeposit(id, data); err != nil {
					return common.ExchangeStatusNA, nil
				}
				bn.l.Infof("Binance deposit tx is not found for over 15mins, it is considered as lost and the deposit failed")
				return common.ExchangeStatusFailed, nil
			}
			return common.ExchangeStatusNA, nil
		}
	}

	// checking as normal
	startTime := timepoint - 86400000
	endTime := timepoint
	deposits, err := bn.interf.DepositHistory(startTime, endTime)
	if err != nil || !deposits.Success {
		return common.ExchangeStatusNA, err
	}
	for _, deposit := range deposits.Deposits {
		if deposit.TxID == txHash {
			if deposit.Status == 1 {
				return common.ExchangeStatusDone, nil
			}
			bn.l.Debugw("got binance deposit status", "status", deposit.Status,
				"deposit_tx", deposit.TxID, "tx_hash", txHash)
			return common.ExchangeStatusNA, nil
		}
	}
	return common.ExchangeStatusNA, nil
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
		return common.ExchangeStatusNA, "", 0, err
	}
	for _, withdraw := range withdraws.Withdrawals {
		if withdraw.WithdrawOrderID == id {
			switch withdraw.Status {
			case binanceRejected, binanceFailure: // 3 = rejected, 5 = failed
				return common.ExchangeStatusFailed, withdraw.TxID, withdraw.Fee, nil
			case binanceCompleted: // 6 = success
				return common.ExchangeStatusDone, withdraw.TxID, withdraw.Fee, nil
			case binanceCancelled: // 1 = cancelled
				return common.ExchangeStatusCancelled, withdraw.TxID, withdraw.Fee, nil
			case binanceAwaitingApproval, binanceProcessing: // no action, just leave it as pending
				return common.ExchangeStatusPending, withdraw.TxID, withdraw.Fee, nil
			default:
				bn.l.Errorw("got unexpected withdraw status", "status", withdraw.Status,
					"withdrawID", id, "assetID", assetID, "amount", amount)
				return common.ExchangeStatusNA, withdraw.TxID, withdraw.Fee, nil
			}
		}
	}
	bn.l.Warnw("Binance Withdrawal doesn't exist. This shouldn't happen unless tx returned from withdrawal from binance and activity ID are not consistently designed",
		"id", id, "asset_id", assetID, "amount", amount, "timepoint", timepoint)
	return common.ExchangeStatusNA, "", 0, nil
}

// OrderStatus return status of an order on binance
func (bn *Binance) OrderStatus(id, base, quote string) (string, float64, error) {
	orderID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return common.ExchangeStatusNA, 0, fmt.Errorf("can not parse orderID (val %s) to uint", id)
	}
	symbol := base + quote
	order, err := bn.interf.OrderStatus(symbol, orderID)
	if err != nil {
		return common.ExchangeStatusNA, 0, err
	}
	qtyLeft, err := remainingQty(order.OrigQty, order.ExecutedQty)
	if err != nil {
		bn.l.Errorw("failed to parse amount", "err", err, "order", order)
	}
	switch order.Status {
	case "CANCELED":
		return common.ExchangeStatusCancelled, qtyLeft, nil
	case "NEW", "PARTIALLY_FILLED", "PENDING_CANCEL":
		return common.ExchangeStatusPending, qtyLeft, nil
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
func NewBinance(id rtypes.ExchangeID, interf BinanceInterface,
	storage BinanceStorage, sr storage.Interface, bc *binblockchain.Blockchain) (*Binance, error) {
	binance := &Binance{
		interf:  interf,
		storage: storage,
		sr:      sr,
		BinanceLive: BinanceLive{
			interf: interf,
		},
		id:            id,
		l:             zap.S(),
		binBlockchain: bc,
	}
	return binance, nil
}
