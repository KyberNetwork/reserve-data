package configuration

import (
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/archive"
	"github.com/KyberNetwork/reserve-data/common/blockchain"
	ccfg "github.com/KyberNetwork/reserve-data/common/config"
	"github.com/KyberNetwork/reserve-data/http"
	"github.com/KyberNetwork/reserve-data/settings"
	settingstorage "github.com/KyberNetwork/reserve-data/settings/storage"
	"github.com/KyberNetwork/reserve-data/world"
)

func GetSetting(ac ccfg.AppConfig, addressSetting *settings.AddressSetting) (*settings.Settings, error) {
	boltSettingStorage, err := settingstorage.NewBoltSettingStorage(ac.SettingDB)
	if err != nil {
		return nil, err
	}
	tokenSetting, err := settings.NewTokenSetting(boltSettingStorage)
	if err != nil {
		return nil, err
	}
	exchangeSetting, err := settings.NewExchangeSetting(boltSettingStorage)
	if err != nil {
		return nil, err
	}
	setting, err := settings.NewSetting(
		tokenSetting,
		addressSetting,
		exchangeSetting,
		settings.WithHandleEmptyToken(mustGetTokenConfig(ac)),
		settings.WithHandleEmptyFee(FeeConfigs),
		settings.WithHandleEmptyMinDeposit(ExchangesMinDepositConfig),
		settings.WithHandleEmptyDepositAddress(mustGetDepositAddress(ac)),
		settings.WithHandleEmptyExchangeInfo())
	return setting, err
}

func newTheWorld(exp ccfg.WorldEndpoints) (*world.TheWorld, error) {
	endpoint := world.NewWorldEndpoint(exp)
	return world.NewTheWorld(endpoint, zap.S()), nil
}

func InitAppState(authEnbl bool, ac ccfg.AppConfig, mainNode *common.EthClient, backupNodes []*common.EthClient) *AppState {
	l := zap.S()
	theWorld, err := newTheWorld(ac.WorldEndpoints)
	if err != nil {
		l.Panicf("Can't init the world (which is used to get global data), err=%+v", err)
	}

	hmac512auth := http.NewKNAuthentication(ac.Authentication.KNSecret, ac.Authentication.KNReadOnly,
		ac.Authentication.KNConfiguration, ac.Authentication.KNConfirmConfiguration)
	addressSetting := settings.NewAddressSetting(common.AddressConfig{
		Reserve: ac.ContractAddresses.Reserve.String(),
		Wrapper: ac.ContractAddresses.Wrapper.String(),
		Pricing: ac.ContractAddresses.Pricing.String(),
		Proxy:   ac.ContractAddresses.Proxy.String(),
	})

	bkClients := map[string]*ethclient.Client{}
	callClient := make([]*common.EthClient, 0, len(backupNodes)+1)
	callClient = append(callClient, mainNode)
	for _, n := range backupNodes {
		callClient = append(callClient, n)
		bkClients[n.URL] = n.Client
	}
	bkClients[mainNode.URL] = mainNode.Client

	bc := blockchain.NewBaseBlockchain(
		mainNode.RPCClient, mainNode.Client, map[string]*blockchain.Operator{},
		blockchain.NewBroadcaster(bkClients),
		blockchain.NewContractCaller(callClient),
	)
	if !authEnbl {
		l.Warnw("WARNING: No authentication mode")
	}
	s3archive := archive.NewS3Archive(ac.AWSConfig)
	aps := &AppState{
		Blockchain:              bc,
		EthereumEndpoint:        ac.Node.Main,
		BackupEthereumEndpoints: ac.Node.Backup,
		AuthEngine:              hmac512auth,
		EnableAuthentication:    authEnbl,
		Archive:                 s3archive,
		World:                   theWorld,
		AddressSetting:          addressSetting,
		AppConfig:               ac,
	}

	l.Infof("configured endpoint: %s, backup: %v", aps.EthereumEndpoint, aps.BackupEthereumEndpoints)

	aps.AddCoreConfig(ac)
	return aps
}
