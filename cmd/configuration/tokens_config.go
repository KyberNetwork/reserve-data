package configuration

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/KyberNetwork/reserve-data/common"
)

const (
	simSettingPath = "shared/deployment_dev.json"
)

//TokenConfigs store token configuration for each modes
//Sim mode require special care.
var TokenConfigs = map[string]map[string]common.Token{
	common.DevMode: map[string]common.Token{
		"KNC":   common.NewToken("KNC", "KyberNetwork", "0xdd974D5C2e2928deA5F71b9825b8b646686BD200", 18, true, true, common.GetTimepoint()),
		"BQX":   common.NewToken("BQX", "Ethos", "0x5Af2Be193a6ABCa9c8817001F45744777Db30756", 8, true, true, common.GetTimepoint()),
		"WAX":   common.NewToken("WAX", "Wax", "0x39bb259f66e1c59d5abef88375979b4d20d98022", 8, true, true, common.GetTimepoint()),
		"CHAT":  common.NewToken("CHAT", "Chatcoin", "0x442bc47357919446eabc18c7211e57a13d983469", 18, true, true, common.GetTimepoint()),
		"DTA":   common.NewToken("DTA", "Data", "0x69b148395ce0015c13e36bffbad63f49ef874e03", 18, true, true, common.GetTimepoint()),
		"WINGS": common.NewToken("WINGS", "WINGS", "0x667088b212ce3d06a1b553a7221E1fD19000d9aF", 18, true, false, common.GetTimepoint()),
		"OMG":   common.NewToken("OMG", "OmiseGO", "0xd26114cd6EE289AccF82350c8d8487fedB8A0C07", 18, true, true, common.GetTimepoint()),
		"GTO":   common.NewToken("GTO", "Gifto", "0xc5bbae50781be1669306b9e001eff57a2957b09d", 5, true, true, common.GetTimepoint()),
		"LINK":  common.NewToken("LINK", "Chain Link", "0x514910771af9ca656af840dff83e8264ecf986ca", 18, true, true, common.GetTimepoint()),
		"LBA":   common.NewToken("LBA", "Libra Credit", "0xfe5f141bf94fe84bc28ded0ab966c16b17490657", 18, true, true, common.GetTimepoint()),
		"COFI":  common.NewToken("COFI", "CoinFi", "0x3136ef851592acf49ca4c825131e364170fa32b3", 0, true, false, common.GetTimepoint()),
		"AION":  common.NewToken("AION", "Aion", "0x4CEdA7906a5Ed2179785Cd3A40A69ee8bc99C466", 8, true, true, common.GetTimepoint()),
		"POLY":  common.NewToken("POLY", "Polymath", "0x9992ec3cf6a55b00978cddf2b27bc6882d88d1ec", 18, true, true, common.GetTimepoint()),
		"CVC":   common.NewToken("CVC", "Civic", "0x41e5560054824eA6B0732E656E3Ad64E20e94E45", 8, true, true, common.GetTimepoint()),
		"SUB":   common.NewToken("SUB", "Substratum", "0x12480e24eb5bec1a9d4369cab6a80cad3c0a377a", 2, true, true, common.GetTimepoint()),
		"ABT":   common.NewToken("ABT", "ArcBlock", "0xb98d4c97425d9908e66e53a6fdf673acca0be986", 18, true, true, common.GetTimepoint()),
		"APPC":  common.NewToken("APPC", "AppCoins", "0x1a7a8bd9106f2b8d977e08582dc7d24c723ab0db", 18, true, true, common.GetTimepoint()),
		"DGX":   common.NewToken("DGX", "Digix Gold", "0x4f3afec4e5a3f2a6a1a411def7d7dfe50ee057bf", 9, true, true, common.GetTimepoint()),
		"IOST":  common.NewToken("IOST", "IOStoken", "0xfa1a856cfa3409cfa145fa4e20eb270df3eb21ab", 18, true, false, common.GetTimepoint()),
		"EDU":   common.NewToken("EDU", "EduCoin", "0xf263292e14d9d8ecd55b58dad1f1df825a874b7c", 18, true, true, common.GetTimepoint()),
		"STORM": common.NewToken("STORM", "Storm", "0xd0a4b8946cb52f0661273bfbc6fd0e0c75fc6433", 18, true, false, common.GetTimepoint()),
		"TUSD":  common.NewToken("TUSD", "TrueUSD", "0x8dd5fbce2f6a956c3022ba3663759011dd51e73e", 18, true, true, common.GetTimepoint()),
		"WABI":  common.NewToken("WABI", "WaBi", "0x286BDA1413a2Df81731D4930ce2F862a35A609fE", 18, true, false, common.GetTimepoint()),
		"LEND":  common.NewToken("LEND", "EthLend", "0x80fB784B7eD66730e8b1DBd9820aFD29931aab03", 18, true, false, common.GetTimepoint()),
		"BAT":   common.NewToken("BAT", "Basic Attention Token", "0x0d8775f648430679a709e98d2b0cb6250d2887ef", 18, true, true, common.GetTimepoint()),
		"BLZ":   common.NewToken("BLZ", "Bluezelle", "0x5732046a883704404f284ce41ffadd5b007fd668", 18, true, true, common.GetTimepoint()),
		"BBO":   common.NewToken("BBO", "Bigbom", "0x84f7c44b6fed1080f647e354d552595be2cc602f", 18, true, false, common.GetTimepoint()),
		"BNT":   common.NewToken("BNT", "Bancor", "0x1f573d6fb3f13d689ff844b4ce37794d79a7ff1c", 18, true, true, common.GetTimepoint()),
		"ZIL":   common.NewToken("ZIL", "Zilliqa", "0x05f4a42e251f2d52b8ed15e9fedaacfcef1fad27", 12, true, true, common.GetTimepoint()),
		"ELF":   common.NewToken("ELF", "AELF", "0xbf2179859fc6d5bee9bf9158632dc51678a4100e", 18, true, true, common.GetTimepoint()),
		"RDN":   common.NewToken("RDN", "Raiden", "0x255aa6df07540cb5d3d297f0d0d4d84cb52bc8e6", 18, true, true, common.GetTimepoint()),
		"ENG":   common.NewToken("ENG", "Enigma", "0xf0ee6b27b759c9893ce4f094b49ad28fd15a23e4", 8, true, true, common.GetTimepoint()),
		"RCN":   common.NewToken("RCN", "Ripio Credit Network", "0xf970b8e36e23f7fc3fd752eea86f8be8d83375a6", 18, true, false, common.GetTimepoint()),
		"ELEC":  common.NewToken("ELEC", "ElectrifyAsia", "0xd49ff13661451313ca1553fd6954bd1d9b6e02b9", 18, true, false, common.GetTimepoint()),
		"EOS":   common.NewToken("EOS", "Eos", "0x86Fa049857E0209aa7D9e616F7eb3b3B78ECfdb0", 18, true, false, common.GetTimepoint()),
		"MANA":  common.NewToken("MANA", "MANA", "0x0f5d2fb29fb7d3cfee444a200298f468908cc942", 18, true, true, common.GetTimepoint()),
		"AST":   common.NewToken("AST", "AirSwap", "0x27054b13b1b798b345b591a4d22e6562d47ea75a", 4, true, true, common.GetTimepoint()),
		"PAL":   common.NewToken("PAL", "PolicyPal Network", "0xfedae5642668f8636a11987ff386bfd215f942ee", 18, true, false, common.GetTimepoint()),
		"REQ":   common.NewToken("REQ", "Request", "0x8f8221afbb33998d8584a2b05749ba73c37a938a", 18, true, true, common.GetTimepoint()),
		"TOMO":  common.NewToken("TOMO", "Tomocoin", "0x8b353021189375591723E7384262F45709A3C3dC", 18, true, false, common.GetTimepoint()),
		"MDS":   common.NewToken("MDS", "MediShares", "0x66186008C1050627F979d464eABb258860563dbE", 18, true, false, common.GetTimepoint()),
		"MOC":   common.NewToken("MOC", "Moss Land", "0x865ec58b06bf6305b886793aa20a2da31d034e68", 0, true, false, common.GetTimepoint()),
		"POWR":  common.NewToken("POWR", "Power Ledger", "0x595832f8fc6bf59c85c527fec3740a1b7a361269", 6, true, true, common.GetTimepoint()),
		"DAI":   common.NewToken("DAI", "DAI", "0x89d24a6b4ccb1b6faa2625fe562bdd9a23260359", 18, true, false, common.GetTimepoint()),
		"SALT":  common.NewToken("SALT", "Salt", "0x4156d3342d5c385a87d264f90653733592000581", 8, true, true, common.GetTimepoint()),
		"MOT":   common.NewToken("MOT", "Olympus Labs", "0x263c618480dbe35c300d8d5ecda19bbb986acaed", 18, true, false, common.GetTimepoint()),
		"ENJ":   common.NewToken("ENJ", "EnjinCoin", "0xf629cbd94d3791c9250152bd8dfbdf380e2a3b9c", 18, true, true, common.GetTimepoint()),
		"PAY":   common.NewToken("PAY", "TenX", "0xB97048628DB6B661D4C2aA833e95Dbe1A905B280", 18, true, true, common.GetTimepoint()),
		"MTL":   common.NewToken("MTL", "Metal", "0xF433089366899D83a9f26A773D59ec7eCF30355e", 8, true, false, common.GetTimepoint()),
		"ETH":   common.NewToken("ETH", "Ethereum", "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee", 18, true, true, common.GetTimepoint()),
		"SNT":   common.NewToken("SNT", "STATUS", "0x744d70fdbe2ba4cf95131626614a1763df805b9e", 18, true, true, common.GetTimepoint()),
		"ADX":   common.NewToken("ADX", "AdEx", "0x4470BB87d77b963A013DB939BE332f927f2b992e", 4, true, false, common.GetTimepoint()),
		"AE":    common.NewToken("AE", "Aeternity", "0x5ca9a71b1d01849c0a95490cc00559717fcf0d1d", 18, true, true, common.GetTimepoint()),
		"POE":   common.NewToken("POE", "Po.et", "0x0e0989b1f9b8a38983c2ba8053269ca62ec9b195", 8, true, true, common.GetTimepoint()),
	},
	common.StagingMode: map[string]common.Token{
		"PAY":   common.NewToken("PAY", "TenX", "0xB97048628DB6B661D4C2aA833e95Dbe1A905B280", 18, true, true, common.GetTimepoint()),
		"MTL":   common.NewToken("MTL", "Metal", "0xF433089366899D83a9f26A773D59ec7eCF30355e", 8, true, false, common.GetTimepoint()),
		"LEND":  common.NewToken("LEND", "EthLend", "0x80fB784B7eD66730e8b1DBd9820aFD29931aab03", 18, true, false, common.GetTimepoint()),
		"ENG":   common.NewToken("ENG", "Enigma", "0xf0ee6b27b759c9893ce4f094b49ad28fd15a23e4", 8, true, true, common.GetTimepoint()),
		"LINK":  common.NewToken("LINK", "Chain Link", "0x514910771af9ca656af840dff83e8264ecf986ca", 18, true, true, common.GetTimepoint()),
		"POLY":  common.NewToken("POLY", "Polymath", "0x9992ec3cf6a55b00978cddf2b27bc6882d88d1ec", 18, true, true, common.GetTimepoint()),
		"BBO":   common.NewToken("BBO", "Bigbom", "0x84f7c44b6fed1080f647e354d552595be2cc602f", 18, true, false, common.GetTimepoint()),
		"REQ":   common.NewToken("REQ", "Request", "0x8f8221afbb33998d8584a2b05749ba73c37a938a", 18, true, true, common.GetTimepoint()),
		"SALT":  common.NewToken("SALT", "Salt", "0x4156D3342D5c385a87D264F90653733592000581", 8, true, true, common.GetTimepoint()),
		"BQX":   common.NewToken("BQX", "Ethos", "0x5af2be193a6abca9c8817001f45744777db30756", 8, true, true, common.GetTimepoint()),
		"KNC":   common.NewToken("KNC", "KyberNetwork", "0xdd974D5C2e2928deA5F71b9825b8b646686BD200", 18, true, true, common.GetTimepoint()),
		"SNT":   common.NewToken("SNT", "STATUS", "0x744d70fdbe2ba4cf95131626614a1763df805b9e", 18, true, true, common.GetTimepoint()),
		"DTA":   common.NewToken("DTA", "Data", "0x69b148395ce0015c13e36bffbad63f49ef874e03", 18, true, true, common.GetTimepoint()),
		"STORM": common.NewToken("STORM", "Storm", "0xd0a4b8946cb52f0661273bfbc6fd0e0c75fc6433", 18, true, false, common.GetTimepoint()),
		"ELEC":  common.NewToken("ELEC", "ElectrifyAsia", "0xd49ff13661451313ca1553fd6954bd1d9b6e02b9", 18, true, false, common.GetTimepoint()),
		"EDU":   common.NewToken("EDU", "EduCoin", "0xf263292e14d9d8ecd55b58dad1f1df825a874b7c", 18, true, true, common.GetTimepoint()),
		"POE":   common.NewToken("POE", "Po.et", "0x0e0989b1f9b8a38983c2ba8053269ca62ec9b195", 8, true, true, common.GetTimepoint()),
		"AST":   common.NewToken("AST", "AirSwap", "0x27054b13b1b798b345b591a4d22e6562d47ea75a", 4, true, true, common.GetTimepoint()),
		"ENJ":   common.NewToken("ENJ", "EnjinCoin", "0xf629cbd94d3791c9250152bd8dfbdf380e2a3b9c", 18, true, true, common.GetTimepoint()),
		"AION":  common.NewToken("AION", "Aion", "0x4CEdA7906a5Ed2179785Cd3A40A69ee8bc99C466", 8, true, true, common.GetTimepoint()),
		"CHAT":  common.NewToken("CHAT", "Chatcoin", "0x442bc47357919446eabc18c7211e57a13d983469", 18, true, true, common.GetTimepoint()),
		"EOS":   common.NewToken("EOS", "Eos", "0x86Fa049857E0209aa7D9e616F7eb3b3B78ECfdb0", 18, true, false, common.GetTimepoint()),
		"ELF":   common.NewToken("ELF", "AELF", "0xbf2179859fc6d5bee9bf9158632dc51678a4100e", 18, true, true, common.GetTimepoint()),
		"BAT":   common.NewToken("BAT", "Basic Attention Token", "0x0d8775f648430679a709e98d2b0cb6250d2887ef", 18, true, true, common.GetTimepoint()),
		"GTO":   common.NewToken("GTO", "Gifto", "0xc5bbae50781be1669306b9e001eff57a2957b09d", 5, true, true, common.GetTimepoint()),
		"MDS":   common.NewToken("MDS", "MediShares", "0x66186008C1050627F979d464eABb258860563dbE", 18, true, false, common.GetTimepoint()),
		"POWR":  common.NewToken("POWR", "Power Ledger", "0x595832f8fc6bf59c85c527fec3740a1b7a361269", 6, true, true, common.GetTimepoint()),
		"MOT":   common.NewToken("MOT", "Olympus Labs", "0x263c618480dbe35c300d8d5ecda19bbb986acaed", 18, true, false, common.GetTimepoint()),
		"IOST":  common.NewToken("IOST", "IOStoken", "0xfa1a856cfa3409cfa145fa4e20eb270df3eb21ab", 18, true, false, common.GetTimepoint()),
		"ETH":   common.NewToken("ETH", "Ethereum", "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee", 18, true, true, common.GetTimepoint()),
		"CVC":   common.NewToken("CVC", "Civic", "0x41e5560054824eA6B0732E656E3Ad64E20e94E45", 8, true, true, common.GetTimepoint()),
		"RDN":   common.NewToken("RDN", "Raiden", "0x255aa6df07540cb5d3d297f0d0d4d84cb52bc8e6", 18, true, true, common.GetTimepoint()),
		"ADX":   common.NewToken("ADX", "AdEx", "0x4470BB87d77b963A013DB939BE332f927f2b992e", 4, true, false, common.GetTimepoint()),
		"ABT":   common.NewToken("ABT", "ArcBlock", "0xb98d4c97425d9908e66e53a6fdf673acca0be986", 18, true, true, common.GetTimepoint()),
		"RCN":   common.NewToken("RCN", "Ripio Credit Network", "0xf970b8e36e23f7fc3fd752eea86f8be8d83375a6", 18, true, false, common.GetTimepoint()),
		"BLZ":   common.NewToken("BLZ", "Bluezelle", "0x5732046a883704404f284ce41ffadd5b007fd668", 18, true, true, common.GetTimepoint()),
		"OMG":   common.NewToken("OMG", "OmiseGO", "0xd26114cd6EE289AccF82350c8d8487fedB8A0C07", 18, true, true, common.GetTimepoint()),
		"TUSD":  common.NewToken("TUSD", "TrueUSD", "0x8dd5fbce2f6a956c3022ba3663759011dd51e73e", 18, true, true, common.GetTimepoint()),
		"COFI":  common.NewToken("COFI", "CoinFi", "0x3136ef851592acf49ca4c825131e364170fa32b3", 0, true, false, common.GetTimepoint()),
		"SUB":   common.NewToken("SUB", "Substratum", "0x12480e24eb5bec1a9d4369cab6a80cad3c0a377a", 2, true, true, common.GetTimepoint()),
		"ZIL":   common.NewToken("ZIL", "Zilliqa", "0x05f4a42e251f2d52b8ed15e9fedaacfcef1fad27", 12, true, true, common.GetTimepoint()),
		"DAI":   common.NewToken("DAI", "DAI", "0x89d24a6b4ccb1b6faa2625fe562bdd9a23260359", 18, true, false, common.GetTimepoint()),
		"BNT":   common.NewToken("BNT", "Bancor", "0x1f573d6fb3f13d689ff844b4ce37794d79a7ff1c", 18, true, true, common.GetTimepoint()),
		"TOMO":  common.NewToken("TOMO", "Tomocoin", "0x8b353021189375591723E7384262F45709A3C3dC", 18, true, false, common.GetTimepoint()),
		"WABI":  common.NewToken("WABI", "WaBi", "0x286BDA1413a2Df81731D4930ce2F862a35A609fE", 18, true, false, common.GetTimepoint()),
		"DGX":   common.NewToken("DGX", "Digix Gold", "0x4f3afec4e5a3f2a6a1a411def7d7dfe50ee057bf", 9, true, true, common.GetTimepoint()),
		"WAX":   common.NewToken("WAX", "Wax", "0x39bb259f66e1c59d5abef88375979b4d20d98022", 8, true, true, common.GetTimepoint()),
		"AE":    common.NewToken("AE", "Aeternity", "0x5ca9a71b1d01849c0a95490cc00559717fcf0d1d", 18, true, true, common.GetTimepoint()),
		"WINGS": common.NewToken("WINGS", "WINGS", "0x667088b212ce3d06a1b553a7221E1fD19000d9aF", 18, true, false, common.GetTimepoint()),
		"MANA":  common.NewToken("MANA", "MANA", "0x0f5d2fb29fb7d3cfee444a200298f468908cc942", 18, true, true, common.GetTimepoint()),
		"APPC":  common.NewToken("APPC", "AppCoins", "0x1a7a8bd9106f2b8d977e08582dc7d24c723ab0db", 18, true, true, common.GetTimepoint()),
		"LBA":   common.NewToken("LBA", "Libra Credit", "0xfe5f141bf94fe84bc28ded0ab966c16b17490657", 18, true, true, common.GetTimepoint()),
		"PAL":   common.NewToken("PAL", "PolicyPal Network", "0xfedae5642668f8636a11987ff386bfd215f942ee", 18, true, false, common.GetTimepoint()),
	},
	common.MainnetMode: map[string]common.Token{
		"ETH":   common.NewToken("ETH", "Ethereum", "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee", 18, true, true, common.GetTimepoint()),
		"ADX":   common.NewToken("ADX", "AdEx", "0x4470BB87d77b963A013DB939BE332f927f2b992e", 4, true, false, common.GetTimepoint()),
		"LINK":  common.NewToken("LINK", "Chain Link", "0x514910771af9ca656af840dff83e8264ecf986ca", 18, true, true, common.GetTimepoint()),
		"ENJ":   common.NewToken("ENJ", "EnjinCoin", "0xf629cbd94d3791c9250152bd8dfbdf380e2a3b9c", 18, true, true, common.GetTimepoint()),
		"TOMO":  common.NewToken("TOMO", "Tomocoin", "0x8b353021189375591723E7384262F45709A3C3dC", 18, true, false, common.GetTimepoint()),
		"DTA":   common.NewToken("DTA", "Data", "0x69b148395ce0015c13e36bffbad63f49ef874e03", 18, true, true, common.GetTimepoint()),
		"EOS":   common.NewToken("EOS", "Eos", "0x86Fa049857E0209aa7D9e616F7eb3b3B78ECfdb0", 18, true, false, common.GetTimepoint()),
		"WINGS": common.NewToken("WINGS", "WINGS", "0x667088b212ce3d06a1b553a7221E1fD19000d9aF", 18, true, false, common.GetTimepoint()),
		"MTL":   common.NewToken("MTL", "Metal", "0xF433089366899D83a9f26A773D59ec7eCF30355e", 8, true, false, common.GetTimepoint()),
		"ZIL":   common.NewToken("ZIL", "Zilliqa", "0x05f4a42e251f2d52b8ed15e9fedaacfcef1fad27", 12, true, true, common.GetTimepoint()),
		"IOST":  common.NewToken("IOST", "IOStoken", "0xfa1a856cfa3409cfa145fa4e20eb270df3eb21ab", 18, true, false, common.GetTimepoint()),
		"OMG":   common.NewToken("OMG", "OmiseGO", "0xd26114cd6EE289AccF82350c8d8487fedB8A0C07", 18, true, true, common.GetTimepoint()),
		"AST":   common.NewToken("AST", "AirSwap", "0x27054b13b1b798b345b591a4d22e6562d47ea75a", 4, true, true, common.GetTimepoint()),
		"POLY":  common.NewToken("POLY", "Polymath", "0x9992ec3cf6a55b00978cddf2b27bc6882d88d1ec", 18, true, true, common.GetTimepoint()),
		"WAX":   common.NewToken("WAX", "Wax", "0x39bb259f66e1c59d5abef88375979b4d20d98022", 8, true, true, common.GetTimepoint()),
		"COFI":  common.NewToken("COFI", "CoinFi", "0x3136ef851592acf49ca4c825131e364170fa32b3", 0, true, false, common.GetTimepoint()),
		"KNC":   common.NewToken("KNC", "KyberNetwork", "0xdd974D5C2e2928deA5F71b9825b8b646686BD200", 18, true, true, common.GetTimepoint()),
		"AE":    common.NewToken("AE", "Aeternity", "0x5ca9a71b1d01849c0a95490cc00559717fcf0d1d", 18, true, true, common.GetTimepoint()),
		"BBO":   common.NewToken("BBO", "Bigbom", "0x84f7c44b6fed1080f647e354d552595be2cc602f", 18, true, false, common.GetTimepoint()),
		"TUSD":  common.NewToken("TUSD", "TrueUSD", "0x8dd5fbce2f6a956c3022ba3663759011dd51e73e", 18, true, true, common.GetTimepoint()),
		"APPC":  common.NewToken("APPC", "AppCoins", "0x1a7a8bd9106f2b8d977e08582dc7d24c723ab0db", 18, true, true, common.GetTimepoint()),
		"SALT":  common.NewToken("SALT", "Salt", "0x4156d3342d5c385a87d264f90653733592000581", 8, true, true, common.GetTimepoint()),
		"CVC":   common.NewToken("CVC", "Civic", "0x41e5560054824eA6B0732E656E3Ad64E20e94E45", 8, true, true, common.GetTimepoint()),
		"POE":   common.NewToken("POE", "Po.et", "0x0e0989b1f9b8a38983c2ba8053269ca62ec9b195", 8, true, true, common.GetTimepoint()),
		"CHAT":  common.NewToken("CHAT", "Chatcoin", "0x442bc47357919446eabc18c7211e57a13d983469", 18, true, true, common.GetTimepoint()),
		"MOC":   common.NewToken("MOC", "Moss Land", "0x865ec58b06bf6305b886793aa20a2da31d034e68", 0, true, false, common.GetTimepoint()),
		"GTO":   common.NewToken("GTO", "Gifto", "0xc5bbae50781be1669306b9e001eff57a2957b09d", 5, true, true, common.GetTimepoint()),
		"DGX":   common.NewToken("DGX", "Digix Gold", "0x4f3afec4e5a3f2a6a1a411def7d7dfe50ee057bf", 9, true, true, common.GetTimepoint()),
		"BLZ":   common.NewToken("BLZ", "Bluezelle", "0x5732046a883704404f284ce41ffadd5b007fd668", 18, true, true, common.GetTimepoint()),
		"BNT":   common.NewToken("BNT", "Bancor", "0x1f573d6fb3f13d689ff844b4ce37794d79a7ff1c", 18, true, true, common.GetTimepoint()),
		"ENG":   common.NewToken("ENG", "Enigma", "0xf0ee6b27b759c9893ce4f094b49ad28fd15a23e4", 8, true, true, common.GetTimepoint()),
		"ABT":   common.NewToken("ABT", "ArcBlock", "0xb98d4c97425d9908e66e53a6fdf673acca0be986", 18, true, true, common.GetTimepoint()),
		"PAL":   common.NewToken("PAL", "PolicyPal Network", "0xfedae5642668f8636a11987ff386bfd215f942ee", 18, true, false, common.GetTimepoint()),
		"SNT":   common.NewToken("SNT", "STATUS", "0x744d70fdbe2ba4cf95131626614a1763df805b9e", 18, true, true, common.GetTimepoint()),
		"ELF":   common.NewToken("ELF", "AELF", "0xbf2179859fc6d5bee9bf9158632dc51678a4100e", 18, true, true, common.GetTimepoint()),
		"REQ":   common.NewToken("REQ", "Request", "0x8f8221afbb33998d8584a2b05749ba73c37a938a", 18, true, true, common.GetTimepoint()),
		"LEND":  common.NewToken("LEND", "EthLend", "0x80fB784B7eD66730e8b1DBd9820aFD29931aab03", 18, true, false, common.GetTimepoint()),
		"ELEC":  common.NewToken("ELEC", "ElectrifyAsia", "0xd49ff13661451313ca1553fd6954bd1d9b6e02b9", 18, true, false, common.GetTimepoint()),
		"PAY":   common.NewToken("PAY", "TenX", "0xB97048628DB6B661D4C2aA833e95Dbe1A905B280", 18, true, true, common.GetTimepoint()),
		"RDN":   common.NewToken("RDN", "Raiden", "0x255aa6df07540cb5d3d297f0d0d4d84cb52bc8e6", 18, true, true, common.GetTimepoint()),
		"STORM": common.NewToken("STORM", "Storm", "0xd0a4b8946cb52f0661273bfbc6fd0e0c75fc6433", 18, true, false, common.GetTimepoint()),
		"MDS":   common.NewToken("MDS", "MediShares", "0x66186008C1050627F979d464eABb258860563dbE", 18, true, false, common.GetTimepoint()),
		"MANA":  common.NewToken("MANA", "MANA", "0x0f5d2fb29fb7d3cfee444a200298f468908cc942", 18, true, true, common.GetTimepoint()),
		"LBA":   common.NewToken("LBA", "Libra Credit", "0xfe5f141bf94fe84bc28ded0ab966c16b17490657", 18, true, true, common.GetTimepoint()),
		"WABI":  common.NewToken("WABI", "WaBi", "0x286BDA1413a2Df81731D4930ce2F862a35A609fE", 18, true, false, common.GetTimepoint()),
		"BAT":   common.NewToken("BAT", "Basic Attention Token", "0x0d8775f648430679a709e98d2b0cb6250d2887ef", 18, true, true, common.GetTimepoint()),
		"RCN":   common.NewToken("RCN", "Ripio Credit Network", "0xf970b8e36e23f7fc3fd752eea86f8be8d83375a6", 18, true, false, common.GetTimepoint()),
		"MOT":   common.NewToken("MOT", "Olympus Labs", "0x263c618480dbe35c300d8d5ecda19bbb986acaed", 18, true, false, common.GetTimepoint()),
		"SUB":   common.NewToken("SUB", "Substratum", "0x12480e24eb5bec1a9d4369cab6a80cad3c0a377a", 2, true, true, common.GetTimepoint()),
		"AION":  common.NewToken("AION", "Aion", "0x4CEdA7906a5Ed2179785Cd3A40A69ee8bc99C466", 8, true, true, common.GetTimepoint()),
		"EDU":   common.NewToken("EDU", "EduCoin", "0xf263292e14d9d8ecd55b58dad1f1df825a874b7c", 18, true, true, common.GetTimepoint()),
		"POWR":  common.NewToken("POWR", "Power Ledger", "0x595832f8fc6bf59c85c527fec3740a1b7a361269", 6, true, true, common.GetTimepoint()),
		"BQX":   common.NewToken("BQX", "Ethos", "0x5Af2Be193a6ABCa9c8817001F45744777Db30756", 8, true, true, common.GetTimepoint()),
		"DAI":   common.NewToken("DAI", "DAI", "0x89d24a6b4ccb1b6faa2625fe562bdd9a23260359", 18, true, false, common.GetTimepoint()),
	},
	common.RopstenMode: map[string]common.Token{
		"SNT":  common.NewToken("SNT", "STATUS", "0xF739577d63cdA4a534B0fB92ABf8BBf6EA48d36c", 18, false, false, common.GetTimepoint()),
		"BAT":  common.NewToken("BAT", "Basic Attention Token", "0x04A34c8f5101Dcc50bF4c64D1C7C124F59bb988c", 18, false, false, common.GetTimepoint()),
		"BITX": common.NewToken("BITX", "BitScreenerToken", "0x7a17267576318efb728bc4a0833e489a46ba138f", 0, false, false, common.GetTimepoint()),
		"ETH":  common.NewToken("ETH", "Ethereum", "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee", 18, false, false, common.GetTimepoint()),
		"OMG":  common.NewToken("OMG", "OmiseGO", "0x5b9a857e0C3F2acc5b94f6693536d3Adf5D6e6Be", 18, false, false, common.GetTimepoint()),
		"KNC":  common.NewToken("KNC", "KyberNetwork", "0xE5585362D0940519d87d29362115D4cc060C56B3", 18, false, false, common.GetTimepoint()),
		"EOS":  common.NewToken("EOS", "Eos", "0xd3c64BbA75859Eb808ACE6F2A6048ecdb2d70817", 18, false, false, common.GetTimepoint()),
		"ELF":  common.NewToken("ELF", "AELF", "0x7174FCb9C2A49c027C9746983D8262597b5EcCb1", 18, false, false, common.GetTimepoint()),
		"MOC":  common.NewToken("MOC", "Moss Coin", "0x1742c81075031b8f173d2327e3479d1fc3feaa76", 0, false, false, common.GetTimepoint()),
		"GTO":  common.NewToken("GTO", "Gifto", "0x6B07b8360832c6bBf05A39D9d443A705032bDc4d", 5, false, false, common.GetTimepoint()),
		"COFI": common.NewToken("COFI", "ConFi", "0xb91786188f8d4e35d6d67799e9f162587bf4da03", 18, false, false, common.GetTimepoint()),
		"POWR": common.NewToken("POWR", "Power Ledger", "0x2C4EfAa21f09c3C6EEF0Edb001E9bffDE7127D3B", 6, false, false, common.GetTimepoint()),
		"MANA": common.NewToken("MANA", "MANA", "0xf5E314c435B3B2EE7c14eA96fCB3307C3a3Ef608", 18, false, false, common.GetTimepoint()),
		"REQ":  common.NewToken("REQ", "Request", "0xa448cD1DB463ae738a171C483C56157d6B83B97f", 18, false, false, common.GetTimepoint()),
	},
}

func mustGetConfigToken(kyberEnv string) map[string]common.Token {
	result, avail := TokenConfigs[kyberEnv]
	if avail {
		return result
	}
	if kyberEnv == common.SimulationMode {
		result, err := loadTokenFromFile(simSettingPath)
		if err != nil {
			log.Panicf("cannot load data from pre-defined simluation setting file, err: %v", err)
		}
		return result

	}
	if kyberEnv == common.ProductionMode {
		return TokenConfigs[common.MainnetMode]
	}
	log.Panicf("cannot get token Config for mode %s", kyberEnv)
	return nil
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
