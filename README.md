# Data fetcher for KyberNetwork reserve
[![Go Report Card](https://goreportcard.com/badge/github.com/KyberNetwork/reserve-data)](https://goreportcard.com/report/github.com/KyberNetwork/reserve-data)
[![Build Status](https://travis-ci.org/KyberNetwork/reserve-data.svg?branch=develop)](https://travis-ci.org/KyberNetwork/reserve-data)

This repo is contains two components:

- core: 
	- interacts with blockchain to get/set rates for tokens pair
	- buy/sell with centralized exchanges (binance, huobi, etc)
(For more detail, take a look to interface ReserveCore in intefaces.go)

- stat:  
	- fetch tradelogs from blockchain and do aggregation and save its data to database and allow client to query

(For more detail, find ReserveStat interface in interfaces.go)

## Compile it

```shell
cd cmd && go build -v
```

a `cmd` executable file will be created in `cmd` module.

## Run the reserve data

1. You need to prepare a `config.json` file inside `cmd` module. The file is described in later section.
2. You need to prepare a JSON keystore file inside `cmd` module. It is the keystore for the reserve owner.
3. Make sure your working directory is `cmd`. Run `KYBER_EXCHANGES=binance,huobi ./cmd` in dev mode.

### Manual

```shell
cd cmd
```

- Run core only

```shell
KYBER_EXCHANGES="binance,huobi" KYBER_ENV=production ./cmd server --log-to-stdout
```

- Run stat only

```shell
KYBER_ENV=production ./cmd server --log-to-stdout --enable-stat --no-core
```

### Docker (recommended)

This repository will build docker images and public on [docker hub](https://hub.docker.com/r/kybernetwork/reserve-data/tags/), you can pull image from docker hub and run:

- Run core only

```shell
docker run -p 8000:8000 -v /location/of/config.json:/go/src/github.com/KyberNetwork/reserve-data/cmd/config.json -e KYBER_EXCHANGES="binance,huobi" KYBER_ENV="production" kybernetwork/reserve-data:develop server --log-to-stdout
```

- Run stat only 

```shell
docker run -p 8000:8000 -v /location/of/config.json:/go/src/github.com/KyberNetwork/reserve-data/cmd/config.json -e KYBER_ENV="production" kybernetwork/reserve-data:develop server --enable-stat --no-core --log-to-stdout
```

**Note** : 

- KYBER_ENV includes "dev, simulation and production", different environment mode uses different settings (check cmd folder for settings file).  

- reserve-data requires config.json file to run, so you need to -v (mount config.json file to docker) so it can run.

## Config file

sample:

```json
{
  "binance_key": "your binance key",
  "binance_secret": "your binance secret",
  "huobi_key": "your huobi key",
  "huobi_secret_key": "your huobi secret",
  "kn_secret": "secret key for people to sign their requests to our apis. It is ignored in dev mode.",
  "kn_readonly": "read only key for people to sign their requests, this key can read everything but cannot execute anything",
  "kn_configuration": "key for people to sign their requests, this key can read everything and set configuration such as target quantity",
  "kn_confirm_configuration": "key for people to sign ther requests, this key can read everything and confirm target quantity, enable/disable setrate or rebalance",
  "keystore_path": "path to the JSON keystore file, recommended to be absolute path",
  "passphrase": "passphrase to unlock the JSON keystore",
  "keystore_deposit_path": "path to the JSON keystore file that will be used to deposit",
  "passphrase_deposit": "passphrase to unlock the JSON keystore",
  "keystore_intermediator_path": "path to JSON keystore file that will be used to deposit to Huobi",
  "passphrase_intermediate_account": "passphrase to unlock JSON keystore",
  "aws_access_key_id": "your aws key ID",
  "aws_secret_access_key": "your aws scret key",
  "aws_expired_stat_data_bucket_name" : "AWS bucket for expired stat data (already created)",
  "aws_expired_reserve_data_bucket_name" : "AWS bucket for expired reserve data (already created)",
  "aws_log_bucket_name" :"AWS bucket for log backup(already created)",
  "aws_region":"AWS region"
}
```

## APIs

https://dev-reserve-data-docs.knstats.com/#introduction