# Set feed configuration

## Create set feed configuration

```shell
curl -X POST "https://gateway.local/v3/setting-change-feed-configuration" \
-H 'Content-Type: application/json' \
-d '{
    "change_list": [
        {
            "type": "set_feed_configuration",
            "data" : {
              "name": "geminiETHUSD",
              "set_rate": "usd_feed"
              "enabled": true,
              "base_volatility_spread": 1.1,
              "normal_spread": 1.2
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

`POST https://gateway.local/v3/setting-change-feed-configuration`

Params | Type | Required | Default | Description
------ | ---- | -------- | ------- | -----------
name | string | true | nil | name of feed will be updated
enabled | bool | false | nil | 
base_volatility_spread | float64 | false | nil | 
normal_spread | float64 | false | nil |  
<aside class="notice">Manage key is required</aside>

## Get pending set feed configuration


```shell
curl -X GET "https://gateway.local/v3/setting-change-feed-configuration"
```

> sample response

```json
{
  "data": [
    {
      "id": 6,
      "created": "2019-08-13T07:25:49.869418Z",
      "change_list":{
        "type": "set_feed_configuration",
        "data": {
          "name": "geminiETHUSD",
          "set_rate": "usd_feed",
          "enabled": true,
          "base_volatility_spread": 1.1,
          "normal_spread": 1.2
        }
      }
    }
  ],
  "success": true
}
```

### HTTP Request

`GET https://gateway.local/v3/setting-change-feed-configuration`
<aside class="notice">All keys are accepted</aside>

## Confirm pending set feed configuration

```shell
curl -X PUT "https://gateway.local/v3/setting-change-feed-configuration/6"
```

> sample response

```json
{
  "success": true
}
```

### HTTP Request

`PUT https://gateway.local/v3/setting-change-feed-configuration/:change_id`
<aside class="notice">Confirm key is required</aside>

## Reject pending set feed configuration

```shell
curl -X DELETE "https://gateway.local/v3/setting-change-feed-configuration/6"
```

> sample response

```json
{
  "success": true
}
```

### HTTP Request

`DELETE https://gateway.local/v3/setting-change-feed-configuration/:change_id`
<aside class="notice">Confirm key is required</aside>

## Update feed status

```shell
curl -X PUT "https://gateway.local/v3/update-feed-status" \
-H 'Content-Type: application/json' \
-d '{"enabled": true, "name: "geminiETHUSD", "set_rate": "usd_feed"}'
```

> sample response

```json
{
  "name": "geminiETHUSD",
  "success": true,
  "set_rate": "usd_feed"
}
```

### HTTP Request

`PUT https://gateway.local/v3/update-feed-status/:name`
<aside class="notice">Manage key is required</aside>