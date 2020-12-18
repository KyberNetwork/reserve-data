package common

import (
	"github.com/KyberNetwork/httpsign-utils/authenticator"
)

// KeyPair ...
type KeyPair struct {
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
}

// ToAuthenticatorKeyPair ...
func (kp KeyPair) ToAuthenticatorKeyPair() authenticator.KeyPair {
	return authenticator.KeyPair{
		AccessKeyID:     kp.AccessKeyID,
		SecretAccessKey: kp.SecretAccessKey,
	}
}
