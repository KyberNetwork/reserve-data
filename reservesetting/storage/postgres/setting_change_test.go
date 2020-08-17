package postgres

import (
	"fmt"
	"testing"

	common3 "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	common2 "github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/testutil"
	"github.com/KyberNetwork/reserve-data/reservesetting/common"
)

var (
	binance = uint64(common2.Binance)
	huobi   = uint64(common2.Huobi)
)

func initData(t *testing.T, s *Storage) {
	id, err := s.CreateSettingChange(common.ChangeCatalogMain, common.SettingChange{ChangeList: []common.SettingChangeEntry{
		{
			Type: common.ChangeTypeUpdateExchange,
			Data: common.UpdateExchangeEntry{
				ExchangeID:      binance,
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
					SizeA:  3,
					SizeB:  4,
					SizeC:  5,
					PriceA: 9,
					PriceB: 15,
					PriceC: 21,
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
				StableParam: &common.StableParam{
					PriceUpdateThreshold: 10,
					AskSpread:            11,
					BidSpread:            12,
					SingleFeedMaxSpread:  13,
				},
				NormalUpdatePerPeriod: 0.5,
				MaxImbalanceRatio:     0.6,
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
				IsQuote:      false,
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
					SizeA:  3,
					SizeB:  4,
					SizeC:  5,
					PriceA: 9,
					PriceB: 15,
					PriceC: 21,
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
				Target: &common.AssetTarget{
					Total:              1,
					Reserve:            2,
					RebalanceThreshold: 3,
					TransferThreshold:  4,
				},
				NormalUpdatePerPeriod: 0.5,
				MaxImbalanceRatio:     0.6,
			},
		},
	}})
	require.NoError(t, err)
	err = s.ConfirmSettingChange(id, true)
	require.NoError(t, err)
}

func TestStorage_SettingChangeCreate(t *testing.T) {
	db, tearDown := testutil.MustNewDevelopmentDB(migrationPath)
	defer func() {
		assert.NoError(t, tearDown())
	}()

	s, err := NewStorage(db)
	assert.NoError(t, err)
	initData(t, s)
	var tests = []struct {
		msg      string
		data     common.SettingChange
		assertFn func(*testing.T, uint64, error)
	}{
		{
			msg: fmt.Sprintf("test missing PWI when SetRate != %s", common.SetRateNotSet),
			data: common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeCreateAsset,
						Data: common.CreateAssetEntry{
							Symbol:                "DAI",
							Name:                  "DAI",
							Address:               common3.HexToAddress("0x1199"),
							Decimals:              18,
							Transferable:          true,
							SetRate:               common.ExchangeFeed,
							Rebalance:             false,
							IsQuote:               false,
							PWI:                   nil,
							NormalUpdatePerPeriod: 0.5,
							MaxImbalanceRatio:     0.6,
						},
					},
				},
			},
			assertFn: func(t *testing.T, u uint64, e error) {
				assert.Equal(t, common.ErrPWIMissing, e)
				assert.NoError(t, s.RejectSettingChange(u))
			},
		},
		{
			msg: "test missing RebalanceQuadratic when Rebalance set",
			data: common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeCreateAsset,
						Data: common.CreateAssetEntry{
							Symbol:                "DAI",
							Name:                  "DAI",
							Address:               common3.HexToAddress("0x1199"),
							Decimals:              18,
							Transferable:          true,
							SetRate:               common.SetRateNotSet,
							Rebalance:             true,
							IsQuote:               false,
							PWI:                   nil,
							NormalUpdatePerPeriod: 0.5,
							MaxImbalanceRatio:     0.6,
							Exchanges: []common.AssetExchange{
								{
									ID:                0,
									AssetID:           0,
									ExchangeID:        binance,
									Symbol:            "DAI",
									DepositAddress:    common3.HexToAddress("0x1234"),
									MinDeposit:        0,
									WithdrawFee:       0,
									TargetRecommended: 0,
									TargetRatio:       0,
								},
							},
							Target: &common.AssetTarget{
								Total:              0,
								Reserve:            0,
								RebalanceThreshold: 0,
								TransferThreshold:  0,
							},
						},
					},
				},
			},
			assertFn: func(t *testing.T, u uint64, e error) {
				assert.Equal(t, common.ErrRebalanceQuadraticMissing, e)
				assert.NoError(t, s.RejectSettingChange(u))
			},
		},
		{
			msg: "test Transferable but no exchange define",
			data: common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeCreateAsset,
						Data: common.CreateAssetEntry{
							Symbol:       "DAI",
							Name:         "DAI",
							Address:      common3.HexToAddress("0x1199"),
							Decimals:     18,
							Transferable: true,
							SetRate:      common.SetRateNotSet,
							Rebalance:    true,
							IsQuote:      false,
							PWI:          nil,
							RebalanceQuadratic: &common.RebalanceQuadratic{
								SizeA:  1,
								SizeB:  2,
								SizeC:  3,
								PriceA: 9,
								PriceB: 15,
								PriceC: 21,
							},
							Target: &common.AssetTarget{
								Total:              0,
								Reserve:            0,
								RebalanceThreshold: 0,
								TransferThreshold:  0,
							},
							NormalUpdatePerPeriod: 0.5,
							MaxImbalanceRatio:     0.6,
						},
					},
				},
			},
			assertFn: func(t *testing.T, u uint64, e error) {
				assert.Equal(t, common.ErrAssetExchangeMissing, e)
				assert.NoError(t, s.RejectSettingChange(u))
			},
		},
		{
			msg: "test missing target when Rebalance set",
			data: common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeCreateAsset,
						Data: common.CreateAssetEntry{
							Symbol:       "DAI",
							Name:         "DAI",
							Address:      common3.HexToAddress("0x1199"),
							Decimals:     18,
							Transferable: true,
							SetRate:      common.SetRateNotSet,
							Rebalance:    true,
							IsQuote:      false,
							PWI:          nil,
							Exchanges: []common.AssetExchange{
								{
									ID:                0,
									AssetID:           0,
									ExchangeID:        binance,
									Symbol:            "DAI",
									DepositAddress:    common3.HexToAddress("0x1234"),
									MinDeposit:        0,
									WithdrawFee:       0,
									TargetRecommended: 0,
									TargetRatio:       0,
								},
							},
							RebalanceQuadratic: &common.RebalanceQuadratic{
								SizeA:  1,
								SizeB:  2,
								SizeC:  3,
								PriceA: 9,
								PriceB: 15,
								PriceC: 21,
							},
							NormalUpdatePerPeriod: 0.5,
							MaxImbalanceRatio:     0.6,
						},
					},
				},
			},
			assertFn: func(t *testing.T, u uint64, e error) {
				assert.Equal(t, common.ErrAssetTargetMissing, e)
				assert.NoError(t, s.RejectSettingChange(u))
			},
		},
		{
			msg: "test missing address when Transferable set",
			data: common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeCreateAsset,
						Data: common.CreateAssetEntry{
							Symbol: "DAI",
							Name:   "DAI",
							//Address:      common3.HexToAddress("0x00"),
							Decimals:     18,
							Transferable: true,
							SetRate:      common.SetRateNotSet,
							Rebalance:    true,
							IsQuote:      false,
							PWI:          nil,
							Exchanges: []common.AssetExchange{
								{
									ExchangeID:        binance,
									Symbol:            "DAI",
									DepositAddress:    common3.HexToAddress("0x1234"),
									MinDeposit:        0,
									WithdrawFee:       0,
									TargetRecommended: 0,
									TargetRatio:       0,
								},
							},
							RebalanceQuadratic: &common.RebalanceQuadratic{
								SizeA:  1,
								SizeB:  2,
								SizeC:  3,
								PriceA: 9,
								PriceB: 15,
								PriceC: 21,
							},
						},
					},
				},
			},
			assertFn: func(t *testing.T, u uint64, e error) {
				assert.Equal(t, common.ErrAddressMissing, e)
				assert.NoError(t, s.RejectSettingChange(u))
			},
		},
		{
			msg: "test missing deposit address when Transferable set",
			data: common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeCreateAsset,
						Data: common.CreateAssetEntry{
							Symbol:                "DAI",
							Name:                  "DAI",
							Address:               common3.HexToAddress("0x12"),
							Decimals:              18,
							Transferable:          true,
							SetRate:               common.SetRateNotSet,
							Rebalance:             true,
							IsQuote:               false,
							PWI:                   nil,
							NormalUpdatePerPeriod: 0.5,
							MaxImbalanceRatio:     0.6,
							Exchanges: []common.AssetExchange{
								{
									ExchangeID:        binance,
									Symbol:            "DAI",
									DepositAddress:    common3.HexToAddress("0x00"),
									MinDeposit:        0,
									WithdrawFee:       0,
									TargetRecommended: 0,
									TargetRatio:       0,
								},
							},
							RebalanceQuadratic: &common.RebalanceQuadratic{
								SizeA:  1,
								SizeB:  2,
								SizeC:  3,
								PriceA: 9,
								PriceB: 15,
								PriceC: 21,
							},
						},
					},
				},
			},
			assertFn: func(t *testing.T, u uint64, e error) {
				assert.Equal(t, common.ErrDepositAddressMissing, e)
				assert.NoError(t, s.RejectSettingChange(u))
			},
		},
		{
			msg: "test invalid trading pair, quote asset does not exist or not is a quote asset",
			data: common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeCreateAsset,
						Data: common.CreateAssetEntry{
							Symbol:                "DAI",
							Name:                  "DAI",
							Address:               common3.HexToAddress("0x12"),
							Decimals:              18,
							Transferable:          true,
							SetRate:               common.SetRateNotSet,
							Rebalance:             true,
							IsQuote:               false,
							PWI:                   nil,
							NormalUpdatePerPeriod: 0.5,
							MaxImbalanceRatio:     0.6,
							Target: &common.AssetTarget{
								Total:              0,
								Reserve:            0,
								RebalanceThreshold: 0,
								TransferThreshold:  0,
							},
							Exchanges: []common.AssetExchange{
								{
									ExchangeID:        binance,
									Symbol:            "DAI",
									DepositAddress:    common3.HexToAddress("0x3344"),
									MinDeposit:        0,
									WithdrawFee:       0,
									TargetRecommended: 0,
									TargetRatio:       0,
									TradingPairs: []common.TradingPair{
										{
											Quote: 3,
										},
									},
								},
							},
							RebalanceQuadratic: &common.RebalanceQuadratic{
								SizeA:  1,
								SizeB:  2,
								SizeC:  3,
								PriceA: 9,
								PriceB: 15,
								PriceC: 21,
							},
						},
					},
				},
			},
			assertFn: func(t *testing.T, u uint64, e error) {
				assert.Equal(t, common.ErrQuoteAssetInvalid, e)
				assert.NoError(t, s.RejectSettingChange(u))
			},
		},
		{
			msg: "test update asset, transferable but asset exchange deposit address not set",
			data: common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeUpdateAsset,
						Data: common.UpdateAssetEntry{
							AssetID:            2,
							Symbol:             nil,
							Name:               nil,
							Address:            nil,
							Decimals:           nil,
							Transferable:       common.BoolPointer(true),
							SetRate:            nil,
							Rebalance:          nil,
							IsQuote:            nil,
							PWI:                nil,
							RebalanceQuadratic: nil,
							Target:             nil,
						},
					},
				},
			},

			assertFn: func(t *testing.T, u uint64, e error) {
				assert.Equal(t, common.ErrDepositAddressMissing, e)
				assert.NoError(t, s.RejectSettingChange(u))
			},
		},
		{
			msg: "test update asset, update normal_update_per_period",
			data: common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeUpdateAsset,
						Data: common.UpdateAssetEntry{
							AssetID:               2,
							NormalUpdatePerPeriod: common.FloatPointer(0.123),
						},
					},
				},
			},

			assertFn: func(t *testing.T, u uint64, e error) {
				assert.NoError(t, e)
				asset, err := s.GetAsset(2)
				assert.NoError(t, err)
				assert.Equal(t, 0.123, asset.NormalUpdatePerPeriod)
			},
		},
		{
			msg: "test update asset, update normal_update_per_period < 0",
			data: common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeUpdateAsset,
						Data: common.UpdateAssetEntry{
							AssetID:               2,
							NormalUpdatePerPeriod: common.FloatPointer(-123),
						},
					},
				},
			},

			assertFn: func(t *testing.T, u uint64, e error) {
				assert.Error(t, e)
				assert.NoError(t, s.RejectSettingChange(u))
			},
		},
		{
			msg: "test update asset, update max_imbalance_ratio",
			data: common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeUpdateAsset,
						Data: common.UpdateAssetEntry{
							AssetID:           2,
							MaxImbalanceRatio: common.FloatPointer(0.456),
						},
					},
				},
			},

			assertFn: func(t *testing.T, u uint64, e error) {
				assert.NoError(t, e)
				asset, err := s.GetAsset(2)
				assert.NoError(t, err)
				assert.Equal(t, 0.456, asset.MaxImbalanceRatio)
			},
		},
		{
			msg: "test update asset, update max_imbalance_ratio < 0",
			data: common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeUpdateAsset,
						Data: common.UpdateAssetEntry{
							AssetID:           2,
							MaxImbalanceRatio: common.FloatPointer(-456),
						},
					},
				},
			},

			assertFn: func(t *testing.T, u uint64, e error) {
				assert.Error(t, e)
				assert.NoError(t, s.RejectSettingChange(u))
			},
		},
	}
	for _, tc := range tests {
		t.Logf("running test case for: %s", tc.msg)
		id, err := s.CreateSettingChange(common.ChangeCatalogMain, tc.data)
		assert.NoError(t, err)
		err = s.ConfirmSettingChange(id, true)
		tc.assertFn(t, id, err)
	}
}

func TestStorage_GetDepositAddresses(t *testing.T) {
	db, tearDown := testutil.MustNewDevelopmentDB(migrationPath)
	defer func() {
		assert.NoError(t, tearDown())
	}()
	s, err := NewStorage(db)
	assert.NoError(t, err)
	initData(t, s)

	depositAddr, err := s.GetDepositAddresses(binance)
	require.NoError(t, err)
	require.Equal(t, 3, len(depositAddr))
	for symbol, addr := range depositAddr {
		t.Logf("symbol=%v deposit address=%v", symbol, addr.Hex())
	}
}
func TestStorage_DeleteTradingPair(t *testing.T) {
	db, tearDown := testutil.MustNewDevelopmentDB(migrationPath)
	defer func() {
		assert.NoError(t, tearDown())
	}()
	s, err := NewStorage(db)
	assert.NoError(t, err)
	initData(t, s)

	_, err = s.GetTradingPair(1, false)
	require.NoError(t, err)
	c, err := s.CreateSettingChange(common.ChangeCatalogMain, common.SettingChange{
		ChangeList: []common.SettingChangeEntry{
			{
				Type: common.ChangeTypeDeleteTradingPair,
				Data: common.DeleteTradingPairEntry{TradingPairID: 1},
			},
		},
		Message: "delete trading pair",
	})
	require.NoError(t, err)
	err = s.ConfirmSettingChange(c, true)
	require.NoError(t, err)
	_, err = s.GetTradingPair(1, false)
	require.Error(t, common.ErrNotFound)
	_, err = s.GetTradingPair(1, true)
	require.NoError(t, err)
}
