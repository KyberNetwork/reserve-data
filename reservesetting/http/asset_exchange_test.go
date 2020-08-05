package http

import (
	"testing"

	eth "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1common "github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/testutil"
	"github.com/KyberNetwork/reserve-data/reservesetting/common"
	"github.com/KyberNetwork/reserve-data/reservesetting/storage/postgres"
)

const (
	migrationPath = "../../cmd/migrations"
)

func TestServer_UpdateAssetExchange(t *testing.T) {

	var (
		supportedExchanges = make(map[v1common.ExchangeID]v1common.LiveExchange)
	)

	// create map of test exchange
	for _, exchangeID := range []v1common.ExchangeID{v1common.Binance, v1common.Huobi} {
		exchange := v1common.TestExchange{}
		supportedExchanges[exchangeID] = exchange
	}

	db, tearDown := testutil.MustNewDevelopmentDB(migrationPath)
	defer func() {
		assert.NoError(t, tearDown())
	}()
	s, err := postgres.NewStorage(db)
	require.NoError(t, err)

	assetID, err := createSampleAsset(s)
	require.NoError(t, err)
	server := NewServer(s, "", supportedExchanges, "", nil)
	c := apiClient{s: server}
	assetResp, err := c.getAsset(assetID)
	require.NoError(t, err)

	selectAssetExchange := assetResp.Asset.AssetExchanges[0]
	updateAxe := common.UpdateAssetExchangeEntry{
		ID:                selectAssetExchange.ID,
		Symbol:            common.StringPointer("XYZ"),
		DepositAddress:    common.AddressPointer(eth.HexToAddress("0x232425")),
		MinDeposit:        common.FloatPointer(3.1),
		WithdrawFee:       common.FloatPointer(3.2),
		TargetRecommended: common.FloatPointer(3.3),
		TargetRatio:       common.FloatPointer(3.4),
	}
	pending, err := c.createSettingChange(common.SettingChange{ChangeList: []common.SettingChangeEntry{
		{
			Type: common.ChangeTypeUpdateAssetExchange,
			Data: updateAxe,
		},
	},
		Message: "Update asset exchange",
	})
	require.NoError(t, err)
	confirm, err := c.confirmSettingChange(pending.ID)
	require.NoError(t, err)
	require.Equal(t, true, confirm.Success)
	updatedResp, err := c.getAsset(assetID)
	require.NoError(t, err)
	found := false

	for _, assetExchange := range updatedResp.Asset.AssetExchanges {
		if assetExchange.ID == selectAssetExchange.ID {
			found = true
			assert.Equal(t, *updateAxe.MinDeposit, assetExchange.MinDeposit)
			assert.Equal(t, *updateAxe.WithdrawFee, assetExchange.WithdrawFee)
			assert.Equal(t, *updateAxe.TargetRecommended, assetExchange.TargetRecommended)
			assert.Equal(t, *updateAxe.TargetRatio, assetExchange.TargetRatio)
			assert.Equal(t, *updateAxe.DepositAddress, assetExchange.DepositAddress)
			assert.Equal(t, *updateAxe.Symbol, assetExchange.Symbol)
			break
		}
	}
	assert.Equal(t, true, found)
	assets, err := c.getAssets()
	assert.NoError(t, err)
	assert.Len(t, assets.Assets, 2)
}
