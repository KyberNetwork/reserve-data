package settings

import (
	"encoding/json"

	"github.com/KyberNetwork/reserve-data/common"
)

type token struct {
	Address  string `json:"address"`
	Name     string `json:"name"`
	Decimals int64  `json:"decimals"`
	Internal bool   `json:"internal use"`
	Active   bool   `json:"listed"`
}

type TokenConfig struct {
	Tokens map[string]token `json:"tokens"`
}

type TokenSetting struct {
	Storage TokenStorage
}

func NewTokenSetting(tokenStorage TokenStorage) (*TokenSetting, error) {
	tokenSetting := TokenSetting{tokenStorage}
	return &tokenSetting, nil

}
func (setting *Settings) loadTokenFromString(data string) error {
	tokens := TokenConfig{}
	if err := json.Unmarshal([]byte(data), &tokens); err != nil {
		return err
	}
	for id, t := range tokens.Tokens {
		token := common.NewToken(id, t.Name, t.Address, t.Decimals, t.Active, t.Internal, common.GetTimepoint())
		if err := setting.UpdateToken(token, 1); err != nil {
			return err
		}
	}
	return nil
}
