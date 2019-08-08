package http

import (
	"log"

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
	id, err := s.storage.CreateUpdateTradingPair(updateTradingPair)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithField("id", id))
}

func (s *Server) getUpdateTradingPairs(c *gin.Context) {
	result, err := s.storage.GetUpdateTradingPairs()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(result))
}

func (s *Server) getUpdateTradingPair(c *gin.Context) {
	var input struct {
		ID uint64 `uri:"id" binding:"exists"`
	}
	if err := c.ShouldBindUri(&input); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	result, err := s.storage.GetUpdateTradingPair(input.ID)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(result))
}

func (s *Server) confirmUpdateTradingPair(c *gin.Context) {
	var input struct {
		ID uint64 `uri:"id" binding:"exists"`
	}
	if err := c.ShouldBindUri(&input); err != nil {
		log.Println(err)
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	err := s.storage.ConfirmUpdateTradingPair(input.ID)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (s *Server) rejectUpdateTradingPair(c *gin.Context) {
	var input struct {
		ID uint64 `uri:"id" binding:"exists"`
	}
	if err := c.ShouldBindUri(&input); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	err := s.storage.RejectUpdateTradingPair(input.ID)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}
