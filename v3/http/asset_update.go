package http

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/KyberNetwork/reserve-data/http/httputil"
	"github.com/KyberNetwork/reserve-data/v3/common"
)

func (s *Server) createUpdateAsset(c *gin.Context) {
	var updateAsset common.CreateUpdateAsset

	err := c.ShouldBindJSON(&updateAsset)

	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}

	for _, updateEntry := range updateAsset.Assets {
		if err = s.checkUpdateAssetParams(updateEntry); err != nil {
			httputil.ResponseFailure(c, httputil.WithError(err))
			return
		}
	}

	id, err := s.storage.CreateUpdateAsset(updateAsset)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithField("id", id))
}

func (s *Server) checkUpdateAssetParams(updateEntry common.UpdateAssetEntry) error {
	asset, err := s.storage.GetAsset(updateEntry.AssetID)
	if err != nil {
		return errors.Wrapf(err, "failed to get asset id: %v from db ", updateEntry.AssetID)
	}

	if updateEntry.Rebalance != nil && *updateEntry.Rebalance {
		if asset.RebalanceQuadratic == nil && updateEntry.RebalanceQuadratic == nil {
			return errors.Errorf("%v at asset id: %v", common.ErrRebalanceQuadraticMissing.Error(), updateEntry.AssetID)
		}

		if asset.Target == nil && updateEntry.Target == nil {
			return errors.Errorf("%v at asset id: %v", common.ErrAssetTargetMissing.Error(), updateEntry.AssetID)
		}
	}

	if updateEntry.SetRate != nil && *updateEntry.SetRate != common.SetRateNotSet && asset.PWI == nil && updateEntry.PWI == nil {
		return errors.Errorf("%v at asset id: %v", common.ErrPWIMissing.Error(), updateEntry.AssetID)
	}

	if updateEntry.Transferable != nil && *updateEntry.Transferable {
		for _, exchange := range asset.Exchanges {
			if common.IsZeroAddress(exchange.DepositAddress) {
				return errors.Errorf("%v at asset id: %v and asset_exchange: %v", common.ErrDepositAddressMissing, updateEntry.AssetID, exchange.ID)
			}
		}
	}
	return nil
}

func (s *Server) getUpdateAssets(c *gin.Context) {
	result, err := s.storage.GetUpdateAssets()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(result))
}

func (s *Server) getUpdateAsset(c *gin.Context) {
	var input struct {
		ID uint64 `uri:"id" binding:"exists"`
	}
	if err := c.ShouldBindUri(&input); err != nil {
		log.Println(err)
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	result, err := s.storage.GetUpdateAsset(input.ID)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(result))
}

func (s *Server) confirmUpdateAsset(c *gin.Context) {
	var input struct {
		ID uint64 `uri:"id" binding:"exists"`
	}
	if err := c.ShouldBindUri(&input); err != nil {
		log.Println(err)
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	err := s.storage.ConfirmUpdateAsset(input.ID)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (s *Server) rejectUpdateAsset(c *gin.Context) {
	var input struct {
		ID uint64 `uri:"id" binding:"exists"`
	}
	if err := c.ShouldBindUri(&input); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	err := s.storage.RejectUpdateAsset(input.ID)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}
