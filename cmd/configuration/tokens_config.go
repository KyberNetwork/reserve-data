package configuration

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/KyberNetwork/reserve-data/common"
)

const simSettingPath = "shared/deployment_dev.json"

// mustGetSimulationTokenConfig gets token configuration from a shared file.
func mustGetSimulationTokenConfig(kyberEnv string) map[string]common.Token {
	if kyberEnv != common.SimulationMode {
		return nil
	}
	result, err := loadTokenFromFile(simSettingPath)
	if err != nil {
		log.Panicf("cannot load data from pre-defined simluation setting file, err: %v", err)
	}
	return result
}

type token struct {
	Address  string `json:"address"`
	Name     string `json:"name"`
	Decimals int64  `json:"decimals"`
	Internal bool   `json:"internal use"`
	Active   bool   `json:"listed"`
}

type tokenData struct {
	Tokens map[string]token `json:"tokens"`
}

func loadTokenFromFile(filePath string) (map[string]common.Token, error) {
	var (
		result = make(map[string]common.Token)
		tokens tokenData
	)
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return result, err
	}
	if err = json.Unmarshal(data, &tokens); err != nil {
		return result, err
	}
	for id, t := range tokens.Tokens {
		token := common.NewToken(id, t.Name, t.Address, t.Decimals, t.Active, t.Internal, common.GetTimepoint())
		result[id] = token
	}
	return result, nil
}
