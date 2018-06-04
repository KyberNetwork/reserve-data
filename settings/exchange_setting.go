package settings

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/KyberNetwork/reserve-data/common"
	ethereum "github.com/ethereum/go-ethereum/common"
)

// ExchangeName is the name of exchanges of which core will use to rebalance.
//go:generate stringer -type=ExchangeName
type ExchangeName int

const (
	Binance ExchangeName = iota
	Bittrex
	Huobi
	StableExchange
)
const exchange_env string = "KYBER_EXCHANGES"

type ExchangeFeesConfig struct {
	Exchanges map[string]common.ExchangeFees `json:"exchanges"`
}

var exchangeNameValue = map[string]ExchangeName{
	"binance":         Binance,
	"bittrex":         Bittrex,
	"huobi":           Huobi,
	"stable_exchange": StableExchange,
}

func RunningExchange() []string {
	exchangesStr, ok := os.LookupEnv(exchange_env)
	if !ok {
		log.Print("WARNING: core is running without exchange")
		return nil
	}
	exchanges := strings.Split(exchangesStr, ",")
	return exchanges
}

func ExchangTypeValues() map[string]ExchangeName {
	return exchangeNameValue
}

type ExchangeSetting struct {
	Storage ExchangeStorage
}

func NewExchangeSetting(exchangeStorage ExchangeStorage) (*ExchangeSetting, error) {
	return &ExchangeSetting{exchangeStorage}, nil
}

func (setting *Settings) LoadFeeFromFile(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	var exFeeConfig ExchangeFeesConfig
	if err = json.Unmarshal(data, &exFeeConfig); err != nil {
		return err
	}
	runningExs := RunningExchange()

	for _, ex := range runningExs {
		//Check if the exchange is in current code deployment.
		exName, ok := exchangeNameValue[ex]
		if !ok {
			return fmt.Errorf("Exchange %s is in KYBER_EXCHANGES, but not avail in current deployment", ex)
		}
		//Check if the current database has a record for such exchange
		if _, err := setting.Exchange.Storage.GetFee(exName); err != nil {
			log.Printf("Exchange %s is in KYBER_EXCHANGES but can't load fee in Database (%s). atempt to load it from config file", exName.String(), err.Error())
			//Check if the config file has config for such exchange
			exFee, ok := exFeeConfig.Exchanges[ex]
			if !ok {
				log.Printf("Warning: Exchange %s is in KYBER_EXCHANGES, but not avail in Fee config file.", ex)
				continue
			}
			if err = setting.Exchange.Storage.StoreFee(exName, exFee); err != nil {
				return err
			}
		}
	}
	return nil
}

type ExchangesMinDepositConfig struct {
	Exchanges map[string]common.ExchangesMinDeposit `json:"exchanges"`
}

func (setting *Settings) LoadMinDepositFromFile(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	var exMinDepositConfig ExchangesMinDepositConfig

	if err = json.Unmarshal(data, &exMinDepositConfig); err != nil {
		return err
	}
	runningExs := RunningExchange()
	for _, ex := range runningExs {
		//Check if the exchange is in current code deployment.
		exName, ok := exchangeNameValue[ex]
		if !ok {
			return fmt.Errorf("Exchange %s is in KYBER_EXCHANGES, but not avail in current deployment", ex)
		}
		//Check if the current database has a record for such exchange
		if _, err := setting.Exchange.Storage.GetMinDeposit(exName); err != nil {
			log.Printf("Exchange %s is in KYBER_EXCHANGES but can't load MinDeposit in Database (%s). atempt to load it from config file", exName.String(), err.Error())
			//Check if the config file has config for such exchange
			minDepo, ok := exMinDepositConfig.Exchanges[ex]
			if !ok {
				log.Printf("Warning: Exchange %s is in KYBER_EXCHANGES, but not avail in MinDepositconfig file", exName.String())
				continue
			}
			if err = setting.Exchange.Storage.StoreMinDeposit(exName, minDepo); err != nil {
				return err
			}
		}
	}
	return nil
}

// exchangeDepositAddress type stores a map[tokenID]depositaddress
// it is used to read address config from a file.
type exchangeDepositAddress map[string]string

// AddressDepositConfig struct contain a map[exchangeName],
// it is used mainly to read addfress config from JSON file.
type AddressDepositConfig struct {
	Exchanges map[string]exchangeDepositAddress `json:"exchanges"`
}

func (setting *Settings) LoadDepositAddressFromFile(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	var exAddressConfig AddressDepositConfig
	if err = json.Unmarshal(data, &exAddressConfig); err != nil {
		return err
	}
	runningExs := RunningExchange()
	for _, ex := range runningExs {
		//Check if the exchange is in current code deployment.
		exName, ok := exchangeNameValue[ex]
		if !ok {
			return fmt.Errorf("Exchange %s is in KYBER_EXCHANGES, but not avail in current deployment", ex)
		}
		//Check if the current database has a record for such exchange
		if _, err := setting.Exchange.Storage.GetDepositAddress(exName); err != nil {
			log.Printf("Exchange %s is in KYBER_EXCHANGES but can't load DepositAddress in Database (%s). atempt to load it from config file", exName.String(), err.Error())
			//Check if the config file has config for such exchange
			exchangeAddressStr, ok := exAddressConfig.Exchanges[ex]
			if !ok {
				log.Printf("Warning: Exchange %s is in KYBER_EXCHANGES, but not avail in DepositAddress config file", ex)
				continue
			}
			exchangeAddresses := convertToAddressMap(exchangeAddressStr)
			if err = setting.Exchange.Storage.StoreDepositAddress(exName, exchangeAddresses); err != nil {
				return err
			}
		}
	}
	return nil
}

func (setting *Settings) HandleEmptyTokenPairs() error {
	runningExs := RunningExchange()
	for _, ex := range runningExs {
		exName, ok := exchangeNameValue[ex]
		if !ok {
			return fmt.Errorf("Exchange %s is in KYBER_EXCHANGES, but not avail in current deployment", ex)
		}
		if _, err := setting.Exchange.Storage.GetTokenPairs(exName); err != nil {
			log.Printf("Exchange %s is in KYBER_EXCHANGES but can't load TokenPairs in Database (%s). atempt to init it", exName.String(), err.Error())
			tokenPairs, err := setting.NewExchangeTokenPairs(exName)
			if err != nil {
				return err
			}
			if err = setting.Exchange.Storage.StoreTokenPairs(exName, tokenPairs); err != nil {
				return err
			}
		}
	}
	return nil
}

func (setting *Settings) NewExchangeTokenPairs(exName ExchangeName) ([]common.TokenPair, error) {
	result := []common.TokenPair{}
	addrs, err := setting.GetDepositAddress(exName)
	if err != nil {
		return result, err
	}
	for tokenID := range addrs {
		_, err := setting.GetInternalTokenByID(tokenID)
		if err != nil {
			return result, fmt.Errorf("Internal Token failed :%s", err)
		}
		if tokenID != "ETH" {
			pair := setting.MustCreateTokenPair(tokenID, "ETH")
			result = append(result, pair)
		}
	}
	return result, nil
}

func convertToAddressMap(data exchangeDepositAddress) common.ExchangeAddresses {
	result := make(common.ExchangeAddresses)
	for token, addrStr := range data {
		result[token] = ethereum.HexToAddress(addrStr)
	}
	return result
}
