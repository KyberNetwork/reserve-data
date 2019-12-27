package postgres

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KyberNetwork/reserve-data/common/testutil"
	"github.com/KyberNetwork/reserve-data/reservesetting/common"
)

func TestPriceFactor(t *testing.T) {
	db, tearDown := testutil.MustNewDevelopmentDB()
	defer func() {
		assert.NoError(t, tearDown())
	}()
	s, err := NewStorage(db)
	assert.NoError(t, err)
	var inserts = []struct {
		msg    string
		data   common.PriceFactorAtTime
		assert func(*testing.T, error)
	}{
		{
			msg: "insert 1",
			data: common.PriceFactorAtTime{
				Timestamp: 3,
				Data: []common.AssetPriceFactor{
					{
						AssetID: 1,
						AfpMid:  31,
						Spread:  32,
					},
					{
						AssetID: 2,
						AfpMid:  33,
						Spread:  34,
					},
				},
			},
			assert: func(t *testing.T, e error) {
				assert.NoError(t, e)
			},
		},
		{
			msg: "insert 2",
			data: common.PriceFactorAtTime{
				Timestamp: 4,
				Data: []common.AssetPriceFactor{
					{
						AssetID: 1,
						AfpMid:  41,
						Spread:  42,
					},
					{
						AssetID: 2,
						AfpMid:  43,
						Spread:  44,
					},
				},
			},
			assert: func(t *testing.T, e error) {
				assert.NoError(t, e)
			},
		},
		{
			msg: "insert 3",
			data: common.PriceFactorAtTime{
				Timestamp: 5,
				Data: []common.AssetPriceFactor{
					{
						AssetID: 1,
						AfpMid:  51,
						Spread:  52,
					},
					{
						AssetID: 2,
						AfpMid:  53,
						Spread:  54,
					},
				},
			},
			assert: func(t *testing.T, e error) {
				assert.NoError(t, e)
			},
		},
	}

	for _, i := range inserts {
		t.Logf("run %s", i.msg)
		_, err := s.CreatePriceFactor(i.data)
		i.assert(t, err)
	}

	res, err := s.GetPriceFactors(3, 4)
	assert.NoError(t, err)
	assert.Len(t, res, 2, "expect 2 asset in result")
	assert.Len(t, res[0].Data, 2, "expect 2 result for each asset")
}

func TestStorage_SetRebalanceControl(t *testing.T) {
	db, tearDown := testutil.MustNewDevelopmentDB()
	defer func() {
		assert.NoError(t, tearDown())
	}()
	s, err := NewStorage(db)
	require.NoError(t, err)

	require.NoError(t, s.SetRebalanceStatus(false))
	rebalance, err := s.GetRebalanceStatus()
	require.NoError(t, err)
	require.Equal(t, false, rebalance)

	require.NoError(t, s.SetRebalanceStatus(true))
	rebalance, err = s.GetRebalanceStatus()
	require.NoError(t, err)
	require.Equal(t, true, rebalance)
}

func TestStorage_SetSetRateControl(t *testing.T) {
	db, tearDown := testutil.MustNewDevelopmentDB()
	defer func() {
		assert.NoError(t, tearDown())
	}()
	s, err := NewStorage(db)
	require.NoError(t, err)

	require.NoError(t, s.SetSetRateStatus(false))
	setRateStatus, err := s.GetSetRateStatus()
	require.NoError(t, err)
	require.Equal(t, false, setRateStatus)

	require.NoError(t, s.SetSetRateStatus(true))
	setRateStatus, err = s.GetSetRateStatus()
	require.NoError(t, err)
	require.Equal(t, true, setRateStatus)
}
