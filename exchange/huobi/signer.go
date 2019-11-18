package huobi

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
)

type Signer struct {
	Key    string `json:"huobi_key"`
	Secret string `json:"huobi_secret"`
}

func (s Signer) Sign(msg string) string {
	mac := hmac.New(sha256.New, []byte(s.Secret))
	if _, err := mac.Write([]byte(msg)); err != nil {
		log.Printf("Encode message error: %s", err.Error())
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
