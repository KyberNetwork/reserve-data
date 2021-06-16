# Rate trigger

## Add schedule job

``` shell
curl -X POST "https://gateway.local/v3/scheduled-job" \
-H 'Content-Type: application/json' \
-d '{
		  "endpoint"`: "/v3/setting-change-main/2",
	    "http_method": "PUT",
      "data": null,
	    "scheduled_time": 123456789
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

`POST https://gateway.local/v3/scheduled-job`

Params | Type | Required | Default | Description
------ | ---- | -------- | ------- | -----------
endpoint | string | yes |  | the endpoint the job should call
http_method | string | yes |  | http method
data | json | no | | data to call
scheduled_time | uint64 | yes | | the time (in millisecond) to execute the job

<aside class="notice">Admin key is required</aside>

## Get all schedule job


```shell
curl -X GET "https://gateway.local/v3/scheduled-job"
```

> sample response

```json
{
  "data": [
    {
      "endpoint"`: "/v3/setting-change-main/2",
	    "http_method": "PUT",
      "data": null,
	    "scheduled_time": 123456789
    },
    ...
  ],
  "success": true
}
```

### HTTP Request

`GET https://gateway.local/v3/scheduled-job`
<aside class="notice">All keys are accepted</aside>


## Get schedule job


```shell
curl -X GET "https://gateway.local/v3/scheduled-job/:id"
```

> sample response

```json
{
  "data": {
      "endpoint"`: "/v3/setting-change-main/2",
	    "http_method": "PUT",
      "data": null,
	    "scheduled_time": 123456789
    },
  "success": true
}
```

### HTTP Request

`GET https://gateway.local/v3/scheduled-job/:id`
<aside class="notice">All keys are accepted</aside>

## Remove schedule job


```shell
curl -X DELETE "https://gateway.local/v3/scheduled-job/:id"
```

> sample response

```json
{
  "success": true
}
```

### HTTP Request

`DELETE https://gateway.local/v3/scheduled-job/:id`
<aside class="notice">Admin and manage keys are accepted</aside>