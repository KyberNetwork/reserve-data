package binance

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"io/ioutil"

	ethereum "github.com/ethereum/go-ethereum/common"
)

// Signer for binance
type Signer struct {
	Key    string `json:"binance_key"`
	Secret string `json:"binance_secret"`
}

// GetKey return binance key
func (s Signer) GetKey() string {
	return s.Key
}

// Sign a message
func (s Signer) Sign(msg string) string {
	mac := hmac.New(sha256.New, []byte(s.Secret))
	if _, err := mac.Write([]byte(msg)); err != nil {
		panic(err) // should never happen
	}
	result := ethereum.Bytes2Hex(mac.Sum(nil))
	return result
}

// NewSigner return binance signer
func NewSigner(key, secret string) Signer {
	return Signer{key, secret}
}

// NewSignerFromFile return signer for binance from file
func NewSignerFromFile(path string) Signer {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	signer := Signer{}
	err = json.Unmarshal(raw, &signer)
	if err != nil {
		panic(err)
	}
	return Signer{
		Key:    signer.Key,
		Secret: signer.Secret,
	}
}
