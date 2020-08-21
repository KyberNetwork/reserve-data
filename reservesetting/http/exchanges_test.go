package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1common "github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/testutil"
	"github.com/KyberNetwork/reserve-data/lib/rtypes"
	"github.com/KyberNetwork/reserve-data/reservesetting/common"
	"github.com/KyberNetwork/reserve-data/reservesetting/storage"
	"github.com/KyberNetwork/reserve-data/reservesetting/storage/postgres"
)

func TestExchanges(t *testing.T) {
	var (
		supportedExchanges = make(map[rtypes.ExchangeID]v1common.LiveExchange)
	)

	// create map of test exchange
	for _, exchangeID := range []rtypes.ExchangeID{rtypes.Binance, rtypes.Huobi} {
		exchange := v1common.TestExchange{}
		supportedExchanges[exchangeID] = exchange
	}

	db, tearDown := testutil.MustNewDevelopmentDB(migrationPath)
	defer func() {
		assert.NoError(t, tearDown())
	}()
	s, err := postgres.NewStorage(db)
	require.NoError(t, err)
	server := NewServer(s, "", supportedExchanges, "", nil, nil)
	c := apiClient{s: server}

	ex, err := c.getExchange(binance)
	require.NoError(t, err)
	assert.Equal(t, ex.Exchange.Name, rtypes.Binance.String())
	exs, err := c.getExchanges()
	require.NoError(t, err)
	assert.Len(t, exs.Exchanges, 3)

	ex3, err := c.getExchange(10)
	assert.NoError(t, err)
	assert.Equal(t, false, ex3.Success)
}

func TestUpdateExchangeStatus(t *testing.T) {
	var (
		supportedExchanges = make(map[rtypes.ExchangeID]v1common.LiveExchange)
	)

	// create map of test exchange
	for _, exchangeID := range []rtypes.ExchangeID{rtypes.Binance} {
		exchange := v1common.TestExchange{}
		supportedExchanges[exchangeID] = exchange
	}

	db, tearDown := testutil.MustNewDevelopmentDB(migrationPath)
	defer func() {
		assert.NoError(t, tearDown())
	}()
	s, err := postgres.NewStorage(db)
	require.NoError(t, err)
	server := NewServer(s, "", supportedExchanges, "", nil, nil)
	c := apiClient{s: server}

	err = s.UpdateExchange(binance, storage.UpdateExchangeOpts{
		TradingFeeMaker: common.FloatPointer(0.2),
		TradingFeeTaker: common.FloatPointer(0.2),
		Disable:         common.BoolPointer(false),
	})
	require.NoError(t, err)

	status, err := c.updateExchangeStatus(binance, exchangeEnabledEntry{
		Disable: false,
	})
	require.NoError(t, err)
	assert.True(t, status.Success)

	ex, err := c.getExchange(binance)
	require.NoError(t, err)
	assert.Equal(t, ex.Exchange.Disable, false)
}
