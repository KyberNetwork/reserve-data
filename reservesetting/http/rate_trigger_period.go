package http

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/KyberNetwork/reserve-data/http/httputil"
	"github.com/KyberNetwork/reserve-data/reservesetting/common"
)

const (
	key = "rate_trigger_period"
)

func (s *Server) setRateTriggerPeriod(c *gin.Context) {
	var input struct {
		Value float64 `json:"value" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	if input.Value <= 0 {
		httputil.ResponseFailure(c, httputil.WithError(fmt.Errorf("value must greater than zero, value=%f", input.Value)))
		return
	}
	data := common.RateTriggerPeriod{
		Key:   key,
		Value: input.Value,
	}
	id, err := s.storage.SetGeneralData(data.ToGeneralData())
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(id))
}

func (s *Server) getRateTriggerPeriod(c *gin.Context) {
	gdata, err := s.storage.GetGeneralData(key)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	data, err := gdata.ToRateTriggerPeriod()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}
