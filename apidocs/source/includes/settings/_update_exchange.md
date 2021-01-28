# Update exchange

## Create update exchange

```shell
curl -X POST "https://gateway.local/v3/setting-change-update-exchange" \
-H 'Content-Type: application/json' \
-d '{
    "change_list": [
        {
            "type": "update_exchange",
            "data" : {
                 "exchange_id": 1,
                 "trading_fee_maker": 0,
                 "trading_fee_taker": 0,
                 "disable": true
            }
        }
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

`POST https://gateway.local/v3/setting-change-update-exchange`

Params | Type | Required | Default | Description
------ | ---- | -------- | ------- | -----------
exchange_id | int | true | nil | ID of exchange will be updated
trading_fee_maker | float64 | false | nil | 
trading_fee_taker | float64 | false | nil | 
disable | bool | false | nil |  
<aside class="notice">Manage key is required</aside>

## Get pending update exchange 


```shell
curl -X GET "https://gateway.local/v3/setting-change-update-exchange"
```

> sample response

```json
{
  "data": [
    {
      "id": 6,
      "created": "2019-08-13T07:25:49.869418Z",
      "change_list":{
        "type": "update_exchange",
        "data": {
          "exchange_id": 1,
          "trading_fee_maker": 0,
          "trading_fee_taker": 0,
          "disable": true
        }
      }
    }
  ],
  "success": true
}
```

### HTTP Request

`GET https://gateway.local/v3/setting-change-update-exchange`
<aside class="notice">All keys are accepted</aside>

## Confirm pending update exchange

```shell
curl -X PUT "https://gateway.local/v3/setting-change-update-exchange/1"
```

> sample response

```json
{
    "success": true
}
```

### HTTP Request

`PUT https://gateway.local/v3/setting-change-update-exchange/:change_id`
<aside class="notice">Confirm key is required</aside>

## Reject pending update exchange 

```shell
curl -X DELETE "https://gateway.local/v3/setting-change-update-exchange/1"
```

> sample response

```json
{
    "success": true
}
```

### HTTP Request

`DELETE https://gateway.local/v3/setting-change-update-exchange/:change_id`
<aside class="notice">Confirm key is required</aside>


## Enable/Disable exchange 

```shell
curl -X PUT "https://gateway.local/v3/set-exchange-enabled/:exchange_id" \
-H 'Content-Type: application/json' \
-d '{"disable": true}'
```

> sample response

```json
{
    "success": true
}
```

### HTTP Request

`PUT https://gateway.local/v3/update-exchange-status/:exchange_id`
<aside class="notice">Manage key is required</aside>
