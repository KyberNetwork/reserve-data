package signer

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type FileSigner struct {
	LiquiKey        string `json:"liqui_key"`
	LiquiSecret     string `json:"liqui_secret"`
	BinanceKey      string `json:"binance_key"`
	BinanceSecret   string `json:"binance_secret"`
	BittrexKey      string `json:"bittrex_key"`
	BittrexSecret   string `json:"bittrex_secret"`
	BitfinexKey     string `json:"bitfinex_key"`
	BitfinexSecret  string `json:"bitfinex_secret"`
	HuobiKey        string `json:"huobi_key"`
	HuobiSecret     string `json:"huobi_secret"`
	Keystore        string `json:"keystore_path"`
	Passphrase      string `json:"passphrase"`
	KeystoreD       string `json:"keystore_deposit_path"`
	PassphraseD     string `json:"passphrase_deposit"`
	KNSecret        string `json:"kn_secret"`
	KNReadOnly      string `json:"kn_readonly"`
	KNConfiguration string `json:"kn_configuration"`
	KNConfirmConf   string `json:"kn_confirm_configuration"`
	KeystoreI       string `json:"keystore_intermediator_path"`
	PassphraseI     string `json:"passphrase_intermediate_account"`
	opts            *bind.TransactOpts
}

func (self FileSigner) GetAddress() ethereum.Address {
	return self.opts.From
}

func (self FileSigner) Sign(tx *types.Transaction) (*types.Transaction, error) {
	return self.opts.Signer(types.HomesteadSigner{}, self.GetAddress(), tx)
}

func (self FileSigner) GetTransactOpts() *bind.TransactOpts {
	return self.opts
}

func (self FileSigner) GetLiquiKey() string {
	return self.LiquiKey
}

func (self FileSigner) GetBitfinexKey() string {
	return self.BitfinexKey
}

func (self FileSigner) GetBittrexKey() string {
	return self.BittrexKey
}

func (self FileSigner) GetBinanceKey() string {
	return self.BinanceKey
}

func (self FileSigner) GetHuobiKey() string {
	return self.HuobiKey
}

func (self FileSigner) KNSign(msg string) string {
	log.Printf("KN secret: %s", self.KNSecret)
	mac := hmac.New(sha512.New, []byte(self.KNSecret))
	mac.Write([]byte(msg))
	return ethereum.Bytes2Hex(mac.Sum(nil))
}

func (self FileSigner) KNReadonlySign(msg string) string {
	mac := hmac.New(sha512.New, []byte(self.KNReadOnly))
	mac.Write([]byte(msg))
	return ethereum.Bytes2Hex(mac.Sum(nil))
}

func (self FileSigner) LiquiSign(msg string) string {
	mac := hmac.New(sha512.New, []byte(self.LiquiSecret))
	mac.Write([]byte(msg))
	return ethereum.Bytes2Hex(mac.Sum(nil))
}

func (self FileSigner) BitfinexSign(msg string) string {
	mac := hmac.New(sha512.New384, []byte(self.BitfinexSecret))
	mac.Write([]byte(msg))
	return ethereum.Bytes2Hex(mac.Sum(nil))
}

func (self FileSigner) BittrexSign(msg string) string {
	mac := hmac.New(sha512.New, []byte(self.BittrexSecret))
	mac.Write([]byte(msg))
	return ethereum.Bytes2Hex(mac.Sum(nil))
}

func (self FileSigner) BinanceSign(msg string) string {
	mac := hmac.New(sha256.New, []byte(self.BinanceSecret))
	mac.Write([]byte(msg))
	result := ethereum.Bytes2Hex(mac.Sum(nil))
	return result
}

func (self FileSigner) HuobiSign(msg string) string {
	mac := hmac.New(sha256.New, []byte(self.HuobiSecret))
	mac.Write([]byte(msg))
	result := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return result
}

func NewFileSigner(file string) (*FileSigner, *FileSigner, *FileSigner) {
	raw, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	signer := FileSigner{}
	err = json.Unmarshal(raw, &signer)
	if err != nil {
		panic(err)
	}
	depositSigner := signer
	keyio, err := os.Open(signer.Keystore)
	if err != nil {
		panic(err)
	}
	auth, err := bind.NewTransactor(keyio, signer.Passphrase)
	if err != nil {
		panic(err)
	}
	keyDIo, err := os.Open(signer.KeystoreD)
	if err != nil {
		panic(err)
	}
	authD, err := bind.NewTransactor(keyDIo, signer.PassphraseD)
	if err != nil {
		panic(err)
	}
	intermediateAccountSigner := signer
	keyIio, err := os.Open(signer.KeystoreI)
	if err != nil {
		panic(err)
	}
	authI, err := bind.NewTransactor(keyIio, signer.PassphraseI)
	if err != nil {
		panic(err)
	}
	// auth.GasLimit = big.NewInt(1000000)
	// auth.GasPrice = big.NewInt(35000000000)
	signer.opts = auth
	depositSigner.opts = authD
	intermediateAccountSigner.opts = authI
	return &signer, &depositSigner, &intermediateAccountSigner
}
