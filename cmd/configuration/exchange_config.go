package configuration

import (
	"github.com/KyberNetwork/reserve-data/common"
)

//ExchangeConfigs store exchange config according to env mode.
var ExchangeConfigs = map[string]string{
	common.DevMode: `{
  "exchanges": {
    "binance": {
      "ETH": "0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90",
      "OMG": "0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90",
      "KNC": "0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90",
      "SNT": "0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90",
      "ELF": "0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90",
      "POWR": "0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90",
      "MANA": "0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90",
      "BAT": "0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90",
      "REQ": "0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90",
      "GTO": "0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90",
      "ENG": "0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90",
      "SALT": "0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90",
      "APPC": "0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90",
      "RDN": "0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90",
      "BQX": "0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90",
      "ZIL": "0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90",
      "AST": "0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90",
      "LINK": "0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90",
      "ENJ": "0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90",
      "AION": "0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90",
      "AE": "0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90",
      "BLZ": "0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90",
      "SUB": "0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90",
      "POE": "0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90",
      "CHAT": "0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90",
      "BNT": "0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90",
      "TUSD": "0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"
    },
    "huobi": {
      "ETH": "0x0c8fd73eaf6089ef1b91231d0a07d0d2ca2b9d66",
      "ABT": "0x0c8fd73eaf6089ef1b91231d0a07d0d2ca2b9d66",
      "POLY": "0x0c8fd73eaf6089ef1b91231d0a07d0d2ca2b9d66",
      "LBA": "0x0c8fd73eaf6089ef1b91231d0a07d0d2ca2b9d66",
      "EDU": "0x0c8fd73eaf6089ef1b91231d0a07d0d2ca2b9d66",
      "CVC": "0x0c8fd73eaf6089ef1b91231d0a07d0d2ca2b9d66",
      "WAX": "0x0c8fd73eaf6089ef1b91231d0a07d0d2ca2b9d66",
      "PAY": "0x0c8fd73eaf6089ef1b91231d0a07d0d2ca2b9d66",
      "DTA": "0x0c8fd73eaf6089ef1b91231d0a07d0d2ca2b9d66"
    },
    "stable_exchange": {
      "ETH": "0xFDF28Bf25779ED4cA74e958d54653260af604C20",
      "DGX": "0xFDF28Bf25779ED4cA74e958d54653260af604C20"
    }
  }
}`,
}
