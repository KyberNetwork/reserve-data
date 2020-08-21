package http

import (
	"github.com/gin-gonic/gin"

	common2 "github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/http/httputil"
	"github.com/KyberNetwork/reserve-data/lib/rtypes"
	"github.com/KyberNetwork/reserve-data/reservesetting/common"
)

func (s *Server) setPriceFactor(c *gin.Context) {
	s.l.Info("storing price factor")
	var params common.PriceFactorAtTime
	if err := c.ShouldBindJSON(&params); err != nil {
		s.l.Warnw("cannot bind request parameter", "err", err)
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	id, err := s.storage.CreatePriceFactor(params)
	if err != nil {
		s.l.Warnw("cannot store price factor", "err", err)
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithField("id", id))
}

type getPriceFactorParams struct {
	From uint64 `form:"from" binding:"required"`
	To   uint64 `form:"to" binding:"required"`
}

func convertToPriceFactorResponse(in []common.PriceFactorAtTime) []*common.AssetPriceFactorListResponse {
	var assetToPriceList = map[rtypes.AssetID]*common.AssetPriceFactorListResponse{}
	var res = make([]*common.AssetPriceFactorListResponse, 0)
	for _, assetList := range in {
		for _, asset := range assetList.Data {
			var e *common.AssetPriceFactorListResponse
			var ok bool
			if e, ok = assetToPriceList[asset.AssetID]; !ok {
				e = &common.AssetPriceFactorListResponse{
					AssetID: asset.AssetID,
					Data:    nil,
				}
				assetToPriceList[asset.AssetID] = e
				res = append(res, e)
			}
			e.Data = append(e.Data, common.AssetPriceFactorResponse{
				Timestamp: assetList.Timestamp,
				AfpMid:    asset.AfpMid,
				Spread:    asset.Spread,
			})
		}
	}
	return res
}
func (s *Server) getPriceFactor(c *gin.Context) {
	s.l.Infow("get price factor")
	var params getPriceFactorParams
	if err := c.ShouldBindQuery(&params); err != nil {
		s.l.Warnw("cannot bind request parameter", "err", err)
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	store, err := s.storage.GetPriceFactors(params.From, params.To)
	if err != nil {
		s.l.Warnw("cannot get price factor", err, err)
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	data := convertToPriceFactorResponse(store)
	httputil.ResponseSuccess(c, httputil.WithMultipleFields(gin.H{
		"timestamp":  common2.NowInMillis(),
		"returnTime": common2.NowInMillis(),
		"data":       data,
	}))
}

func (s *Server) getSetRateStatus(c *gin.Context) {
	status, err := s.storage.GetSetRateStatus()
	if err != nil {
		s.l.Warnw("failed to get set rate status", "err", err)
		httputil.ResponseFailure(c, httputil.WithError(err))
	}
	httputil.ResponseSuccess(c, httputil.WithData(status))
}

func (s *Server) holdSetRate(c *gin.Context) {
	if err := s.storage.SetSetRateStatus(false); err != nil {
		s.l.Warnw("failed to set rate staus", "err", err)
		httputil.ResponseFailure(c, httputil.WithError(err))
	}
	httputil.ResponseSuccess(c)
}

func (s *Server) enableSetRate(c *gin.Context) {
	if err := s.storage.SetSetRateStatus(true); err != nil {
		s.l.Warnw("failed to set SetRate status", "err", err)
		httputil.ResponseFailure(c, httputil.WithError(err))
	}
	httputil.ResponseSuccess(c)
}

func (s *Server) getRebalanceStatus(c *gin.Context) {
	status, err := s.storage.GetRebalanceStatus()
	if err != nil {
		s.l.Warnw("failed to get rebalance status", "err", err)
		httputil.ResponseFailure(c, httputil.WithError(err))
	}
	httputil.ResponseSuccess(c, httputil.WithData(status))
}

func (s *Server) holdRebalance(c *gin.Context) {
	if err := s.storage.SetRebalanceStatus(false); err != nil {
		s.l.Warnw("failed to set rebalance status", "err", err)
		httputil.ResponseFailure(c, httputil.WithError(err))
	}
	httputil.ResponseSuccess(c)
}

func (s *Server) enableRebalance(c *gin.Context) {
	if err := s.storage.SetRebalanceStatus(true); err != nil {
		s.l.Warnw("failed to set rebalance status", "err", err)
		httputil.ResponseFailure(c, httputil.WithError(err))
	}
	httputil.ResponseSuccess(c)
}
