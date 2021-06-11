package postgres

import (
	"database/sql"
	"time"

	v3 "github.com/KyberNetwork/reserve-data/reservesetting/common"
)

type scheduledJobDataInput struct {
	Endpoint      string    `json:"endpoint" db:"endpoint"`
	HTTPMethod    string    `json:"http_method" db:"http_method"`
	Data          []byte    `json:"data" db:"data"`
	ScheduledTime time.Time `json:"scheduled_time" db:"scheduled_time"`
}

func (s *Storage) AddScheduledJob(input v3.ScheduledJobData) (uint64, error) {
	var id uint64
	if err := s.stmts.newScheduledJob.Get(&id, scheduledJobDataInput{
		Endpoint:      input.Endpoint,
		HTTPMethod:    input.HTTPMethod,
		Data:          input.Data,
		ScheduledTime: input.ScheduledTime,
	}); err != nil {
		return 0, err
	}
	return id, nil
}

type scheduledJobDB struct {
	ID            uint64    `db:"id"`
	Endpoint      string    `db:"endpoint"`
	HTTPMethod    string    `db:"http_method"`
	Data          []byte    `db:"data"`
	ScheduledTime time.Time `db:"scheduled_time"`
}

func (s *Storage) GetAllScheduledJob() ([]v3.ScheduledJobData, error) {
	var (
		dbResult []scheduledJobDB
		result   []v3.ScheduledJobData
	)
	if err := s.stmts.getAllScheduledJob.Select(&dbResult, nil); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	for _, dbr := range dbResult {
		result = append(result, v3.ScheduledJobData{
			ID:            dbr.ID,
			Endpoint:      dbr.Endpoint,
			HTTPMethod:    dbr.HTTPMethod,
			Data:          dbr.Data,
			ScheduledTime: dbr.ScheduledTime,
		})
	}
	return result, nil
}

func (s *Storage) GetEligibleScheduledJob() ([]v3.ScheduledJobData, error) {
	var (
		dbResult []scheduledJobDB
		result   []v3.ScheduledJobData
	)
	if err := s.stmts.getAllScheduledJob.Select(&dbResult, time.Now()); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	for _, dbr := range dbResult {
		result = append(result, v3.ScheduledJobData{
			ID:            dbr.ID,
			Endpoint:      dbr.Endpoint,
			HTTPMethod:    dbr.HTTPMethod,
			Data:          dbr.Data,
			ScheduledTime: dbr.ScheduledTime,
		})
	}
	return result, nil
}

func (s *Storage) GetScheduledJob(id uint64) (v3.ScheduledJobData, error) {
	var (
		dbResult scheduledJobDB
	)
	if err := s.stmts.getScheduledJob.Get(&dbResult, id); err != nil {
		return v3.ScheduledJobData{}, err
	}
	return v3.ScheduledJobData{
		ID:            dbResult.ID,
		Endpoint:      dbResult.Endpoint,
		HTTPMethod:    dbResult.HTTPMethod,
		Data:          dbResult.Data,
		ScheduledTime: dbResult.ScheduledTime,
	}, nil
}

func (s *Storage) RemoveScheduledJob(id uint64) error {
	if _, err := s.stmts.deleteScheduledJob.Exec(id); err != nil {
		return err
	}
	return nil
}
