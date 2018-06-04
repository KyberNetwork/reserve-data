package settings

import (
	"encoding/json"
	"io/ioutil"
)

// AddressName is the name of ethereum address used in core.
//go:generate stringer -type=AddressName
type AddressName int

const (
	Reserve AddressName = iota
	Burner
	Bank
	Network
	Wrapper
	Pricing
	Whitelist
	SetRate
	Intermediator
)

var addressNameValues = map[string]AddressName{
	"reserve":       Reserve,
	"burner":        Burner,
	"bank":          Bank,
	"network":       Network,
	"wrapper":       Wrapper,
	"pricing":       Pricing,
	"whitelist":     Whitelist,
	"setrate":       SetRate,
	"intermediator": Intermediator,
}

// AddressTypeValues returns the mapping of the string presentation
// of address name and its value.
func AddressTypeValues() map[string]AddressName {
	return addressNameValues
}

// AddressSetName is the name of ethereum address set used in core.
//go:generate stringer -type=AddressSetName
type AddressSetName int

const (
	ThirdPartyReserves AddressSetName = iota
	OldNetWorks
	OldBurners
)

var addressSetNameValues = map[string]AddressSetName{
	"third_party_reserves": ThirdPartyReserves,
	"oldNetworks":          OldNetWorks,
	"oldBurners":           OldBurners,
}

// AddressSetTypeValues returns the mapping of the string presentation
// of address set name and its value.
func AddressSetTypeValues() map[string]AddressSetName {
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
	SetRate            string   `json:"setrate"`
	Intermediator      string   `json:"intermediator"`
}

// AddressSetting type defines component to handle all address setting in core.
// It contains the storage interface used to query addresses.
type AddressSetting struct {
	Storage AddressStorage
}

func NewAddressSetting(addressStorage AddressStorage) (*AddressSetting, error) {
	return &AddressSetting{addressStorage}, nil
}

func (setting *Settings) LoadAddressFromFile(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	addrs := AddressConfig{}
	if err = json.Unmarshal(data, &addrs); err != nil {
		return err
	}
	if err = setting.Address.Storage.UpdateOneAddress(Bank, addrs.Bank); err != nil {
		return err
	}
	if err = setting.Address.Storage.UpdateOneAddress(Reserve, addrs.Reserve); err != nil {
		return err
	}
	if err = setting.Address.Storage.UpdateOneAddress(Network, addrs.Network); err != nil {
		return err
	}
	if err = setting.Address.Storage.UpdateOneAddress(Wrapper, addrs.Wrapper); err != nil {
		return err
	}
	if err = setting.Address.Storage.UpdateOneAddress(Pricing, addrs.Pricing); err != nil {
		return err
	}
	if err = setting.Address.Storage.UpdateOneAddress(Burner, addrs.FeeBurner); err != nil {
		return err
	}
	if err = setting.Address.Storage.UpdateOneAddress(Whitelist, addrs.Whitelist); err != nil {
		return err
	}
	if err = setting.Address.Storage.UpdateOneAddress(SetRate, addrs.SetRate); err != nil {
		return err
	}
	if err = setting.Address.Storage.UpdateOneAddress(Intermediator, addrs.Intermediator); err != nil {
		return err
	}
	for _, addr := range addrs.ThirdPartyReserves {
		if err = setting.Address.Storage.AddAddressToSet(ThirdPartyReserves, addr); err != nil {
			return err
		}
	}
	return nil
}
