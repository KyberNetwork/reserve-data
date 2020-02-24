package http

import (
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"time"

	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/getsentry/raven-go"
	"github.com/gin-contrib/pprof"
	"github.com/gin-contrib/sentry"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data"
	"github.com/KyberNetwork/reserve-data/cmd/deployment"
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/http/httputil"
	"github.com/KyberNetwork/reserve-data/lib/caller"
	v3common "github.com/KyberNetwork/reserve-data/reservesetting/common"
	"github.com/KyberNetwork/reserve-data/reservesetting/storage"
)

const (
	maxTimespot      uint64 = 4124425154000
	defaultTimeRange uint64 = 86400000
)

// Server struct for http package
type Server struct {
	app            reserve.Data
	core           reserve.Core
	host           string
	r              *gin.Engine
	blockchain     Blockchain
	settingStorage storage.Interface
	l              *zap.SugaredLogger
}

func getTimePoint(c *gin.Context, useDefault bool, l *zap.SugaredLogger) uint64 {
	timestamp := c.DefaultQuery("timestamp", "")
	if timestamp == "" {
		if useDefault {
			l.Debugw("Interpreted timestamp to default", "maxTimespot", maxTimespot)
			return maxTimespot
		}
		timepoint := common.NowInMillis()
		l.Debugw("Interpreted timestamp to current time", "timepoint", timepoint)
		return timepoint
	}
	timepoint, err := strconv.ParseUint(timestamp, 10, 64)
	if err != nil {
		l.Debugw("Interpreted timestamp to default", "timestamp", timestamp, "maxTimespot", maxTimespot)
		return maxTimespot
	}
	l.Debugw("Interpreted timestamp", "timestamp", timestamp, "timepoint", timepoint)
	return timepoint
}

// AllPricesVersion return current version of all token
func (s *Server) AllPricesVersion(c *gin.Context) {
	s.l.Infow("Getting all prices version")
	data, err := s.app.CurrentPriceVersion(getTimePoint(c, true, s.l))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithField("version", data))
	}
}

type price struct {
	Base     uint64              `json:"base"`
	Quote    uint64              `json:"quote"`
	Exchange uint64              `json:"exchange"`
	Bids     []common.PriceEntry `json:"bids"`
	Asks     []common.PriceEntry `json:"asks"`
}

// AllPrices return prices of all tokens
func (s *Server) AllPrices(c *gin.Context) {
	s.l.Infow("Getting all prices")
	data, err := s.app.GetAllPrices(getTimePoint(c, true, s.l))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}

	var responseData []price
	for tp, onePrice := range data.Data {
		pair, err := s.settingStorage.GetTradingPair(tp)
		if err != nil {
			httputil.ResponseFailure(c, httputil.WithError(err))
			return
		}
		for exchangeID, exchangePrice := range onePrice {
			responseData = append(responseData, price{
				Base:     pair.Base,
				Quote:    pair.Quote,
				Exchange: uint64(exchangeID),
				Bids:     exchangePrice.Bids,
				Asks:     exchangePrice.Asks,
			})
		}
	}

	httputil.ResponseSuccess(c, httputil.WithMultipleFields(gin.H{
		"version":   data.Version,
		"timestamp": data.Timestamp,
		"data":      responseData,
		"block":     data.Block,
	}))

}

// Price return price of a token
func (s *Server) Price(c *gin.Context) {
	base := c.Param("base")
	quote := c.Param("quote")
	s.l.Infow("Getting price", "base", base, "quote", quote)
	// TODO: change getting price to accept asset id
	//pair, err := s.setting.NewTokenPairFromID(base, quote)
	//if err != nil {
	//	httputil.ResponseFailure(c, httputil.WithReason("Token pair is not supported"))
	//} else {
	//	data, err := s.app.GetOnePrice(pair.PairID(), getTimePoint(c, true))
	//	if err != nil {
	//		httputil.ResponseFailure(c, httputil.WithError(err))
	//	} else {
	//		httputil.ResponseSuccess(c, httputil.WithMultipleFields(gin.H{
	//			"version":   data.Version,
	//			"timestamp": data.Timestamp,
	//			"exchanges": data.Data,
	//		}))
	//	}
	//}
}

// AuthDataVersion return current version of auth data
func (s *Server) AuthDataVersion(c *gin.Context) {
	s.l.Infow("Getting current auth data snapshot version")
	data, err := s.app.CurrentAuthDataVersion(getTimePoint(c, true, s.l))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithField("version", data))
	}
}

// AuthData return current auth data
func (s *Server) AuthData(c *gin.Context) {
	s.l.Infow("Getting current auth data snapshot \n")
	data, err := s.app.GetAuthData(getTimePoint(c, true, s.l))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithMultipleFields(gin.H{
			"version": data.Version,
			"data":    data,
		}))
	}
}

// GetRates return all rates
func (s *Server) GetRates(c *gin.Context) {
	s.l.Infow("Getting all rates")
	fromTime, _ := strconv.ParseUint(c.Query("fromTime"), 10, 64)
	toTime, _ := strconv.ParseUint(c.Query("toTime"), 10, 64)
	if toTime == 0 {
		toTime = common.TimeToMillis(time.Now())
	}
	if fromTime == 0 {
		fromTime = toTime - defaultTimeRange
	}
	data, err := s.app.GetRates(fromTime, toTime)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithData(data))
	}
}

// GetRate return rate of a token
func (s *Server) GetRate(c *gin.Context) {
	s.l.Infow("Getting all rates")
	data, err := s.app.GetRate(getTimePoint(c, true, s.l))
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

// TradeRequest form
type TradeRequest struct {
	Pair   uint64  `json:"pair"`
	Amount float64 `json:"amount"`
	Rate   float64 `json:"rate"`
	Type   string  `json:"type"`
}

// Trade create an order in cexs
func (s *Server) Trade(c *gin.Context) {
	var request TradeRequest
	var err error
	if err := c.ShouldBindJSON(&request); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	var pair v3common.TradingPairSymbols
	pair, err = s.settingStorage.GetTradingPair(request.Pair)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}

	if request.Type != "sell" && request.Type != "buy" {
		httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("Trade type of %s is not supported.", request.Type)))
		return
	}

	exchange, ok := common.SupportedExchanges[common.ExchangeID(pair.ExchangeID)]
	if !ok {
		httputil.ResponseFailure(c, httputil.WithError(errors.Errorf("exchange %v is not supported", pair.ExchangeID)))
		return
	}

	id, done, remaining, finished, err := s.core.Trade(
		exchange, request.Type, pair, request.Rate, request.Amount)
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

// OpenOrdersRequest request for open orders
type OpenOrdersRequest struct {
	ExchangeID uint64 `json:"exchange_id" binding:"required"`
	Pair       uint64 `json:"pair" binding:"required"`
}

// OpenOrders request for open orders
func (s *Server) OpenOrders(c *gin.Context) {
	var (
		logger = s.l.With("func", caller.GetCurrentFunctionName())
		query  OpenOrdersRequest
	)
	if err := c.ShouldBindQuery(&query); err != nil {
		logger.Errorw("query is is not correct format", "error", err)
		httputil.ResponseFailure(c, httputil.WithError(err))
	}

	exchange, ok := common.SupportedExchanges[common.ExchangeID(query.ExchangeID)]
	if !ok {
		httputil.ResponseFailure(c, httputil.WithError(errors.Errorf("exchange %v is not supported", query.ExchangeID)))
		return
	}

	pair, err := s.settingStorage.GetTradingPair(query.Pair)
	if err != nil {
		logger.Errorw("failed to get trading token pair from setting data base", "err", err)
		httputil.ResponseFailure(c, httputil.WithError(err))
	}
	logger.Infow("getting open orders for pair", "base", pair.BaseSymbol, "quote", pair.QuoteSymbol)

	openOrders, err := exchange.OpenOrders(pair)
	if err != nil {
		logger.Errorw("failed to get open orders", "exchange", exchange.ID().String, "base", pair.BaseSymbol, "quote", pair.QuoteSymbol)
		httputil.ResponseFailure(c, httputil.WithError(err))
	}

	httputil.ResponseSuccess(c, httputil.WithData(openOrders))
}

// CancelOrderRequest type
type CancelOrderRequest struct {
	Exchange string `json:"exchange"`
	OrderID  string `json:"order_id"`
}

// CancelOrder cancel an order from cexs
func (s *Server) CancelOrder(c *gin.Context) {
	var request CancelOrderRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}

	exchange, err := common.GetExchange(request.Exchange)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	s.l.Infow("Cancel order", "id", request.OrderID, "from", exchange.ID().String())
	activityID, err := common.StringToActivityID(request.OrderID)
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

// WithdrawRequest type
type WithdrawRequest struct {
	ExchangeID uint64   `json:"exchange"`
	Asset      uint64   `json:"asset"`
	Amount     *big.Int `json:"amount"`
}

// Withdraw asset to reserve from cex
func (s *Server) Withdraw(c *gin.Context) {
	var request WithdrawRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}

	exchange, ok := common.SupportedExchanges[common.ExchangeID(request.ExchangeID)]
	if !ok {
		httputil.ResponseFailure(c, httputil.WithError(errors.Errorf("exchange %v is not supported", request.ExchangeID)))
		return
	}

	asset, err := s.settingStorage.GetAsset(request.Asset)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	s.l.Infow("Withdraw", "amount", request.Amount.Text(10), "asset_id", asset.ID,
		"asset_symbol", asset.Symbol, "exchange", exchange.ID().String())
	id, err := s.core.Withdraw(exchange, asset, request.Amount)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithField("id", id))
}

// DepositRequest type
type DepositRequest struct {
	ExchangeID uint64   `json:"exchange"`
	Amount     *big.Int `json:"amount"`
	Asset      uint64   `json:"asset"`
}

// Deposit asset into cex
func (s *Server) Deposit(c *gin.Context) {
	var request DepositRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}

	exchange, ok := common.SupportedExchanges[common.ExchangeID(request.ExchangeID)]
	if !ok {
		httputil.ResponseFailure(c, httputil.WithError(errors.Errorf("exchange %v is not supported", request.ExchangeID)))
		return
	}

	asset, err := s.settingStorage.GetAsset(request.Asset)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}

	s.l.Infow("Depositing", "amount", request.Amount.Text(10), "asset_id", asset.ID,
		"asset_symbol", asset.Symbol, "exchange", exchange.ID().String())
	id, err := s.core.Deposit(exchange, asset, request.Amount, getTimePoint(c, false, s.l))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithField("id", id))
}

type getActivitiesRequest struct {
	FromTime uint64 `form:"fromTime" binding:"required"`
	ToTime   uint64 `form:"toTime" binding:"required"`
}

// GetActivities return all activities record
func (s *Server) GetActivities(c *gin.Context) {
	var query getActivitiesRequest
	s.l.Infow("Getting all activity records")
	if err := c.ShouldBindQuery(&query); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}

	if query.ToTime-query.FromTime > defaultTimeRange {
		httputil.ResponseFailure(c, httputil.WithReason("time range to big, it should be < 86400000 milisecond"))
		return
	}

	data, err := s.app.GetRecords(query.FromTime, query.ToTime)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithData(data))
	}
}

// StopFetcher stop fetcher from fetch data
func (s *Server) StopFetcher(c *gin.Context) {
	err := s.app.Stop()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c)
	}
}

// ImmediatePendingActivities return activities which are pending
func (s *Server) ImmediatePendingActivities(c *gin.Context) {
	s.l.Infow("Getting all immediate pending activity records")
	data, err := s.app.GetPendingActivities()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithData(data))
	}
}

// GetTradeHistory return trade history
func (s *Server) GetTradeHistory(c *gin.Context) {
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

// GetTimeServer return server time
func (s *Server) GetTimeServer(c *gin.Context) {
	httputil.ResponseSuccess(c, httputil.WithData(common.GetTimestamp()))
}

// ValidateTimeInput check if the params fromTime, toTime is valid or not
func (s *Server) ValidateTimeInput(c *gin.Context) (uint64, uint64, bool) {
	fromTime, ok := strconv.ParseUint(c.Query("fromTime"), 10, 64)
	if ok != nil {
		httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("fromTime param is invalid: %s", ok)))
		return 0, 0, false
	}
	toTime, _ := strconv.ParseUint(c.Query("toTime"), 10, 64)
	if toTime == 0 {
		toTime = common.NowInMillis()
	}
	return fromTime, toTime, true
}

func (s *Server) updateTokenIndice(c *gin.Context) {
	if err := s.blockchain.LoadAndSetTokenIndices(); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	}
	httputil.ResponseSuccess(c)
}

type checkTokenIndiceRequest struct {
	Address string `form:"address" binding:"required"`
}

func (s *Server) checkTokenIndice(c *gin.Context) {
	var (
		query checkTokenIndiceRequest
	)
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(
			http.StatusBadRequest,
			nil,
		)
		return
	}
	if err := s.blockchain.CheckTokenIndices(ethereum.HexToAddress(query.Address)); err != nil {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"error": err.Error(),
			},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		nil,
	)
}

func (s *Server) register() {
	if s.core != nil && s.app != nil {
		g := s.r.Group("/v3")
		g.GET("/prices-version", s.AllPricesVersion)
		g.GET("/prices", s.AllPrices)
		g.GET("/prices/:base/:quote", s.Price)
		g.GET("/getrates", s.GetRate)
		g.GET("/get-all-rates", s.GetRates)

		g.GET("/authdata-version", s.AuthDataVersion)
		g.GET("/authdata", s.AuthData)
		g.GET("/activities", s.GetActivities)
		g.GET("/immediate-pending-activities", s.ImmediatePendingActivities)

		g.GET("/open-orders", s.OpenOrders)
		g.POST("/cancelorder", s.CancelOrder)
		g.POST("/cancel-all-orders", s.CancelAllOrders)
		g.POST("/deposit", s.Deposit)
		g.POST("/withdraw", s.Withdraw)
		g.POST("/trade", s.Trade)
		g.POST("/setrates", s.SetRate)
		g.GET("/tradehistory", s.GetTradeHistory)

		g.GET("/timeserver", s.GetTimeServer)

		g.GET("/gold-feed", s.GetGoldData)
		g.GET("/btc-feed", s.GetBTCData)
		g.GET("/usd-feed", s.GetUSDData)

		g.GET("/addresses", s.GetAddresses)

		g.PUT("/update-token-indice", s.updateTokenIndice)
		g.GET("/check-token-indice", s.checkTokenIndice)
	}
}

// Run the server
func (s *Server) Run() {
	s.register()
	if err := s.r.Run(s.host); err != nil {
		log.Panic(err)
	}
}

// EnableProfiler enable profiler
func (s *Server) EnableProfiler() {
	pprof.Register(s.r)
}

// NewHTTPServer return new server
func NewHTTPServer(
	app reserve.Data,
	core reserve.Core,
	host string,
	dpl deployment.Deployment,
	bc Blockchain,
	settingStorage storage.Interface,
) *Server {
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

	return &Server{
		app:            app,
		core:           core,
		host:           host,
		r:              r,
		blockchain:     bc,
		settingStorage: settingStorage,
		l:              zap.S(),
	}
}
