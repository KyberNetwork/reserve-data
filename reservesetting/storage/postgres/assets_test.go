package postgres

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KyberNetwork/reserve-data/common/testutil"
	"github.com/KyberNetwork/reserve-data/lib/rtypes"
	"github.com/KyberNetwork/reserve-data/reservesetting/common"
)

const (
	migrationPath = "../../../cmd/migrations"
)

func TestUpdateAsset(t *testing.T) {
	db, tearDown := testutil.MustNewDevelopmentDB(migrationPath)
	defer func() {
		assert.NoError(t, tearDown())
	}()
	s, err := NewStorage(db)
	assert.NoError(t, err)
	initData(t, s)
	assetID := rtypes.AssetID(1)
	target := &common.AssetTarget{
		Total:              100.0,
		Reserve:            101.0,
		RebalanceThreshold: 102.0,
		TransferThreshold:  103.0,
	}
	pwi := &common.AssetPWI{
		Ask: common.PWIEquation{
			A:                   1.1,
			B:                   1.2,
			C:                   1.3,
			MinMinSpread:        1.4,
			PriceMultiplyFactor: 1.5,
		},
		Bid: common.PWIEquation{
			A:                   2.1,
			B:                   2.2,
			C:                   2.3,
			MinMinSpread:        2.4,
			PriceMultiplyFactor: 2.5,
		},
	}
	rebalance := &common.RebalanceQuadratic{
		SizeA:  3.1,
		SizeB:  3.2,
		SizeC:  3.3,
		PriceA: 9.1,
		PriceB: 15.2,
		PriceC: 21.3,
	}
	stableParam := &common.UpdateStableParam{
		PriceUpdateThreshold: common.FloatPointer(1),
		AskSpread:            common.FloatPointer(2),
		BidSpread:            common.FloatPointer(3),
	}

	var tests = []struct {
		msg      string
		data     common.SettingChange
		assertFn func(*testing.T, common.Asset, error)
	}{
		{
			msg: "test update target",
			data: common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeUpdateAsset,
						Data: common.UpdateAssetEntry{
							AssetID: assetID,
							Target:  target,
						},
					},
				},
			},
			assertFn: func(t *testing.T, a common.Asset, e error) {
				assert.NoError(t, e)
				assert.Equal(t, target, a.Target)
			},
		},
		{
			msg: "test update pwis",
			data: common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeUpdateAsset,
						Data: common.UpdateAssetEntry{
							AssetID: assetID,
							PWI:     pwi,
						},
					},
				},
			},
			assertFn: func(t *testing.T, a common.Asset, e error) {
				assert.NoError(t, e)
				assert.Equal(t, pwi, a.PWI)
			},
		},
		{
			msg: "test update rebalanceQuadratic",
			data: common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeUpdateAsset,
						Data: common.UpdateAssetEntry{
							AssetID:            assetID,
							RebalanceQuadratic: rebalance,
						},
					},
				},
			},
			assertFn: func(t *testing.T, a common.Asset, e error) {
				assert.NoError(t, e)
				assert.Equal(t, rebalance, a.RebalanceQuadratic)
			},
		},
		{
			msg: "test update stable params",
			data: common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeUpdateAsset,
						Data: common.UpdateAssetEntry{
							AssetID:     assetID,
							StableParam: stableParam,
						},
					},
				},
			},
			assertFn: func(t *testing.T, asset common.Asset, e error) {
				assert.NoError(t, e)
				assert.Equal(t, *stableParam.AskSpread, asset.StableParam.AskSpread)
				assert.Equal(t, *stableParam.PriceUpdateThreshold, asset.StableParam.PriceUpdateThreshold)
				assert.Equal(t, float64(0), asset.StableParam.MultipleFeedsMaxDiff)
			},
		},
	}
	for _, tc := range tests {
		t.Logf("running test case for: %s", tc.msg)
		id, err := s.CreateSettingChange(common.ChangeCatalogMain, tc.data)
		assert.NoError(t, err)
		err = s.ConfirmSettingChange(id, true)
		require.NoError(t, err)
		asset, err := s.GetAsset(assetID)
		require.NoError(t, err)
		tc.assertFn(t, asset, err)
	}
}

func TestGetAssetBySymbol(t *testing.T) {
	db, tearDown := testutil.MustNewDevelopmentDB(migrationPath)
	defer func() {
		assert.NoError(t, tearDown())
	}()
	s, err := NewStorage(db)
	assert.NoError(t, err)
	initData(t, s)
	assetByID, err := s.GetAsset(3)
	require.NoError(t, err)
	assetBySymbol, err := s.GetAssetBySymbol(assetByID.Symbol)
	require.NoError(t, err)
	require.Equal(t, assetByID.ID, assetBySymbol.ID)
	require.Equal(t, assetByID.Decimals, assetBySymbol.Decimals)
}
