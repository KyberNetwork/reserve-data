## Pending create asset

```shell
curl -X POST "https://gateway.local/v3/setting-change-main" \
-H 'Content-Type: application/json' \
-d '{
    "change_list":[{
        "type": "create_asset",
        "data": {
            "symbol": "OMG",
            "name": "Omisego",
            "address": "0xdd974d5c2e2928dea5f71b9825b8b646686bd200",
            "decimals": 18,
            "transferable": true,
            "set_rate": "exchange_feed",
            "rebalance": true,
            "is_quote": true,
            "is_enabled": true,
            "pwi": {
                "ask": {
                    "a": 13.88888,
                    "b": -0.11111,
                    "c": 0,
                    "min_min_spread": 0.005,
                    "price_multiply_factor": 0.45
                },
                "bid": {
                    "a": 13.88888,
                    "b": -0.11111,
                    "c": 0,
                    "min_min_spread": 0.005,
                    "price_multiply_factor": 0.45
                }
            },
            "rebalance_quadratic": {
                "size_a": 0.000001754386,
                "size_b": 0.0004894737,
                "size_c": 0.9995088,
                "price_a": 0.1234,
                "price_a": 0.1235,
                "price_a": 0.1236
            },
            "exchanges": [
                {
                    "id": 5,
                    "asset_id": 3,
                    "exchange_id": 0,
                    "symbol": "OMG",
                    "deposit_address": "0x023ab1f7acaad1f7a01d3bfa4afd2ab575780090",
                    "min_deposit": 0,
                    "withdraw_fee": 18,
                    "target_recommended": 0,
                    "target_ratio": 0,
                    "trading_pairs": []
                }
            ],
            "target": {
                "total": 70000,
                "reserve": 48000,
                "rebalance_threshold": 0.33,
                "transfer_threshold": 0.25,
                "min_withdraw_threshold": 1.0
            },
            "stable_param": {
                "price_update_threshold": 0.2,
                "ask_spread": 34.0,
                "bid_spread": 0.1,
                "single_feed_max_spread": 0.4,
                "multiple_feeds_max_diff": 0.6
            },
            "normal_update_per_period": 1.234, // default value is 1
            "max_imbalance_ratio": 3.456, // default value is 2
            "order_duration_millis": 20000,
            "sanity_info": {
                "provider": "binance",
                "threshold": "threshold",
                "path": {"OMG", "ETH"}
            }
        }
    }]
}'
```

> sample response

```json
{
  "id": 8,
  "success": true
}
```

### HTTP Request

`POST https://gateway.local/v3/setting-change-main`
<aside class="notice">Manage key is required</aside>
<aside class="warning">
Constraints:<br>
- "rebalance": true => "rebalance_quadratic": required<br>
- "rebalance": true => "target": required<br>
- "set_rate" is not null => "pwi": required<br>
- "asset_exchange" child objs need to be valid with asset_exchange constraints
</aside>