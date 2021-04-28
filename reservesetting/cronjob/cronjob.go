package cronjob

import (
	"fmt"
	"sync"
	"time"

	cjv3 "github.com/robfig/cron/v3"
	"go.uber.org/zap"

	nh "github.com/KyberNetwork/reserve-data/lib/noauth-http"
	"github.com/KyberNetwork/reserve-data/reservesetting/common"
	"github.com/KyberNetwork/reserve-data/reservesetting/storage"
)

type jobID uint64

type CronJob struct {
	l          *zap.SugaredLogger
	cli        *nh.Client
	s          storage.Interface
	settingURL string
	jobs       map[jobID]*cjv3.Cron
	mu         sync.Mutex
}

func NewCronJob(s storage.Interface, settingURL string) *CronJob {
	return &CronJob{
		l:          zap.S(),
		cli:        nh.New(),
		s:          s,
		settingURL: settingURL,
		jobs:       make(map[jobID]*cjv3.Cron),
		mu:         sync.Mutex{},
	}
}

func (cj *CronJob) Init() error {
	jobs, err := cj.s.GetCronJob()
	if err != nil {
		return err
	}
	for _, j := range jobs {
		if err := cj.AddJob(j, false); err != nil {
			return err
		}
	}
	fmt.Println(cj.jobs)
	return nil
}

func (cj *CronJob) AddJob(jd common.CronJobData, shouldStore bool) error {
	var (
		jID uint64 = jd.ID
		err error
	)
	if shouldStore {
		jID, err = cj.s.AddCronJob(jd)
		if err != nil {
			cj.l.Errorw("failed to add job to db", "err", err)
			return err
		}
		jd.ID = jID
	}
	c := cjv3.New(cjv3.WithLocation(time.UTC))
	cj.l.Infow("spec", "value", timeToSpec(jd.ScheduleTime.UTC()))
	_, err = c.AddFunc(timeToSpec(jd.ScheduleTime.UTC()), cj.generateJob(jd))
	if err != nil {
		return err
	}
	cj.mu.Lock()
	cj.jobs[jobID(jID)] = c
	cj.mu.Unlock()
	c.Start()
	return nil
}

func (cj *CronJob) generateJob(jd common.CronJobData) func() {
	return func() {
		cj.l.Infow("do itttttt")
		_, err := cj.cli.DoReq(fmt.Sprintf("%s/%s", cj.settingURL, jd.Endpoint), jd.HTTPMethod, jd.Data)
		if err != nil {
			cj.l.Errorw("request failed", "err", err)
		}
		if err := cj.RemoveJob(jd.ID); err != nil {
			cj.l.Errorw("failed to remove job", "err", err)
		}
	}
}

func (cj *CronJob) RemoveJob(id uint64) error {
	c, ok := cj.jobs[jobID(id)]
	if !ok {
		cj.l.Errorw("job is not available", "id", id)
		return fmt.Errorf("job %d is not available", id)
	}
	c.Stop()
	cj.mu.Lock()
	delete(cj.jobs, jobID(id))
	cj.mu.Unlock()
	if err := cj.s.RemoveCronJob(uint64(id)); err != nil {
		cj.l.Errorw("failed to remove cronjob from db", "id", id)
		return err
	}
	return nil
}

func timeToSpec(t time.Time) string {
	return fmt.Sprintf("%d %d %d %d *", t.Minute(), t.Hour(), t.Day(), t.Month())
}
