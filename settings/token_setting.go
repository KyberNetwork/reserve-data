package settings

import (
	"github.com/KyberNetwork/reserve-data/common"
)

//TokenSetting is the object to implement token setting interface
type TokenSetting struct {
	Storage TokenStorage
}

//NewTokenSetting return a TokenSetting instance
func NewTokenSetting(tokenStorage TokenStorage) (*TokenSetting, error) {
	tokenSetting := TokenSetting{tokenStorage}
	return &tokenSetting, nil

}

func (s *Settings) savePreconfigToken(data map[string]common.Token) error {
	const (
		version = 1
	)
	for _, t := range data {
		if err := s.UpdateToken(t, version); err != nil {
			return err
		}
	}
	return nil
}
