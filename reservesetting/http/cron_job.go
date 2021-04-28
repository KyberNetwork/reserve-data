package http

import (
	rcommon "github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/http/httputil"
	"github.com/KyberNetwork/reserve-data/reservesetting/common"

	"github.com/gin-gonic/gin"
)

type cronJobInputData struct {
	Endpoint     string      `json:"endpoint"`
	HTTPMethod   string      `json:"http_method"`
	Data         interface{} `json:"data"`
	ScheduleTime uint64      `json:"schedule_time"`
}

func (s *Server) addCronJob(c *gin.Context) {
	var input cronJobInputData
	if err := c.ShouldBindJSON(&input); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	if err := s.cj.AddJob(common.CronJobData{
		Endpoint:     input.Endpoint,
		HTTPMethod:   input.HTTPMethod,
		Data:         input.Data,
		ScheduleTime: rcommon.MillisToTime(input.ScheduleTime).UTC(),
	}, true); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (s *Server) getCronJob(c *gin.Context) {
	data, err := s.storage.GetCronJob()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (s *Server) stopCronJob(c *gin.Context) {
	var input struct {
		ID uint64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&input); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	if err := s.cj.RemoveJob(input.ID); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}
