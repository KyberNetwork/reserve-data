package settings

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data/common"
)

// ExchangeName is the name of exchanges of which core will use to rebalance.
//go:generate stringer -type=ExchangeName -linecomment
type ExchangeName int

const (
	//Binance is the enumerated key for binance
	Binance ExchangeName = iota //binance
	//Bittrex is the enumerated key for bittrex (deprecated)
	Bittrex //bittrex
	//Huobi is the enumerated key for huobi
	Huobi //huobi
	//StableExchange is the enumerated key for stable_exchange
	StableExchange //stable_exchange
)
const exchangeEnv string = "KYBER_EXCHANGES"

//ErrExchangeRecordNotFound will be return on empty db query
var ErrExchangeRecordNotFound = errors.New("exchange record not found")

var exchangeNameValue = map[string]ExchangeName{
	"binance":         Binance,
	"bittrex":         Bittrex,
	"huobi":           Huobi,
	"stable_exchange": StableExchange,
}

// RunningExchanges get the exchangeEnvironment params and return the list of exchanges ID for the current run
// It returns empty string slice if the ENV is empty string or not found
// DO NOT CALL this once httpserver has ran.
func RunningExchanges() []string {
	exchangesStr, ok := os.LookupEnv(exchangeEnv)
	if (!ok) || (len(exchangesStr) == 0) {
		zap.S().Warnf("WARNING: core is running without exchange")
		return nil
	}
	exchanges := strings.Split(exchangesStr, ",")
	return exchanges
}

//ExchangeTypeValues return exchange Name value config
func ExchangeTypeValues() map[string]ExchangeName {
	return exchangeNameValue
}

//ExchangeSetting is the struct to implement exchange related setting
type ExchangeSetting struct {
	Storage ExchangeStorage
}

//NewExchangeSetting return a new exchange setting
func NewExchangeSetting(exchangeStorage ExchangeStorage) (*ExchangeSetting, error) {
	return &ExchangeSetting{exchangeStorage}, nil
}

func (s *Settings) savePreconfigFee(exFeeConfig map[string]common.ExchangeFees) error {
	runningExs := RunningExchanges()

	for _, ex := range runningExs {
		//Check if the exchange is in current code deployment.
		exName, ok := exchangeNameValue[ex]
		if !ok {
			return fmt.Errorf("exchange %s is in KYBER_EXCHANGES, but not avail in current deployment", ex)
		}
		//Check if the current database has a record for such exchange
		if _, err := s.Exchange.Storage.GetFee(exName); err != nil {
			s.l.Warnf("Exchange %s is in KYBER_EXCHANGES but can't load fee in Database (%+v). atempt to load it from config file", exName.String(), err)
			//Check if the config file has config for such exchange
			exFee, ok := exFeeConfig[ex]
			if !ok {
				s.l.Warnf("Warning: Exchange %s is in KYBER_EXCHANGES, but not avail in Fee config file.", ex)
				continue
			}
			//multiply all Funding fee by 2 to avoid fee increasing from exchanges
			for tokenID, value := range exFee.Funding.Deposit {
				exFee.Funding.Deposit[tokenID] = value * 2
			}
			for tokenID, value := range exFee.Funding.Withdraw {
				exFee.Funding.Withdraw[tokenID] = value * 2
			}
			//version =1 means it is init from config file
			if err = s.Exchange.Storage.StoreFee(exName, exFee, 1); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Settings) savePrecofigMinDeposit(exMinDepositConfig map[string]common.ExchangesMinDeposit) error {

	runningExs := RunningExchanges()
	for _, ex := range runningExs {
		//Check if the exchange is in current code deployment.
		exName, ok := exchangeNameValue[ex]
		if !ok {
			return fmt.Errorf("exchange %s is in KYBER_EXCHANGES, but not avail in current deployment", ex)
		}
		//Check if the current database has a record for such exchange
		if _, err := s.Exchange.Storage.GetMinDeposit(exName); err != nil {
			s.l.Warnf("Exchange %s is in KYBER_EXCHANGES but can't load MinDeposit in Database (%s). "+
				"atempt to load it from config file", exName.String(), err)
			//Check if the config file has config for such exchange
			minDepo, ok := exMinDepositConfig[ex]
			if !ok {
				s.l.Warnf("Warning: Exchange %s is in KYBER_EXCHANGES, but not avail in MinDepositconfig file", exName.String())
				continue
			}
			//multiply all minimum deposit by 2 to avoid min deposit increasing from Exchange
			for token, value := range minDepo {
				minDepo[token] = value * 2
			}
			//version =1 means it is init from config file
			if err = s.Exchange.Storage.StoreMinDeposit(exName, minDepo, 1); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Settings) savePreconfigExchangeDepositAddress(data map[common.ExchangeID]common.ExchangeAddresses) error {
	const (
		version = 1
	)
	runningExs := RunningExchanges()
	for _, ex := range runningExs {
		//Check if the exchange is in current code deployment.
		exName, ok := exchangeNameValue[ex]
		if !ok {
			return fmt.Errorf("exchange %s is in KYBER_EXCHANGES, but not avail in current deployment", ex)
		}
		//Check if the current database has a record for such exchange
		if _, err := s.Exchange.Storage.GetDepositAddresses(exName); err != nil {
			s.l.Warnf("Exchange %s is in KYBER_EXCHANGES but can't load DepositAddress in Database (%s). "+
				"attempt to load it from config file", exName.String(), err)
			//Check if the config file has config for such exchange
			exchangeAddress, ok := data[common.ExchangeID(ex)]
			if !ok {
				s.l.Warnf("Exchange %s is in KYBER_EXCHANGES, but not avail in preconfig data", ex)
				continue
			}
			//version =1 means it is init from config file
			if err = s.Exchange.Storage.StoreDepositAddress(exName, exchangeAddress, version); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Settings) handleEmptyExchangeInfo() error {
	runningExs := RunningExchanges()
	for _, ex := range runningExs {
		exName, ok := exchangeNameValue[ex]
		if !ok {
			return fmt.Errorf("exchange %s is in KYBER_EXCHANGES, but not avail in current deployment", ex)
		}
		if _, err := s.Exchange.Storage.GetExchangeInfo(exName); err != nil {
			s.l.Warnf("Exchange %s is in KYBER_EXCHANGES but can't load its exchangeInfo in Database (%s). "+
				"attempt to init it", exName.String(), err)
			exInfo, err := s.NewExchangeInfo(exName)
			if err != nil {
				return err
			}
			//version =1 means it is init from config file
			if err = s.Exchange.Storage.StoreExchangeInfo(exName, exInfo, 1); err != nil {
				return err
			}
		}
	}
	return nil
}

//NewExchangeInfo return an an ExchangeInfo
func (s *Settings) NewExchangeInfo(exName ExchangeName) (common.ExchangeInfo, error) {
	result := common.NewExchangeInfo()
	addrs, err := s.GetDepositAddresses(exName)
	if err != nil {
		return result, err
	}
	for tokenID := range addrs {
		if tokenID != "ETH" {
			token, err := s.GetTokenByID(tokenID)
			if err != nil {
				s.l.Warnf("can not find token %s (%s). This will skip preparing its exchange info", tokenID, err)
				continue
			}
			if !token.Internal {
				s.l.Warnf("INFO: Token %s is external. This will skip preparing its exchange info", tokenID)
				continue
			}
			pairID := common.NewTokenPairID(tokenID, "ETH")
			result[pairID] = common.ExchangePrecisionLimit{}
		}
	}
	return result, nil
}
