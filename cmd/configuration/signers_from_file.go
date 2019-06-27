package configuration

import (
	"encoding/json"
	"io/ioutil"

	"github.com/KyberNetwork/reserve-data/common/blockchain"
)

type jsonPricingDetail struct {
	Keystore   string `json:"keystore_path"`
	Passphrase string `json:"passphrase"`
}

func PricingSignerFromConfigFile(secretPath string) (*blockchain.EthereumSigner, error) {
	raw, err := ioutil.ReadFile(secretPath)
	if err != nil {
		return nil, err
	}
	detail := jsonPricingDetail{}
	err = json.Unmarshal(raw, &detail)
	if err != nil {
		return nil, err
	}
	return blockchain.NewEthereumSigner(detail.Keystore, detail.Passphrase), nil
}

type jsonDepositDetail struct {
	Keystore   string `json:"keystore_deposit_path"`
	Passphrase string `json:"passphrase_deposit"`
}

func DepositSignerFromConfigFile(secretPath string) (*blockchain.EthereumSigner, error) {
	raw, err := ioutil.ReadFile(secretPath)
	if err != nil {
		return nil, err
	}
	detail := jsonDepositDetail{}
	err = json.Unmarshal(raw, &detail)
	if err != nil {
		return nil, err
	}
	return blockchain.NewEthereumSigner(detail.Keystore, detail.Passphrase), nil
}

type jsonHuobiIntermediatorDetail struct {
	Keystore   string `json:"keystore_intermediator_path"`
	Passphrase string `json:"passphrase_intermediate_account"`
}

func HuobiIntermediatorSignerFromFile(secretPath string) (*blockchain.EthereumSigner, error) {
	raw, err := ioutil.ReadFile(secretPath)
	if err != nil {
		return nil, err
	}
	detail := jsonHuobiIntermediatorDetail{}
	err = json.Unmarshal(raw, &detail)
	if err != nil {
		return nil, err
	}
	return blockchain.NewEthereumSigner(detail.Keystore, detail.Passphrase), nil
}
