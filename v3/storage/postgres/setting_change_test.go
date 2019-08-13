package postgres

import (
	"testing"

	common3 "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	common2 "github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/testutil"
	"github.com/KyberNetwork/reserve-data/v3/common"
)

func TestStorage_ObjectChangeCreate(t *testing.T) {
	db, tearDown := testutil.MustNewDevelopmentDB()
	defer func() {
		assert.NoError(t, tearDown())
	}()

	binance := uint64(common2.Binance)
	huobi := uint64(common2.Huobi)
	s, err := NewStorage(db)
	assert.NoError(t, err)
	id, err := s.CreateSettingChange(common.SettingChange{ChangeList: []common.SettingChangeEntry{
		{
			Type: common.ChangeTypeUpdateExchange,
			Data: common.UpdateExchangeEntry{
				ExchangeID:      0,
				TradingFeeMaker: common.FloatPointer(1.5),
				TradingFeeTaker: common.FloatPointer(1.5),
				Disable:         common.BoolPointer(false),
			},
		},
		{
			Type: common.ChangeTypeCreateAssetExchange,
			Data: common.CreateAssetExchangeEntry{
				AssetID:           1,
				ExchangeID:        binance,
				Symbol:            "ETH",
				DepositAddress:    common3.HexToAddress("0xee"),
				MinDeposit:        100.0,
				WithdrawFee:       12.0,
				TargetRecommended: 13.0,
				TargetRatio:       14.0,
			},
		},
		{
			Type: common.ChangeTypeCreateAssetExchange,
			Data: common.CreateAssetExchangeEntry{
				AssetID:           1,
				ExchangeID:        huobi,
				Symbol:            "ETH",
				DepositAddress:    common3.HexToAddress("0xee"),
				MinDeposit:        100.0,
				WithdrawFee:       12.0,
				TargetRecommended: 13.0,
				TargetRatio:       14.0,
			},
		},
		{
			Type: common.ChangeTypeCreateAsset,
			Data: common.CreateAssetEntry{
				Symbol:       "BTC",
				Name:         "BTC Super",
				Address:      common3.Address{},
				Decimals:     18,
				Transferable: false,
				SetRate:      common.ExchangeFeed,
				Rebalance:    false,
				IsQuote:      true,
				PWI: &common.AssetPWI{
					Ask: common.PWIEquation{
						A:                   1,
						B:                   2,
						C:                   3,
						MinMinSpread:        4,
						PriceMultiplyFactor: 5,
					},
					Bid: common.PWIEquation{
						A:                   2,
						B:                   3,
						C:                   4,
						MinMinSpread:        5,
						PriceMultiplyFactor: 6,
					},
				},
				RebalanceQuadratic: &common.RebalanceQuadratic{
					A: 3,
					B: 4,
					C: 5,
				},
				Exchanges: []common.AssetExchange{
					{
						ExchangeID:        binance,
						Symbol:            "BTC",
						DepositAddress:    common3.Address{},
						MinDeposit:        3,
						WithdrawFee:       4,
						TargetRecommended: 5,
						TargetRatio:       6,
						TradingPairs: []common.TradingPair{
							{
								Quote: 1,
							},
						},
					},
					{
						ExchangeID:        huobi,
						Symbol:            "BTC",
						DepositAddress:    common3.Address{},
						MinDeposit:        3,
						WithdrawFee:       4,
						TargetRecommended: 5,
						TargetRatio:       6,
						TradingPairs: []common.TradingPair{
							{
								Base: 1,
							},
						},
					},
				},
				Target: nil,
			},
		},
		{
			Type: common.ChangeTypeCreateAsset,
			Data: common.CreateAssetEntry{
				Symbol:       "KNC",
				Name:         "KNC Super",
				Address:      common3.HexToAddress("0x11223344"),
				Decimals:     18,
				Transferable: false,
				SetRate:      common.ExchangeFeed,
				Rebalance:    false,
				IsQuote:      true,
				PWI: &common.AssetPWI{
					Ask: common.PWIEquation{
						A:                   1,
						B:                   2,
						C:                   3,
						MinMinSpread:        4,
						PriceMultiplyFactor: 5,
					},
					Bid: common.PWIEquation{
						A:                   2,
						B:                   3,
						C:                   4,
						MinMinSpread:        5,
						PriceMultiplyFactor: 6,
					},
				},
				RebalanceQuadratic: &common.RebalanceQuadratic{
					A: 3,
					B: 4,
					C: 5,
				},
				Exchanges: []common.AssetExchange{
					{
						ExchangeID:        binance,
						Symbol:            "KNC",
						DepositAddress:    common3.HexToAddress("0x223344"),
						MinDeposit:        3,
						WithdrawFee:       4,
						TargetRecommended: 5,
						TargetRatio:       6,
						TradingPairs: []common.TradingPair{
							{
								Quote: 1,
							},
							{
								Quote: 2,
							},
						},
					},
				},
				Target: nil,
			},
		},
	}})
	require.NoError(t, err)
	err = s.ConfirmSettingChange(id, true)
	require.NoError(t, err)
}
