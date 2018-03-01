package core

import (
	"errors"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"time"

	"github.com/KyberNetwork/reserve-data/common"
	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type ReserveCore struct {
	blockchain      Blockchain
	activityStorage ActivityStorage
	rm              ethereum.Address
	imt             ethereum.Address
}

func NewReserveCore(
	blockchain Blockchain,
	storage ActivityStorage,
	rm ethereum.Address,
	imt ethereum.Address) *ReserveCore {
	return &ReserveCore{
		blockchain,
		storage,
		rm,
		imt,
	}
}

func timebasedID(id string) common.ActivityID {
	return common.NewActivityID(uint64(time.Now().UnixNano()), id)
}

func (self ReserveCore) CancelOrder(id common.ActivityID, exchange common.Exchange) error {
	return exchange.CancelOrder(id)
}

func (self ReserveCore) GetAddresses() *common.Addresses {
	return self.blockchain.GetAddresses()
}

func (self ReserveCore) Trade(
	exchange common.Exchange,
	tradeType string,
	base common.Token,
	quote common.Token,
	rate float64,
	amount float64,
	timepoint uint64) (common.ActivityID, float64, float64, bool, error) {

	var id string
	var done, remaining float64
	var finished bool
	var err error

	err = sanityCheckTrading(exchange, base, quote, rate, amount)
	if err == nil {
		id, done, remaining, finished, err = exchange.Trade(tradeType, base, quote, rate, amount, timepoint)
	}

	var status string
	if err != nil {
		status = "failed"
	} else {
		if finished {
			status = "done"
		} else {
			status = "submitted"
		}
	}
	uid := timebasedID(id)
	self.activityStorage.Record(
		"trade",
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
			"error":     err,
		},
		status,
		"",
		timepoint,
	)
	log.Printf(
		"Core ----------> %s on %s: base: %s, quote: %s, rate: %s, amount: %s, timestamp: %d ==> Result: id: %s, done: %s, remaining: %s, finished: %t, error: %s",
		tradeType, exchange.ID(), base.ID, quote.ID,
		strconv.FormatFloat(rate, 'f', -1, 64),
		strconv.FormatFloat(amount, 'f', -1, 64), timepoint,
		uid,
		strconv.FormatFloat(done, 'f', -1, 64),
		strconv.FormatFloat(remaining, 'f', -1, 64),
		finished, err,
	)
	return uid, done, remaining, finished, err
}

func (self ReserveCore) Deposit(
	exchange common.Exchange,
	token common.Token,
	amount *big.Int,
	timepoint uint64) (common.ActivityID, error) {

	address, supported := exchange.Address(token)

	var tx *types.Transaction
	var txhex string = ethereum.Hash{}.Hex()
	var txnonce string = "0"
	var txprice string = "0"
	var err error
	var status string

	if !supported {
		err = errors.New(fmt.Sprintf("Exchange %s doesn't support token %s", exchange.ID(), token.ID))
	} else if self.activityStorage.HasPendingDeposit(token, exchange) {
		err = errors.New(fmt.Sprintf("There is a pending %s deposit to %s currently, please try again", token.ID, exchange.ID()))
	} else {
		err = sanityCheckAmount(exchange, token, amount)
		if err == nil {
			if exchange.GetImtorMode() {
				tx, err = self.blockchain.Send(token, amount, self.imt)
			} else {
				tx, err = self.blockchain.Send(token, amount, address)
			}
		}
	}
	if err != nil {
		status = "failed"
	} else {
		status = "submitted"
		txhex = tx.Hash().Hex()
		txnonce = strconv.FormatUint(tx.Nonce(), 10)
		txprice = tx.GasPrice().Text(10)
	}
	amountFloat := common.BigToFloat(amount, token.Decimal)
	uid := timebasedID(txhex + "|" + token.ID + "|" + strconv.FormatFloat(amountFloat, 'f', -1, 64))
	self.activityStorage.Record(
		"deposit",
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
			"error":    err,
		},
		"",
		status,
		timepoint,
	)
	log.Printf(
		"Core ----------> Deposit to %s: token: %s, amount: %s, timestamp: %d ==> Result: tx: %s, error: %s",
		exchange.ID(), token.ID, amount.Text(10), timepoint, txhex, err,
	)
	return uid, err
}

func (self ReserveCore) Withdraw(
	exchange common.Exchange, token common.Token,
	amount *big.Int, timepoint uint64) (common.ActivityID, error) {

	_, supported := exchange.Address(token)
	var err error
	var id string
	if !supported {
		err = errors.New(fmt.Sprintf("Exchange %s doesn't support token %s", exchange.ID(), token.ID))
	} else {
		err = sanityCheckAmount(exchange, token, amount)
		if err == nil {
			id, err = exchange.Withdraw(token, amount, self.rm, timepoint)
		}
	}
	var status string
	if err != nil {
		status = "failed"
	} else {
		status = "submitted"
	}
	uid := timebasedID(id)
	self.activityStorage.Record(
		"withdraw",
		uid,
		string(exchange.ID()),
		map[string]interface{}{
			"exchange":  exchange,
			"token":     token,
			"amount":    strconv.FormatFloat(common.BigToFloat(amount, token.Decimal), 'f', -1, 64),
			"timepoint": timepoint,
		}, map[string]interface{}{
			"error": err,
			"id":    id,
			// this field will be updated with real tx when data fetcher can fetch it
			// from exchanges
			"tx": "",
		},
		status,
		"",
		timepoint,
	)
	log.Printf(
		"Core ----------> Withdraw from %s: token: %s, amount: %s, timestamp: %d ==> Result: id: %s, error: %s",
		exchange.ID(), token.ID, amount.Text(10), timepoint, id, err,
	)
	return uid, err
}

func (self ReserveCore) pendingSetrateInfo(minedNonce uint64) (*big.Int, *big.Int, error) {
	act, err := self.activityStorage.PendingSetrate(minedNonce)
	if err != nil {
		return nil, nil, err
	}
	if act != nil {
		nonce, _ := strconv.ParseUint(act.Result["nonce"].(string), 10, 64)
		gasPrice, _ := strconv.ParseUint(act.Result["gasPrice"].(string), 10, 64)
		return big.NewInt(int64(nonce)), big.NewInt(int64(gasPrice)), nil
	} else {
		return nil, nil, nil
	}
}

func (self ReserveCore) SetRates(
	tokens []common.Token,
	buys []*big.Int,
	sells []*big.Int,
	block *big.Int,
	afpMids []*big.Int) (common.ActivityID, error) {

	lentokens := len(tokens)
	lenbuys := len(buys)
	lensells := len(sells)
	lenafps := len(afpMids)

	var tx *types.Transaction
	var txhex string = ethereum.Hash{}.Hex()
	var txnonce string = "0"
	var txprice string = "0"
	var err error
	var status string

	if lentokens != lenbuys || lentokens != lensells || lentokens != lenafps {
		err = errors.New("Tokens, buys sells and afpMids must have the same length")
	} else {
		err = sanityCheck(buys, afpMids, sells)
		if err == nil {
			tokenAddrs := []ethereum.Address{}
			for _, token := range tokens {
				tokenAddrs = append(tokenAddrs, ethereum.HexToAddress(token.Address))
			}
			// if there is a pending set rate tx, we replace it
			var oldNonce *big.Int
			var oldPrice *big.Int
			var minedNonce uint64
			minedNonce, err = self.blockchain.SetRateMinedNonce()
			if err != nil {
				err = errors.New("Couldn't get mined nonce of set rate operator")
			} else {
				oldNonce, oldPrice, err = self.pendingSetrateInfo(minedNonce)
				if err != nil {
					err = errors.New("Couldn't check pending set rate tx pool. Please try later")
				} else {
					if oldNonce != nil {
						newPrice := big.NewInt(0).Add(oldPrice, big.NewInt(10000000000))
						log.Printf("Trying to replace old tx with new price: %s", newPrice.Text(10))
						tx, err = self.blockchain.SetRates(
							tokenAddrs, buys, sells, block,
							oldNonce,
							newPrice,
						)
					} else {
						tx, err = self.blockchain.SetRates(
							tokenAddrs, buys, sells, block,
							nil,
							big.NewInt(50100000000),
						)
					}
				}
			}
		}
	}
	if err != nil {
		status = "failed"
	} else {
		status = "submitted"
		txhex = tx.Hash().Hex()
		txnonce = strconv.FormatUint(tx.Nonce(), 10)
		txprice = tx.GasPrice().Text(10)
	}
	uid := timebasedID(txhex)
	self.activityStorage.Record(
		"set_rates",
		uid,
		"blockchain",
		map[string]interface{}{
			"tokens": tokens,
			"buys":   buys,
			"sells":  sells,
			"block":  block,
			"afpMid": afpMids,
		}, map[string]interface{}{
			"tx":       txhex,
			"nonce":    txnonce,
			"gasPrice": txprice,
			"error":    err,
		},
		"",
		status,
		common.GetTimepoint(),
	)
	log.Printf(
		"Core ----------> Set rates: ==> Result: tx: %s, error: %s",
		txhex, err,
	)
	return uid, err
}

func sanityCheck(buys, afpMid, sells []*big.Int) error {
	eth := big.NewFloat(0).SetInt(big.NewInt(1000000000000000000))
	for i, s := range sells {
		check := checkZeroValue(buys[i], s)
		switch check {
		case 1:
			sFloat := big.NewFloat(0).SetInt(s)
			sRate := calculateRate(sFloat, eth)
			bFloat := big.NewFloat(0).SetInt(buys[i])
			bRate := calculateRate(eth, bFloat)
			aMFloat := big.NewFloat(0).SetInt(afpMid[i])
			aMRate := calculateRate(aMFloat, eth)
			if bRate.Cmp(sRate) <= 0 || bRate.Cmp(aMRate) <= 0 {
				return errors.New("Sell price must be bigger than buy price and afpMid price")
			}
		case 0:
			return nil
		case -1:
			return errors.New("Rate cannot be zero on only sell or buy side")
		}
	}
	return nil
}

func sanityCheckTrading(exchange common.Exchange, base, quote common.Token, rate, amount float64) error {
	tokenPairID := makeTokenPair(base.ID, quote.ID)
	exchangeInfo, err := exchange.GetExchangeInfo(tokenPairID)
	if err != nil {
		return err
	}
	currentNotional := rate * amount
	minNotional := exchangeInfo.MinNotional
	if minNotional != float64(0) {
		if currentNotional < minNotional {
			return errors.New("Notional must be bigger than exchange's MinNotional")
		}
	}
	return nil
}

func sanityCheckAmount(exchange common.Exchange, token common.Token, amount *big.Int) error {
	exchangeFee := exchange.GetFee()
	amountFloat := big.NewFloat(0).SetInt(amount)
	feeWithdrawing := exchangeFee.Funding.GetTokenFee(string(token.ID))
	expDecimal := big.NewInt(0).Exp(big.NewInt(10), big.NewInt(token.Decimal), nil)
	minAmountWithdraw := big.NewFloat(0)

	minAmountWithdraw.Mul(big.NewFloat(feeWithdrawing), big.NewFloat(0).SetInt(expDecimal))
	if amountFloat.Cmp(minAmountWithdraw) < 0 {
		return errors.New("Amount is too small!!!")
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

func makeTokenPair(base, quote string) common.TokenPairID {
	if base == "ETH" {
		return common.NewTokenPairID(quote, base)
	}
	return common.NewTokenPairID(base, quote)
}
