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
		Reserve:            "0x63825c174ab367968EC60f061753D3bbD36A0D8F",
		Network:            "0x818E6FECD516Ecc3849DAf6845e3EC868087B755",
		Wrapper:            "0x6172AFC8c00c46E0D07ce3AF203828198194620a",
		Pricing:            "0x798AbDA6Cc246D0EDbA912092A2a3dBd3d11191B",
		FeeBurner:          "0xed4f53268bfdFF39B36E8786247bA3A02Cf34B04",
		Whitelist:          "0x6e106a75d369d09a9ea1dcc16da844792aa669a3",
		ThirdPartyReserves: []string{"0x2aab2b157a03915c8a73adae735d0cf51c872f31"},
		InternalNetwork:    "0x91a502C678605fbCe581eae053319747482276b9",
	},
	common.StagingMode: {
		Reserve:   "0x2C5a182d280EeB5824377B98CD74871f78d6b8BC",
		Network:   "0xC14f34233071543E979F6A79AA272b0AB1B4947D",
		Wrapper:   "0x6172AFC8c00c46E0D07ce3AF203828198194620a",
		Pricing:   "0xe3E415a7a6c287a95DC68a01ff036828073fD2e6",
		FeeBurner: "0xd6703974Dc30155d768c058189A2936Cf7C62Da6",
		Whitelist: "0x1b8a28d2185a7ab86ae40303b39f33278cc42130",
		ThirdPartyReserves: []string{"0xe1213e46efcb8785b47ae0620a51f490f747f1da",
			"0x6f50e41885fdc44dbdf7797df0393779a9c0a3a6",
			"0x1f58a138c976ceface828fd9e0b82295e85e7c81"},
		InternalNetwork: "0x706aBcE058DB29eB36578c463cf295F180a1Fe9C",
	},
	common.MainnetMode: {
		Reserve:   "0x63825c174ab367968EC60f061753D3bbD36A0D8F",
		Network:   "0x818E6FECD516Ecc3849DAf6845e3EC868087B755",
		Wrapper:   "0x6172AFC8c00c46E0D07ce3AF203828198194620a",
		Pricing:   "0x798AbDA6Cc246D0EDbA912092A2a3dBd3d11191B",
		FeeBurner: "0xed4f53268bfdFF39B36E8786247bA3A02Cf34B04",
		Whitelist: "0x6e106a75d369d09a9ea1dcc16da844792aa669a3",
		ThirdPartyReserves: []string{"0x2aab2b157a03915c8a73adae735d0cf51c872f31",
			"0x6f50e41885fdc44dbdf7797df0393779a9c0a3a6",
			"0x4d864b5b4f866f65f53cbaad32eb9574760865e6",
			"0xc935cad589bebd8673104073d5a5eccfe67fb7b1"},
		InternalNetwork: "0x91a502C678605fbCe581eae053319747482276b9",
	},
	common.RopstenMode: {
		Reserve:   "0x0FC1CF3e7DD049F7B42e6823164A64F76fC06Be0",
		Network:   "0x0a56d8a49E71da8d7F9C65F95063dB48A3C9560B",
		Wrapper:   "0x9de0a60F4A489e350cD8E3F249f4080858Af41d3",
		Pricing:   "0x535DE1F5a982c2a896da62790a42723A71c0c12B",
		FeeBurner: "0x89B5c470559b80e541E53eF78244edD112c7C58A",
	},
}

func mustGetAddressesConfig(kyberEnv string) common.AddressConfig {
	result, avail := AddressConfigs[kyberEnv]
	if avail {
		return result
	}
	if kyberEnv == common.SimulationMode {
		result, err := loadAddressFromFile(simSettingPath)
		if err != nil {
			log.Panicf("cannot load address from file, err: %v", err)
		}
		return result
	}
	if kyberEnv == common.ProductionMode {
		return AddressConfigs[common.MainnetMode]
	}
	log.Panicf("cannot get address config for mode %s", kyberEnv)

	return result
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
