package middleware

const maxTimeIntervalAllow = 30000

// ValidateNonce interface for checking if a nonce is valid or not
type ValidateNonce interface {
	IsValid(int64) bool
}

// ValidateNonceByTime checking validate by time range
type ValidateNonceByTime struct {
	// RangeAllow maximum time diffirent between server and client
	RangeAllow uint64
}

// NewValidateNonceByTime return ValidateNonceByTime with default value
func NewValidateNonceByTime() ValidateNonceByTime {
	return ValidateNonceByTime{
		RangeAllow: maxTimeIntervalAllow,
	}
}

// IsValid return nonce is valid or not by time range
func (v ValidateNonceByTime) IsValid(nonce int64) bool {
	serverTime := getTimepoint()
	difference := nonce - int64(serverTime)
	if difference < -maxTimeIntervalAllow || difference > maxTimeIntervalAllow {
		return false
	}
	return true
}
