package http

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/KyberNetwork/reserve-data/http/httputil"
	"github.com/KyberNetwork/reserve-data/v3/common"
)

func (s *Server) createCreateAssetExchange(c *gin.Context) {
	var createAssetExchange common.CreateCreateAssetExchange

	err := c.ShouldBindJSON(&createAssetExchange)

	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}

	for index, entry := range createAssetExchange.AssetExchanges {
		if err := s.checkCreateAssetExchangeParams(entry); err != nil {
			httputil.ResponseFailure(c, httputil.WithError(err), httputil.WithField("index", index),
				httputil.WithField("asset_id", entry.AssetID), httputil.WithField("exchange_id", entry.ExchangeID))
			return
		}
	}

	id, err := s.storage.CreateCreateAssetExchange(createAssetExchange)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithField("id", id))
}

func (s *Server) checkCreateAssetExchangeParams(createEntry common.CreateAssetExchangeEntry) error {
	asset, err := s.storage.GetAsset(createEntry.AssetID)
	if err != nil {
		return errors.Wrap(err, "asset not found")
	}

	_, err = s.storage.GetExchange(createEntry.ExchangeID)
	if err != nil {
		return errors.Wrap(err, "exchange not found")
	}

	for _, exchange := range asset.Exchanges {
		if exchange.ExchangeID == createEntry.ExchangeID {
			return common.ErrAssetExchangeAlreadyExist
		}
	}
	if asset.Transferable && common.IsZeroAddress(createEntry.DepositAddress) {
		return common.ErrDepositAddressMissing
	}
	return nil
}

func (s *Server) getCreateAssetExchanges(c *gin.Context) {
	result, err := s.storage.GetCreateAssetExchanges()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(result))
}

func (s *Server) getCreateAssetExchange(c *gin.Context) {
	var input struct {
		ID uint64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&input); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	result, err := s.storage.GetCreateAssetExchange(input.ID)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(result))
}

func (s *Server) confirmCreateAssetExchange(c *gin.Context) {
	var input struct {
		ID uint64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&input); err != nil {
		log.Println(err)
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	err := s.storage.ConfirmCreateAssetExchange(input.ID)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (s *Server) rejectCreateAssetExchange(c *gin.Context) {
	var input struct {
		ID uint64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&input); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	err := s.storage.RejectCreateAssetExchange(input.ID)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}
