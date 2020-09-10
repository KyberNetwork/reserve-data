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

	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/getsentry/raven-go"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-contrib/sentry"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data"
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/gasinfo"
	gaspricedataclient "github.com/KyberNetwork/reserve-data/common/gaspricedata-client"
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

// Server object
type Server struct {
	app            reserve.Data
	core           reserve.Core
	metric         metric.Storage
	bindAddr       string
	authEnabled    bool
	auth           Authentication
	profilerPrefix string
	r              *gin.Engine
	blockchain     Blockchain
	setting        Setting
	l              *zap.SugaredLogger
	gasInfo        *gasinfo.GasPriceInfo
}

func getTimePoint(c *gin.Context, useDefault bool) uint64 {
	timestamp := c.DefaultQuery("timestamp", "")
	l := zap.S()
	if timestamp == "" {
		if useDefault {
			l.Debugf("Interpreted timestamp to default - %d\n", maxTimespot)
			return maxTimespot
		}
		timepoint := common.GetTimepoint()
		l.Debugf("Interpreted timestamp to current time - %d\n", timepoint)
		return timepoint
	}
	timepoint, err := strconv.ParseUint(timestamp, 10, 64)
	if err != nil {
		l.Debugf("Interpreted timestamp(%s) to default - %d", timestamp, maxTimespot)
		return maxTimespot
	}
	l.Debugf("Interpreted timestamp(%s) to %d", timestamp, timepoint)
	return timepoint
}

// IsIntime check if request time is in range of 30s, otherwise the request is invalid
func IsIntime(l *zap.SugaredLogger, nonce string) bool {
	serverTime := common.GetTimepoint()
	nonceInt, err := strconv.ParseInt(nonce, 10, 64)
	if err != nil {
		l.Debugf("IsIntime returns false, err: %v", err)
		return false
	}
	difference := nonceInt - int64(serverTime)
	if difference < -30000 || difference > 30000 {
		l.Debugf("IsIntime returns false, nonce: %d, serverTime: %d, difference: %d", nonceInt, int64(serverTime), difference)
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
		httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("Malformed request package: %+v", err)))
		return c.Request.Form, false
	}

	if !s.authEnabled {
		return c.Request.Form, true
	}

	params := c.Request.Form
	s.l.Debugf("Form params: %s\n", params)
	if !IsIntime(s.l, params.Get("nonce")) {
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

// AllPricesVersion return current version all price
func (s *Server) AllPricesVersion(c *gin.Context) {
	s.l.Infof("Getting all prices version")
	data, err := s.app.CurrentPriceVersion(getTimePoint(c, true))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithField("version", data))
	}
}

// AllPrices return all prices of token
func (s *Server) AllPrices(c *gin.Context) {
	s.l.Infof("Getting all prices \n")
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

// Price return price for a certain pair of token
func (s *Server) Price(c *gin.Context) {
	base := c.Param("base")
	quote := c.Param("quote")
	s.l.Infof("Getting price for %s - %s", base, quote)
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

// AuthDataVersion return current version of auth data
func (s *Server) AuthDataVersion(c *gin.Context) {
	s.l.Infof("Getting current auth data snapshot version")
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

// AuthData return authenticated data
// include: reserve balance on blockchain
// reserve balance on centralized exchanges
// pending activities (set rates, buy, sell, deposit, withdraw)
func (s *Server) AuthData(c *gin.Context) {
	s.l.Infof("Getting current auth data snapshot \n")
	_, ok := s.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, RebalancePermission, ConfigurePermission, ConfirmConfPermission})
	if !ok {
		return
	}
	now := common.GetTimepoint()
	tp := getTimePoint(c, true)
	updateWindow := uint64(30000) // auth data get update every 10s, but we allow it get late at max 30s
	data, err := s.app.GetAuthData(tp)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithMultipleFields(gin.H{
			"version":   data.Version,
			"timestamp": data.Timestamp,
			"data":      data.Data,
		}))
		if now-uint64(data.Version) > updateWindow {
			s.l.Warnw("auth data not updated", "now", now, "version", data.Version, "requested_time_point", tp)
		}
	}
}

// GetRates return all rates
func (s *Server) GetRates(c *gin.Context) {
	s.l.Infof("Getting all rates")
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

// GetRate return all rates
func (s *Server) GetRate(c *gin.Context) {
	s.l.Infof("Getting all rates")
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

func tokenExisted(tokenAddr ethereum.Address, tokens []common.Token) bool {
	exist := false
	for _, token := range tokens {
		if ethereum.HexToAddress(token.Address) == tokenAddr {
			exist = true
			break
		}
	}
	return exist
}

func (s *Server) checkTokenDelisted(tokens []common.Token, bigBuys, bigSells, bigAfpMid []*big.Int) ([]common.Token, []*big.Int, []*big.Int, []*big.Int, error) {
	listedTokens := s.blockchain.ListedTokens()
	if len(listedTokens) <= len(tokens) {
		return tokens, bigBuys, bigSells, bigAfpMid, nil
	}

	for _, tokenAddr := range listedTokens {
		if !tokenExisted(tokenAddr, tokens) {
			tokens = append(tokens, common.Token{
				Address: tokenAddr.Hex(),
			})
			bigBuys = append(bigBuys, big.NewInt(0))
			bigSells = append(bigSells, big.NewInt(0))
			bigAfpMid = append(bigAfpMid, big.NewInt(0))
		}
	}

	return tokens, bigBuys, bigSells, bigAfpMid, nil
}

// SetRate call set rate token to blockchain
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
	var tokens []common.Token
	for _, tok := range strings.Split(tokenAddrs, "-") {
		token, err := s.setting.GetInternalTokenByID(tok)
		if err != nil {
			httputil.ResponseFailure(c, httputil.WithError(err))
			return
		}
		tokens = append(tokens, token)
	}
	var bigBuys []*big.Int
	for _, rate := range strings.Split(buys, "-") {
		r, err := hexutil.DecodeBig(rate)
		if err != nil {
			httputil.ResponseFailure(c, httputil.WithError(err))
			return
		}
		bigBuys = append(bigBuys, r)
	}
	var bigSells []*big.Int
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
	var bigAfpMid []*big.Int
	for _, rate := range strings.Split(afpMid, "-") {
		var r *big.Int
		if r, err = hexutil.DecodeBig(rate); err != nil {
			httputil.ResponseFailure(c, httputil.WithError(err))
			return
		}
		bigAfpMid = append(bigAfpMid, r)
	}
	tokens, bigBuys, bigSells, bigAfpMid, err = s.checkTokenDelisted(tokens, bigBuys, bigSells, bigAfpMid)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}

	id, err := s.core.SetRates(tokens, bigBuys, bigSells, big.NewInt(intBlock), bigAfpMid, msgs)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithField("id", id))
}

func (s *Server) cancelSetRate(c *gin.Context) {
	_, ok := s.Authenticated(c, []string{}, []Permission{RebalancePermission, ConfirmConfPermission})
	if !ok {
		return
	}
	id, err := s.core.CancelSetRate()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithField("id", id))
}

// Trade do trade action to centralize exchanges
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
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	s.l.Infof("http server: Trade: rate: %f, raw rate: %s", rate, rateParam)
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

// GetOpenOrders return open orders from exchange
func (s *Server) GetOpenOrders(c *gin.Context) {
	_, ok := s.Authenticated(c, nil, []Permission{RebalancePermission, ConfigurePermission, ConfirmConfPermission, ReadOnlyPermission})
	if !ok {
		return
	}

	exchangeParam := c.Query("exchange")
	exchanges := make(map[common.ExchangeID]common.Exchange)
	if exchangeParam != "" {
		exchange, err := common.GetExchange(exchangeParam)
		if err != nil {
			httputil.ResponseFailure(c, httputil.WithError(err))
			return
		}
		exchanges[common.ExchangeID(exchangeParam)] = exchange
	} else {
		exchanges = common.SupportedExchanges
	}
	result := make(map[common.ExchangeID][]common.Order)
	for exchangeID, exchange := range exchanges {
		orders, err := exchange.OpenOrders()
		if err != nil {
			httputil.ResponseFailure(c, httputil.WithError(err))
			return
		}
		result[exchangeID] = orders
	}
	httputil.ResponseSuccess(c, httputil.WithData(result))
}

// CancelOrder cancel an open order on exchanges
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
	s.l.Infof("Cancel order id: %s from %s\n", id, exchange.ID())
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

// Withdraw withdraw token from exchanges
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
	s.l.Infof("Withdraw %s %s from %s\n", amount.Text(10), token.ID, exchange.ID())
	id, err := s.core.Withdraw(exchange, token, amount, getTimePoint(c, false))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithField("id", id))
}

// Deposit token to exchange
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
	s.l.Infof("Depositing %s %s to %s\n", amount.Text(10), token.ID, exchange.ID())
	id, err := s.core.Deposit(exchange, token, amount, getTimePoint(c, false))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithField("id", id))
}

// GetActivities return all activities record
// in a time frame
func (s *Server) GetActivities(c *gin.Context) {
	s.l.Infof("Getting all activity records \n")
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

type gasStatusResult struct {
	GasPrice      gaspricedataclient.GasResult `json:"gas_price"`
	HighThreshold float64                      `json:"high"`
	LowThreshold  float64                      `json:"low"`
}

func (s *Server) GetGasStatus(c *gin.Context) {
	_, ok := s.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, RebalancePermission, ConfigurePermission, ConfirmConfPermission})
	if !ok {
		return
	}
	gasPrice, err := s.gasInfo.AllSourceGas()
	if err != nil {
		s.l.Errorw("query gas price failed", "err", err)
		httputil.ResponseFailure(c, httputil.WithReason(err.Error()))
		return
	}
	gasThreshold, err := s.app.GetGasThreshold()
	if err != nil {
		s.l.Errorw("failed to get gas threshold", "err", err)
		httputil.ResponseFailure(c, httputil.WithReason(err.Error()))
		return
	}
	result := gasStatusResult{
		GasPrice:      gasPrice,
		HighThreshold: gasThreshold.High,
		LowThreshold:  gasThreshold.Low,
	}
	httputil.ResponseSuccess(c, httputil.WithData(result))
}

func (s *Server) SetGasThreshold(c *gin.Context) {
	high := c.Request.FormValue("high")
	low := c.Request.FormValue("low")
	_, ok := s.Authenticated(c, []string{"high", "low"}, []Permission{ConfirmConfPermission})
	if !ok {
		return
	}
	h, err := strconv.ParseFloat(high, 64)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithReason("invalid high value"))
		return
	}
	l, err := strconv.ParseFloat(low, 64)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithReason("invalid low value"))
		return
	}
	if l >= h {
		httputil.ResponseFailure(c, httputil.WithReason("high must > low value"))
		return
	}
	if err := s.app.SetGasThreshold(common.GasThreshold{High: h, Low: l}); err != nil {
		httputil.ResponseFailure(c, httputil.WithReason(err.Error()))
		return
	}
	httputil.ResponseSuccess(c)
}

func (s *Server) SetPreferGasSource(c *gin.Context) {
	name := c.Request.FormValue("name")
	_, ok := s.Authenticated(c, []string{"name"}, []Permission{ConfirmConfPermission})
	if !ok {
		return
	}
	if name == "" {
		httputil.ResponseFailure(c, httputil.WithReason("name is required"))
	}
	if err := s.app.SetPreferGasSource(common.PreferGasSource{Name: name}); err != nil {
		httputil.ResponseFailure(c, httputil.WithReason(err.Error()))
		return
	}
	httputil.ResponseSuccess(c)
}

func (s *Server) GetPreferGasSource(c *gin.Context) {
	_, ok := s.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, RebalancePermission, ConfigurePermission, ConfirmConfPermission})
	if !ok {
		return
	}
	preferredGasSource, err := s.app.GetPreferGasSource()
	if err != nil {
		s.l.Errorw("failed to get prefered gas source", "err", err)
		httputil.ResponseFailure(c, httputil.WithReason(err.Error()))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(preferredGasSource))
}

// StopFetcher request to stop fetcher
func (s *Server) StopFetcher(c *gin.Context) {
	err := s.app.Stop()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c)
	}
}

// ImmediatePendingActivities return current pending activities
func (s *Server) ImmediatePendingActivities(c *gin.Context) {
	s.l.Infof("Getting all immediate pending activity records \n")
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

// Metrics return metrics
func (s *Server) Metrics(c *gin.Context) {
	response := common.MetricResponse{
		Timestamp: common.GetTimepoint(),
	}
	s.l.Infof("Getting metrics")
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

// StoreMetrics store token metrics
func (s *Server) StoreMetrics(c *gin.Context) {
	s.l.Infof("Storing metrics")
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

// GetFee return centralized exchanges fee config
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

// GetMinDeposit return min deposit config of centralized echanges
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

// GetTradeHistory return trade history in centralized exchanges
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

// GetTimeServer return current time server
func (s *Server) GetTimeServer(c *gin.Context) {
	httputil.ResponseSuccess(c, httputil.WithData(common.GetTimestamp()))
}

// GetRebalanceStatus return rebalance configuration status (enabled, disabled)
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

// HoldRebalance disable rebalance - notify analytics to stop sending rebalance request
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

// EnableRebalance enable rebalance request
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

// GetSetrateStatus return set rate status configuration (enabled, disabled)
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

// HoldSetrate turn setrate config into disabled
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

// EnableSetrate turn set rate configuration to enabled
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

// ValidateTimeInput validate from-to time value
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

// GetExchangesStatus return exchange status (enabled, disabled)
// analytics component will only request for enabled exchanges
func (s *Server) GetExchangesStatus(c *gin.Context) {
	data, err := s.app.GetExchangeStatus()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

// UpdateExchangeStatus update exchange status (enable, disable)
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

// ExchangeNotification get exchange notification config
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

// GetNotifications get notifications
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

// SetStableTokenParams set stable token params
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

// ConfirmStableTokenParams confirm change to stable token params
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

// RejectStableTokenParams reject request changes stable token params
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

// GetPendingStableTokenParams return pending change stable token params
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

// GetStableTokenParams return stable token params
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

// GetPendingTargetQtyV2 get pending change target quantity
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

// ConfirmTargetQtyV2 confirm change target quantity
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

// CancelTargetQtyV2 cancel update target quantity request
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

func (s *Server) getVersion(c *gin.Context) {
	httputil.ResponseSuccess(c, httputil.WithField("version", common.AppVersion))
}

// GetTargetQtyV2 return target quantity with v2 format
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

		s.r.GET("/open-orders", s.GetOpenOrders)
		s.r.POST("/cancelorder/:exchangeid", s.CancelOrder)
		s.r.POST("/cancel-order-by-order-id", s.CancelOrderByOrderID)
		s.r.POST("/cancel-all-orders", s.CancelAllOrders)

		s.r.POST("/deposit/:exchangeid", s.Deposit)
		s.r.POST("/withdraw/:exchangeid", s.Withdraw)
		s.r.POST("/trade/:exchangeid", s.Trade)
		s.r.POST("/setrates", s.SetRate)
		s.r.POST("/cancel-setrates", s.cancelSetRate)
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
		s.r.GET("/usd-feed", s.GetUSDData)
		s.r.POST("/set-feed-configuration", s.UpdateFeedConfiguration)
		s.r.GET("/get-feed-configuration", s.GetFeedConfiguration)

		v2.POST("/set-feed-setting", s.SetFeedSetting)
		v2.POST("/confirm-feed-setting", s.ConfirmPendingFeedSetting)
		v2.POST("/reject-feed-setting", s.RejectPendingFeedSetting)
		v2.GET("/pending-feed-setting", s.GetPendingFeedSetting)
		v2.GET("/feed-setting", s.GetFeedSetting)

		s.r.POST("/set-fetcher-configuration", s.UpdateFetcherConfiguration)
		s.r.GET("/get-all-fetcher-configuration", s.GetAllFetcherConfiguration)
		s.r.GET("/version", s.getVersion)
		s.r.GET("/gas-threshold", s.GetGasStatus)
		s.r.POST("/gas-threshold", s.SetGasThreshold)
		s.r.GET("/gas-source", s.GetPreferGasSource)
		s.r.POST("/gas-source", s.SetPreferGasSource)
	}
}

// Run the server
func (s *Server) Run() {
	s.register()
	if len(s.profilerPrefix) != 0 {
		pprof.Register(s.r, s.profilerPrefix)
	}
	if err := s.r.Run(s.bindAddr); err != nil {
		log.Panic(err)
	}
}

// NewHTTPServer create new server instance
func NewHTTPServer(app reserve.Data, core reserve.Core, metric metric.Storage, bindAddr string, enableAuth bool,
	profilerPrefix string, authEngine Authentication, env string, bc Blockchain, setting Setting, client *gasinfo.GasPriceInfo) *Server {
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

	s := &Server{
		app:            app,
		core:           core,
		metric:         metric,
		bindAddr:       bindAddr,
		authEnabled:    enableAuth,
		auth:           authEngine,
		profilerPrefix: profilerPrefix,
		r:              r,
		blockchain:     bc,
		setting:        setting,
		l:              zap.S(),
		gasInfo:        client,
	}

	return s
}
