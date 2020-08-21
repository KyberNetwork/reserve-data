package postgres

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/testutil"
	rtypes "github.com/KyberNetwork/reserve-data/lib/rtypes"
	commonv3 "github.com/KyberNetwork/reserve-data/reservesetting/common"
	"github.com/KyberNetwork/reserve-data/reservesetting/storage"
)

func TestStorage_UpdateExchange(t *testing.T) {
	db, tearDown := testutil.MustNewDevelopmentDB(migrationPath)
	defer func() {
		assert.NoError(t, tearDown())
	}()

	s, err := NewStorage(db)
	require.NoError(t, err)

	// expect that exchange are initialized
	exchanges, err := s.GetExchanges()
	require.NoError(t, err)
	assert.Len(t, exchanges, len(common.ValidExchangeNames))

	for _, exchange := range exchanges {
		assert.Zero(t, exchange.TradingFeeMaker)
		assert.Zero(t, exchange.TradingFeeTaker)
		assert.True(t, exchange.Disable)
	}

	// exchange should not be allowed to enable if trading fees are not all set
	err = s.UpdateExchange(rtypes.Huobi,
		storage.UpdateExchangeOpts{
			TradingFeeTaker: commonv3.FloatPointer(0.02),
			Disable:         commonv3.BoolPointer(false),
		})
	assert.Error(t, err)
	assert.Equal(t, commonv3.ErrExchangeFeeMissing, err)

	err = s.UpdateExchange(rtypes.Huobi, storage.UpdateExchangeOpts{
		TradingFeeMaker: commonv3.FloatPointer(0.01),
		TradingFeeTaker: commonv3.FloatPointer(0.02),
	},
	)
	require.NoError(t, err)

	exchanges, err = s.GetExchanges()
	require.NoError(t, err)

	for _, exchange := range exchanges {
		switch exchange.ID {
		case rtypes.Huobi:
			assert.Equal(t, 0.01, exchange.TradingFeeMaker)
			assert.Equal(t, 0.02, exchange.TradingFeeTaker)
			assert.True(t, exchange.Disable)
		case rtypes.Binance:
			assert.Zero(t, exchange.TradingFeeMaker)
			assert.Zero(t, exchange.TradingFeeTaker)
			assert.True(t, exchange.Disable)
		}
	}

	err = s.UpdateExchange(rtypes.Huobi,
		storage.UpdateExchangeOpts{
			TradingFeeMaker: commonv3.FloatPointer(0.01),
			TradingFeeTaker: commonv3.FloatPointer(0.02),
			Disable:         commonv3.BoolPointer(false),
		})
	require.NoError(t, err)

	exchanges, err = s.GetExchanges()
	require.NoError(t, err)
	for _, exchange := range exchanges {
		if exchange.ID == rtypes.Huobi {
			assert.Equal(t, 0.01, exchange.TradingFeeMaker)
			assert.Equal(t, 0.02, exchange.TradingFeeTaker)
			assert.False(t, exchange.Disable)
		}
	}

	huobiExchange, err := s.GetExchange(rtypes.Huobi)
	require.NoError(t, err)
	assert.Equal(t, 0.01, huobiExchange.TradingFeeMaker)
	assert.Equal(t, 0.02, huobiExchange.TradingFeeTaker)
	assert.False(t, huobiExchange.Disable)
}

func TestStorage_GetUpdateByName(t *testing.T) {
	db, tearDown := testutil.MustNewDevelopmentDB(migrationPath)
	defer func() {
		assert.NoError(t, tearDown())
	}()
	s, err := NewStorage(db)
	assert.NoError(t, err)
	initData(t, s)
	exchangeByID, err := s.GetExchange(binance)
	require.NoError(t, err)
	exchangeByName, err := s.GetExchangeByName(exchangeByID.Name)
	require.NoError(t, err)
	require.Equal(t, exchangeByID, exchangeByName)
}
