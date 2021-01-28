## Pending update asset exchange

```shell
curl -X POST "https://gateway.local/v3/setting-change-main" \
-H 'Content-Type: application/json' \
-d '{
    "change_list": [{
        "type": "update_asset_exchange",
        "data": {
            "id": 1,
            "symbol": "KNC",
            "deposit_address": "0xC5094e852F71346dF9eD6795Be6F014994C43e09",
            "min_deposit": 0,
            "withdraw_fee": 0,
            "target_recommended": 0,
            "target_ratio": 0
        }
    }]
}'
```

> sample response

```json
{
  "id": 2,
  "success": true
}
```

### HTTP Request

`POST https://gateway.local/v3/setting-change-main`
<aside class="notice">Manage key is required</aside>
<aside class="warning">
Constraints:<br>
- if asset has "transferable" = true => "deposit_address" != nil<br>
- "trading_pairs" child obj must be valid with trading pair contraints<br>
</aside>

### Data fields:

Params | Type | Required | Default | Description
------ | ---- | -------- | ------- | -----------
id | int | true | nil | id of asset exchange will be updated
symbol | string | false | nil |
deposit_address | string | false | nil |
min_deposit | float64 | false | nil |
withdraw_fee | float64 | false | nil |
target_recommended | float64 | false | nil |
target_ratio | float64 | false | nil |