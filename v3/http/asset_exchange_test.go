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

func TestHTTPServerAssetExchange(t *testing.T) {

	const (
		createAssetExchange = "/v3/create-asset-exchange"
		updateAssetExchange = "/v3/update-asset-exchange"
		// updateTradingPair = "/v3/update-trading-pair"
	)

	db, tearDown := testutil.MustNewDevelopmentDB()
	defer func() {
		assert.NoError(t, tearDown())
	}()

	s, err := postgres.NewStorage(db)
	require.NoError(t, err)

	assetID, err := createSampleAsset(s)
	require.NoError(t, err)

	sugar := testutil.MustNewDevelopmentSugaredLogger()
	server := NewServer(sugar, s, nil)

	var tests = []testCase{

		{
			msg:      "create asset exchange",
			endpoint: createAssetExchange,
			method:   http.MethodPost,
			assert:   httputil.ExpectSuccess,
			data: &common.CreateCreateAssetExchange{
				AssetExchanges: []common.CreateAssetExchangeEntry{
					{
						AssetID:           assetID,
						ExchangeID:        1,
						Symbol:            "ETH",
						DepositAddress:    eth.HexToAddress("0x001"),
						MinDeposit:        10.0,
						WithdrawFee:       11.0,
						TargetRecommended: 12.0,
						TargetRatio:       13.0,
					},
				}},
		},
		{
			msg:      "confirm pending asset exchange",
			endpoint: createAssetExchange + "/1",
			method:   http.MethodPut,
			data:     nil,
			assert:   httputil.ExpectSuccess,
		},
		{ // test deposit addr = 0
			data: common.CreateUpdateAssetExchange{
				AssetExchanges: []common.UpdateAssetExchangeEntry{
					{
						ID:             1,
						DepositAddress: common.AddressPointer(eth.HexToAddress("0x0000000000000")),
						MinDeposit:     common.FloatPointer(6.0),
						WithdrawFee:    common.FloatPointer(9.0),
					},
				},
			},
			endpoint: updateAssetExchange,
			method:   http.MethodPost,
			msg:      "create invalid update asset exchange",
			assert:   httputil.ExpectFailureWithReason(common.ErrDepositAddressMissing.Error()),
		},
		{ // test create update asset exchange
			data: common.CreateUpdateAssetExchange{
				AssetExchanges: []common.UpdateAssetExchangeEntry{
					{
						ID:             2,
						DepositAddress: common.AddressPointer(eth.HexToAddress("0x0000000000001")),
						MinDeposit:     common.FloatPointer(6.0),
						WithdrawFee:    common.FloatPointer(9.0),
					},
				},
			},
			endpoint: updateAssetExchange,
			method:   http.MethodPost,
			msg:      "create update asset exchange",
			assert:   httputil.ExpectSuccess,
		},
		{ // test confirm update asset exchange
			msg:      "confirm pending update asset exchange",
			endpoint: updateAssetExchange + "/2",
			method:   http.MethodPut,
			data:     nil,
			assert:   httputil.ExpectSuccess,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.msg, func(t *testing.T) { testHTTPRequest(t, tc, server.r) })
	}
}
