package configuration

import (
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/config"
)

func mustGetTokenConfig(ac config.AppConfig) map[string]common.Token {
	result := make(map[string]common.Token)
	for id, t := range ac.TokenSet {
		token := common.NewToken(id, t.Name, t.Address, t.Decimals, t.Active, t.Internal, common.GetTimepoint())
		result[id] = token
	}
	return result
}
