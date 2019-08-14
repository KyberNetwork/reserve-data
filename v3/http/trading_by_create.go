package http

import (
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
		if err = s.checkCreateTradingByParams(entry); err != nil {
			httputil.ResponseFailure(c, httputil.WithError(err), httputil.WithField("index", index),
				httputil.WithField("asset_id", entry.AssetID), httputil.WithField("trading_pair", entry.TradingPairID))
			return
		}
	}

	id, err := s.storage.CreatePendingObject(createTradingBy, common.PendingTypeCreateTradingBy)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithField("id", id))
}

func (s *Server) checkCreateTradingByParams(createEntry common.CreateTradingByEntry) error {
	tpSymBol, err := s.storage.GetTradingPair(createEntry.TradingPairID)
	if err != nil {
		return err
	}
	if tpSymBol.Base != createEntry.AssetID && tpSymBol.Quote != createEntry.AssetID {
		return common.ErrTradingByAssetIDInvalid
	}
	return nil
}
