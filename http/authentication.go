package http

import (
	"crypto/hmac"
	"crypto/sha512"

	ethereum "github.com/ethereum/go-ethereum/common"
)

type Authentication interface {
	KNSign(message string) string
	KNReadonlySign(message string) string
	KNConfigurationSign(message string) string
	KNConfirmConfSign(message string) string
	GetPermission(signed string, message string) []Permission
}

type KNAuthentication struct {
	KNSecret        string
	KNReadOnly      string
	KNConfiguration string
	KNConfirmConf   string
}

func (self KNAuthentication) KNSign(msg string) string {
	mac := hmac.New(sha512.New, []byte(self.KNSecret))
	mac.Write([]byte(msg))
	return ethereum.Bytes2Hex(mac.Sum(nil))
}

func (self KNAuthentication) KNReadonlySign(msg string) string {
	mac := hmac.New(sha512.New, []byte(self.KNReadOnly))
	mac.Write([]byte(msg))
	return ethereum.Bytes2Hex(mac.Sum(nil))
}

func (self KNAuthentication) KNConfigurationSign(msg string) string {
	mac := hmac.New(sha512.New, []byte(self.KNConfiguration))
	mac.Write([]byte(msg))
	return ethereum.Bytes2Hex(mac.Sum(nil))
}

func (self KNAuthentication) KNConfirmConfSign(msg string) string {
	mac := hmac.New(sha512.New, []byte(self.KNConfirmConf))
	mac.Write([]byte(msg))
	return ethereum.Bytes2Hex(mac.Sum(nil))
}

func (self KNAuthentication) GetPermission(signed string, message string) []Permission {
	result := []Permission{}
	rebalanceSigned := self.KNSign(message)
	if signed == rebalanceSigned {
		result = append(result, RebalancePermission)
	}
	readonlySigned := self.KNReadonlySign(message)
	if signed == readonlySigned {
		result = append(result, ReadOnlyPermission)
	}
	configureSigned := self.KNConfigurationSign(message)
	if signed == configureSigned {
		result = append(result, ConfigurePermission)
	}
	confirmConfSigned := self.KNConfirmConfSign(message)
	if signed == confirmConfSigned {
		result = append(result, ConfirmConfPermission)
	}
	return result
}
