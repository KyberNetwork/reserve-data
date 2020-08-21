package http

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/http/httputil"
	"github.com/KyberNetwork/reserve-data/lib/rtypes"
)

type cancelAllOrderRequest struct {
	ExchangeID rtypes.ExchangeID `json:"exchange_id"`
	Symbol     string            `json:"symbol" binding:"required"`
}

// CancelAllOrders cancel all orders
func (s *Server) CancelAllOrders(c *gin.Context) {
	var (
		query cancelAllOrderRequest
	)
	if err := c.ShouldBindJSON(&query); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	}
	exchange, ok := common.SupportedExchanges[query.ExchangeID]
	if !ok {
		httputil.ResponseFailure(c, httputil.WithError(errors.Errorf("exchange %v is not supported", query.ExchangeID)))
		return
	}
	if err := exchange.CancelAllOrders(query.Symbol); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}
