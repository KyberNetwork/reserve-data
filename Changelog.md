# Reserve-data Changelog

## Unreleased

### Features:

### Bug fixes:

### Improvements:

### Compatibility

## 0.10.0 (2018-07-03)

### Features:

- Add rebalance quadratic API set (#327)

### Bug fixes:

- Fix type/interface conversion error swallow. Now these errors will be reported (#325)

### Improvements:

- Remove many unused methods (#309)
- Refactor GasOracle (#318)
- Improve input validation for many APIs (#330, #334, #336, #340)
- Fix many hardcoded API endpoints (#331)

### Compatibility

- KyberNetwork smart contracts (>= 2.0.0)
- KyberNetwork analytic (0.9.0)

## 0.9.1 (2018-06-26)

### Compatibility:
- KyberNetwork smart contracts (>= 0.3.0)
- KyberNetwork analytic (0.8.0)

### Features:

- Listed new tokens
  - POLY
  - LBA
  - EDU
  - CVC

### Bug fixes:

### Improvements:

### Compatibility

## 0.9.0 (2018-06-19)

### Compatibility:
- KyberNetwork smart contracts (>= 0.3.0)
- KyberNetwork analytic (0.8.0)

### Features:

### Bug fixes:

- Enable errcheck checker and fix all unhandled errors (#307)
- Fix GetAccounts error is ignored for houbi exchange (#308)
- Fix missing required fields in PWIs v2 APIs (#314)

- Handle trade error in huobi (#317)
- Fix Binance is using wrong storage implementation (#310)
- Fix failed to store rate on a blank database (#287)
- Fix resource leaks in CallContract (#291)
- Fix inconsistent exchange status (#285)
### Improvements:

- Enable vet shadow (#296)
- Remove unused verification package (#312)
- Refactor activity record saving (#295)
- Add tests for Qty V2 APIs (#289)
- Add document for tbindex (#284)

### Compatibility:

## 0.8.0 (2018-06-03)

### Features:
- Support dry mode to check all configs and initializations (#223)
- Support more price feeds (XAU-ETH, USD-ETH...)
- Set target qty v2 to support more targets on ETH (#227)
- Add API to monitor pricing gas cost (#207)
- Add support for pwis in both side (sell/buy)

### Bug fixes:
- Fixed where trade logs are skipped if db writes failed (#216)
- Handled a lot of shadow errors
- Fixed a lot of minor bugs and design defects

### Improvements:
- Push old log to s3 (#217)
- Improved a lot code base quality
- Removed redundant signal field of TickerRunner
- Start having APIs version 2 to use json instead of string concatenation format

### Compatibility:
- KyberNetwork smart contracts (>= 0.3.0)
- KyberNetwork analytic (0.8.0)

## 0.7.0 (2018-04-30)

### Features:
- Support global data monitoring (gold feeds...)
- Support stable exchange (the virtual exchange to handle tokens that are not listed on big cexs)

### Bug fixes:
- TODO

### Compatibility:
- This version only works with KyberNetwork smart contracts version 0.3.0 or later

## 0.6.1 (2018-04-25)
### Features:
- Getting aws info from config file

### Bug fixes:
- Get correct filename from AWS's Item key for intergrity check

### Compatibility:
- This version only works with KyberNetwork smart contracts version 0.3.0 or later

## 0.6.0 (2018-04-25)

### Features:
- support Huobi as backend centralized exchange
- API for analytic to submit notifications on exchange status
- API to enable/disable a particular exchange
- More graph data such as (timezone, heatmap, trade summary, trade summary for a wallet)
- Support more tokens (including DAI)
- Support multiple token mode (internal use, external use, unlisted tokens)
- API to support status server, dashboard notifications
- Analytic dev mode to support dev environment for analytic team
- API for analytic to submit pricing data

### Bug fixes:
- fix bunch of errors in order to ensure stat server will not miss any tradelogs (including making log aggregation and last log id persistence atomic)
- return log in correct order to ensure consistent stat results
- enable authentication in stat server (it was ignored before)

### Improvements:
- Fallback to other nodes when primary node is down with contract calls
- Massive refactor the codebase which is related to blockchain interaction
- Add more tests to make the whole component more stable
- Improve a lot log aggregration performance by process in group and reduce number of database transactions
- Using ethereum.Address type in database interfaces to ensure it's consistency
- Update bittrex constants and settings
- Persist error messages for all kind of activities (deposit/withdraw/trade)

### Compatibility:
- This version only works with KyberNetwork smart contracts version 0.3.0 or later

## 0.5.0 (2018-03-08)

### Features:
- Add api to halt setrate
- Reorganize commandline to make core runable in different modes and configs
- Rotate logs for better log management
- Add API to get all core relevant addresses
- Add api to check core version
- Add a tool to monitor base/compact to detect bugs
- Add binance trade history API
### Bug fixes:
- Fix minors with deposit signer
- Fix order of pending activities in its API
- Add API timerange limit to all relevant apis to mitigate dos attack from key keepers
- Add sanity check to validate response from node
- Remove eth-eth pair in requesting to exchange
### Improvements:
- Update binance limits
- Reduce bianance api rate
- Organize configuration better to list/delist token more easily
- Wait sometime before fetching new rate to hopefully mitigrate reorg
- Added sanity check on deposit/trade/withdraw
- Improved gas limit estimation for deposit and setrate
- Removed duplicated records in get rate API
- Query rate at a specific block instead of relying on latest block

## 0.4.1 (2018-02-19)
### Features:
- Listed more 4 tokens (eng, salt, appc, rdn)
- Added more tools for monitoring and testing such as deposit/withdraw trigger, rate validator

### Bug fixes:
- Fixed submit empty setrate for the first one
- Fixed bug in rare case that panics when core couldn't get mined nonce
- Fixed incompatibility between geth and parity in tx receipt data
- Enable microsecond info in log

### Improvements:
- Separated cex token pairs to config
- Separated cex fee to config
- Added sanity checks on setrates, deposit, withdraw and trade
- Added env tag to sentry

### Compatibility:
- This version only works with KyberNetwork smart contracts version 0.3.0 or later

## 0.4.0 (2018-02-08)

### Features:
- Support rebalance toggle, dynamic target qty with set/confirm key model
- Support multiple keys for different roles

### Bug fixes:
- Fixed minor bugs
- Detect throwing txs

### Improvements:
- Done sanity check in with setrate api
- Rebroadcasting tx to multiple node to improve tx propagation
- Replace staled/long mining set rate txs
- Made improvements to the code base
- Applied timeout to communication to nodes to ensure analytic doesn't have to wait for too long to set another rate

### Compatibility:
- This version only works with KyberNetwork smart contracts version 0.3.0 or later

## 0.3.0 (2018-01-31)

### Features:
- Introduce various key permissions
- New API for getting KN rate historical data
- New API for getting trade history on cexs

### Bugfixes:
- Handle lost transactions

### Improvements:
- Using multiple nodes to broadcast tx
- Avoid storing redundant rate data
- More code refactoring

### Compatibility:
- This version only works with KyberNetwork smart contracts version 0.3.0


