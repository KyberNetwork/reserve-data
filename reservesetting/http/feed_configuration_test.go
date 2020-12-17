package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1common "github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/feed"
	"github.com/KyberNetwork/reserve-data/common/testutil"
	rtypes "github.com/KyberNetwork/reserve-data/lib/rtypes"
	"github.com/KyberNetwork/reserve-data/reservesetting/common"
	"github.com/KyberNetwork/reserve-data/reservesetting/storage/postgres"
)

func TestCheckFeedConfiguration(t *testing.T) {
	db, tearDown := testutil.MustNewDevelopmentDB(migrationPath)
	defer func() {
		assert.NoError(t, tearDown())
	}()
	s, err := postgres.NewStorage(db)
	require.NoError(t, err)
	supportedExchanges := make(map[rtypes.ExchangeID]v1common.LiveExchange)
	server := NewServer(s, "", supportedExchanges, "", nil, nil, nil)

	feedConfigurationEntryNotSupportedTest := common.SetFeedConfigurationEntry{
		Name: "Not supported feed",
	}
	err = server.checkSetFeedConfigurationParams(feedConfigurationEntryNotSupportedTest)
	require.Error(t, err)

	// test match supported feeds name and set rate value
	// gold feed
	supportedGoldFeed := []feed.Feed{
		feed.OneForgeXAUETH,
		feed.OneForgeXAUUSD, // OneForgeXAUUSD
		feed.GDAXETHUSD,     // GDAXETHUSD
		feed.KrakenETHUSD,   // KrakenETHUSD
		feed.GeminiETHUSD,   // GeminiETHUSD
	}

	for _, feedName := range supportedGoldFeed {
		feedNameSupportedEntry := common.SetFeedConfigurationEntry{
			Name:    feedName.String(),
			SetRate: common.GoldFeed,
		}
		err := server.checkSetFeedConfigurationParams(feedNameSupportedEntry)
		assert.NoError(t, err)
	}

	// test match supported feeds name and set rate value
	// USD feed
	supportedUSDFeed := []feed.Feed{
		feed.CoinbaseETHUSDDAI5000,
		feed.CurveDAIUSDC10000,
	}

	for _, feedName := range supportedUSDFeed {
		feedNameSupportedEntry := common.SetFeedConfigurationEntry{
			Name:    feedName.String(),
			SetRate: common.USDFeed,
		}
		err := server.checkSetFeedConfigurationParams(feedNameSupportedEntry)
		assert.NoError(t, err)
	}

	// test match supported feeds name and set rate value
	// USD feed
	supportedBTCFeed := []feed.Feed{
		feed.CoinbaseETHBTC3,
		feed.BinanceETHBTC3,
	}

	for _, feedName := range supportedBTCFeed {
		feedNameSupportedEntry := common.SetFeedConfigurationEntry{
			Name:    feedName.String(),
			SetRate: common.BTCFeed,
		}
		err := server.checkSetFeedConfigurationParams(feedNameSupportedEntry)
		assert.NoError(t, err)
	}

	// test feed name and set rate value does not match
	feedNameSupportedEntry := common.SetFeedConfigurationEntry{
		Name:    feed.GeminiETHUSD.String(),
		SetRate: common.BTCFeed,
	}
	err = server.checkSetFeedConfigurationParams(feedNameSupportedEntry)
	assert.Error(t, err)
}
