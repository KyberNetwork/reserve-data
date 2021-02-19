#Exchanges

APIs relate to centralized exchanges

## Trade

Create an order into centralized exchange

```shell
curl -X POST "http://gateway.local/v3/trade"
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

`POST https://gateway.local/v3/trade`
<aside class="notice">Rebalance key is required</aside>

Params | Type | Required | Default | Description
------ | ---- | -------- | ------- | -----------
pair | uint64 | true | nil | id of pair
rate | float64 | true | nil | rate to create order
amount | float64 | true | nil | order amount 
type | string | true | nil | order type (buy or sell)


## Get Open Orders

```shell
curl -X POST "http://gateway.local/v3/open-orders?exchange_id=1&pair=2"
```

> sample response

data is map exchange id with its corresponding open orders

```json
{
    "data": {
        "1": [
            {
                "symbol": "KNCETH",
                "trading_pair_id": 3132131,
                "base": "KNC",
                "quote": "ETH",
                "order_id": "54649705",
                "price": 0.0018,
                "orig_qty": 118,
                "executed_qty": 0,
                "type": "LIMIT",
                "side": "BUY",
                "time": 142342342
            }
        ]
    },
    "success": true
}
```


### HTTP Request

`GET https://gateway.local/v3/open-orders`

Param | Type | Required | Default | Description
----- | ---- | -------- | ------- | -----------
exchange_id | uint64 | false | nil | id of exchange (ex: binance - 1, huobi - 2)
pair | uint64 | true | nil | id of pair (ex: KNCETH - 1, OMGETH - 2)


## Cancel orders 

```shell
curl -X POST "http://gateway.local/v3/cancel-orders"
-H 'Content-Type: application/json'
-d '{
    "orders": [
        {
            "id": "43223",
            "symbol": "KNCETH"
        },
        {
            "id": "43243",
            "symbol": "KNCETH"
        }
    ],
    "exchange_id": 1
}'

// ex: 1 is binance id
```

> sample response

```json
{
    "success": true,
    "data": [
        {
            "43142": {
                "success": true
            }
        },
        {
            "43224": {
                "success": false,
                "error": "this order does not exist"
            }
        }
    ]
}
```

### HTTP request

`POST https://gateway.local/v3/cancel-orders`
<aside class="notice">Rebalance key is required</aside>

Param | Type | Required | Default | Description
----- | ---- | -------- | ------- | -----------
order_ids | string array | true | nil | list of order ids to be cancelled
exchange | string | true | nil | exchange to cancel order

## Cancel all open orders of a symbol

```shell
curl -X POST "http://gateway.local/v3/cancel-all-orders"
-H 'Content-Type: application/json'
-d '{
    "exchange_id": 1,
    "symbol": "KNC"
}'

// ex: 1 is binance id
```

> sample response

```json
{
    "success": true
}
```  
or  
```json
{
    "success": false,
    "error": "cannot cancel order"
}
```



### HTTP request

`POST https://gateway.local/v3/cance-all-orders`
<aside class="notice">Rebalance key is required</aside>

Param | Type | Required | Default | Description
----- | ---- | -------- | ------- | -----------
symbol| string | true | nil | symbol to cancel open orders
exchange | string | true | nil | exchange to cancel orders

## Withdraw

```shell
curl -X POST "https://gateway.local/v3/withdraw"
-H 'Content-type: application/json'
-d '{
    "exchange": 1,
    "amount": 41.42342,
    "asset": 1
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

`POST https://gateway.local/v3/withdraw`
<aside class="notice">Rebalance key is required</aside>

Params | Type | Required | Default | Description
------ | ---- | -------- | ------- | -----------
exchange | uint64(int) | true | nil | id of exchange
amount | string(big int) | true | nil | amount we want to withdraw
asset | uint64 (asset id) | true | nil | asset we want to withdraw

## CEX Transfer between account

```shell
curl -X POST "https://gateway.local/v3/cex-transfer"
-H 'Content-type: application/json'
-d '{
    "exchange": 1,
    "amount": 1000000000000000000,
    "asset": 1,
    "from_account": "main",
    "to_account": "account_1"
}'
```

> sample response

```json
{
    "success": true,
    "id": 1432423
}
```

## Binance main account info

```shell
curl "https://gateway.local/v3/binance/main"
```

> sample response

```json
{
    "success": true,
    "data": {
      "balances": [
        {
          "asset": "BTC",
          "free": 0.02066858,
          "locked": 0
        },
        {
          "asset": "LTC",
          "free": 0,
          "locked": 0
        },
        {
          "asset": "ETH",
          "free": 0.15406002,
          "locked": 0
        }
      ]
    }
}
```

### HTTP Request

`GET https://gateway.local/v3/binance/main`
<aside class="notice">All keys are accepted</aside>


## Deposit

```shell
curl -X POST "https://gateway.local/v3/deposit"
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

`POST https://gateway.local/v3/deposit`
<aside class="notice">Rebalance key is required</aside>

Param | Type | Required | Default | Description
----- | ---- | -------- | ------- | -----------
exchange | uint64(int) | true | nil | id of exchange
asset | integer (asset id) | true | nil | asset id
amount | string (big integer) | true | nil | amount to deposit


## Get trade history

```shell
curl -X GET "https://gateway.local/v3/tradehistory?fromTime=1565149153000&toTime=1565235553000"
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


## Get Asset Withdraw Status

```shell
curl -X GET "https://gateway.local/v3/all-asset-withdraw-status/:exchange"
```

> sample response

```json
{
    "success": true,
    "data": {
        "1": true,
        "2": true,
        ...
    }
}
```

### HTTP Request

`GET https://gateway.local/v3/all-asset-withdraw-status/:exchange`
<aside class="notice">All keys are accepted</aside>
