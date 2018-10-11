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
	"github.com/getsentry/raven-go"
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
	blockchain  Blockchain
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

// CheckRequiredParams signed message (message = url encoded both query params and post params, keys are sorted) in "signed" header
// using HMAC512
// params must contain "nonce" which is the unixtime in millisecond. The nonce will be invalid
// if it differs from server time more than 10s
func (self *HTTPServer) CheckRequiredParams(c *gin.Context, requiredParams []string) (url.Values, bool) {
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

	for _, p := range requiredParams {
		if params.Get(p) == "" {
			httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("Required param (%s) is missing. Param name is case sensitive", p)))
			return c.Request.Form, false
		}
	}

	return params, false
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
	pair, err := self.setting.NewTokenPairFromID(base, quote)
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

	data, err := self.app.CurrentAuthDataVersion(getTimePoint(c, true))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithField("version", data))
	}
}

func (self *HTTPServer) AuthData(c *gin.Context) {
	log.Printf("Getting current auth data snapshot \n")

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
		toTime = maxTimespot
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
	postForm, ok := self.CheckRequiredParams(c, []string{"tokens", "buys", "sells", "block", "afp_mid", "msgs"})
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
		token, err := self.setting.GetInternalTokenByID(tok)
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
	id, err := self.core.SetRates(tokens, bigBuys, bigSells, big.NewInt(intBlock), bigAfpMid, msgs)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithField("id", id))
}

func (self *HTTPServer) Trade(c *gin.Context) {
	postForm, ok := self.CheckRequiredParams(c, []string{"base", "quote", "amount", "rate", "type"})
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
	base, err := self.setting.GetInternalTokenByID(baseTokenParam)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	quote, err := self.setting.GetInternalTokenByID(quoteTokenParam)
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
	postForm, ok := self.CheckRequiredParams(c, []string{"order_id"})
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
	postForm, ok := self.CheckRequiredParams(c, []string{"token", "amount"})
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
	token, err := self.setting.GetInternalTokenByID(tokenParam)
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
	postForm, ok := self.CheckRequiredParams(c, []string{"amount", "token"})
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
	token, err := self.setting.GetInternalTokenByID(tokenParam)
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
	data, err := self.app.GetPendingActivities()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithData(data))
	}
}

func (self *HTTPServer) Metrics(c *gin.Context) {
	response := common.MetricResponse{
		Timestamp: common.GetTimepoint(),
	}
	log.Printf("Getting metrics")
	postForm, ok := self.CheckRequiredParams(c, []string{"tokens", "from", "to"})
	if !ok {
		return
	}
	tokenParam := postForm.Get("tokens")
	fromParam := postForm.Get("from")
	toParam := postForm.Get("to")
	tokens := []common.Token{}
	for _, tok := range strings.Split(tokenParam, "-") {
		token, err := self.setting.GetInternalTokenByID(tok)
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
	postForm, ok := self.CheckRequiredParams(c, []string{"timestamp", "data"})
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

	err = self.metric.StoreMetric(&metricEntry, common.GetTimepoint())
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c)
	}
}

//ValidateExchangeInfo validate if data is complete exchange info with all token pairs supported
// func ValidateExchangeInfo(exchange common.Exchange, data map[common.TokenPairID]common.ExchangePrecisionLimit) error {
// 	exInfo, err :=self
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
func (self *HTTPServer) GetExchangeInfo(c *gin.Context) {
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

func (self *HTTPServer) GetFee(c *gin.Context) {
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

func (self *HTTPServer) GetMinDeposit(c *gin.Context) {
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

func (self *HTTPServer) GetTradeHistory(c *gin.Context) {
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
	data, err := self.metric.GetRebalanceControl()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data.Status))
}

func (self *HTTPServer) HoldRebalance(c *gin.Context) {
	if err := self.metric.StoreRebalanceControl(false); err != nil {
		httputil.ResponseFailure(c, httputil.WithReason(err.Error()))
		return
	}
	httputil.ResponseSuccess(c)
	return
}

func (self *HTTPServer) EnableRebalance(c *gin.Context) {
	if err := self.metric.StoreRebalanceControl(true); err != nil {
		httputil.ResponseFailure(c, httputil.WithReason(err.Error()))
	}
	httputil.ResponseSuccess(c)
	return
}

func (self *HTTPServer) GetSetrateStatus(c *gin.Context) {
	data, err := self.metric.GetSetrateControl()
	if err != nil {
		httputil.ResponseFailure(c)
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data.Status))
}

func (self *HTTPServer) HoldSetrate(c *gin.Context) {
	if err := self.metric.StoreSetrateControl(false); err != nil {
		httputil.ResponseFailure(c, httputil.WithReason(err.Error()))
	}
	httputil.ResponseSuccess(c)
	return
}

func (self *HTTPServer) EnableSetrate(c *gin.Context) {
	if err := self.metric.StoreSetrateControl(true); err != nil {
		httputil.ResponseFailure(c, httputil.WithReason(err.Error()))
	}
	httputil.ResponseSuccess(c)
	return
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
	if (tzparam < startTimezone) || (tzparam > endTimezone) {
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

func (self *HTTPServer) GetCapByAddress(c *gin.Context) {
	addr := c.Param("addr")
	address := ethereum.HexToAddress(addr)
	if address.Big().Cmp(ethereum.Big0) == 0 {
		httputil.ResponseFailure(c, httputil.WithReason("address is not valid"))
		return
	}
	data, kyced, err := self.stat.GetTxCapByAddress(address)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithMultipleFields(
			gin.H{
				"data": data,
				"kyc":  kyced,
			},
		))
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

func (self *HTTPServer) GetWalletStats(c *gin.Context) {
	fromTime, toTime, ok := self.ValidateTimeInput(c)
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

	data, err := self.stat.GetWalletStats(fromTime, toTime, walletAddr.Hex(), tzparam)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (self *HTTPServer) GetWalletAddresses(c *gin.Context) {
	data, err := self.stat.GetWalletAddresses()
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
	postForm, ok := self.CheckRequiredParams(c, []string{"exchange", "status", "timestamp"})
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
	if (tzparam < startTimezone) || (tzparam > endTimezone) {
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
	if (tzparam < startTimezone) || (tzparam > endTimezone) {
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
	postForm, ok := self.CheckRequiredParams(c, []string{})
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
	err = self.stat.UpdatePriceAnalyticData(timestamp, value)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}
func (self *HTTPServer) GetPriceAnalyticData(c *gin.Context) {
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
	postForm, ok := self.CheckRequiredParams(c, []string{
		"exchange", "action", "token", "fromTime", "toTime", "isWarning"})
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
	data, err := self.app.GetNotifications()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (self *HTTPServer) GetUserList(c *gin.Context) {
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
	tokenID := c.Query("token")
	if tokenID == "" {
		httputil.ResponseFailure(c, httputil.WithReason("token is required"))
		return
	}

	data, err := self.stat.GetReserveVolume(fromTime, toTime, freq, reserveAddr, tokenID)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (self *HTTPServer) SetStableTokenParams(c *gin.Context) {
	postForm, ok := self.CheckRequiredParams(c, []string{})
	if !ok {
		return
	}
	value := []byte(postForm.Get("value"))
	if len(value) > maxDataSize {
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
	postForm, ok := self.CheckRequiredParams(c, []string{})
	if !ok {
		return
	}
	value := []byte(postForm.Get("value"))
	if len(value) > maxDataSize {
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
	_, ok := self.CheckRequiredParams(c, []string{})
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
	data, err := self.metric.GetPendingStableTokenParams()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (self *HTTPServer) GetStableTokenParams(c *gin.Context) {
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

//SetTargetQtyV2 set token target quantity version 2
func (self *HTTPServer) SetTargetQtyV2(c *gin.Context) {
	postForm, ok := self.CheckRequiredParams(c, []string{})
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
		if _, err := self.setting.GetInternalTokenByID(tokenID); err != nil {
			httputil.ResponseFailure(c, httputil.WithError(err))
			return
		}
	}

	err := self.metric.StorePendingTargetQtyV2(value)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (self *HTTPServer) GetPendingTargetQtyV2(c *gin.Context) {
	data, err := self.metric.GetPendingTargetQtyV2()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (self *HTTPServer) ConfirmTargetQtyV2(c *gin.Context) {
	postForm, ok := self.CheckRequiredParams(c, []string{})
	if !ok {
		return
	}
	value := []byte(postForm.Get("value"))
	if len(value) > maxDataSize {
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
	err := self.metric.RemovePendingTargetQtyV2()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (self *HTTPServer) GetTargetQtyV2(c *gin.Context) {
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
		//Configure group
		configure := self.r.Group("/")
		{
			configure.POST("setting/set-token-update", self.SetTokenUpdate)
			configure.POST("setting/update-exchange-fee", self.UpdateExchangeFee)
			configure.POST("setting/update-exchange-mindeposit", self.UpdateExchangeMinDeposit)
			configure.POST("setting/update-deposit-address", self.UpdateDepositAddress)
			configure.POST("setting/update-exchange-info", self.UpdateExchangeInfo)

			configure.POST("v2/settargetqty", self.SetTargetQtyV2)
			configure.POST("v2/set-pwis-equation", self.SetPWIEquationV2)

			configure.POST("set-rebalance-quadratic", self.SetRebalanceQuadratic)
			configure.POST("set-stable-token-params", self.SetStableTokenParams)
		}

		confirm := self.r.Group("/")
		{
			confirm.POST("setting/confirm-token-update", self.ConfirmTokenUpdate)
			confirm.POST("setting/reject-token-update", self.RejectTokenUpdate)

			confirm.POST("v2/confirmtargetqty", self.ConfirmTargetQtyV2)
			confirm.POST("v2/canceltargetqty", self.CancelTargetQtyV2)
			confirm.POST("v2/confirm-pwis-equation", self.ConfirmPWIEquationV2)
			confirm.POST("v2/reject-pwis-equation", self.RejectPWIEquationV2)

			confirm.POST("holdrebalance", self.HoldRebalance)
			confirm.POST("enablerebalance", self.EnableRebalance)
			confirm.POST("holdsetrate", self.HoldSetrate)
			confirm.POST("enablesetrate", self.EnableSetrate)
			confirm.POST("confirm-rebalance-quadratic", self.ConfirmRebalanceQuadratic)
			confirm.POST("reject-rebalance-quadratic", self.RejectRebalanceQuadratic)
			confirm.POST("update-exchange-status", self.UpdateExchangeStatus)
			confirm.POST("confirm-stable-token-params", self.ConfirmStableTokenParams)
			confirm.POST("reject-stable-token-params", self.RejectStableTokenParams)
		}

		//Read group
		read := self.r.Group("/")
		{
			read.GET("setting/pending-token-update", self.GetPendingTokenUpdates)
			read.GET("setting/token-settings", self.TokenSettings)
			read.GET("setting/all-settings", self.GetAllSetting)
			read.GET("setting/internal-tokens", self.GetInternalTokens)
			read.GET("setting/active-tokens", self.GetActiveTokens)
			read.GET("setting/token-by-address", self.GetTokenByAddress)
			read.GET("setting/active-token-by-id", self.GetActiveTokenByID)
			read.GET("setting/address", self.GetAddress)
			read.GET("setting/addresses", self.GetAddresses)
			read.GET("setting/ping", self.ReadyToServe)

			read.GET("authdata-version", self.AuthDataVersion)
			read.GET("authdata", self.AuthData)
			read.GET("activities", self.GetActivities)
			read.GET("immediate-pending-activities", self.ImmediatePendingActivities)
			read.GET("metrics", self.Metrics)
			read.GET("tradehistory", self.GetTradeHistory)
			read.GET("rebalancestatus", self.GetRebalanceStatus)
			read.GET("rebalance-quadratic", self.GetRebalanceQuadratic)
			read.GET("pending-rebalance-quadratic", self.GetPendingRebalanceQuadratic)
			read.GET("setratestatus", self.GetSetrateStatus)
			read.GET("exchange-notifications", self.GetNotifications)
			read.GET("get-step-function-data", self.GetStepFunctionData)
			read.GET("pending-stable-token-params", self.GetPendingStableTokenParams)
			read.GET("stable-token-params", self.GetStableTokenParams)

			read.GET("v2/targetqty", self.GetTargetQtyV2)
			read.GET("v2/pendingtargetqty", self.GetPendingTargetQtyV2)
			read.GET("v2/pwis-equation", self.GetPWIEquationV2)
			read.GET("v2/pending-pwis-equation", self.GetPendingPWIEquationV2)
		}

		//Rebalance group
		rebalance := self.r.Group("/")
		{
			rebalance.POST("metrics", self.StoreMetrics)
			rebalance.POST("cancelorder/:exchangeid", self.CancelOrder)
			rebalance.POST("deposit/:exchangeid", self.Deposit)
			rebalance.POST("withdraw/:exchangeid", self.Withdraw)
			rebalance.POST("trade/:exchangeid", self.Trade)
			rebalance.POST("setrates", self.SetRate)
			rebalance.POST("exchange-notification", self.ExchangeNotification)
		}

		//No auth
		self.r.GET("/prices-version", self.AllPricesVersion)
		self.r.GET("/prices", self.AllPrices)
		self.r.GET("/prices/:base/:quote", self.Price)
		self.r.GET("/getrates", self.GetRate)
		self.r.GET("/get-all-rates", self.GetRates)
		self.r.GET("/exchangeinfo", self.GetExchangeInfo)
		self.r.GET("/exchangefees", self.GetFee)
		self.r.GET("/exchange-min-deposit", self.GetMinDeposit)
		self.r.GET("/timeserver", self.GetTimeServer)
		self.r.GET("/gold-feed", self.GetGoldData)
		self.r.GET("/get-exchange-status", self.GetExchangesStatus)

		if self.authEnabled{
			readPermissionsCheck, err := self.auth.ReadPermissionCheck()
			if err != nil {
				panic(err)
			}
			read.Use(readPermissionsCheck)

			configurePermissionCheck, err := self.auth.ConfigurePermissionCheck()
			if err != nil {
				panic(err)
			}
			configure.Use(configurePermissionCheck)

			confirmPermissionCheck, err := self.auth.ConfirmPermissionCheck()
			if err != nil {
				panic(err)
			}
			confirm.Use(confirmPermissionCheck)

			rebalancePermissionCheck, err := self.auth.RebalancePermissionCheck()
			if err != nil {
				panic(err)
			}
			rebalance.Use(rebalancePermissionCheck)
		}
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
		self.r.GET("/get-wallet-address", self.GetWalletAddresses)
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
	if err := self.r.Run(self.host); err != nil {
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
