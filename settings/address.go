package settings

import (
	"errors"

	ethereum "github.com/ethereum/go-ethereum/common"
)

var ErrNoAddr = errors.New("cannot find the address")

func (setting *Settings) GetAddress(name AddressName) (ethereum.Address, error) {
	return setting.Address.GetAddress(name)
}

func (addrsetting *AddressSetting) GetAddress(name AddressName) (ethereum.Address, error) {
	addr, ok := addrsetting.Addresses[name]
	if !ok {
		return ethereum.Address{}, ErrNoAddr
	}
	return addr, nil
}

// GetAllAddresses return all the address setting in cores.
func (setting *Settings) GetAllAddresses() (map[string]interface{}, error) {
	allAddress := make(map[string]interface{})
	for name, addr := range setting.Address.Addresses {
		allAddress[name.String()] = addr
	}
	return allAddress, nil
}
