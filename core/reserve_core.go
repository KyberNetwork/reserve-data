package core

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"time"

	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data/common"
	commonv3 "github.com/KyberNetwork/reserve-data/reservesetting/common"
)

const (
	// highBoundGasPrice is the price we will try to use to get higher priority
	// than trade tx to avoid price front running from users.
	highBoundGasPrice float64 = 100.1

	statusFailed    = "failed"
	statusSubmitted = "submitted"
	statusDone      = "done"
)

// ReserveCore instance
type ReserveCore struct {
	blockchain      Blockchain
	activityStorage ActivityStorage
	addressConf     *common.ContractAddressConfiguration
	l               *zap.SugaredLogger
}

// NewReserveCore return reserve core
func NewReserveCore(
	blockchain Blockchain,
	storage ActivityStorage,
	addressConf *common.ContractAddressConfiguration) *ReserveCore {
	return &ReserveCore{
		blockchain:      blockchain,
		activityStorage: storage,
		addressConf:     addressConf,
		l:               zap.S(),
	}
}

func timebasedID(id string) common.ActivityID {
	return common.NewActivityID(uint64(time.Now().UnixNano()), id)
}

// CancelOrder cancel an order on centralized exchanges
func (rc ReserveCore) CancelOrder(orderID string, exchange common.Exchange) error {
	activity, err := rc.activityStorage.GetActivity(exchange.ID(), orderID)
	if err != nil {
		return err
	}
	if activity.Action != common.ActionTrade {
		return errors.New("this is not an order activity so cannot cancel")
	}
	base := activity.Params.Base
	quote := activity.Params.Quote
	return exchange.CancelOrder(orderID, base, quote)
}

// Trade token on centralized exchange
func (rc ReserveCore) Trade(
	exchange common.Exchange,
	tradeType string,
	pair commonv3.TradingPairSymbols,
	rate float64,
	amount float64) (common.ActivityID, float64, float64, bool, error) {
	var err error

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

		return rc.activityStorage.Record(
			common.ActionTrade,
			uid,
			exchange.ID().String(),
			common.ActivityParams{
				Exchange:  exchange.ID(),
				Type:      tradeType,
				Base:      pair.BaseSymbol,
				Quote:     pair.QuoteSymbol,
				Rate:      rate,
				Amount:    amount,
				Timepoint: timepoint,
			},
			activityResult,
			status,
			"",
			timepoint,
		)
	}

	if err = sanityCheckTrading(pair, rate, amount); err != nil {
		if sErr := recordActivity("", statusFailed, 0, 0, false, err); sErr != nil {
			rc.l.Warnw("failed to save activity record", "err", sErr)
			return common.ActivityID{}, 0, 0, false, common.CombineActivityStorageErrs(err, sErr)
		}
		return common.ActivityID{}, 0, 0, false, err
	}

	id, done, remaining, finished, err := exchange.Trade(tradeType, pair, rate, amount)
	uid := timebasedID(id)
	if err != nil {
		if sErr := recordActivity(id, statusFailed, done, remaining, finished, err); sErr != nil {
			rc.l.Warnw("failed to save activity record", "err", sErr)
			return uid, done, remaining, finished, common.CombineActivityStorageErrs(err, sErr)
		}
		return uid, done, remaining, finished, err
	}

	var status string
	if finished {
		status = statusDone
	} else {
		status = statusSubmitted
	}

	sErr := recordActivity(id, status, done, remaining, finished, nil)
	return uid, done, remaining, finished, common.CombineActivityStorageErrs(err, sErr)
}

// Deposit deposit token into centralized exchange
func (rc ReserveCore) Deposit(
	exchange common.Exchange,
	asset commonv3.Asset,
	amount *big.Int,
	timepoint uint64) (common.ActivityID, error) {
	address, supported := exchange.Address(asset)
	var (
		err         error
		ok          bool
		tx          *types.Transaction
		amountFloat = common.BigToFloat(amount, int64(asset.Decimals))
	)

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
			"",
			status,
			timepoint,
		)
	}

	if !supported {
		err = fmt.Errorf("exchange %s doesn't support token %s", exchange.ID().String(), asset.Symbol)
		sErr := recordActivity(statusFailed, "", 0, "", err)
		if sErr != nil {
			rc.l.Warnw("failed to save activity record", "err", sErr)
		}
		return common.ActivityID{}, common.CombineActivityStorageErrs(err, sErr)
	}

	if ok, err = rc.activityStorage.HasPendingDeposit(asset, exchange); err != nil {
		sErr := recordActivity(statusFailed, "", 0, "", err)
		if sErr != nil {
			rc.l.Warnw("failed to save activity record", "err", sErr)
		}
		return common.ActivityID{}, common.CombineActivityStorageErrs(err, sErr)
	}
	if ok {
		err = fmt.Errorf("there is a pending %s deposit to %s currently, please try again", asset.Symbol, exchange.ID().String())
		sErr := recordActivity(statusFailed, "", 0, "", err)
		if sErr != nil {
			rc.l.Warnw("failed to save activity record", "err", sErr)
		}
		return common.ActivityID{}, common.CombineActivityStorageErrs(err, sErr)
	}

	if err = sanityCheckAmount(exchange, asset, amount); err != nil {
		sErr := recordActivity(statusFailed, "", 0, "", err)
		if sErr != nil {
			rc.l.Warnw("failed to save activity record", "err", sErr)
		}
		return common.ActivityID{}, common.CombineActivityStorageErrs(err, sErr)
	}
	if tx, err = rc.blockchain.Send(asset, amount, address); err != nil {
		sErr := recordActivity(statusFailed, "", 0, "", err)
		if sErr != nil {
			rc.l.Warnw("failed to save activity record", "err", sErr)
		}
		return common.ActivityID{}, common.CombineActivityStorageErrs(err, sErr)
	}

	sErr := recordActivity(
		statusSubmitted,
		tx.Hash().Hex(),
		tx.Nonce(),
		tx.GasPrice().Text(10),
		nil,
	)
	return uidGenerator(tx.Hash().Hex()), common.CombineActivityStorageErrs(err, sErr)
}

// Withdraw token from exchange
func (rc ReserveCore) Withdraw(exchange common.Exchange, asset commonv3.Asset, amount *big.Int) (common.ActivityID, error) {
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
			"",
			timepoint,
		)
	}

	_, supported := exchange.Address(asset)
	if !supported {
		err = fmt.Errorf("exchange %s doesn't support asset %d", exchange.ID().String(), asset.ID)
		sErr := activityRecord("", statusFailed, err)
		if sErr != nil {
			rc.l.Warnw("failed to store activity record", "err", sErr)
		}
		return common.ActivityID{}, common.CombineActivityStorageErrs(err, sErr)
	}

	if err = sanityCheckAmount(exchange, asset, amount); err != nil {
		sErr := activityRecord("", statusFailed, err)
		if sErr != nil {
			rc.l.Warnw("failed to store activity record", "err", sErr)
		}
		return common.ActivityID{}, common.CombineActivityStorageErrs(err, sErr)
	}

	reserveAddr := rc.addressConf.Reserve

	id, err := exchange.Withdraw(asset, amount, reserveAddr)
	if err != nil {
		sErr := activityRecord("", statusFailed, err)
		if sErr != nil {
			rc.l.Warnw("failed to store activity record", "err", sErr)
		}
		return common.ActivityID{}, common.CombineActivityStorageErrs(err, sErr)
	}

	sErr := activityRecord(id, statusSubmitted, nil)
	return timebasedID(id), common.CombineActivityStorageErrs(err, sErr)
}

func calculateNewGasPrice(initPrice *big.Int, count uint64) *big.Int {
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

// return: old nonce, init price, step, error
func (rc ReserveCore) pendingSetrateInfo(minedNonce uint64) (*big.Int, *big.Int, uint64, error) {
	act, count, err := rc.activityStorage.PendingSetRate(minedNonce)
	if err != nil {
		return nil, nil, 0, err
	}
	if act == nil {
		return nil, nil, 0, nil
	}
	gasPriceStr := act.Result.GasPrice

	gasPrice, err := strconv.ParseUint(gasPriceStr, 10, 64)
	if err != nil {
		return nil, nil, count, err
	}
	return big.NewInt(int64(act.Result.Nonce)), big.NewInt(int64(gasPrice)), count, nil
}

// GetSetRateResult return result of set rate action
func (rc ReserveCore) GetSetRateResult(tokens []commonv3.Asset,
	buys, sells, afpMids []*big.Int,
	block *big.Int) (*types.Transaction, error) {
	var (
		tx  *types.Transaction
		err error
	)
	if len(tokens) != len(buys) {
		return tx, fmt.Errorf("number of buys (%d) is not equal to number of tokens (%d)", len(buys), len(tokens))
	}
	if len(tokens) != len(sells) {
		return tx, fmt.Errorf("number of sell (%d) is not equal to number of tokens (%d)", len(sells), len(tokens))

	}
	if len(tokens) != len(afpMids) {
		return tx, fmt.Errorf("number of afpMids (%d) is not equal to number of tokens (%d)", len(afpMids), len(tokens))
	}
	if err = sanityCheck(buys, afpMids, sells, rc.l); err != nil {
		return tx, err
	}
	var tokenAddrs []ethereum.Address
	for _, token := range tokens {
		tokenAddrs = append(tokenAddrs, token.Address)
	}
	// if there is a pending set rate tx, we replace it
	var (
		oldNonce   *big.Int
		initPrice  *big.Int
		minedNonce uint64
		count      uint64
	)
	minedNonce, err = rc.blockchain.SetRateMinedNonce()
	if err != nil {
		return tx, fmt.Errorf("couldn't get mined nonce of set rate operator (%s)", err.Error())
	}
	oldNonce, initPrice, count, err = rc.pendingSetrateInfo(minedNonce)
	rc.l.Infof("old nonce: %v, init price: %v, count: %d, err: %v", oldNonce, initPrice, count, err)
	if err != nil {
		return tx, fmt.Errorf("couldn't check pending set rate tx pool (%s). Please try later", err.Error())
	}
	if oldNonce != nil {
		newPrice := calculateNewGasPrice(initPrice, count)
		tx, err = rc.blockchain.SetRates(
			tokenAddrs, buys, sells, block,
			oldNonce,
			newPrice,
		)
		if err != nil {
			rc.l.Warnw("Trying to replace old tx failed", "err", err)
		} else {
			rc.l.Infof("Trying to replace old tx with new price: %s, tx: %s, init price: %s, count: %d",
				newPrice.String(), tx.Hash().Hex(), initPrice.String(), count)
		}
	} else {
		recommendedPrice := rc.blockchain.StandardGasPrice()
		var initPrice *big.Int
		if recommendedPrice == 0 || recommendedPrice > highBoundGasPrice {
			initPrice = common.GweiToWei(10)
		} else {
			initPrice = common.GweiToWei(recommendedPrice)
		}
		rc.l.Infof("initial set rate tx, init price: %s", initPrice.String())
		tx, err = rc.blockchain.SetRates(
			tokenAddrs, buys, sells, block,
			big.NewInt(int64(minedNonce)),
			initPrice,
		)
	}
	return tx, err
}

// SetRates to reserve
func (rc ReserveCore) SetRates(
	assets []commonv3.Asset,
	buys []*big.Int,
	sells []*big.Int,
	block *big.Int,
	afpMids []*big.Int,
	additionalMsgs []string) (common.ActivityID, error) {

	var (
		tx           *types.Transaction
		txhex        = ethereum.Hash{}.Hex()
		txnonce      = uint64(0)
		txprice      = "0"
		err          error
		miningStatus string
	)

	tx, err = rc.GetSetRateResult(assets, buys, sells, afpMids, block)
	if err != nil {
		miningStatus = common.MiningStatusFailed
	} else {
		miningStatus = common.MiningStatusSubmitted
		txhex = tx.Hash().Hex()
		txnonce = tx.Nonce()
		txprice = tx.GasPrice().Text(10)
	}
	uid := timebasedID(txhex)
	assetsID := []uint64{}
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
			Assets: assetsID,
			Buys:   buys,
			Sells:  sells,
			Block:  block,
			AFPMid: afpMids,
			Msgs:   additionalMsgs,
		},
		activityResult,
		"",
		miningStatus,
		common.NowInMillis(),
	)
	rc.l.Infof(
		"Core ----------> Set rates: ==> Result: tx: %s, nonce: %d, price: %s, error: %v, storage error: %v",
		txhex, txnonce, txprice, err, sErr,
	)

	return uid, common.CombineActivityStorageErrs(err, sErr)
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
			aMFloat := big.NewFloat(0).SetInt(afpMid[i])
			aMRate := calculateRate(aMFloat, eth)
			if bRate.Cmp(sRate) <= 0 || bRate.Cmp(aMRate) <= 0 {
				return errors.New("buy price must be bigger than sell price and afpMid price")
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
		if common.ExchangeID(exchg.ExchangeID).String() == exchange.ID().String() {
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
