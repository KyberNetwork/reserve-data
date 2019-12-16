package binance

import (
	"crypto/hmac"
	"crypto/sha256"
	"errors"

	ethereum "github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

type Signer struct {
	Key    string `json:"binance_key"`
	Secret string `json:"binance_secret"`
}

func (s Signer) GetKey() string {
	return s.Key
}

func (s Signer) Sign(msg string) string {
	mac := hmac.New(sha256.New, []byte(s.Secret))
	if _, err := mac.Write([]byte(msg)); err != nil {
		zap.S().Panic(err)
	}
	result := ethereum.Bytes2Hex(mac.Sum(nil))
	return result
}

func NewSigner(key, secret string) (*Signer, error) {
	if key == "" || secret == "" {
		return nil, errors.New("key and secret must not empty")
	}
	return &Signer{key, secret}, nil
}
