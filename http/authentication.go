package http

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/json"
	"io/ioutil"
	"log"

	ethereum "github.com/ethereum/go-ethereum/common"
)

// Authentication is the authentication layer of HTTP APIs.
type Authentication interface {
	KNSign(message string) string
	GetPermission(signed string, message string) []Permission
}

// KNAuthentication save authentication key for each permission
type KNAuthentication struct {
	KNSecret        string `json:"kn_secret"`
	KNReadOnly      string `json:"kn_readonly"`
	KNConfiguration string `json:"kn_configuration"`
	KNConfirmConf   string `json:"kn_confirm_configuration"`
}

// NewKNAuthenticationFromFile read authentication from file and save to KNAuthentication object
func NewKNAuthenticationFromFile(path string) KNAuthentication {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	result := KNAuthentication{}
	if err = json.Unmarshal(raw, &result); err != nil {
		panic(err)
	}
	return result
}

// KNSign return signed key for rebalance permission
func (self KNAuthentication) KNSign(msg string) string {
	mac := hmac.New(sha512.New, []byte(self.KNSecret))
	if _, err := mac.Write([]byte(msg)); err != nil {
		log.Printf("Encode message error: %s", err.Error())
	}
	return ethereum.Bytes2Hex(mac.Sum(nil))
}

func (self KNAuthentication) knReadonlySign(msg string) string {
	mac := hmac.New(sha512.New, []byte(self.KNReadOnly))
	if _, err := mac.Write([]byte(msg)); err != nil {
		log.Printf("Encode message error: %s", err.Error())
	}
	return ethereum.Bytes2Hex(mac.Sum(nil))
}

func (self KNAuthentication) knConfigurationSign(msg string) string {
	mac := hmac.New(sha512.New, []byte(self.KNConfiguration))
	if _, err := mac.Write([]byte(msg)); err != nil {
		log.Printf("Encode message error: %s", err.Error())
	}
	return ethereum.Bytes2Hex(mac.Sum(nil))
}

func (self KNAuthentication) knConfirmConfSign(msg string) string {
	mac := hmac.New(sha512.New, []byte(self.KNConfirmConf))
	if _, err := mac.Write([]byte(msg)); err != nil {
		log.Printf("Encode message error: %s", err.Error())
	}
	return ethereum.Bytes2Hex(mac.Sum(nil))
}

func (self KNAuthentication) GetPermission(signed string, message string) []Permission {
	result := []Permission{}
	rebalanceSigned := self.KNSign(message)
	if signed == rebalanceSigned {
		result = append(result, RebalancePermission)
	}
	readonlySigned := self.knReadonlySign(message)
	if signed == readonlySigned {
		result = append(result, ReadOnlyPermission)
	}
	configureSigned := self.knConfigurationSign(message)
	if signed == configureSigned {
		result = append(result, ConfigurePermission)
	}
	confirmConfSigned := self.knConfirmConfSign(message)
	if signed == confirmConfSigned {
		result = append(result, ConfirmConfPermission)
	}
	return result
}
