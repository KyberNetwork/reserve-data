## create pending update trading pair

```shell
curl -X POST "https://gateway.local/v3/setting-change-tpair" \
-H 'Content-Type: application/json' \
-d '{
    "change_list": [{
        "type": "update_trading_pair",
        "data": {
            "trading_pair_id": 1,
            "stale_threshold": 45000
        }
    }]
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

`POST https://gateway.local/v3/setting-change-tpair`
<aside class="notice">Manage key is required</aside>

### Data fields:

Params | Type | Required | Default | Description
------ | ---- | -------- | ------- | -----------
trading_pair_id | int | true | nil | id of trading pair
stale_threshold | float | false | nil | the new stale threshold value


## Get pending setting trading pair


```shell
curl -X GET "https://gateway.local/v3/setting-change-tpair"
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
          "type": "update_trading_pair",
          "data": {
            "trading_pair_id": 1,
            "stale_threshold": 45000
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

`GET https://gateway.local/v3/setting-change-tpair`
<aside class="notice">All keys are accepted</aside>

## Confirm pending setting tpair

```shell
curl -X PUT "https://gateway.local/v3/setting-change-tpair/6"
```

> sample response

```json
{
  "success": true
}
```

### HTTP Request

`PUT https://gateway.local/v3/setting-change-tpair/:change_id`
<aside class="notice">Manage key is required</aside>

## Reject pending setting trading pair

```shell
curl -X DELETE "https://gateway.local/v3/setting-change-tpair/6"
```

> sample response

```json
{
  "success": true
}
```

### HTTP Request

`DELETE https://gateway.local/v3/setting-change-tpair/:change_id`
<aside class="notice">Manage key is required</aside>