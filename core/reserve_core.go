package core

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"time"

	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/blockchain"
	"github.com/KyberNetwork/reserve-data/common/gasinfo"
	"github.com/KyberNetwork/reserve-data/data/storage"
	"github.com/KyberNetwork/reserve-data/lib/caller"
	"github.com/KyberNetwork/reserve-data/lib/rtypes"
	commonv3 "github.com/KyberNetwork/reserve-data/reservesetting/common"
)

// ReserveCore instance
type ReserveCore struct {
	blockchain      Blockchain
	activityStorage ActivityStorage
	addressConf     *common.ContractAddressConfiguration
	l               *zap.SugaredLogger
	gasPriceInfo    *gasinfo.GasPriceInfo
	// nonce value will be use in deposit transaction
	depositNonce int64
}

// NewReserveCore return reserve core
func NewReserveCore(
	blockchain Blockchain,
	storage ActivityStorage,
	addressConf *common.ContractAddressConfiguration,
	gasPriceInfo *gasinfo.GasPriceInfo) *ReserveCore {
	return &ReserveCore{
		blockchain:      blockchain,
		activityStorage: storage,
		addressConf:     addressConf,
		l:               zap.S(),
		gasPriceInfo:    gasPriceInfo,
	}
}

func timebasedID(id string) common.ActivityID {
	return common.NewActivityID(uint64(time.Now().UnixNano()), id)
}

// Transfer move fund between main and sub account
func (rc *ReserveCore) Transfer(fromAccount, toAccount string, asset commonv3.Asset, amount *big.Int, exchange common.Exchange) (string, error) {
	return exchange.Transfer(fromAccount, toAccount, asset, amount, false, "")
}

// CancelOrders cancel orders on centralized exchanges
func (rc *ReserveCore) CancelOrders(orders []common.RequestOrder, exchange common.Exchange) map[string]common.CancelOrderResult {
	var (
		logger = rc.l.With("func", caller.GetCurrentFunctionName())
	)
	result := make(map[string]common.CancelOrderResult)
	for _, order := range orders {
		_, err := rc.activityStorage.GetActivity(exchange.ID(), order.ID)
		if err != nil && err != storage.ErrorNotFound {
			logger.Warnw("failed to get order", "order id", order.ID, "exchange", exchange.ID().String(), "error", err)
			result[order.ID] = common.CancelOrderResult{
				Success: false,
				Error:   err.Error(),
			}
			continue
		}
		if err := exchange.CancelOrder(order.ID, order.Symbol); err != nil {
			result[order.ID] = common.CancelOrderResult{
				Success: false,
				Error:   err.Error(),
			}
		} else {
			result[order.ID] = common.CancelOrderResult{
				Success: true,
			}
		}
	}
	return result
}

// Trade token on centralized exchange
func (rc *ReserveCore) Trade(
	exchange common.Exchange,
	tradeType string,
	pair commonv3.TradingPairSymbols,
	rate float64,
	amount float64) (common.ActivityID, float64, float64, bool, error) {
	var (
		err       error
		isPending bool = true
	)

	timepoint := common.NowInMillis()
	recordActivity := func(id, status string, done, remaining float64, finished bool, err error) error {
		uid := timebasedID(id)
		rc.l.Infof(
			"Core ----------> %s on %s: base: %s, quote: %s, rate: %s, amount: %s, timestamp: %d ==> Result: id: %s, done: %s, remaining: %s, finished: %t, error: %v",
			tradeType, exchange.ID().String(), pair.BaseSymbol, pair.QuoteSymbol,
			strconv.FormatFloat(rate, 'f', -1, 64),
			strconv.FormatFloat(amount, 'f', -1, 64), timepoint,
			uid,
			strconv.FormatFloat(done, 'f', -1, 64),
			strconv.FormatFloat(remaining, 'f', -1, 64),
			finished, err,
		)

		activityResult := common.ActivityResult{
			ID:        id,
			Done:      done,
			Remaining: remaining,
			Finished:  finished,
			Error:     "",
		}

		if err != nil {
			activityResult.Error = err.Error()
		}

		if finished {
			isPending = false
		}

		return rc.activityStorage.Record(
			common.ActionTrade,
			uid,
			exchange.ID().String(),
			common.ActivityParams{
				Exchange:      exchange.ID(),
				Type:          tradeType,
				Base:          pair.BaseSymbol,
				Quote:         pair.QuoteSymbol,
				Rate:          rate,
				Amount:        amount,
				Timepoint:     timepoint,
				TradingPairID: pair.ID,
			},
			activityResult,
			status,
			"",
			timepoint,
			isPending,
			common.NowInMillis(),
		)
	}

	if err = sanityCheckTrading(pair, rate, amount); err != nil {
		if sErr := recordActivity("", common.ExchangeStatusFailed, 0, 0, false, err); sErr != nil {
			rc.l.Warnw("failed to save activity record", "err", sErr)
			return common.ActivityID{}, 0, 0, false, common.CombineActivityStorageErrs(err, sErr)
		}
		return common.ActivityID{}, 0, 0, false, err
	}

	id, done, remaining, finished, err := exchange.Trade(tradeType, pair, rate, amount)
	uid := timebasedID(id)
	if err != nil {
		if sErr := recordActivity(id, common.ExchangeStatusFailed, done, remaining, true, err); sErr != nil {
			rc.l.Warnw("failed to save activity record", "err", sErr)
			return uid, done, remaining, finished, common.CombineActivityStorageErrs(err, sErr)
		}
		return uid, done, remaining, finished, err
	}

	var status string
	if finished {
		status = common.ExchangeStatusDone
	} else {
		status = common.ExchangeStatusSubmitted
	}

	sErr := recordActivity(id, status, done, remaining, finished, nil)
	return uid, done, remaining, finished, common.CombineActivityStorageErrs(err, sErr)
}

// TransferToSelf utility func to override nonce, this trigger manual in case core can't resolve account nonce automatically.
func (rc *ReserveCore) TransferToSelf(op string, nonce uint64, recommendedPrice float64) (*types.Transaction, error) {
	var action string
	switch op {
	case blockchain.DepositOP:
		action = common.ActionDeposit
	case blockchain.PricingOP:
		action = common.ActionSetRate
	default:
		return nil, fmt.Errorf("op %s is invalid", op)
	}
	l := rc.l.With("op", op, "nonce", nonce, "gas_price", recommendedPrice)
	// transfer to self use to override tx with nonce in rage (minedNonce, maxPendingNonce]
	err := rc.validateNonceInRange(op, nonce, action)
	if err != nil {
		l.Errorw("check nonce to override tx", "err", err)
		return nil, err
	}

	if recommendedPrice == 0 {
		recommendedPrice, err = rc.gasPriceInfo.GetCurrentGas()
		if err != nil {
			l.Errorw("transfer-self to get gas price", "err", err)
			return nil, fmt.Errorf("failed to get gas: %w", err)
		}
	}
	highBoundGasPrice := rc.maxGasPrice()
	if recommendedPrice == 0 || recommendedPrice > highBoundGasPrice {
		l.Errorw("transfer-self failed due gasprice invalid", "gas", recommendedPrice, "highbound", highBoundGasPrice)
		return nil, fmt.Errorf("gasprice invalid, current price %v, highbound %v", recommendedPrice, highBoundGasPrice)
	}
	gasPrice := common.GweiToWei(recommendedPrice)
	tx, err := rc.blockchain.TransferToSelf(op, gasPrice, big.NewInt(0).SetUint64(nonce))
	if err != nil {
		l.Errorw("transfer-self failed", "err", err, "op", op, "gasPrice", gasPrice.String(), "nonce", nonce)
	} else {
		l.Infow("transfer-self", "tx", tx.Hash().String(), "op", op, "gasPrice", gasPrice.String(), "nonce", nonce)
	}
	return tx, err
}

func (rc *ReserveCore) validateNonceInRange(op string, nonce uint64, action string) error {
	minedNonce, err := rc.blockchain.GetMinedNonceWithOP(op)
	if err != nil {
		return fmt.Errorf("couldn't get mined nonce, %w", err)
	}
	rc.l.Debugw("mined nonce account", "op", op, "nonce", minedNonce)
	if nonce <= minedNonce {
		return fmt.Errorf("override nonce must greater than minedNonce %v", minedNonce)
	}

	maxPendingNonce, err := rc.activityStorage.MaxPendingNonce(action)
	if err != nil || maxPendingNonce == 0 {
		maxPendingNonce = int64(minedNonce)
	}
	if nonce > uint64(maxPendingNonce) {
		return fmt.Errorf("cannot override for nonce > maxPendingNonce %v", maxPendingNonce)
	}
	return nil
}

// Deposit deposit token into centralized exchange
func (rc *ReserveCore) Deposit(
	exchange common.Exchange,
	asset commonv3.Asset,
	amount *big.Int,
	timepoint uint64) (common.ActivityID, error) {
	amountFloat := common.BigToFloat(amount, int64(asset.Decimals))
	uidGenerator := func(txhex string) common.ActivityID {
		id := fmt.Sprintf("%s|%s|%s",
			txhex,
			asset.Symbol,
			strconv.FormatFloat(amountFloat, 'f', -1, 64),
		)
		return timebasedID(id)
	}
	recordActivity := func(status, txhex string, txnonce uint64, txprice string, err error) error {
		uid := uidGenerator(txhex)
		rc.l.Infof(
			"Core ----------> Deposit to %s: token: %s, amount: %s, timestamp: %d ==> Result: tx: %s, error: %v",
			exchange.ID().String(), asset.Symbol, amount.Text(10), timepoint, txhex, err,
		)

		activityResult := common.ActivityResult{
			Tx:       txhex,
			Nonce:    txnonce,
			GasPrice: txprice,
			Error:    "",
			TxTime:   common.TimeToMillis(time.Now()),
		}

		if err != nil {
			activityResult.Error = err.Error()
		}

		return rc.activityStorage.Record(
			common.ActionDeposit,
			uid,
			exchange.ID().String(),
			common.ActivityParams{
				Exchange:  exchange.ID(),
				Asset:     asset.ID,
				Amount:    amountFloat,
				Timepoint: timepoint,
			},
			activityResult,
			common.ExchangeStatusNA,
			status,
			timepoint,
			status != common.MiningStatusFailed,
			common.NowInMillis(),
		)
	}

	tx, err := rc.doDeposit(exchange, asset, amount)
	if err != nil {
		sErr := recordActivity(common.MiningStatusFailed, "", 0, "", err)
		if sErr != nil {
			rc.l.Errorw("failed to save activity record", "err", sErr)
		}
		return common.ActivityID{}, common.CombineActivityStorageErrs(err, sErr)
	}
	gasPrice := ""
	if tx.GasPrice() != nil {
		gasPrice = tx.GasPrice().String()
	}
	rc.l.Infow("deposit_initialized", "exchange", exchange.ID().String(), "asset",
		asset.Address.String(), "name", asset.Name, "amount", amount.String(), "nonce", tx.Nonce(),
		"gas_price", gasPrice, "tx", tx.Hash().String())

	sErr := recordActivity(
		common.MiningStatusSubmitted,
		tx.Hash().Hex(),
		tx.Nonce(),
		tx.GasPrice().Text(10),
		nil,
	)
	return uidGenerator(tx.Hash().Hex()), common.CombineActivityStorageErrs(err, sErr)
}

func (rc *ReserveCore) maxGasPrice() float64 {
	// MaxGasPrice will fetch gasPrice from kyber network contract(with cache for configurable seconds)
	max, err := rc.gasPriceInfo.MaxGas()
	if err != nil {
		rc.l.Errorw("failed to receive maxGasPrice from network, fallback to hard code value",
			"err", err, "maxGasPrice", common.HighBoundGasPrice)
		return common.HighBoundGasPrice
	}
	return max
}
func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
func (rc *ReserveCore) doDeposit(exchange common.Exchange, asset commonv3.Asset, amount *big.Int) (tx *types.Transaction, err error) {

	address, supported := exchange.Address(asset)
	if !supported {
		return nil, fmt.Errorf("exchange %s doesn't support token %d", exchange.ID(), asset.ID)
	}
	found, err := rc.activityStorage.HasPendingDeposit(asset, exchange)
	if err != nil {
		return nil, err
	}
	if found {
		return nil, fmt.Errorf("there is a pending %d deposit to %s currently, please try again", asset.ID, exchange.ID())
	}
	if err = sanityCheckAmount(exchange, asset, amount); err != nil {
		return nil, err
	}
	// if there is a pending deposit tx, we replace it
	var (
		initPrice     *big.Int
		selectedNonce int64
	)
	minedNonce, err := rc.blockchain.GetMinedNonceWithOP(blockchain.DepositOP)
	if err != nil {
		rc.l.Errorw("couldn't get mined nonce of deposit operator", "err", err)
		return tx, fmt.Errorf("couldn't get mined nonce of deposit operator (%+v)", err)
	}
	rc.l.Debugw("mined nonce for deposit account", "nonce", minedNonce)
	if rc.depositNonce == 0 {
		maxPendingNonce, err := rc.activityStorage.MaxPendingNonce(common.ActionDeposit)
		if err != nil || maxPendingNonce == 0 {
			selectedNonce = int64(minedNonce)
		} else {
			selectedNonce = maxPendingNonce + 1
			rc.l.Debugw("use max pending activity nonce", "nonce", selectedNonce)
		}
	} else { // we have a nonce, use it
		selectedNonce = rc.depositNonce
	}
	selectedNonce = maxInt64(int64(minedNonce), selectedNonce) // select the bigger value, just in case our local nonce get delay
	rc.l.Debugw("selected nonce for deposit", "nonce", selectedNonce)

	recommendedPrice, err := rc.gasPriceInfo.GetCurrentGas()
	if err != nil {
		rc.l.Errorw("deposit failed to get gas price, use default", "err", err)
		return nil, fmt.Errorf("failed to get gas: %w", err)
	}
	highBoundGasPrice := rc.maxGasPrice()
	if recommendedPrice == 0 || recommendedPrice > highBoundGasPrice {
		rc.l.Errorw("deposit failed due gasprice invalid", "gas", recommendedPrice, "highbound", highBoundGasPrice)
		return nil, fmt.Errorf("gasprice invalid, current price %v, highbound %v", recommendedPrice, highBoundGasPrice)
	}
	initPrice = common.GweiToWei(recommendedPrice)
	rc.l.Infof("initial deposit tx, init price: %s", initPrice.String())

	if tx, err = rc.blockchain.Send(asset, amount, address, big.NewInt(selectedNonce), initPrice); err != nil {
		return nil, err
	}
	rc.depositNonce = selectedNonce + 1 // increase nonce if we success send a transaction, can be use for next transaction
	return tx, nil
}

// Withdraw token from exchange
func (rc *ReserveCore) Withdraw(exchange common.Exchange, asset commonv3.Asset, amount *big.Int) (common.ActivityID, error) {
	var err error
	timepoint := common.NowInMillis()
	activityRecord := func(id, status string, err error) error {
		uid := timebasedID(id)
		rc.l.Infof("Core ----------> Withdraw from %s: asset: %d, amount: %s, timestamp: %d ==> Result: id: %s, error: %s",
			exchange.ID().String(), asset.ID, amount.Text(10), timepoint, id, err,
		)
		acitivityResult := common.ActivityResult{
			ID: id,
			// this field will be updated with real tx when data fetcher can fetch it
			// from exchanges
			Tx:    "",
			Error: "",
		}
		// omitempty if err == nil
		if err != nil {
			acitivityResult.Error = err.Error()
		}
		return rc.activityStorage.Record(
			common.ActionWithdraw,
			uid,
			exchange.ID().String(),
			common.ActivityParams{
				Exchange:  exchange.ID(),
				Asset:     asset.ID,
				Amount:    common.BigToFloat(amount, int64(asset.Decimals)),
				Timepoint: timepoint,
			},
			acitivityResult,
			status,
			common.MiningStatusNA,
			timepoint,
			status != common.ExchangeStatusFailed && status != common.ExchangeStatusCancelled,
			common.NowInMillis(),
		)
	}

	_, supported := exchange.Address(asset)
	if !supported {
		err = fmt.Errorf("exchange %s doesn't support asset %d", exchange.ID().String(), asset.ID)
		sErr := activityRecord("", common.ExchangeStatusFailed, err)
		if sErr != nil {
			rc.l.Warnw("failed to store activity record", "err", sErr)
		}
		return common.ActivityID{}, common.CombineActivityStorageErrs(err, sErr)
	}

	if err = sanityCheckAmount(exchange, asset, amount); err != nil {
		sErr := activityRecord("", common.ExchangeStatusFailed, err)
		if sErr != nil {
			rc.l.Warnw("failed to store activity record", "err", sErr)
		}
		return common.ActivityID{}, common.CombineActivityStorageErrs(err, sErr)
	}

	reserveAddr := rc.addressConf.Reserve

	id, err := exchange.Withdraw(asset, amount, reserveAddr)
	if err != nil {
		rc.l.Errorw("init withdraw failed", "err", err, "asset", asset.Address.String(),
			"amount", amount.String(), "reserveAddr", reserveAddr)
		sErr := activityRecord("", common.ExchangeStatusFailed, err)
		if sErr != nil {
			rc.l.Warnw("failed to store activity record", "err", sErr)
		}
		return common.ActivityID{}, common.CombineActivityStorageErrs(err, sErr)
	}

	sErr := activityRecord(id, common.ExchangeStatusSubmitted, nil)
	return timebasedID(id), common.CombineActivityStorageErrs(err, sErr)
}

func calculateNewGasPrice(initPrice *big.Int, count uint64, highBoundGasPrice float64) *big.Int {
	// in this case after 5 tries the tx is still not mined.
	// at this point, 100.1 gwei is not enough but it doesn't matter
	// if the tx is mined or not because users' tx is not mined neither
	// so we can just increase the gas price a tiny amount (1 gwei) to make
	// the node accept tx with up to date price
	if count > 4 {
		return big.NewInt(0).Add(
			common.GweiToWei(highBoundGasPrice),
			common.GweiToWei(float64(count)-4.0))
	}
	// new = initPrice * (high bound / initPrice)^(step / 4)
	initPriceFloat := common.BigToFloat(initPrice, 9) // convert Gwei int to float
	base := highBoundGasPrice / initPriceFloat
	newPrice := initPriceFloat * math.Pow(base, float64(count)/4.0)
	return common.FloatToBigInt(newPrice, 9)
}

func requireSameLength(tokens []commonv3.Asset, buys, sells, afpMids []*big.Int) error {
	if len(tokens) != len(buys) {
		return fmt.Errorf("number of buys (%d) is not equal to number of tokens (%d)", len(buys), len(tokens))
	}
	if len(tokens) != len(sells) {
		return fmt.Errorf("number of sell (%d) is not equal to number of tokens (%d)", len(sells), len(tokens))

	}
	if len(tokens) != len(afpMids) {
		return fmt.Errorf("number of afpMids (%d) is not equal to number of tokens (%d)", len(afpMids), len(tokens))
	}
	return nil
}

// CancelSetRate create and send a tx with higher gas price to cancel all pending set rate tx
func (rc *ReserveCore) CancelSetRate() (common.ActivityID, error) {
	highBoundGasPrice := rc.maxGasPrice()

	recommendedPrice, err := rc.gasPriceInfo.GetCurrentGas()
	if err != nil {
		rc.l.Errorw("failed to get gas price", "err", err)
		return common.ActivityID{}, fmt.Errorf("setRate failed to get gas price %w", err)
	}
	if recommendedPrice == 0 || recommendedPrice > highBoundGasPrice {
		rc.l.Errorw("failed to get gas price", "err", err, "gas",
			recommendedPrice, "highBound", highBoundGasPrice)
		return common.ActivityID{}, fmt.Errorf("cancel setRate failed to query gas price got value %v highBound %v",
			recommendedPrice, highBoundGasPrice)
	}
	minedNonce, err := rc.blockchain.GetMinedNonceWithOP(blockchain.PricingOP)
	if err != nil {
		return common.ActivityID{}, fmt.Errorf("couldn't get mined nonce of set rate operator (%+v)", err)
	}

	targetActivity, err := rc.activityStorage.GetActivityForOverride(common.ActionCancelSetRate, minedNonce)
	if err != nil {
		rc.l.Errorw("failed to get pending activity", "err", err)
		return common.ActivityID{}, fmt.Errorf("failed to get pending cancel setrate activity, %w", err)
	}
	if targetActivity == nil { // if no pending cancel setRate found, fallback to cancel recent setRate
		targetActivity, err = rc.activityStorage.GetActivityForOverride(common.ActionSetRate, minedNonce)
	}
	if err != nil {
		rc.l.Errorw("failed to get pending activity", "err", err)
		return common.ActivityID{}, fmt.Errorf("failed to get pending activity, %w", err)
	}
	if targetActivity == nil {
		msg := "no pending setRate to cancel"
		rc.l.Infow(msg)
		return common.ActivityID{}, fmt.Errorf(msg)
	}

	pendingTxGasPriceBig, ok := new(big.Int).SetString(targetActivity.Result.GasPrice, 10)
	if !ok {
		rc.l.Errorw("get invalid gasPrice from activity", "tx_result", targetActivity.Result,
			"activity", targetActivity.ID.String())
		return common.ActivityID{}, fmt.Errorf("get invalid gasPrice from activity, stored price %s",
			targetActivity.Result.GasPrice)
	}
	pendingTxGasFloat := common.BigToFloat(pendingTxGasPriceBig, 9)
	newGasPrice := common.CalculateNewPrice(pendingTxGasFloat, recommendedPrice)
	if newGasPrice > highBoundGasPrice {
		rc.l.Errorw("abort override cancel setRate as gasPrice too high", "pendingTxPrice", pendingTxGasFloat,
			"newGasPrice", newGasPrice, "highBound", highBoundGasPrice)
		return common.ActivityID{}, fmt.Errorf("abort override setRate as gasPrice too high")
	}
	newPrice := common.FloatToBigInt(newGasPrice, 9)

	rc.l.Infow("cancel setRate tx with info", "newPrice", newPrice.String(),
		"highBoundGasPrice", highBoundGasPrice, "nonce", targetActivity.Result.Nonce)

	tx, err := rc.blockchain.BuildSendETHTx(blockchain.TxOpts{
		Nonce:    big.NewInt(int64(targetActivity.Result.Nonce)),
		Value:    big.NewInt(0),
		GasPrice: newPrice,
	}, rc.blockchain.GetPricingOPAddress())
	if err != nil {
		rc.l.Errorw("failed to build cancel setRate tx", "err", err)
		return common.ActivityID{}, err
	}

	var (
		txhex        = ethereum.Hash{}.Hex()
		txprice      = "0"
		miningStatus string
		txNonce      = uint64(0)
		errResult    = ""
	)

	btx, err := rc.blockchain.SignAndBroadcast(tx, blockchain.PricingOP)
	if err != nil {
		rc.l.Errorw("failed to sign and broadcast tx", "err", err)
		miningStatus = common.MiningStatusFailed
		errResult = err.Error()
	} else {
		miningStatus = common.MiningStatusSubmitted
		txhex = btx.Hash().Hex()
		txNonce = btx.Nonce()
		txprice = btx.GasPrice().Text(10)
	}
	uid := timebasedID(txhex)
	activityResult := common.ActivityResult{
		Tx:       txhex,
		Nonce:    txNonce,
		GasPrice: txprice,
		Error:    errResult,
	}
	sErr := rc.activityStorage.Record(
		common.ActionCancelSetRate,
		uid,
		"blockchain",
		common.ActivityParams{},
		activityResult,
		common.ExchangeStatusNA,
		miningStatus,
		common.NowInMillis(),
		miningStatus != common.MiningStatusFailed,
		common.NowInMillis(),
	)

	rc.l.Infow("sent cancel setRate tx", "tx", btx.Hash().String(), "gasPrice",
		newPrice.String(), "nonce", targetActivity.Result.Nonce)

	return uid, common.CombineActivityStorageErrs(err, sErr)
}

// GetSetRateResult return result of set rate action
func (rc *ReserveCore) GetSetRateResult(tokens []commonv3.Asset,
	buys, sells, afpMids []*big.Int,
	block *big.Int) (*types.Transaction, uint64, error) {
	var (
		tx  *types.Transaction
		err error
	)
	err = requireSameLength(tokens, buys, sells, afpMids)
	if err != nil {
		return tx, common.NowInMillis(), err
	}
	if err = sanityCheck(buys, afpMids, sells, rc.l); err != nil {
		return tx, common.NowInMillis(), err
	}
	var tokenAddrs []ethereum.Address
	for _, token := range tokens {
		tokenAddrs = append(tokenAddrs, token.Address)
	}
	// if there is a pending set rate tx, we replace it
	var (
		minedNonce uint64
	)
	highBoundGasPrice := rc.maxGasPrice()
	minedNonce, err = rc.blockchain.GetMinedNonceWithOP(blockchain.PricingOP)
	if err != nil {
		return tx, common.NowInMillis(), fmt.Errorf("couldn't get mined nonce of set rate operator (%s)", err.Error())
	}
	pendingSetRate, err := rc.activityStorage.GetActivityForOverride(common.ActionSetRate, minedNonce)
	if err != nil {
		rc.l.Errorw("failed to get pending activity", "err", err)
		return nil, common.NowInMillis(), fmt.Errorf("failed to get pending activity, %w", err)
	}
	recommendedPrice, err := rc.gasPriceInfo.GetCurrentGas()
	if err != nil {
		rc.l.Errorw("failed to get gas price", "err", err)
		return nil, common.NowInMillis(), fmt.Errorf("setRate failed to get gas price %w", err)
	}
	if recommendedPrice == 0 || recommendedPrice > highBoundGasPrice {
		rc.l.Errorw("failed to get gas price", "err", err, "gas",
			recommendedPrice, "highBound", highBoundGasPrice)
		return nil, common.NowInMillis(), fmt.Errorf("setrate failed to query gas price got value %v highBound %v",
			recommendedPrice, highBoundGasPrice)
	}
	if pendingSetRate != nil {
		pendingTxGasPriceBig, ok := new(big.Int).SetString(pendingSetRate.Result.GasPrice, 10)
		if !ok {
			rc.l.Errorw("get invalid gasPrice from activity", "tx_result", pendingSetRate.Result, "activity", pendingSetRate.ID.String())
			return nil, common.NowInMillis(), fmt.Errorf("get invalid gasPrice from activity")
		}
		pendingTxGasFloat := common.BigToFloat(pendingTxGasPriceBig, 9)
		newGasPrice := common.CalculateNewPrice(pendingTxGasFloat, recommendedPrice)
		l := rc.l.With("current_price", pendingTxGasFloat, "new_price", newGasPrice, "current_tx", pendingSetRate.Result.Tx)
		if newGasPrice > highBoundGasPrice {
			l.Errorw("abort override setRate as gasPrice too high", "highBound", highBoundGasPrice)
			return nil, common.NowInMillis(), fmt.Errorf("abort override setRate as gasPrice too high")
		}
		newPrice := common.FloatToBigInt(newGasPrice, 9)
		tx, err = rc.blockchain.SetRates(tokenAddrs, buys, sells, block, big.NewInt(int64(pendingSetRate.Result.Nonce)), newPrice)
		if err != nil {
			l.Errorw("Trying to replace old tx failed", "err", err)
			return tx, pendingSetRate.OrgTime, err
		}
		l.Infow("sent tx with new price", "new_tx", tx.Hash().String())
		return tx, pendingSetRate.OrgTime, err
	}

	initPrice := common.GweiToWei(recommendedPrice)
	rc.l.Infof("initial set rate tx, init price: %s", initPrice.String())
	tx, err = rc.blockchain.SetRates(
		tokenAddrs, buys, sells, block,
		big.NewInt(int64(minedNonce)),
		initPrice,
	)
	return tx, common.NowInMillis(), err
}

// SetRates to reserve
func (rc *ReserveCore) SetRates(assets []commonv3.Asset, buys, sells []*big.Int, block *big.Int, afpMids []*big.Int,
	msgs []string, triggers []bool) (common.ActivityID, error) {

	var (
		tx           *types.Transaction
		txhex        = ethereum.Hash{}.Hex()
		txnonce      = uint64(0)
		orgTime      uint64
		txprice      = "0"
		err          error
		miningStatus string
	)
	if block.Int64() == 0 {
		cb, err := rc.blockchain.CurrentBlock()
		if err != nil {
			return common.ActivityID{}, fmt.Errorf("cannot get current block: %v", err)
		}
		block = big.NewInt(int64(cb))
	}
	tx, orgTime, err = rc.GetSetRateResult(assets, buys, sells, afpMids, block)
	if err != nil {
		rc.l.Errorw("failed to get set rate result", "err", err)
		miningStatus = common.MiningStatusFailed
	} else {
		miningStatus = common.MiningStatusSubmitted
	}
	if tx != nil {
		txhex = tx.Hash().Hex()
		txnonce = tx.Nonce()
		txprice = tx.GasPrice().Text(10)
	}
	uid := timebasedID(txhex)
	assetsID := make([]rtypes.AssetID, 0, len(assets))
	for _, asset := range assets {
		assetsID = append(assetsID, asset.ID)
	}
	activityResult := common.ActivityResult{
		Tx:       txhex,
		Nonce:    txnonce,
		GasPrice: txprice,
		Error:    "",
	}
	if err != nil {
		activityResult.Error = err.Error()
	}
	sErr := rc.activityStorage.Record(
		common.ActionSetRate,
		uid,
		"blockchain",
		common.ActivityParams{
			Assets:   assetsID,
			Buys:     buys,
			Sells:    sells,
			Block:    block,
			AFPMid:   afpMids,
			Msgs:     msgs,
			Triggers: triggers,
		},
		activityResult,
		common.ExchangeStatusNA,
		miningStatus,
		common.NowInMillis(),
		miningStatus != common.MiningStatusFailed,
		orgTime,
	)
	rc.l.Infof(
		"Core ----------> Set rates: ==> Result: tx: %s, nonce: %d, price: %s, error: %v, storage error: %v",
		txhex, txnonce, txprice, err, sErr,
	)

	return uid, common.CombineActivityStorageErrs(err, sErr)
}

// SpeedupTx send a new tx with same info, with higher gas
func (rc *ReserveCore) SpeedupTx(act common.ActivityRecord) (*big.Int, error) {
	action := act.Action
	var opAccount string
	switch action {
	case common.ActionDeposit:
		opAccount = blockchain.DepositOP
	case common.ActionSetRate:
		opAccount = blockchain.PricingOP
	default:
		return nil, fmt.Errorf("speedup not support for action: %s", action)
	}
	newGas, err := rc.gasPriceInfo.GetCurrentGas()
	if err != nil {
		return nil, fmt.Errorf("speedup failed due can't get gas price, %w", err)
	}
	maxGas, err := rc.gasPriceInfo.MaxGas()
	if err != nil {
		maxGas = common.HighBoundGasPrice
	}
	l := rc.l.With("action", action, "activity", act.ID.String())
	tx := ethereum.HexToHash(act.Result.Tx)
	signedTx, err := rc.blockchain.SpeedupTx(tx, newGas, maxGas, opAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to speedup tx, %w", err)
	}
	l.Debugw("sent speedup tx", "new_tx", signedTx.Hash().String(), "src_tx", tx.String())

	// store new act to storage
	replaceUID := func(actID common.ActivityID, txhex string) common.ActivityID {
		parts := strings.Split(actID.EID, "|")
		if len(parts) != 3 {
			return actID // uid somehow malform
		}
		id := fmt.Sprintf("%s|%s|%s",
			txhex,
			parts[1],
			parts[2],
		)
		return timebasedID(id)
	}
	activityResult := common.ActivityResult{
		Tx:       signedTx.Hash().Hex(),
		Nonce:    act.Result.Nonce,
		GasPrice: signedTx.GasPrice().String(),
		Error:    "",
		TxTime:   common.TimeToMillis(time.Now()),
	}
	return signedTx.GasPrice(), rc.activityStorage.Record(
		action,
		replaceUID(act.ID, signedTx.Hash().Hex()),
		act.Params.Exchange.String(),
		*act.Params,
		activityResult,
		common.ExchangeStatusNA,
		common.MiningStatusSubmitted,
		common.TimeToMillis(time.Now()),
		true,
		act.OrgTime,
	)
}

func sanityCheck(buys, afpMid, sells []*big.Int, l *zap.SugaredLogger) error {
	eth := big.NewFloat(0).SetInt(common.EthToWei(1))
	for i, s := range sells {
		check := checkZeroValue(buys[i], s)
		switch check {
		case 1: // both buy/sell rate > 0
			sFloat := big.NewFloat(0).SetInt(s)
			sRate := calculateRate(sFloat, eth)
			bFloat := big.NewFloat(0).SetInt(buys[i])
			bRate := calculateRate(eth, bFloat)
			// aMFloat := big.NewFloat(0).SetInt(afpMid[i])
			// aMRate := calculateRate(aMFloat, eth)
			if bRate.Cmp(sRate) <= 0 {
				return errors.New("buy price must be bigger than sell price")
			}
		case 0: // both buy/sell rate is 0
			return nil
		case -1: // either buy/sell rate is 0
			if buys[i].Cmp(big.NewInt(0)) == 0 {
				return errors.New("buy rate can not be zero")
			}
			l.Warnw("sanityCheck sell rate is zero", "index", i, "buy_rate", buys[i].String())
		}
	}
	return nil
}

func sanityCheckTrading(pair commonv3.TradingPairSymbols, rate, amount float64) error {
	currentNotional := rate * amount
	minNotional := pair.MinNotional
	if minNotional != float64(0) {
		if currentNotional < minNotional {
			return errors.New("notional must be bigger than exchange's MinNotional")
		}
	}
	return nil
}

func sanityCheckAmount(exchange common.Exchange, asset commonv3.Asset, amount *big.Int) error {
	var feeWithdrawing float64
	for _, exchg := range asset.Exchanges {
		if exchg.ExchangeID == exchange.ID() {
			feeWithdrawing = exchg.WithdrawFee
		}
	}

	amountFloat := big.NewFloat(0).SetInt(amount)
	expDecimal := big.NewInt(0).Exp(big.NewInt(10), big.NewInt(0).SetUint64(asset.Decimals), nil)
	minAmountWithdraw := big.NewFloat(0)

	minAmountWithdraw.Mul(big.NewFloat(feeWithdrawing), big.NewFloat(0).SetInt(expDecimal))
	if amountFloat.Cmp(minAmountWithdraw) < 0 {
		return errors.New("amount is too small")
	}
	return nil
}

func calculateRate(theDividend, divisor *big.Float) *big.Float {
	div := big.NewFloat(0)
	div.Quo(theDividend, divisor)
	return div
}

func checkZeroValue(buy, sell *big.Int) int {
	zero := big.NewInt(0)
	if buy.Cmp(zero) == 0 && sell.Cmp(zero) == 0 {
		return 0
	}
	if buy.Cmp(zero) > 0 && sell.Cmp(zero) > 0 {
		return 1
	}
	return -1
}
