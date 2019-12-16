package huobi

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"

	"go.uber.org/zap"
)

type Signer struct {
	Key    string `json:"huobi_key"`
	Secret string `json:"huobi_secret"`
}

func (s Signer) Sign(msg string) string {
	mac := hmac.New(sha256.New, []byte(s.Secret))
	if _, err := mac.Write([]byte(msg)); err != nil {
		zap.S().Panic(err)
	}
	result := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return result
}

func (s Signer) GetKey() string {
	return s.Key
}

func NewSigner(key, secret string) *Signer {
	return &Signer{key, secret}
}
