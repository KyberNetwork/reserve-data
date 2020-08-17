package postgres

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KyberNetwork/reserve-data/common/testutil"
	"github.com/KyberNetwork/reserve-data/reservesetting/common"
	"github.com/KyberNetwork/reserve-data/reservesetting/storage"
)

func TestStorage_UpdateTradingPair(t *testing.T) {
	db, tearDown := testutil.MustNewDevelopmentDB(migrationPath)
	defer func() {
		assert.NoError(t, tearDown())
	}()
	s, err := NewStorage(db)
	require.NoError(t, err)

	initData(t, s)
	tp, err := s.GetTradingPair(1, false)
	require.NoError(t, err)
	expectedTradingPair := tp
	expectedTradingPair.PricePrecision = 1
	expectedTradingPair.AmountPrecision = 1
	expectedTradingPair.AmountLimitMin = 0.1
	expectedTradingPair.AmountLimitMax = 0.1
	expectedTradingPair.PriceLimitMin = 0.1
	expectedTradingPair.PriceLimitMax = 0.1
	expectedTradingPair.MinNotional = 0.1
	err = s.UpdateTradingPair(tp.ID, storage.UpdateTradingPairOpts{
		ID:              tp.ID,
		PricePrecision:  common.Uint64Pointer(1),
		AmountPrecision: common.Uint64Pointer(1),
		AmountLimitMin:  common.FloatPointer(0.1),
		AmountLimitMax:  common.FloatPointer(0.1),
		PriceLimitMin:   common.FloatPointer(0.1),
		PriceLimitMax:   common.FloatPointer(0.1),
		MinNotional:     common.FloatPointer(0.1),
	})
	require.NoError(t, err)
	updatedTradingPair, err := s.GetTradingPair(1, false)
	require.NoError(t, err)
	require.Equal(t, expectedTradingPair, updatedTradingPair)
	require.Equal(t, binance, updatedTradingPair.ExchangeID)
}
