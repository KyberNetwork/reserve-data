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

var (
	binance = uint64(common2.Binance)
	huobi   = uint64(common2.Huobi)
)

func initData(t *testing.T, s *Storage) {
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

func TestStorage_ObjectChangeCreate(t *testing.T) {
	db, tearDown := testutil.MustNewDevelopmentDB()
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
			msg: "test missing PWI",
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
							SetRate:      common.ExchangeFeed,
							Rebalance:    false,
							IsQuote:      false,
							PWI:          nil,
						},
					},
				},
			},
			assertFn: func(t *testing.T, u uint64, e error) {
				assert.Equal(t, common.ErrPWIMissing, e)
			},
		},
		{
			msg: "test missing RebalanceQuadratic",
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
			},
		},
		{
			msg: "test missing exchange",
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
								A: 1,
								B: 2,
								C: 3,
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
				assert.Equal(t, common.ErrAssetExchangeMissing, e)
			},
		},
		{
			msg: "test missing target",
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
								A: 1,
								B: 2,
								C: 3,
							},
						},
					},
				},
			},
			assertFn: func(t *testing.T, u uint64, e error) {
				assert.Equal(t, common.ErrAssetTargetMissing, e)
			},
		},
		{
			msg: "test missing address",
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
								A: 1,
								B: 2,
								C: 3,
							},
						},
					},
				},
			},
			assertFn: func(t *testing.T, u uint64, e error) {
				assert.Equal(t, common.ErrAddressMissing, e)
			},
		},
		{
			msg: "test missing deposit address",
			data: common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeCreateAsset,
						Data: common.CreateAssetEntry{
							Symbol:       "DAI",
							Name:         "DAI",
							Address:      common3.HexToAddress("0x12"),
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
									DepositAddress:    common3.HexToAddress("0x00"),
									MinDeposit:        0,
									WithdrawFee:       0,
									TargetRecommended: 0,
									TargetRatio:       0,
								},
							},
							RebalanceQuadratic: &common.RebalanceQuadratic{
								A: 1,
								B: 2,
								C: 3,
							},
						},
					},
				},
			},
			assertFn: func(t *testing.T, u uint64, e error) {
				assert.Equal(t, common.ErrDepositAddressMissing, e)
			},
		},
		{
			msg: "test invalid trading pair",
			data: common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeCreateAsset,
						Data: common.CreateAssetEntry{
							Symbol:       "DAI",
							Name:         "DAI",
							Address:      common3.HexToAddress("0x12"),
							Decimals:     18,
							Transferable: true,
							SetRate:      common.SetRateNotSet,
							Rebalance:    true,
							IsQuote:      false,
							PWI:          nil,
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
								A: 1,
								B: 2,
								C: 3,
							},
						},
					},
				},
			},
			assertFn: func(t *testing.T, u uint64, e error) {
				assert.Equal(t, common.ErrQuoteAssetInvalid, e)
			},
		},
	}
	for _, tc := range tests {
		t.Logf("running test case for: %s", tc.msg)
		id, err := s.CreateSettingChange(tc.data)
		assert.NoError(t, err)
		err = s.ConfirmSettingChange(id, true)
		tc.assertFn(t, id, err)
	}
}
