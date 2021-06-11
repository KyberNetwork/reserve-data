# Setting change main

## Create setting main

```shell
curl -X POST "https://gateway.local/v3/setting-change-main" \
-H 'Content-Type: application/json' \
-d '{
    "change_list": [
        {
            "type": "...",
            "data" : {
                 ...
            }
        },
        ...
    ],
    "scheduled_time": 123456789 // to do as a schedule job
}'
```

> sample response

```json
{
  "id": 6,
  "success": true
}
```

### HTTP Request

`POST https://gateway.local/v3/setting-change-main`

"change_list" is a JSON array including changes, each change consists the following fields:


Params | Type | Required | Default | Description
------ | ---- | -------- | ------- | -----------
type | string | true | nil | type of setting change
data | json object | true | nil | information about the changes

Change types includes:

<a href="#pending-create-asset">create_asset</a><br>
<a href="#pending-update-asset">update_asset</a><br>
<a href="#pending-create-asset-exchange">create_asset_exchange</a><br>
<a href="#pending-update-asset-exchange">update_asset_exchange</a><br>
<a href="#pending-delete-asset-exchange">delete_asset_exchange</a><br>
<a href="#pending-create-trading-pair">create_trading_pair</a><br>
<a href="#pending-delete-trading-pair">delete_trading_pair</a><br>
<a href="#pending-change-asset-address">change_asset_addr</a><br>

## Get pending setting change 


```shell
curl -X GET "https://gateway.local/v3/setting-change-main?status=pending"
```

> sample response

```json
{
  "data": [
    {
      "id": 6,
      "created": "2019-08-13T07:25:49.869418Z",
      "change_list":{
        "type": "delete_trading_pair",
        "data": {
          "id": 2
        }
      },
      "proposer": "proposer",
      "list_approval":[
        {
          "key_id": "manage_access_key_id",
          "timestamp": "2020-12-14T09:55:14.222009Z"
        }
      ]
    }
  ],
  "success": true
}
```

### HTTP Request

`GET https://gateway.local/v3/setting-change-main`
<aside class="notice">All keys are accepted</aside>

#### Params
Params | Type | Required | Default | Description
------ | ---- | -------- | ------- | -----------
status | string | false | pending | status of setting change (include: pending, accepted, rejected)

## Approve pending setting change

Approve for a setting change, when the number of approvals reaches the requirement the setting change will be applied automatically.

```shell
curl -X PUT "https://gateway.local/v3/setting-change-main/1"
```

> sample response

```json
{
    "success": true
}
```

## Disapprove pending setting change

Disapprove a setting change that is approved by you before it's applied.

```shell
curl -X DELETE "https://gateway.local/v3/disapprove-setting-change/1"
```

> sample response

```json
{
    "success": true
}
```

### HTTP Request

`PUT https://gateway.local/v3/setting-change-main/:change_id`
<aside class="notice">Admin key is required</aside>

## Reject pending setting change 

```shell
curl -X DELETE "https://gateway.local/v3/setting-change-main/1"
```

> sample response

```json
{
    "success": true
}
```

### HTTP Request

`DELETE https://gateway.local/v3/setting-change-main/:change_id`
<aside class="notice">Admin key is required</aside>