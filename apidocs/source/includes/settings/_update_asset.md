## Pending update asset

```shell
curl -X POST "https://gateway.local/v3/setting-change-main" \
-H 'Content-Type: application/json' \
-d '{
    "change_list":[{
        "type": "update_asset",
        "data": {
            "asset_id": 3,
            "symbol": "KNC",
            "name": "Kyber Network Crystal",
            "address": "",
            "decimals": 18,
            "transferable": true,
            "set_rate": "exchange_feed",
            "rebalance": true,
            "is_quote": true,
            "is_enabled": false,
            "pwi": {
                "ask": {
                    "a": 0.3241,
                    "b": 0.342472,
                    "c": 0.3414,
                    "min_min_spread": 0.34241,
                    "price_multiply_factor": 0.43141
                },
                "bid": {
                    "a": 0.3241,
                    "b": 0.342472,
                    "c": 0.3414,
                    "min_min_spread": 0.34241,
                    "price_multiply_factor": 0.43141
                }
            },
            "rebalance_quadratic": {
                    "size_a": 0.3241,
                    "size_b": 0.342472,
                    "size_c": 0.3414,
                    "price_a": 0.3242,
                    "price_b": 0.342473,
                    "price_c": 0.3415,
                    "price_offset": 1
            },
            "target": {
                "total": 4134242.432,
                "reserve": 34535,
                "rebalance_threshold": 343.43,
                "transfer_threshold": 4353.4353,
                "min_withdraw_threshold": 100.0
            },
            "stable_param": {
                "price_update_threshold": 0.2,
                "ask_spread": 34.0,
                "bid_spread": 0.1,
                "single_feed_max_spread": 0.4,
                "multiple_feeds_max_diff": 0.6
            },
            "normal_update_per_period": 1.234,
            "max_imbalance_ratio": 3.456,
            "order_duration_millis": 21000,
            "sanity_info": {
                "provider": "binance",
                "threshold": 0.05,
                "path": ["OMG", "ETH"]
            }
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
- "asset_id": required
- "rebalance": true => "rebalance_quadratic": required or this asset already has rebalance_quadratic config<br>
- "rebalance": true => "target": required or this asset already has target config<br>
- "set_rate" is not null => "pwi": required or this asset already has set_rate config<br>
</aside>