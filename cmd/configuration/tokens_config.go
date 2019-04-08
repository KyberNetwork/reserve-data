package configuration

import (
	"github.com/KyberNetwork/reserve-data/common"
)

//TokenConfigs store token configs according to env mode.
var TokenConfigs = map[string]string{
	common.DevMode: `
{
  "tokens": {
    "OMG": {
      "name": "OmiseGO",
      "decimals": 18,
      "address": "0xd26114cd6EE289AccF82350c8d8487fedB8A0C07",
      "minimalRecordResolution": "1000000000000000",
      "maxPerBlockImbalance": "439794468212403470336",
      "maxTotalImbalance": "722362414038872621056",
      "internal use": true,
      "listed": true
    },
    "KNC": {
      "name": "KyberNetwork",
      "decimals": 18,
      "address": "0xdd974D5C2e2928deA5F71b9825b8b646686BD200",
      "minimalRecordResolution": "1000000000000000",
      "maxPerBlockImbalance": "3475912029567568052224",
      "maxTotalImbalance": "5709185508564730380288",
      "internal use": true,
      "listed": true
    },
    "EOS": {
      "name": "Eos",
      "decimals": 18,
      "address": "0x86Fa049857E0209aa7D9e616F7eb3b3B78ECfdb0",
      "minimalRecordResolution": "1000000000000000",
      "maxPerBlockImbalance": "938890140546807627776",
      "maxTotalImbalance": "1542127055848131526656",
      "internal use": false,
      "listed": true
    },
    "SNT": {
      "name": "STATUS",
      "decimals": 18,
      "address": "0x744d70fdbe2ba4cf95131626614a1763df805b9e",
      "minimalRecordResolution": "10000000000000000",
      "maxPerBlockImbalance": "43262133595415336976384",
      "maxTotalImbalance": "52109239915677776609280",
      "internal use": true,
      "listed": true
    },
    "ELF": {
      "name": "AELF",
      "decimals": 18,
      "address": "0xbf2179859fc6d5bee9bf9158632dc51678a4100e",
      "minimalRecordResolution": "1000000000000000",
      "maxPerBlockImbalance": "5906192156691986907136",
      "maxTotalImbalance": "7114008452735498715136",
      "internal use": true,
      "listed": true
    },
    "POWR": {
      "name": "Power Ledger",
      "decimals": 6,
      "address": "0x595832f8fc6bf59c85c527fec3740a1b7a361269",
      "minimalRecordResolution": "1000",
      "maxPerBlockImbalance": "7989613502",
      "maxTotalImbalance": "7989613502",
      "internal use": true,
      "listed": true
    },
    "MANA": {
      "name": "MANA",
      "decimals": 18,
      "address": "0x0f5d2fb29fb7d3cfee444a200298f468908cc942",
      "minimalRecordResolution": "1000000000000000",
      "maxPerBlockImbalance": "46289152908501773713408",
      "maxTotalImbalance": "46289152908501773713408",
      "internal use": true,
      "listed": true
    },
    "BAT": {
      "name": "Basic Attention Token",
      "decimals": 18,
      "address": "0x0d8775f648430679a709e98d2b0cb6250d2887ef",
      "minimalRecordResolution": "1000000000000000",
      "maxPerBlockImbalance": "13641944431813013274624",
      "maxTotalImbalance": "13641944431813013274624",
      "internal use": true,
      "listed": true
    },
    "REQ": {
      "name": "Request",
      "decimals": 18,
      "address": "0x8f8221afbb33998d8584a2b05749ba73c37a938a",
      "minimalRecordResolution": "1000000000000000",
      "maxPerBlockImbalance": "27470469074054960644096",
      "maxTotalImbalance": "33088179999699195920384",
      "internal use": true,
      "listed": true
    },
    "GTO": {
      "name": "Gifto",
      "decimals": 5,
      "address": "0xc5bbae50781be1669306b9e001eff57a2957b09d",
      "minimalRecordResolution": "10",
      "maxPerBlockImbalance": "1200696404",
      "maxTotalImbalance": "1200696404",
      "internal use": true,
      "listed": true
    },
    "RDN": {
      "name": "Raiden",
      "decimals": 18,
      "address": "0x255aa6df07540cb5d3d297f0d0d4d84cb52bc8e6",
      "minimalRecordResolution": "1000000000000000",
      "maxPerBlockImbalance": "2392730983766020325376",
      "maxTotalImbalance": "2882044469946171260928",
      "internal use": true,
      "listed": true
    },
    "APPC": {
      "name": "AppCoins",
      "decimals": 18,
      "address": "0x1a7a8bd9106f2b8d977e08582dc7d24c723ab0db",
      "minimalRecordResolution": "1000000000000000",
      "maxPerBlockImbalance": "10010270788085346205696",
      "maxTotalImbalance": "12057371164248796823552",
      "internal use": true,
      "listed": true
    },
    "ENG": {
      "name": "Enigma",
      "decimals": 8,
      "address": "0xf0ee6b27b759c9893ce4f094b49ad28fd15a23e4",
      "minimalRecordResolution": "10000",
      "maxPerBlockImbalance": "288970915691",
      "maxTotalImbalance": "348065467950",
      "internal use": true,
      "listed": true
    },
    "SALT": {
      "name": "Salt",
      "decimals": 8,
      "address": "0x4156d3342d5c385a87d264f90653733592000581",
      "minimalRecordResolution": "10000",
      "maxPerBlockImbalance": "123819203326",
      "maxTotalImbalance": "123819203326",
      "internal use": true,
      "listed": true
    },
    "BQX": {
      "name": "Ethos",
      "decimals": 8,
      "address": "0x5Af2Be193a6ABCa9c8817001F45744777Db30756",
      "minimalRecordResolution": "10000",
      "maxPerBlockImbalance": "127620338002",
      "maxTotalImbalance": "127620338002",
      "internal use": true,
      "listed": true
    },
    "ADX": {
      "name": "AdEx",
      "decimals": 4,
      "address": "0x4470BB87d77b963A013DB939BE332f927f2b992e",
      "minimalRecordResolution": "10",
      "maxPerBlockImbalance": "1925452883",
      "maxTotalImbalance": "1925452883",
      "internal use": false,
      "listed": true
    },
    "AST": {
      "name": "AirSwap",
      "decimals": 4,
      "address": "0x27054b13b1b798b345b591a4d22e6562d47ea75a",
      "minimalRecordResolution": "10",
      "maxPerBlockImbalance": "1925452883",
      "maxTotalImbalance": "1925452883",
      "internal use": true,
      "listed": true
    },
    "RCN": {
      "name": "Ripio Credit Network",
      "decimals": 18,
      "address": "0xf970b8e36e23f7fc3fd752eea86f8be8d83375a6",
      "minimalRecordResolution": "10",
      "maxPerBlockImbalance": "1925452883",
      "maxTotalImbalance": "1925452883",
      "internal use": false,
      "listed": true
    },
    "ZIL": {
      "name": "Zilliqa",
      "decimals": 12,
      "address": "0x05f4a42e251f2d52b8ed15e9fedaacfcef1fad27",
      "minimalRecordResolution": "10",
      "maxPerBlockImbalance": "1925452883",
      "maxTotalImbalance": "1925452883",
      "internal use": true,
      "listed": true
    },
    "DAI": {
      "name": "DAI",
      "decimals": 18,
      "address": "0x89d24a6b4ccb1b6faa2625fe562bdd9a23260359",
      "minimalRecordResolution": "1000000000000000",
      "maxPerBlockImbalance": "2711997842670896021504",
      "maxTotalImbalance": "3833713935933528080384",
      "internal use": false,
      "listed": true
    },
    "LINK": {
      "name": "Chain Link",
      "decimals": 18,
      "address": "0x514910771af9ca656af840dff83e8264ecf986ca",
      "minimalRecordResolution": "10",
      "maxPerBlockImbalance": "1925452883",
      "maxTotalImbalance": "1925452883",
      "internal use": true,
      "listed": true
    },
    "IOST": {
      "name": "IOStoken",
      "decimals": 18,
      "address": "0xfa1a856cfa3409cfa145fa4e20eb270df3eb21ab",
      "internal use": false,
      "listed": true
    },
    "STORM": {
      "name": "Storm",
      "decimals": 18,
      "address": "0xd0a4b8946cb52f0661273bfbc6fd0e0c75fc6433",
      "internal use": false,
      "listed": true
    },
    "MOT": {
      "name": "Olympus Labs",
      "decimals": 18,
      "address": "0x263c618480dbe35c300d8d5ecda19bbb986acaed",
      "internal use": false,
      "listed": true
    },
    "DGX": {
      "name": "Digix Gold",
      "decimals": 9,
      "address": "0x4f3afec4e5a3f2a6a1a411def7d7dfe50ee057bf",
      "minimalRecordResolution": "1000000000000000",
      "maxPerBlockImbalance": "2711997842670896021504",
      "maxTotalImbalance": "3833713935933528080384",
      "internal use": true,
      "listed": true
    },
    "ABT": {
      "name": "ArcBlock",
      "decimals": 18,
      "address": "0xb98d4c97425d9908e66e53a6fdf673acca0be986",
      "minimalRecordResolution": "100000000000000",
      "maxPerBlockImbalance": "5461044951947117854720",
      "maxTotalImbalance": "6043192343824681664512",
      "internal use": true,
      "listed": true
    },
    "ENJ": {
      "name": "EnjinCoin",
      "decimals": 18,
      "address": "0xf629cbd94d3791c9250152bd8dfbdf380e2a3b9c",
      "minimalRecordResolution": "100000000000000",
      "maxPerBlockImbalance": "5461044951947117854720",
      "maxTotalImbalance": "6043192343824681664512",
      "internal use": true,
      "listed": true
    },
    "AION": {
      "name": "Aion",
      "decimals": 8,
      "address": "0x4CEdA7906a5Ed2179785Cd3A40A69ee8bc99C466",
      "internal use": true,
      "listed": true
    },
    "AE": {
      "name": "Aeternity",
      "decimals": 18,
      "address": "0x5ca9a71b1d01849c0a95490cc00559717fcf0d1d",
      "internal use": true,
      "listed": true
    },
    "BLZ": {
      "name": "Bluezelle",
      "decimals": 18,
      "address": "0x5732046a883704404f284ce41ffadd5b007fd668",
      "internal use": true,
      "listed": true
    },
    "PAL": {
      "name": "PolicyPal Network",
      "decimals": 18,
      "address": "0xfedae5642668f8636a11987ff386bfd215f942ee",
      "internal use": false,
      "listed": true
    },
    "ELEC": {
      "name": "ElectrifyAsia",
      "decimals": 18,
      "address": "0xd49ff13661451313ca1553fd6954bd1d9b6e02b9",
      "internal use": false,
      "listed": true
    },
    "BBO": {
      "name": "Bigbom",
      "decimals": 18,
      "address": "0x84f7c44b6fed1080f647e354d552595be2cc602f",
      "internal use": false,
      "listed": true
    },
    "POLY": {
      "name": "Polymath",
      "decimals": 18,
      "address": "0x9992ec3cf6a55b00978cddf2b27bc6882d88d1ec",
      "internal use": true,
      "listed": true
    },
    "LBA": {
      "name": "Libra Credit",
      "decimals": 18,
      "address": "0xfe5f141bf94fe84bc28ded0ab966c16b17490657",
      "internal use": true,
      "listed": true
    },
    "EDU": {
      "name": "EduCoin",
      "decimals": 18,
      "address": "0xf263292e14d9d8ecd55b58dad1f1df825a874b7c",
      "internal use": true,
      "listed": true
    },
    "CVC": {
      "name": "Civic",
      "decimals": 8,
      "address": "0x41e5560054824eA6B0732E656E3Ad64E20e94E45",
      "internal use": true,
      "listed": true
    },
    "WAX": {
      "name": "Wax",
      "decimals": 8,
      "address": "0x39bb259f66e1c59d5abef88375979b4d20d98022",
      "internal use": true,
      "listed": true
    },
    "SUB": {
      "name": "Substratum",
      "decimals": 2,
      "address": "0x12480e24eb5bec1a9d4369cab6a80cad3c0a377a",
      "internal use": true,
      "listed": true
    },
    "POE": {
      "name": "Po.et",
      "decimals": 8,
      "address": "0x0e0989b1f9b8a38983c2ba8053269ca62ec9b195",
      "internal use": true,
      "listed": true
    },
    "PAY": {
      "name": "TenX",
      "decimals": 18,
      "address": "0xB97048628DB6B661D4C2aA833e95Dbe1A905B280",
      "internal use": true,
      "listed": true
    },
    "CHAT": {
      "name": "Chatcoin",
      "decimals": 18,
      "address": "0x442bc47357919446eabc18c7211e57a13d983469",
      "internal use": true,
      "listed": true
    },
    "DTA": {
      "name": "Data",
      "decimals": 18,
      "address": "0x69b148395ce0015c13e36bffbad63f49ef874e03",
      "internal use": true,
      "listed": true
    },
    "BNT": {
      "name": "Bancor",
      "decimals": 18,
      "address": "0x1f573d6fb3f13d689ff844b4ce37794d79a7ff1c",
      "internal use": true,
      "listed": true
    },
    "TUSD": {
      "name": "TrueUSD",
      "decimals": 18,
      "address": "0x8dd5fbce2f6a956c3022ba3663759011dd51e73e",
      "internal use": true,
      "listed": true
    },
    "TOMO": {
      "name": "Tomocoin",
      "decimals": 18,
      "address": "0x8b353021189375591723E7384262F45709A3C3dC",
      "internal use": false,
      "listed": true
    },
    "MDS": {
      "name": "MediShares",
      "decimals": 18,
      "address": "0x66186008C1050627F979d464eABb258860563dbE",
      "internal use": false,
      "listed": true
    },
    "LEND": {
      "name": "EthLend",
      "decimals": 18,
      "address": "0x80fB784B7eD66730e8b1DBd9820aFD29931aab03",
      "internal use": false,
      "listed": true
    },
    "WINGS": {
      "name": "WINGS",
      "decimals": 18,
      "address": "0x667088b212ce3d06a1b553a7221E1fD19000d9aF",
      "internal use": false,
      "listed": true
    },
    "MTL": {
      "name": "Metal",
      "decimals": 8,
      "address": "0xF433089366899D83a9f26A773D59ec7eCF30355e",
      "internal use": false,
      "listed": true
    },
    "WABI": {
      "name": "WaBi",
      "decimals": 18,
      "address": "0x286BDA1413a2Df81731D4930ce2F862a35A609fE",
      "internal use": false,
      "listed": true
    },
    "COFI": {
      "name": "CoinFi",
      "decimal": 18,
      "address": "0x3136ef851592acf49ca4c825131e364170fa32b3",
      "internal use": false,
      "listed": true
    },
    "MOC": {
      "name": "Moss Land",
      "decimal": 18,
      "address": "0x865ec58b06bf6305b886793aa20a2da31d034e68",
      "internal use": false,
      "listed": true
    },
    "ETH": {
      "name": "Ethereum",
      "decimals": 18,
      "address": "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
      "internal use": true,
      "listed": true
    }
  }
}
`,
}
