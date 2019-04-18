package settings

import (
	"github.com/KyberNetwork/reserve-data/common"
	ethereum "github.com/ethereum/go-ethereum/common"
)

// AddressName is the name of ethereum address used in core.
//go:generate stringer -type=AddressName -linecomment
type AddressName int

const (
	//Reserve is the enumerated key for reserve
	Reserve AddressName = iota //reserve
	//Network is the enumerated key for network
	Wrapper //wrapper
	//Pricing is the enumerated key for pricing
	Pricing //pricing
)

var addressNameValues = map[string]AddressName{
	"reserve": Reserve,
	"wrapper": Wrapper,
	"pricing": Pricing,
}

// AddressNameValues returns the mapping of the string presentation
// of address name and its value.
func AddressNameValues() map[string]AddressName {
	return addressNameValues
}

// AddressSetting type defines component to handle all address setting in core.
// It contains the storage interface used to query addresses.
type AddressSetting struct {
	Addresses map[AddressName]ethereum.Address
}

//NewAddressSetting return an implementation of Address Setting
func NewAddressSetting(data common.AddressConfig) (*AddressSetting, error) {
	address := make(map[AddressName]ethereum.Address)
	addressSetting := &AddressSetting{
		Addresses: address,
	}
	if err := addressSetting.saveAddressFromAddressConfig(data); err != nil {
		return addressSetting, err
	}
	return addressSetting, nil
}

func (addrSetting *AddressSetting) saveAddressFromAddressConfig(addrs common.AddressConfig) error {
	addrSetting.Addresses[Reserve] = ethereum.HexToAddress(addrs.Reserve)
	addrSetting.Addresses[Wrapper] = ethereum.HexToAddress(addrs.Wrapper)
	addrSetting.Addresses[Pricing] = ethereum.HexToAddress(addrs.Pricing)
	return nil
}
