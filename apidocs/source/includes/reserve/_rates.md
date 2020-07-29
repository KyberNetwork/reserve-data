# Rates and Prices

## Get prices for all base-quote pairs

```shell
curl -X -GET "http://gateway.local/v3/prices"
```

> sample response

```json
{
    "block": 9290920,
    "data": [
        {
            "base":     3,
            "quote":    1,
            "exchange": 1,
            "bids": [
                {
                    "quantity": 31,
                    "rate":     0.00123,
                },
                ...
            ],
            "asks": [
                {
                    "quantity": 31,
                    "rate":     0.00123,
                },
                ...
            ]
        }
    ],
    "success": true,
    "timestamp": "1514114582015",
    "version": 64
}
```

### HTTP Request

`GET https://gateway.local/v3/prices`

## Get prices for a specific base-quote pair

```shell
curl -X GET "https://gateway.local/prices/omg/eth"
```

> sample response


### HTTP Request

`GET https://gateway.local/v3/prices/:base/:quote`

## Get token rates from blockchain

```shell
curl -X GET "http://gateway.local/v3/getrates"
```

> sample response

```json
{
    "data": {
        "3": {
            "Timestamp": "1579161010159",
            "ReturnTime": "1579161010414",
            "BaseBuy": 3.2992994010339163,
            "CompactBuy": -32,
            "BaseSell": 0.2994836524507138,
            "CompactSell": 35,
            "Rate": 0,
            "Block": 9290940,
        },
    },
    "success": true,
    "timestamp": "1515412583215",
    "version": 1515412582435
}
```

### HTTP Request

`GET http://gateway.local/v3/getrates`


## Get all token rates from blockchain

```shell
curl -X GET "https://gateway.local/v3/get-all-rates"
```

> sample response

```json
{
    "data": [
        {
            "Version": 0,
            "Timestamp": "1517280618739",
            "ReturnTime": "1517280619071",
            "Data": {
                "3": {
                    "Timestamp": "1517280618739",
                    "ReturnTime": "1517280619071",
                    "BaseBuy": 4053.2170631085987,
                    "CompactBuy": 43,
                    "BaseSell": 0.000233599514875301,
                    "CompactSell": -3,
                    "Rate": 0,
                    "Block": 5635245
                }
            }
        },
        {
            "Version": 0,
            "Timestamp": "1517280621738",
            "ReturnTime": "1517280622251",
            "Data": {
                "1": {
                    "Timestamp": "1517280621738",
                    "ReturnTime": "1517280622251",
                    "BaseBuy": 87.21360760013062,
                    "CompactBuy": 0,
                    "BaseSell": 0.0128686459657361,
                    "CompactSell": 0,
                    "Rate": 0,
                    "Block": 5635245
                },
                "2": {
                    "Timestamp": "1517280621738",
                    "ReturnTime": "1517280622251",
                    "BaseBuy": 0,
                    "CompactBuy": 32,
                    "BaseSell": 0,
                    "CompactSell": -14,
                    "Rate": 0,
                    "Block": 5635245
                },
                "3": {
                    "Timestamp": "1517280621738",
                    "ReturnTime": "1517280622251",
                    "BaseBuy": 307.05930436561505,
                    "CompactBuy": -34,
                    "BaseSell": 0.003084981280661941,
                    "CompactSell": 81,
                    "Rate": 0,
                    "Block": 5635245
                },
                "4": {
                    "Timestamp": "1517280621738",
                    "ReturnTime": "1517280622251",
                    "BaseBuy": 65.0580993582104,
                    "CompactBuy": 32,
                    "BaseSell": 0.014925950060437398,
                    "CompactSell": -14,
                    "Rate": 0,
                    "Block": 5635245
                },
                "5": {
                    "Timestamp": "1517280621738",
                    "ReturnTime": "1517280622251",
                    "BaseBuy": 152.3016783627643,
                    "CompactBuy": 9,
                    "BaseSell": 0.006196212698403499,
                    "CompactSell": 23,
                    "Rate": 0,
                    "Block": 5635245
                },
                "6": {
                    "Timestamp": "1517280621738",
                    "ReturnTime": "1517280622251",
                    "BaseBuy": 4053.2170631085987,
                    "CompactBuy": 43,
                    "BaseSell": 0.000233599514875301,
                    "CompactSell": -3,
                    "Rate": 0,
                    "Block": 5635245
                }
            }
        }
    ],
    "success": true
}
```

### HTTP Request

`GET https://gateway.local/v3/get-all-rates`

Params | Type | Required | Default | Description
------ | ---- | -------- | ------- | -----------
fromTime | uint64 | false | 24 hours ago | fromTime to get all rates
toTime | uint64 | false | current time | toTime to get all rates

## Set Rate

Set rate create a contract call to set rate into conversion rate contract.

``` shell
curl -X POST "https://gateway.local/v3/setrates" \
-H 'Content-Type: application/json' \
-d '{
        "block": 9000000,
        "rates": [
            {
            "asset_id": 1
            "buy": "122",
            "sell": "123",
            "mid": "123",
            "msg": "set rate"
            "trigger": true
            }
        ]
    }'
```

> sample response

```json
{
  "id": "rate-rate-id",
  "success": true
}
```

### HTTP Request

`POST https://gateway.local/v3/setrates`

Params | Type | Required | Default | Description
------ | ---- | -------- | ------- | -----------
block | int64 | yes |  | block that use to validate rate
rates | []RateRequest | yes |  | a list of set rate request

#### RateRequest

Field | Type | Required | Default | Description
------ | ---- | -------- | ------- | -----------
asset_id | int64 | yes |  | asset id
buy | big number | yes |  | buy rate
sell | big number | yes |  | sell rate
mid | big number | yes |  | mid value
msg | big number | yes |  | a string will be attach into activity record, general purpose
trigger | bool | yes |  | mark set rate on asset as trigger


<aside class="notice">Rebalance key is required</aside>
