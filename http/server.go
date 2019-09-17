package http

import (
	"fmt"
	"log"
	"math/big"
	"strconv"

	"github.com/getsentry/raven-go"
	"github.com/gin-contrib/pprof"
	"github.com/gin-contrib/sentry"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/KyberNetwork/reserve-data"
	"github.com/KyberNetwork/reserve-data/cmd/deployment"
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/http/httputil"
	v3common "github.com/KyberNetwork/reserve-data/reservesetting/common"
	"github.com/KyberNetwork/reserve-data/reservesetting/storage"
)

const (
	maxTimespot uint64 = 4124425154000
)

// Server struct for http package
type Server struct {
	app            reserve.Data
	core           reserve.Core
	host           string
	r              *gin.Engine
	blockchain     Blockchain
	settingStorage storage.Interface
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

// AllPricesVersion return current version of all token
func (s *Server) AllPricesVersion(c *gin.Context) {
	log.Printf("Getting all prices version")
	data, err := s.app.CurrentPriceVersion(getTimePoint(c, true))
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
	log.Printf("Getting all prices \n")
	data, err := s.app.GetAllPrices(getTimePoint(c, true))
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
	log.Printf("Getting price for %s - %s \n", base, quote)
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
	log.Printf("Getting current auth data snapshot version")
	data, err := s.app.CurrentAuthDataVersion(getTimePoint(c, true))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithField("version", data))
	}
}

// AuthData return current auth data
func (s *Server) AuthData(c *gin.Context) {
	log.Printf("Getting current auth data snapshot \n")
	data, err := s.app.GetAuthData(getTimePoint(c, true))
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

// GetRate return rate of a token
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
	log.Printf("Cancel order id: %s from %s\n", request.OrderID, exchange.ID().String())
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
	log.Printf("Withdraw %s id=%d symbol=%v from %s\n", request.Amount.Text(10), asset.ID, asset.Symbol, exchange.ID().String())
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

	log.Printf("Depositing %s id=%d symbol=%v to %s\n", request.Amount.Text(10), asset.ID, asset.Symbol, exchange.ID().String())
	id, err := s.core.Deposit(exchange, asset, request.Amount, getTimePoint(c, false))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithField("id", id))
}

// GetActivities return all activities record
func (s *Server) GetActivities(c *gin.Context) {
	log.Printf("Getting all activity records \n")
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
	log.Printf("Getting all immediate pending activity records \n")
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
		toTime = common.GetTimepoint()
	}
	return fromTime, toTime, true
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
		g.POST("/set-feed-configuration", s.UpdateFeedConfiguration)
		g.GET("/get-feed-configuration", s.GetFeedConfiguration)

		g.GET("/addresses", s.GetAddresses)

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
	return &Server{
		app:            app,
		core:           core,
		host:           host,
		r:              r,
		blockchain:     bc,
		settingStorage: settingStorage,
	}
}
