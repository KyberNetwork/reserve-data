package core

import (
	"errors"
	"fmt"
	"log"
	"math"
	"math/big"
	"strconv"
	"time"

	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/settings"
)

const (
	// highBoundGasPrice is the price we will try to use to get higher priority
	// than trade tx to avoid price front running from users.
	highBoundGasPrice float64 = 100.1

	statusFailed    = "failed"
	statusSubmitted = "submitted"
	statusDone      = "done"
)

type ReserveCore struct {
	blockchain      Blockchain
	activityStorage ActivityStorage
	setting         Setting
}

func NewReserveCore(
	blockchain Blockchain,
	storage ActivityStorage,
	setting Setting) *ReserveCore {
	return &ReserveCore{
		blockchain,
		storage,
		setting,
	}
}

func timebasedID(id string) common.ActivityID {
	return common.NewActivityID(uint64(time.Now().UnixNano()), id)
}

func (rc ReserveCore) CancelOrder(id common.ActivityID, exchange common.Exchange) error {
	activity, err := rc.activityStorage.GetActivity(id)
	if err != nil {
		return err
	}
	if activity.Action != common.ActionTrade {
		return errors.New("this is not an order activity so cannot cancel")
	}
	base, ok := activity.Params["base"].(string)
	if !ok {
		return fmt.Errorf("cannot convert params base (value: %v) to tokenID (type string)", activity.Params["base"])
	}
	quote, ok := activity.Params["quote"].(string)
	if !ok {
		return fmt.Errorf("cannot convert params quote (value: %v) to tokenID (type string)", activity.Params["quote"])
	}
	orderID := id.EID
	return exchange.CancelOrder(orderID, base, quote)
}

func (rc ReserveCore) Trade(
	exchange common.Exchange,
	tradeType string,
	base common.Token,
	quote common.Token,
	rate float64,
	amount float64,
	timepoint uint64) (common.ActivityID, float64, float64, bool, error) {
	var err error

	recordActivity := func(id, status string, done, remaining float64, finished bool, err error) error {
		uid := timebasedID(id)
		log.Printf(
			"Core ----------> %s on %s: base: %s, quote: %s, rate: %s, amount: %s, timestamp: %d ==> Result: id: %s, done: %s, remaining: %s, finished: %t, error: %s",
			tradeType, exchange.ID(), base.ID, quote.ID,
			strconv.FormatFloat(rate, 'f', -1, 64),
			strconv.FormatFloat(amount, 'f', -1, 64), timepoint,
			uid,
			strconv.FormatFloat(done, 'f', -1, 64),
			strconv.FormatFloat(remaining, 'f', -1, 64),
			finished, common.ErrorToString(err),
		)

		return rc.activityStorage.Record(
			common.ActionTrade,
			uid,
			string(exchange.ID()),
			map[string]interface{}{
				"exchange":  exchange,
				"type":      tradeType,
				"base":      base,
				"quote":     quote,
				"rate":      rate,
				"amount":    strconv.FormatFloat(amount, 'f', -1, 64),
				"timepoint": timepoint,
			}, map[string]interface{}{
				"id":        id,
				"done":      done,
				"remaining": remaining,
				"finished":  finished,
				"error":     common.ErrorToString(err),
			},
			status,
			"",
			timepoint,
		)
	}

	if err = sanityCheckTrading(exchange, base, quote, rate, amount); err != nil {
		if sErr := recordActivity("", statusFailed, 0, 0, false, err); sErr != nil {
			log.Printf("failed to save activity record: %s", sErr)
			return common.ActivityID{}, 0, 0, false, common.CombineActivityStorageErrs(err, sErr)
		}
		return common.ActivityID{}, 0, 0, false, err
	}

	id, done, remaining, finished, err := exchange.Trade(tradeType, base, quote, rate, amount, timepoint)
	uid := timebasedID(id)
	if err != nil {
		if sErr := recordActivity(id, statusFailed, done, remaining, finished, err); sErr != nil {
			log.Printf("failed to save activity record: %s", sErr)
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

func (rc ReserveCore) Deposit(
	exchange common.Exchange,
	token common.Token,
	amount *big.Int,
	timepoint uint64) (common.ActivityID, error) {
	address, supported := exchange.Address(token)
	var (
		err         error
		ok          bool
		tx          *types.Transaction
		amountFloat = common.BigToFloat(amount, token.Decimals)
	)

	uidGenerator := func(txhex string) common.ActivityID {
		return timebasedID(txhex + "|" + token.ID + "|" + strconv.FormatFloat(amountFloat, 'f', -1, 64))
	}
	recordActivity := func(status, txhex, txnonce, txprice string, err error) error {
		uid := uidGenerator(txhex)
		log.Printf(
			"Core ----------> Deposit to %s: token: %s, amount: %s, timestamp: %d ==> Result: tx: %s, error: %s",
			exchange.ID(), token.ID, amount.Text(10), timepoint, txhex, common.ErrorToString(err),
		)

		return rc.activityStorage.Record(
			common.ActionDeposit,
			uid,
			string(exchange.ID()),
			map[string]interface{}{
				"exchange":  exchange,
				"token":     token,
				"amount":    strconv.FormatFloat(amountFloat, 'f', -1, 64),
				"timepoint": timepoint,
			}, map[string]interface{}{
				"tx":       txhex,
				"nonce":    txnonce,
				"gasPrice": txprice,
				"error":    common.ErrorToString(err),
			},
			"",
			status,
			timepoint,
		)
	}

	if !supported {
		err = fmt.Errorf("exchange %s doesn't support token %s", exchange.ID(), token.ID)
		sErr := recordActivity(statusFailed, "", "", "", err)
		if sErr != nil {
			log.Printf("failed to save activity record: %s", sErr)
		}
		return common.ActivityID{}, common.CombineActivityStorageErrs(err, sErr)
	}

	if ok, err = rc.activityStorage.HasPendingDeposit(token, exchange); err != nil {
		sErr := recordActivity(statusFailed, "", "", "", err)
		if sErr != nil {
			log.Printf("failed to save activity record: %s", sErr)
		}
		return common.ActivityID{}, common.CombineActivityStorageErrs(err, sErr)
	}
	if ok {
		err = fmt.Errorf("there is a pending %s deposit to %s currently, please try again", token.ID, exchange.ID())
		sErr := recordActivity(statusFailed, "", "", "", err)
		if sErr != nil {
			log.Printf("failed to save activity record: %s", sErr)
		}
		return common.ActivityID{}, common.CombineActivityStorageErrs(err, sErr)
	}

	if err = sanityCheckAmount(exchange, token, amount); err != nil {
		sErr := recordActivity(statusFailed, "", "", "", err)
		if sErr != nil {
			log.Printf("failed to save activity record: %s", sErr)
		}
		return common.ActivityID{}, common.CombineActivityStorageErrs(err, sErr)
	}
	if tx, err = rc.blockchain.Send(token, amount, address); err != nil {
		sErr := recordActivity(statusFailed, "", "", "", err)
		if sErr != nil {
			log.Printf("failed to save activity record: %s", sErr)
		}
		return common.ActivityID{}, common.CombineActivityStorageErrs(err, sErr)
	}

	sErr := recordActivity(
		statusSubmitted,
		tx.Hash().Hex(),
		strconv.FormatUint(tx.Nonce(), 10),
		tx.GasPrice().Text(10),
		nil,
	)
	return uidGenerator(tx.Hash().Hex()), common.CombineActivityStorageErrs(err, sErr)
}

func (rc ReserveCore) Withdraw(
	exchange common.Exchange, token common.Token,
	amount *big.Int, timepoint uint64) (common.ActivityID, error) {
	var err error

	activityRecord := func(id, status string, err error) error {
		uid := timebasedID(id)
		log.Printf(
			"Core ----------> Withdraw from %s: token: %s, amount: %s, timestamp: %d ==> Result: id: %s, error: %s",
			exchange.ID(), token.ID, amount.Text(10), timepoint, id, err,
		)
		return rc.activityStorage.Record(
			common.ActionWithdraw,
			uid,
			string(exchange.ID()),
			map[string]interface{}{
				"exchange":  exchange,
				"token":     token,
				"amount":    strconv.FormatFloat(common.BigToFloat(amount, token.Decimals), 'f', -1, 64),
				"timepoint": timepoint,
			}, map[string]interface{}{
				"error": common.ErrorToString(err),
				"id":    id,
				// this field will be updated with real tx when data fetcher can fetch it
				// from exchanges
				"tx": "",
			},
			status,
			"",
			timepoint,
		)
	}

	_, supported := exchange.Address(token)
	if !supported {
		err = fmt.Errorf("exchange %s doesn't support token %s", exchange.ID(), token.ID)
		sErr := activityRecord("", statusFailed, err)
		if sErr != nil {
			log.Printf("failed to store activiry record: %s", sErr.Error())
		}
		return common.ActivityID{}, common.CombineActivityStorageErrs(err, sErr)
	}

	if err = sanityCheckAmount(exchange, token, amount); err != nil {
		sErr := activityRecord("", statusFailed, err)
		if sErr != nil {
			log.Printf("failed to store activiry record: %s", sErr.Error())
		}
		return common.ActivityID{}, common.CombineActivityStorageErrs(err, sErr)
	}

	reserveAddr, err := rc.setting.GetAddress(settings.Reserve)
	if err != nil {
		sErr := activityRecord("", statusFailed, err)
		if sErr != nil {
			log.Printf("failed to store activiry record: %s", sErr.Error())
		}
		return common.ActivityID{}, common.CombineActivityStorageErrs(err, sErr)
	}

	id, err := exchange.Withdraw(token, amount, reserveAddr, timepoint)
	if err != nil {
		sErr := activityRecord("", statusFailed, err)
		if sErr != nil {
			log.Printf("failed to store activiry record: %s", sErr.Error())
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
	nonceStr, ok := act.Result["nonce"].(string)
	if !ok {
		nErr := fmt.Errorf("cannot convert result[nonce] (value %v) to string type", act.Result["nonce"])
		return nil, nil, count, nErr
	}
	gasPriceStr, ok := act.Result["gasPrice"].(string)
	if !ok {
		nErr := fmt.Errorf("cannot convert result[gasPrice] (value %v) to string type", act.Result["gasPrice"])
		return nil, nil, count, nErr
	}
	nonce, err := strconv.ParseUint(nonceStr, 10, 64)
	if err != nil {
		return nil, nil, count, err
	}
	gasPrice, err := strconv.ParseUint(gasPriceStr, 10, 64)
	if err != nil {
		return nil, nil, count, err
	}
	return big.NewInt(int64(nonce)), big.NewInt(int64(gasPrice)), count, nil
}

func (rc ReserveCore) GetSetRateResult(tokens []common.Token,
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
	if err = sanityCheck(buys, sells); err != nil {
		return tx, err
	}
	var tokenAddrs []ethereum.Address
	for _, token := range tokens {
		tokenAddrs = append(tokenAddrs, ethereum.HexToAddress(token.Address))
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
	log.Printf("old nonce: %v, init price: %v, count: %d, err: %s", oldNonce, initPrice, count, common.ErrorToString(err))
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
			log.Printf("Trying to replace old tx failed, err: %s", err)
		} else {
			log.Printf("Trying to replace old tx with new price: %s, tx: %s, init price: %s, count: %d",
				newPrice.String(),
				tx.Hash().Hex(),
				initPrice.String(),
				count,
			)
		}
	} else {
		recommendedPrice := rc.blockchain.StandardGasPrice()
		var initPrice *big.Int
		if recommendedPrice == 0 || recommendedPrice > highBoundGasPrice {
			initPrice = common.GweiToWei(10)
		} else {
			initPrice = common.GweiToWei(recommendedPrice)
		}
		log.Printf("initial set rate tx, init price: %s", initPrice.String())
		tx, err = rc.blockchain.SetRates(
			tokenAddrs, buys, sells, block,
			big.NewInt(int64(minedNonce)),
			initPrice,
		)
	}
	return tx, err
}

func (rc ReserveCore) SetRates(
	tokens []common.Token,
	buys []*big.Int,
	sells []*big.Int,
	block *big.Int,
	afpMids []*big.Int,
	additionalMsgs []string) (common.ActivityID, error) {

	var (
		tx           *types.Transaction
		txhex        = ethereum.Hash{}.Hex()
		txnonce      = "0"
		txprice      = "0"
		err          error
		miningStatus string
	)

	tx, err = rc.GetSetRateResult(tokens, buys, sells, afpMids, block)
	if err != nil {
		miningStatus = common.MiningStatusFailed
	} else {
		miningStatus = common.MiningStatusSubmitted
		txhex = tx.Hash().Hex()
		txnonce = strconv.FormatUint(tx.Nonce(), 10)
		txprice = tx.GasPrice().Text(10)
	}
	uid := timebasedID(txhex)
	sErr := rc.activityStorage.Record(
		common.ActionSetRate,
		uid,
		"blockchain",
		map[string]interface{}{
			"tokens": tokens,
			"buys":   buys,
			"sells":  sells,
			"block":  block,
			"afpMid": afpMids,
			"msgs":   additionalMsgs,
		}, map[string]interface{}{
			"tx":       txhex,
			"nonce":    txnonce,
			"gasPrice": txprice,
			"error":    common.ErrorToString(err),
		},
		"",
		miningStatus,
		common.GetTimepoint(),
	)
	log.Printf(
		"Core ----------> Set rates: ==> Result: tx: %s, nonce: %s, price: %s, error: %s, storage error: %s",
		txhex, txnonce, txprice, common.ErrorToString(err), common.ErrorToString(sErr),
	)

	return uid, common.CombineActivityStorageErrs(err, sErr)
}

func sanityCheck(buys, sells []*big.Int) error {
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
			if bRate.Cmp(sRate) < 0 {
				return errors.New("buy price must be bigger than sell price")
			}
		case 0: // both buy/sell rate is 0
			return nil
		case -1: // either buy/sell rate is 0
			if buys[i].Cmp(big.NewInt(0)) == 0 {
				return errors.New("buy rate can not be zero")
			}
			log.Printf("WARNING: sell rate is zero, index: %d, buy rate: %s", i, buys[i].String())
		}
	}
	return nil
}

func sanityCheckTrading(exchange common.Exchange, base, quote common.Token, rate, amount float64) error {
	tokenPair := makeTokenPair(base, quote)
	exchangeInfo, err := exchange.GetExchangeInfo(tokenPair.PairID())
	if err != nil {
		return err
	}
	currentNotional := rate * amount
	minNotional := exchangeInfo.MinNotional
	if minNotional != float64(0) {
		if currentNotional < minNotional {
			return errors.New("notional must be bigger than exchange's MinNotional")
		}
	}
	return nil
}

func sanityCheckAmount(exchange common.Exchange, token common.Token, amount *big.Int) error {
	exchangeFee, err := exchange.GetFee()
	if err != nil {
		return err
	}
	amountFloat := big.NewFloat(0).SetInt(amount)
	feeWithdrawing := exchangeFee.Funding.GetTokenFee(token.ID)
	expDecimal := big.NewInt(0).Exp(big.NewInt(10), big.NewInt(token.Decimals), nil)
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

func makeTokenPair(base, quote common.Token) common.TokenPair {
	if base.ID == "ETH" {
		return common.NewTokenPair(quote, base)
	}
	return common.NewTokenPair(base, quote)
}
