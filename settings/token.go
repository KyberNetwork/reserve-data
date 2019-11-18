package settings

import (
	"errors"
	"fmt"
	"log"

	"github.com/KyberNetwork/reserve-data/common"
	ethereum "github.com/ethereum/go-ethereum/common"
)

// ErrTokenNotFound is the error returned for get operation where the
// token is not found in database.
var ErrTokenNotFound = errors.New("token not found")

func (s *Settings) GetAllTokens() ([]common.Token, error) {
	return s.Tokens.Storage.GetAllTokens()
}

func (s *Settings) GetActiveTokens() ([]common.Token, error) {
	return s.Tokens.Storage.GetActiveTokens()
}

func (s *Settings) GetInternalTokens() ([]common.Token, error) {
	return s.Tokens.Storage.GetInternalTokens()
}

func (s *Settings) GetExternalTokens() ([]common.Token, error) {
	return s.Tokens.Storage.GetExternalTokens()
}

func (s *Settings) GetTokenByID(id string) (common.Token, error) {
	return s.Tokens.Storage.GetTokenByID(id)
}

func (s *Settings) GetActiveTokenByID(id string) (common.Token, error) {
	return s.Tokens.Storage.GetActiveTokenByID(id)
}

func (s *Settings) GetInternalTokenByID(id string) (common.Token, error) {
	return s.Tokens.Storage.GetInternalTokenByID(id)
}

func (s *Settings) GetExternalTokenByID(id string) (common.Token, error) {
	return s.Tokens.Storage.GetExternalTokenByID(id)
}

func (s *Settings) GetTokenByAddress(addr ethereum.Address) (common.Token, error) {
	return s.Tokens.Storage.GetTokenByAddress(addr)
}

func (s *Settings) GetActiveTokenByAddress(addr ethereum.Address) (common.Token, error) {
	return s.Tokens.Storage.GetActiveTokenByAddress(addr)
}

func (s *Settings) GetInternalTokenByAddress(addr ethereum.Address) (common.Token, error) {
	return s.Tokens.Storage.GetInternalTokenByAddress(addr)
}

func (s *Settings) GetExternalTokenByAddress(addr ethereum.Address) (common.Token, error) {
	return s.Tokens.Storage.GetExternalTokenByAddress(addr)
}

func (s *Settings) ETHToken() common.Token {
	eth, err := s.Tokens.Storage.GetInternalTokenByID("ETH")
	if err != nil {
		log.Panicf("There is no ETH token in token DB, this should not happen (%s)", err)
	}
	return eth
}

func (s *Settings) NewTokenPairFromID(base, quote string) (common.TokenPair, error) {
	bToken, err1 := s.GetInternalTokenByID(base)
	qToken, err2 := s.GetInternalTokenByID(quote)
	if err1 != nil || err2 != nil {
		return common.TokenPair{}, fmt.Errorf("%s or %s is not supported", base, quote)
	}
	return common.TokenPair{Base: bToken, Quote: qToken}, nil
}

func (s *Settings) MustCreateTokenPair(base, quote string) common.TokenPair {
	pair, err := s.NewTokenPairFromID(base, quote)
	if err != nil {
		panic(err)
	}
	return pair
}

func (s *Settings) UpdateToken(t common.Token, timestamp uint64) error {
	if timestamp == 0 {
		timestamp = common.GetTimepoint()
	}
	return s.Tokens.Storage.UpdateToken(t, timestamp)
}

func (s *Settings) ApplyTokenWithExchangeSetting(tokens []common.Token, exSetting map[ExchangeName]*common.ExchangeSetting, timestamp uint64) error {
	if timestamp == 0 {
		timestamp = common.GetTimepoint()
	}
	exStatus, err := s.GetExchangeStatus()
	if err != nil {
		return err
	}
	var availExs []ExchangeName
	for name, status := range exStatus {
		if status.Status {
			availExs = append(availExs, exchangeNameValue[name])
		}
	}
	return s.Tokens.Storage.UpdateTokenWithExchangeSetting(tokens, exSetting, availExs, timestamp)
}

func (s *Settings) UpdatePendingTokenUpdates(tokenUpdates map[string]common.TokenUpdate) error {
	return s.Tokens.Storage.StorePendingTokenUpdates(tokenUpdates)
}

func (s *Settings) GetPendingTokenUpdates() (map[string]common.TokenUpdate, error) {
	return s.Tokens.Storage.GetPendingTokenUpdates()
}

func (s *Settings) RemovePendingTokenUpdates() error {
	return s.Tokens.Storage.RemovePendingTokenUpdates()
}

func (s *Settings) GetTokenVersion() (uint64, error) {
	return s.Tokens.Storage.GetTokenVersion()
}
