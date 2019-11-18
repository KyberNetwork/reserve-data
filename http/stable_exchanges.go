package http

import (
	"encoding/json"
	"log"

	"github.com/KyberNetwork/reserve-data/common"

	"github.com/KyberNetwork/reserve-data/http/httputil"
	"github.com/gin-gonic/gin"
)

func (s *Server) GetGoldData(c *gin.Context) {
	log.Printf("Getting gold data")

	data, err := s.app.GetGoldData(getTimePoint(c, true))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithData(data))
	}
}

func (s *Server) GetBTCData(c *gin.Context) {
	data, err := s.app.GetBTCData(getTimePoint(c, true))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithData(data))
	}
}

func (s *Server) GetUSDData(c *gin.Context) {
	data, err := s.app.GetUSDData(getTimePoint(c, true))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
	} else {
		httputil.ResponseSuccess(c, httputil.WithData(data))
	}
}

func (s *Server) UpdateFeedConfiguration(c *gin.Context) {
	const dataPostFormKey = "data"

	postForm, ok := s.Authenticated(c, []string{dataPostFormKey}, []Permission{ConfirmConfPermission})
	if !ok {
		return
	}

	data := []byte(postForm.Get(dataPostFormKey))
	if len(data) > maxDataSize {
		httputil.ResponseFailure(c, httputil.WithError(errDataSizeExceed))
		return
	}

	var input common.FeedConfiguration
	if err := json.Unmarshal(data, &input); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}

	if err := s.app.UpdateFeedConfiguration(input.Name, input.Enabled); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (s *Server) GetFeedConfiguration(c *gin.Context) {
	_, ok := s.Authenticated(c, []string{}, []Permission{ReadOnlyPermission, RebalancePermission, ConfigurePermission, ConfirmConfPermission})
	if !ok {
		return
	}

	data, err := s.app.GetFeedConfiguration()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}
