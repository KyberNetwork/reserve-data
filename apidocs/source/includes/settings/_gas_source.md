# Gas Source

## Set Gas Source

the request must be application/json format.

``` shell
curl -X POST "https://gateway.local/v3/gas-source" \
-H 'Content-Type: application/json' \
-d '{
      "name": "etherscan"
    }'
```

> sample response

```json
{
  "success": true
}
```

### HTTP Request

`POST https://gateway.local/v3/gas-source`

Params | Type | Required | Default | Description
------ | ---- | -------- | ------- | -----------
name | string | yes |  | name of prefer source
<aside class="notice">Admin key is required</aside>

## Get Gas Source


```shell
curl -X GET "https://gateway.local/v3/gas-source"
```

> sample response

```json
{
  "success": true,
  "data": {
    "name": "etherscan"
  }
}
```

### HTTP Request

`GET https://gateway.local/v3/gas-source`
<aside class="notice">All keys are accepted</aside>
