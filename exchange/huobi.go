package exchange

import (
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
	"golang.org/x/sync/errgroup"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/blockchain"
	huobiblockchain "github.com/KyberNetwork/reserve-data/exchange/huobi/blockchain"
	huobihttp "github.com/KyberNetwork/reserve-data/exchange/huobi/http"
	"github.com/KyberNetwork/reserve-data/lib/caller"
	commonv3 "github.com/KyberNetwork/reserve-data/reservesetting/common"
	"github.com/KyberNetwork/reserve-data/reservesetting/storage"
)

const (
	huobiEpsilon float64 = 0.0000000001 // 10e-10
)

// Huobi is instance for Huobi exchange
type Huobi struct {
	interf     HuobiInterface
	blockchain HuobiBlockchain
	storage    HuobiStorage
	sr         storage.SettingReader
	l          *zap.SugaredLogger
	HuobiLive
}

// TokenAddresses return deposit of all token supported by Huobi
func (h *Huobi) TokenAddresses() (map[common.AssetID]ethereum.Address, error) {
	result, err := h.sr.GetDepositAddresses(uint64(common.Huobi))
	if err != nil {
		return nil, err
	}
	return result, nil
}

// MarshalText marshal Huobi exchange name
func (h *Huobi) MarshalText() (text []byte, err error) {
	return []byte(h.ID().String()), nil
}

// RealDepositAddress return the actual Huobi deposit address of a token
// It should only be used to send 2nd transaction.
func (h *Huobi) RealDepositAddress(tokenID string, asset commonv3.Asset) (ethereum.Address, error) {
	liveAddress, err := h.interf.GetDepositAddress(tokenID)
	if err != nil || len(liveAddress.Data) == 0 || liveAddress.Data[0].Address == "" {
		if err != nil {
			h.l.Warnw("Get Huobi live deposit address for token failed. Check the currently available address instead", "tokenID", tokenID, "err", err)
		} else {
			h.l.Warnw("Get Huobi live deposit address for token failed: the replied address is empty. Check the currently available address instead", "tokenID", tokenID)
		}
		addrs, uErr := h.sr.GetDepositAddresses(uint64(common.Huobi))
		if uErr != nil {
			return ethereum.Address{}, uErr
		}
		result, supported := addrs[common.AssetID(asset.ID)]
		if !supported || commonv3.IsZeroAddress(result) {
			return result, fmt.Errorf("real deposit address of token %s is not available", tokenID)
		}
		return result, nil
	}
	return ethereum.HexToAddress(liveAddress.Data[0].Address), nil
}

// Address return the deposit address of a token in Huobi exchange.
// Due to the logic of Huobi exchange, every token if supported will be
// deposited to an Intermediator address instead.
func (h *Huobi) Address(asset commonv3.Asset) (ethereum.Address, bool) {
	var exhSymbol string
	for _, exchange := range asset.Exchanges {
		if exchange.ExchangeID == uint64(common.Huobi) {
			exhSymbol = exchange.Symbol
		}
	}
	result := h.blockchain.GetIntermediatorAddr()
	_, err := h.RealDepositAddress(exhSymbol, asset)
	//if the realDepositAddress can not be querried, that mean the token isn't supported on Huobi
	if err != nil {
		return result, false
	}
	return result, true
}

// TokenPairs return all token pair support by Huobi
func (h *Huobi) TokenPairs() ([]commonv3.TradingPairSymbols, error) {
	pairs, err := h.sr.GetTradingPairs(uint64(common.Huobi))
	if err != nil {
		return nil, err
	}
	return pairs, nil
}

// QueryOrder return order status
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

// Trade on Huobi
func (h *Huobi) Trade(tradeType string, pair commonv3.TradingPairSymbols, rate float64, amount float64) (id string, done float64, remaining float64, finished bool, err error) {
	result, err := h.interf.Trade(tradeType, pair, rate, amount)

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
		pair.BaseSymbol+pair.QuoteSymbol,
		orderID,
	)
	if err != nil {
		h.l.Warnw("Huobi Query order error", "err", err)
	}
	return result.OrderID, done, remaining, finished, err
}

//Withdraw return withdraw id from huobi
func (h *Huobi) Withdraw(asset commonv3.Asset, amount *big.Int, address ethereum.Address) (string, error) {
	withdrawID, err := h.interf.Withdraw(asset, amount, address)
	if err != nil {
		return "", err
	}
	return withdrawID, err
}

// CancelOrder cancel an order from Huobi
func (h *Huobi) CancelOrder(id, base, quote string) error {
	idNo, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return err
	}
	symbol := base + quote
	result, err := h.interf.CancelOrder(symbol, idNo)
	if err != nil {
		return err
	}
	if result.Status != "ok" {
		return errors.New("Huobi Couldn't cancel order id " + id)
	}
	return nil
}

// FetchOnePairData return data of one pair
func (h *Huobi) FetchOnePairData(
	wg *sync.WaitGroup,
	pair commonv3.TradingPairSymbols,
	data *sync.Map,
	timepoint uint64) {

	defer wg.Done()
	result := common.ExchangePrice{}

	timestamp := common.Timestamp(fmt.Sprintf("%d", timepoint))
	result.Timestamp = timestamp
	result.Valid = true
	respData, err := h.interf.GetDepthOnePair(pair.BaseSymbol, pair.QuoteSymbol)
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
	data.Store(pair.ID, result)
}

// FetchPriceData return price data from Huobi
func (h *Huobi) FetchPriceData(timepoint uint64) (map[uint64]common.ExchangePrice, error) {
	wait := sync.WaitGroup{}
	data := sync.Map{}
	pairs, err := h.TokenPairs()
	if err != nil {
		return nil, err
	}
	for _, pair := range pairs {
		wait.Add(1)
		go h.FetchOnePairData(&wait, pair, &data, timepoint)
	}
	wait.Wait()
	result := map[uint64]common.ExchangePrice{}
	data.Range(func(key, value interface{}) bool {
		tokenPairID, ok := key.(uint64)
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

// FetchEBalanceData return account balance
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
		result.AvailableBalance = map[common.AssetID]float64{}
		result.LockedBalance = map[common.AssetID]float64{}
		result.DepositBalance = map[common.AssetID]float64{}
		result.Status = true
		if respData.Status != "ok" {
			result.Valid = false
			result.Error = fmt.Sprintf("Cannot fetch ebalance")
			result.Status = false
		} else {
			assets, err := h.sr.GetAssets()
			if err != nil {
				return common.EBalanceEntry{}, err
			}

			balances := respData.Data.List
			for _, b := range balances {
				tokenSymbol := strings.ToUpper(b.Currency)
				for _, asset := range assets {
					for _, exchg := range asset.Exchanges {
						if exchg.ExchangeID == uint64(common.Huobi) && exchg.Symbol == tokenSymbol {
							balance, _ := strconv.ParseFloat(b.Balance, 64)
							if b.Type == "trade" {
								result.AvailableBalance[common.AssetID(asset.ID)] = balance
							} else {
								result.LockedBalance[common.AssetID(asset.ID)] = balance
							}
							result.DepositBalance[common.AssetID(asset.ID)] = 0
						}
					}
				}
			}
		}
	}
	return result, nil
}

// FetchOnePairTradeHistory return trade history of one pair token
func (h *Huobi) FetchOnePairTradeHistory(pair commonv3.TradingPairSymbols) ([]common.TradeHistory, error) {
	var result []common.TradeHistory
	resp, err := h.interf.GetAccountTradeHistory(pair.BaseSymbol, pair.QuoteSymbol)
	if err != nil {
		return nil, err
	}
	for _, trade := range resp.Data {
		price, err := strconv.ParseFloat(trade.Price, 64)
		if err != nil {
			return nil, errors.Wrapf(err, "Can not parse price: %v", trade.Price)
		}
		quantity, err := strconv.ParseFloat(trade.Amount, 64)
		if err != nil {
			return nil, errors.Wrapf(err, "Can not parse price error: %v", trade.Amount)
		}
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
	return result, nil
}

//FetchTradeHistory get all trade history for all pairs from huobi exchange
func (h *Huobi) FetchTradeHistory() {
	pairs, err := h.TokenPairs()
	if err != nil {
		h.l.Warnw("Huobi fetch trade history failed. This might due to pairs setting hasn't been init yet", "err", err)
		return
	}
	var (
		result = map[uint64][]common.TradeHistory{}
		guard  = &sync.Mutex{}
		wait   = &sync.WaitGroup{}
	)

	for _, pair := range pairs {
		wait.Add(1)
		go func(pair commonv3.TradingPairSymbols) {
			defer wait.Done()
			histories, err := h.FetchOnePairTradeHistory(pair)
			if err != nil {
				h.l.Warnw("Cannot fetch data for pair", "pair", fmt.Sprintf("%s%s", pair.BaseSymbol, pair.QuoteSymbol), "err", err)
				return
			}
			guard.Lock()
			result[pair.ID] = histories
			guard.Unlock()
		}(pair)
	}
	wait.Wait()
	if err := h.storage.StoreTradeHistory(result); err != nil {
		h.l.Warnw("Store trade history error", "err", err)
	}
}

// GetTradeHistory return list of trade history
func (h *Huobi) GetTradeHistory(fromTime, toTime uint64) (common.ExchangeTradeHistory, error) {
	return h.storage.GetTradeHistory(fromTime, toTime)
}

// Send2ndTransaction send the second transaction
func (h *Huobi) Send2ndTransaction(amount float64, asset commonv3.Asset, exchangeAddress ethereum.Address) (*types.Transaction, error) {
	IAmount := common.FloatToBigInt(amount, int64(asset.Decimals))
	// Check balance, removed from huobi's blockchain object.
	// currBalance := h.blockchain.CheckBalance(token)
	// log.Printf("current balance of token %s is %d", token.ID, currBalance)
	// //h.blockchain.
	// if currBalance.Cmp(IAmount) < 0 {
	// 	log.Printf("balance is not enough, wait till next check")
	// 	return nil, errors.New("balance is not enough")
	// }
	var tx *types.Transaction
	var err error
	// TODO: add a check isETH that matching id instead of symbol
	if asset.Symbol == "ETH" {
		tx, err = h.blockchain.SendETHFromAccountToExchange(IAmount, exchangeAddress)
	} else {
		tx, err = h.blockchain.SendTokenFromAccountToExchange(IAmount, exchangeAddress, asset.Address)
	}
	if err != nil {
		h.l.Warnw("ERROR: Can not send transaction to exchange", "err", err)
		return nil, err
	}
	h.l.Infof("Transaction submitted. Tx is: %v", tx)
	return tx, nil

}

// PendingIntermediateTxs return list of intermediate pending txs
func (h *Huobi) PendingIntermediateTxs() (map[common.ActivityID]common.TXEntry, error) {
	result, err := h.storage.GetPendingIntermediateTXs()
	if err != nil {
		return nil, err
	}
	return result, nil
}

// FindTx2InPending find the second transaction from pending list
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

func (h *Huobi) exchangeDepositStatus(id common.ActivityID, tx2Entry common.TXEntry, assetID uint64, sentAmount float64) (string, error) {
	assets, err := h.sr.GetAssets()
	if err != nil {
		h.l.Warnw("Huobi ERROR: Can not get list of assets from setting", "err", err)
		return "", err
	}

	// make sure the size is enough for storing all deposit history
	deposits, err := h.interf.DepositHistory(len(assets) * 2)
	if err != nil || deposits.Status != "ok" {
		h.l.Warnw("Huobi Getting deposit history from huobi failed", "err", err, "status", deposits)
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
					h.ID().String(),
					assetID,
					"mined",
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
	h.l.Infof("Huobi Deposit doesn't exist. Huobi hasn't recognized the deposit yet or in theory, you have more than %d deposits at the same time.", len(assets)*2)
	return "", nil
}

func (h *Huobi) process1stTx(id common.ActivityID, tx1Hash string, assetID uint64, sentAmount float64) (string, error) {
	status, blockno, err := h.blockchain.TxStatus(ethereum.HexToHash(tx1Hash))
	if err != nil {
		h.l.Warnw("Huobi Can not get TX status", "err", err, "tx", tx1Hash)
		return "", nil
	}
	h.l.Infof("Huobi Status for Tx1 was %s at block %d ", status, blockno)
	if status == common.MiningStatusMined {
		//if it is mined, send 2nd tx.
		h.l.Infof("Found a new deposit status, which deposit %f %d. Procceed to send it to Huobi", sentAmount, assetID)
		//check if the asset is supported, the asset can be active or inactivee
		asset, err := h.sr.GetAsset(assetID)
		if err != nil {
			return "", err
		}

		var exhSymbol string
		for _, exchg := range asset.Exchanges {
			if exchg.ExchangeID == uint64(common.Huobi) {
				exhSymbol = exchg.Symbol
			}
		}

		exchangeAddress, err := h.RealDepositAddress(exhSymbol, asset)
		if err != nil {
			return "", err
		}
		tx2, err := h.Send2ndTransaction(sentAmount, asset, exchangeAddress)
		if err != nil {
			h.l.Infof("Huobi Trying to send 2nd tx failed, error: %s. Will retry next time", err.Error())
			return "", nil
		}
		//store tx2 to pendingIntermediateTx
		data := common.NewTXEntry(
			tx2.Hash().Hex(),
			h.ID().String(),
			assetID,
			common.MiningStatusSubmitted,
			"",
			sentAmount,
			common.GetTimestamp(),
		)
		if err = h.storage.StorePendingIntermediateTx(id, data); err != nil {
			h.l.Warnw("Trying to store 2nd tx to pending tx storage failed, error: %s. It will be ignored and "+
				"can make us to send to huobi again and the deposit will be marked as failed because the fund is not efficient", "err", err)
		}
		return "", nil
	}
	//No need to handle other blockchain status of TX1 here, since Fetcher will handle it from blockchain Status.
	return "", nil
}

// DepositStatus return status of a deposit
func (h *Huobi) DepositStatus(id common.ActivityID, tx1Hash string, assetID uint64, sentAmount float64, timepoint uint64) (string, error) {
	var data common.TXEntry
	tx2Entry, found := h.FindTx2(id)
	//if not found, meaning there is no tx2 yet, process 1st Tx and send 2nd Tx.
	if !found {
		return h.process1stTx(id, tx1Hash, assetID, sentAmount)
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
			h.ID().String(),
			assetID,
			common.MiningStatusMined,
			"",
			sentAmount,
			common.GetTimestamp())
		if uErr := h.storage.StorePendingIntermediateTx(id, data); uErr != nil {
			h.l.Warnw("Huobi Trying to store intermediate tx to huobi storage, error. Ignore it and try later", "err", uErr)
			return "", nil
		}
		return h.exchangeDepositStatus(id, tx2Entry, assetID, sentAmount)
	case common.MiningStatusFailed:
		data = common.NewTXEntry(
			tx2Entry.Hash,
			h.ID().String(),
			assetID,
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
		elapsed := common.NowInMillis() - tx2Entry.Timestamp.Millis()
		if elapsed > uint64(15*time.Minute/time.Millisecond) {
			data = common.NewTXEntry(
				tx2Entry.Hash,
				h.ID().String(),
				assetID,
				common.MiningStatusLost,
				common.ExchangeStatusLost,
				sentAmount,
				common.GetTimestamp(),
			)
			if err = h.storage.StoreIntermediateTx(id, data); err != nil {
				h.l.Warnw("Huobi Trying to store intermediate tx failed. Ignore it and treat it like it is still pending", "err", err)
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
	id string, assetID uint64, amount float64, timepoint uint64) (string, string, error) {
	withdrawID, _ := strconv.ParseUint(id, 10, 64)
	assets, err := h.sr.GetAssets()
	if err != nil {
		return "", "", fmt.Errorf("huobi Can't get list of assets from setting (%s)", err)
	}
	// make sure the size is enough for storing all huobi withdrawal history
	withdraws, err := h.interf.WithdrawHistory(len(assets) * 2)
	if err != nil {
		return "", "", fmt.Errorf("can't get withdraw history from huobi: %s", err.Error())
	}
	h.l.Infof("Huobi Withdrawal id: %d", withdrawID)
	for _, withdraw := range withdraws.Data {
		if withdraw.ID == withdrawID {
			if withdraw.State == "confirmed" {
				if withdraw.TxHash[0:2] != "0x" {
					withdraw.TxHash = "0x" + withdraw.TxHash
				}
				return common.ExchangeStatusDone, withdraw.TxHash, nil
			}
			return "", withdraw.TxHash, nil
		}
	}
	return "", "", errors.New("huobi Withdrawal doesn't exist. This shouldn't happen unless tx returned from withdrawal from huobi and activity ID are not consistently designed")
}

// OrderStatus return order status from Huobi
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

// ID return exchange ID
func (h *Huobi) ID() common.ExchangeID {
	return common.Huobi
}

// OpenOrders get open orders from binance
func (h *Huobi) OpenOrders(pair commonv3.TradingPairSymbols) ([]common.Order, error) {
	var (
		logger     = h.l.With("func", caller.GetCurrentFunctionName())
		result     []common.Order
		pairs      []commonv3.TradingPairSymbols
		err        error
		errGroup   errgroup.Group
		openOrders sync.Map
	)
	if pair.BaseSymbol == "" {
		logger.Info("No pair token provided, get open orders for all token")
		pairs, err = h.TokenPairs()
		if err != nil {
			return nil, err
		}
	} else {
		pairs = append(pairs, pair)
	}
	for _, pair := range pairs {
		errGroup.Go(
			func(pair commonv3.TradingPairSymbols) func() error {
				return func() error {
					logger.Infow("get open orders for pair", "pair", pair.BaseSymbol+pair.QuoteSymbol)
					orders, err := h.interf.OpenOrdersForOnePair(pair)
					if err != nil {
						return err
					}
					for _, order := range orders.Data {
						originalQty, err := strconv.ParseFloat(order.OrigQty, 64)
						if err != nil {
							return err
						}
						price, err := strconv.ParseFloat(order.Price, 64)
						if err != nil {
							return err
						}
						openOrders.Store(order.OrderID, common.Order{
							OrderID: strconv.FormatUint(order.OrderID, 10),
							Type:    order.Type,
							OrigQty: originalQty,
							Price:   price,
							Base:    pair.BaseSymbol,
							Quote:   pair.QuoteSymbol,
						})
					}
					return nil
				}
			}(pair),
		)
	}

	if err := errGroup.Wait(); err != nil {
		return result, err
	}

	openOrders.Range(func(key, value interface{}) bool {
		order, ok := value.(common.Order)
		if !ok {
			return false
		}
		result = append(result, order)
		return true
	})

	return result, nil
}

//NewHuobi creates new Huobi exchange instance
func NewHuobi(
	interf HuobiInterface,
	blockchain *blockchain.BaseBlockchain,
	signer blockchain.Signer,
	nonce blockchain.NonceCorpus,
	storage HuobiStorage,
	sr storage.SettingReader,
) (*Huobi, error) {

	bc, err := huobiblockchain.NewBlockchain(blockchain, signer, nonce)
	if err != nil {
		return nil, err
	}

	huobiObj := Huobi{
		interf:     interf,
		blockchain: bc,
		storage:    storage,
		sr:         sr,
		HuobiLive: HuobiLive{
			interf: interf,
		},
		l: zap.S(),
	}
	huobiServer := huobihttp.NewHuobiHTTPServer(&huobiObj)
	go huobiServer.Run()
	return &huobiObj, nil
}
