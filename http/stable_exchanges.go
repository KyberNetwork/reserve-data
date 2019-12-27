package http

import (
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"

	"github.com/KyberNetwork/reserve-data/http/httputil"
)

// GetGoldData return gold data feed
func (s *Server) GetGoldData(c *gin.Context) {
	zap.S().Info("Getting gold data")

	data, err := s.app.GetGoldData(getTimePoint(c, true, s.l))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithData(data))
	}
}

// GetBTCData return BTC data feed
func (s *Server) GetBTCData(c *gin.Context) {
	data, err := s.app.GetBTCData(getTimePoint(c, true, s.l))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithData(data))
	}
}

// GetUSDData return BTC data feed
func (s *Server) GetUSDData(c *gin.Context) {
	data, err := s.app.GetUSDData(getTimePoint(c, true, s.l))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithData(data))
	}
}
