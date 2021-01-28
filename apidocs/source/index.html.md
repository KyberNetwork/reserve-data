---
title: API Reference

language_tabs: # must be one of https://git.io/vQNgJ
  - shell

toc_footers:
  - <a href='https://github.com/lord/slate'>Documentation Powered by Slate</a>

includes:
  - settings/apis
  - settings/setting_change_main
  - settings/create_asset
  - settings/update_asset
  - settings/change_asset_address
  - settings/create_asset_exchange
  - settings/update_asset_exchange
  - settings/delete_asset_exchange
  - settings/create_trading_pair
  - settings/delete_trading_pair
  - settings/update_exchange
  - settings/setting_change_pwis
  - settings/setting_change_rbquadratic
  - settings/set_feed_configuration
  - reserve/rates
  - settings/rate_trigger
  - exchanges/exchanges
  - exchanges/rebalance
  - stable-token-and-btc/apis
  - errors
  - settings/gas_threshold
  - settings/setting_change_tpair

search: true
---

# Introduction

# Authentication
Authentication follow: https://tools.ietf.org/html/draft-cavage-http-signatures-10

Required headers:

- **Digest**
- **Authorization**
- **Signature**
- **Nonce**

# APIs

## Get time server

```shell
curl -X GET "http://gateway.local/v3/timeserver"
```

> sample response

```json
{
  "data": "1517479497447",
  "success": true
}
```

### HTTP request

`GET https://gateway.local/v3/timeserver`

## Get auth data

```shell
curl -X GET "https://gateway.local/v3/authdata"
```

> sample response

```json
{
    "data": {
        "balances": [
            {
                "asset_id": 1,
                "exchanges": [
                    {
                        "exchange_id": 1,
                        "available": 99.01583652879484,
                        "locked": 0
                    },
                    {
                        "exchange_id": 2,
                        "available": 181.8038571,
                        "locked": 0
                    }
                ],
                "reserve": 1717.0532129686085,
                "error": ""
            },
            {
                "asset_id": 2,
                "exchanges": [
                    {
                        "exchange_id": 1,
                        "available": 11574.235393,
                        "locked": 0
                    }
                ],
                "reserve": 15360.30933492698,
                "error": ""
            },
            {
                "asset_id": 3,
                "exchanges": [
                    {
                        "exchange_id": 2,
                        "available": 4.602309,
                        "locked": 0
                    }
                ],
                "reserve": 2559.238495969894,
                "error": ""
            }
        ],
        "pending_activities": {
            "set_rates": [
                {
                    "id": "1553164292266910443|0x303b6541817b328cc00d627ddab452a146fb403b16f1fc6a8db6e36216fe54ab",
                    "params": {
                        "block": 7411639,
                        "afp_mid": [
                            165405861856343,
                            108819058000000000,
                            4377249572437696
                        ],
                        "buys": [
                            6.022698129896374e+21,
                            9154550619948383000,
                            227387262849142330000
                        ],
                        "sells": [
                            164217513027505,
                            108181761942374060,
                            4360506592823122
                        ],
                        "tokens": [
                            1,
                            2,
                            3
                        ]
                    },
                    "gasPrice": "8000000000",
                    "nonce": "345612",
                    "tx": "0x303b6541817b328cc00d627ddab452a146fb403b16f1fc6a8db6e36216fe54ab",
                    "mining_status": "submitted",
                    "error": ""
                }
            ],
            "withdraw": [
                {
                    "id": "1553164258017310762|721813701",
                    "params": {
                        "exchange_id": 1,
                        "amount": 13504.7761,
                        "timepoint": 1553164257651,
                        "token": 2
                    },
                    "exchange_tx": "721813701",
                    "tx": "0xc481ec82aa2b4b33a4fccf94ecd4bdd278af0b6d8f381463ba934bf6d66880e9",
                    "exchange_status": "submitted",
                    "mining_status": "",
                    "error": ""
                }
            ],
            "deposit": [
                {
                    "id": "1553164258017310762|721813701",
                    "params": {
                        "exchange_id": 1,
                        "amount": 3504.7761,
                        "timepoint": 1553164257651,
                        "token": 3
                    },
                    "exchange_tx": "721813701",
                    "tx": "0xc481ec82aa2b4b33a4fccf94ecd4bdd278af0b6d8f381463ba934bf6d66880e9",
                    "exchange_status": "",
                    "mining_status": "mined",
                    "error": ""
                }
            ],
            "trades": [
                {
                    "action": "trade",
                    "id": "1600254671157255169|106845707",
                    "eid": "",
                    "destination": "binance",
                    "params": {
                        "exchange": 1,
                        "amount": 32,
                        "timepoint": 1600254670913,
                        "type": "buy",
                        "base": "KNC",
                        "quote": "ETH",
                        "rate": 0.0031299,
                        "trading_pair_id": 123243
                    },
                    "result": {
                        "id": "106845707",
                        "remaining": 32
                    },
                    "exchange_status": "done",
                    "timestamp": "1600254670913",
                    "actionDetail": "buy KNC-ETH"
                }
            ]
        }
    },
    "success": true,
    "version": 1553164294136
}
```

### HTTP Request

`GET https://gateway.local/v3/authdata`

## Get all activities

```shell
curl -X GET "https://gateway.local/v3/activities?fromTime=1564889953000&toTime=1565235553000"
```

> sample response

```json
```

### HTTP Request

`GET https://gateway.local/v3/activities`

Param | Type | Required | Description
----- | ---- | -------- | -----------
fromTime | uint64 | true | fromTime to get activities
toTime | uint64 | true | toTime to get activities
