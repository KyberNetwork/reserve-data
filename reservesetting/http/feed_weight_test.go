package http

import (
	"testing"

	"github.com/KyberNetwork/reserve-data/common/feed"
	"github.com/KyberNetwork/reserve-data/reservesetting/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFeedWeight(t *testing.T) {

	var (
		testSetRateBTCFeed           = common.BTCFeed
		testFeedWeightIsNotSupported = common.FeedWeight{
			"NotSupportedFeed": 0.341432,
		}
	)
	// Test feed weight is not supported
	err := checkFeedWeight(&testSetRateBTCFeed, &testFeedWeightIsNotSupported)
	require.NotNil(t, err)

	// Test feed weight is supported
	testFeedWeightIsSupported := common.FeedWeight{
		feed.CoinbaseETHBTC3.String(): 0.1,
		feed.BinanceETHBTC3.String():  0.2,
	}
	err = checkFeedWeight(&testSetRateBTCFeed, &testFeedWeightIsSupported)
	assert.NoError(t, err)

	// Test feed weight USD Feed
	testSetRateUSDFeed := common.USDFeed
	err = checkFeedWeight(&testSetRateUSDFeed, &testFeedWeightIsNotSupported)
	require.NotNil(t, err)

	testFeedWeightUSDIsSupported := common.FeedWeight{
		feed.CoinbaseETHUSDDAI5000.String(): 0.1,
		feed.CurveDAIUSDC10000.String():     0.2,
	}
	err = checkFeedWeight(&testSetRateUSDFeed, &testFeedWeightUSDIsSupported)
	assert.NoError(t, err)
}
