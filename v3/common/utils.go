package common

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// IsZeroAddress return if hash is zero.
func IsZeroAddress(addr common.Address) bool {
	return addr.Hash().Big().Cmp(big.NewInt(0)) == 0
}

// FloatPointer is helper use in optional parameter
func FloatPointer(f float64) *float64 {
	return &f
}

// StringPointer is helper use in optional parameter
func StringPointer(s string) *string {
	return &s
}

// BoolPointer is helper use in optional parameter
func BoolPointer(b bool) *bool {
	return &b
}

// AddressPointer convert address to pointer
func AddressPointer(a common.Address) *common.Address {
	return &a
}

// Uint64Pointer convert uint64 to pointer
func Uint64Pointer(i uint64) *uint64 {
	return &i
}

// SetRatePointer return SetRate pointer
func SetRatePointer(i SetRate) *SetRate {
	return &i
}

// SettingChangeFromType create an empty object for correspond type
func SettingChangeFromType(t ChangeType) (SettingChangeType, error) {
	var i SettingChangeType
	switch t {
	case ChangeTypeUnknown:
		return nil, fmt.Errorf("got unknow change type")
	case ChangeTypeCreateAsset:
		i = &CreateAssetEntry{}
	case ChangeTypeUpdateAsset:
		i = &UpdateAssetEntry{}
	case ChangeTypeCreateAssetExchange:
		i = &CreateAssetExchangeEntry{}
	case ChangeTypeUpdateAssetExchange:
		i = &UpdateAssetExchangeEntry{}
	case ChangeTypeCreateTradingPair:
		i = &CreateTradingPairEntry{}
	case ChangeTypeCreateTradingBy:
		i = &CreateTradingByEntry{}
	case ChangeTypeUpdateExchange:
		i = &UpdateExchangeEntry{}
	case ChangeTypeChangeAssetAddr:
		i = &ChangeAssetAddressEntry{}
	case ChangeTypeDeleteTradingPair:
		// TODO: process delete entry
	case ChangeTypeDeleteAssetExchange:

	case ChangeTypeDeleteTradingBy:
	}
	return i, nil
}
