package http

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	eth "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KyberNetwork/reserve-data/common/testutil"
	"github.com/KyberNetwork/reserve-data/http/httputil"
	"github.com/KyberNetwork/reserve-data/v3/common"
	"github.com/KyberNetwork/reserve-data/v3/storage/postgres"
)

// TODO write more test cases
func TestServer_SettingChangeBasic(t *testing.T) {
	t.Skip()
	const (
		settingChangePath = "/v3/setting-change"
	)

	db, tearDown := testutil.MustNewDevelopmentDB()
	defer func() {
		assert.NoError(t, tearDown())
	}()

	s, err := postgres.NewStorage(db)
	require.NoError(t, err)

	assetID, err := createSampleAsset(s)
	require.NoError(t, err)

	server := NewServer(s, nil)

	var tests = []testCase{
		{
			msg:      "create asset exchange",
			endpoint: settingChangePath,
			method:   http.MethodPost,
			data: &common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeCreateAssetExchange,
						Data: common.CreateAssetExchangeEntry{
							AssetID:           assetID,
							ExchangeID:        1,
							Symbol:            "ETH",
							DepositAddress:    eth.HexToAddress("0x007"),
							MinDeposit:        10.0,
							WithdrawFee:       11.0,
							TargetRecommended: 12.0,
							TargetRatio:       13.0,
						},
					},
					{
						Type: common.ChangeTypeUpdateAssetExchange,
						Data: common.UpdateAssetExchangeEntry{
							ID:             2,
							DepositAddress: common.AddressPointer(eth.HexToAddress("0x0000000000001")),
							MinDeposit:     common.FloatPointer(6.0),
							WithdrawFee:    common.FloatPointer(9.0),
						},
					},
				},
			},
			assert: httputil.ExpectSuccess,
		},
		{
			msg:      "confirm setting change",
			endpoint: fmt.Sprint(settingChangePath, "/", 1),
			method:   http.MethodPut,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				httputil.ExpectSuccess(t, resp)
				//check update asset exchange
				assetExchange, err := s.GetAssetExchange(2)
				require.NoError(t, err)
				require.Equal(t, 6.0, assetExchange.MinDeposit)
				require.Equal(t, 9.0, assetExchange.WithdrawFee)
				//check create asset exchange
				assetExchange, err = s.GetAssetExchange(4)
				require.NoError(t, err)
				require.Equal(t, "ETH", assetExchange.Symbol)
				require.Equal(t, uint64(1), assetExchange.ExchangeID)
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.msg, func(t *testing.T) {
			testHTTPRequest(t, tc, server.r)
		})
	}

}
