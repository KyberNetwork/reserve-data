// Code generated by "enumer -type=ChangeCatalog -linecomment -json=true"; DO NOT EDIT.

//
package common

import (
	"encoding/json"
	"fmt"
)

const _ChangeCatalogName = "set_targetset_pwisset_stable_tokenset_rebalance_quadraticupdate_exchangemainset_feed_configuration"

var _ChangeCatalogIndex = [...]uint8{0, 10, 18, 34, 57, 72, 76, 98}

func (i ChangeCatalog) String() string {
	if i < 0 || i >= ChangeCatalog(len(_ChangeCatalogIndex)-1) {
		return fmt.Sprintf("ChangeCatalog(%d)", i)
	}
	return _ChangeCatalogName[_ChangeCatalogIndex[i]:_ChangeCatalogIndex[i+1]]
}

var _ChangeCatalogValues = []ChangeCatalog{0, 1, 2, 3, 4, 5, 6}

var _ChangeCatalogNameToValueMap = map[string]ChangeCatalog{
	_ChangeCatalogName[0:10]:  0,
	_ChangeCatalogName[10:18]: 1,
	_ChangeCatalogName[18:34]: 2,
	_ChangeCatalogName[34:57]: 3,
	_ChangeCatalogName[57:72]: 4,
	_ChangeCatalogName[72:76]: 5,
	_ChangeCatalogName[76:98]: 6,
}

// ChangeCatalogString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func ChangeCatalogString(s string) (ChangeCatalog, error) {
	if val, ok := _ChangeCatalogNameToValueMap[s]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("%s does not belong to ChangeCatalog values", s)
}

// ChangeCatalogValues returns all values of the enum
func ChangeCatalogValues() []ChangeCatalog {
	return _ChangeCatalogValues
}

// IsAChangeCatalog returns "true" if the value is listed in the enum definition. "false" otherwise
func (i ChangeCatalog) IsAChangeCatalog() bool {
	for _, v := range _ChangeCatalogValues {
		if i == v {
			return true
		}
	}
	return false
}

// MarshalJSON implements the json.Marshaler interface for ChangeCatalog
func (i ChangeCatalog) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface for ChangeCatalog
func (i *ChangeCatalog) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("ChangeCatalog should be a string, got %s", data)
	}

	var err error
	*i, err = ChangeCatalogString(s)
	return err
}
