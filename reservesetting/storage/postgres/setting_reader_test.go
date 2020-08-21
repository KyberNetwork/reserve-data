package postgres

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KyberNetwork/reserve-data/common/testutil"
	"github.com/KyberNetwork/reserve-data/lib/rtypes"
	// "github.com/KyberNetwork/reserve-data/reservesetting/common"
	// "github.com/KyberNetwork/reserve-data/reservesetting/storage"
)

func TestStorage_GetAssetBySymbol(t *testing.T) {
	db, tearDown := testutil.MustNewDevelopmentDB(migrationPath)
	defer func() {
		assert.NoError(t, tearDown())
	}()
	s, err := NewStorage(db)
	require.NoError(t, err)

	initData(t, s)

	asset, err := s.GetAssetBySymbol("BTC")
	require.NoError(t, err)
	realAsset, err := s.GetAsset(asset.ID)
	require.NoError(t, err)
	require.Equal(t, float64(13), realAsset.StableParam.SingleFeedMaxSpread)
	require.Equal(t, float64(0), realAsset.StableParam.MultipleFeedsMaxDiff)
}

func TestStorage_GetTransferableAssets(t *testing.T) {
	db, tearDown := testutil.MustNewDevelopmentDB(migrationPath)
	defer func() {
		assert.NoError(t, tearDown())
	}()
	s, err := NewStorage(db)
	require.NoError(t, err)

	initData(t, s)
	assets, err := s.GetTransferableAssets()
	require.NoError(t, err)
	for _, asset := range assets {
		a, err := s.GetAsset(asset.ID)
		require.NoError(t, err)
		require.True(t, a.Transferable)
	}
}

func TestStorage_GetTradingPair(t *testing.T) {
	db, tearDown := testutil.MustNewDevelopmentDB(migrationPath)
	defer func() {
		assert.NoError(t, tearDown())
	}()
	s, err := NewStorage(db)
	require.NoError(t, err)

	initData(t, s)
	_, err = s.GetTradingPair(1, false)
	require.NoError(t, err)
}

func TestStorage_GetTradingPairs(t *testing.T) {
	db, tearDown := testutil.MustNewDevelopmentDB(migrationPath)
	defer func() {
		assert.NoError(t, tearDown())
	}()
	s, err := NewStorage(db)
	require.NoError(t, err)

	initData(t, s)
	tps, err := s.GetTradingPairs(1)
	require.NoError(t, err)
	require.NotZero(t, len(tps))
}

func TestStorage_GetMinNotional(t *testing.T) {
	db, tearDown := testutil.MustNewDevelopmentDB(migrationPath)
	defer func() {
		assert.NoError(t, tearDown())
	}()
	s, err := NewStorage(db)
	require.NoError(t, err)

	initData(t, s)
	minNotional, err := s.GetMinNotional(rtypes.Binance, 2, 1)
	require.NoError(t, err)
	require.Equal(t, float64(0), minNotional)
}
