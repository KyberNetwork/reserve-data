##Pending create asset exchange

```shell
curl -X POST "https://gateway.local/v3/setting-change-main" \
-H 'Content-Type: application/json' \
-d '{
    "change_list": [{
        "type": "create_asset_exchange",
        "data": {
            "asset_id": 3,
            "exchange_id": 1,
            "symbol": "KNC",
            "deposit_address": "",
            "min_deposit": 0,
            "withdraw_fee": 0,
            "target_recommended": 0,
            "target_ratio": 0,
            "trading_pairs" : [{
                "id": 2,
                "base": 1,
                "quote": 3
            }]
        }
    }]
}'
```

> sample response

```json
{
  "id": 5,
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