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

type feedStatusEntry struct {
	Enabled bool `json:"enabled" binding:"required"`
}

func (s *Server) updateFeedStatus(c *gin.Context) {
	var input struct {
		Name string `uri:"name" binding:"required"`
	}
	if err := c.ShouldBindUri(&input); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	feed, err := s.storage.GetFeedConfiguration(input.Name)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}

	var fse feedStatusEntry
	if err = c.ShouldBindJSON(&fse); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}

	if err := s.storage.UpdateFeedStatus(feed.Name, fse.Enabled); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}
