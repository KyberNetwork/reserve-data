package settings

import (
	"log"
)

type Settings struct {
	Tokens   *TokenSetting
	Address  *AddressSetting
	Exchange *ExchangeSetting
}

// HandleEmptyToken will load the token settings from default file if the
// database is empty.
func WithHandleEmptyToken(pathJSON string) SettingOption {
	return func(setting *Settings) {
		allToks, err := setting.GetAllTokens()
		if err != nil || len(allToks) < 1 {
			if err != nil {
				log.Printf("Setting Init: Token DB is faulty (%s), attempt to load token from file", err)
			} else {
				log.Printf("Setting Init: Token DB is empty, attempt to load token from file")
			}
			if err = setting.LoadTokenFromFile(pathJSON); err != nil {
				log.Printf("Setting Init: Can not load Token from file: %s, Token DB is needed to be updated manually", err)
			}
		}
	}
}

// HandleEmptyAddress will load the address settings from default file if the
// database is empty.
func WithHandleEmptyAddress(pathJSON string) SettingOption {
	return func(setting *Settings) {
		addressCount, err := setting.Address.Storage.CountAddress()
		if addressCount == 0 || err != nil {
			if err != nil {
				log.Printf("Setting Init: Address DB is faulty (%s), attempt to load Address from file", err)
			} else {
				log.Printf("Setting Init: Address DB is empty, attempt to load address from file")
			}
			if err = setting.LoadAddressFromFile(pathJSON); err != nil {
				log.Printf("Setting Init: Can not load Address from file: %s, address DB is needed to be updated manually", err)
			}
		}
	}
}

// WithHandleEmptyFee will load the Fee settings from default file
// if the fee database is empty
func WithHandleEmptyFee(pathJSON string) SettingOption {
	return func(setting *Settings) {
		if err := setting.LoadFeeFromFile(pathJSON); err != nil {
			log.Printf("WARNING: Setting Init: cannot load Fee from file: %s, Fee is needed to be updated manually", err)
		}
	}
}

// WithHandleEmptyMinDeposit will load the MinDeposit setting from fefault file
// if the Mindeposit database is empty
func WithHandleEmptyMinDeposit(pathJSON string) SettingOption {
	return func(setting *Settings) {
		if err := setting.LoadMinDepositFromFile(pathJSON); err != nil {
			log.Printf("WARNING: Setting Init: cannot load MinDeposit from file: %s, Fee is needed to be updated manually", err)
		}
	}
}

// WithHandleEmptyDepositAddress will load the MinDeposit setting from fefault file
// if the DepositAddress database is empty
func WithHandleEmptyDepositAddress(pathJSON string) SettingOption {
	return func(setting *Settings) {
		if err := setting.LoadDepositAddressFromFile(pathJSON); err != nil {
			log.Printf("WARNING: Setting Init: cannot load DepositAddress from file: %s, Fee is needed to be updated manually", err)
		}
	}
}

// WithHandleEmptyTokenPairs will create TokenPairs from list of Token DepositAddress with each exchange
// if that exchange TokenPairs is empty
func WithHandleEmptyTokenPairs() SettingOption {
	return func(setting *Settings) {
		if err := setting.HandleEmptyTokenPairs(); err != nil {
			log.Printf("WARNING: Setting Init: cannot init TokenPairs %s, Token Pair is needed to be updated manualluy", err.Error())
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
