package http

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"

	"github.com/KyberNetwork/reserve-stats/lib/httputil/middleware"
	ethereum "github.com/ethereum/go-ethereum/common"
)

// Authentication is the authentication layer of HTTP APIs.
type Authentication interface {
	KNSign(message string) string
	secrectKeys() middleware.SecrectKeys
	ReadPermissionCheck() (gin.HandlerFunc, error)
	RebalancePermissionCheck() (gin.HandlerFunc, error)
	ConfigurePermissionCheck() (gin.HandlerFunc, error)
	ConfirmPermissionCheck() (gin.HandlerFunc, error)


}

type KNAuthentication struct {
	KNSecret        string `json:"kn_secret"`
	KNReadOnly      string `json:"kn_readonly"`
	KNConfiguration string `json:"kn_configuration"`
	KNConfirmConf   string `json:"kn_confirm_configuration"`
}

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

func (k KNAuthentication) secrectKeys() middleware.SecrectKeys {
	return middleware.SecrectKeys{
		readOnlyPermission:    k.KNReadOnly,
		rebalancePermission:   k.KNSecret,
		configurePermission:   k.KNConfiguration,
		confirmConfPermission: k.KNConfirmConf,
	}
}

func (k KNAuthentication) ReadPermissionCheck() (gin.HandlerFunc, error) {
	return middleware.Authenticated(canReadPermissions,k.secrectKeys(),middleware.NewValidateNonceByTime())
}

func (k KNAuthentication) RebalancePermissionCheck() (gin.HandlerFunc, error) {
	return middleware.Authenticated(rebalancePermissions,k.secrectKeys(),middleware.NewValidateNonceByTime())
}

func (k KNAuthentication) ConfigurePermissionCheck() (gin.HandlerFunc, error) {
	return middleware.Authenticated(configurePermissions,k.secrectKeys(),middleware.NewValidateNonceByTime())
}

func (k KNAuthentication) ConfirmPermissionCheck() (gin.HandlerFunc, error) {
	return middleware.Authenticated(confirmPermissions,k.secrectKeys(),middleware.NewValidateNonceByTime())
}

func (self KNAuthentication) KNSign(msg string) string {
	mac := hmac.New(sha512.New, []byte(self.KNSecret))
	if _, err := mac.Write([]byte(msg)); err != nil {
		log.Printf("Encode message error: %s", err.Error())
	}
	return ethereum.Bytes2Hex(mac.Sum(nil))
}
