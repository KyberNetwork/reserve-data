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
	"github.com/KyberNetwork/reserve-data/common/gasinfo"
	"github.com/KyberNetwork/reserve-data/exchange"
	"github.com/KyberNetwork/reserve-data/exchange/binance"
	"github.com/KyberNetwork/reserve-data/http/httputil"
	"github.com/KyberNetwork/reserve-data/lib/caller"
	"github.com/KyberNetwork/reserve-data/lib/rtypes"
	v3common "github.com/KyberNetwork/reserve-data/reservesetting/common"
	"github.com/KyberNetwork/reserve-data/reservesetting/storage"
)

const (
	defaultTimeRange uint64 = 86400000
)

// Server struct for http package
type Server struct {
	app                reserve.Data
	core               reserve.Core
	host               string
	r                  *gin.Engine
	blockchain         Blockchain
	settingStorage     storage.Interface
	l                  *zap.SugaredLogger
	gasInfo            *gasinfo.GasPriceInfo
	binanceMainAccount *binance.Endpoint
	rcf                common.RawConfig
}

func getTimePoint(c *gin.Context, l *zap.SugaredLogger) uint64 {
	timestamp := c.DefaultQuery("timestamp", "")
	if timestamp == "" {
		timepoint := common.NowInMillis()
		l.Debugw("Interpreted timestamp to current time", "timepoint", timepoint)
		return timepoint
	}
	timepoint, err := strconv.ParseUint(timestamp, 10, 64)
	if err != nil {
		l.Debugw("Interpreted timestamp to default", "timestamp", timestamp)
		return common.NowInMillis()
	}
	l.Debugw("Interpreted timestamp", "timestamp", timestamp, "timepoint", timepoint)
	return timepoint
}

// AllPricesVersion return current version of all token
func (s *Server) AllPricesVersion(c *gin.Context) {
	s.l.Debugw("Getting all prices version")
	data, err := s.app.CurrentPriceVersion(getTimePoint(c, s.l))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithField("version", data))
	}
}

type price struct {
	Base      rtypes.AssetID      `json:"base"`
	Quote     rtypes.AssetID      `json:"quote"`
	Exchange  rtypes.ExchangeID   `json:"exchange"`
	Bids      []common.PriceEntry `json:"bids"`
	Asks      []common.PriceEntry `json:"asks"`
	Timestamp uint64              `json:"timestamp"`
}

// AllPrices return prices of all tokens
func (s *Server) AllPrices(c *gin.Context) {
	s.l.Debugw("Getting all prices")
	data, err := s.app.GetAllPrices(getTimePoint(c, s.l))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}

	var responseData []price
	for tp, onePrice := range data.Data {
		pair, err := s.settingStorage.GetTradingPair(tp, false)
		if err != nil {
			httputil.ResponseFailure(c, httputil.WithError(err))
			return
		}
		for exchangeID, exchangePrice := range onePrice {
			ts := uint64(0)
			if exchangePrice.Timestamp != "" {
				ts = exchangePrice.Timestamp.Millis()
			}
			responseData = append(responseData, price{
				Base:      pair.Base,
				Quote:     pair.Quote,
				Exchange:  exchangeID,
				Bids:      exchangePrice.Bids,
				Asks:      exchangePrice.Asks,
				Timestamp: ts,
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
	s.l.Debugw("Getting current auth data snapshot version")
	data, err := s.app.CurrentAuthDataVersion(getTimePoint(c, s.l))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithField("version", data))
	}
}

// AuthData return current auth data
func (s *Server) AuthData(c *gin.Context) {
	s.l.Debugw("Getting current auth data snapshot \n")
	data, err := s.app.GetAuthData(getTimePoint(c, s.l))
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

// TransferSelfRequest form
type TransferSelfRequest struct {
	Nonce    uint64 `json:"nonce" binding:"required"`
	Op       string `json:"op" binding:"required"`
	GasPrice uint64 `json:"gas_price" binding:"numeric,min=0,max=1000"`
}

// Trade create an order in cexs
func (s *Server) transferSelf(c *gin.Context) {
	var request TransferSelfRequest
	var err error
	if err := c.ShouldBindJSON(&request); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}

	tx, err := s.core.TransferToSelf(request.Op, request.Nonce, float64(request.GasPrice))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithMultipleFields(gin.H{
		"tx": tx.Hash().String(),
	}))
}

// GetRate return rate of a token
func (s *Server) GetRate(c *gin.Context) {
	s.l.Infow("Getting all rates")
	data, err := s.app.GetRate(getTimePoint(c, s.l))
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
	Pair   rtypes.TradingPairID `json:"pair"`
	Amount float64              `json:"amount"`
	Rate   float64              `json:"rate"`
	Type   string               `json:"type"`
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
	pair, err = s.settingStorage.GetTradingPair(request.Pair, false)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}

	if request.Type != "sell" && request.Type != "buy" {
		httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("Trade type of %s is not supported.", request.Type)))
		return
	}

	exchange, ok := common.SupportedExchanges[pair.ExchangeID]
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
	ExchangeID rtypes.ExchangeID    `form:"exchange_id"`
	Pair       rtypes.TradingPairID `form:"pair"`
}

// OpenOrders request for open orders
func (s *Server) OpenOrders(c *gin.Context) {
	var (
		logger = s.l.With("func", caller.GetCurrentFunctionName())
		query  OpenOrdersRequest
		pair   v3common.TradingPairSymbols
		err    error
	)
	if err := c.ShouldBindQuery(&query); err != nil {
		logger.Errorw("query is is not correct format", "error", err)
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	getExchange := make(map[rtypes.ExchangeID]common.Exchange)
	if query.ExchangeID != 0 {
		exchange, ok := common.SupportedExchanges[query.ExchangeID]
		if !ok {
			httputil.ResponseFailure(c, httputil.WithError(errors.Errorf("exchange %v is not supported", query.ExchangeID)))
			return
		}
		getExchange[query.ExchangeID] = exchange
	} else {
		getExchange = common.SupportedExchanges
	}
	if query.Pair != 0 {
		logger.Infow("query pair", "pair", query.Pair)
		pair, err = s.settingStorage.GetTradingPair(query.Pair, false)
		if err != nil {
			logger.Errorw("failed to get trading token pair from setting data base", "err", err)
			httputil.ResponseFailure(c, httputil.WithError(err))
			return
		}
		logger.Infow("getting open orders for pair", "base", pair.BaseSymbol, "quote", pair.QuoteSymbol)
	} else {
		logger.Debug("pair id not provide, getting open orders for all supported pairs")
	}
	result := make(map[rtypes.ExchangeID][]common.Order)
	for exchangeID, exchange := range getExchange {
		openOrders, err := exchange.OpenOrders(pair)
		if err != nil {
			logger.Errorw("failed to get open orders",
				"exchange", exchange.ID().String(),
				"base", pair.BaseSymbol,
				"quote", pair.QuoteSymbol,
				"error", err)
			httputil.ResponseFailure(c, httputil.WithError(err))
			return
		}
		result[exchangeID] = openOrders
	}

	httputil.ResponseSuccess(c, httputil.WithData(result))
}

// CancelOrderRequest type
type CancelOrderRequest struct {
	ExchangeID rtypes.ExchangeID     `json:"exchange_id"`
	Orders     []common.RequestOrder `json:"orders"`
}

// CancelOrder cancel an order from cexs
func (s *Server) CancelOrder(c *gin.Context) {
	var request CancelOrderRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}

	exchange, ok := common.SupportedExchanges[request.ExchangeID]
	if !ok {
		httputil.ResponseFailure(c, httputil.WithError(errors.Errorf("exchange %v is not supported", request.ExchangeID)))
		return
	}
	s.l.Infow("Cancel order", "order", request.Orders, "from", exchange.ID().String())
	result := s.core.CancelOrders(request.Orders, exchange)
	httputil.ResponseSuccess(c, httputil.WithData(result))
}

// WithdrawRequest type
type WithdrawRequest struct {
	ExchangeID rtypes.ExchangeID `json:"exchange"`
	Asset      rtypes.AssetID    `json:"asset"`
	Amount     *big.Int          `json:"amount"`
}

// Withdraw asset to reserve from cex
func (s *Server) Withdraw(c *gin.Context) {
	var request WithdrawRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}

	exchange, ok := common.SupportedExchanges[request.ExchangeID]
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

type transferRequest struct {
	ExchangeID  rtypes.ExchangeID `json:"exchange"`
	Asset       rtypes.AssetID    `json:"asset"`
	Amount      *big.Int          `json:"amount"`
	FromAccount string            `json:"from_account"`
	ToAccount   string            `json:"to_account"`
}

// Withdraw asset to reserve from cex
func (s *Server) cexTransfer(c *gin.Context) {
	var request transferRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}

	exh, ok := common.SupportedExchanges[request.ExchangeID]
	if !ok {
		httputil.ResponseFailure(c, httputil.WithError(errors.Errorf("exchange %v is not supported", request.ExchangeID)))
		return
	}

	asset, err := s.settingStorage.GetAsset(request.Asset)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	s.l.Infow("cexTransfer", "amount", request.Amount.Text(10), "asset_id", asset.ID,
		"asset_symbol", asset.Symbol, "exchange", exh.ID().String(), "from_account", request.FromAccount,
		"to_account", request.ToAccount)
	id, err := s.core.Transfer(request.FromAccount, request.ToAccount, asset, request.Amount, exh)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithField("id", id))
}

func (s *Server) getBinanceMainAccountInfo(c *gin.Context) {
	resp, err := s.binanceMainAccount.GetInfo()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	mb := make(map[string]exchange.Balance)
	for _, b := range resp.Balances {
		mb[b.Asset] = b
	}
	assets, err := s.settingStorage.GetAssets()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	var result []exchange.BinanceMainAccountBalance
	for _, a := range assets {
	loop:
		for _, aex := range a.Exchanges {
			if aex.ExchangeID == rtypes.Binance || aex.ExchangeID == rtypes.Binance2 {
				if balance, ok := mb[aex.Symbol]; ok {
					result = append(result, exchange.BinanceMainAccountBalance{
						AssetID: a.ID,
						Symbol:  aex.Symbol,
						Free:    balance.Free,
						Locked:  balance.Locked,
					})
					break loop
				}
			}
		}
	}
	httputil.ResponseSuccess(c, httputil.WithField("data", result))
}

// DepositRequest type
type DepositRequest struct {
	ExchangeID rtypes.ExchangeID `json:"exchange"`
	Amount     *big.Int          `json:"amount"`
	Asset      rtypes.AssetID    `json:"asset"`
}

// Deposit asset into cex
func (s *Server) Deposit(c *gin.Context) {
	var request DepositRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}

	exchange, ok := common.SupportedExchanges[request.ExchangeID]
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
	id, err := s.core.Deposit(exchange, asset, request.Amount, getTimePoint(c, s.l))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithField("id", id))
}

type getActivitiesRequest struct {
	FromTime uint64   `form:"fromTime" binding:"required"`
	ToTime   uint64   `form:"toTime" binding:"required"`
	Actions  []string `form:"actions"`
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

	data, err := s.app.GetRecords(query.FromTime, query.ToTime, query.Actions)
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

func (s *Server) getTriggers(c *gin.Context) {
	fromTime, toTime, ok := s.ValidateTimeInput(c)
	if !ok {
		return
	}
	triggers, err := s.app.GetAssetRateTriggers(fromTime, toTime)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(triggers))
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

func (s *Server) getNonWithdrawableAssets(c *gin.Context) {
	exchange := c.Param("exchange")
	assets, err := s.settingStorage.GetAssets()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	switch exchange {
	case rtypes.Binance.String(), rtypes.Binance2.String():
		data, err := s.binanceMainAccount.GetAllAssetWithdrawStatus()
		if err != nil {
			httputil.ResponseFailure(c, httputil.WithError(err))
			return
		}
		nwa := []rtypes.AssetID{}
		for _, a := range assets {
			for _, aex := range a.Exchanges {
				if aex.ExchangeID == rtypes.Binance || aex.ExchangeID == rtypes.Binance2 {
					if !data[aex.Symbol] {
						nwa = append(nwa, aex.AssetID)
					}
					break
				}
			}
		}
		httputil.ResponseSuccess(c, httputil.WithData(nwa))
	case rtypes.Huobi.String():
		httputil.ResponseSuccess(c, httputil.WithData([]rtypes.AssetID{}))
	default:
		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": fmt.Sprintf("exchange %s is not supported", exchange),
			},
		)
	}
}

func (s *Server) getVersion(context *gin.Context) {
	context.String(http.StatusOK, "%s", common.AppVersion)
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
		g.POST("/cancel-orders", s.CancelOrder)
		g.POST("/cancel-all-orders", s.CancelAllOrders)
		g.POST("/deposit", s.Deposit)
		g.POST("/withdraw", s.Withdraw)
		g.POST("/trade", s.Trade)
		g.POST("/transfer-self", s.transferSelf)
		g.POST("/setrates", s.SetRate)
		g.POST("/cancel-setrates", s.cancelSetRate)
		g.GET("/tradehistory", s.GetTradeHistory)

		g.GET("/timeserver", s.GetTimeServer)

		g.GET("/gold-feed", s.GetGoldData)
		g.GET("/btc-feed", s.GetBTCData)
		g.GET("/usd-feed", s.GetUSDData)

		g.GET("/addresses", s.GetAddresses)

		g.PUT("/update-token-indice", s.updateTokenIndice)
		g.GET("/check-token-indice", s.checkTokenIndice)
		g.GET("/token-rate-trigger", s.getTriggers)
		g.POST("/cex-transfer", s.cexTransfer)
		g.GET("/binance/main", s.getBinanceMainAccountInfo)

		g.GET("/non-withdrawable-assets/:exchange", s.getNonWithdrawableAssets)
		g.GET("/version", s.getVersion)
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
	gasInfo *gasinfo.GasPriceInfo,
	binanceMainAccount *binance.Endpoint,
	rcf common.RawConfig,
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
		app:                app,
		core:               core,
		host:               host,
		r:                  r,
		blockchain:         bc,
		settingStorage:     settingStorage,
		l:                  zap.S(),
		gasInfo:            gasInfo,
		binanceMainAccount: binanceMainAccount,
		rcf:                rcf,
	}
}
