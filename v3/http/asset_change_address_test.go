package http

import (
	"net/http"
	"testing"

	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KyberNetwork/reserve-data/common/testutil"
	"github.com/KyberNetwork/reserve-data/http/httputil"
	"github.com/KyberNetwork/reserve-data/v3/common"
	"github.com/KyberNetwork/reserve-data/v3/storage/postgres"
)

func TestHTTPServerChangeAssetAddress(t *testing.T) {
	const (
		changeAssetAddress = "/v3/change-asset-address"
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
			msg:      "create change asset address",
			endpoint: changeAssetAddress,
			method:   http.MethodPost,
			assert:   httputil.ExpectSuccess,
			data: &common.CreateChangeAssetAddress{
				Assets: []common.ChangeAssetAddressEntry{
					{
						ID:      assetID,
						Address: "0x4ddbda50ddbec0289a80de667a72d158819e381d",
					},
				},
			},
		},
		{
			msg:      "confirm pending change asset address",
			endpoint: changeAssetAddress + "/1",
			method:   http.MethodPut,
			data:     nil,
			assert:   httputil.ExpectSuccess,
		},
		{
			msg: "create invalid change asset address (with invalid address)",
			data: &common.CreateChangeAssetAddress{
				Assets: []common.ChangeAssetAddressEntry{
					{
						ID:      assetID,
						Address: "invalid address",
					},
				},
			},
			endpoint: changeAssetAddress,
			method:   http.MethodPost,
			assert:   httputil.ExpectFailureWithReason(common.ErrInvalidAddress.Error()),
		},
		{
			msg: "create invalid change asset address (with not exist asset)",
			data: &common.CreateChangeAssetAddress{
				Assets: []common.ChangeAssetAddressEntry{
					{
						ID:      1234,
						Address: "0x3f3150ea2b596f6bdb6c4af21b744019f29694c1",
					},
				},
			},
			endpoint: changeAssetAddress,
			method:   http.MethodPost,
			assert:   httputil.ExpectFailureWithReason(common.ErrNotFound.Error()),
		},
		{
			msg:      "create other valid change asset address",
			endpoint: changeAssetAddress,
			method:   http.MethodPost,
			assert:   httputil.ExpectSuccess,
			data: &common.CreateChangeAssetAddress{
				Assets: []common.ChangeAssetAddressEntry{
					{
						ID:      assetID,
						Address: "0x5dbcb95364cbc5604bacbb8c6eb9aa788f347a17",
					},
				},
			},
		},
		{
			msg:      "reject pending change asset address",
			endpoint: changeAssetAddress + "/2",
			method:   http.MethodDelete,
			data:     nil,
			assert:   httputil.ExpectSuccess,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.msg, func(t *testing.T) { testHTTPRequest(t, tc, server.r) })
	}
}

func TestChangeAssetAddress_Successfully(t *testing.T) {
	db, tearDown := testutil.MustNewDevelopmentDB()
	defer func() {
		assert.NoError(t, tearDown())
	}()

	s, err := postgres.NewStorage(db)
	require.NoError(t, err)

	assetID, err := createSampleAsset(s)
	require.NoError(t, err)

	id, err := s.CreatePendingObject(common.CreateChangeAssetAddress{
		Assets: []common.ChangeAssetAddressEntry{
			{
				ID:      assetID,
				Address: "0x5dbcb95364cbc5604bacbb8c6eb9aa788f347a17",
			},
		},
	}, common.PendingTypeChangeAssetAddr)
	require.NoError(t, err)
	require.NotZero(t, id)

	_, err = s.GetPendingObject(id, common.PendingTypeChangeAssetAddr)
	require.NoError(t, err)

	err = s.ConfirmChangeAssetAddress(id)
	require.NoError(t, err)

	asset, err := s.GetAsset(assetID)
	require.NoError(t, err)
	require.Equal(t, ethereum.HexToAddress("0x5dbcb95364cbc5604bacbb8c6eb9aa788f347a17"), asset.Address)
}

func TestChangeAssetAddress_FailedWithDuplicateAddress(t *testing.T) {
	db, tearDown := testutil.MustNewDevelopmentDB()
	defer func() {
		assert.NoError(t, tearDown())
	}()

	s, err := postgres.NewStorage(db)
	require.NoError(t, err)

	assetID, err := createSampleAsset(s)
	require.NoError(t, err)

	asset, err := s.GetAsset(assetID)
	require.NoError(t, err)

	id, err := s.CreatePendingObject(common.CreateChangeAssetAddress{
		Assets: []common.ChangeAssetAddressEntry{
			{
				ID:      assetID,
				Address: asset.Address.Hex(),
			},
		},
	}, common.PendingTypeChangeAssetAddr)
	require.NoError(t, err)

	_, err = s.GetPendingObject(id, common.PendingTypeChangeAssetAddr)
	require.NoError(t, err)

	err = s.ConfirmChangeAssetAddress(id)
	require.Equal(t, err, common.ErrAddressExists)
}
