package settings

import (
	"log"

	"github.com/KyberNetwork/reserve-data/common"
)

// Settings contains shared common settings between all components.
type Settings struct {
	Tokens   *TokenSetting
	Address  *AddressSetting
	Exchange *ExchangeSetting
}

// WithHandleEmptyToken will load the token settings from default file if the
// database is empty.
func WithHandleEmptyToken(data map[string]common.Token) SettingOption {
	return func(setting *Settings) {
		allToks, err := setting.GetAllTokens()
		if err != nil || len(allToks) < 1 {
			if err != nil {
				log.Printf("Setting Init: Token DB is faulty (%s), attempt to load token from file", err.Error())
			} else {
				log.Printf("Setting Init: Token DB is empty, attempt to load token from file")
			}
			if err = setting.savePreconfigToken(data); err != nil {
				log.Printf("Setting Init: Can not load Token from file: %s, Token DB is needed to be updated manually", err.Error())
			}
		}
	}
}

// WithHandleEmptyFee will load the Fee settings from default file
// if the fee database is empty. It will mutiply the Funding fee value by 2
func WithHandleEmptyFee(feeConfig map[string]common.ExchangeFees) SettingOption {
	return func(setting *Settings) {
		if err := setting.savePreconfigFee(feeConfig); err != nil {
			log.Printf("WARNING: Setting Init: cannot load Fee from file: %s, Fee is needed to be updated manually", err.Error())
		}
	}
}

// WithHandleEmptyMinDeposit will load the MinDeposit setting from fefault file
// if the Mindeposit database is empty. It will mutiply the MinDeposit value by 2
func WithHandleEmptyMinDeposit(data map[string]common.ExchangesMinDeposit) SettingOption {
	return func(setting *Settings) {
		if err := setting.savePrecofigMinDeposit(data); err != nil {
			log.Printf("WARNING: Setting Init: cannot load MinDeposit from file: %s, Fee is needed to be updated manually", err.Error())
		}
	}
}

// WithHandleEmptyDepositAddress will load the MinDeposit setting from fefault file
// if the DepositAddress database is empty
func WithHandleEmptyDepositAddress(data map[common.ExchangeID]common.ExchangeAddresses) SettingOption {
	return func(setting *Settings) {
		if err := setting.savePreconfigExchangeDepositAddress(data); err != nil {
			log.Printf("WARNING: Setting Init: cannot load DepositAddress from file: %s, Fee is needed to be updated manually", err.Error())
		}
	}
}

// WithHandleEmptyExchangeInfo will create a map of TokenPairs from token deposit address to an empty ExchangePrecisionLimit
// it will return error if occur
func WithHandleEmptyExchangeInfo() SettingOption {
	return func(setting *Settings) {
		if err := setting.handleEmptyExchangeInfo(); err != nil {
			log.Panicf("Setting Init: cannot init Exchange info %s, this will stop the core function", err.Error())
		}
	}
}

// SettingOption sets the initialization behavior of the Settings instance.
type SettingOption func(s *Settings)

// NewSetting create setting object from its component, and handle if the setting database is empty
// returns a pointer to setting object from which all core setting can be read and write to; and error if occurs.
func NewSetting(token *TokenSetting, address *AddressSetting, exchange *ExchangeSetting, options ...SettingOption) (*Settings, error) {
	setting := &Settings{
		Tokens:   token,
		Address:  address,
		Exchange: exchange,
	}
	for _, option := range options {
		option(setting)
	}
	return setting, nil
}
