package postgres

import (
	"encoding/json"
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

	testJSON, err := json.Marshal(nil)
	require.NoError(t, err)

	testData := common.ScheduledJobData{
		Endpoint:      "test endpoint",
		HTTPMethod:    http.MethodPost,
		Data:          testJSON,
		ScheduledTime: time.Now(),
	}
	id, err := s.CreateScheduledJob(testData)
	require.NoError(t, err)

	testData.ID = id
	sj, err := s.GetScheduledJob(id)
	require.NoError(t, err)

	require.Equal(t, testData.ScheduledTime.Unix(), sj.ScheduledTime.Unix())
	require.Equal(t, testData.Endpoint, sj.Endpoint)
	require.Equal(t, testData.HTTPMethod, sj.HTTPMethod)
	require.Equal(t, testData.Data, sj.Data)

	err = s.UpdateScheduledJobStatus("canceled", id)
	require.NoError(t, err)

	allSJ, err := s.GetAllScheduledJob("")
	require.NoError(t, err)
	require.Zero(t, len(allSJ))
}
