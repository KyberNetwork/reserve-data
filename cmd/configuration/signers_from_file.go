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

func PricingSignerFromConfigFile(secretPath string) *blockchain.EthereumSigner {
	raw, err := ioutil.ReadFile(secretPath)
	if err != nil {
		panic(err)
	}
	detail := jsonPricingDetail{}
	err = json.Unmarshal(raw, &detail)
	if err != nil {
		panic(err)
	}
	return blockchain.NewEthereumSigner(detail.Keystore, detail.Passphrase)
}

type jsonDepositDetail struct {
	Keystore   string `json:"keystore_deposit_path"`
	Passphrase string `json:"passphrase_deposit"`
}

func DepositSignerFromConfigFile(secretPath string) *blockchain.EthereumSigner {
	raw, err := ioutil.ReadFile(secretPath)
	if err != nil {
		panic(err)
	}
	detail := jsonDepositDetail{}
	err = json.Unmarshal(raw, &detail)
	if err != nil {
		panic(err)
	}
	return blockchain.NewEthereumSigner(detail.Keystore, detail.Passphrase)
}

type jsonHuobiIntermediatorDetail struct {
	Keystore   string `json:"keystore_intermediator_path"`
	Passphrase string `json:"passphrase_intermediate_account"`
}

func HuobiIntermediatorSignerFromFile(secretPath string) *blockchain.EthereumSigner {
	raw, err := ioutil.ReadFile(secretPath)
	if err != nil {
		panic(err)
	}
	detail := jsonHuobiIntermediatorDetail{}
	err = json.Unmarshal(raw, &detail)
	if err != nil {
		panic(err)
	}
	return blockchain.NewEthereumSigner(detail.Keystore, detail.Passphrase)
}
