package scheduledjob

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"time"

	"go.uber.org/zap"

	marketdata "github.com/KyberNetwork/reserve-data/reservesetting/market-data"
	"github.com/KyberNetwork/reserve-data/reservesetting/storage"
)

type ScheduledJob struct {
	l                      *zap.SugaredLogger
	s                      storage.Interface
	md                     *marketdata.MarketData
	httpHandler            http.Handler
	numberApprovalRequired int
}

func NewScheduledJob(s storage.Interface, md *marketdata.MarketData, httpHandler http.Handler, numberApprovalRequired int) *ScheduledJob {
	return &ScheduledJob{
		l:                      zap.S(),
		s:                      s,
		md:                     md,
		httpHandler:            httpHandler,
		numberApprovalRequired: numberApprovalRequired,
	}
}

func (sj *ScheduledJob) Run(interval time.Duration) {
	for {
		sj.ExecuteScheduledSettingChange()
		sj.ExecuteEligibleScheduledJob()
		time.Sleep(interval)
	}
}

func (sj *ScheduledJob) ExecuteEligibleScheduledJob() {
	l := sj.l.With("func", "ExecuteEligibleScheduledJob")
	jobs, err := sj.s.GetEligibleScheduledJob()
	if err != nil {
		l.Errorw("cannot get job from db", "err", err)
		return
	}
	for _, j := range jobs {
		l.Infow("job info", "info", j)
		req, err := http.NewRequest(j.HTTPMethod, j.Endpoint, bytes.NewReader(j.Data))
		if err != nil {
			sj.l.Errorw("failed to build request", "err", err)
			continue
		}
		req.Header.Add("Content-Type", "application/json")
		res := httptest.NewRecorder()
		sj.httpHandler.ServeHTTP(res, req)
		if res.Code != http.StatusOK {
			l.Errorw("received unexpected code", "code", res.Code, "job id", j.ID)
		}
		if err := sj.s.UpdateScheduledJobStatus("done", j.ID); err != nil {
			l.Errorw("failed to update job's status", "job id", j.ID)
		}
	}
}

func (sj *ScheduledJob) ExecuteScheduledSettingChange() {
	l := sj.l.With("func", "ExecuteScheduledSettingChange")
	ids, err := sj.s.GetScheduledSettingChange()
	if err != nil {
		l.Errorw("cannot get job from db", "err", err)
		return
	}
	if len(ids) > 0 {
		l.Infow("list scheduled setting change will be executed", "ids", ids)
	}
	for _, id := range ids {
		la, err := sj.s.GetListApprovalSettingChange(uint64(id))
		if err != nil {
			if len(la) < sj.numberApprovalRequired {
				l.Warnw("it's time to apply the scheduled setting change but it still doesn't have enough approval", "id", id)
				continue
			}
		}
		additionalDataReturn, err := sj.s.ConfirmSettingChange(id, true)
		if err != nil {
			l.Errorw("cannot confirm setting change", "err", err, "id", id)
			continue
		}
		if err := sj.md.TryToAddFeed(additionalDataReturn); err != nil {
			l.Errorw("cannot add feed to market data", "err", err)
			continue
		}
	}
}
