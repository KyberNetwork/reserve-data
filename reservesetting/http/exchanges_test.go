package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1common "github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/testutil"
	"github.com/KyberNetwork/reserve-data/reservesetting/storage/postgres"
)

func TestExchanges(t *testing.T) {
	var (
		supportedExchanges = make(map[v1common.ExchangeID]v1common.LiveExchange)
	)

	// create map of test exchange
	for _, exchangeID := range []v1common.ExchangeID{v1common.Binance, v1common.Huobi, v1common.StableExchange} {
		exchange := v1common.TestExchange{}
		supportedExchanges[exchangeID] = exchange
	}

	db, tearDown := testutil.MustNewDevelopmentDB()
	defer func() {
		assert.NoError(t, tearDown())
	}()
	s, err := postgres.NewStorage(db, "")
	require.NoError(t, err)
	server := NewServer(s, "", supportedExchanges, nil, "")
	c := apiClient{s: server}

	ex, err := c.getExchange(binance)
	require.NoError(t, err)
	assert.Equal(t, ex.Exchange.Name, v1common.Binance.String())
	exs, err := c.getExchanges()
	require.NoError(t, err)
	assert.Len(t, exs.Exchanges, 3)

	ex3, err := c.getExchange(10)
	assert.NoError(t, err)
	assert.Equal(t, false, ex3.Success)
}
