package binance

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"io/ioutil"

	ethereum "github.com/ethereum/go-ethereum/common"
)

type Signer struct {
	Key    string `json:"binance_key"`
	Secret string `json:"binance_secret"`
}

func (self Signer) GetKey() string {
	return self.Key
}

func (self Signer) Sign(msg string) string {
	mac := hmac.New(sha256.New, []byte(self.Secret))
	mac.Write([]byte(msg))
	result := ethereum.Bytes2Hex(mac.Sum(nil))
	return result
}

func NewSigner(key, secret string) *Signer {
	return &Signer{key, secret}
}

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
	return signer
}
