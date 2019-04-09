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
	//Burner is the enumberated key for burner
	Burner //burner
	//Bank is the enumberated key for bank
	Bank //bank
	//Network is the enumberated key for network
	Network //network
	//Wrapper is the enumberated key for wrapper
	Wrapper //wrapper
	//Pricing is the enumberated key for pricing
	Pricing //pricing
	//Whitelist is the enumberated key for whitelist
	Whitelist //whitelist
	//InternalNetwork is the enumberated key for internal_network
	InternalNetwork //internal_network
)

var addressNameValues = map[string]AddressName{
	"reserve":          Reserve,
	"burner":           Burner,
	"bank":             Bank,
	"network":          Network,
	"wrapper":          Wrapper,
	"pricing":          Pricing,
	"whitelist":        Whitelist,
	"internal_network": InternalNetwork,
}

// AddressNameValues returns the mapping of the string presentation
// of address name and its value.
func AddressNameValues() map[string]AddressName {
	return addressNameValues
}

// AddressSetName is the name of ethereum address set used in core.
//go:generate stringer -type=AddressSetName -linecomment
type AddressSetName int

const (
	//ThirdPartyReserves is the enumerated key for third_party_reserves
	ThirdPartyReserves AddressSetName = iota //third_party_reserves
	//OldNetworks is the enumerated key for old_networks
	OldNetworks //old_networks
	//OldBurners is the enumerated key for old_burners
	OldBurners //old_burners
)

var addressSetNameValues = map[string]AddressSetName{
	"third_party_reserves": ThirdPartyReserves,
	"old_networks":         OldNetworks,
	"old_burners":          OldBurners,
}

// AddressSetNameValues returns the mapping of the string presentation
// of address set name and its value.
func AddressSetNameValues() map[string]AddressSetName {
	return addressSetNameValues
}

// AddressSetting type defines component to handle all address setting in core.
// It contains the storage interface used to query addresses.
type AddressSetting struct {
	Addresses   map[AddressName]ethereum.Address
	AddressSets map[AddressSetName]([]ethereum.Address)
}

//NewAddressSetting return an implementation of Address Setting
func NewAddressSetting(data common.AddressConfig) (*AddressSetting, error) {
	address := make(map[AddressName]ethereum.Address)
	addressSets := make(map[AddressSetName]([]ethereum.Address))
	addressSetting := &AddressSetting{
		Addresses:   address,
		AddressSets: addressSets,
	}
	if err := addressSetting.saveAddressFromAddressConfig(data); err != nil {
		return addressSetting, err
	}
	return addressSetting, nil
}

func (addrSetting *AddressSetting) saveAddressFromAddressConfig(addrs common.AddressConfig) error {
	addrSetting.Addresses[Bank] = ethereum.HexToAddress(addrs.Bank)
	addrSetting.Addresses[Reserve] = ethereum.HexToAddress(addrs.Reserve)
	addrSetting.Addresses[Network] = ethereum.HexToAddress(addrs.Network)
	addrSetting.Addresses[Wrapper] = ethereum.HexToAddress(addrs.Wrapper)
	addrSetting.Addresses[Pricing] = ethereum.HexToAddress(addrs.Pricing)
	addrSetting.Addresses[Burner] = ethereum.HexToAddress(addrs.FeeBurner)
	addrSetting.Addresses[Whitelist] = ethereum.HexToAddress(addrs.Whitelist)
	addrSetting.Addresses[InternalNetwork] = ethereum.HexToAddress(addrs.InternalNetwork)
	thirdParty := []ethereum.Address{}

	for _, addr := range addrs.ThirdPartyReserves {
		thirdParty = append(thirdParty, ethereum.HexToAddress(addr))
	}
	addrSetting.AddressSets[ThirdPartyReserves] = thirdParty
	return nil
}
