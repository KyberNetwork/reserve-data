#  Settings Read APIs

## Get asset by id

```shell
curl -X GET "https://gateway.local/v3/asset/1"
```

> sample response

```json
{
  "data": {
    "id": 1,
    "symbol": "ETH",
    "name": "Ethereum",
    "address": "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
    "decimals": 18,
    "transferable": true,
    "set_rate": "not_set",
    "rebalance": true,
    "is_quote": true,
    "is_enabled": true,
    "pwi": {
      "ask": {
        "a": 13.88888,
        "b": -0.11111,
        "c": 0,
        "min_min_spread": 0.005,
        "price_multiply_factor": 0.45
      },
      "bid": {
        "a": 13.88888,
        "b": -0.11111,
        "c": 0,
        "min_min_spread": 0.005,
        "price_multiply_factor": 0.45
      }
    },
    "rebalance_quadratic": {
      "size_a": 0.000001754386,
      "size_b": 0.0004894737,
      "size_c": 0.9995088,
      "price_a": 0.1234,
      "price_b": 0.1235,
      "price_c": 0.1236,
      "price_offset": 0
    },
    "exchanges": [
      {
        "id": 2,
        "asset_id": 1,
        "exchange_id": 1,
        "symbol": "ETH",
        "deposit_address": "0x0000000000000000000000000000000000000001",
        "min_deposit": 1000,
        "withdraw_fee": 0.5,
        "target_recommended": 1000,
        "target_ratio": 5,
        "trading_pairs": [
          {
            "id": 2,
            "base": 1,
            "quote": 2,
            "price_precision": 6,
            "amount_precision": 4,
            "amount_limit_min": 0,
            "amount_limit_max": 0,
            "price_limit_min": 0,
            "price_limit_max": 0,
            "min_notional": 0.02
          }
        ]
      },
      {
        "id": 1,
        "asset_id": 1,
        "exchange_id": 0,
        "symbol": "ETH",
        "deposit_address": "0x023ab1f7acaad1f7a01d3bfa4afd2ab575780090",
        "min_deposit": 1000,
        "withdraw_fee": 0.5,
        "target_recommended": 2000,
        "target_ratio": 15,
        "trading_pairs": [
          {
            "id": 4,
            "base": 3,
            "quote": 2,
            "price_precision": 8,
            "amount_precision": 0,
            "amount_limit_min": 1,
            "amount_limit_max": 90000000,
            "price_limit_min": 1e-8,
            "price_limit_max": 1000,
            "min_notional": 0.0001
          },
          {
            "id": 1,
            "base": 1,
            "quote": 2,
            "price_precision": 6,
            "amount_precision": 3,
            "amount_limit_min": 0.001,
            "amount_limit_max": 100000,
            "price_limit_min": 0.000001,
            "price_limit_max": 100000,
            "min_notional": 0.0001
          },
          {
            "id": 3,
            "base": 3,
            "quote": 1,
            "price_precision": 7,
            "amount_precision": 0,
            "amount_limit_min": 1,
            "amount_limit_max": 90000000,
            "price_limit_min": 0,
            "price_limit_max": 0,
            "min_notional": 0.01
          }
        ]
      }
    ],
    "target": {
      "total": 300,
      "reserve": 300,
      "rebalance_threshold": 300,
      "transfer_threshold": 300,
      "min_withdraw_threshold": 100
    },
    "stable_param": {
      "price_update_threshold": 0.2,
      "ask_spread": 34.0,
      "bid_spread": 0.1,
      "single_feed_max_spread": 0.4,
      "multiple_feeds_max_diff": 0.6
    },
    "zero_x": {
      "ref_eth_amount": 1.0,
      "eth_step": 1.1,
      "max_eth_size_buy": 1.2,
      "max_eth_size_sell": 1.3,
      "enabled": true
    },
    "created": "2019-07-31T07:50:28.604784Z",
    "updated": "2019-07-31T08:53:20.585862Z",
    "normal_update_per_period": 1.234,
    "max_imbalance_ratio": 3.456,
    "order_duration_millis": 20000
  },
  "success": true
}
```

### HTTP Request

`GET https://gateway.local/v3/asset/:asset_id`

## Get all assets

```shell
curl -X GET "https://gateway.local/v3/asset"
```

> sample response

```json
{
  "data": [
    {
      "id": 1,
      "symbol": "ETH",
      "name": "Ethereum",
      "address": "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
      "decimals": 18,
      "transferable": true,
      "set_rate": "not_set",
      "rebalance": true,
      "is_quote": true,
      "is_enabled": true,
      "pwi": {
        "ask": {
          "a": 13.88888,
          "b": -0.11111,
          "c": 0,
          "min_min_spread": 0.005,
          "price_multiply_factor": 0.45
        },
        "bid": {
          "a": 13.88888,
          "b": -0.11111,
          "c": 0,
          "min_min_spread": 0.005,
          "price_multiply_factor": 0.45
        }
      },
      "rebalance_quadratic": {
        "size_a": 0.000001754386,
        "size_b": 0.0004894737,
        "size_c": 0.9995088,
        "price_a": 0.1234,
        "price_b": 0.1235,
        "price_c": 0.1236,
        "price_offset": 0
      },
      "exchanges": [
        {
          "id": 2,
          "asset_id": 1,
          "exchange_id": 1,
          "symbol": "ETH",
          "deposit_address": "0x0000000000000000000000000000000000000001",
          "min_deposit": 1000,
          "withdraw_fee": 0.5,
          "target_recommended": 1000,
          "target_ratio": 5
        },
        {
          "id": 1,
          "asset_id": 1,
          "exchange_id": 0,
          "symbol": "ETH",
          "deposit_address": "0x023ab1f7acaad1f7a01d3bfa4afd2ab575780090",
          "min_deposit": 1000,
          "withdraw_fee": 0.5,
          "target_recommended": 2000,
          "target_ratio": 15
        }
      ],
      "target": {
        "total": 300,
        "reserve": 300,
        "rebalance_threshold": 300,
        "transfer_threshold": 300,
        "min_withdraw_threshold": 100
      },
      "stable_param": {
        "price_update_threshold": 0.2,
        "ask_spread": 34.0,
        "bid_spread": 0.1,
        "single_feed_max_spread": 0.4,
        "multiple_feeds_max_diff": 0.6
      },
      "zero_x": {
        "ref_eth_amount": 1.0,
        "eth_step": 1.1,
        "max_eth_size_buy": 1.2,
        "max_eth_size_sell": 1.3,
        "enabled": true
      },
      "normal_update_per_period": 1.234,
      "max_imbalance_ratio": 3.456,
      "order_duration_millis": 20000,
      "created": "2019-07-31T07:50:28.604784Z",
      "updated": "2019-07-31T08:53:20.585862Z"
    },
    {
      "id": 2,
      "symbol": "BTC",
      "name": "Bitcoin",
      "address": "0x0000000000000000000000000000000000000000",
      "decimals": 8,
      "transferable": false,
      "set_rate": "not_set",
      "rebalance": true,
      "is_quote": true,
      "is_enabled": true,
      "rebalance_quadratic": {
        "size_a": 0.000001754386,
        "size_b": 0.0004894737,
        "size_c": 0.9995088,
        "price_a": 0.1234,
        "price_b": 0.1235,
        "price_c": 0.1236,
        "price_offset": 0
      },
      "exchanges": [
        {
          "id": 3,
          "asset_id": 2,
          "exchange_id": 0,
          "symbol": "BTC",
          "deposit_address": "0x0000000000000000000000000000000000000000",
          "min_deposit": 0,
          "withdraw_fee": 0.01,
          "target_recommended": 10,
          "target_ratio": 5
        },
        {
          "id": 4,
          "asset_id": 2,
          "exchange_id": 1,
          "symbol": "BTC",
          "deposit_address": "0x0000000000000000000000000000000000000000",
          "min_deposit": 0,
          "withdraw_fee": 0.01,
          "target_recommended": 10,
          "target_ratio": 5
        }
      ],
      "target": {
        "total": 10,
        "reserve": 6,
        "rebalance_threshold": 0.33,
        "transfer_threshold": 0.25,
        "min_withdraw_threshold": 1.0
      },
      "stable_param": {
        "price_update_threshold": 0,
        "ask_spread": 0,
        "bid_spread": 0,
        "single_feed_max_spread": 0,
        "multiple_feeds_max_diff": 0
      },
      "zero_x": {
        "ref_eth_amount": 1.0,
        "eth_step": 1.1,
        "max_eth_size_buy": 1.2,
        "max_eth_size_sell": 1.3,
        "enabled": true
      },
      "normal_update_per_period": 1.234,
      "max_imbalance_ratio": 3.456,
      "order_duration_millis": 20000,
      "created": "2019-07-31T08:53:33.653734Z",
      "updated": "2019-07-31T08:53:33.653734Z"
    }
  ],
  "success": true
}
```

### HTTP Request

`GET https://gateway.local/v3/asset`


## Get exchange by id

```shell
curl -X GET "https://gateway.local/v3/exchange/0"
```

> sample response

```json
{
    "data": {
        "id": 0,
        "name": "binance",
        "trading_fee_maker": 0,
        "trading_fee_taker": 0,
        "disable": true
    },
    "success": true
}
```

### HTTP Request

`GET https://gateway.local/v3/exchange/:exchange_id`


## Get all exchanges

```shell
curl -X GET "https://gateway.local/v3/exchange"
```

> sample response

```json
{
    "data": [
        {
            "id": 0,
            "name": "binance",
            "trading_fee_maker": 0,
            "trading_fee_taker": 0,
            "disable": true
        },
        {
            "id": 1,
            "name": "huobi",
            "trading_fee_maker": 0,
            "trading_fee_taker": 0,
            "disable": true
        },
        {
            "id": 2,
            "name": "stable_exchange",
            "trading_fee_maker": 1,
            "trading_fee_taker": 1,
            "disable": false
        }
    ],
    "success": true
}
```

### HTTP Request

`GET https://gateway.local/v3/exchange`

## Get trading pair by id

Params | Type | Required | Default | Description
------ | ---- | -------- | ------- | -----------
including_deleted | bool | false | false | include deleted object in result


```shell
curl -X GET "https://gateway.local/v3/trading-pair/1?including_deleted=false"
```

> sample response

```json
{
    "data": {
        "id": 1,
        "base": 1,
        "quote": 2,
        "price_precision": 6,
        "amount_precision": 3,
        "amount_limit_min": 0.001,
        "amount_limit_max": 100000,
        "price_limit_min": 0.000001,
        "price_limit_max": 100000,
        "min_notional": 0.0001,
        "base_symbol": "ETH",
        "quote_symbol": "BTC"
    },
    "success": true
}
```

### HTTP Request

`GET https://gateway.local/v3/trading-pair/:trading_pair_id`

## Get all trading pairs by exchange id

Params | Type | Required | Default | Description
------ | ---- | -------- | ------- | -----------
exchange_id| integer | true | none | exchange id 


```shell
curl -X GET "https://gateway.local/v3/trading-pair?exchange_id=1"
```

> sample response

```json
{
    "data": [
        {
            "id": 1,
            "base": 1,
            "quote": 2,
            "price_precision": 6,
            "amount_precision": 3,
            "amount_limit_min": 0.001,
            "amount_limit_max": 100000,
            "price_limit_min": 0.000001,
            "price_limit_max": 100000,
            "min_notional": 0.0001,
            "base_symbol": "ETH",
            "quote_symbol": "BTC"
        },
        {
            "id": 2,
            "base": 1,
            "quote": 3,
            "price_precision": 6,
            "amount_precision": 3,
            "amount_limit_min": 0.001,
            "amount_limit_max": 100000,
            "price_limit_min": 0.000001,
            "price_limit_max": 100000,
            "min_notional": 0.0001,
            "base_symbol": "ETH",
            "quote_symbol": "KNC"
        }
    ],
    "success": true
}
```

### HTTP Request

`GET https://gateway.local/v3/trading-pair?exchange_id=1`

## Get feed configurations

```shell
curl -X GET "https://gateway.local/v3/feed-configurations"
```

> sample response

```json
{
    "data": [
        {
            "name": "geminiETHUSD",
            "set_rate": "usd_feed",
            "enabled": true,
            "base_volatility_spread": 0,
            "normal_spread": 0
        },
        ...
    ],
    "success": true
}
```
### HTTP Request

`GET https://gateway.local/v3/feed-configurations`
