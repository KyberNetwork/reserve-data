# Rate trigger period

## Set rate trigger period

```shell
curl -X POST "https://gateway.local/v3/rate-trigger-period" \
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

`POST https://gateway.local/v3/rate-trigger-period`

Params | Type | Required | Default | Description
------ | ---- | -------- | ------- | -----------
value | float64 | yes |  | value of rate trigger period
<aside class="notice">Write key is required</aside>

## Get rate trigger period


```shell
curl -X GET "https://gateway.local/v3/rate-trigger-period"
```

> sample response

```json
{
  "data": {
    "key": "rate_trigger_period",
    "value": 5.345
  },
  "success": true
}
```

### HTTP Request

`GET https://gateway.local/v3/rate-trigger-period`
<aside class="notice">All keys are accepted</aside>
