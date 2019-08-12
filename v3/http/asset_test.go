package http

import (
	"net/http"
	"testing"

	eth "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KyberNetwork/reserve-data/common/testutil"
	"github.com/KyberNetwork/reserve-data/http/httputil"
	"github.com/KyberNetwork/reserve-data/v3/common"
	"github.com/KyberNetwork/reserve-data/v3/storage/postgres"
)

var (
	createPEA = common.CreateCreateAsset{AssetInputs: []common.CreateAssetEntry{
		{
			Address:   eth.HexToAddress("0x01"),
			Symbol:    "ABC",
			Name:      "ABC test",
			Decimals:  18,
			SetRate:   common.ExchangeFeed,
			Rebalance: true,
			IsQuote:   true,
			PWI: &common.AssetPWI{
				Ask: common.PWIEquation{
					A:                   1.0,
					B:                   1.0,
					C:                   1.0,
					MinMinSpread:        2.0,
					PriceMultiplyFactor: 2.0,
				},
				Bid: common.PWIEquation{
					A:                   1.0,
					B:                   1.0,
					C:                   1.0,
					MinMinSpread:        2.0,
					PriceMultiplyFactor: 2.0,
				},
			},
			RebalanceQuadratic: &common.RebalanceQuadratic{
				A: 100.0,
				B: 200.0,
				C: 150.0,
			},
			Target: &common.AssetTarget{
				Total:              1000.0,
				Reserve:            1000.0,
				RebalanceThreshold: 1000.0,
				TransferThreshold:  1000.0,
			},
			Exchanges: []common.AssetExchange{
				{
					ExchangeID: 0, // pre-define exchange
					TradingPairs: []common.TradingPair{
						{
							Base:  0,
							Quote: 1,
						},
					},
				},
			},
		},
	}}
)

func getCreatePEAWithQuoteFalse() common.CreateCreateAsset {
	var createPEAWithQuoteFalse common.CreateCreateAsset
	createPEAWithQuoteFalse.AssetInputs = append(createPEAWithQuoteFalse.AssetInputs, createPEA.AssetInputs[0])
	createPEAWithQuoteFalse.AssetInputs[0].IsQuote = false
	createPEAWithQuoteFalse.AssetInputs[0].Address = eth.HexToAddress("0x000005")
	createPEAWithQuoteFalse.AssetInputs[0].Symbol = "QUOTE FALSE"
	createPEAWithQuoteFalse.AssetInputs[0].Exchanges = []common.AssetExchange{
		{
			ExchangeID:   0, // pre-define exchange
			TradingPairs: []common.TradingPair{},
		},
	}
	return createPEAWithQuoteFalse
}

func TestReCreateCreateAsset(t *testing.T) {

	db, tearDown := testutil.MustNewDevelopmentDB()
	defer func() {
		assert.NoError(t, tearDown())
	}()
	s, err := postgres.NewStorage(db)
	require.NoError(t, err)
	_, err = s.CreatePendingObject(createPEA, common.PendingTypeCreateAsset)
	require.NoError(t, err)
	id2, err := s.CreatePendingObject(createPEA, common.PendingTypeCreateAsset)
	require.NoError(t, err)
	pending, err := s.GetPendingObjects(common.PendingTypeCreateAsset)
	require.NoError(t, err)
	if len(pending) != 1 || pending[0].ID != id2 {
		t.Fatal("expect 1 element with latest create one")
	}
}

func TestHTTPServerAsset(t *testing.T) {

	const (
		assetBase       = "/v3/asset"
		createAssetBase = "/v3/create-asset"
	)

	db, tearDown := testutil.MustNewDevelopmentDB()
	defer func() {
		assert.NoError(t, tearDown())
	}()

	s, err := postgres.NewStorage(db)
	require.NoError(t, err)
	// asset = 1 for ETH is pre-insert in DB.
	_, err = s.CreateAssetExchange(0, 1, "ETH", eth.HexToAddress("0x00"), 10,
		0.2, 5.0, 0.3)
	require.NoError(t, err)
	sugar := testutil.MustNewDevelopmentSugaredLogger()
	server := NewServer(sugar, s, nil)

	var createPEAWithQuoteFalse = getCreatePEAWithQuoteFalse()
	createPEAWithQuoteFalse.AssetInputs[0].Exchanges[0].TradingPairs = []common.TradingPair{
		{
			Quote: 1, Base: 0,
		},
	}
	var createPEAWithQuoteFalse2 = getCreatePEAWithQuoteFalse()
	createPEAWithQuoteFalse2.AssetInputs[0].Exchanges[0].TradingPairs = []common.TradingPair{
		{
			Quote: 0, Base: 1,
		},
	}

	var tests = []testCase{
		{
			msg:      "asset not found",
			endpoint: assetBase + "/-1",
			method:   http.MethodGet,
			assert:   httputil.ExpectFailure,
		},
		{
			msg:      "list all asset",
			endpoint: assetBase,
			method:   http.MethodGet,
			assert:   httputil.ExpectSuccess,
		},
		{
			msg:      "create pending asset",
			endpoint: createAssetBase,
			method:   http.MethodPost,
			assert:   httputil.ExpectSuccess,
			data:     createPEA,
		},
		{
			msg:      "list all pending asset",
			endpoint: createAssetBase,
			method:   http.MethodGet,
			assert:   httputil.ExpectSuccess,
		},
		{
			msg:      "confirm invalid pending asset",
			endpoint: createAssetBase + "/-1",
			method:   http.MethodPut,
			data:     nil,
			assert:   httputil.ExpectFailure,
		},
		{
			msg:      "confirm pending asset",
			endpoint: createAssetBase + "/1",
			method:   http.MethodPut,
			data:     nil,
			assert:   httputil.ExpectSuccess,
		},
		{
			msg:      "receive asset",
			endpoint: assetBase + "/1",
			method:   http.MethodGet,
			assert:   httputil.ExpectSuccess,
		},
		{
			msg:      "create pending asset",
			endpoint: createAssetBase,
			method:   http.MethodPost,
			assert:   httputil.ExpectFailure,
			data:     createPEAWithQuoteFalse2,
		},
		{
			msg:      "create pending asset",
			endpoint: createAssetBase,
			method:   http.MethodPost,
			assert:   httputil.ExpectSuccess,
			data:     createPEAWithQuoteFalse,
		},
		{
			msg:      "create pending asset",
			endpoint: createAssetBase + "/2",
			method:   http.MethodDelete,
			assert:   httputil.ExpectSuccess,
		},
		{
			msg:      "create pending asset",
			endpoint: createAssetBase,
			method:   http.MethodPost,
			assert:   httputil.ExpectSuccess,
			data:     createPEA,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.msg, func(t *testing.T) { testHTTPRequest(t, tc, server.r) })
	}
}
