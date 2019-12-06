package http

import (
	"crypto/hmac"
	"crypto/sha512"

	ethereum "github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

// Authentication is the authentication layer of HTTP APIs.
type Authentication interface {
	GetPermission(signed string, message string) []Permission
}

type KNAuthentication struct {
	KNSecret        string `json:"kn_secret"`
	KNReadOnly      string `json:"kn_readonly"`
	KNConfiguration string `json:"kn_configuration"`
	KNConfirmConf   string `json:"kn_confirm_configuration"`
}

// NewKNAuthentication ...
func NewKNAuthentication(secret, readOnly, configuration, confirm string) KNAuthentication {
	return KNAuthentication{
		KNSecret:        secret,
		KNReadOnly:      readOnly,
		KNConfiguration: configuration,
		KNConfirmConf:   confirm,
	}
}

func (auth KNAuthentication) KNSign(msg string) string {
	mac := hmac.New(sha512.New, []byte(auth.KNSecret))
	if _, err := mac.Write([]byte(msg)); err != nil {
		zap.S().Panic(err)
	}
	return ethereum.Bytes2Hex(mac.Sum(nil))
}

func (auth KNAuthentication) knReadonlySign(msg string) string {
	mac := hmac.New(sha512.New, []byte(auth.KNReadOnly))
	if _, err := mac.Write([]byte(msg)); err != nil {
		zap.S().Panic(err)
	}
	return ethereum.Bytes2Hex(mac.Sum(nil))
}

func (auth KNAuthentication) knConfigurationSign(msg string) string {
	mac := hmac.New(sha512.New, []byte(auth.KNConfiguration))
	if _, err := mac.Write([]byte(msg)); err != nil {
		zap.S().Panic(err)
	}
	return ethereum.Bytes2Hex(mac.Sum(nil))
}

func (auth KNAuthentication) knConfirmConfSign(msg string) string {
	mac := hmac.New(sha512.New, []byte(auth.KNConfirmConf))
	if _, err := mac.Write([]byte(msg)); err != nil {
		zap.S().Panic(err)
	}
	return ethereum.Bytes2Hex(mac.Sum(nil))
}

func (auth KNAuthentication) GetPermission(signed string, message string) []Permission {
	result := []Permission{}
	rebalanceSigned := auth.KNSign(message)
	if signed == rebalanceSigned {
		result = append(result, RebalancePermission)
	}
	readonlySigned := auth.knReadonlySign(message)
	if signed == readonlySigned {
		result = append(result, ReadOnlyPermission)
	}
	configureSigned := auth.knConfigurationSign(message)
	if signed == configureSigned {
		result = append(result, ConfigurePermission)
	}
	confirmConfSigned := auth.knConfirmConfSign(message)
	if signed == confirmConfSigned {
		result = append(result, ConfirmConfPermission)
	}
	return result
}
