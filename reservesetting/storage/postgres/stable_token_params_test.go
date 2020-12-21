package postgres

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KyberNetwork/reserve-data/common/testutil"
	"github.com/KyberNetwork/reserve-data/reservesetting/common"
)

func TestStorage_GetStableTokenParams(t *testing.T) {
	var params = map[string]interface{}{
		"a": "1",
		"b": 2,
	}
	db, tearDown := testutil.MustNewDevelopmentDB(migrationPath)
	defer func() {
		assert.NoError(t, tearDown())
	}()
	s, err := NewStorage(db)
	assert.NoError(t, err)

	id, err := s.CreateSettingChange(common.ChangeCatalogStableToken, common.SettingChange{
		ChangeList: []common.SettingChangeEntry{
			{
				Type: common.ChangeTypeUpdateStableTokenParams,
				Data: common.UpdateStableTokenParamsEntry{
					Params: params,
				},
			},
		},
	}, "")
	require.NoError(t, err)
	_, err = s.ConfirmSettingChange(id, true)
	require.NoError(t, err)
	outputParams, err := s.GetStableTokenParams()
	require.NoError(t, err)
	t.Log(outputParams)
	require.Contains(t, outputParams, "a")
	require.Equal(t, "1", outputParams["a"])
}
