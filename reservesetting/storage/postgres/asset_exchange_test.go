package postgres

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/KyberNetwork/reserve-data/common/testutil"
	rtypes "github.com/KyberNetwork/reserve-data/lib/rtypes"
)

func TestUpdateAssetExchange(t *testing.T) {
	db, tearDown := testutil.MustNewDevelopmentDB(migrationPath)
	defer func() {
		assert.NoError(t, tearDown())
	}()
	s, err := NewStorage(db)
	assert.NoError(t, err)
	initData(t, s)
	ae, err := s.GetAssetExchangeBySymbol(rtypes.Binance, "ETH")
	assert.NoError(t, err)
	ae.WithdrawFee = 0.004
	err = s.UpdateAssetExchangeWithdrawFee(0.004, ae.ID)
	assert.NoError(t, err)
	newAE, err := s.GetAssetExchangeBySymbol(rtypes.Binance, "ETH")
	assert.NoError(t, err)
	assert.Equal(t, ae, newAE)
}
