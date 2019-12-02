package configuration

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/KyberNetwork/reserve-data/common"
)

//AddressConfigs store token configs according to env mode.
var AddressConfigs = map[string]common.AddressConfig{
	common.DevMode: {
		Reserve:         "0x63825c174ab367968EC60f061753D3bbD36A0D8F",
		Wrapper:         "0x6172AFC8c00c46E0D07ce3AF203828198194620a",
		Pricing:         "0x798AbDA6Cc246D0EDbA912092A2a3dBd3d11191B",
		Network:         "0x818E6FECD516Ecc3849DAf6845e3EC868087B755",
		InternalNetwork: "0x91a502C678605fbCe581eae053319747482276b9",
	},
	common.StagingMode: {
		Reserve:         "0x2C5a182d280EeB5824377B98CD74871f78d6b8BC",
		Wrapper:         "0x6172AFC8c00c46E0D07ce3AF203828198194620a",
		Pricing:         "0xe3E415a7a6c287a95DC68a01ff036828073fD2e6",
		Network:         "0xC14f34233071543E979F6A79AA272b0AB1B4947D",
		InternalNetwork: "0x706aBcE058DB29eB36578c463cf295F180a1Fe9C",
	},
	common.MainnetMode: {
		Reserve:         "0x63825c174ab367968EC60f061753D3bbD36A0D8F",
		Wrapper:         "0x6172AFC8c00c46E0D07ce3AF203828198194620a",
		Pricing:         "0x798AbDA6Cc246D0EDbA912092A2a3dBd3d11191B",
		Network:         "0x818E6FECD516Ecc3849DAf6845e3EC868087B755",
		InternalNetwork: "0x65bF64Ff5f51272f729BDcD7AcFB00677ced86Cd",
	},
	common.RopstenMode: {
		Reserve: "0x0FC1CF3e7DD049F7B42e6823164A64F76fC06Be0",
		Wrapper: "0x9de0a60F4A489e350cD8E3F249f4080858Af41d3",
		Pricing: "0x535DE1F5a982c2a896da62790a42723A71c0c12B",
		Network: "0x0a56d8a49E71da8d7F9C65F95063dB48A3C9560B",
	},
}

func mustGetAddressesConfig(kyberEnv string, cliAddresses common.AddressConfig) common.AddressConfig {

	if cliAddresses.Network != "" && cliAddresses.Pricing != "" && cliAddresses.Wrapper != "" && cliAddresses.Reserve != "" {
		return cliAddresses
	}

	if kyberEnv == common.SimulationMode {
		resultFromFile, err := loadAddressFromFile(simSettingPath)
		if err != nil {
			log.Panicf("cannot load address from file, err: %v", err)
		}
		return resultFromFile
	}

	log.Panicf("cannot get address config for mode %s", kyberEnv)
	return common.AddressConfig{}
}

func loadAddressFromFile(path string) (common.AddressConfig, error) {
	data, err := ioutil.ReadFile(path)
	addrs := common.AddressConfig{}

	if err != nil {
		return addrs, err
	}
	err = json.Unmarshal(data, &addrs)
	return addrs, err
}
