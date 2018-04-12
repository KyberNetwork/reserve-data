# Verification Script Usage

## Build

Change directory to verification folder then run build command:

```
  go build
```

## Run 
```
  ./verification
```

## Options
### Commands
1. verify
  Run throught all supported exchange and try checking deposit, withdraw api

  Eg:
  ```
    ./verification verify
  ```
2. deposit
  Run deposit command

  Eg:
  ```
    ./verification deposit -exchange huobi -amount 0.5 -token ETH -base_url http://localhost:8000
  ```
3. withdraw
  Run withdraw command

  Eg:
  ```
    ./verification withdraw -exchange huobi -amount 0.5 -token ETH -base_url http://localhost:8000
  ```
