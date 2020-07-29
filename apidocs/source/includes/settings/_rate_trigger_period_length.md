# Rate trigger period length

## Set rate trigger period length

```shell
curl -X POST "https://gateway.local/v3/rate-trigger-period-length" \
-H 'Content-Type: application/json' \
-d '{
		  "value": 5.345
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

`POST https://gateway.local/v3/rate-trigger-period-length`

Params | Type | Required | Default | Description
------ | ---- | -------- | ------- | -----------
value | float64 | yes | nil | value of rate trigger period length
<aside class="notice">Write key is required</aside>

## Get rate trigger period length


```shell
curl -X GET "https://gateway.local/v3/rate-trigger-period-length"
```

> sample response

```json
{
  "data": {
    "id": 6,
    "key": "rate_trigger_period_length",
    "value": 5.345
  },
  "success": true
}
```

### HTTP Request

`GET https://gateway.local/v3/rate-trigger-period-length`
<aside class="notice">All keys are accepted</aside>
