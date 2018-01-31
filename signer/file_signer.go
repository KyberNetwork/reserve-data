package signer

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/json"
	"io/ioutil"
	// "math/big"
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
	KeystoreSR      string `json:"keystore_setrate_path"`
	PassphraseSR    string `json:"passphrase_setrate"`
	KeystoreD       string `json:"keystore_deposit_path"`
	PassphraseD     string `json:"passphrase_deposit"`
	KNSecret        string `json:"kn_secret"`
	KNReadOnly      string `json:"kn_readonly"`
	KNConfiguration string `json:"kn_configuration"`
	opts            map[string]*bind.TransactOpts
}

func (self FileSigner) GetAddress(transaction string) ethereum.Address {
	return self.opts[transaction].From
}

func (self FileSigner) Sign(address ethereum.Address, tx *types.Transaction, transaction string) (*types.Transaction, error) {
	return self.opts[transaction].Signer(types.HomesteadSigner{}, address, tx)
}

func (self FileSigner) GetTransactOpts(transaction string) *bind.TransactOpts {
	return self.opts[transaction]
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

func NewFileSigner(file string) *FileSigner {
	raw, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	signer := FileSigner{}
	err = json.Unmarshal(raw, &signer)
	if err != nil {
		panic(err)
	}
	keySetRateIo, err := os.Open(signer.KeystoreSR)
	if err != nil {
		panic(err)
	}
	authSetRate, err := bind.NewTransactor(keySetRateIo, signer.PassphraseSR)
	if err != nil {
		panic(err)
	}
	keyDepositIo, err := os.Open(signer.KeystoreD)
	if err != nil {
		panic(err)
	}
	authDeposit, err := bind.NewTransactor(keyDepositIo, signer.PassphraseD)
	if err != nil {
		panic(err)
	}
	// auth.GasLimit = big.NewInt(1000000)
	// auth.GasPrice = big.NewInt(35000000000)
	opts := map[string]*bind.TransactOpts{}
	opts["setrate"] = authSetRate
	opts["deposit"] = authDeposit
	signer.opts = opts
	return &signer
}
