package postgres

import (
	"database/sql"
	"encoding/json"
	"time"

	pgutil "github.com/KyberNetwork/reserve-data/common/postgres"
	v3 "github.com/KyberNetwork/reserve-data/reservesetting/common"

	"github.com/pkg/errors"
)

type scheduleJobDataInput struct {
	Endpoint     string    `json:"endpoint" db:"endpoint"`
	HTTPMethod   string    `json:"http_method" db:"http_method"`
	Data         []byte    `json:"data" db:"data"`
	ScheduleTime time.Time `json:"schedule_time" db:"schedule_time"`
}

func (s *Storage) AddScheduleJob(input v3.ScheduleJobData) (uint64, error) {
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
	if err = tx.NamedStmt(s.stmts.newScheduleJob).Get(&id, scheduleJobDataInput{
		Endpoint:     input.Endpoint,
		HTTPMethod:   input.HTTPMethod,
		Data:         byteData,
		ScheduleTime: input.ScheduleTime,
	}); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		s.l.Errorw("cannot add new job", "err", err)
		return 0, err
	}
	return id, nil
}

type scheduleJobDB struct {
	ID           uint64    `db:"id"`
	Endpoint     string    `db:"endpoint"`
	HTTPMethod   string    `db:"http_method"`
	Data         []byte    `db:"data"`
	ScheduleTime time.Time `db:"schedule_time"`
}

func (s *Storage) GetAllScheduleJob() ([]v3.ScheduleJobData, error) {
	var (
		dbResult []scheduleJobDB
		result   []v3.ScheduleJobData
	)
	if err := s.stmts.getAllScheduleJob.Select(&dbResult, nil); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	for _, dbr := range dbResult {
		var data interface{}
		if err := json.Unmarshal(dbr.Data, &data); err != nil {
			return nil, err
		}
		result = append(result, v3.ScheduleJobData{
			ID:           dbr.ID,
			Endpoint:     dbr.Endpoint,
			HTTPMethod:   dbr.HTTPMethod,
			Data:         data,
			ScheduleTime: dbr.ScheduleTime,
		})
	}
	return result, nil
}

func (s *Storage) GetEligibleScheduleJob() ([]v3.ScheduleJobData, error) {
	var (
		dbResult []scheduleJobDB
		result   []v3.ScheduleJobData
	)
	if err := s.stmts.getAllScheduleJob.Select(&dbResult, time.Now()); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	for _, dbr := range dbResult {
		var data interface{}
		if err := json.Unmarshal(dbr.Data, &data); err != nil {
			return nil, err
		}
		result = append(result, v3.ScheduleJobData{
			ID:           dbr.ID,
			Endpoint:     dbr.Endpoint,
			HTTPMethod:   dbr.HTTPMethod,
			Data:         data,
			ScheduleTime: dbr.ScheduleTime,
		})
	}
	return result, nil
}

func (s *Storage) GetScheduleJob(id uint64) (v3.ScheduleJobData, error) {
	var (
		dbResult scheduleJobDB
	)
	if err := s.stmts.getScheduleJob.Get(&dbResult, id); err != nil {
		return v3.ScheduleJobData{}, err
	}
	var data interface{}
	if err := json.Unmarshal(dbResult.Data, &data); err != nil {
		return v3.ScheduleJobData{}, err
	}
	return v3.ScheduleJobData{
		ID:           dbResult.ID,
		Endpoint:     dbResult.Endpoint,
		HTTPMethod:   dbResult.HTTPMethod,
		Data:         data,
		ScheduleTime: dbResult.ScheduleTime,
	}, nil
}

func (s *Storage) RemoveScheduleJob(id uint64) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return errors.Wrap(err, "create transaction error")
	}
	defer pgutil.RollbackUnlessCommitted(tx)
	if _, err = tx.Stmtx(s.stmts.deleteScheduleJob).Exec(id); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		s.l.Errorw("cannot remove job", "err", err)
		return err
	}
	return nil
}
