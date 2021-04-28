package postgres

import (
	"database/sql"
	"encoding/json"
	"time"

	pgutil "github.com/KyberNetwork/reserve-data/common/postgres"
	v3 "github.com/KyberNetwork/reserve-data/reservesetting/common"

	"github.com/pkg/errors"
)

type cronJobDataInput struct {
	Endpoint     string    `json:"endpoint" db:"endpoint"`
	HTTPMethod   string    `json:"http_method" db:"http_method"`
	Data         []byte    `json:"data" db:"data"`
	ScheduleTime time.Time `json:"schedule_time" db:"schedule_time"`
}

func (s *Storage) AddCronJob(input v3.CronJobData) (uint64, error) {
	byteData, err := json.Marshal(input.Data)
	if err != nil {
		return 0, errors.Wrap(err, "cannot marshal data")
	}
	tx, err := s.db.Beginx()
	if err != nil {
		return 0, errors.Wrap(err, "create transaction error")
	}
	defer pgutil.RollbackUnlessCommitted(tx)
	var id uint64
	if err = tx.NamedStmt(s.stmts.newCronJob).Get(&id, cronJobDataInput{
		Endpoint:     input.Endpoint,
		HTTPMethod:   input.HTTPMethod,
		Data:         byteData,
		ScheduleTime: input.ScheduleTime,
	}); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		s.l.Errorw("cannot add new cronjob", "err", err)
		return 0, err
	}
	return id, nil
}

type cronJobDB struct {
	ID           uint64    `db:"id"`
	Endpoint     string    `db:"endpoint"`
	HTTPMethod   string    `db:"http_method"`
	Data         []byte    `db:"data"`
	ScheduleTime time.Time `db:"schedule_time"`
}

func (s *Storage) GetCronJob() ([]v3.CronJobData, error) {
	var (
		dbResult []cronJobDB
		result   []v3.CronJobData
	)
	if err := s.stmts.getCronJob.Select(&dbResult); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	s.l.Infow("data", "data", dbResult)
	for _, dbr := range dbResult {
		var data interface{}
		if err := json.Unmarshal(dbr.Data, &data); err != nil {
			return nil, err
		}
		result = append(result, v3.CronJobData{
			ID:           dbr.ID,
			Endpoint:     dbr.Endpoint,
			HTTPMethod:   dbr.HTTPMethod,
			Data:         data,
			ScheduleTime: dbr.ScheduleTime,
		})
	}
	return result, nil
}

func (s *Storage) RemoveCronJob(id uint64) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return errors.Wrap(err, "create transaction error")
	}
	defer pgutil.RollbackUnlessCommitted(tx)
	if _, err = tx.Stmtx(s.stmts.deleteCronJob).Exec(id); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		s.l.Errorw("cannot remove cronjob", "err", err)
		return err
	}
	return nil
}
