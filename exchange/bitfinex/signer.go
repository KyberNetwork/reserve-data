package bitfinex

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/json"
	"io/ioutil"

	ethereum "github.com/ethereum/go-ethereum/common"
)

type Signer struct {
	Key    string `json:"bitfinex_key"`
	Secret string `json:"bitfinex_secret"`
}

func (self Signer) GetKey() string {
	return self.Key
}

func (self Signer) Sign(msg string) string {
	mac := hmac.New(sha512.New384, []byte(self.Secret))
	mac.Write([]byte(msg))
	return ethereum.Bytes2Hex(mac.Sum(nil))
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
