package configuration

import (
	"log"

	"github.com/KyberNetwork/reserve-data/common"
)

const (
	simSettingPath = "shared/deployment_dev.json"
)

var ethTokenConfig = map[string]common.Token{
	"ETH": common.NewToken("ETH", "Ethereum", "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee", 18, true, true, common.GetTimepoint()),
}

//TokenConfigs store token configuration for each modes
//Sim mode require special care.
var TokenConfigs = map[string]map[string]common.Token{
	common.DevMode:        ethTokenConfig,
	common.StagingMode:    ethTokenConfig,
	common.ProductionMode: ethTokenConfig,
	common.RopstenMode:    ethTokenConfig,
	common.SimulationMode: ethTokenConfig,
}

func mustGetTokenConfig(kyberEnv string) map[string]common.Token {
	result, avail := TokenConfigs[kyberEnv]
	if avail {
		return result
	}
	if kyberEnv == common.ProductionMode {
		return TokenConfigs[common.MainnetMode]
	}
	log.Panicf("cannot get token config for mode %s", kyberEnv)
	return nil
}
