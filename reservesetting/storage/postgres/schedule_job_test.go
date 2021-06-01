package postgres

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KyberNetwork/reserve-data/common/testutil"
	"github.com/KyberNetwork/reserve-data/reservesetting/common"
)

func TestScheduleJob(t *testing.T) {
	db, tearDown := testutil.MustNewDevelopmentDB(migrationPath)
	defer func() {
		assert.NoError(t, tearDown())
	}()

	s, err := NewStorage(db)
	require.NoError(t, err)

	testData := common.ScheduleJobData{
		Endpoint:     "test endpoint",
		HTTPMethod:   http.MethodPost,
		Data:         nil,
		ScheduleTime: time.Now(),
	}
	id, err := s.AddScheduleJob(testData)
	require.NoError(t, err)

	testData.ID = id
	sj, err := s.GetScheduleJob(id)
	require.NoError(t, err)

	require.Equal(t, testData.ScheduleTime.Unix(), sj.ScheduleTime.Unix())
	require.Equal(t, testData.Endpoint, sj.Endpoint)
	require.Equal(t, testData.HTTPMethod, sj.HTTPMethod)
	require.Equal(t, testData.Data, sj.Data)

	err = s.RemoveScheduleJob(id)
	require.NoError(t, err)

	allSJ, err := s.GetAllScheduleJob()
	require.NoError(t, err)
	require.Zero(t, len(allSJ))
}
