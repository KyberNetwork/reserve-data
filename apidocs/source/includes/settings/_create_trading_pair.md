## Pending create trading pair

```shell
curl -X POST "https://gateway.local/v3/setting-change-main" \
-H 'Content-Type: application/json' \
-d '{
    "change_list": [{
        "type": "create_trading_pair",
        "data": {
            "base": 1,
            "quote": 3,
            "asset_id": 1,
            "exchange_id": 1
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
<aside class="warning">
Constraints:<br>
- quote asset must have field "is_quote" = true<br>
- quote and base asset must have asset_exchange with correspond exchange_id<br>
- trading by asset must be ether base asset or quote asset<br>
</aside>

### Data fields:

Params | Type | Required | Default | Description
------ | ---- | -------- | ------- | -----------
exchange_id | int | true | nil | id of exchange
base | int | true | nil | id of base asset
quote | int | true | nil | id of quote asset
asset_id | int | true | nil | id of trading by asset 