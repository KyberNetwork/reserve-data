package stat

import ethereum "github.com/ethereum/go-ethereum/common"

type UserStorage interface {
	UpdateAddressCategory(address ethereum.Address, cat string) error
	UpdateUserAddresses(user string, addresses []ethereum.Address, timestamps []uint64) error
	SetLastProcessedCatLogTimepoint(timepoint uint64) error

	// returns lowercased category of an address
	GetCategory(addr ethereum.Address) (string, error)
	GetAddressesOfUser(user string) (addresses []ethereum.Address, registeredTimes []uint64, err error)
	// returns lowercased user identity of the address
	GetUserOfAddress(addr ethereum.Address) (email string, registeredTime uint64, err error)
	GetLastProcessedCatLogTimepoint() (timepoint uint64, err error)

	GetKycUsers() (map[string]uint64, error)

	// returns all of addresses that's not pushed to the chain
	// for kyced category
	GetPendingAddresses() ([]ethereum.Address, error)
}
