// Code generated by "stringer -type=ExchangeID -linecomment"; DO NOT EDIT.

package common

import "strconv"

const _ExchangeID_name = "binancehuobistable_exchange"

var _ExchangeID_index = [...]uint8{0, 7, 12, 27}

func (i ExchangeID) String() string {
	i -= 1
	if i < 0 || i >= ExchangeID(len(_ExchangeID_index)-1) {
		return "ExchangeID(" + strconv.FormatInt(int64(i+1), 10) + ")"
	}
	return _ExchangeID_name[_ExchangeID_index[i]:_ExchangeID_index[i+1]]
}
