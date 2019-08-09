package postgres

import (
	"testing"

	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KyberNetwork/reserve-data/common/testutil"
	"github.com/KyberNetwork/reserve-data/v3/common"
)

// setUp adds data for testing returns baseID, quoteID2, tradingPairID
func setUp(t *testing.T, storage *Storage) (uint64, uint64, uint64) {
	assetID, err := storage.CreateAsset("ABC", "ABC Advanced Token",
		ethereum.HexToAddress("0xB8c77482e45F1F44dE1745F52C74426C631bDD52"),
		12, true, common.BTCFeed, true, true,
		testPWI, testRb, testAssetExchanges, testAssetTarget)
	require.NoError(t, err)
	asset, err := storage.GetAsset(assetID)
	require.NoError(t, err)
	require.Equal(t, "ABC", asset.Symbol)
	require.NotEqual(t, len(asset.Exchanges), 0)

	assetID2, err := storage.CreateAsset("DEF", "DEF Advanced Token",
		ethereum.HexToAddress("0xB8c77482e45F1F44dE1745F52C74426C631bDD51"),
		12, true, common.BTCFeed, true, true,
		testPWI, testRb, testAssetExchanges, testAssetTarget)
	require.NoError(t, err)
	asset2, err := storage.GetAsset(assetID2)
	require.NoError(t, err)
	require.Equal(t, "DEF", asset2.Symbol)
	//create trading pair for test
	assetExchangeID := asset.Exchanges[0].ID
	tx, err := storage.db.Beginx()
	require.NoError(t, err)
	defer rollbackUnlessCommitted(tx)

	tradingPairID, err := storage.createTradingPair(tx, assetExchangeID, assetID, assetID2, 10, 10,
		0.1, 0.1, 0.1, 0.1, 0.1)

	require.NoError(t, tx.Commit())
	require.NoError(t, err)
	tradingPair, err := storage.GetTradingPair(tradingPairID)
	require.NoError(t, err)
	require.Equal(t, assetID, tradingPair.Base)
	return assetID, assetID2, tradingPairID
}

func TestStorage_CreateTradingBy(t *testing.T) {
	db, tearDown := testutil.MustNewDevelopmentDB()
	defer func() {
		assert.NoError(t, tearDown())
	}()

	s, err := NewStorage(db)
	require.NoError(t, err)

	baseID, quoteID, tradingPairID := setUp(t, s)

	tests := []struct {
		assetID        uint64
		tradingPairID  uint64
		expectedResult bool
		assetFn        func(tradingByID uint64, returnErr error)
	}{
		{
			assetID:        baseID,
			tradingPairID:  1234,
			expectedResult: false,
		},
		{
			assetID:        baseID,
			tradingPairID:  tradingPairID,
			expectedResult: true,
			assetFn: func(tradingByID uint64, _ error) {
				tradingBy, err := s.GetTradingBy(tradingByID)
				require.NoError(t, err)
				require.Equal(t, baseID, tradingBy.AssetID)
				require.Equal(t, tradingPairID, tradingBy.TradingPairID)
			},
		},
		{
			assetID:        quoteID,
			tradingPairID:  tradingPairID,
			expectedResult: true,
			assetFn: func(tradingByID uint64, _ error) {
				tradingBy, err := s.GetTradingBy(tradingByID)
				require.NoError(t, err)
				require.Equal(t, quoteID, tradingBy.AssetID)
				require.Equal(t, tradingPairID, tradingBy.TradingPairID)
			},
		},
		{
			assetID:        quoteID,
			tradingPairID:  tradingPairID,
			expectedResult: false,
			assetFn: func(_ uint64, returnErr error) {
				require.Equal(t, common.ErrTradingByAlreadyExists, returnErr)
			},
		},
	}

	for _, test := range tests {
		tradingByID, err := s.CreateTradingBy(test.assetID, test.tradingPairID)
		if test.expectedResult {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
		}
		if test.assetFn != nil {
			test.assetFn(tradingByID, err)
		}
	}
}

func TestStorage_ConfirmCreateTradingBy(t *testing.T) {
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

	_, err = s.GetPendingObject(id, common.PendingTypeCreateTradingBy)
	require.NoError(t, err)

	err = s.ConfirmCreateTradingBy(id)
	require.NoError(t, err)
}
