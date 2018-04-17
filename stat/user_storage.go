package stat

type UserStorage interface {
	UpdateAddressCategory(address string, cat string) error
	UpdateUserAddresses(user string, addresses []string, timestamps []uint64) error
	SetLastProcessedCatLogTimepoint(timepoint uint64) error

	// returns lowercased category of an address
	GetCategory(addr string) (string, error)
	GetAddressesOfUser(user string) (addresses []string, registeredTimes []uint64, err error)
	// returns lowercased user identity of the address
	GetUserOfAddress(addr string) (email string, registeredTime uint64, err error)
	GetLastProcessedCatLogTimepoint() (timepoint uint64, err error)

	GetKycUsers() (map[string]uint64, error)

	// returns all of addresses that's not pushed to the chain
	// for kyced category
	GetPendingAddresses() ([]string, error)
}
