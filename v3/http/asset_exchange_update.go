package http

import (
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

	id, err := s.storage.CreatePendingObject(updateAssetExchange, common.PendingTypeUpdateAssetExchange)
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
