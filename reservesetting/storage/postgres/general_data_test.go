package postgres

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/KyberNetwork/reserve-data/common/testutil"
	"github.com/KyberNetwork/reserve-data/reservesetting/common"
)

func TestGeneralData(t *testing.T) {
	db, tearDown := testutil.MustNewDevelopmentDB(migrationPath)
	defer func() {
		assert.NoError(t, tearDown())
	}()
	s, err := NewStorage(db)
	assert.NoError(t, err)
	var (
		dataTest1 = common.GeneralData{
			Key:   "test data",
			Value: "123.456",
		}
		dataTest2 = common.GeneralData{
			Key:   "test data",
			Value: "456.789",
		}
	)

	_, err = s.GetGeneralData("test data")
	assert.EqualError(t, err, common.ErrNotFound.Error())

	_, err = s.SetGeneralData(dataTest1)
	assert.NoError(t, err)
	dataDB, err := s.GetGeneralData("test data")
	assert.NoError(t, err)
	assert.Equal(t, dataTest1, dataDB)

	_, err = s.SetGeneralData(dataTest2)
	assert.NoError(t, err)
	dataDB, err = s.GetGeneralData("test data")
	assert.NoError(t, err)
	assert.Equal(t, dataTest2, dataDB)
}
