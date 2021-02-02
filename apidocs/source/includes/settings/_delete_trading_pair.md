## Pending delete trading pair

```shell
curl -X POST "https://gateway.local/v3/setting-change-main" \
-H 'Content-Type: application/json' \
-d '{
    "change_list": [{
        "type": "delete_trading_pair",
        "data": {
            "id": 1
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

`POST https://gateway.local/v3/setting-change-main`
<aside class="notice">Manage key is required</aside>

### Data fields:

Params | Type | Required | Default | Description
------ | ---- | -------- | ------- | -----------
id | int | true | nil | id of trading pair