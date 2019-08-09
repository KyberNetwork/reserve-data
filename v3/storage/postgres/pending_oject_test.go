package postgres

import (
	"testing"

	eth "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KyberNetwork/reserve-data/common/testutil"
	"github.com/KyberNetwork/reserve-data/v3/common"
)

func TestStorage_CreatePendingObject(t *testing.T) {
	db, tearDown := testutil.MustNewDevelopmentDB()
	defer func() {
		assert.NoError(t, tearDown())
	}()

	s, err := NewStorage(db)
	require.NoError(t, err)

	baseID, _, tradingPairID := setUp(t, s)

	id, err := s.CreatePendingObject(common.CreateCreateTradingBy{
		TradingBys: []common.CreateTradingByEntry{
			{
				AssetID:       baseID,
				TradingPairID: tradingPairID,
			},
		},
	}, common.PendingTypeCreateTradingBy)
	require.NoError(t, err)

	createAssetID, err := s.CreatePendingObject(common.CreateUpdateAsset{
		Assets: []common.UpdateAssetEntry{
			{
				AssetID:      3,
				Symbol:       common.StringPointer("XYZ"),
				Name:         common.StringPointer("ZXC"),
				Address:      common.AddressPointer(eth.HexToAddress("0x02")),
				Decimals:     common.Uint64Pointer(19),
				Transferable: common.BoolPointer(false),
				SetRate:      common.SetRatePointer(common.BTCFeed),
				Rebalance:    common.BoolPointer(true),
				IsQuote:      common.BoolPointer(false),
				Target: &common.AssetTarget{
					Total:              5.0,
					Reserve:            5.0,
					RebalanceThreshold: 5.0,
					TransferThreshold:  5.0,
				},
				RebalanceQuadratic: &common.RebalanceQuadratic{
					A: 5.0,
					B: 5.0,
					C: 5.0,
				},
				PWI: &common.AssetPWI{
					Ask: common.PWIEquation{
						A:                   5.0,
						B:                   5.0,
						C:                   5.0,
						MinMinSpread:        5.0,
						PriceMultiplyFactor: 5.0,
					},
					Bid: common.PWIEquation{
						A:                   5.0,
						B:                   5.0,
						C:                   5.0,
						MinMinSpread:        5.0,
						PriceMultiplyFactor: 5.0,
					},
				},
			},
		},
	}, common.PendingTypeUpdateAsset)
	require.NoError(t, err)

	_, err = s.GetPendingObject(id, common.PendingTypeCreateTradingBy)
	require.NoError(t, err)

	_, err = s.GetPendingObject(createAssetID, common.PendingTypeUpdateAsset)
	require.NoError(t, err)
	// Delete create trading by
	err = s.RejectPendingObject(id, common.PendingTypeCreateTradingBy)
	require.NoError(t, err)
	// Get a deleted pending obj should return err
	_, err = s.GetPendingObject(id, common.PendingTypeCreateTradingBy)
	require.Error(t, err)
	// Get update asset should not be deleted
	_, err = s.GetPendingObject(createAssetID, common.PendingTypeUpdateAsset)
	require.NoError(t, err)
	// Create pending object with the same type
	_, err = s.CreatePendingObject(common.CreateUpdateAsset{
		Assets: []common.UpdateAssetEntry{},
	}, common.PendingTypeUpdateAsset)
	require.NoError(t, err)
	// old pending obj should be deleted
	_, err = s.GetPendingObject(createAssetID, common.PendingTypeUpdateAsset)
	require.Error(t, err)
}
