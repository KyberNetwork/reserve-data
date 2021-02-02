# GasThreshold

## Set GasThreshold

the request must be application/json format.

``` shell
curl -X POST "https://gateway.local/v3/gas-threshold" \
-H 'Content-Type: application/json' \
-d '{
    "high": 10.0,
    "low": 5.345
    }'
```

> sample response

```json
{
  "id": 1,
  "success": true
}
```

### HTTP Request

`POST https://gateway.local/v3/gas-threshold`

Params | Type | Required | Default | Description
------ | ---- | -------- | ------- | -----------
high | float64 | yes |  | the high value
low | float64 | yes |  | the low value
<aside class="notice">Admin key is required</aside>

## Get GasThreshold


```shell
curl -X GET "https://gateway.local/v3/gas-threshold"
```

> sample response

```json
{
  "success": true,
  "data": {
    "gas_price": {
      "etherscan": {
        "value": {
          "fast": 117,
          "standard": 105,
          "slow": 71
        },
        "timestamp": 1600748883444
      },
      "ethgasstation": {
        "value": {
          "fast": 105,
          "standard": 97,
          "slow": 90
        },
        "timestamp": 1600748883447
      },
      "gasnow": {
        "value": {
          "fast": 91,
          "standard": 63,
          "slow": 61
        },
        "timestamp": 1600748883691
      }
    },
    "high": 150,
    "low": 120
  }
}
```

### HTTP Request

`GET https://gateway.local/v3/gas-threshold`
<aside class="notice">All keys are accepted</aside>