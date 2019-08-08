package http

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/KyberNetwork/reserve-data/http/httputil"
	"github.com/KyberNetwork/reserve-data/v3/common"
)

func (s *Server) createUpdateAssetExchange(c *gin.Context) {
	var updateAssetExchange common.CreateUpdateAssetExchange

	err := c.ShouldBindJSON(&updateAssetExchange)

	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}

	for index, entry := range updateAssetExchange.AssetExchanges {
		if err := s.checkUpdateAssetExchangeParams(entry); err != nil {
			httputil.ResponseFailure(c, httputil.WithError(err), httputil.WithField("index", index),
				httputil.WithField("asset_exchange_id", entry.ID))
			return
		}
	}

	id, err := s.storage.CreateUpdateAssetExchange(updateAssetExchange)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithField("id", id))
}

func (s *Server) checkUpdateAssetExchangeParams(updateEntry common.UpdateAssetExchangeEntry) error {
	assetExchange, err := s.storage.GetAssetExchange(updateEntry.ID)
	if err != nil {
		return errors.Wrap(err, "asset exchange not found")
	}

	asset, err := s.storage.GetAsset(assetExchange.AssetID)
	if err != nil {
		return errors.Wrap(err, "asset not found")
	}

	if asset.Transferable && updateEntry.DepositAddress != nil && common.IsZeroAddress(*updateEntry.DepositAddress) {
		return common.ErrDepositAddressMissing
	}
	return nil
}

func (s *Server) getUpdateAssetExchanges(c *gin.Context) {
	result, err := s.storage.GetUpdateAssetExchanges()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(result))
}

func (s *Server) getUpdateAssetExchange(c *gin.Context) {
	var input struct {
		ID uint64 `uri:"id" binding:"exists"`
	}
	if err := c.ShouldBindUri(&input); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	result, err := s.storage.GetUpdateAssetExchange(input.ID)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(result))
}

func (s *Server) confirmUpdateAssetExchange(c *gin.Context) {
	var input struct {
		ID uint64 `uri:"id" binding:"exists"`
	}
	if err := c.ShouldBindUri(&input); err != nil {
		log.Println(err)
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	err := s.storage.ConfirmUpdateAssetExchange(input.ID)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (s *Server) rejectUpdateAssetExchange(c *gin.Context) {
	var input struct {
		ID uint64 `uri:"id" binding:"exists"`
	}
	if err := c.ShouldBindUri(&input); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	err := s.storage.RejectUpdateAssetExchange(input.ID)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}
