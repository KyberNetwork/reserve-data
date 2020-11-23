package http

import (
	"github.com/gin-gonic/gin"

	"github.com/KyberNetwork/reserve-data/http/httputil"
	"github.com/KyberNetwork/reserve-data/lib/rtypes"
	"github.com/KyberNetwork/reserve-data/reservesetting/common"
)

func (s *Server) getTradingPair(c *gin.Context) {
	var input struct {
		ID rtypes.TradingPairID `uri:"id" binding:"required"`
	}
	var filter struct {
		IncludingDeleted bool `form:"including_deleted" json:"including_deleted"`
	}
	if err := c.ShouldBindUri(&input); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	if err := c.ShouldBindQuery(&filter); err != nil {
		s.l.Errorw("failed to bind query", "err", err)
	}
	result, err := s.storage.GetTradingPair(input.ID, filter.IncludingDeleted)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(result))
}

type getTradingPairParam struct {
	ID rtypes.ExchangeID `form:"id" binding:"required"`
}

func (s *Server) getTradingPairs(c *gin.Context) {
	var (
		query getTradingPairParam
	)
	if err := c.ShouldBindQuery(&query); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}

	result, err := s.storage.GetTradingPairs(query.ID)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(result))
}

func (s *Server) checkDeleteTradingPairParams(entry common.DeleteTradingPairEntry) error {
	_, err := s.storage.GetTradingPair(entry.TradingPairID, false)
	return err
}
