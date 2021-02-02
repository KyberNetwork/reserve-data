# Setting change pwis

## Create setting pwis

```shell
curl -X POST "https://gateway.local/v3/setting-change-pwis" \
-H 'Content-Type: application/json' \
-d '{
    "change_list": [
        {
         "type":"update_asset",
         "data":{
            "asset_id":3,
            "pwi":{
               "ask":{
                  "a":6,
                  "b":12,
                  "c":0,
                  "min_min_spread":0.005,
                  "price_multiply_factor":1
               },
               "bid":{
                  "a":8,
                  "b":16,
                  "c":0,
                  "min_min_spread":0.005,
                  "price_multiply_factor":1
               }
            }
         }
      },
      ...
    ]
}'
```

> sample response

```json
{
  "id": 6,
  "success": true
}
```

### HTTP Request

`POST https://gateway.local/v3/setting-change-pwis`

Params | Type | Required | Default | Description
------ | ---- | -------- | ------- | -----------
asset_id | uint64 | true | nil | ID of asset
a | float64 | false | nil | 
b | float64 | false | nil | 
c | float64 | false | nil | 
min_min_spread | float64 | false | nil | 
price_multiply_factor | float64 | false | nil | 
<aside class="notice">Manage key is required</aside>

## Get pending setting pwis


```shell
curl -X GET "https://gateway.local/v3/setting-change-pwis"
```

> sample response

```json
{
  "data": [
    {
      "id": 6,
      "created": "2019-08-13T07:25:49.869418Z",
      "change_list": [
        {
          "type":"update_asset",
          "data":{
              "asset_id":3,
              "pwi":{
                "ask":{
                    "a":6,
                    "b":12,
                    "c":0,
                    "min_min_spread":0.005,
                    "price_multiply_factor":1
                },
                "bid":{
                    "a":8,
                    "b":16,
                    "c":0,
                    "min_min_spread":0.005,
                    "price_multiply_factor":1
                }
              }
          }
        },
        ...
      ]
    }
  ],
  "success": true
}
```

### HTTP Request

`GET https://gateway.local/v3/setting-change-pwis`
<aside class="notice">All keys are accepted</aside>

## Confirm pending setting pwis

```shell
curl -X PUT "https://gateway.local/v3/setting-change-pwis/6"
```

> sample response

```json
{
  "success": true
}
```

### HTTP Request

`PUT https://gateway.local/v3/setting-change-pwis/:change_id`
<aside class="notice">Confirm key is required</aside>

## Reject pending setting pwis

```shell
curl -X DELETE "https://gateway.local/v3/setting-change-pwis/6"
```

> sample response

```json
{
  "success": true
}
```

### HTTP Request

`DELETE https://gateway.local/v3/setting-change-pwis/:change_id`
<aside class="notice">Confirm key is required</aside>