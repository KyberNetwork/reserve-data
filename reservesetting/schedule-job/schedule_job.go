package schedulejob

import (
	"fmt"
	"time"

	"go.uber.org/zap"

	nh "github.com/KyberNetwork/reserve-data/lib/noauth-http"
	"github.com/KyberNetwork/reserve-data/reservesetting/storage"
)

type ScheduleJob struct {
	l          *zap.SugaredLogger
	cli        *nh.Client
	s          storage.Interface
	settingURL string
}

func NewScheduleJob(s storage.Interface, settingURL string) *ScheduleJob {
	return &ScheduleJob{
		l:          zap.S(),
		cli:        nh.New(),
		s:          s,
		settingURL: settingURL,
	}
}

func (cj *ScheduleJob) Run(interval time.Duration) {
	l := cj.l.With("func", "Run")
	for {
		func() {
			jobs, err := cj.s.GetAllScheduleJob()
			if err != nil {
				l.Errorw("cannot get job from db", "err", err)
				return
			}
			for _, j := range jobs {
				if time.Since(j.ScheduleTime) > 0 {
					_, err := cj.cli.DoReq(fmt.Sprintf("%s/%s", cj.settingURL, j.Endpoint), j.HTTPMethod, j.Data)
					if err != nil {
						cj.l.Errorw("failed to execute request", "err", err)
					}
					if err := cj.s.RemoveScheduleJob(j.ID); err != nil {
						cj.l.Errorw("failed to remove the job from db", "id", j.ID)
					}
				}
			}
		}()
		time.Sleep(interval)
	}
}
