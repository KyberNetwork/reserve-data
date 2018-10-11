package middleware

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var errInvalidNonce = errors.New("Invalid nonce")
var errInvalidSign = errors.New("Invalid signed token")
var errAuthenticatedPermission = errors.New("secerts does not have key for permission")

// Permission define permission type
type Permission uint32

const (
	// ReadOnly permission only
	ReadOnly Permission = 1 + iota
	// ReadAndWrite permission readandwrite
	ReadAndWrite
)

// SecrectKeys map of secret key with permission
type SecrectKeys map[Permission]string
type requiredFields struct {
	Nonce int64 `form:"nonce" binding:"required"`
}

// Authenticated return middleware func for gin
func Authenticated(permissions []Permission, secrets SecrectKeys, v ValidateNonce) (gin.HandlerFunc, error) {
	for _, p := range permissions {
		if _, ok := secrets[p]; !ok {
			return nil, errAuthenticatedPermission
		}
	}
	return func(c *gin.Context) {
		err := c.Request.ParseForm()
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		var r requiredFields
		err = c.ShouldBind(&r)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		if !v.IsValid(r.Nonce) {
			c.AbortWithError(http.StatusBadRequest, errInvalidNonce)
			return
		}
		signed := c.GetHeader("signed")
		message := c.Request.Form.Encode()
		auth := checkAuth(signed, message, permissions, secrets)
		if !auth {
			c.AbortWithError(http.StatusUnauthorized, errInvalidSign)
			return
		}
		c.Next()
	}, nil
}

func checkAuth(signed string, message string, permissions []Permission, secrets SecrectKeys) bool {
	for _, p := range permissions {
		secret, _ := secrets[p]
		serverSigned := serverSign(message, secret)
		if serverSigned == signed {
			return true
		}
	}
	return false
}

func serverSign(msg string, secret string) string {
	mac := hmac.New(sha512.New, []byte(secret))
	if _, err := mac.Write([]byte(msg)); err != nil {
		log.Printf("Encode message error: %s", err.Error())
	}
	return hex.EncodeToString(mac.Sum(nil))
}

func getTimepoint() uint64 {
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	return uint64(timestamp)
}
