package postgres

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KyberNetwork/reserve-data/common/testutil"
	"github.com/KyberNetwork/reserve-data/reservesetting/common"

	ethereum "github.com/ethereum/go-ethereum/common"
)

func TestApproveSettingChange(t *testing.T) {
	db, tearDown := testutil.MustNewDevelopmentDB(migrationPath)
	defer func() {
		assert.NoError(t, tearDown())
	}()

	s, err := NewStorage(db)
	require.NoError(t, err)

	stID, err := s.CreateSettingChange(common.ChangeCatalogMain, common.SettingChange{
		ChangeList: []common.SettingChangeEntry{
			{
				Type: common.ChangeTypeCreateAsset,
				Data: common.CreateAssetEntry{
					Symbol:                "KNC",
					Name:                  "KNC",
					Address:               ethereum.HexToAddress(""),
					Decimals:              18,
					NormalUpdatePerPeriod: 1,
					MaxImbalanceRatio:     1,
					OrderDurationMillis:   123123,
				},
			},
		},
	})
	require.NoError(t, err)
	err = s.ApproveSettingChange("123", uint64(stID))
	require.NoError(t, err)
	listApproval, err := s.GetLisApprovalSettingChange(uint64(stID))
	require.NoError(t, err)
	require.Equal(t, 1, len(listApproval))
	require.Equal(t, "123", listApproval[0].KeyID)

	// test disapprove
	err = s.DisapproveSettingChange("123", uint64(stID))
	require.NoError(t, err)
	listApproval, err = s.GetLisApprovalSettingChange(uint64(stID))
	require.NoError(t, err)
	require.Equal(t, 0, len(listApproval))
}
