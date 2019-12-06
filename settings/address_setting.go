package settings

import (
	ethereum "github.com/ethereum/go-ethereum/common"

	"github.com/KyberNetwork/reserve-data/common"
)

// AddressName is the name of ethereum address used in core.
//go:generate stringer -type=AddressName -linecomment
type AddressName int

const (
	//Reserve is the enumerated key for reserve
	Reserve AddressName = iota //reserve
	// Proxy address
	// as we used to call it network
	// we keep its string name as is so other component won't need to change
	Proxy //network
	//Wrapper is the enumberated key for wrapper
	Wrapper //wrapper
	//Pricing is the enumberated key for pricing
	Pricing //pricing
)

var addressNameValues = map[string]AddressName{
	"reserve": Reserve,
	"wrapper": Wrapper,
	"network": Proxy,
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
func NewAddressSetting(data common.AddressConfig) *AddressSetting {
	address := make(map[AddressName]ethereum.Address)
	addressSetting := &AddressSetting{
		Addresses: address,
	}
	addressSetting.saveAddressFromAddressConfig(data)
	return addressSetting
}

func (addrSetting *AddressSetting) saveAddressFromAddressConfig(addrs common.AddressConfig) {
	addrSetting.Addresses[Reserve] = ethereum.HexToAddress(addrs.Reserve)
	addrSetting.Addresses[Wrapper] = ethereum.HexToAddress(addrs.Wrapper)
	addrSetting.Addresses[Pricing] = ethereum.HexToAddress(addrs.Pricing)
	addrSetting.Addresses[Proxy] = ethereum.HexToAddress(addrs.Proxy)
	return
}
