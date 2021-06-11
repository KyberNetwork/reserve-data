package http

import (
	"log"
	"net/http"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	v1common "github.com/KyberNetwork/reserve-data/common"
	gaspricedataclient "github.com/KyberNetwork/reserve-data/common/gaspricedata-client"
	coreclient "github.com/KyberNetwork/reserve-data/lib/core-client"
	marketdatacli "github.com/KyberNetwork/reserve-data/lib/market-data"
	"github.com/KyberNetwork/reserve-data/lib/rtypes"
	"github.com/KyberNetwork/reserve-data/reservesetting/common"
	"github.com/KyberNetwork/reserve-data/reservesetting/storage"
)

// Server is the HTTP server of token V3.
type Server struct {
	storage                storage.Interface
	r                      *gin.Engine
	host                   string
	supportedExchanges     map[rtypes.ExchangeID]v1common.LiveExchange
	l                      *zap.SugaredLogger
	coreClient             *coreclient.Client
	gasClient              gaspricedataclient.Client
	marketDataClient       *marketdatacli.Client
	numberApprovalRequired int
}

// NewServer creates new HTTP server for reservesetting APIs.
func NewServer(storage storage.Interface, host string, supportedExchanges map[rtypes.ExchangeID]v1common.LiveExchange,
	sentryDSN string, coreClient *coreclient.Client, gasClient gaspricedataclient.Client,
	marketDataClient *marketdatacli.Client, numberApprovalRequired int) *Server {
	l := zap.S()
	r := gin.Default()
	if sentryDSN != "" {
		// To initialize Sentry's handler, you need to initialize Sentry itself beforehand
		if err := sentry.Init(sentry.ClientOptions{
			Dsn: sentryDSN,
		}); err != nil {
			l.Warnw("Sentry initialization failed", "err", err)
		}
		r.Use(sentrygin.New(sentrygin.Options{}))
	}
	server := &Server{
		storage:                storage,
		r:                      r,
		host:                   host,
		supportedExchanges:     supportedExchanges,
		l:                      l,
		coreClient:             coreClient,
		gasClient:              gasClient,
		marketDataClient:       marketDataClient,
		numberApprovalRequired: numberApprovalRequired,
	}
	g := r.Group("/v3")

	g.GET("/asset/:id", server.getAsset)
	g.GET("/asset", server.getAssets)
	g.GET("/exchange/:id", server.getExchange)
	g.GET("/exchange", server.getExchanges)
	g.GET("/trading-pair/:id", server.getTradingPair)
	g.GET("/trading-pair", server.getTradingPairs)
	g.GET("/stable-token-params", server.getStableTokenParams)
	g.GET("/feed-configurations", server.getFeedConfigurations)

	// because we don't allow to create asset directly, it must go through pending operation
	// so all 'create' operation mean to operate on pending object.

	g.POST("/setting-change-main", server.createSettingChangeWithType(common.ChangeCatalogMain))
	g.GET("/setting-change-main", server.getSettingChangeWithType(common.ChangeCatalogMain))
	g.GET("/setting-change-main/:id", server.getSettingChange)
	g.PUT("/setting-change-main/:id", server.confirmSettingChange)
	g.DELETE("/setting-change-main/:id", server.rejectSettingChange)

	g.POST("/setting-change-target", server.createSettingChangeWithType(common.ChangeCatalogSetTarget))
	g.GET("/setting-change-target", server.getSettingChangeWithType(common.ChangeCatalogSetTarget))
	g.GET("/setting-change-target/:id", server.getSettingChange)
	g.PUT("/setting-change-target/:id", server.confirmSettingChange)
	g.DELETE("/setting-change-target/:id", server.rejectSettingChange)

	g.POST("/setting-change-pwis", server.createSettingChangeWithType(common.ChangeCatalogSetPWIS))
	g.GET("/setting-change-pwis", server.getSettingChangeWithType(common.ChangeCatalogSetPWIS))
	g.GET("/setting-change-pwis/:id", server.getSettingChange)
	g.PUT("/setting-change-pwis/:id", server.confirmSettingChange)
	g.DELETE("/setting-change-pwis/:id", server.rejectSettingChange)

	g.POST("/setting-change-stable", server.createSettingChangeWithType(common.ChangeCatalogStableToken))
	g.GET("/setting-change-stable", server.getSettingChangeWithType(common.ChangeCatalogStableToken))
	g.GET("/setting-change-stable/:id", server.getSettingChange)
	g.PUT("/setting-change-stable/:id", server.confirmSettingChange)
	g.DELETE("/setting-change-stable/:id", server.rejectSettingChange)

	g.POST("/setting-change-rbquadratic", server.createSettingChangeWithType(common.ChangeCatalogRebalanceQuadratic))
	g.GET("/setting-change-rbquadratic", server.getSettingChangeWithType(common.ChangeCatalogRebalanceQuadratic))
	g.GET("/setting-change-rbquadratic/:id", server.getSettingChange)
	g.PUT("/setting-change-rbquadratic/:id", server.confirmSettingChange)
	g.DELETE("/setting-change-rbquadratic/:id", server.rejectSettingChange)

	g.POST("/setting-change-update-exchange", server.createSettingChangeWithType(common.ChangeCatalogUpdateExchange))
	g.GET("/setting-change-update-exchange", server.getSettingChangeWithType(common.ChangeCatalogUpdateExchange))
	g.GET("/setting-change-update-exchange/:id", server.getSettingChange)
	g.PUT("/setting-change-update-exchange/:id", server.confirmSettingChange)
	g.DELETE("/setting-change-update-exchange/:id", server.rejectSettingChange)
	g.PUT("/set-exchange-enabled/:id", server.setExchangeEnabled)

	g.POST("/setting-change-feed-configuration", server.createSettingChangeWithType(common.ChangeCatalogFeedConfiguration))
	g.GET("/setting-change-feed-configuration", server.getSettingChangeWithType(common.ChangeCatalogFeedConfiguration))
	g.GET("/setting-change-feed-configuration/:id", server.getSettingChange)
	g.PUT("/setting-change-feed-configuration/:id", server.confirmSettingChange)
	g.DELETE("/setting-change-feed-configuration/:id", server.rejectSettingChange)

	g.POST("/setting-change-tpair", server.createSettingChangeWithType(common.ChangeCatalogTradingPair))
	g.GET("/setting-change-tpair", server.getSettingChangeWithType(common.ChangeCatalogTradingPair))
	g.GET("/setting-change-tpair/:id", server.getSettingChange)
	g.PUT("/setting-change-tpair/:id", server.confirmSettingChange)
	g.DELETE("/setting-change-tpair/:id", server.rejectSettingChange)

	g.PUT("/update-feed-status/:name", server.updateFeedStatus)

	g.GET("/price-factor", server.getPriceFactor)
	g.POST("/price-factor", server.setPriceFactor)

	g.GET("/set-rate-status", server.getSetRateStatus)
	g.POST("/hold-set-rate", server.holdSetRate)
	g.POST("/enable-set-rate", server.enableSetRate)

	g.GET("/rebalance-status", server.getRebalanceStatus)
	g.POST("/hold-rebalance", server.holdRebalance)
	g.POST("/enable-rebalance", server.enableRebalance)

	g.GET("/rate-trigger-period", server.getRateTriggerPeriod)
	g.POST("/rate-trigger-period", server.setRateTriggerPeriod)
	g.GET("/gas-threshold", server.getGasStatus)
	g.POST("/gas-threshold", server.setGasThreshold)
	g.GET("/gas-source", server.getPreferGasSource)
	g.POST("/gas-source", server.setPreferGasSource)
	g.GET("/version", server.getVersion)

	g.POST("/scheduled-job", server.addScheduledJob)
	g.GET("/scheduled-job", server.getAllScheduledJob)
	g.GET("/scheduled-job/:id", server.getScheduledJob)
	g.DELETE("/scheduled-job/:id", server.removeScheduledJob)

	g.DELETE("/disapprove-setting-change/:id", server.disapproveSettingChange)
	g.GET("/number-approval-required", server.getNumberApprovalRequired)

	return server
}
func (s *Server) getVersion(context *gin.Context) {
	context.String(http.StatusOK, "%s", v1common.AppVersion)
}

// EnableProfiler enable profiler on path "/debug/pprof"
func (s *Server) EnableProfiler() {
	pprof.Register(s.r)
}

// Run the server
func (s *Server) Run() {
	if err := s.r.Run(s.host); err != nil {
		log.Panic(err)
	}
}
