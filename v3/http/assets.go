package http

import (
	"github.com/gin-gonic/gin"

	"github.com/KyberNetwork/reserve-data/http/httputil"
)

func (s *Server) getAsset(c *gin.Context) {

	var input struct {
		ID uint64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&input); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	asset, err := s.storage.GetAsset(input.ID)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(asset))
}

func (s *Server) getAssets(c *gin.Context) {
	assets, err := s.storage.GetAssets()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	}
	httputil.ResponseSuccess(c, httputil.WithData(assets))
}
