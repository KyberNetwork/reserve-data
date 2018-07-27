package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/KyberNetwork/reserve-data"
	"github.com/KyberNetwork/reserve-data/blockchain"
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
	maxTimespot   uint64 = 18446744073709551615
	maxDataSize   int    = 1000000 //1 Megabyte in byte
	startTimezone int64  = -11
	endTimezone   int64  = 14
)

var (
	// errDataSizeExceed is returned when the post data is larger than maxDataSize.
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
	blockchain  *blockchain.Blockchain
	setting     Setting
}

func getTimePoint(c *gin.Context, useDefault bool) uint64 {
	timestamp := c.DefaultQuery("timestamp", "")
	if timestamp == "" {
		if useDefault {
			log.Printf("Interpreted timestamp to default - %d\n", maxTimespot)
			return maxTimespot
		} else {
			timepoint := common.GetTimepoint()
			log.Printf("Interpreted timestamp to current time - %d\n", timepoint)
			return uint64(timepoint)
		}
	} else {
		timepoint, err := strconv.ParseUint(timestamp, 10, 64)
		if err != nil {
			log.Printf("Interpreted timestamp(%s) to default - %d", timestamp, maxTimespot)
			return maxTimespot
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
func (h *HTTPServer) Authenticated(c *gin.Context, requiredParams []string, perms []Permission) (url.Values, bool) {
	err := c.Request.ParseForm()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("Malformed request package: %s", err.Error())))
		return c.Request.Form, false
	}

	if !h.authEnabled {
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
	userPerms := h.auth.GetPermission(signed, message)
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

func (h *HTTPServer) AllPricesVersion(c *gin.Context) {
	log.Printf("Getting all prices version")
	data, err := h.app.CurrentPriceVersion(getTimePoint(c, true))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithField("version", data))
	}
}

func (h *HTTPServer) AllPrices(c *gin.Context) {
	log.Printf("Getting all prices \n")
	data, err := h.app.GetAllPrices(getTimePoint(c, true))
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

func (h *HTTPServer) Price(c *gin.Context) {
	base := c.Param("base")
	quote := c.Param("quote")
	log.Printf("Getting price for %s - %s \n", base, quote)
	pair, err := h.setting.NewTokenPairFromID(base, quote)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithReason("Token pair is not supported"))
	} else {
		data, err := h.app.GetOnePrice(pair.PairID(), getTimePoint(c, true))
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

func (h *HTTPServer) AuthDataVersion(c *gin.Context) {
	log.Printf("Getting current auth data snapshot version")
	_, ok := h.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, RebalancePermission, ConfigurePermission, ConfirmConfPermission})
	if !ok {
		return
	}

	data, err := h.app.CurrentAuthDataVersion(getTimePoint(c, true))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithField("version", data))
	}
}

func (h *HTTPServer) AuthData(c *gin.Context) {
	log.Printf("Getting current auth data snapshot \n")
	_, ok := h.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, RebalancePermission, ConfigurePermission, ConfirmConfPermission})
	if !ok {
		return
	}

	data, err := h.app.GetAuthData(getTimePoint(c, true))
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

func (h *HTTPServer) GetRates(c *gin.Context) {
	log.Printf("Getting all rates \n")
	fromTime, _ := strconv.ParseUint(c.Query("fromTime"), 10, 64)
	toTime, _ := strconv.ParseUint(c.Query("toTime"), 10, 64)
	if toTime == 0 {
		toTime = maxTimespot
	}
	data, err := h.app.GetRates(fromTime, toTime)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithData(data))
	}
}

func (h *HTTPServer) GetRate(c *gin.Context) {
	log.Printf("Getting all rates \n")
	data, err := h.app.GetRate(getTimePoint(c, true))
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

func (h *HTTPServer) SetRate(c *gin.Context) {
	postForm, ok := h.Authenticated(c, []string{"tokens", "buys", "sells", "block", "afp_mid", "msgs"}, []Permission{RebalancePermission})
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
		token, err := h.setting.GetInternalTokenByID(tok)
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
		var r *big.Int
		if r, err = hexutil.DecodeBig(rate); err != nil {
			httputil.ResponseFailure(c, httputil.WithError(err))
			return
		}
		bigAfpMid = append(bigAfpMid, r)
	}
	id, err := h.core.SetRates(tokens, bigBuys, bigSells, big.NewInt(intBlock), bigAfpMid, msgs)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithField("id", id))
}

func (h *HTTPServer) Trade(c *gin.Context) {
	postForm, ok := h.Authenticated(c, []string{"base", "quote", "amount", "rate", "type"}, []Permission{RebalancePermission})
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
	base, err := h.setting.GetInternalTokenByID(baseTokenParam)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	quote, err := h.setting.GetInternalTokenByID(quoteTokenParam)
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
	id, done, remaining, finished, err := h.core.Trade(
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

func (h *HTTPServer) CancelOrder(c *gin.Context) {
	postForm, ok := h.Authenticated(c, []string{"order_id"}, []Permission{RebalancePermission})
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
	err = h.core.CancelOrder(activityID, exchange)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (h *HTTPServer) Withdraw(c *gin.Context) {
	postForm, ok := h.Authenticated(c, []string{"token", "amount"}, []Permission{RebalancePermission})
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
	token, err := h.setting.GetInternalTokenByID(tokenParam)
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
	id, err := h.core.Withdraw(exchange, token, amount, getTimePoint(c, false))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithField("id", id))
}

func (h *HTTPServer) Deposit(c *gin.Context) {
	postForm, ok := h.Authenticated(c, []string{"amount", "token"}, []Permission{RebalancePermission})
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
	token, err := h.setting.GetInternalTokenByID(tokenParam)
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
	id, err := h.core.Deposit(exchange, token, amount, getTimePoint(c, false))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithField("id", id))
}

func (h *HTTPServer) GetActivities(c *gin.Context) {
	log.Printf("Getting all activity records \n")
	_, ok := h.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, RebalancePermission, ConfigurePermission, ConfirmConfPermission})
	if !ok {
		return
	}
	fromTime, _ := strconv.ParseUint(c.Query("fromTime"), 10, 64)
	toTime, _ := strconv.ParseUint(c.Query("toTime"), 10, 64)
	if toTime == 0 {
		toTime = common.GetTimepoint()
	}

	data, err := h.app.GetRecords(fromTime*1000000, toTime*1000000)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithData(data))
	}
}

func (h *HTTPServer) CatLogs(c *gin.Context) {
	log.Printf("Getting cat logs")
	fromTime, err := strconv.ParseUint(c.Query("fromTime"), 10, 64)
	if err != nil {
		fromTime = 0
	}
	toTime, err := strconv.ParseUint(c.Query("toTime"), 10, 64)
	if err != nil || toTime == 0 {
		toTime = common.GetTimepoint()
	}

	data, err := h.stat.GetCatLogs(fromTime, toTime)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithData(data))
	}
}

func (h *HTTPServer) TradeLogs(c *gin.Context) {
	log.Printf("Getting trade logs")
	fromTime, err := strconv.ParseUint(c.Query("fromTime"), 10, 64)
	if err != nil {
		fromTime = 0
	}
	toTime, err := strconv.ParseUint(c.Query("toTime"), 10, 64)
	if err != nil || toTime == 0 {
		toTime = common.GetTimepoint()
	}

	data, err := h.stat.GetTradeLogs(fromTime, toTime)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithData(data))
	}
}

func (h *HTTPServer) StopFetcher(c *gin.Context) {
	err := h.app.Stop()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c)
	}
}

func (h *HTTPServer) ImmediatePendingActivities(c *gin.Context) {
	log.Printf("Getting all immediate pending activity records \n")
	_, ok := h.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, RebalancePermission, ConfigurePermission, ConfirmConfPermission})
	if !ok {
		return
	}

	data, err := h.app.GetPendingActivities()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithData(data))
	}
}

func (h *HTTPServer) Metrics(c *gin.Context) {
	response := common.MetricResponse{
		Timestamp: common.GetTimepoint(),
	}
	log.Printf("Getting metrics")
	postForm, ok := h.Authenticated(c, []string{"tokens", "from", "to"}, []Permission{ReadOnlyPermission, RebalancePermission, ConfigurePermission, ConfirmConfPermission})
	if !ok {
		return
	}
	tokenParam := postForm.Get("tokens")
	fromParam := postForm.Get("from")
	toParam := postForm.Get("to")
	tokens := []common.Token{}
	for _, tok := range strings.Split(tokenParam, "-") {
		token, err := h.setting.GetInternalTokenByID(tok)
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
	data, err := h.metric.GetMetric(tokens, from, to)
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

func (h *HTTPServer) StoreMetrics(c *gin.Context) {
	log.Printf("Storing metrics")
	postForm, ok := h.Authenticated(c, []string{"timestamp", "data"}, []Permission{RebalancePermission})
	if !ok {
		return
	}
	timestampParam := postForm.Get("timestamp")
	dataParam := postForm.Get("data")

	timestamp, err := strconv.ParseUint(timestampParam, 10, 64)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	}
	metricEntry := common.MetricEntry{}
	metricEntry.Timestamp = timestamp
	metricEntry.Data = map[string]common.TokenMetric{}
	// data must be in form of <token>_afpmid_spread|<token>_afpmid_spread|...
	for _, tokenData := range strings.Split(dataParam, "|") {
		var (
			afpmid float64
			spread float64
		)

		parts := strings.Split(tokenData, "_")
		if len(parts) != 3 {
			httputil.ResponseFailure(c, httputil.WithReason("submitted data is not in correct format"))
			return
		}
		token := parts[0]
		afpmidStr := parts[1]
		spreadStr := parts[2]

		if afpmid, err = strconv.ParseFloat(afpmidStr, 64); err != nil {
			httputil.ResponseFailure(c, httputil.WithReason("Afp mid "+afpmidStr+" is not float64"))
			return
		}

		if spread, err = strconv.ParseFloat(spreadStr, 64); err != nil {
			httputil.ResponseFailure(c, httputil.WithReason("Spread "+spreadStr+" is not float64"))
			return
		}
		metricEntry.Data[token] = common.TokenMetric{
			AfpMid: afpmid,
			Spread: spread,
		}
	}

	err = h.metric.StoreMetric(&metricEntry, common.GetTimepoint())
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c)
	}
}

//ValidateExchangeInfo validate if data is complete exchange info with all token pairs supported
// func ValidateExchangeInfo(exchange common.Exchange, data map[common.TokenPairID]common.ExchangePrecisionLimit) error {
// 	exInfo, err :=h
// 	pairs := exchange.Pairs()
// 	for _, pair := range pairs {
// 		// stable exchange is a simulated exchange which is not a real exchange
// 		// we do not do rebalance on stable exchange then it also does not need to have exchange info (and it actully does not have one)
// 		// therefore we skip checking it for supported tokens
// 		if exchange.ID() == common.ExchangeID("stable_exchange") {
// 			continue
// 		}
// 		if _, exist := data[pair.PairID()]; !exist {
// 			return fmt.Errorf("exchange info of %s lack of token %s", exchange.ID(), string(pair.PairID()))
// 		}
// 	}
// 	return nil
// }

//GetExchangeInfo return exchange info of one exchange if it is given exchangeID
//otherwise return all exchanges info
func (h *HTTPServer) GetExchangeInfo(c *gin.Context) {
	exchangeParam := c.Query("exchangeid")
	if exchangeParam == "" {
		data := map[string]common.ExchangeInfo{}
		for _, ex := range common.SupportedExchanges {
			exchangeInfo, err := ex.GetInfo()
			if err != nil {
				httputil.ResponseFailure(c, httputil.WithError(err))
				return
			}
			responseData := exchangeInfo.GetData()
			// if err := ValidateExchangeInfo(exchangeInfo, responseData); err != nil {
			// 	httputil.ResponseFailure(c, httputil.WithError(err))
			// 	return
			// }
			data[string(ex.ID())] = responseData
		}
		httputil.ResponseSuccess(c, httputil.WithData(data))
		return
	}
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

func (h *HTTPServer) GetFee(c *gin.Context) {
	data := map[string]common.ExchangeFees{}
	for _, exchange := range common.SupportedExchanges {
		fee, err := exchange.GetFee()
		if err != nil {
			httputil.ResponseFailure(c, httputil.WithError(err))
			return
		}
		data[string(exchange.ID())] = fee
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
	return
}

func (h *HTTPServer) GetMinDeposit(c *gin.Context) {
	data := map[string]common.ExchangesMinDeposit{}
	for _, exchange := range common.SupportedExchanges {
		minDeposit, err := exchange.GetMinDeposit()
		if err != nil {
			httputil.ResponseFailure(c, httputil.WithError(err))
			return
		}
		data[string(exchange.ID())] = minDeposit
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
	return
}

func (h *HTTPServer) GetTradeHistory(c *gin.Context) {
	_, ok := h.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, RebalancePermission, ConfigurePermission, ConfirmConfPermission})
	if !ok {
		return
	}
	fromTime, toTime, ok := h.ValidateTimeInput(c)
	if !ok {
		return
	}
	data, err := h.app.GetTradeHistory(fromTime, toTime)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (h *HTTPServer) GetGoldData(c *gin.Context) {
	log.Printf("Getting gold data")

	data, err := h.app.GetGoldData(getTimePoint(c, true))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithData(data))
	}
}

func (h *HTTPServer) GetTimeServer(c *gin.Context) {
	httputil.ResponseSuccess(c, httputil.WithData(common.GetTimestamp()))
}

func (h *HTTPServer) GetRebalanceStatus(c *gin.Context) {
	_, ok := h.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, RebalancePermission, ConfigurePermission, ConfirmConfPermission})
	if !ok {
		return
	}
	data, err := h.metric.GetRebalanceControl()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data.Status))
}

func (h *HTTPServer) HoldRebalance(c *gin.Context) {
	_, ok := h.Authenticated(c, []string{}, []Permission{ConfirmConfPermission})
	if !ok {
		return
	}
	if err := h.metric.StoreRebalanceControl(false); err != nil {
		httputil.ResponseFailure(c, httputil.WithReason(err.Error()))
		return
	}
	httputil.ResponseSuccess(c)
	return
}

func (h *HTTPServer) EnableRebalance(c *gin.Context) {
	_, ok := h.Authenticated(c, []string{}, []Permission{ConfirmConfPermission})
	if !ok {
		return
	}
	if err := h.metric.StoreRebalanceControl(true); err != nil {
		httputil.ResponseFailure(c, httputil.WithReason(err.Error()))
	}
	httputil.ResponseSuccess(c)
	return
}

func (h *HTTPServer) GetSetrateStatus(c *gin.Context) {
	_, ok := h.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, RebalancePermission, ConfigurePermission, ConfirmConfPermission})
	if !ok {
		return
	}
	data, err := h.metric.GetSetrateControl()
	if err != nil {
		httputil.ResponseFailure(c)
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data.Status))
}

func (h *HTTPServer) HoldSetrate(c *gin.Context) {
	_, ok := h.Authenticated(c, []string{}, []Permission{ConfirmConfPermission})
	if !ok {
		return
	}
	if err := h.metric.StoreSetrateControl(false); err != nil {
		httputil.ResponseFailure(c, httputil.WithReason(err.Error()))
	}
	httputil.ResponseSuccess(c)
	return
}

func (h *HTTPServer) EnableSetrate(c *gin.Context) {
	_, ok := h.Authenticated(c, []string{}, []Permission{ConfirmConfPermission})
	if !ok {
		return
	}
	if err := h.metric.StoreSetrateControl(true); err != nil {
		httputil.ResponseFailure(c, httputil.WithReason(err.Error()))
	}
	httputil.ResponseSuccess(c)
	return
}

func (h *HTTPServer) GetAssetVolume(c *gin.Context) {
	fromTime, _ := strconv.ParseUint(c.Query("fromTime"), 10, 64)
	toTime, _ := strconv.ParseUint(c.Query("toTime"), 10, 64)
	freq := c.Query("freq")
	asset := c.Query("asset")
	data, err := h.stat.GetAssetVolume(fromTime, toTime, freq, asset)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (h *HTTPServer) GetBurnFee(c *gin.Context) {
	fromTime, _ := strconv.ParseUint(c.Query("fromTime"), 10, 64)
	toTime, _ := strconv.ParseUint(c.Query("toTime"), 10, 64)
	freq := c.Query("freq")
	reserveAddr := c.Query("reserveAddr")
	if reserveAddr == "" {
		httputil.ResponseFailure(c, httputil.WithReason("reserveAddr is required"))
		return
	}
	data, err := h.stat.GetBurnFee(fromTime, toTime, freq, reserveAddr)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (h *HTTPServer) GetWalletFee(c *gin.Context) {
	fromTime, _ := strconv.ParseUint(c.Query("fromTime"), 10, 64)
	toTime, _ := strconv.ParseUint(c.Query("toTime"), 10, 64)
	freq := c.Query("freq")
	reserveAddr := c.Query("reserveAddr")
	walletAddr := c.Query("walletAddr")
	data, err := h.stat.GetWalletFee(fromTime, toTime, freq, reserveAddr, walletAddr)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (h *HTTPServer) ExceedDailyLimit(c *gin.Context) {
	addr := c.Param("addr")
	log.Printf("Checking daily limit for %s", addr)
	address := ethereum.HexToAddress(addr)
	if address.Big().Cmp(ethereum.Big0) == 0 {
		httputil.ResponseFailure(c, httputil.WithReason("address is not valid"))
		return
	}
	exceeded, err := h.stat.ExceedDailyLimit(address)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithData(exceeded))
	}
}

func (h *HTTPServer) GetUserVolume(c *gin.Context) {
	fromTime, _ := strconv.ParseUint(c.Query("fromTime"), 10, 64)
	toTime, _ := strconv.ParseUint(c.Query("toTime"), 10, 64)
	freq := c.Query("freq")
	userAddr := c.Query("userAddr")
	if userAddr == "" {
		httputil.ResponseFailure(c, httputil.WithReason("User address is required"))
		return
	}
	data, err := h.stat.GetUserVolume(fromTime, toTime, freq, userAddr)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (h *HTTPServer) GetUsersVolume(c *gin.Context) {
	fromTime, _ := strconv.ParseUint(c.Query("fromTime"), 10, 64)
	toTime, _ := strconv.ParseUint(c.Query("toTime"), 10, 64)
	freq := c.Query("freq")
	userAddr := c.Query("userAddr")
	if userAddr == "" {
		httputil.ResponseFailure(c, httputil.WithReason("User address is required"))
		return
	}
	userAddrs := strings.Split(userAddr, ",")
	data, err := h.stat.GetUsersVolume(fromTime, toTime, freq, userAddrs)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (h *HTTPServer) ValidateTimeInput(c *gin.Context) (uint64, uint64, bool) {
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

func (h *HTTPServer) GetTradeSummary(c *gin.Context) {
	fromTime, toTime, ok := h.ValidateTimeInput(c)
	if !ok {
		return
	}
	tzparam, _ := strconv.ParseInt(c.Query("timeZone"), 10, 64)
	if (tzparam < startTimezone) || (tzparam > endTimezone) {
		httputil.ResponseFailure(c, httputil.WithReason("Timezone is not supported"))
		return
	}
	data, err := h.stat.GetTradeSummary(fromTime, toTime, tzparam)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (h *HTTPServer) GetCapByAddress(c *gin.Context) {
	addr := c.Param("addr")
	address := ethereum.HexToAddress(addr)
	if address.Big().Cmp(ethereum.Big0) == 0 {
		httputil.ResponseFailure(c, httputil.WithReason("address is not valid"))
		return
	}
	data, err := h.stat.GetCapByAddress(address)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithData(data))
	}
}

func (h *HTTPServer) GetCapByUser(c *gin.Context) {
	user := c.Param("user")
	data, err := h.stat.GetCapByUser(user)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithData(data))
	}
}

func (h *HTTPServer) GetPendingAddresses(c *gin.Context) {
	data, err := h.stat.GetPendingAddresses()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithData(data))
	}
}

func (h *HTTPServer) UpdateUserAddresses(c *gin.Context) {
	var err error
	postForm, ok := h.Authenticated(c, []string{"user", "addresses", "timestamps"}, []Permission{ConfirmConfPermission})
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
		var (
			t uint64
			a = ethereum.HexToAddress(addr)
		)
		t, err = strconv.ParseUint(timesStr[i], 10, 64)
		if a.Big().Cmp(ethereum.Big0) != 0 && err == nil {
			addrs = append(addrs, a)
			timestamps = append(timestamps, t)
		}
	}
	if len(addrs) == 0 {
		httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("user %s doesn't have any valid addresses in %s", user, addresses)))
		return
	}
	err = h.stat.UpdateUserAddresses(user, addrs, timestamps)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c)
	}
}

func (h *HTTPServer) GetWalletStats(c *gin.Context) {
	fromTime, toTime, ok := h.ValidateTimeInput(c)
	if !ok {
		return
	}
	tzparam, _ := strconv.ParseInt(c.Query("timeZone"), 10, 64)
	if (tzparam < startTimezone) || (tzparam > endTimezone) {
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

	data, err := h.stat.GetWalletStats(fromTime, toTime, walletAddr.Hex(), tzparam)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (h *HTTPServer) GetWalletAddresses(c *gin.Context) {
	data, err := h.stat.GetWalletAddresses()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (h *HTTPServer) GetReserveRate(c *gin.Context) {
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
	data, err := h.stat.GetReserveRates(fromTime, toTime, reserveAddr)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (h *HTTPServer) GetExchangesStatus(c *gin.Context) {
	data, err := h.app.GetExchangeStatus()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (h *HTTPServer) UpdateExchangeStatus(c *gin.Context) {
	postForm, ok := h.Authenticated(c, []string{"exchange", "status", "timestamp"}, []Permission{ConfirmConfPermission})
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
	err = h.app.UpdateExchangeStatus(exchange, status, timestamp)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (h *HTTPServer) GetCountryStats(c *gin.Context) {
	fromTime, toTime, ok := h.ValidateTimeInput(c)
	if !ok {
		return
	}
	country := c.Query("country")
	tzparam, _ := strconv.ParseInt(c.Query("timeZone"), 10, 64)
	if (tzparam < startTimezone) || (tzparam > endTimezone) {
		httputil.ResponseFailure(c, httputil.WithReason("Timezone is not supported"))
		return
	}
	data, err := h.stat.GetGeoData(fromTime, toTime, country, tzparam)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (h *HTTPServer) GetHeatMap(c *gin.Context) {
	fromTime, toTime, ok := h.ValidateTimeInput(c)
	if !ok {
		return
	}
	tzparam, _ := strconv.ParseInt(c.Query("timeZone"), 10, 64)
	if (tzparam < startTimezone) || (tzparam > endTimezone) {
		httputil.ResponseFailure(c, httputil.WithReason("Timezone is not supported"))
		return
	}

	data, err := h.stat.GetHeatMap(fromTime, toTime, tzparam)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (h *HTTPServer) GetCountries(c *gin.Context) {
	data, _ := h.stat.GetCountries()
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (h *HTTPServer) UpdatePriceAnalyticData(c *gin.Context) {
	postForm, ok := h.Authenticated(c, []string{}, []Permission{RebalancePermission})
	if !ok {
		return
	}
	timestamp, err := strconv.ParseUint(postForm.Get("timestamp"), 10, 64)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	value := []byte(postForm.Get("value"))
	if len(value) > maxDataSize {
		httputil.ResponseFailure(c, httputil.WithReason(errDataSizeExceed.Error()))
		return
	}
	err = h.stat.UpdatePriceAnalyticData(timestamp, value)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}
func (h *HTTPServer) GetPriceAnalyticData(c *gin.Context) {
	_, ok := h.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, ConfigurePermission, ConfirmConfPermission, RebalancePermission})
	if !ok {
		return
	}
	fromTime, toTime, ok := h.ValidateTimeInput(c)
	if !ok {
		return
	}
	if toTime == 0 {
		toTime = common.GetTimepoint()
	}

	data, err := h.stat.GetPriceAnalyticData(fromTime, toTime)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (h *HTTPServer) ExchangeNotification(c *gin.Context) {
	postForm, ok := h.Authenticated(c, []string{
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

	err := h.app.UpdateExchangeNotification(exchange, action, tokenPair, from, to, isWarning, msg)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (h *HTTPServer) GetNotifications(c *gin.Context) {
	_, ok := h.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, RebalancePermission, ConfigurePermission, ConfirmConfPermission})
	if !ok {
		return
	}
	data, err := h.app.GetNotifications()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (h *HTTPServer) GetUserList(c *gin.Context) {
	_, ok := h.Authenticated(c, []string{"fromTime", "toTime", "timeZone"}, []Permission{ReadOnlyPermission, RebalancePermission, ConfigurePermission, ConfirmConfPermission})
	if !ok {
		return
	}
	fromTime, toTime, ok := h.ValidateTimeInput(c)
	if !ok {
		return
	}
	timeZone, err := strconv.ParseInt(c.Query("timeZone"), 10, 64)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("timeZone is required: %s", err.Error())))
		return
	}
	data, err := h.stat.GetUserList(fromTime, toTime, timeZone)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (h *HTTPServer) GetReserveVolume(c *gin.Context) {
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
	token, err := h.setting.GetActiveTokenByID(tokenName)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	data, err := h.stat.GetReserveVolume(fromTime, toTime, freq, reserveAddr, token.Address)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (h *HTTPServer) SetStableTokenParams(c *gin.Context) {
	postForm, ok := h.Authenticated(c, []string{}, []Permission{ConfigurePermission})
	if !ok {
		return
	}
	value := []byte(postForm.Get("value"))
	if len(value) > maxDataSize {
		httputil.ResponseFailure(c, httputil.WithReason(errDataSizeExceed.Error()))
		return
	}
	err := h.metric.SetStableTokenParams(value)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (h *HTTPServer) ConfirmStableTokenParams(c *gin.Context) {
	postForm, ok := h.Authenticated(c, []string{}, []Permission{ConfirmConfPermission})
	if !ok {
		return
	}
	value := []byte(postForm.Get("value"))
	if len(value) > maxDataSize {
		httputil.ResponseFailure(c, httputil.WithReason(errDataSizeExceed.Error()))
		return
	}
	err := h.metric.ConfirmStableTokenParams(value)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (h *HTTPServer) RejectStableTokenParams(c *gin.Context) {
	_, ok := h.Authenticated(c, []string{}, []Permission{ConfirmConfPermission})
	if !ok {
		return
	}
	err := h.metric.RemovePendingStableTokenParams()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (h *HTTPServer) GetPendingStableTokenParams(c *gin.Context) {
	_, ok := h.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, ConfigurePermission, ConfirmConfPermission, RebalancePermission})
	if !ok {
		return
	}

	data, err := h.metric.GetPendingStableTokenParams()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (h *HTTPServer) GetStableTokenParams(c *gin.Context) {
	_, ok := h.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, ConfigurePermission, ConfirmConfPermission, RebalancePermission})
	if !ok {
		return
	}

	data, err := h.metric.GetStableTokenParams()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (h *HTTPServer) GetTokenHeatmap(c *gin.Context) {
	fromTime, toTime, ok := h.ValidateTimeInput(c)
	if !ok {
		return
	}
	freq := c.Query("freq")
	token := c.Query("token")
	if token == "" {
		httputil.ResponseFailure(c, httputil.WithReason("token param is required"))
		return
	}

	data, err := h.stat.GetTokenHeatmap(fromTime, toTime, token, freq)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

//SetTargetQtyV2 set token target quantity version 2
func (h *HTTPServer) SetTargetQtyV2(c *gin.Context) {
	postForm, ok := h.Authenticated(c, []string{}, []Permission{ConfigurePermission})
	if !ok {
		return
	}
	value := []byte(postForm.Get("value"))
	if len(value) > maxDataSize {
		httputil.ResponseFailure(c, httputil.WithReason(errDataSizeExceed.Error()))
		return
	}
	var tokenTargetQty common.TokenTargetQtyV2
	if err := json.Unmarshal(value, &tokenTargetQty); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}

	for tokenID := range tokenTargetQty {
		if _, err := h.setting.GetInternalTokenByID(tokenID); err != nil {
			httputil.ResponseFailure(c, httputil.WithError(err))
			return
		}
	}

	err := h.metric.StorePendingTargetQtyV2(value)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (h *HTTPServer) GetPendingTargetQtyV2(c *gin.Context) {
	_, ok := h.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, ConfigurePermission, ConfirmConfPermission, RebalancePermission})
	if !ok {
		return
	}

	data, err := h.metric.GetPendingTargetQtyV2()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (h *HTTPServer) ConfirmTargetQtyV2(c *gin.Context) {
	postForm, ok := h.Authenticated(c, []string{}, []Permission{ConfirmConfPermission})
	if !ok {
		return
	}
	value := []byte(postForm.Get("value"))
	if len(value) > maxDataSize {
		httputil.ResponseFailure(c, httputil.WithReason(errDataSizeExceed.Error()))
		return
	}
	err := h.metric.ConfirmTargetQtyV2(value)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	}
	httputil.ResponseSuccess(c)
}

func (h *HTTPServer) CancelTargetQtyV2(c *gin.Context) {
	_, ok := h.Authenticated(c, []string{}, []Permission{ConfirmConfPermission})
	if !ok {
		return
	}
	err := h.metric.RemovePendingTargetQtyV2()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (h *HTTPServer) GetTargetQtyV2(c *gin.Context) {
	_, ok := h.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, ConfigurePermission, ConfirmConfPermission, RebalancePermission})
	if !ok {
		return
	}

	data, err := h.metric.GetTargetQtyV2()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (h *HTTPServer) GetFeeSetRateByDay(c *gin.Context) {
	fromTime, toTime, ok := h.ValidateTimeInput(c)
	if !ok {
		return
	}
	data, err := h.stat.GetFeeSetRateByDay(fromTime, toTime)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithReason(err.Error()))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (h *HTTPServer) register() {
	stt := h.r.Group("/setting")
	stt.POST("/set-token-update", h.SetTokenUpdate)
	stt.GET("/pending-token-update", h.GetPendingTokenUpdates)
	stt.POST("/confirm-token-update", h.ConfirmTokenUpdate)
	stt.POST("/reject-token-update", h.RejectTokenUpdate)
	stt.GET("/token-settings", h.TokenSettings)
	stt.POST("/update-address", h.UpdateAddress)
	stt.POST("/add-address-to-set", h.AddAddressToSet)
	stt.POST("/update-exchange-fee", h.UpdateExchangeFee)
	stt.POST("/update-exchange-mindeposit", h.UpdateExchangeMinDeposit)
	stt.POST("/update-deposit-address", h.UpdateDepositAddress)
	stt.POST("/update-exchange-info", h.UpdateExchangeInfo)
	stt.GET("/all-settings", h.GetAllSetting)

	if h.core != nil && h.app != nil {
		v2 := h.r.Group("/v2")

		h.r.GET("/prices-version", h.AllPricesVersion)
		h.r.GET("/prices", h.AllPrices)
		h.r.GET("/prices/:base/:quote", h.Price)
		h.r.GET("/getrates", h.GetRate)
		h.r.GET("/get-all-rates", h.GetRates)

		h.r.GET("/authdata-version", h.AuthDataVersion)
		h.r.GET("/authdata", h.AuthData)
		h.r.GET("/activities", h.GetActivities)
		h.r.GET("/immediate-pending-activities", h.ImmediatePendingActivities)
		h.r.GET("/metrics", h.Metrics)
		h.r.POST("/metrics", h.StoreMetrics)

		h.r.POST("/cancelorder/:exchangeid", h.CancelOrder)
		h.r.POST("/deposit/:exchangeid", h.Deposit)
		h.r.POST("/withdraw/:exchangeid", h.Withdraw)
		h.r.POST("/trade/:exchangeid", h.Trade)
		h.r.POST("/setrates", h.SetRate)
		h.r.GET("/exchangeinfo", h.GetExchangeInfo)
		h.r.GET("/exchangefees", h.GetFee)
		h.r.GET("/exchange-min-deposit", h.GetMinDeposit)
		h.r.GET("/tradehistory", h.GetTradeHistory)

		v2.GET("/targetqty", h.GetTargetQtyV2)
		v2.GET("/pendingtargetqty", h.GetPendingTargetQtyV2)
		v2.POST("/settargetqty", h.SetTargetQtyV2)
		v2.POST("/confirmtargetqty", h.ConfirmTargetQtyV2)
		v2.POST("/canceltargetqty", h.CancelTargetQtyV2)

		h.r.GET("/timeserver", h.GetTimeServer)

		h.r.GET("/rebalancestatus", h.GetRebalanceStatus)
		h.r.POST("/holdrebalance", h.HoldRebalance)
		h.r.POST("/enablerebalance", h.EnableRebalance)

		h.r.GET("/setratestatus", h.GetSetrateStatus)
		h.r.POST("/holdsetrate", h.HoldSetrate)
		h.r.POST("/enablesetrate", h.EnableSetrate)

		v2.GET("/pwis-equation", h.GetPWIEquationV2)
		v2.GET("/pending-pwis-equation", h.GetPendingPWIEquationV2)
		v2.POST("/set-pwis-equation", h.SetPWIEquationV2)
		v2.POST("/confirm-pwis-equation", h.ConfirmPWIEquationV2)
		v2.POST("/reject-pwis-equation", h.RejectPWIEquationV2)

		h.r.GET("/rebalance-quadratic", h.GetRebalanceQuadratic)
		h.r.GET("/pending-rebalance-quadratic", h.GetPendingRebalanceQuadratic)
		h.r.POST("/set-rebalance-quadratic", h.SetRebalanceQuadratic)
		h.r.POST("/confirm-rebalance-quadratic", h.ConfirmRebalanceQuadratic)
		h.r.POST("/reject-rebalance-quadratic", h.RejectRebalanceQuadratic)

		h.r.GET("/get-exchange-status", h.GetExchangesStatus)
		h.r.POST("/update-exchange-status", h.UpdateExchangeStatus)

		h.r.POST("/exchange-notification", h.ExchangeNotification)
		h.r.GET("/exchange-notifications", h.GetNotifications)

		h.r.POST("/set-stable-token-params", h.SetStableTokenParams)
		h.r.POST("/confirm-stable-token-params", h.ConfirmStableTokenParams)
		h.r.POST("/reject-stable-token-params", h.RejectStableTokenParams)
		h.r.GET("/pending-stable-token-params", h.GetPendingStableTokenParams)
		h.r.GET("/stable-token-params", h.GetStableTokenParams)

		h.r.GET("/gold-feed", h.GetGoldData)
	}

	if h.stat != nil {
		h.r.GET("/cap-by-address/:addr", h.GetCapByAddress)
		h.r.GET("/cap-by-user/:user", h.GetCapByUser)
		h.r.GET("/richguy/:addr", h.ExceedDailyLimit)
		h.r.GET("/tradelogs", h.TradeLogs)
		h.r.GET("/catlogs", h.CatLogs)
		h.r.GET("/get-asset-volume", h.GetAssetVolume)
		h.r.GET("/get-burn-fee", h.GetBurnFee)
		h.r.GET("/get-wallet-fee", h.GetWalletFee)
		h.r.GET("/get-user-volume", h.GetUserVolume)
		h.r.GET("/get-users-volume", h.GetUsersVolume)
		h.r.GET("/get-trade-summary", h.GetTradeSummary)
		h.r.POST("/update-user-addresses", h.UpdateUserAddresses)
		h.r.GET("/get-pending-addresses", h.GetPendingAddresses)
		h.r.GET("/get-reserve-rate", h.GetReserveRate)
		h.r.GET("/get-wallet-stats", h.GetWalletStats)
		h.r.GET("/get-wallet-address", h.GetWalletAddresses)
		h.r.GET("/get-country-stats", h.GetCountryStats)
		h.r.GET("/get-heat-map", h.GetHeatMap)
		h.r.GET("/get-countries", h.GetCountries)
		h.r.POST("/update-price-analytic-data", h.UpdatePriceAnalyticData)
		h.r.GET("/get-price-analytic-data", h.GetPriceAnalyticData)
		h.r.GET("/get-reserve-volume", h.GetReserveVolume)
		h.r.GET("/get-user-list", h.GetUserList)
		h.r.GET("/get-token-heatmap", h.GetTokenHeatmap)
		h.r.GET("/get-fee-setrate", h.GetFeeSetRateByDay)
	}
}

func (h *HTTPServer) Run() {
	h.register()
	if err := h.r.Run(h.host); err != nil {
		log.Panic(err)
	}
}

func NewHTTPServer(
	app reserve.ReserveData,
	core reserve.ReserveCore,
	stat reserve.ReserveStats,
	metric metric.MetricStorage,
	host string,
	enableAuth bool,
	authEngine Authentication,
	env string,
	bc *blockchain.Blockchain,
	setting Setting) *HTTPServer {
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
		app, core, stat, metric, host, enableAuth, authEngine, r, bc, setting,
	}
}
