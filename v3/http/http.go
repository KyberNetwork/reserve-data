package http

import (
	"github.com/KyberNetwork/reserve-data/v3/common"
	"github.com/gin-gonic/gin"

	"github.com/KyberNetwork/reserve-data/v3/storage"
)

// Server is the HTTP server of token V3.
type Server struct {
	storage storage.Interface
	r       *gin.Engine
}

// NewServer creates new HTTP server for v3 APIs.
func NewServer(storage storage.Interface, r *gin.Engine) *Server {
	if r == nil {
		r = gin.Default()
	}
	server := &Server{
		storage: storage,
		r:       r,
	}
	g := r.Group("/v3")

	g.GET("/asset/:id", server.getAsset)
	g.GET("/asset", server.getAssets)
	g.GET("/exchange/:id", server.getExchange)
	g.GET("/exchange", server.getExchanges)
	g.GET("/trading-pair/:id", server.getTradingPair)

	// because we don't allow to create asset directly, it must go through pending operation
	// so all 'create' operation mean to operate on pending object.

	// API for CreateAsset object
	g.POST("/create-asset", server.createCreateAsset)
	g.GET("/create-asset", server.getPendingObjects(common.PendingTypeCreateAsset))
	g.GET("/create-asset/:id", server.getPendingObject(common.PendingTypeCreateAsset))
	g.PUT("/create-asset/:id", server.confirmPendingObject(common.PendingTypeCreateAsset))
	g.DELETE("/create-asset/:id", server.rejectPendingObject(common.PendingTypeCreateAsset))

	// API for UpdateAsset object
	g.POST("/update-asset", server.createUpdateAsset)
	g.GET("/update-asset", server.getPendingObjects(common.PendingTypeUpdateAsset))
	g.GET("/update-asset/:id", server.getPendingObject(common.PendingTypeUpdateAsset))
	g.PUT("/update-asset/:id", server.confirmPendingObject(common.PendingTypeUpdateAsset))
	g.DELETE("/update-asset/:id", server.rejectPendingObject(common.PendingTypeUpdateAsset))

	// API for ChangeAssetAddress object
	g.POST("/change-asset-address", server.createChangeAssetAddress)
	g.GET("/change-asset-address", server.getPendingObjects(common.PendingTypeChangeAssetAddr))
	g.GET("/change-asset-address/:id", server.getPendingObject(common.PendingTypeChangeAssetAddr))
	g.PUT("/change-asset-address/:id", server.confirmPendingObject(common.PendingTypeChangeAssetAddr))
	g.DELETE("/change-asset-address/:id", server.rejectPendingObject(common.PendingTypeChangeAssetAddr))

	// API for CreateAssetExchange object
	g.POST("/create-asset-exchange", server.createCreateAssetExchange)
	g.GET("/create-asset-exchange", server.getPendingObjects(common.PendingTypeCreateAssetExchange))
	g.GET("/create-asset-exchange/:id", server.getPendingObject(common.PendingTypeCreateAssetExchange))
	g.PUT("/create-asset-exchange/:id", server.confirmPendingObject(common.PendingTypeCreateAssetExchange))
	g.DELETE("/create-asset-exchange/:id", server.rejectPendingObject(common.PendingTypeCreateAssetExchange))

	g.GET("/update-asset-exchange", server.getPendingObjects(common.PendingTypeUpdateAssetExchange))
	g.GET("/update-asset-exchange/:id", server.getPendingObject(common.PendingTypeUpdateAssetExchange))
	g.POST("/update-asset-exchange", server.createUpdateAssetExchange)
	g.PUT("/update-asset-exchange/:id", server.confirmPendingObject(common.PendingTypeUpdateAssetExchange))
	g.DELETE("/update-asset-exchange/:id", server.rejectPendingObject(common.PendingTypeUpdateAssetExchange))

	// API for UpdateExchange object
	g.PUT("/update-exchange/:id", server.confirmPendingObject(common.PendingTypeUpdateExchange))
	g.GET("/update-exchange", server.getPendingObjects(common.PendingTypeUpdateExchange))
	g.GET("/update-exchange/:id", server.getPendingObject(common.PendingTypeUpdateExchange))
	g.POST("/update-exchange", server.createUpdateExchange)
	g.DELETE("/update-exchange/:id", server.rejectPendingObject(common.PendingTypeUpdateExchange))

	g.PUT("/create-trading-pair/:id", server.confirmPendingObject(common.PendingTypeCreateTradingPair))
	g.GET("/create-trading-pair", server.getPendingObjects(common.PendingTypeCreateTradingPair))
	g.GET("/create-trading-pair/:id", server.getPendingObject(common.PendingTypeCreateTradingPair))
	g.POST("/create-trading-pair", server.createCreateTradingPair)
	g.DELETE("/create-trading-pair/:id", server.rejectPendingObject(common.PendingTypeCreateTradingPair))

	g.PUT("/update-trading-pair/:id", server.confirmPendingObject(common.PendingTypeUpdateTradingPair))
	g.GET("/update-trading-pair", server.getPendingObjects(common.PendingTypeUpdateTradingPair))
	g.GET("/update-trading-pair/:id", server.getPendingObject(common.PendingTypeUpdateTradingPair))
	g.POST("/update-trading-pair", server.createUpdateTradingPair)
	g.DELETE("/update-trading-pair/:id", server.rejectPendingObject(common.PendingTypeUpdateTradingPair))

	g.POST("/create-trading-by", server.createCreateTradingBy)
	g.GET("/create-trading-by", server.getPendingObjects(common.PendingTypeCreateTradingBy))
	g.GET("/create-trading-by/:id", server.getPendingObject(common.PendingTypeCreateTradingBy))
	g.PUT("/create-trading-by/:id", server.confirmPendingObject(common.PendingTypeCreateTradingBy))
	g.DELETE("/create-trading-by/:id", server.rejectPendingObject(common.PendingTypeCreateTradingBy))

	return server
}
