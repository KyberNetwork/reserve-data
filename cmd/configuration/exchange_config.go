package configuration

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/KyberNetwork/reserve-data/common"

	ethereum "github.com/ethereum/go-ethereum/common"
)

//ExchangeConfigs store exchange config according to env mode.
var ExchangeConfigs = map[string]map[common.ExchangeID]common.ExchangeAddresses{
	common.DevMode: map[common.ExchangeID]common.ExchangeAddresses{
		"binance": map[string]ethereum.Address{
			"ETH":  ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"OMG":  ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"KNC":  ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"SNT":  ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"ELF":  ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"POWR": ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"MANA": ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"BAT":  ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"REQ":  ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"GTO":  ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"ENG":  ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"SALT": ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"APPC": ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"RDN":  ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"BQX":  ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"ZIL":  ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"AST":  ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"LINK": ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"ENJ":  ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"AION": ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"AE":   ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"BLZ":  ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"SUB":  ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"POE":  ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"CHAT": ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"BNT":  ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"TUSD": ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
		},
		"huobi": map[string]ethereum.Address{
			"ETH":  ethereum.HexToAddress("0x0c8fd73eaf6089ef1b91231d0a07d0d2ca2b9d66"),
			"ABT":  ethereum.HexToAddress("0x0c8fd73eaf6089ef1b91231d0a07d0d2ca2b9d66"),
			"POLY": ethereum.HexToAddress("0x0c8fd73eaf6089ef1b91231d0a07d0d2ca2b9d66"),
			"LBA":  ethereum.HexToAddress("0x0c8fd73eaf6089ef1b91231d0a07d0d2ca2b9d66"),
			"EDU":  ethereum.HexToAddress("0x0c8fd73eaf6089ef1b91231d0a07d0d2ca2b9d66"),
			"CVC":  ethereum.HexToAddress("0x0c8fd73eaf6089ef1b91231d0a07d0d2ca2b9d66"),
			"WAX":  ethereum.HexToAddress("0x0c8fd73eaf6089ef1b91231d0a07d0d2ca2b9d66"),
			"PAY":  ethereum.HexToAddress("0x0c8fd73eaf6089ef1b91231d0a07d0d2ca2b9d66"),
			"DTA":  ethereum.HexToAddress("0x0c8fd73eaf6089ef1b91231d0a07d0d2ca2b9d66"),
		},
		"stable_exchange": map[string]ethereum.Address{
			"ETH": ethereum.HexToAddress("0xFDF28Bf25779ED4cA74e958d54653260af604C20"),
			"DGX": ethereum.HexToAddress("0xFDF28Bf25779ED4cA74e958d54653260af604C20"),
		},
	},
	common.StagingMode: map[common.ExchangeID]common.ExchangeAddresses{
		"binance": map[string]ethereum.Address{
			"ETH":  ethereum.HexToAddress("0x1ae659f93ba2fc0a1f379545cf9335adb75fa547"),
			"OMG":  ethereum.HexToAddress("0x1ae659f93ba2fc0a1f379545cf9335adb75fa547"),
			"KNC":  ethereum.HexToAddress("0x1ae659f93ba2fc0a1f379545cf9335adb75fa547"),
			"SNT":  ethereum.HexToAddress("0x1ae659f93ba2fc0a1f379545cf9335adb75fa547"),
			"ELF":  ethereum.HexToAddress("0x1ae659f93ba2fc0a1f379545cf9335adb75fa547"),
			"POWR": ethereum.HexToAddress("0x1ae659f93ba2fc0a1f379545cf9335adb75fa547"),
			"MANA": ethereum.HexToAddress("0x1ae659f93ba2fc0a1f379545cf9335adb75fa547"),
			"BAT":  ethereum.HexToAddress("0x1ae659f93ba2fc0a1f379545cf9335adb75fa547"),
			"REQ":  ethereum.HexToAddress("0x1ae659f93ba2fc0a1f379545cf9335adb75fa547"),
			"GTO":  ethereum.HexToAddress("0x1ae659f93ba2fc0a1f379545cf9335adb75fa547"),
			"ENG":  ethereum.HexToAddress("0x1ae659f93ba2fc0a1f379545cf9335adb75fa547"),
			"SALT": ethereum.HexToAddress("0x1ae659f93ba2fc0a1f379545cf9335adb75fa547"),
			"APPC": ethereum.HexToAddress("0x1ae659f93ba2fc0a1f379545cf9335adb75fa547"),
			"RDN":  ethereum.HexToAddress("0x1ae659f93ba2fc0a1f379545cf9335adb75fa547"),
			"BQX":  ethereum.HexToAddress("0x1ae659f93ba2fc0a1f379545cf9335adb75fa547"),
			"ZIL":  ethereum.HexToAddress("0x1ae659f93ba2fc0a1f379545cf9335adb75fa547"),
			"AST":  ethereum.HexToAddress("0x1ae659f93ba2fc0a1f379545cf9335adb75fa547"),
			"LINK": ethereum.HexToAddress("0x1ae659f93ba2fc0a1f379545cf9335adb75fa547"),
			"ENJ":  ethereum.HexToAddress("0x1ae659f93ba2fc0a1f379545cf9335adb75fa547"),
			"AION": ethereum.HexToAddress("0x1ae659f93ba2fc0a1f379545cf9335adb75fa547"),
			"AE":   ethereum.HexToAddress("0x1ae659f93ba2fc0a1f379545cf9335adb75fa547"),
			"BLZ":  ethereum.HexToAddress("0x1ae659f93ba2fc0a1f379545cf9335adb75fa547"),
			"SUB":  ethereum.HexToAddress("0x1ae659f93ba2fc0a1f379545cf9335adb75fa547"),
			"POE":  ethereum.HexToAddress("0x1ae659f93ba2fc0a1f379545cf9335adb75fa547"),
			"CHAT": ethereum.HexToAddress("0x1ae659f93ba2fc0a1f379545cf9335adb75fa547"),
			"BNT":  ethereum.HexToAddress("0x1ae659f93ba2fc0a1f379545cf9335adb75fa547"),
			"TUSD": ethereum.HexToAddress("0x1ae659f93ba2fc0a1f379545cf9335adb75fa547"),
		},
		"huobi": map[string]ethereum.Address{
			"ETH":  ethereum.HexToAddress("0xb48ee85467bf613a22244084c1a46c2deac18dd0"),
			"ABT":  ethereum.HexToAddress("0xb48ee85467bf613a22244084c1a46c2deac18dd0"),
			"POLY": ethereum.HexToAddress("0xb48ee85467bf613a22244084c1a46c2deac18dd0"),
			"LBA":  ethereum.HexToAddress("0xb48ee85467bf613a22244084c1a46c2deac18dd0"),
			"EDU":  ethereum.HexToAddress("0xb48ee85467bf613a22244084c1a46c2deac18dd0"),
			"CVC":  ethereum.HexToAddress("0xb48ee85467bf613a22244084c1a46c2deac18dd0"),
			"WAX":  ethereum.HexToAddress("0xb48ee85467bf613a22244084c1a46c2deac18dd0"),
			"PAY":  ethereum.HexToAddress("0xb48ee85467bf613a22244084c1a46c2deac18dd0"),
			"DTA":  ethereum.HexToAddress("0xb48ee85467bf613a22244084c1a46c2deac18dd0"),
		},
		"stable_exchange": map[string]ethereum.Address{
			"ETH": ethereum.HexToAddress("0xFDF28Bf25779ED4cA74e958d54653260af604C20"),
			"DGX": ethereum.HexToAddress("0xFDF28Bf25779ED4cA74e958d54653260af604C20"),
		},
	},
	common.MainnetMode: map[common.ExchangeID]common.ExchangeAddresses{
		"binance": {
			"ETH":  ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"OMG":  ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"KNC":  ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"SNT":  ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"ELF":  ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"POWR": ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"MANA": ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"BAT":  ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"REQ":  ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"GTO":  ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"ENG":  ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"SALT": ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"APPC": ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"RDN":  ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"BQX":  ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"ZIL":  ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"AST":  ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"LINK": ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"ENJ":  ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"AION": ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"AE":   ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"BLZ":  ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"SUB":  ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"POE":  ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"CHAT": ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"BNT":  ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"TUSD": ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
		},
		"huobi": {
			"ETH":  ethereum.HexToAddress("0x0c8fd73eaf6089ef1b91231d0a07d0d2ca2b9d66"),
			"ABT":  ethereum.HexToAddress("0x0c8fd73eaf6089ef1b91231d0a07d0d2ca2b9d66"),
			"POLY": ethereum.HexToAddress("0x0c8fd73eaf6089ef1b91231d0a07d0d2ca2b9d66"),
			"LBA":  ethereum.HexToAddress("0x0c8fd73eaf6089ef1b91231d0a07d0d2ca2b9d66"),
			"EDU":  ethereum.HexToAddress("0x0c8fd73eaf6089ef1b91231d0a07d0d2ca2b9d66"),
			"CVC":  ethereum.HexToAddress("0x0c8fd73eaf6089ef1b91231d0a07d0d2ca2b9d66"),
			"WAX":  ethereum.HexToAddress("0x0c8fd73eaf6089ef1b91231d0a07d0d2ca2b9d66"),
			"PAY":  ethereum.HexToAddress("0x0c8fd73eaf6089ef1b91231d0a07d0d2ca2b9d66"),
			"DTA":  ethereum.HexToAddress("0x0c8fd73eaf6089ef1b91231d0a07d0d2ca2b9d66"),
		},
		"stable_exchange": {
			"ETH": ethereum.HexToAddress("0xFDF28Bf25779ED4cA74e958d54653260af604C20"),
			"DGX": ethereum.HexToAddress("0xFDF28Bf25779ED4cA74e958d54653260af604C20"),
		},
	},
	common.RopstenMode: map[common.ExchangeID]common.ExchangeAddresses{
		"binance": {
			"ETH":  ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"OMG":  ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"KNC":  ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"SNT":  ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"EOS":  ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"ELF":  ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"POWR": ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"MANA": ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"BAT":  ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"REQ":  ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
			"GTO":  ethereum.HexToAddress("0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"),
		},
	},
}

func mustGetExchangeConfig(kyberEnv string) map[common.ExchangeID]common.ExchangeAddresses {
	result, avail := ExchangeConfigs[kyberEnv]
	if avail {
		return result
	}
	if kyberEnv == common.SimulationMode {
		result, err := loadDepositAddressFromFile(simSettingPath)
		if err != nil {
			log.Panicf("cannot load data from pre-defined simluation setting file, err: %v", err)
		}
		return result
	}
	if kyberEnv == common.ProductionMode {
		return ExchangeConfigs[common.MainnetMode]
	}
	log.Panicf("cannot get exchange config for mode %s", kyberEnv)

	return nil
}

// exchangeDepositAddress type stores a map[tokenID]depositaddress
// it is used to read address config from a file.
type exchangeDepositAddress map[string]string

// AddressDepositConfig struct contain a map[exchangeName],
// it is used mainly to read addfress config from JSON file.
type AddressDepositConfig struct {
	Exchanges map[string]exchangeDepositAddress `json:"exchanges"`
}

func loadDepositAddressFromFile(path string) (map[common.ExchangeID]common.ExchangeAddresses, error) {
	var (
		result          = make(map[common.ExchangeID]common.ExchangeAddresses)
		exAddressConfig AddressDepositConfig
	)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return result, err
	}
	if err := json.Unmarshal([]byte(data), &exAddressConfig); err != nil {
		return result, err
	}
	for exchangeID, addrs := range exAddressConfig.Exchanges {
		exchangeAddresses := convertToAddressMap(addrs)
		result[common.ExchangeID(exchangeID)] = exchangeAddresses
	}
	return result, nil
}

func convertToAddressMap(data exchangeDepositAddress) common.ExchangeAddresses {
	result := make(common.ExchangeAddresses)
	for token, addrStr := range data {
		result[token] = ethereum.HexToAddress(addrStr)
	}
	return result
}
