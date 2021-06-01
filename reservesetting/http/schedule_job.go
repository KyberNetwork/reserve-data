package http

import (
	"fmt"
	"strings"

	rcommon "github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/http/httputil"
	"github.com/KyberNetwork/reserve-data/reservesetting/common"

	"github.com/gin-gonic/gin"
)

type cronJobInputData struct {
	Endpoint     string      `json:"endpoint" binding:"required"`
	HTTPMethod   string      `json:"http_method" binding:"required"`
	Data         interface{} `json:"data"`
	ScheduleTime uint64      `json:"schedule_time" binding:"required"`
}

func (s *Server) addScheduleJob(c *gin.Context) {
	var input cronJobInputData
	if err := c.ShouldBindJSON(&input); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	if input.ScheduleTime < rcommon.NowInMillis() {
		httputil.ResponseFailure(c, httputil.WithError(fmt.Errorf("the schedule time is in the past, time=%d", input.ScheduleTime)))
		return
	}
	id, err := s.storage.AddScheduleJob(common.ScheduleJobData{
		Endpoint:     strings.TrimPrefix(input.Endpoint, "/"),
		HTTPMethod:   input.HTTPMethod,
		Data:         input.Data,
		ScheduleTime: rcommon.MillisToTime(input.ScheduleTime),
	})
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithField("id", id))
}

func (s *Server) getAllScheduleJob(c *gin.Context) {
	data, err := s.storage.GetAllScheduleJob()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (s *Server) getScheduleJob(c *gin.Context) {
	var input struct {
		ID uint64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&input); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	data, err := s.storage.GetScheduleJob(input.ID)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (s *Server) removeScheduleJob(c *gin.Context) {
	var input struct {
		ID uint64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&input); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	if err := s.storage.RemoveScheduleJob(input.ID); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}
