package http

import (
	"github.com/gin-gonic/gin"

	"github.com/KyberNetwork/reserve-data/http/httputil"
)

func (s *Server) getFeedConfigurations(c *gin.Context) {
	feedConfigurations, err := s.storage.GetFeedConfigurations()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(feedConfigurations))
}
