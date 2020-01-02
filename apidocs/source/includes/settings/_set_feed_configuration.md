# Set Feed Configuration

## Create set feed configuration

```shell
curl -X POST "https://gateway.local/v3/setting-change-feed-configuration" \
-H 'Content-Type: application/json' \
-d '{
    "change_list": [
        {
            "type": "set_feed_configuration",
            "data" : {
              "name": "DGX",
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
<aside class="notice">Write key is required</aside>

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
          "name": "DGX",
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