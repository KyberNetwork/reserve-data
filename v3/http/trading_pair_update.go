package http

import (
	"github.com/gin-gonic/gin"

	"github.com/KyberNetwork/reserve-data/http/httputil"
	"github.com/KyberNetwork/reserve-data/v3/common"
)

func (s *Server) createUpdateTradingPair(c *gin.Context) {
	var updateTradingPair common.CreateUpdateTradingPair

	err := c.ShouldBindJSON(&updateTradingPair)

	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	// TODO validate if the update request satisfy constraint
	id, err := s.storage.CreatePendingObject(updateTradingPair, common.PendingTypeUpdateTradingPair)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithField("id", id))
}
