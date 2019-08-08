package http

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/KyberNetwork/reserve-data/http/httputil"
	"github.com/KyberNetwork/reserve-data/v3/common"
)

func (s *Server) createCreateTradingBy(c *gin.Context) {
	var createTradingBy common.CreateCreateTradingBy
	err := c.ShouldBindJSON(&createTradingBy)

	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}

	for index, entry := range createTradingBy.TradingBys {
		if err = s.checkCreateTradingByEntry(entry); err != nil {
			httputil.ResponseFailure(c, httputil.WithError(err), httputil.WithField("index", index),
				httputil.WithField("asset_id", entry.AssetID), httputil.WithField("trading_pair", entry.TradingPairID))
			return
		}
	}

	id, err := s.storage.CreateCreateTradingBy(createTradingBy)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithField("id", id))
}

func (s *Server) checkCreateTradingByEntry(createEntry common.CreateTradingByEntry) error {
	tpSymBol, err := s.storage.GetTradingPair(createEntry.TradingPairID)
	if err != nil {
		return err
	}
	if tpSymBol.Base != createEntry.AssetID && tpSymBol.Quote != createEntry.AssetID {
		return common.ErrTradingByAssetIDInvalid
	}
	return nil
}

func (s *Server) getCreateTradingBys(c *gin.Context) {
	result, err := s.storage.GetCreateTradingBys()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(result))
}

func (s *Server) getCreateTradingBy(c *gin.Context) {
	var input struct {
		ID uint64 `uri:"id" binding:"exists"`
	}
	if err := c.ShouldBindUri(&input); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	result, err := s.storage.GetCreateTradingBy(input.ID)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(result))
}

func (s *Server) confirmCreateTradingBy(c *gin.Context) {
	var input struct {
		ID uint64 `uri:"id" binding:"exists"`
	}
	if err := c.ShouldBindUri(&input); err != nil {
		log.Println(err)
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	err := s.storage.ConfirmCreateTradingBy(input.ID)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (s *Server) rejectCreateTradingBy(c *gin.Context) {
	var input struct {
		ID uint64 `uri:"id" binding:"exists"`
	}
	if err := c.ShouldBindUri(&input); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	err := s.storage.RejectCreateTradingBy(input.ID)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}
