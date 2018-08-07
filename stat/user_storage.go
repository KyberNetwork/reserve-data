package stat

import ethereum "github.com/ethereum/go-ethereum/common"

// UserStorage is the interface of database contains information
// of users that doing the currency exchanges.
// Category is used for setting different limitations to address.
// When a user is KYC'ed, the database stores a mapping of email
// and Ethereum addresses for later retrieval purpose.
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

	// return email of address if address is kyced else return empty string
	GetKYCAddress(address ethereum.Address) (string, error)
}
