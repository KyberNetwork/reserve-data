package postgres

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KyberNetwork/reserve-data/common/feed"
	"github.com/KyberNetwork/reserve-data/common/testutil"
	"github.com/KyberNetwork/reserve-data/reservesetting/common"
)

func TestGetFeedConfigurations(t *testing.T) {
	db, tearDown := testutil.MustNewDevelopmentDB(migrationPath)
	defer func() {
		assert.NoError(t, tearDown())
	}()

	s, err := NewStorage(db)
	require.NoError(t, err)

	fcs, err := s.GetFeedConfigurations()
	require.NoError(t, err)

	require.Equal(t, len(feed.AllFeeds().Gold)+len(feed.AllFeeds().USD)+len(feed.AllFeeds().BTC), len(fcs))
}

func TestSetFeedConfigurations(t *testing.T) {
	db, tearDown := testutil.MustNewDevelopmentDB(migrationPath)
	defer func() {
		assert.NoError(t, tearDown())
	}()

	s, err := NewStorage(db)
	require.NoError(t, err)

	fcs, err := s.GetFeedConfigurations()
	require.NoError(t, err)

	require.NotZero(t, len(fcs))
	require.Equal(t, len(feed.AllFeeds().Gold)+len(feed.AllFeeds().USD)+len(feed.AllFeeds().BTC), len(fcs))

	var (
		fname                 = feed.CoinbaseETHDAI10000.String()
		fenabled              = false
		fbaseVolatilitySpread = 1.1
		fnormalSpread         = 1.2
		testFeedData          = common.SetFeedConfigurationEntry{
			Name:                 fname,
			SetRate:              common.USDFeed,
			Enabled:              &fenabled,
			BaseVolatilitySpread: &fbaseVolatilitySpread,
			NormalSpread:         &fnormalSpread,
		}
		expectFC = common.FeedConfiguration{
			Name:                 fname,
			SetRate:              common.USDFeed,
			Enabled:              fenabled,
			BaseVolatilitySpread: fbaseVolatilitySpread,
			NormalSpread:         fnormalSpread,
		}
	)

	err = s.setFeedConfiguration(nil, testFeedData)
	require.NoError(t, err)
	newFC, err := s.GetFeedConfiguration(fname, common.USDFeed)
	require.NoError(t, err)

	require.Equal(t, expectFC, newFC)
}

func TestUpdateFeedStatus(t *testing.T) {
	db, tearDown := testutil.MustNewDevelopmentDB(migrationPath)
	defer func() {
		assert.NoError(t, tearDown())
	}()

	s, err := NewStorage(db)
	require.NoError(t, err)
	fname := feed.CoinbaseETHDAI10000.String()
	fcs, err := s.GetFeedConfiguration(fname, common.USDFeed)
	require.NoError(t, err)
	var (
		newStatus = !fcs.Enabled
	)
	err = s.UpdateFeedStatus(fname, common.USDFeed, newStatus)
	require.NoError(t, err)

	newFeed, err := s.GetFeedConfiguration(fname, common.USDFeed)
	require.NoError(t, err)

	require.Equal(t, newStatus, newFeed.Enabled)
}
