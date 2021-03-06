package http

import (
	"encoding/json"
	"fmt"

	rcommon "github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/http/httputil"
	"github.com/KyberNetwork/reserve-data/reservesetting/common"

	"github.com/gin-gonic/gin"
)

type cronJobInputData struct {
	Endpoint      string      `json:"endpoint" binding:"required"`
	HTTPMethod    string      `json:"http_method" binding:"required"`
	Data          interface{} `json:"data"`
	ScheduledTime uint64      `json:"scheduled_time" binding:"required"`
}

func (s *Server) createScheduledJob(c *gin.Context) {
	var input cronJobInputData
	if err := c.ShouldBindJSON(&input); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	if input.ScheduledTime < rcommon.NowInMillis() {
		httputil.ResponseFailure(c, httputil.WithError(fmt.Errorf("the schedule time is in the past, time=%d", input.ScheduledTime)))
		return
	}
	byteData, err := json.Marshal(input.Data)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	id, err := s.storage.CreateScheduledJob(common.ScheduledJobData{
		Endpoint:      input.Endpoint,
		HTTPMethod:    input.HTTPMethod,
		Data:          byteData,
		ScheduledTime: rcommon.MillisToTime(input.ScheduledTime),
	})
	if err != nil {
		s.l.Errorw("cannot add scheduled job to db", "err", err)
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithField("id", id))
}

func (s *Server) getAllScheduledJob(c *gin.Context) {
	status := c.Query("status")
	data, err := s.storage.GetAllScheduledJob(status)
	if err != nil {
		s.l.Errorw("cannot get scheduled job from db", "err", err)
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (s *Server) getScheduledJob(c *gin.Context) {
	var input struct {
		ID uint64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&input); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	data, err := s.storage.GetScheduledJob(input.ID)
	if err != nil {
		s.l.Errorw("cannot get scheduled job from db", "err", err)
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (s *Server) rejectScheduledJob(c *gin.Context) {
	var input struct {
		ID uint64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&input); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	if err := s.storage.UpdateScheduledJobStatus("canceled", input.ID); err != nil {
		s.l.Errorw("cannot update scheduled job status", "err", err)
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}
