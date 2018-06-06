package http

import (
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/KyberNetwork/reserve-data"
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/http/httputil"
	"github.com/KyberNetwork/reserve-data/metric"
	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	raven "github.com/getsentry/raven-go"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sentry"
	"github.com/gin-gonic/gin"
)

const (
	MAX_TIMESPOT   uint64 = 18446744073709551615
	MAX_DATA_SIZE  int    = 1000000 //1 Megabyte in byte
	START_TIMEZONE int64  = -11
	END_TIMEZONE   int64  = 14
)

var (
	// errDataSizeExceed is returned when the post data is larger than MAX_DATA_SIZE.
	errDataSizeExceed = errors.New("the data size must be less than 1 MB")
)

type HTTPServer struct {
	app         reserve.ReserveData
	core        reserve.ReserveCore
	stat        reserve.ReserveStats
	metric      metric.MetricStorage
	host        string
	authEnabled bool
	auth        Authentication
	r           *gin.Engine
}

func getTimePoint(c *gin.Context, useDefault bool) uint64 {
	timestamp := c.DefaultQuery("timestamp", "")
	if timestamp == "" {
		if useDefault {
			log.Printf("Interpreted timestamp to default - %d\n", MAX_TIMESPOT)
			return MAX_TIMESPOT
		} else {
			timepoint := common.GetTimepoint()
			log.Printf("Interpreted timestamp to current time - %d\n", timepoint)
			return uint64(timepoint)
		}
	} else {
		timepoint, err := strconv.ParseUint(timestamp, 10, 64)
		if err != nil {
			log.Printf("Interpreted timestamp(%s) to default - %d", timestamp, MAX_TIMESPOT)
			return MAX_TIMESPOT
		} else {
			log.Printf("Interpreted timestamp(%s) to %d", timestamp, timepoint)
			return timepoint
		}
	}
}

func IsIntime(nonce string) bool {
	serverTime := common.GetTimepoint()
	log.Printf("Server time: %d, None: %s", serverTime, nonce)
	nonceInt, err := strconv.ParseInt(nonce, 10, 64)
	if err != nil {
		log.Printf("IsIntime returns false, err: %v", err)
		return false
	}
	difference := nonceInt - int64(serverTime)
	if difference < -30000 || difference > 30000 {
		log.Printf("IsIntime returns false, nonce: %d, serverTime: %d, difference: %d", nonceInt, int64(serverTime), difference)
		return false
	}
	return true
}

func eligible(ups, allowedPerms []Permission) bool {
	for _, up := range ups {
		for _, ap := range allowedPerms {
			if up == ap {
				return true
			}
		}
	}
	return false
}

// Authenticated signed message (message = url encoded both query params and post params, keys are sorted) in "signed" header
// using HMAC512
// params must contain "nonce" which is the unixtime in millisecond. The nonce will be invalid
// if it differs from server time more than 10s
func (self *HTTPServer) Authenticated(c *gin.Context, requiredParams []string, perms []Permission) (url.Values, bool) {
	err := c.Request.ParseForm()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("Malformed request package: %s", err.Error())))
		return c.Request.Form, false
	}

	if !self.authEnabled {
		return c.Request.Form, true
	}

	params := c.Request.Form
	log.Printf("Form params: %s\n", params)
	if !IsIntime(params.Get("nonce")) {
		httputil.ResponseFailure(c, httputil.WithReason("Your nonce is invalid"))
		return c.Request.Form, false
	}

	for _, p := range requiredParams {
		if params.Get(p) == "" {
			httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("Required param (%s) is missing. Param name is case sensitive", p)))
			return c.Request.Form, false
		}
	}

	signed := c.GetHeader("signed")
	message := c.Request.Form.Encode()
	userPerms := self.auth.GetPermission(signed, message)
	if eligible(userPerms, perms) {
		return params, true
	} else {
		if len(userPerms) == 0 {
			httputil.ResponseFailure(c, httputil.WithReason("Invalid signed token"))
		} else {
			httputil.ResponseFailure(c, httputil.WithReason("You don't have permission to proceed"))
		}
		return params, false
	}
}

func (self *HTTPServer) AllPricesVersion(c *gin.Context) {
	log.Printf("Getting all prices version")
	data, err := self.app.CurrentPriceVersion(getTimePoint(c, true))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithField("version", data))
	}
}

func (self *HTTPServer) AllPrices(c *gin.Context) {
	log.Printf("Getting all prices \n")
	data, err := self.app.GetAllPrices(getTimePoint(c, true))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithMultipleFields(gin.H{
			"version":   data.Version,
			"timestamp": data.Timestamp,
			"data":      data.Data,
			"block":     data.Block,
		}))
	}
}

func (self *HTTPServer) Price(c *gin.Context) {
	base := c.Param("base")
	quote := c.Param("quote")
	log.Printf("Getting price for %s - %s \n", base, quote)
	pair, err := common.NewTokenPair(base, quote)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithReason("Token pair is not supported"))
	} else {
		data, err := self.app.GetOnePrice(pair.PairID(), getTimePoint(c, true))
		if err != nil {
			httputil.ResponseFailure(c, httputil.WithError(err))
		} else {
			httputil.ResponseSuccess(c, httputil.WithMultipleFields(gin.H{
				"version":   data.Version,
				"timestamp": data.Timestamp,
				"exchanges": data.Data,
			}))
		}
	}
}

func (self *HTTPServer) AuthDataVersion(c *gin.Context) {
	log.Printf("Getting current auth data snapshot version")
	_, ok := self.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, RebalancePermission, ConfigurePermission, ConfirmConfPermission})
	if !ok {
		return
	}

	data, err := self.app.CurrentAuthDataVersion(getTimePoint(c, true))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithField("version", data))
	}
}

func (self *HTTPServer) AuthData(c *gin.Context) {
	log.Printf("Getting current auth data snapshot \n")
	_, ok := self.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, RebalancePermission, ConfigurePermission, ConfirmConfPermission})
	if !ok {
		return
	}

	data, err := self.app.GetAuthData(getTimePoint(c, true))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithMultipleFields(gin.H{
			"version":   data.Version,
			"timestamp": data.Timestamp,
			"data":      data.Data,
		}))
	}
}

func (self *HTTPServer) GetRates(c *gin.Context) {
	log.Printf("Getting all rates \n")
	fromTime, _ := strconv.ParseUint(c.Query("fromTime"), 10, 64)
	toTime, _ := strconv.ParseUint(c.Query("toTime"), 10, 64)
	if toTime == 0 {
		toTime = MAX_TIMESPOT
	}
	data, err := self.app.GetRates(fromTime, toTime)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithData(data))
	}
}

func (self *HTTPServer) GetRate(c *gin.Context) {
	log.Printf("Getting all rates \n")
	data, err := self.app.GetRate(getTimePoint(c, true))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithMultipleFields(gin.H{
			"version":   data.Version,
			"timestamp": data.Timestamp,
			"data":      data.Data,
		}))
	}
}

func (self *HTTPServer) SetRate(c *gin.Context) {
	postForm, ok := self.Authenticated(c, []string{"tokens", "buys", "sells", "block", "afp_mid", "msgs"}, []Permission{RebalancePermission})
	if !ok {
		return
	}
	tokenAddrs := postForm.Get("tokens")
	buys := postForm.Get("buys")
	sells := postForm.Get("sells")
	block := postForm.Get("block")
	afpMid := postForm.Get("afp_mid")
	msgs := strings.Split(postForm.Get("msgs"), "-")
	tokens := []common.Token{}
	for _, tok := range strings.Split(tokenAddrs, "-") {
		token, err := common.GetInternalToken(tok)
		if err != nil {
			httputil.ResponseFailure(c, httputil.WithError(err))
			return
		}
		tokens = append(tokens, token)
	}
	bigBuys := []*big.Int{}
	for _, rate := range strings.Split(buys, "-") {
		r, err := hexutil.DecodeBig(rate)
		if err != nil {
			httputil.ResponseFailure(c, httputil.WithError(err))
			return
		}
		bigBuys = append(bigBuys, r)
	}
	bigSells := []*big.Int{}
	for _, rate := range strings.Split(sells, "-") {
		r, err := hexutil.DecodeBig(rate)
		if err != nil {
			httputil.ResponseFailure(c, httputil.WithError(err))
			return
		}
		bigSells = append(bigSells, r)
	}
	intBlock, err := strconv.ParseInt(block, 10, 64)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	bigAfpMid := []*big.Int{}
	for _, rate := range strings.Split(afpMid, "-") {
		r, vErr := hexutil.DecodeBig(rate)
		if vErr != nil {
			httputil.ResponseFailure(c, httputil.WithError(vErr))
			return
		}
		bigAfpMid = append(bigAfpMid, r)
	}
	id, err := self.core.SetRates(tokens, bigBuys, bigSells, big.NewInt(intBlock), bigAfpMid, msgs)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithField("id", id))
}

func (self *HTTPServer) Trade(c *gin.Context) {
	postForm, ok := self.Authenticated(c, []string{"base", "quote", "amount", "rate", "type"}, []Permission{RebalancePermission})
	if !ok {
		return
	}

	exchangeParam := c.Param("exchangeid")
	baseTokenParam := postForm.Get("base")
	quoteTokenParam := postForm.Get("quote")
	amountParam := postForm.Get("amount")
	rateParam := postForm.Get("rate")
	typeParam := postForm.Get("type")

	exchange, err := common.GetExchange(exchangeParam)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	base, err := common.GetInternalToken(baseTokenParam)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	quote, err := common.GetInternalToken(quoteTokenParam)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	amount, err := strconv.ParseFloat(amountParam, 64)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	rate, err := strconv.ParseFloat(rateParam, 64)
	log.Printf("http server: Trade: rate: %f, raw rate: %s", rate, rateParam)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	if typeParam != "sell" && typeParam != "buy" {
		httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("Trade type of %s is not supported.", typeParam)))
		return
	}
	id, done, remaining, finished, err := self.core.Trade(
		exchange, typeParam, base, quote, rate, amount, getTimePoint(c, false))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithMultipleFields(gin.H{
		"id":        id,
		"done":      done,
		"remaining": remaining,
		"finished":  finished,
	}))
}

func (self *HTTPServer) CancelOrder(c *gin.Context) {
	postForm, ok := self.Authenticated(c, []string{"order_id"}, []Permission{RebalancePermission})
	if !ok {
		return
	}

	exchangeParam := c.Param("exchangeid")
	id := postForm.Get("order_id")

	exchange, err := common.GetExchange(exchangeParam)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	log.Printf("Cancel order id: %s from %s\n", id, exchange.ID())
	activityID, err := common.StringToActivityID(id)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	err = self.core.CancelOrder(activityID, exchange)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (self *HTTPServer) Withdraw(c *gin.Context) {
	postForm, ok := self.Authenticated(c, []string{"token", "amount"}, []Permission{RebalancePermission})
	if !ok {
		return
	}

	exchangeParam := c.Param("exchangeid")
	tokenParam := postForm.Get("token")
	amountParam := postForm.Get("amount")

	exchange, err := common.GetExchange(exchangeParam)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	token, err := common.GetInternalToken(tokenParam)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	amount, err := hexutil.DecodeBig(amountParam)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	log.Printf("Withdraw %s %s from %s\n", amount.Text(10), token.ID, exchange.ID())
	id, err := self.core.Withdraw(exchange, token, amount, getTimePoint(c, false))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithField("id", id))
}

func (self *HTTPServer) Deposit(c *gin.Context) {
	postForm, ok := self.Authenticated(c, []string{"amount", "token"}, []Permission{RebalancePermission})
	if !ok {
		return
	}

	exchangeParam := c.Param("exchangeid")
	amountParam := postForm.Get("amount")
	tokenParam := postForm.Get("token")

	exchange, err := common.GetExchange(exchangeParam)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	token, err := common.GetInternalToken(tokenParam)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	amount, err := hexutil.DecodeBig(amountParam)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	log.Printf("Depositing %s %s to %s\n", amount.Text(10), token.ID, exchange.ID())
	id, err := self.core.Deposit(exchange, token, amount, getTimePoint(c, false))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithField("id", id))
}

func (self *HTTPServer) GetActivities(c *gin.Context) {
	log.Printf("Getting all activity records \n")
	_, ok := self.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, RebalancePermission, ConfigurePermission, ConfirmConfPermission})
	if !ok {
		return
	}
	fromTime, _ := strconv.ParseUint(c.Query("fromTime"), 10, 64)
	toTime, _ := strconv.ParseUint(c.Query("toTime"), 10, 64)
	if toTime == 0 {
		toTime = common.GetTimepoint()
	}

	data, err := self.app.GetRecords(fromTime*1000000, toTime*1000000)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithData(data))
	}
}

func (self *HTTPServer) CatLogs(c *gin.Context) {
	log.Printf("Getting cat logs")
	fromTime, err := strconv.ParseUint(c.Query("fromTime"), 10, 64)
	if err != nil {
		fromTime = 0
	}
	toTime, err := strconv.ParseUint(c.Query("toTime"), 10, 64)
	if err != nil || toTime == 0 {
		toTime = common.GetTimepoint()
	}

	data, err := self.stat.GetCatLogs(fromTime, toTime)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithData(data))
	}
}

func (self *HTTPServer) TradeLogs(c *gin.Context) {
	log.Printf("Getting trade logs")
	fromTime, err := strconv.ParseUint(c.Query("fromTime"), 10, 64)
	if err != nil {
		fromTime = 0
	}
	toTime, err := strconv.ParseUint(c.Query("toTime"), 10, 64)
	if err != nil || toTime == 0 {
		toTime = common.GetTimepoint()
	}

	data, err := self.stat.GetTradeLogs(fromTime, toTime)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithData(data))
	}
}

func (self *HTTPServer) StopFetcher(c *gin.Context) {
	err := self.app.Stop()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c)
	}
}

func (self *HTTPServer) ImmediatePendingActivities(c *gin.Context) {
	log.Printf("Getting all immediate pending activity records \n")
	_, ok := self.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, RebalancePermission, ConfigurePermission, ConfirmConfPermission})
	if !ok {
		return
	}

	data, err := self.app.GetPendingActivities()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithData(data))
	}
}

func (self *HTTPServer) Metrics(c *gin.Context) {
	response := metric.MetricResponse{
		Timestamp: common.GetTimepoint(),
	}
	log.Printf("Getting metrics")
	postForm, ok := self.Authenticated(c, []string{"tokens", "from", "to"}, []Permission{ReadOnlyPermission, RebalancePermission, ConfigurePermission, ConfirmConfPermission})
	if !ok {
		return
	}
	tokenParam := postForm.Get("tokens")
	fromParam := postForm.Get("from")
	toParam := postForm.Get("to")
	tokens := []common.Token{}
	for _, tok := range strings.Split(tokenParam, "-") {
		token, err := common.GetInternalToken(tok)
		if err != nil {
			httputil.ResponseFailure(c, httputil.WithError(err))
			return
		}
		tokens = append(tokens, token)
	}
	from, err := strconv.ParseUint(fromParam, 10, 64)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	}
	to, err := strconv.ParseUint(toParam, 10, 64)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	}
	data, err := self.metric.GetMetric(tokens, from, to)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	}
	response.ReturnTime = common.GetTimepoint()
	response.Data = data
	httputil.ResponseSuccess(c, httputil.WithMultipleFields(gin.H{
		"timestamp":  response.Timestamp,
		"returnTime": response.ReturnTime,
		"data":       response.Data,
	}))
}

func (self *HTTPServer) StoreMetrics(c *gin.Context) {
	log.Printf("Storing metrics")
	postForm, ok := self.Authenticated(c, []string{"timestamp", "data"}, []Permission{RebalancePermission})
	if !ok {
		return
	}
	timestampParam := postForm.Get("timestamp")
	dataParam := postForm.Get("data")

	timestamp, err := strconv.ParseUint(timestampParam, 10, 64)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	}
	metricEntry := metric.MetricEntry{}
	metricEntry.Timestamp = timestamp
	metricEntry.Data = map[string]metric.TokenMetric{}
	// data must be in form of <token>_afpmid_spread|<token>_afpmid_spread|...
	for _, tokenData := range strings.Split(dataParam, "|") {
		parts := strings.Split(tokenData, "_")
		if len(parts) != 3 {
			httputil.ResponseFailure(c, httputil.WithReason("submitted data is not in correct format"))
			return
		}
		token := parts[0]
		afpmidStr := parts[1]
		spreadStr := parts[2]

		afpmid, vErr := strconv.ParseFloat(afpmidStr, 64)
		if vErr != nil {
			httputil.ResponseFailure(c, httputil.WithReason("Afp mid "+afpmidStr+" is not float64"))
			return
		}
		spread, vErr := strconv.ParseFloat(spreadStr, 64)
		if vErr != nil {
			httputil.ResponseFailure(c, httputil.WithReason("Spread "+spreadStr+" is not float64"))
			return
		}
		metricEntry.Data[token] = metric.TokenMetric{
			AfpMid: afpmid,
			Spread: spread,
		}
	}

	err = self.metric.StoreMetric(&metricEntry, common.GetTimepoint())
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c)
	}
}

func (self *HTTPServer) GetExchangeInfo(c *gin.Context) {
	exchangeParam := c.Query("exchangeid")
	if exchangeParam == "" {
		data := map[string]map[common.TokenPairID]common.ExchangePrecisionLimit{}
		for _, ex := range common.SupportedExchanges {
			exchangeInfo, err := ex.GetInfo()
			if err != nil {
				httputil.ResponseFailure(c, httputil.WithError(err))
				return
			}
			data[string(ex.ID())] = exchangeInfo.GetData()
		}
		httputil.ResponseSuccess(c, httputil.WithData(data))
	} else {
		exchange, err := common.GetExchange(exchangeParam)
		if err != nil {
			httputil.ResponseFailure(c, httputil.WithError(err))
			return
		}
		exchangeInfo, err := exchange.GetInfo()
		if err != nil {
			httputil.ResponseFailure(c, httputil.WithError(err))
			return
		}
		httputil.ResponseSuccess(c, httputil.WithData(exchangeInfo.GetData()))
	}
}

func (self *HTTPServer) GetPairInfo(c *gin.Context) {
	exchangeParam := c.Param("exchangeid")
	base := c.Param("base")
	quote := c.Param("quote")
	exchange, err := common.GetExchange(exchangeParam)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	pair, err := common.NewTokenPair(base, quote)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	pairInfo, err := exchange.GetExchangeInfo(pair.PairID())
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(pairInfo))
	return
}

func (self *HTTPServer) GetExchangeFee(c *gin.Context) {
	exchangeParam := c.Param("exchangeid")
	exchange, err := common.GetExchange(exchangeParam)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	fee := exchange.GetFee()
	httputil.ResponseSuccess(c, httputil.WithData(fee))
	return
}

func (self *HTTPServer) GetFee(c *gin.Context) {
	data := map[string]common.ExchangeFees{}
	for _, exchange := range common.SupportedExchanges {
		fee := exchange.GetFee()
		data[string(exchange.ID())] = fee
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
	return
}

func (self *HTTPServer) GetMinDeposit(c *gin.Context) {
	data := map[string]common.ExchangesMinDeposit{}
	for _, exchange := range common.SupportedExchanges {
		minDeposit := exchange.GetMinDeposit()
		data[string(exchange.ID())] = minDeposit
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
	return
}

func (self *HTTPServer) GetTargetQty(c *gin.Context) {
	log.Println("Getting target quantity")
	_, ok := self.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, RebalancePermission, ConfigurePermission, ConfirmConfPermission})
	if !ok {
		return
	}
	data, err := self.metric.GetTokenTargetQty()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithData(data))
	}
}

func (self *HTTPServer) GetPendingTargetQty(c *gin.Context) {
	log.Println("Getting pending target qty")
	_, ok := self.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, RebalancePermission, ConfigurePermission, ConfirmConfPermission})
	if !ok {
		return
	}
	data, err := self.metric.GetPendingTargetQty()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
	return
}

// func targetQtySanityCheck(total, reserve, rebalanceThresold, transferThresold float64) error {
// 	if total <= reserve {
// 		return errors.New("Total quantity must bigger than reserver quantity")
// 	}
// 	if rebalanceThresold < 0 || rebalanceThresold > 1 || transferThresold < 0 || transferThresold > 1 {
// 		return errors.New("Rebalance and transfer thresold must bigger than 0 and smaller than 1")
// 	}
// 	return nil
// }

func (self *HTTPServer) ConfirmTargetQty(c *gin.Context) {
	log.Println("Confirm target quantity")
	postForm, ok := self.Authenticated(c, []string{"data", "type"}, []Permission{ConfirmConfPermission})
	if !ok {
		return
	}
	data := postForm.Get("data")
	id := postForm.Get("id")
	err := self.metric.StoreTokenTargetQty(id, data)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
	return
}

func (self *HTTPServer) CancelTargetQty(c *gin.Context) {
	log.Println("Cancel target quantity")
	_, ok := self.Authenticated(c, []string{}, []Permission{ConfirmConfPermission})
	if !ok {
		return
	}
	err := self.metric.RemovePendingTargetQty()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
	return
}

func (self *HTTPServer) SetTargetQty(c *gin.Context) {
	postForm, ok := self.Authenticated(c, []string{"data", "type"}, []Permission{ConfigurePermission})
	if !ok {
		return
	}
	data := postForm.Get("data")
	dataType := postForm.Get("type")
	log.Println("Setting target qty")
	var err error
	for _, dataConfig := range strings.Split(data, "|") {
		dataParts := strings.Split(dataConfig, "_")
		if dataType == "" || (dataType == "1" && len(dataParts) != 5) {
			httputil.ResponseFailure(c, httputil.WithReason("Data submitted not enough information"))
			return
		}
		token := dataParts[0]
		// total, _ := strconv.ParseFloat(dataParts[1], 64)
		// reserve, _ := strconv.ParseFloat(dataParts[2], 64)
		// rebalanceThresold, _ := strconv.ParseFloat(dataParts[3], 64)
		// transferThresold, _ := strconv.ParseFloat(dataParts[4], 64)
		_, err = common.GetInternalToken(token)
		if err != nil {
			httputil.ResponseFailure(c, httputil.WithError(err))
			return
		}
	}
	err = self.metric.StorePendingTargetQty(data, dataType)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}

	pendingData, err := self.metric.GetPendingTargetQty()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(pendingData))
	return
}

func (self *HTTPServer) GetAddress(c *gin.Context) {
	httputil.ResponseSuccess(c, httputil.WithData(self.core.GetAddresses()))
	return
}

func (self *HTTPServer) GetTradeHistory(c *gin.Context) {
	_, ok := self.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, RebalancePermission, ConfigurePermission, ConfirmConfPermission})
	if !ok {
		return
	}
	fromTime, toTime, ok := self.ValidateTimeInput(c)
	if !ok {
		return
	}
	data, err := self.app.GetTradeHistory(fromTime, toTime)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (self *HTTPServer) GetGoldData(c *gin.Context) {
	log.Printf("Getting gold data")

	data, err := self.app.GetGoldData(getTimePoint(c, true))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithData(data))
	}
}

func (self *HTTPServer) GetTimeServer(c *gin.Context) {
	httputil.ResponseSuccess(c, httputil.WithData(common.GetTimestamp()))
}

func (self *HTTPServer) GetRebalanceStatus(c *gin.Context) {
	_, ok := self.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, RebalancePermission, ConfigurePermission, ConfirmConfPermission})
	if !ok {
		return
	}
	data, err := self.metric.GetRebalanceControl()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data.Status))
}

func (self *HTTPServer) HoldRebalance(c *gin.Context) {
	_, ok := self.Authenticated(c, []string{}, []Permission{ConfirmConfPermission})
	if !ok {
		return
	}
	if err := self.metric.StoreRebalanceControl(false); err != nil {
		httputil.ResponseFailure(c, httputil.WithReason(err.Error()))
		return
	}
	httputil.ResponseSuccess(c)
	return
}

func (self *HTTPServer) EnableRebalance(c *gin.Context) {
	_, ok := self.Authenticated(c, []string{}, []Permission{ConfirmConfPermission})
	if !ok {
		return
	}
	if err := self.metric.StoreRebalanceControl(true); err != nil {
		httputil.ResponseFailure(c, httputil.WithReason(err.Error()))
	}
	httputil.ResponseSuccess(c)
	return
}

func (self *HTTPServer) GetSetrateStatus(c *gin.Context) {
	_, ok := self.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, RebalancePermission, ConfigurePermission, ConfirmConfPermission})
	if !ok {
		return
	}
	data, err := self.metric.GetSetrateControl()
	if err != nil {
		httputil.ResponseFailure(c)
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data.Status))
}

func (self *HTTPServer) HoldSetrate(c *gin.Context) {
	_, ok := self.Authenticated(c, []string{}, []Permission{ConfirmConfPermission})
	if !ok {
		return
	}
	if err := self.metric.StoreSetrateControl(false); err != nil {
		httputil.ResponseFailure(c, httputil.WithReason(err.Error()))
	}
	httputil.ResponseSuccess(c)
	return
}

func (self *HTTPServer) EnableSetrate(c *gin.Context) {
	_, ok := self.Authenticated(c, []string{}, []Permission{ConfirmConfPermission})
	if !ok {
		return
	}
	if err := self.metric.StoreSetrateControl(true); err != nil {
		httputil.ResponseFailure(c, httputil.WithReason(err.Error()))
	}
	httputil.ResponseSuccess(c)
	return
}

func (self *HTTPServer) GetPWIEquation(c *gin.Context) {
	_, ok := self.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, RebalancePermission, ConfigurePermission, ConfirmConfPermission})
	if !ok {
		return
	}
	data, err := self.metric.GetPWIEquation()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (self *HTTPServer) GetAssetVolume(c *gin.Context) {
	fromTime, _ := strconv.ParseUint(c.Query("fromTime"), 10, 64)
	toTime, _ := strconv.ParseUint(c.Query("toTime"), 10, 64)
	freq := c.Query("freq")
	asset := c.Query("asset")
	data, err := self.stat.GetAssetVolume(fromTime, toTime, freq, asset)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (self *HTTPServer) GetPendingPWIEquation(c *gin.Context) {
	_, ok := self.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, RebalancePermission, ConfigurePermission, ConfirmConfPermission})
	if !ok {
		return
	}
	data, err := self.metric.GetPendingPWIEquation()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (self *HTTPServer) GetBurnFee(c *gin.Context) {
	fromTime, _ := strconv.ParseUint(c.Query("fromTime"), 10, 64)
	toTime, _ := strconv.ParseUint(c.Query("toTime"), 10, 64)
	freq := c.Query("freq")
	reserveAddr := c.Query("reserveAddr")
	if reserveAddr == "" {
		httputil.ResponseFailure(c, httputil.WithReason("reserveAddr is required"))
		return
	}
	data, err := self.stat.GetBurnFee(fromTime, toTime, freq, reserveAddr)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (self *HTTPServer) SetPWIEquation(c *gin.Context) {
	var err error
	postForm, ok := self.Authenticated(c, []string{}, []Permission{ConfigurePermission})
	if !ok {
		return
	}
	data := postForm.Get("data")
	for _, dataConfig := range strings.Split(data, "|") {
		dataParts := strings.Split(dataConfig, "_")
		if len(dataParts) != 4 {
			httputil.ResponseFailure(c, httputil.WithReason("The input data is not correct"))
			return
		}
		token := dataParts[0]
		_, err = common.GetInternalToken(token)
		if err != nil {
			httputil.ResponseFailure(c, httputil.WithError(err))
			return
		}
	}
	err = self.metric.StorePendingPWIEquation(data)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (self *HTTPServer) GetWalletFee(c *gin.Context) {
	fromTime, _ := strconv.ParseUint(c.Query("fromTime"), 10, 64)
	toTime, _ := strconv.ParseUint(c.Query("toTime"), 10, 64)
	freq := c.Query("freq")
	reserveAddr := c.Query("reserveAddr")
	walletAddr := c.Query("walletAddr")
	data, err := self.stat.GetWalletFee(fromTime, toTime, freq, reserveAddr, walletAddr)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (self *HTTPServer) ConfirmPWIEquation(c *gin.Context) {
	postForm, ok := self.Authenticated(c, []string{}, []Permission{ConfirmConfPermission})
	if !ok {
		return
	}
	postData := postForm.Get("data")
	err := self.metric.StorePWIEquation(postData)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (self *HTTPServer) ExceedDailyLimit(c *gin.Context) {
	addr := c.Param("addr")
	log.Printf("Checking daily limit for %s", addr)
	address := ethereum.HexToAddress(addr)
	if address.Big().Cmp(ethereum.Big0) == 0 {
		httputil.ResponseFailure(c, httputil.WithReason("address is not valid"))
		return
	}
	exceeded, err := self.stat.ExceedDailyLimit(address)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithData(exceeded))
	}
}

func (self *HTTPServer) GetUserVolume(c *gin.Context) {
	fromTime, _ := strconv.ParseUint(c.Query("fromTime"), 10, 64)
	toTime, _ := strconv.ParseUint(c.Query("toTime"), 10, 64)
	freq := c.Query("freq")
	userAddr := c.Query("userAddr")
	if userAddr == "" {
		httputil.ResponseFailure(c, httputil.WithReason("User address is required"))
		return
	}
	data, err := self.stat.GetUserVolume(fromTime, toTime, freq, userAddr)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (self *HTTPServer) GetUsersVolume(c *gin.Context) {
	fromTime, _ := strconv.ParseUint(c.Query("fromTime"), 10, 64)
	toTime, _ := strconv.ParseUint(c.Query("toTime"), 10, 64)
	freq := c.Query("freq")
	userAddr := c.Query("userAddr")
	if userAddr == "" {
		httputil.ResponseFailure(c, httputil.WithReason("User address is required"))
		return
	}
	userAddrs := strings.Split(userAddr, ",")
	data, err := self.stat.GetUsersVolume(fromTime, toTime, freq, userAddrs)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (self *HTTPServer) ValidateTimeInput(c *gin.Context) (uint64, uint64, bool) {
	fromTime, ok := strconv.ParseUint(c.Query("fromTime"), 10, 64)
	if ok != nil {
		httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("fromTime param is invalid: %s", ok)))
		return 0, 0, false
	}
	toTime, _ := strconv.ParseUint(c.Query("toTime"), 10, 64)
	if toTime == 0 {
		toTime = common.GetTimepoint()
	}
	return fromTime, toTime, true
}

func (self *HTTPServer) GetTradeSummary(c *gin.Context) {
	fromTime, toTime, ok := self.ValidateTimeInput(c)
	if !ok {
		return
	}
	tzparam, _ := strconv.ParseInt(c.Query("timeZone"), 10, 64)
	if (tzparam < START_TIMEZONE) || (tzparam > END_TIMEZONE) {
		httputil.ResponseFailure(c, httputil.WithReason("Timezone is not supported"))
		return
	}
	data, err := self.stat.GetTradeSummary(fromTime, toTime, tzparam)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (self *HTTPServer) RejectPWIEquation(c *gin.Context) {
	_, ok := self.Authenticated(c, []string{}, []Permission{ConfirmConfPermission})
	if !ok {
		return
	}
	// postData := postForm.RunningMode("data")
	err := self.metric.RemovePendingPWIEquation()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (self *HTTPServer) GetCapByAddress(c *gin.Context) {
	addr := c.Param("addr")
	address := ethereum.HexToAddress(addr)
	if address.Big().Cmp(ethereum.Big0) == 0 {
		httputil.ResponseFailure(c, httputil.WithReason("address is not valid"))
		return
	}
	data, err := self.stat.GetCapByAddress(address)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithData(data))
	}
}

func (self *HTTPServer) GetCapByUser(c *gin.Context) {
	user := c.Param("user")
	data, err := self.stat.GetCapByUser(user)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithData(data))
	}
}

func (self *HTTPServer) GetPendingAddresses(c *gin.Context) {
	data, err := self.stat.GetPendingAddresses()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithData(data))
	}
}

func (self *HTTPServer) UpdateUserAddresses(c *gin.Context) {
	var err error
	postForm, ok := self.Authenticated(c, []string{"user", "addresses", "timestamps"}, []Permission{ConfirmConfPermission})
	if !ok {
		return
	}
	user := postForm.Get("user")
	addresses := postForm.Get("addresses")
	times := postForm.Get("timestamps")
	addrs := []ethereum.Address{}
	timestamps := []uint64{}
	addrsStr := strings.Split(addresses, "-")
	timesStr := strings.Split(times, "-")
	if len(addrsStr) != len(timesStr) {
		httputil.ResponseFailure(c, httputil.WithReason("addresses and timestamps must have the same number of elements"))
		return
	}
	for i, addr := range addrsStr {
		a := ethereum.HexToAddress(addr)
		t, vErr := strconv.ParseUint(timesStr[i], 10, 64)
		if a.Big().Cmp(ethereum.Big0) != 0 && vErr == nil {
			addrs = append(addrs, a)
			timestamps = append(timestamps, t)
		}
	}
	if len(addrs) == 0 {
		httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("user %s doesn't have any valid addresses in %s", user, addresses)))
		return
	}
	err = self.stat.UpdateUserAddresses(user, addrs, timestamps)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c)
	}
}

func (self *HTTPServer) GetWalletStats(c *gin.Context) {
	fromTime, toTime, ok := self.ValidateTimeInput(c)
	if !ok {
		return
	}
	tzparam, _ := strconv.ParseInt(c.Query("timeZone"), 10, 64)
	if (tzparam < START_TIMEZONE) || (tzparam > END_TIMEZONE) {
		httputil.ResponseFailure(c, httputil.WithReason("Timezone is not supported"))
		return
	}
	if toTime == 0 {
		toTime = common.GetTimepoint()
	}
	walletAddr := ethereum.HexToAddress(c.Query("walletAddr"))
	wcap := big.NewInt(0)
	wcap.Exp(big.NewInt(2), big.NewInt(128), big.NewInt(0))
	if walletAddr.Big().Cmp(wcap) < 0 {
		httputil.ResponseFailure(c, httputil.WithReason("Wallet address is invalid, its integer form must be larger than 2^128"))
		return
	}

	data, err := self.stat.GetWalletStats(fromTime, toTime, walletAddr.Hex(), tzparam)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (self *HTTPServer) GetWalletAddress(c *gin.Context) {
	data, err := self.stat.GetWalletAddress()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (self *HTTPServer) GetReserveRate(c *gin.Context) {
	fromTime, _ := strconv.ParseUint(c.Query("fromTime"), 10, 64)
	toTime, _ := strconv.ParseUint(c.Query("toTime"), 10, 64)
	if toTime == 0 {
		toTime = common.GetTimepoint()
	}
	reserveAddr := ethereum.HexToAddress(c.Query("reserveAddr"))
	if reserveAddr.Big().Cmp(ethereum.Big0) == 0 {
		httputil.ResponseFailure(c, httputil.WithReason("Reserve address is invalid"))
		return
	}
	data, err := self.stat.GetReserveRates(fromTime, toTime, reserveAddr)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (self *HTTPServer) GetExchangesStatus(c *gin.Context) {
	data, err := self.app.GetExchangeStatus()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (self *HTTPServer) UpdateExchangeStatus(c *gin.Context) {
	postForm, ok := self.Authenticated(c, []string{"exchange", "status", "timestamp"}, []Permission{ConfirmConfPermission})
	if !ok {
		return
	}
	exchange := postForm.Get("exchange")
	status, err := strconv.ParseBool(postForm.Get("status"))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	timestamp, err := strconv.ParseUint(postForm.Get("timestamp"), 10, 64)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	_, err = common.GetExchange(strings.ToLower(exchange))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	err = self.app.UpdateExchangeStatus(exchange, status, timestamp)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (self *HTTPServer) GetCountryStats(c *gin.Context) {
	fromTime, toTime, ok := self.ValidateTimeInput(c)
	if !ok {
		return
	}
	country := c.Query("country")
	tzparam, _ := strconv.ParseInt(c.Query("timeZone"), 10, 64)
	if (tzparam < START_TIMEZONE) || (tzparam > END_TIMEZONE) {
		httputil.ResponseFailure(c, httputil.WithReason("Timezone is not supported"))
		return
	}
	data, err := self.stat.GetGeoData(fromTime, toTime, country, tzparam)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (self *HTTPServer) GetHeatMap(c *gin.Context) {
	fromTime, toTime, ok := self.ValidateTimeInput(c)
	if !ok {
		return
	}
	tzparam, _ := strconv.ParseInt(c.Query("timeZone"), 10, 64)
	if (tzparam < START_TIMEZONE) || (tzparam > END_TIMEZONE) {
		httputil.ResponseFailure(c, httputil.WithReason("Timezone is not supported"))
		return
	}

	data, err := self.stat.GetHeatMap(fromTime, toTime, tzparam)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (self *HTTPServer) GetCountries(c *gin.Context) {
	data, _ := self.stat.GetCountries()
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (self *HTTPServer) UpdatePriceAnalyticData(c *gin.Context) {
	postForm, ok := self.Authenticated(c, []string{}, []Permission{RebalancePermission})
	if !ok {
		return
	}
	timestamp, err := strconv.ParseUint(postForm.Get("timestamp"), 10, 64)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	value := []byte(postForm.Get("value"))
	if len(value) > MAX_DATA_SIZE {
		httputil.ResponseFailure(c, httputil.WithReason(errDataSizeExceed.Error()))
		return
	}
	err = self.stat.UpdatePriceAnalyticData(timestamp, value)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}
func (self *HTTPServer) GetPriceAnalyticData(c *gin.Context) {
	_, ok := self.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, ConfigurePermission, ConfirmConfPermission, RebalancePermission})
	if !ok {
		return
	}
	fromTime, toTime, ok := self.ValidateTimeInput(c)
	if !ok {
		return
	}
	if toTime == 0 {
		toTime = common.GetTimepoint()
	}

	data, err := self.stat.GetPriceAnalyticData(fromTime, toTime)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (self *HTTPServer) ExchangeNotification(c *gin.Context) {
	postForm, ok := self.Authenticated(c, []string{
		"exchange", "action", "token", "fromTime", "toTime", "isWarning"}, []Permission{RebalancePermission})
	if !ok {
		return
	}

	exchange := postForm.Get("exchange")
	action := postForm.Get("action")
	tokenPair := postForm.Get("token")
	from, _ := strconv.ParseUint(postForm.Get("fromTime"), 10, 64)
	to, _ := strconv.ParseUint(postForm.Get("toTime"), 10, 64)
	isWarning, _ := strconv.ParseBool(postForm.Get("isWarning"))
	msg := postForm.Get("msg")

	err := self.app.UpdateExchangeNotification(exchange, action, tokenPair, from, to, isWarning, msg)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (self *HTTPServer) GetNotifications(c *gin.Context) {
	_, ok := self.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, RebalancePermission, ConfigurePermission, ConfirmConfPermission})
	if !ok {
		return
	}
	data, err := self.app.GetNotifications()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (self *HTTPServer) GetUserList(c *gin.Context) {
	_, ok := self.Authenticated(c, []string{"fromTime", "toTime", "timeZone"}, []Permission{ReadOnlyPermission, RebalancePermission, ConfigurePermission, ConfirmConfPermission})
	if !ok {
		return
	}
	fromTime, toTime, ok := self.ValidateTimeInput(c)
	if !ok {
		return
	}
	timeZone, err := strconv.ParseInt(c.Query("timeZone"), 10, 64)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("timeZone is required: %s", err.Error())))
		return
	}
	data, err := self.stat.GetUserList(fromTime, toTime, timeZone)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (self *HTTPServer) GetReserveVolume(c *gin.Context) {
	fromTime, _ := strconv.ParseUint(c.Query("fromTime"), 10, 64)
	toTime, _ := strconv.ParseUint(c.Query("toTime"), 10, 64)
	freq := c.Query("freq")
	reserveAddr := c.Query("reserveAddr")
	if reserveAddr == "" {
		httputil.ResponseFailure(c, httputil.WithReason("reserve address is required"))
		return
	}
	tokenName := c.Query("token")
	if tokenName == "" {
		httputil.ResponseFailure(c, httputil.WithReason("token is required"))
		return
	}
	token, err := common.GetNetworkToken(tokenName)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	data, err := self.stat.GetReserveVolume(fromTime, toTime, freq, reserveAddr, token.Address)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (self *HTTPServer) SetStableTokenParams(c *gin.Context) {
	postForm, ok := self.Authenticated(c, []string{}, []Permission{ConfigurePermission})
	if !ok {
		return
	}
	value := []byte(postForm.Get("value"))
	if len(value) > MAX_DATA_SIZE {
		httputil.ResponseFailure(c, httputil.WithReason(errDataSizeExceed.Error()))
		return
	}
	err := self.metric.SetStableTokenParams(value)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (self *HTTPServer) ConfirmStableTokenParams(c *gin.Context) {
	postForm, ok := self.Authenticated(c, []string{}, []Permission{ConfirmConfPermission})
	if !ok {
		return
	}
	value := []byte(postForm.Get("value"))
	if len(value) > MAX_DATA_SIZE {
		httputil.ResponseFailure(c, httputil.WithReason(errDataSizeExceed.Error()))
		return
	}
	err := self.metric.ConfirmStableTokenParams(value)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (self *HTTPServer) RejectStableTokenParams(c *gin.Context) {
	_, ok := self.Authenticated(c, []string{}, []Permission{ConfirmConfPermission})
	if !ok {
		return
	}
	err := self.metric.RemovePendingStableTokenParams()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (self *HTTPServer) GetPendingStableTokenParams(c *gin.Context) {
	_, ok := self.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, ConfigurePermission, ConfirmConfPermission, RebalancePermission})
	if !ok {
		return
	}

	data, err := self.metric.GetPendingStableTokenParams()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (self *HTTPServer) GetStableTokenParams(c *gin.Context) {
	_, ok := self.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, ConfigurePermission, ConfirmConfPermission, RebalancePermission})
	if !ok {
		return
	}

	data, err := self.metric.GetStableTokenParams()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (self *HTTPServer) GetTokenHeatmap(c *gin.Context) {
	fromTime, toTime, ok := self.ValidateTimeInput(c)
	if !ok {
		return
	}
	freq := c.Query("freq")
	token := c.Query("token")
	if token == "" {
		httputil.ResponseFailure(c, httputil.WithReason("token param is required"))
		return
	}

	data, err := self.stat.GetTokenHeatmap(fromTime, toTime, token, freq)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (self *HTTPServer) SetTargetQtyV2(c *gin.Context) {
	postForm, ok := self.Authenticated(c, []string{}, []Permission{ConfigurePermission})
	if !ok {
		return
	}
	value := []byte(postForm.Get("value"))
	if len(value) > MAX_DATA_SIZE {
		httputil.ResponseFailure(c, httputil.WithReason(errDataSizeExceed.Error()))
		return
	}
	err := self.metric.StorePendingTargetQtyV2(value)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (self *HTTPServer) GetPendingTargetQtyV2(c *gin.Context) {
	_, ok := self.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, ConfigurePermission, ConfirmConfPermission, RebalancePermission})
	if !ok {
		return
	}

	data, err := self.metric.GetPendingTargetQtyV2()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (self *HTTPServer) ConfirmTargetQtyV2(c *gin.Context) {
	postForm, ok := self.Authenticated(c, []string{}, []Permission{ConfirmConfPermission})
	if !ok {
		return
	}
	value := []byte(postForm.Get("value"))
	if len(value) > MAX_DATA_SIZE {
		httputil.ResponseFailure(c, httputil.WithReason(errDataSizeExceed.Error()))
		return
	}
	err := self.metric.ConfirmTargetQtyV2(value)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	}
	httputil.ResponseSuccess(c)
}

func (self *HTTPServer) CancelTargetQtyV2(c *gin.Context) {
	_, ok := self.Authenticated(c, []string{}, []Permission{ConfirmConfPermission})
	if !ok {
		return
	}
	err := self.metric.RemovePendingTargetQtyV2()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (self *HTTPServer) GetTargetQtyV2(c *gin.Context) {
	_, ok := self.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, ConfigurePermission, ConfirmConfPermission, RebalancePermission})
	if !ok {
		return
	}

	data, err := self.metric.GetTargetQtyV2()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (self *HTTPServer) GetFeeSetRateByDay(c *gin.Context) {
	fromTime, toTime, ok := self.ValidateTimeInput(c)
	if !ok {
		return
	}
	data, err := self.stat.GetFeeSetRateByDay(fromTime, toTime)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithReason(err.Error()))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (self *HTTPServer) register() {
	if self.core != nil && self.app != nil {
		v2 := self.r.Group("/v2")

		self.r.GET("/prices-version", self.AllPricesVersion)
		self.r.GET("/prices", self.AllPrices)
		self.r.GET("/prices/:base/:quote", self.Price)
		self.r.GET("/getrates", self.GetRate)
		self.r.GET("/get-all-rates", self.GetRates)

		self.r.GET("/authdata-version", self.AuthDataVersion)
		self.r.GET("/authdata", self.AuthData)
		self.r.GET("/activities", self.GetActivities)
		self.r.GET("/immediate-pending-activities", self.ImmediatePendingActivities)
		self.r.GET("/metrics", self.Metrics)
		self.r.POST("/metrics", self.StoreMetrics)

		self.r.POST("/cancelorder/:exchangeid", self.CancelOrder)
		self.r.POST("/deposit/:exchangeid", self.Deposit)
		self.r.POST("/withdraw/:exchangeid", self.Withdraw)
		self.r.POST("/trade/:exchangeid", self.Trade)
		self.r.POST("/setrates", self.SetRate)
		self.r.GET("/exchangeinfo", self.GetExchangeInfo)
		self.r.GET("/exchangeinfo/:exchangeid/:base/:quote", self.GetPairInfo)
		self.r.GET("/exchangefees", self.GetFee)
		self.r.GET("/exchange-min-deposit", self.GetMinDeposit)
		self.r.GET("/exchangefees/:exchangeid", self.GetExchangeFee)
		self.r.GET("/core/addresses", self.GetAddress)
		self.r.GET("/tradehistory", self.GetTradeHistory)

		self.r.GET("/targetqty", self.GetTargetQty)
		self.r.GET("/pendingtargetqty", self.GetPendingTargetQty)
		self.r.POST("/settargetqty", self.SetTargetQty)
		self.r.POST("/confirmtargetqty", self.ConfirmTargetQty)
		self.r.POST("/canceltargetqty", self.CancelTargetQty)

		v2.GET("/targetqty", self.GetTargetQtyV2)
		v2.GET("/pendingtargetqty", self.GetPendingTargetQtyV2)
		v2.POST("/settargetqty", self.SetTargetQtyV2)
		v2.POST("/confirmtargetqty", self.ConfirmTargetQtyV2)
		v2.POST("/canceltargetqty", self.CancelTargetQtyV2)

		self.r.GET("/timeserver", self.GetTimeServer)

		self.r.GET("/rebalancestatus", self.GetRebalanceStatus)
		self.r.POST("/holdrebalance", self.HoldRebalance)
		self.r.POST("/enablerebalance", self.EnableRebalance)

		self.r.GET("/setratestatus", self.GetSetrateStatus)
		self.r.POST("/holdsetrate", self.HoldSetrate)
		self.r.POST("/enablesetrate", self.EnableSetrate)

		self.r.GET("/pwis-equation", self.GetPWIEquation)
		self.r.GET("/pending-pwis-equation", self.GetPendingPWIEquation)
		self.r.POST("/set-pwis-equation", self.SetPWIEquation)
		self.r.POST("/confirm-pwis-equation", self.ConfirmPWIEquation)
		self.r.POST("/reject-pwis-equation", self.RejectPWIEquation)

		v2.GET("/pwis-equation", self.GetPWIEquationV2)
		v2.GET("/pending-pwis-equation", self.GetPendingPWIEquationV2)
		v2.POST("/set-pwis-equation", self.SetPWIEquationV2)
		v2.POST("/confirm-pwis-equation", self.ConfirmPWIEquationV2)
		v2.POST("/reject-pwis-equation", self.RejectPWIEquationV2)

		self.r.GET("/get-exchange-status", self.GetExchangesStatus)
		self.r.POST("/update-exchange-status", self.UpdateExchangeStatus)

		self.r.POST("/exchange-notification", self.ExchangeNotification)
		self.r.GET("/exchange-notifications", self.GetNotifications)

		self.r.POST("/set-stable-token-params", self.SetStableTokenParams)
		self.r.POST("/confirm-stable-token-params", self.ConfirmStableTokenParams)
		self.r.POST("/reject-stable-token-params", self.RejectStableTokenParams)
		self.r.GET("/pending-stable-token-params", self.GetPendingStableTokenParams)
		self.r.GET("/stable-token-params", self.GetStableTokenParams)

		self.r.GET("/gold-feed", self.GetGoldData)
	}

	if self.stat != nil {
		self.r.GET("/cap-by-address/:addr", self.GetCapByAddress)
		self.r.GET("/cap-by-user/:user", self.GetCapByUser)
		self.r.GET("/richguy/:addr", self.ExceedDailyLimit)
		self.r.GET("/tradelogs", self.TradeLogs)
		self.r.GET("/catlogs", self.CatLogs)
		self.r.GET("/get-asset-volume", self.GetAssetVolume)
		self.r.GET("/get-burn-fee", self.GetBurnFee)
		self.r.GET("/get-wallet-fee", self.GetWalletFee)
		self.r.GET("/get-user-volume", self.GetUserVolume)
		self.r.GET("/get-users-volume", self.GetUsersVolume)
		self.r.GET("/get-trade-summary", self.GetTradeSummary)
		self.r.POST("/update-user-addresses", self.UpdateUserAddresses)
		self.r.GET("/get-pending-addresses", self.GetPendingAddresses)
		self.r.GET("/get-reserve-rate", self.GetReserveRate)
		self.r.GET("/get-wallet-stats", self.GetWalletStats)
		self.r.GET("/get-wallet-address", self.GetWalletAddress)
		self.r.GET("/get-country-stats", self.GetCountryStats)
		self.r.GET("/get-heat-map", self.GetHeatMap)
		self.r.GET("/get-countries", self.GetCountries)
		self.r.POST("/update-price-analytic-data", self.UpdatePriceAnalyticData)
		self.r.GET("/get-price-analytic-data", self.GetPriceAnalyticData)
		self.r.GET("/get-reserve-volume", self.GetReserveVolume)
		self.r.GET("/get-user-list", self.GetUserList)
		self.r.GET("/get-token-heatmap", self.GetTokenHeatmap)
		self.r.GET("/get-fee-setrate", self.GetFeeSetRateByDay)
	}
}

func (self *HTTPServer) Run() {
	self.register()
	self.r.Run(self.host)
}

func NewHTTPServer(
	app reserve.ReserveData,
	core reserve.ReserveCore,
	stat reserve.ReserveStats,
	metric metric.MetricStorage,
	host string,
	enableAuth bool,
	authEngine Authentication,
	env string) *HTTPServer {

	r := gin.Default()
	sentryCli, err := raven.NewWithTags(
		"https://bf15053001464a5195a81bc41b644751:eff41ac715114b20b940010208271b13@sentry.io/228067",
		map[string]string{
			"env": env,
		},
	)
	if err != nil {
		panic(err)
	}
	r.Use(sentry.Recovery(
		sentryCli,
		false,
	))
	corsConfig := cors.DefaultConfig()
	corsConfig.AddAllowHeaders("signed")
	corsConfig.AllowAllOrigins = true
	corsConfig.MaxAge = 5 * time.Minute
	r.Use(cors.New(corsConfig))

	return &HTTPServer{
		app, core, stat, metric, host, enableAuth, authEngine, r,
	}
}
