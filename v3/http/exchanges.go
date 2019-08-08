package http

import (
	"github.com/gin-gonic/gin"

	"github.com/KyberNetwork/reserve-data/http/httputil"
	"github.com/KyberNetwork/reserve-data/v3/common"
)

func (s *Server) getExchanges(c *gin.Context) {
	exhs, err := s.storage.GetExchanges()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(exhs))
}

func (s *Server) getExchange(c *gin.Context) {

	var input struct {
		ID uint64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&input); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	asset, err := s.storage.GetExchange(input.ID)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(asset))
}

func (s *Server) createUpdateExchange(c *gin.Context) {
	var updateExchange common.CreateUpdateExchange

	err := c.ShouldBindJSON(&updateExchange)

	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	// TODO validate if the update request satisfy constraint
	id, err := s.storage.CreatePendingObject(updateExchange, common.PendingTypeUpdateExchange)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithField("id", id))
}
