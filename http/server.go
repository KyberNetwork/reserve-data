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

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/getsentry/raven-go"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sentry"
	"github.com/gin-gonic/gin"

	"github.com/KyberNetwork/reserve-data"
	"github.com/KyberNetwork/reserve-data/cmd/deployment"
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/http/httputil"
	"github.com/KyberNetwork/reserve-data/metric"
)

const (
	maxTimespot uint64 = 18446744073709551615
	maxDataSize int    = 1000000 //1 Megabyte in byte
)

var (
	// errDataSizeExceed is returned when the post data is larger than maxDataSize.
	errDataSizeExceed = errors.New("the data size must be less than 1 MB")
)

type Server struct {
	app                 reserve.Data
	core                reserve.Core
	metric              metric.Storage
	host                string
	authEnabled         bool
	auth                Authentication
	r                   *gin.Engine
	blockchain          Blockchain
	setting             Setting
	contractAddressConf *common.ContractAddressConfiguration
}

func getTimePoint(c *gin.Context, useDefault bool) uint64 {
	timestamp := c.DefaultQuery("timestamp", "")
	if timestamp == "" {
		if useDefault {
			log.Printf("Interpreted timestamp to default - %d\n", maxTimespot)
			return maxTimespot
		}
		timepoint := common.GetTimepoint()
		log.Printf("Interpreted timestamp to current time - %d\n", timepoint)
		return timepoint
	}
	timepoint, err := strconv.ParseUint(timestamp, 10, 64)
	if err != nil {
		log.Printf("Interpreted timestamp(%s) to default - %d", timestamp, maxTimespot)
		return maxTimespot
	}
	log.Printf("Interpreted timestamp(%s) to %d", timestamp, timepoint)
	return timepoint
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
func (s *Server) Authenticated(c *gin.Context, requiredParams []string, perms []Permission) (url.Values, bool) {
	err := c.Request.ParseForm()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("Malformed request package: %s", err.Error())))
		return c.Request.Form, false
	}

	if !s.authEnabled {
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
	userPerms := s.auth.GetPermission(signed, message)
	if eligible(userPerms, perms) {
		return params, true
	}
	if len(userPerms) == 0 {
		httputil.ResponseFailure(c, httputil.WithReason("Invalid signed token"))
	} else {
		httputil.ResponseFailure(c, httputil.WithReason("You don't have permission to proceed"))
	}
	return params, false
}

func (s *Server) AllPricesVersion(c *gin.Context) {
	log.Printf("Getting all prices version")
	data, err := s.app.CurrentPriceVersion(getTimePoint(c, true))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithField("version", data))
	}
}

func (s *Server) AllPrices(c *gin.Context) {
	log.Printf("Getting all prices \n")
	data, err := s.app.GetAllPrices(getTimePoint(c, true))
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

func (s *Server) Price(c *gin.Context) {
	base := c.Param("base")
	quote := c.Param("quote")
	log.Printf("Getting price for %s - %s \n", base, quote)
	pair, err := s.setting.NewTokenPairFromID(base, quote)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithReason("Token pair is not supported"))
	} else {
		data, err := s.app.GetOnePrice(pair.PairID(), getTimePoint(c, true))
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

func (s *Server) AuthDataVersion(c *gin.Context) {
	log.Printf("Getting current auth data snapshot version")
	_, ok := s.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, RebalancePermission, ConfigurePermission, ConfirmConfPermission})
	if !ok {
		return
	}

	data, err := s.app.CurrentAuthDataVersion(getTimePoint(c, true))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithField("version", data))
	}
}

func (s *Server) AuthData(c *gin.Context) {
	log.Printf("Getting current auth data snapshot \n")
	_, ok := s.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, RebalancePermission, ConfigurePermission, ConfirmConfPermission})
	if !ok {
		return
	}

	data, err := s.app.GetAuthData(getTimePoint(c, true))
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

func (s *Server) GetRates(c *gin.Context) {
	log.Printf("Getting all rates \n")
	fromTime, _ := strconv.ParseUint(c.Query("fromTime"), 10, 64)
	toTime, _ := strconv.ParseUint(c.Query("toTime"), 10, 64)
	if toTime == 0 {
		toTime = maxTimespot
	}
	data, err := s.app.GetRates(fromTime, toTime)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithData(data))
	}
}

func (s *Server) GetRate(c *gin.Context) {
	log.Printf("Getting all rates \n")
	data, err := s.app.GetRate(getTimePoint(c, true))
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

func (s *Server) SetRate(c *gin.Context) {
	postForm, ok := s.Authenticated(c, []string{"tokens", "buys", "sells", "block", "afp_mid", "msgs"}, []Permission{RebalancePermission})
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
		token, err := s.setting.GetInternalTokenByID(tok)
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
	id, err := s.core.SetRates(tokens, bigBuys, bigSells, big.NewInt(intBlock), bigAfpMid, msgs)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithField("id", id))
}

func (s *Server) Trade(c *gin.Context) {
	postForm, ok := s.Authenticated(c, []string{"base", "quote", "amount", "rate", "type"}, []Permission{RebalancePermission})
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
	base, err := s.setting.GetInternalTokenByID(baseTokenParam)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	quote, err := s.setting.GetInternalTokenByID(quoteTokenParam)
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
	id, done, remaining, finished, err := s.core.Trade(
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

func (s *Server) CancelOrder(c *gin.Context) {
	postForm, ok := s.Authenticated(c, []string{"order_id"}, []Permission{RebalancePermission})
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
	err = s.core.CancelOrder(activityID, exchange)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (s *Server) Withdraw(c *gin.Context) {
	postForm, ok := s.Authenticated(c, []string{"token", "amount"}, []Permission{RebalancePermission})
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
	token, err := s.setting.GetInternalTokenByID(tokenParam)
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
	id, err := s.core.Withdraw(exchange, token, amount, getTimePoint(c, false))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithField("id", id))
}

func (s *Server) Deposit(c *gin.Context) {
	postForm, ok := s.Authenticated(c, []string{"amount", "token"}, []Permission{RebalancePermission})
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
	token, err := s.setting.GetInternalTokenByID(tokenParam)
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
	id, err := s.core.Deposit(exchange, token, amount, getTimePoint(c, false))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithField("id", id))
}

func (s *Server) GetActivities(c *gin.Context) {
	log.Printf("Getting all activity records \n")
	_, ok := s.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, RebalancePermission, ConfigurePermission, ConfirmConfPermission})
	if !ok {
		return
	}
	fromTime, _ := strconv.ParseUint(c.Query("fromTime"), 10, 64)
	toTime, _ := strconv.ParseUint(c.Query("toTime"), 10, 64)
	if toTime == 0 {
		toTime = common.GetTimepoint()
	}

	data, err := s.app.GetRecords(fromTime*1000000, toTime*1000000)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithData(data))
	}
}

func (s *Server) StopFetcher(c *gin.Context) {
	err := s.app.Stop()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c)
	}
}

func (s *Server) ImmediatePendingActivities(c *gin.Context) {
	log.Printf("Getting all immediate pending activity records \n")
	_, ok := s.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, RebalancePermission, ConfigurePermission, ConfirmConfPermission})
	if !ok {
		return
	}

	data, err := s.app.GetPendingActivities()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithData(data))
	}
}

func (s *Server) Metrics(c *gin.Context) {
	response := common.MetricResponse{
		Timestamp: common.GetTimepoint(),
	}
	log.Printf("Getting metrics")
	postForm, ok := s.Authenticated(c, []string{"tokens", "from", "to"}, []Permission{ReadOnlyPermission, RebalancePermission, ConfigurePermission, ConfirmConfPermission})
	if !ok {
		return
	}
	tokenParam := postForm.Get("tokens")
	fromParam := postForm.Get("from")
	toParam := postForm.Get("to")
	tokens := []common.Token{}
	for _, tok := range strings.Split(tokenParam, "-") {
		token, err := s.setting.GetInternalTokenByID(tok)
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
	data, err := s.metric.GetMetric(tokens, from, to)
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

func (s *Server) StoreMetrics(c *gin.Context) {
	log.Printf("Storing metrics")
	postForm, ok := s.Authenticated(c, []string{"timestamp", "data"}, []Permission{RebalancePermission})
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

	err = s.metric.StoreMetric(&metricEntry, common.GetTimepoint())
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
func (s *Server) GetExchangeInfo(c *gin.Context) {
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

func (s *Server) GetFee(c *gin.Context) {
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
}

func (s *Server) GetMinDeposit(c *gin.Context) {
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
}

func (s *Server) GetTradeHistory(c *gin.Context) {
	_, ok := s.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, RebalancePermission, ConfigurePermission, ConfirmConfPermission})
	if !ok {
		return
	}
	fromTime, toTime, ok := s.ValidateTimeInput(c)
	if !ok {
		return
	}
	data, err := s.app.GetTradeHistory(fromTime, toTime)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (s *Server) GetTimeServer(c *gin.Context) {
	httputil.ResponseSuccess(c, httputil.WithData(common.GetTimestamp()))
}

func (s *Server) GetRebalanceStatus(c *gin.Context) {
	_, ok := s.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, RebalancePermission, ConfigurePermission, ConfirmConfPermission})
	if !ok {
		return
	}
	data, err := s.metric.GetRebalanceControl()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data.Status))
}

func (s *Server) HoldRebalance(c *gin.Context) {
	_, ok := s.Authenticated(c, []string{}, []Permission{ConfirmConfPermission})
	if !ok {
		return
	}
	if err := s.metric.StoreRebalanceControl(false); err != nil {
		httputil.ResponseFailure(c, httputil.WithReason(err.Error()))
		return
	}
	httputil.ResponseSuccess(c)
}

func (s *Server) EnableRebalance(c *gin.Context) {
	_, ok := s.Authenticated(c, []string{}, []Permission{ConfirmConfPermission})
	if !ok {
		return
	}
	if err := s.metric.StoreRebalanceControl(true); err != nil {
		httputil.ResponseFailure(c, httputil.WithReason(err.Error()))
	}
	httputil.ResponseSuccess(c)
}

func (s *Server) GetSetrateStatus(c *gin.Context) {
	_, ok := s.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, RebalancePermission, ConfigurePermission, ConfirmConfPermission})
	if !ok {
		return
	}
	data, err := s.metric.GetSetrateControl()
	if err != nil {
		httputil.ResponseFailure(c)
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data.Status))
}

func (s *Server) HoldSetrate(c *gin.Context) {
	_, ok := s.Authenticated(c, []string{}, []Permission{ConfirmConfPermission})
	if !ok {
		return
	}
	if err := s.metric.StoreSetrateControl(false); err != nil {
		httputil.ResponseFailure(c, httputil.WithReason(err.Error()))
	}
	httputil.ResponseSuccess(c)
}

func (s *Server) EnableSetrate(c *gin.Context) {
	_, ok := s.Authenticated(c, []string{}, []Permission{ConfirmConfPermission})
	if !ok {
		return
	}
	if err := s.metric.StoreSetrateControl(true); err != nil {
		httputil.ResponseFailure(c, httputil.WithReason(err.Error()))
	}
	httputil.ResponseSuccess(c)
}

func (s *Server) ValidateTimeInput(c *gin.Context) (uint64, uint64, bool) {
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

func (s *Server) GetExchangesStatus(c *gin.Context) {
	data, err := s.app.GetExchangeStatus()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (s *Server) UpdateExchangeStatus(c *gin.Context) {
	postForm, ok := s.Authenticated(c, []string{"exchange", "status", "timestamp"}, []Permission{ConfirmConfPermission})
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
	err = s.app.UpdateExchangeStatus(exchange, status, timestamp)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (s *Server) ExchangeNotification(c *gin.Context) {
	postForm, ok := s.Authenticated(c, []string{
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

	err := s.app.UpdateExchangeNotification(exchange, action, tokenPair, from, to, isWarning, msg)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (s *Server) GetNotifications(c *gin.Context) {
	_, ok := s.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, RebalancePermission, ConfigurePermission, ConfirmConfPermission})
	if !ok {
		return
	}
	data, err := s.app.GetNotifications()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (s *Server) SetStableTokenParams(c *gin.Context) {
	postForm, ok := s.Authenticated(c, []string{}, []Permission{ConfigurePermission})
	if !ok {
		return
	}
	value := []byte(postForm.Get("value"))
	if len(value) > maxDataSize {
		httputil.ResponseFailure(c, httputil.WithReason(errDataSizeExceed.Error()))
		return
	}
	err := s.metric.SetStableTokenParams(value)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (s *Server) ConfirmStableTokenParams(c *gin.Context) {
	postForm, ok := s.Authenticated(c, []string{}, []Permission{ConfirmConfPermission})
	if !ok {
		return
	}
	value := []byte(postForm.Get("value"))
	if len(value) > maxDataSize {
		httputil.ResponseFailure(c, httputil.WithReason(errDataSizeExceed.Error()))
		return
	}
	err := s.metric.ConfirmStableTokenParams(value)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (s *Server) RejectStableTokenParams(c *gin.Context) {
	_, ok := s.Authenticated(c, []string{}, []Permission{ConfirmConfPermission})
	if !ok {
		return
	}
	err := s.metric.RemovePendingStableTokenParams()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (s *Server) GetPendingStableTokenParams(c *gin.Context) {
	_, ok := s.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, ConfigurePermission, ConfirmConfPermission, RebalancePermission})
	if !ok {
		return
	}

	data, err := s.metric.GetPendingStableTokenParams()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (s *Server) GetStableTokenParams(c *gin.Context) {
	_, ok := s.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, ConfigurePermission, ConfirmConfPermission, RebalancePermission})
	if !ok {
		return
	}

	data, err := s.metric.GetStableTokenParams()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

//SetTargetQtyV2 set token target quantity version 2
func (s *Server) SetTargetQtyV2(c *gin.Context) {
	postForm, ok := s.Authenticated(c, []string{}, []Permission{ConfigurePermission})
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
		if _, err := s.setting.GetInternalTokenByID(tokenID); err != nil {
			err = fmt.Errorf("TokenID: %s, error: %s", tokenID, err)
			httputil.ResponseFailure(c, httputil.WithError(err))
			return
		}
	}

	err := s.metric.StorePendingTargetQtyV2(value)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (s *Server) GetPendingTargetQtyV2(c *gin.Context) {
	_, ok := s.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, ConfigurePermission, ConfirmConfPermission, RebalancePermission})
	if !ok {
		return
	}

	data, err := s.metric.GetPendingTargetQtyV2()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (s *Server) ConfirmTargetQtyV2(c *gin.Context) {
	postForm, ok := s.Authenticated(c, []string{}, []Permission{ConfirmConfPermission})
	if !ok {
		return
	}
	value := []byte(postForm.Get("value"))
	if len(value) > maxDataSize {
		httputil.ResponseFailure(c, httputil.WithReason(errDataSizeExceed.Error()))
		return
	}
	err := s.metric.ConfirmTargetQtyV2(value)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	}
	httputil.ResponseSuccess(c)
}

func (s *Server) CancelTargetQtyV2(c *gin.Context) {
	_, ok := s.Authenticated(c, []string{}, []Permission{ConfirmConfPermission})
	if !ok {
		return
	}
	err := s.metric.RemovePendingTargetQtyV2()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (s *Server) GetTargetQtyV2(c *gin.Context) {
	_, ok := s.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, ConfigurePermission, ConfirmConfPermission, RebalancePermission})
	if !ok {
		return
	}

	data, err := s.metric.GetTargetQtyV2()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (s *Server) register() {

	if s.core != nil && s.app != nil {
		stt := s.r.Group("/setting")
		stt.POST("/set-token-update", s.SetTokenUpdate)
		stt.GET("/pending-token-update", s.GetPendingTokenUpdates)
		stt.POST("/confirm-token-update", s.ConfirmTokenUpdate)
		stt.POST("/reject-token-update", s.RejectTokenUpdate)
		stt.GET("/token-settings", s.TokenSettings)
		stt.POST("/update-exchange-fee", s.UpdateExchangeFee)
		stt.POST("/update-exchange-mindeposit", s.UpdateExchangeMinDeposit)
		stt.POST("/update-deposit-address", s.UpdateDepositAddress)
		stt.POST("/update-exchange-info", s.UpdateExchangeInfo)
		stt.GET("/all-settings", s.GetAllSetting)
		stt.GET("/internal-tokens", s.GetInternalTokens)
		stt.GET("/active-tokens", s.GetActiveTokens)
		stt.GET("/token-by-address", s.GetTokenByAddress)
		stt.GET("/active-token-by-id", s.GetActiveTokenByID)
		stt.GET("/address", s.GetAddress)
		stt.GET("/ping", s.ReadyToServe)
		v2 := s.r.Group("/v2")

		s.r.GET("/prices-version", s.AllPricesVersion)
		s.r.GET("/prices", s.AllPrices)
		s.r.GET("/prices/:base/:quote", s.Price)
		s.r.GET("/getrates", s.GetRate)
		s.r.GET("/get-all-rates", s.GetRates)

		s.r.GET("/authdata-version", s.AuthDataVersion)
		s.r.GET("/authdata", s.AuthData)
		s.r.GET("/activities", s.GetActivities)
		s.r.GET("/immediate-pending-activities", s.ImmediatePendingActivities)
		s.r.GET("/metrics", s.Metrics)
		s.r.POST("/metrics", s.StoreMetrics)

		s.r.POST("/cancelorder/:exchangeid", s.CancelOrder)
		s.r.POST("/deposit/:exchangeid", s.Deposit)
		s.r.POST("/withdraw/:exchangeid", s.Withdraw)
		s.r.POST("/trade/:exchangeid", s.Trade)
		s.r.POST("/setrates", s.SetRate)
		s.r.GET("/exchangeinfo", s.GetExchangeInfo)
		s.r.GET("/exchangefees", s.GetFee)
		s.r.GET("/exchange-min-deposit", s.GetMinDeposit)
		s.r.GET("/tradehistory", s.GetTradeHistory)

		v2.GET("/targetqty", s.GetTargetQtyV2)
		v2.GET("/pendingtargetqty", s.GetPendingTargetQtyV2)
		v2.POST("/settargetqty", s.SetTargetQtyV2)
		v2.POST("/confirmtargetqty", s.ConfirmTargetQtyV2)
		v2.POST("/canceltargetqty", s.CancelTargetQtyV2)

		s.r.GET("/timeserver", s.GetTimeServer)

		s.r.GET("/rebalancestatus", s.GetRebalanceStatus)
		s.r.POST("/holdrebalance", s.HoldRebalance)
		s.r.POST("/enablerebalance", s.EnableRebalance)

		s.r.GET("/setratestatus", s.GetSetrateStatus)
		s.r.POST("/holdsetrate", s.HoldSetrate)
		s.r.POST("/enablesetrate", s.EnableSetrate)

		v2.GET("/pwis-equation", s.GetPWIEquationV2)
		v2.GET("/pending-pwis-equation", s.GetPendingPWIEquationV2)
		v2.POST("/set-pwis-equation", s.SetPWIEquationV2)
		v2.POST("/confirm-pwis-equation", s.ConfirmPWIEquationV2)
		v2.POST("/reject-pwis-equation", s.RejectPWIEquationV2)

		s.r.GET("/rebalance-quadratic", s.GetRebalanceQuadratic)
		s.r.GET("/pending-rebalance-quadratic", s.GetPendingRebalanceQuadratic)
		s.r.POST("/set-rebalance-quadratic", s.SetRebalanceQuadratic)
		s.r.POST("/confirm-rebalance-quadratic", s.ConfirmRebalanceQuadratic)
		s.r.POST("/reject-rebalance-quadratic", s.RejectRebalanceQuadratic)

		s.r.GET("/get-exchange-status", s.GetExchangesStatus)
		s.r.POST("/update-exchange-status", s.UpdateExchangeStatus)

		s.r.POST("/exchange-notification", s.ExchangeNotification)
		s.r.GET("/exchange-notifications", s.GetNotifications)

		s.r.POST("/set-stable-token-params", s.SetStableTokenParams)
		s.r.POST("/confirm-stable-token-params", s.ConfirmStableTokenParams)
		s.r.POST("/reject-stable-token-params", s.RejectStableTokenParams)
		s.r.GET("/pending-stable-token-params", s.GetPendingStableTokenParams)
		s.r.GET("/stable-token-params", s.GetStableTokenParams)

		s.r.GET("/gold-feed", s.GetGoldData)
		s.r.GET("/btc-feed", s.GetBTCData)
		s.r.POST("/set-feed-configuration", s.UpdateFeedConfiguration)
		s.r.GET("/get-feed-configuration", s.GetFeedConfiguration)

		s.r.POST("/set-fetcher-configuration", s.UpdateFetcherConfiguration)
		s.r.GET("/get-all-fetcher-configuration", s.GetAllFetcherConfiguration)
	}
}

func (s *Server) Run() {
	s.register()
	if err := s.r.Run(s.host); err != nil {
		log.Panic(err)
	}
}

func NewHTTPServer(
	app reserve.Data,
	core reserve.Core,
	metric metric.Storage,
	host string,
	enableAuth bool,
	authEngine Authentication,
	dpl deployment.Deployment,
	bc Blockchain,
	setting Setting,
	contractAddressConf *common.ContractAddressConfiguration) *Server {
	r := gin.Default()
	sentryCli, err := raven.NewWithTags(
		"https://bf15053001464a5195a81bc41b644751:eff41ac715114b20b940010208271b13@sentry.io/228067",
		map[string]string{
			"env": dpl.String(),
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

	return &Server{
		app:                 app,
		core:                core,
		metric:              metric,
		host:                host,
		authEnabled:         enableAuth,
		auth:                authEngine,
		r:                   r,
		blockchain:          bc,
		setting:             setting,
		contractAddressConf: contractAddressConf,
	}
}
