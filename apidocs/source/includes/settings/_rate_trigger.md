# Rate trigger

## Set rate trigger period

``` shell
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
<aside class="notice">Admin key is required</aside>

## Get rate trigger period


```shell
curl -X GET "https://gateway.local/v3/rate-trigger-period"
```

> sample response

```json
{
  "data": {
    "rate_trigger_period": 5.345
  },
  "success": true
}
```

### HTTP Request

`GET https://gateway.local/v3/rate-trigger-period`
<aside class="notice">All keys are accepted</aside>


## Get Token Rate Trigger

Calculate number of set rate called for each asset, return a map of assetID->Count

``` shell
curl "https://gateway.local/v3/token-rate-trigger?fromTime=1596015657136&toTime=1596015687136"
```
Params | Type | Required | Default | Description
------ | ---- | -------- | ------- | -----------
fromTime | int | yes |  | from time in millis
toTime | int | yes |  | to time in millis

> sample response

```json
{
  "success": true,
  "data": {
    "1": 5,
    "2": 3
  }
}

```

### HTTP Request

`GET "https://gateway.local/v3/token-rate-trigger?fromTime=1596015657136&toTime=1596015687136"`

<aside class="notice">Read key is required</aside>