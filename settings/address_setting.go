package settings

import (
	"encoding/json"

	ethereum "github.com/ethereum/go-ethereum/common"
)

// AddressName is the name of ethereum address used in core.
//go:generate stringer -type=AddressName -linecomment
type AddressName int

const (
	Reserve         AddressName = iota //reserve
	Burner                             //burner
	Bank                               //bank
	Network                            //network
	Wrapper                            //wrapper
	Pricing                            //pricing
	Whitelist                          //whitelist
	InternalNetwork                    //internal_network
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
	ThirdPartyReserves AddressSetName = iota //third_party_reserves
	OldNetworks                              //old_networks
	OldBurners                               //old_burners
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

// AddressConfig type defines a list of address attribute avaiable in core.
// It is used mainly for
type AddressConfig struct {
	Bank               string   `json:"bank"`
	Reserve            string   `json:"reserve"`
	Network            string   `json:"network"`
	Wrapper            string   `json:"wrapper"`
	Pricing            string   `json:"pricing"`
	FeeBurner          string   `json:"feeburner"`
	Whitelist          string   `json:"whitelist"`
	ThirdPartyReserves []string `json:"third_party_reserves"`
	InternalNetwork    string   `json:"internal network"`
}

// AddressSetting type defines component to handle all address setting in core.
// It contains the storage interface used to query addresses.
type AddressSetting struct {
	Addresses   map[AddressName]ethereum.Address
	AddressSets map[AddressSetName]([]ethereum.Address)
}

func NewAddressSetting(data string) (*AddressSetting, error) {
	address := make(map[AddressName]ethereum.Address)
	addressSets := make(map[AddressSetName]([]ethereum.Address))
	addressSetting := &AddressSetting{
		Addresses:   address,
		AddressSets: addressSets,
	}
	if err := addressSetting.loadAddressFromString(data); err != nil {
		return addressSetting, err
	}
	return addressSetting, nil
}

func (addrSetting *AddressSetting) loadAddressFromString(data string) error {
	addrs := AddressConfig{}
	if err := json.Unmarshal([]byte(data), &addrs); err != nil {
		return err
	}
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
