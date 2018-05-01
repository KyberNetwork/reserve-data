package common

import (
	"encoding/json"
	"io/ioutil"

	ethereum "github.com/ethereum/go-ethereum/common"
)

type token struct {
	Address          string `json:"address"`
	Name             string `json:"name"`
	Decimals         int64  `json:"decimals"`
	KNReserveSupport bool   `json:"internal use"`
	Active           bool   `json:"listed"`
}

type exchange map[string]string

type TokenInfo struct {
	Address  ethereum.Address `json:"address"`
	Decimals int64            `json:"decimals"`
}

type AddressConfig struct {
	Tokens             map[string]token    `json:"tokens"`
	Exchanges          map[string]exchange `json:"exchanges"`
	Bank               string              `json:"bank"`
	Reserve            string              `json:"reserve"`
	Network            string              `json:"network"`
	Wrapper            string              `json:"wrapper"`
	Pricing            string              `json:"pricing"`
	FeeBurner          string              `json:"feeburner"`
	Whitelist          string              `json:"whitelist"`
	ThirdPartyReserves []string            `json:"third_party_reserves"`
	Intermediator      string              `json:"intermediator"`
	SetRate            string              `json:"setrate"`
}

func GetAddressConfigFromFile(path string) (AddressConfig, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return AddressConfig{}, err
	} else {
		result := AddressConfig{}
		err := json.Unmarshal(data, &result)
		return result, err
	}
}

type Addresses struct {
	Tokens               map[string]TokenInfo          `json:"tokens"`
	Exchanges            map[ExchangeID]TokenAddresses `json:"exchanges"`
	WrapperAddress       ethereum.Address              `json:"wrapper"`
	PricingAddress       ethereum.Address              `json:"pricing"`
	ReserveAddress       ethereum.Address              `json:"reserve"`
	FeeBurnerAddress     ethereum.Address              `json:"feeburner"`
	NetworkAddress       ethereum.Address              `json:"network"`
	PricingOperator      ethereum.Address              `json:"pricing_operator"`
	DepositOperator      ethereum.Address              `json:"deposit_opeartor"`
	IntermediateOperator ethereum.Address              `json:"intermediate_operator"`
}

type TokenAddresses map[string]ethereum.Address
