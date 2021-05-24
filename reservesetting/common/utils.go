package common

import (
	"github.com/ethereum/go-ethereum/common"
)

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
	case ChangeTypeUpdateExchange:
		i = &UpdateExchangeEntry{}
	case ChangeTypeChangeAssetAddr:
		i = &ChangeAssetAddressEntry{}
	case ChangeTypeDeleteTradingPair:
		i = &DeleteTradingPairEntry{}
	case ChangeTypeDeleteAssetExchange:
		i = &DeleteAssetExchangeEntry{}
	case ChangeTypeUpdateStableTokenParams:
		i = &UpdateStableTokenParamsEntry{}
	case ChangeTypeSetFeedConfiguration:
		i = &SetFeedConfigurationEntry{}
	case ChangeTypeUpdateTradingPair:
		i = &UpdateTradingPairEntry{}
	}
	return i, nil
}
