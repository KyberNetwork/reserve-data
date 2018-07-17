package settings

import (
	"encoding/json"
	"io/ioutil"

	"github.com/KyberNetwork/reserve-data/common"
)

type token struct {
	Address  string `json:"address"`
	Name     string `json:"name"`
	Decimals int64  `json:"decimals"`
	Active   bool   `json:"internal use"`
	Internal bool   `json:"listed"`
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
func (setting *Settings) loadTokenFromFile(filePath string) error {
	data, err := ioutil.ReadFile(filePath)
	tokens := TokenConfig{}
	if err != nil {
		return err
	}
	if err = json.Unmarshal(data, &tokens); err != nil {
		return err
	}
	for id, t := range tokens.Tokens {
		token := common.NewToken(id, t.Name, t.Address, t.Decimals, t.Active, t.Internal, common.GetTimepoint())
		if err = setting.UpdateToken(token); err != nil {
			return err
		}
	}
	return nil
}
