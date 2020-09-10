package blockchain

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KyberNetwork/reserve-data/common"
	baseblockchain "github.com/KyberNetwork/reserve-data/common/blockchain"
	"github.com/KyberNetwork/reserve-data/settings"
	settingsstorage "github.com/KyberNetwork/reserve-data/settings/storage"
)

const (
	rpcEndpoint = "https://mainnet.infura.io/v3/3243b06bf0334cff8e468bf90ce6e08c"
)

func newTestSetting(t *testing.T, tmpDir string) *settings.Settings {
	t.Helper()
	boltSettingStorage, err := settingsstorage.NewBoltSettingStorage(filepath.Join(tmpDir, "setting.db"))
	if err != nil {
		t.Fatal(err)
	}
	tokenSetting, err := settings.NewTokenSetting(boltSettingStorage)
	if err != nil {
		t.Fatal(err)
	}
	addressSetting := &settings.AddressSetting{
		Addresses: map[settings.AddressName]ethereum.Address{
			settings.Pricing: ethereum.HexToAddress("0x798AbDA6Cc246D0EDbA912092A2a3dBd3d11191B"),
			settings.Reserve: ethereum.HexToAddress(""),
			settings.Wrapper: ethereum.HexToAddress(""),
			settings.Proxy:   ethereum.HexToAddress(""),
		},
	}

	exchangeSetting, err := settings.NewExchangeSetting(boltSettingStorage)
	if err != nil {
		t.Fatal(err)
	}
	setting, err := settings.NewSetting(tokenSetting, addressSetting, exchangeSetting)
	if err != nil {
		t.Fatal(err)
	}
	return setting
}

// this is an external test
func TestGeneratedGetListedTokens(t *testing.T) {
	t.Skip() // skip as it is an external test
	log.Println("test generated get listed tokens")

	rpcClient, err := rpc.Dial(rpcEndpoint)
	require.NoError(t, err)

	ethClient := ethclient.NewClient(rpcClient)

	contractCallerClient, err := common.NewEthClient(rpcEndpoint)
	require.NoError(t, err)

	contractCaller := baseblockchain.NewContractCaller([]*common.EthClient{contractCallerClient})

	baseBlockchain := baseblockchain.NewBaseBlockchain(
		rpcClient,      // rpc client
		ethClient,      // eth client
		nil,            // map[string]*Operator
		nil,            // broadcaster
		contractCaller, // contract caller
	)

	tmpDir, err := ioutil.TempDir("", "test_setting")
	require.NoError(t, err)

	setting := newTestSetting(t, tmpDir)
	defer func() {
		if rErr := os.RemoveAll(tmpDir); rErr != nil {
			t.Error(rErr)
		}
	}()

	blockchain, err := NewBlockchain(baseBlockchain, setting)
	require.NoError(t, err)

	opts := blockchain.GetCallOpts(0)
	listedToken, err := blockchain.GeneratedGetListedTokens(opts)
	assert.NoError(t, err)
	assert.NotEmpty(t, listedToken)
	log.Printf("listed tokens: %d", len(listedToken))
	for _, token := range listedToken {
		log.Printf("token: %s", token.Hex())
	}
}
