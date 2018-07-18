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
//go:generate stringer -type=ExchangeName -linecomment
type ExchangeName int

const (
	Binance        ExchangeName = iota //binance
	Bittrex                            //bittrex
	Huobi                              //huobi
	StableExchange                     //stable_exchange
)
const exchangeEnv string = "KYBER_EXCHANGES"

type ExchangeFeesConfig struct {
	Exchanges map[string]common.ExchangeFees `json:"exchanges"`
}

var exchangeNameValue = map[string]ExchangeName{
	"binance":         Binance,
	"bittrex":         Bittrex,
	"huobi":           Huobi,
	"stable_exchange": StableExchange,
}

func RunningExchanges() []string {
	exchangesStr, ok := os.LookupEnv(exchangeEnv)
	if !ok {
		log.Print("WARNING: core is running without exchange")
		return nil
	}
	exchanges := strings.Split(exchangesStr, ",")
	return exchanges
}

func ExchangeTypeValues() map[string]ExchangeName {
	return exchangeNameValue
}

type ExchangeSetting struct {
	Storage ExchangeStorage
}

func NewExchangeSetting(exchangeStorage ExchangeStorage) (*ExchangeSetting, error) {
	return &ExchangeSetting{exchangeStorage}, nil
}

func (setting *Settings) loadFeeFromFile(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	var exFeeConfig ExchangeFeesConfig
	if err = json.Unmarshal(data, &exFeeConfig); err != nil {
		return err
	}
	runningExs := RunningExchanges()

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
			//multiply all Funding fee by 2 to avoid fee increasing from exchanges
			for tokenID, value := range exFee.Funding.Deposit {
				exFee.Funding.Deposit[tokenID] = value * 2
			}
			for tokenID, value := range exFee.Funding.Withdraw {
				exFee.Funding.Withdraw[tokenID] = value * 2
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

func (setting *Settings) loadMinDepositFromFile(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	var exMinDepositConfig ExchangesMinDepositConfig

	if err = json.Unmarshal(data, &exMinDepositConfig); err != nil {
		return err
	}
	runningExs := RunningExchanges()
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
			//multiply all minimum deposit by 2 to avoid min deposit increasing from Exchange
			for token, value := range minDepo {
				minDepo[token] = value * 2
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

func (setting *Settings) loadDepositAddressFromFile(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	var exAddressConfig AddressDepositConfig
	if err = json.Unmarshal(data, &exAddressConfig); err != nil {
		return err
	}
	runningExs := RunningExchanges()
	for _, ex := range runningExs {
		//Check if the exchange is in current code deployment.
		exName, ok := exchangeNameValue[ex]
		if !ok {
			return fmt.Errorf("Exchange %s is in KYBER_EXCHANGES, but not avail in current deployment", ex)
		}
		//Check if the current database has a record for such exchange
		if _, err := setting.Exchange.Storage.GetDepositAddresses(exName); err != nil {
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

func convertToAddressMap(data exchangeDepositAddress) common.ExchangeAddresses {
	result := make(common.ExchangeAddresses)
	for token, addrStr := range data {
		result[token] = ethereum.HexToAddress(addrStr)
	}
	return result
}

func (setting *Settings) handleEmptyExchangeInfo() error {
	runningExs := RunningExchanges()
	for _, ex := range runningExs {
		exName, ok := exchangeNameValue[ex]
		if !ok {
			return fmt.Errorf("Exchange %s is in KYBER_EXCHANGES, but not avail in current deployment", ex)
		}
		if _, err := setting.Exchange.Storage.GetExchangeInfo(exName); err != nil {
			log.Printf("Exchange %s is in KYBER_EXCHANGES but can't load its exchangeInfo in Database (%s). attempt to init it", exName.String(), err.Error())
			exInfo, err := setting.NewExchangeInfo(exName)
			if err != nil {
				return err
			}
			if err = setting.Exchange.Storage.StoreExchangeInfo(exName, exInfo); err != nil {
				return err
			}
		}
	}
	return nil
}

func (setting *Settings) NewExchangeInfo(exName ExchangeName) (common.ExchangeInfo, error) {
	result := common.NewExchangeInfo()
	addrs, err := setting.GetDepositAddresses(exName)
	if err != nil {
		return result, err
	}
	for tokenID := range addrs {
		if tokenID != "ETH" {
			token, err := setting.GetTokenByID(tokenID)
			if err != nil {
				log.Printf("WARNING: can not find token %s (%s). This will skip preparing its exchange info", tokenID, err)
				continue
			}
			if token.Internal == false {
				log.Printf("INFO: Token %s is external. This will skip preparing its exchange info", tokenID)
				continue
			}
			pairID := common.NewTokenPairID(tokenID, "ETH")
			result[pairID] = common.ExchangePrecisionLimit{}
		}
	}
	return result, nil
}
