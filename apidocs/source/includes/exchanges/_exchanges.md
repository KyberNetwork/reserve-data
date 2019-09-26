#Exchanges

APIs relate to centralized exchanges

## Trade

Create an order into centralized exchange

```shell
curl -X POST "http://gateway.local/trade"
-H 'Content-Type: application/json'
-d '{
   "pair" : 1
   "rate": 0.443,
   "amount": 141,
   "type": "buy" 
}'
```

> sample response

```json
{
    "id": "19234634",
    "success": true,
    "done": 0,
    "remaining": 0.01,
    "finished": false
}
```

### HTTP Request

`POST https://gateway.local/trade`
<aside class="notice">Rebalance key is required</aside>

Params | Type | Required | Default | Description
------ | ---- | -------- | ------- | -----------
pair | uint64 | true | nil | id of pair
rate | float64 | true | nil | rate to create order
amount | float64 | true | nil | order amount 
type | string | true | nil | order type (buy or sell)


## Cancel order 

```shell
curl -X POST "http://gateway.local/cancelorder/binance"
-H 'Content-Type: application/json'
-d '{
    "order_id": 43142
}'
```

> sample response

```json
{
    "success": true
}
```

### HTTP request

`POST https://gateway.local/cancelorder/:exchange`
<aside class="notice">Rebalance key is required</aside>

Param | Type | Required | Default | Description
----- | ---- | -------- | ------- | -----------
order_id | string | true | nil | order id to be cancelled

## Withdraw

```shell
curl -X POST "https://gateway.local/withdraw"
-H 'Content-type: application/json'
-d '{
    "exchange": 1,
    "amount": 41.42342,
    "asset": "ETH"
}'
```

> sample response

```json
{
    "success": true,
    "id": 1432423
}
```

### HTTP Request

`POST https://gateway.local/withdraw`
<aside class="notice">Rebalance key is required</aside>

Params | Type | Required | Default | Description
------ | ---- | -------- | ------- | -----------
exchange | uint64(int) | true | nil | id of exchange
amount | string(big int) | true | nil | amount we want to withdraw
asset | uint64 (asset id) | true | nil | asset we want to withdraw


## Deposit

```shell
curl -X POST "https://gateway.local/deposit"
-H 'Content-Type: application/json'
-d '{
    "exchange": 1,
    "asset": 1,
    "amount": "41342342"
}'
```

> sample response

```json
{
    "success": true,
    "id": 34142342
}
```

### HTTP Request

`POST https://gateway.local/deposit`
<aside class="notice">Rebalance key is required</aside>

Param | Type | Required | Default | Description
----- | ---- | -------- | ------- | -----------
exchange | uint64(int) | true | nil | id of exchange
asset | integer (asset id) | true | nil | asset id
amount | string (big integer) | true | nil | amount to deposit


## Get trade history

```shell
curl -X GET "https://gateway.local/tradehistory?fromTime=1565149153000&toTime=1565235553000"
```

> sample response

```json
{
    "data": {
        "Version": 1517298257114,
        "Valid": true,
        "Timestamp": "1517298257115",
        "Data": {
            "binance": {
                "KNC-ETH": [
                    {
                        "ID": "548006",
                        "Price": 0.003065,
                        "Qty": 29,
                        "Type": "buy",
                        "Timestamp": 1516116380102
                    },
                    {
                        "ID": "548007",
                        "Price": 0.003065,
                        "Qty": 130,
                        "Type": "buy",
                        "Timestamp": 1516116380102
                    }
                ],
                "OMG-ETH": [
                    {
                        "ID": "295923",
                        "Price": 0.020446,
                        "Qty": 4,
                        "Type": "buy",
                        "Timestamp": 1514360742162
                    }
                ],
                "bittrex": {
                    "OMG-ETH": [
                        {
                            "ID": "eb948865-6261-4991-8615-b36c8ccd1256",
                            "Price": 0.01822057,
                            "Qty": 1,
                            "Type": "buy",
                            "Timestamp": 18446737278344972745
                        }
                    ]
                }
            }
        },
        "success": true
    }
}
```

### HTTP Request

`GET https://gateway.local/v3/tradehistory`

Param | Type | Required | Default | Description
----- | ---- | -------- | ------- | -----------
fromTime | uint64 | true | nil | fromTime to get trade history
toTime | uint64 | true | nil | toTime to get trade history

**Limit: toTime - fromTime <= 3 days**

