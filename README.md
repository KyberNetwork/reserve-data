# Data fetcher for KyberNetwork reserve
[![Go Report Card](https://goreportcard.com/badge/github.com/KyberNetwork/reserve-data)](https://goreportcard.com/report/github.com/KyberNetwork/reserve-data)
[![Build Status](https://travis-ci.org/KyberNetwork/reserve-data.svg?branch=develop)](https://travis-ci.org/KyberNetwork/reserve-data)

## Compile it

```
cd cmd && go build -v
```
a `cmd` executable file will be created in `cmd` module.

## Run the reserve data

1. You need to prepare a `config.json` file inside `cmd` module. The file is described in later section.
2. You need to prepare a JSON keystore file inside `cmd` module. It is the keystore for the reserve owner.
3. Make sure your working directory is `cmd`. Run `KYBER_EXCHANGES=binance,bittrex ./cmd` in dev mode.

## Config file

sample:
```
{
  "binance_key": "your binance key",
  "binance_secret": "your binance secret",
  "kn_secret": "secret key for people to sign their requests to our apis. It is ignored in dev mode.",
  "kn_readonly": "read only key for people to sign their requests, this key can read everything but cannot execute anything",
  "kn_configuration": "key for people to sign their requests, this key can read everything and set configuration such as target quantity",
  "kn_confirm_configuration": "key for people to sign ther requests, this key can read everything and confirm target quantity, enable/disable setrate or rebalance",
  "keystore_path": "path to the JSON keystore file, recommended to be absolute path",
  "passphrase": "passphrase to unlock the JSON keystore"
  "keystore_deposit_path": "path to the JSON keystore file that will be used to deposit",
  "passphrase_deposit": "passphrase to unlock the JSON keystore"
}
```

## APIs

### Get time server
```
<host>:8000/timeserver
```

eg:
```
curl -X GET "http://localhost:8000/timeserver"
```

response:
```
{
  "data": "1517479497447",
  "success": true
}
```

### Get all addresses are being used by core

```
<host>:8000/core/addresses
```

eg:
```
curl -X GET "http://localhost:8000/core/addresses"
```
response:
```
{"data":{"tokens":{"EOS":"0x15fb2a9d7dadbb88f260f78dcbb574b3b76a8e06","ETH":"0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee","KNC":"0x8dc114d77e857558aefbe8e1a50b460ff9578f1a","OMG":"0x7606bd550f467546212649a9c25623dfca88dcd7","SALT":"0xcc112cd38362bf3c07d226768fd5869e65296083","SNT":"0x676f650000f420485b99ef0377a2e1c96eb3e821"},"exchanges":{"binance":{"EOS":"0x1ae659f93ba2fc0a1f379545cf9335adb75fa547","ETH":"0x1ae659f93ba2fc0a1f379545cf9335adb75fa547","KNC":"0x1ae659f93ba2fc0a1f379545cf9335adb75fa547","OMG":"0x1ae659f93ba2fc0a1f379545cf9335adb75fa547","SALT":"0x1ae659f93ba2fc0a1f379545cf9335adb75fa547","SNT":"0x1ae659f93ba2fc0a1f379545cf9335adb75fa547"},"bittrex":{"EOS":"0xef6ee90c5bb23da2eb71b3daa8e57b204e5ac647","ETH":"0xe0355aa3cc0a4e0e4b2a70acd90c2fa961f61b23","KNC":"0x132478f1ec4b8e1256b11fdf3e00d97e4df5988f","OMG":"0x9db6e8d2d133448dbcf755f19d540253da4ba043","SALT":"0x385d619b530f00ab7d082683f7cdc37995ac76f2","SNT":"0x3ef96f9de64c44b1ad392b10e2277a73ec14ff5f"}},"wrapper":"0xa54f27b5a72fc1ddc5c4bc6ed50391f457e4a46a","pricing":"0x77925520469d0fcbb0311814c053bf9bafcd867b","reserve":"0x2d1ceabd5a1cd16581ad199031601615a434a2cd","feeburner":"0xa33a2f0745ee8e31b753ec33d22d363a62a123a4","network":"0x643211b405c9a14139142e1104250bbcd94bd0ef"},"success":true}
```

### Get prices for specific base-quote pair

```
<host>:8000/prices/<base>/<quote>
```

Where *<base>* is symbol of the base token and *<quote>* is symbol of the quote token

eg:
```
curl -X GET "http://13.229.54.28:8000/prices/omg/eth"
```

### Get prices for all base-quote pairs
```
<host>:8000/prices
```

eg:
```
curl -X GET "http://localhost:8000/prices"
```
response:
```
{"data":{"ADX-ETH":{"bittrex":{"Valid":true,"Error":"","Timestamp":"1514114579228","Bids":[{"Quantity":11.16312371,"Rate":0.00264449},{"Quantity":202.35427853,"Rate":0.00264448},{"Quantity":75,"Rate":0.00264352},{"Quantity":97.32884421,"Rate":0.00263935},{"Quantity":1534.265,"Rate":0.00250437},{"Quantity":1147.78359847,"Rate":0.00250435},{"Quantity":426.37538021,"Rate":0.00250429},{"Quantity":38.11472359,"Rate":0.00250428},{"Quantity":467.8397624,"Rate":0.0025},{"Quantity":400,"Rate":0.002492},{"Quantity":4113.521,"Rate":0.00247909},{"Quantity":1452.15160462,"Rate":0.00247908},{"Quantity":586.36558478,"Rate":0.00247906},{"Quantity":91.18953202,"Rate":0.00246122},{"Quantity":2195.04164481,"Rate":0.00246009},{"Quantity":513.46207919,"Rate":0.00246008},{"Quantity":105.60426867,"Rate":0.0024212},{"Quantity":508.00203601,"Rate":0.00241},{"Quantity":84.89325577,"Rate":0.00235001},{"Quantity":10,"Rate":0.00235},{"Quantity":454.72,"Rate":0.00233853},{"Quantity":29.30270379,"Rate":0.00229974},{"Quantity":5000,"Rate":0.00228201},{"Quantity":88.64641273,"Rate":0.00225051},{"Quantity":102.41475552,"Rate":0.00224439},{"Quantity":48.34010514,"Rate":0.00220357},{"Quantity":45.33779727,"Rate":0.00220015},{"Quantity":2591.95297273,"Rate":0.0022},{"Quantity":45.59562282,"Rate":0.00218771},{"Quantity":91.35912149,"Rate":0.00218369},{"Quantity":73.71614193,"Rate":0.00216752},{"Quantity":156.0554184,"Rate":0.0021573},{"Quantity":139.59345821,"Rate":0.00215},{"Quantity":1396.28311069,"Rate":0.00214319},{"Quantity":88.57979941,"Rate":0.00213785},{"Quantity":20,"Rate":0.00213618},{"Quantity":36.56576324,"Rate":0.00213069},{"Quantity":1969.24295775,"Rate":0.00213},{"Quantity":30,"Rate":0.0021217},{"Quantity":415.6671988,"Rate":0.00212131},{"Quantity":30,"Rate":0.00212121},{"Quantity":10,"Rate":0.00212},{"Quantity":514.25,"Rate":0.0021},{"Quantity":120.35835049,"Rate":0.00209785},{"Quantity":10,"Rate":0.00207728},{"Quantity":42.33687186,"Rate":0.00207337},{"Quantity":86.533295,"Rate":0.00206796},{"Quantity":10,"Rate":0.00206066},{"Quantity":80.67770235,"Rate":0.00204444},{"Quantity":70,"Rate":0.00202382},{"Quantity":44.38001453,"Rate":0.00202287},{"Quantity":31.88963038,"Rate":0.00202163},{"Quantity":30.52335887,"Rate":0.00202146},{"Quantity":1521.31424734,"Rate":0.002},{"Quantity":9.97549877,"Rate":0.0019999},{"Quantity":49.98772231,"Rate":0.00199549},{"Quantity":1503.39559434,"Rate":0.00199548},{"Quantity":15.73375618,"Rate":0.00199436},{"Quantity":71.78447794,"Rate":0.00197853},{"Quantity":80.88958992,"Rate":0.0019753},{"Quantity":50,"Rate":0.00197},{"Quantity":5,"Rate":0.00195495},{"Quantity":250,"Rate":0.001947},{"Quantity":1541.07965277,"Rate":0.00194668},{"Quantity":8.92871246,"Rate":0.00193499},{"Quantity":31.69266153,"Rate":0.00190323},{"Quantity":63,"Rate":0.0019},{"Quantity":1580.70145711,"Rate":0.00189789},{"Quantity":5,"Rate":0.00188139},{"Quantity":53.02438324,"Rate":0.00188121},{"Quantity":5,"Rate":0.00187012},{"Quantity":3785.6254394,"Rate":0.0018491},{"Quantity":1622.41441223,"Rate":0.00184909},{"Quantity":11.27118644,"Rate":0.00177},{"Quantity":221,"Rate":0.00175},{"Quantity":10,"Rate":0.00172998},{"Quantity":5,"Rate":0.00172583},{"Quantity":5,"Rate":0.001711},{"Quantity":11.66678333,"Rate":0.00171},{"Quantity":76.22677571,"Rate":0.00170121},{"Quantity":50,"Rate":0.001688},{"Quantity":307.75711692,"Rate":0.00168542},{"Quantity":29.92745405,"Rate":0.00166653},{"Quantity":120.10836845,"Rate":0.001661},{"Quantity":5,"Rate":0.00166},{"Quantity":151.40522261,"Rate":0.00164707},{"Quantity":2000,"Rate":0.00161188},{"Quantity":9.65322581,"Rate":0.00155},{"Quantity":665,"Rate":0.0015},{"Quantity":2000,"Rate":0.00141188},{"Quantity":357.88605052,"Rate":0.0013936},{"Quantity":551.10497238,"Rate":0.0011},{"Quantity":2000,"Rate":0.00091188},{"Quantity":153.24934706,"Rate":0.0006509},{"Quantity":4000,"Rate":0.00061188},{"Quantity":4000,"Rate":0.00021188},{"Quantity":1000,"Rate":0.00001},{"Quantity":41562.5,"Rate":0.0000012},{"Quantity":48897.05882353,"Rate":0.00000102},{"Quantity":10000,"Rate":0.00000101}],"Asks":[{"Quantity":4850.84,"Rate":0.00277997},{"Quantity":144.04135361,"Rate":0.00277998},{"Quantity":14.50780994,"Rate":0.00278059},{"Quantity":59.72507722,"Rate":0.00289147},{"Quantity":2333.89607257,"Rate":0.00289148},{"Quantity":1715.55875691,"Rate":0.0029145},{"Quantity":17.95991987,"Rate":0.00291781},{"Quantity":23.05644335,"Rate":0.00293006},{"Quantity":3403.71231078,"Rate":0.00293796},{"Quantity":200,"Rate":0.00295},{"Quantity":2868.40462067,"Rate":0.0029502},{"Quantity":50,"Rate":0.00295078},{"Quantity":128.26061528,"Rate":0.0029682},{"Quantity":30,"Rate":0.00297069},{"Quantity":1186.939,"Rate":0.00297416},{"Quantity":180.90249945,"Rate":0.00297417},{"Quantity":159.95625,"Rate":0.0029758},{"Quantity":8.38243676,"Rate":0.00298},{"Quantity":21.07394366,"Rate":0.002982},{"Quantity":3350.69921968,"Rate":0.00298445},{"Quantity":18.81232821,"Rate":0.00298585},{"Quantity":53.43540075,"Rate":0.00299898},{"Quantity":401.71248796,"Rate":0.003},{"Quantity":100,"Rate":0.00301},{"Quantity":13.21577148,"Rate":0.00301801},{"Quantity":1655.88714797,"Rate":0.00301952},{"Quantity":7.95670277,"Rate":0.00302047},{"Quantity":2000,"Rate":0.00302188},{"Quantity":10000,"Rate":0.00303},{"Quantity":3299.31216912,"Rate":0.00303093},{"Quantity":13.06347493,"Rate":0.00305066},{"Quantity":3249.47747718,"Rate":0.00307741},{"Quantity":56.56258781,"Rate":0.00308981},{"Quantity":1480.88701282,"Rate":0.0031},{"Quantity":53.83657683,"Rate":0.00312},{"Quantity":2000,"Rate":0.00312188},{"Quantity":3201.12584741,"Rate":0.0031239},{"Quantity":1600.22707577,"Rate":0.00312455},{"Quantity":19265.02145923,"Rate":0.00313131},{"Quantity":70.42205483,"Rate":0.00313723},{"Quantity":207.0336135,"Rate":0.00315},{"Quantity":254.1,"Rate":0.00316},{"Quantity":3154.19204739,"Rate":0.00317038},{"Quantity":2855.17427501,"Rate":0.00317202},{"Quantity":29.33307688,"Rate":0.00318699},{"Quantity":766.12452394,"Rate":0.00318787},{"Quantity":65.79269502,"Rate":0.00319959},{"Quantity":7890.05540482,"Rate":0.0032},{"Quantity":3.3222,"Rate":0.00321088},{"Quantity":2000,"Rate":0.00322188},{"Quantity":1699.65771917,"Rate":0.00322958},{"Quantity":230.42819284,"Rate":0.00325},{"Quantity":41.59490937,"Rate":0.00325655},{"Quantity":36.94444444,"Rate":0.00325996},{"Quantity":322.04953504,"Rate":0.00326263},{"Quantity":450,"Rate":0.00327},{"Quantity":785.75583452,"Rate":0.00328675},{"Quantity":300,"Rate":0.00328763},{"Quantity":167.20774693,"Rate":0.003299},{"Quantity":200,"Rate":0.0033},{"Quantity":60,"Rate":0.0033001},{"Quantity":16.69889259,"Rate":0.00330184},{"Quantity":450.97353375,"Rate":0.00332},{"Quantity":2000,"Rate":0.00332188},{"Quantity":55394.03726924,"Rate":0.00333},{"Quantity":60.37780398,"Rate":0.00333331},{"Quantity":1499.42537021,"Rate":0.00333461},{"Quantity":338.01986439,"Rate":0.00336113},{"Quantity":388.74945439,"Rate":0.00336119},{"Quantity":150,"Rate":0.00336187},{"Quantity":496.239545,"Rate":0.0034},{"Quantity":35.98572236,"Rate":0.00340036},{"Quantity":25,"Rate":0.0034009},{"Quantity":44,"Rate":0.00343434},{"Quantity":1453.64138944,"Rate":0.00343963},{"Quantity":826.75849011,"Rate":0.00345},{"Quantity":16.18602161,"Rate":0.00345113},{"Quantity":866.12758335,"Rate":0.00345989},{"Quantity":300,"Rate":0.00347103},{"Quantity":218.10721193,"Rate":0.003485},{"Quantity":1429.21960073,"Rate":0.00348577},{"Quantity":2000,"Rate":0.003492},{"Quantity":53317.05940099,"Rate":0.0035},{"Quantity":128.92197233,"Rate":0.00350009},{"Quantity":410,"Rate":0.00353},{"Quantity":1410.57053346,"Rate":0.00354466},{"Quantity":24.3596882,"Rate":0.00356},{"Quantity":100,"Rate":0.00357331},{"Quantity":295.75360093,"Rate":0.00359876},{"Quantity":26374.51481418,"Rate":0.0036},{"Quantity":2000,"Rate":0.00362188},{"Quantity":12.08923814,"Rate":0.00363607},{"Quantity":109.88868892,"Rate":0.00364},{"Quantity":600,"Rate":0.00367561},{"Quantity":6895.79817762,"Rate":0.00368},{"Quantity":6000,"Rate":0.00369},{"Quantity":6767,"Rate":0.0037},{"Quantity":278.19435664,"Rate":0.00372727},{"Quantity":155.5901487,"Rate":0.00375},{"Quantity":10000,"Rate":0.0038}],"ReturnTime":"1514114579641"}},"BAT-ETH":{"bittrex":{"Valid":true,"Error":"","Timestamp":"1514114579228","Bids":[{"Quantity":15173.85912685,"Rate":0.00047374},{"Quantity":130552,"Rate":0.00047363},{"Quantity":2149.78448276,"Rate":0.0004734},{"Quantity":0.97633788,"Rate":0.00047164},{"Quantity":7201.502,"Rate":0.00046531},{"Quantity":130552,"Rate":0.00046518},{"Quantity":215.90909091,"Rate":0.00046424},{"Quantity":2579.74137931,"Rate":0.000464},{"Quantity":4336.95652174,"Rate":0.00046},{"Quantity":109.23366697,"Rate":0.00045659},{"Quantity":1154.72771497,"Rate":0.000451},{"Quantity":1381.47874906,"Rate":0.00045001},{"Quantity":266.48138573,"Rate":0.00045},{"Quantity":2676.54778537,"Rate":0.00044536},{"Quantity":300,"Rate":0.00044534},{"Quantity":3714.71025259,"Rate":0.00044529},{"Quantity":2255.50342837,"Rate":0.00044336},{"Quantity":473.96047067,"Rate":0.00044197},{"Quantity":500,"Rate":0.00044196},{"Quantity":64.88347441,"Rate":0.00044088},{"Quantity":1000,"Rate":0.00044081},{"Quantity":1200,"Rate":0.0004401},{"Quantity":20637.7038762,"Rate":0.00044},{"Quantity":58452.01298605,"Rate":0.00043584},{"Quantity":10000,"Rate":0.0004344},{"Quantity":1000,"Rate":0.00043081},{"Quantity":1162.08803979,"Rate":0.00043026},{"Quantity":1782.2954076,"Rate":0.00043002},{"Quantity":200,"Rate":0.0004296},{"Quantity":15000,"Rate":0.000425},{"Quantity":1000,"Rate":0.00042081},{"Quantity":5437.61878837,"Rate":0.00042009},{"Quantity":150,"Rate":0.00042},{"Quantity":2418.18181818,"Rate":0.0004125},{"Quantity":242.38816125,"Rate":0.00041153},{"Quantity":1000,"Rate":0.00041081},{"Quantity":700.24147415,"Rate":0.00041022},{"Quantity":224.96510753,"Rate":0.00041013},{"Quantity":486.44299229,"Rate":0.00041012},{"Quantity":243.30798604,"Rate":0.00041},{"Quantity":1832.33540363,"Rate":0.00040829},{"Quantity":36534.35072646,"Rate":0.00040756},{"Quantity":369.20120617,"Rate":0.00040641},{"Quantity":26000,"Rate":0.000405},{"Quantity":23000,"Rate":0.00040499},{"Quantity":20000,"Rate":0.0004044},{"Quantity":183.23593452,"Rate":0.00040331},{"Quantity":515.15735087,"Rate":0.0004025},{"Quantity":100,"Rate":0.00040184},{"Quantity":1000,"Rate":0.00040081},{"Quantity":1200,"Rate":0.00040029},{"Quantity":45,"Rate":0.00040011},{"Quantity":1500,"Rate":0.0004001},{"Quantity":8355.67664769,"Rate":0.0004},{"Quantity":505.06329114,"Rate":0.000395},{"Quantity":126.8616954,"Rate":0.00039413},{"Quantity":778,"Rate":0.00039231},{"Quantity":50.11631089,"Rate":0.00039088},{"Quantity":1000,"Rate":0.00039081},{"Quantity":627.94591783,"Rate":0.00039027},{"Quantity":3197.11538461,"Rate":0.00039},{"Quantity":1190.64670916,"Rate":0.00038602},{"Quantity":45,"Rate":0.00038574},{"Quantity":553.40496818,"Rate":0.000385},{"Quantity":3136.76688436,"Rate":0.000382},{"Quantity":26073.70936169,"Rate":0.0003812},{"Quantity":1000,"Rate":0.00038081},{"Quantity":5900,"Rate":0.00038},{"Quantity":2037.74607726,"Rate":0.00037925},{"Quantity":1320.98209556,"Rate":0.00037756},{"Quantity":26.45888594,"Rate":0.000377},{"Quantity":2094.44060503,"Rate":0.0003756},{"Quantity":1000,"Rate":0.0003755},{"Quantity":2670.68273092,"Rate":0.0003735},{"Quantity":200,"Rate":0.00037347},{"Quantity":2681.4516129,"Rate":0.000372},{"Quantity":269.59459459,"Rate":0.00037},{"Quantity":3458.9818752,"Rate":0.000369},{"Quantity":1354.08465235,"Rate":0.00036833},{"Quantity":608,"Rate":0.00036818},{"Quantity":1359.99236496,"Rate":0.00036673},{"Quantity":273.28767123,"Rate":0.000365},{"Quantity":5000,"Rate":0.0003645},{"Quantity":6866.27862154,"Rate":0.00036319},{"Quantity":2755.52486188,"Rate":0.000362},{"Quantity":1000,"Rate":0.00036081},{"Quantity":823.56297384,"Rate":0.00036004},{"Quantity":1302.08333333,"Rate":0.00036},{"Quantity":435.67847868,"Rate":0.00035585},{"Quantity":1400,"Rate":0.00035555},{"Quantity":280.55914946,"Rate":0.00035554},{"Quantity":100,"Rate":0.00035535},{"Quantity":2820.18659881,"Rate":0.0003537},{"Quantity":2833.80681818,"Rate":0.000352},{"Quantity":4147.5,"Rate":0.00035},{"Quantity":368.36956522,"Rate":0.000345},{"Quantity":26.53679107,"Rate":0.00034492},{"Quantity":1453.86969829,"Rate":0.00034305},{"Quantity":2916.66666667,"Rate":0.000342},{"Quantity":4107.11761919,"Rate":0.00034}],"Asks":[{"Quantity":660.96951182,"Rate":0.00048652},{"Quantity":476.36673132,"Rate":0.00048663},{"Quantity":53661.5,"Rate":0.00048668},{"Quantity":174.46513359,"Rate":0.00048679},{"Quantity":359.78557803,"Rate":0.0004868},{"Quantity":535.3054103,"Rate":0.00048741},{"Quantity":2095.84928611,"Rate":0.0004875},{"Quantity":56.96731939,"Rate":0.00048795},{"Quantity":201.81944156,"Rate":0.00048889},{"Quantity":1591.40581442,"Rate":0.0004889},{"Quantity":80000,"Rate":0.0004895},{"Quantity":52.71165175,"Rate":0.00048956},{"Quantity":45,"Rate":0.00049179},{"Quantity":844.73910536,"Rate":0.00049194},{"Quantity":587.112123,"Rate":0.00049195},{"Quantity":15864.31819881,"Rate":0.000493},{"Quantity":2000,"Rate":0.00049478},{"Quantity":4000,"Rate":0.0004948},{"Quantity":7207.90019147,"Rate":0.00049504},{"Quantity":387.86495134,"Rate":0.00049776},{"Quantity":8026.26169709,"Rate":0.00049778},{"Quantity":397.73608311,"Rate":0.00049887},{"Quantity":493.77334994,"Rate":0.000499},{"Quantity":44.45671093,"Rate":0.00049974},{"Quantity":2961.99607852,"Rate":0.0004999},{"Quantity":851.25,"Rate":0.0005},{"Quantity":60.04321245,"Rate":0.00050548},{"Quantity":5000,"Rate":0.0005078},{"Quantity":10000,"Rate":0.0005084},{"Quantity":109.96388554,"Rate":0.00050846},{"Quantity":6000,"Rate":0.00050932},{"Quantity":2500,"Rate":0.00050934},{"Quantity":11105.72426414,"Rate":0.00050944},{"Quantity":19600,"Rate":0.00051},{"Quantity":586.45229638,"Rate":0.00051242},{"Quantity":205.93281566,"Rate":0.000515},{"Quantity":104.5298277,"Rate":0.000516},{"Quantity":2500,"Rate":0.00051934},{"Quantity":18221.97680839,"Rate":0.00051999},{"Quantity":19318.47826087,"Rate":0.00052},{"Quantity":47.62373841,"Rate":0.000522},{"Quantity":1000,"Rate":0.00052333},{"Quantity":44.01235435,"Rate":0.00052489},{"Quantity":146.61423898,"Rate":0.000525},{"Quantity":1600,"Rate":0.00052597},{"Quantity":100,"Rate":0.0005268},{"Quantity":7000,"Rate":0.0005274},{"Quantity":600,"Rate":0.00053},{"Quantity":2296.22733995,"Rate":0.00053499},{"Quantity":46.94117647,"Rate":0.000535},{"Quantity":4213.33557484,"Rate":0.00053587},{"Quantity":47.05235731,"Rate":0.000537},{"Quantity":2500,"Rate":0.00053934},{"Quantity":11437.90849673,"Rate":0.0005398},{"Quantity":17521.03196859,"Rate":0.00053998},{"Quantity":1000,"Rate":0.00053999},{"Quantity":4633,"Rate":0.00054},{"Quantity":550,"Rate":0.00054122},{"Quantity":102.39054989,"Rate":0.00054135},{"Quantity":45,"Rate":0.00054215},{"Quantity":492.20918042,"Rate":0.00054226},{"Quantity":4000,"Rate":0.000543},{"Quantity":1000,"Rate":0.000545},{"Quantity":425.5215085,"Rate":0.00054656},{"Quantity":11677.56725401,"Rate":0.00055},{"Quantity":5437.70103221,"Rate":0.00055308},{"Quantity":350,"Rate":0.000555},{"Quantity":100,"Rate":0.000558},{"Quantity":7000,"Rate":0.0005584},{"Quantity":190,"Rate":0.00055917},{"Quantity":5714.25094967,"Rate":0.00055968},{"Quantity":100,"Rate":0.00056},{"Quantity":1835.38332559,"Rate":0.00056263},{"Quantity":1500,"Rate":0.0005682},{"Quantity":190,"Rate":0.00056832},{"Quantity":4844.9821889,"Rate":0.00057},{"Quantity":67,"Rate":0.00057048},{"Quantity":110,"Rate":0.0005707},{"Quantity":5000,"Rate":0.00057192},{"Quantity":4999,"Rate":0.000572},{"Quantity":74.65362721,"Rate":0.000574},{"Quantity":50,"Rate":0.00057402},{"Quantity":341.84347826,"Rate":0.000575},{"Quantity":2000,"Rate":0.000577},{"Quantity":100,"Rate":0.00057832},{"Quantity":50,"Rate":0.00057863},{"Quantity":10500,"Rate":0.000579},{"Quantity":3500,"Rate":0.00057934},{"Quantity":10611.70212766,"Rate":0.00057999},{"Quantity":2337.02300462,"Rate":0.00058},{"Quantity":250,"Rate":0.0005802},{"Quantity":206,"Rate":0.00058204},{"Quantity":500,"Rate":0.000585},{"Quantity":2921.11493944,"Rate":0.00058569},{"Quantity":800,"Rate":0.00058689},{"Quantity":100,"Rate":0.00058832},{"Quantity":100,"Rate":0.000589},{"Quantity":50,"Rate":0.00058934},{"Quantity":3860.91525424,"Rate":0.00059},{"Quantity":264.86,"Rate":0.00059016}],"ReturnTime":"1514114579480"}},"CVC-ETH":{"bittrex":{"Valid":true,"Error":"","Timestamp":"1514114579228","Bids":[{"Quantity":128.67287333,"Rate":0.00099655},{"Quantity":500,"Rate":0.00098795},{"Quantity":45.30924007,"Rate":0.00098539},{"Quantity":4000,"Rate":0.00098017},{"Quantity":7512.582,"Rate":0.00098016},{"Quantity":1093.73996907,"Rate":0.00098001},{"Quantity":221.403105,"Rate":0.00098},{"Quantity":12716.00367794,"Rate":0.00094971},{"Quantity":2000,"Rate":0.0009497},{"Quantity":76064,"Rate":0.00094967},{"Quantity":12164.29699842,"Rate":0.0009495},{"Quantity":688.28600444,"Rate":0.00094948},{"Quantity":8460.91023362,"Rate":0.00094946},{"Quantity":10321.7652325,"Rate":0.00094945},{"Quantity":13360.76177549,"Rate":0.00094306},{"Quantity":76064,"Rate":0.00094051},{"Quantity":500,"Rate":0.00094009},{"Quantity":7457.69858803,"Rate":0.00092201},{"Quantity":27.13558188,"Rate":0.000919},{"Quantity":1644.23076923,"Rate":0.00091},{"Quantity":250,"Rate":0.000905},{"Quantity":110.63051073,"Rate":0.00090165},{"Quantity":754.82655499,"Rate":0.00090101},{"Quantity":1000,"Rate":0.0009001},{"Quantity":1030.22182667,"Rate":0.0009},{"Quantity":500,"Rate":0.000893},{"Quantity":392.73862422,"Rate":0.00088895},{"Quantity":3435.32424694,"Rate":0.00088569},{"Quantity":125.8876843,"Rate":0.000885},{"Quantity":113.81549891,"Rate":0.00087861},{"Quantity":285.23144494,"Rate":0.00087429},{"Quantity":457.8208188,"Rate":0.00087152},{"Quantity":86.99127907,"Rate":0.00086},{"Quantity":1000,"Rate":0.00085888},{"Quantity":100,"Rate":0.00085424},{"Quantity":373.78799944,"Rate":0.00085396},{"Quantity":937.72032902,"Rate":0.000851},{"Quantity":1025,"Rate":0.00085},{"Quantity":296.11290874,"Rate":0.00084105},{"Quantity":11517.29988715,"Rate":0.00083805},{"Quantity":1369.71488304,"Rate":0.00083438},{"Quantity":500,"Rate":0.00082618},{"Quantity":100,"Rate":0.00082},{"Quantity":10000,"Rate":0.00081526},{"Quantity":100,"Rate":0.000815},{"Quantity":100,"Rate":0.00080946},{"Quantity":100,"Rate":0.000805},{"Quantity":242.80282507,"Rate":0.00080099},{"Quantity":710.70986613,"Rate":0.00080001},{"Quantity":632.17874369,"Rate":0.0008},{"Quantity":500,"Rate":0.00079768},{"Quantity":5000.57782738,"Rate":0.00079457},{"Quantity":225.23965351,"Rate":0.00078964},{"Quantity":3706.98323692,"Rate":0.00078621},{"Quantity":100,"Rate":0.00078},{"Quantity":12014.08834254,"Rate":0.000767},{"Quantity":914.79844573,"Rate":0.00076623},{"Quantity":10556.12617302,"Rate":0.00076457},{"Quantity":945.19816657,"Rate":0.00076225},{"Quantity":333.6165032,"Rate":0.00075509},{"Quantity":100,"Rate":0.000755},{"Quantity":6000.57782738,"Rate":0.00075457},{"Quantity":100,"Rate":0.000754},{"Quantity":26.600266,"Rate":0.00075},{"Quantity":36.15100671,"Rate":0.000745},{"Quantity":1500,"Rate":0.00072},{"Quantity":500,"Rate":0.00071855},{"Quantity":1393.58226045,"Rate":0.00071614},{"Quantity":220,"Rate":0.00071609},{"Quantity":23.78750843,"Rate":0.00071184},{"Quantity":102.21986725,"Rate":0.00070961},{"Quantity":423.28528792,"Rate":0.00070697},{"Quantity":100,"Rate":0.000705},{"Quantity":26.7839734,"Rate":0.00070383},{"Quantity":2837.7167404,"Rate":0.00070303},{"Quantity":212.8681178,"Rate":0.0007029},{"Quantity":9455.52675988,"Rate":0.0007},{"Quantity":28.70503597,"Rate":0.000695},{"Quantity":14456.52173913,"Rate":0.00069},{"Quantity":11326.12820989,"Rate":0.00068708},{"Quantity":50,"Rate":0.00068139},{"Quantity":50,"Rate":0.00068138},{"Quantity":1500,"Rate":0.0006795},{"Quantity":36.7966801,"Rate":0.00067819},{"Quantity":2962.00614672,"Rate":0.00067353},{"Quantity":19657.95164143,"Rate":0.00066131},{"Quantity":9459.92397346,"Rate":0.0006613},{"Quantity":151.13636364,"Rate":0.00066},{"Quantity":152,"Rate":0.00065973},{"Quantity":110,"Rate":0.00065703},{"Quantity":340,"Rate":0.00065001},{"Quantity":23260.18769231,"Rate":0.00065},{"Quantity":1547.06329389,"Rate":0.00064477},{"Quantity":311.71875,"Rate":0.00064},{"Quantity":25,"Rate":0.00063201},{"Quantity":70,"Rate":0.000631},{"Quantity":316.66666667,"Rate":0.00063},{"Quantity":3197.88410676,"Rate":0.00062385},{"Quantity":150,"Rate":0.00062274},{"Quantity":3267.74193548,"Rate":0.00062}],"Asks":[{"Quantity":153.22180315,"Rate":0.001},{"Quantity":7010.72355807,"Rate":0.00101567},{"Quantity":2679.69026772,"Rate":0.00101568},{"Quantity":174.46513359,"Rate":0.00101569},{"Quantity":16728.9,"Rate":0.00101637},{"Quantity":497.31772495,"Rate":0.00101638},{"Quantity":29.14286678,"Rate":0.00104338},{"Quantity":73.21460741,"Rate":0.0010452},{"Quantity":38.40883433,"Rate":0.00105036},{"Quantity":500,"Rate":0.001059},{"Quantity":119.14484817,"Rate":0.00105973},{"Quantity":500,"Rate":0.00109},{"Quantity":91.87147805,"Rate":0.0011025},{"Quantity":37.640389,"Rate":0.00111},{"Quantity":27.50199389,"Rate":0.00111265},{"Quantity":2554.85097257,"Rate":0.00111499},{"Quantity":1000,"Rate":0.001115},{"Quantity":1000,"Rate":0.00111881},{"Quantity":33.56584821,"Rate":0.00112},{"Quantity":284.76,"Rate":0.00112026},{"Quantity":30,"Rate":0.00112878},{"Quantity":4892.5,"Rate":0.00113},{"Quantity":542.48906906,"Rate":0.00113766},{"Quantity":111.25777139,"Rate":0.00113975},{"Quantity":609.0559582,"Rate":0.00114},{"Quantity":6000,"Rate":0.00116636},{"Quantity":4000,"Rate":0.00116742},{"Quantity":29.46,"Rate":0.00117116},{"Quantity":913.17,"Rate":0.00117118},{"Quantity":153.35539109,"Rate":0.00118164},{"Quantity":1000,"Rate":0.001185},{"Quantity":200,"Rate":0.00118523},{"Quantity":50,"Rate":0.001189},{"Quantity":100,"Rate":0.00119},{"Quantity":353.50683624,"Rate":0.00119989},{"Quantity":2620.94264601,"Rate":0.0012},{"Quantity":4235.77886346,"Rate":0.00120999},{"Quantity":15000,"Rate":0.00121},{"Quantity":1305.32309377,"Rate":0.001215},{"Quantity":2200,"Rate":0.001216},{"Quantity":500,"Rate":0.001222},{"Quantity":147.29,"Rate":0.00122208},{"Quantity":130.93722492,"Rate":0.00122596},{"Quantity":40,"Rate":0.001228},{"Quantity":500,"Rate":0.00123371},{"Quantity":2898.7363,"Rate":0.00123456},{"Quantity":100,"Rate":0.00124},{"Quantity":466.85885689,"Rate":0.001245},{"Quantity":50,"Rate":0.001249},{"Quantity":958.64286208,"Rate":0.00125},{"Quantity":353.25,"Rate":0.00126},{"Quantity":80.02,"Rate":0.001263},{"Quantity":1000,"Rate":0.00126793},{"Quantity":5084.63509139,"Rate":0.001272},{"Quantity":392.76,"Rate":0.00127302},{"Quantity":50,"Rate":0.00128},{"Quantity":5960,"Rate":0.00129},{"Quantity":50,"Rate":0.00129214},{"Quantity":100,"Rate":0.00129889},{"Quantity":17332.70530102,"Rate":0.0013},{"Quantity":2418.31808155,"Rate":0.00130115},{"Quantity":246.55914279,"Rate":0.0013048},{"Quantity":5000,"Rate":0.00131},{"Quantity":200,"Rate":0.00131309},{"Quantity":98.19,"Rate":0.0013239},{"Quantity":98.19,"Rate":0.00132391},{"Quantity":196.38,"Rate":0.00132394},{"Quantity":60,"Rate":0.00133783},{"Quantity":5000,"Rate":0.00134},{"Quantity":199.6996997,"Rate":0.00134382},{"Quantity":76.30830824,"Rate":0.00134496},{"Quantity":39.67658443,"Rate":0.001347},{"Quantity":36.11655942,"Rate":0.00134758},{"Quantity":2000,"Rate":0.00134845},{"Quantity":1543.7234,"Rate":0.00135},{"Quantity":1680.5561263,"Rate":0.00136173},{"Quantity":3239.59178596,"Rate":0.00138},{"Quantity":2000,"Rate":0.00138845},{"Quantity":622,"Rate":0.00139},{"Quantity":7084.56628915,"Rate":0.0014},{"Quantity":250,"Rate":0.0014154},{"Quantity":15.23669083,"Rate":0.00143},{"Quantity":30683.46774194,"Rate":0.00144},{"Quantity":3000,"Rate":0.00144435},{"Quantity":942.81954375,"Rate":0.00144829},{"Quantity":2000,"Rate":0.00144845},{"Quantity":80.78011404,"Rate":0.00145},{"Quantity":52.09489746,"Rate":0.00146797},{"Quantity":525,"Rate":0.00147},{"Quantity":2000,"Rate":0.00148845},{"Quantity":300,"Rate":0.00148998},{"Quantity":2715.20603992,"Rate":0.00149},{"Quantity":15,"Rate":0.00149345},{"Quantity":12903.00242556,"Rate":0.00149999},{"Quantity":6299,"Rate":0.0015},{"Quantity":2418.31808155,"Rate":0.00150115},{"Quantity":50,"Rate":0.001505},{"Quantity":500,"Rate":0.00150769},{"Quantity":3837.60784909,"Rate":0.00152},{"Quantity":13712.67187668,"Rate":0.001524}],"ReturnTime":"1514114579642"}},"DGD-ETH":{"bittrex":{"Valid":true,"Error":"","Timestamp":"1514114579228","Bids":[{"Quantity":0.27508146,"Rate":0.21463293},{"Quantity":8.61103292,"Rate":0.21463292},{"Quantity":1,"Rate":0.21462222},{"Quantity":1.1705604,"Rate":0.2146222},{"Quantity":0.1859172,"Rate":0.214613},{"Quantity":2,"Rate":0.21454321},{"Quantity":0.58173319,"Rate":0.21075807},{"Quantity":0.79271622,"Rate":0.21075805},{"Quantity":3.24257964,"Rate":0.21074272},{"Quantity":11.70859117,"Rate":0.21074269},{"Quantity":19.12399583,"Rate":0.21074268},{"Quantity":10,"Rate":0.21},{"Quantity":0.2155875,"Rate":0.20821012},{"Quantity":4.58311056,"Rate":0.20733483},{"Quantity":7.88138352,"Rate":0.20733482},{"Quantity":22.26755771,"Rate":0.20684801},{"Quantity":14.46714496,"Rate":0.206848},{"Quantity":0.89773585,"Rate":0.20555829},{"Quantity":24.29387626,"Rate":0.20529865},{"Quantity":2,"Rate":0.20529223},{"Quantity":16.11095268,"Rate":0.20528206},{"Quantity":0.33934889,"Rate":0.20528205},{"Quantity":37.16346635,"Rate":0.20528201},{"Quantity":14.51724706,"Rate":0.205282},{"Quantity":9.71843336,"Rate":0.20528},{"Quantity":0.1,"Rate":0.20525837},{"Quantity":0.48611111,"Rate":0.2052},{"Quantity":0.21714056,"Rate":0.20500221},{"Quantity":20,"Rate":0.205},{"Quantity":5,"Rate":0.20485472},{"Quantity":0.24426444,"Rate":0.20469619},{"Quantity":0.109,"Rate":0.2035},{"Quantity":0.98518519,"Rate":0.2025},{"Quantity":9.8615917,"Rate":0.2023},{"Quantity":29.49765505,"Rate":0.20201545},{"Quantity":48.8576565,"Rate":0.20201542},{"Quantity":49.37742126,"Rate":0.20201541},{"Quantity":9.87553314,"Rate":0.20201441},{"Quantity":4.22211969,"Rate":0.2020144},{"Quantity":10.92174988,"Rate":0.201},{"Quantity":49.84733473,"Rate":0.200111},{"Quantity":0.87237583,"Rate":0.20010011},{"Quantity":5.98200571,"Rate":0.20000777},{"Quantity":5.4,"Rate":0.2000001},{"Quantity":18.64028334,"Rate":0.2},{"Quantity":0.25,"Rate":0.19992434},{"Quantity":0.25193035,"Rate":0.19846755},{"Quantity":10.05468337,"Rate":0.198415},{"Quantity":50.37369963,"Rate":0.19802},{"Quantity":40,"Rate":0.198},{"Quantity":0.14630443,"Rate":0.19662884},{"Quantity":25,"Rate":0.19600001},{"Quantity":0.5,"Rate":0.196},{"Quantity":0.30031136,"Rate":0.19571478},{"Quantity":40.8604975,"Rate":0.19529865},{"Quantity":1.96169594,"Rate":0.19377701},{"Quantity":25,"Rate":0.19089501},{"Quantity":0.15,"Rate":0.19055},{"Quantity":7.15736684,"Rate":0.19},{"Quantity":10.55980183,"Rate":0.188924},{"Quantity":53.24155319,"Rate":0.1873061},{"Quantity":24.99244987,"Rate":0.187},{"Quantity":2.66944357,"Rate":0.18683669},{"Quantity":3,"Rate":0.1851},{"Quantity":12.5528796,"Rate":0.18500001},{"Quantity":32.1740718,"Rate":0.18400005},{"Quantity":10.41089242,"Rate":0.18251},{"Quantity":25,"Rate":0.18250592},{"Quantity":5,"Rate":0.1823864},{"Quantity":3,"Rate":0.182},{"Quantity":0.1102221,"Rate":0.181},{"Quantity":6,"Rate":0.1801},{"Quantity":25,"Rate":0.18000013},{"Quantity":21.11177484,"Rate":0.18},{"Quantity":55.47830923,"Rate":0.1798},{"Quantity":85.7852,"Rate":0.1776001},{"Quantity":35,"Rate":0.17168884},{"Quantity":2,"Rate":0.171},{"Quantity":5,"Rate":0.1708113},{"Quantity":1.20180723,"Rate":0.166},{"Quantity":10,"Rate":0.1651},{"Quantity":85.7852,"Rate":0.1603},{"Quantity":5,"Rate":0.15917494},{"Quantity":3,"Rate":0.1557},{"Quantity":0.19726434,"Rate":0.1517},{"Quantity":1,"Rate":0.15},{"Quantity":50,"Rate":0.07168884},{"Quantity":2.2560524,"Rate":0.06},{"Quantity":75,"Rate":0.04168884},{"Quantity":10,"Rate":0.00600002},{"Quantity":1000,"Rate":0.00002},{"Quantity":1000,"Rate":0.00001},{"Quantity":100000,"Rate":1e-7},{"Quantity":100000,"Rate":4e-8},{"Quantity":33250,"Rate":3e-8},{"Quantity":1000000,"Rate":1e-8}],"Asks":[{"Quantity":1.43683366,"Rate":0.22554555},{"Quantity":0.10879304,"Rate":0.22554557},{"Quantity":0.06252449,"Rate":0.22554606},{"Quantity":13.60836551,"Rate":0.22999999},{"Quantity":10.1097584,"Rate":0.23},{"Quantity":0.23664215,"Rate":0.2315547},{"Quantity":0.13901393,"Rate":0.23275807},{"Quantity":1.001563,"Rate":0.23298608},{"Quantity":4.61805556,"Rate":0.2337},{"Quantity":0.25833296,"Rate":0.23434085},{"Quantity":10,"Rate":0.2348874},{"Quantity":5,"Rate":0.2348988},{"Quantity":14.14259945,"Rate":0.235},{"Quantity":2,"Rate":0.23564},{"Quantity":0.12955468,"Rate":0.23688812},{"Quantity":1.71126276,"Rate":0.23988},{"Quantity":1.05767954,"Rate":0.24},{"Quantity":0.22678495,"Rate":0.24199089},{"Quantity":5,"Rate":0.24342226},{"Quantity":0.4442651,"Rate":0.249},{"Quantity":25,"Rate":0.249899},{"Quantity":1.08234037,"Rate":0.25},{"Quantity":15,"Rate":0.25099999},{"Quantity":10,"Rate":0.25199999},{"Quantity":3,"Rate":0.2521},{"Quantity":10,"Rate":0.25299999},{"Quantity":0.15896505,"Rate":0.25377017},{"Quantity":10,"Rate":0.25399999},{"Quantity":0.16487079,"Rate":0.2549452},{"Quantity":10,"Rate":0.25499999},{"Quantity":0.22493387,"Rate":0.25539483},{"Quantity":10,"Rate":0.25599999},{"Quantity":10,"Rate":0.25699999},{"Quantity":10,"Rate":0.25799999},{"Quantity":0.1877907,"Rate":0.258},{"Quantity":5,"Rate":0.25881362},{"Quantity":10,"Rate":0.25899999},{"Quantity":10,"Rate":0.25999999},{"Quantity":0.83996695,"Rate":0.26},{"Quantity":3,"Rate":0.260999},{"Quantity":11.1959478,"Rate":0.26099999},{"Quantity":0.22097746,"Rate":0.26158763},{"Quantity":0.22096935,"Rate":0.26159551},{"Quantity":0.22094971,"Rate":0.26161462},{"Quantity":10,"Rate":0.26199999},{"Quantity":10,"Rate":0.26299999},{"Quantity":10,"Rate":0.26399999},{"Quantity":10,"Rate":0.26499999},{"Quantity":10,"Rate":0.26599999},{"Quantity":76,"Rate":0.26643723},{"Quantity":10,"Rate":0.26699999},{"Quantity":10,"Rate":0.26799999},{"Quantity":5,"Rate":0.26820484},{"Quantity":10,"Rate":0.26899999},{"Quantity":10,"Rate":0.26999999},{"Quantity":1,"Rate":0.27},{"Quantity":10,"Rate":0.27099999},{"Quantity":0.3389791,"Rate":0.27178252},{"Quantity":20,"Rate":0.27199999},{"Quantity":3,"Rate":0.2721},{"Quantity":20,"Rate":0.27299999},{"Quantity":10,"Rate":0.27399999},{"Quantity":10,"Rate":0.27499999},{"Quantity":10,"Rate":0.27599999},{"Quantity":0.20221566,"Rate":0.27666702},{"Quantity":10,"Rate":0.27699999},{"Quantity":10,"Rate":0.27799999},{"Quantity":10,"Rate":0.27899999},{"Quantity":5,"Rate":0.27932651},{"Quantity":10,"Rate":0.27999999},{"Quantity":0.95250138,"Rate":0.28},{"Quantity":10,"Rate":0.28093655},{"Quantity":10,"Rate":0.28099999},{"Quantity":10,"Rate":0.28199999},{"Quantity":10,"Rate":0.28299999},{"Quantity":10,"Rate":0.28399999},{"Quantity":10,"Rate":0.28499999},{"Quantity":10,"Rate":0.28599999},{"Quantity":10,"Rate":0.28699999},{"Quantity":10,"Rate":0.28799999},{"Quantity":10,"Rate":0.28899999},{"Quantity":10,"Rate":0.289},{"Quantity":34,"Rate":0.28905187},{"Quantity":10,"Rate":0.28999999},{"Quantity":0.1,"Rate":0.29},{"Quantity":10,"Rate":0.29093655},{"Quantity":10,"Rate":0.29099999},{"Quantity":10,"Rate":0.29199999},{"Quantity":3,"Rate":0.2921},{"Quantity":5,"Rate":0.29236391},{"Quantity":0.19597777,"Rate":0.29261538},{"Quantity":10,"Rate":0.29299999},{"Quantity":10,"Rate":0.29399999},{"Quantity":10,"Rate":0.29499999},{"Quantity":58.49999707,"Rate":0.29599998},{"Quantity":10,"Rate":0.29599999},{"Quantity":0.12521046,"Rate":0.296},{"Quantity":10,"Rate":0.29699999},{"Quantity":10,"Rate":0.29799999},{"Quantity":25,"Rate":0.2989}],"ReturnTime":"1514114579641"}},"FUN-ETH":{"bittrex":{"Valid":true,"Error":"","Timestamp":"1514114579228","Bids":[{"Quantity":3550.94852427,"Rate":0.00008065},{"Quantity":24900,"Rate":0.00008064},{"Quantity":489855.39183168,"Rate":0.00008063},{"Quantity":1162.04306583,"Rate":0.00008062},{"Quantity":1260.13346331,"Rate":0.00008053},{"Quantity":38797.51349813,"Rate":0.0000802},{"Quantity":10000,"Rate":0.00008},{"Quantity":4117.23305821,"Rate":0.00007823},{"Quantity":8993.80726037,"Rate":0.00007821},{"Quantity":441370,"Rate":0.00007745},{"Quantity":1591.52751716,"Rate":0.00007703},{"Quantity":300,"Rate":0.00007673},{"Quantity":2000,"Rate":0.0000766},{"Quantity":1742.20209602,"Rate":0.00007655},{"Quantity":177412.36647165,"Rate":0.0000765},{"Quantity":128930.40389422,"Rate":0.00007601},{"Quantity":142179.96541875,"Rate":0.000076},{"Quantity":5276.38190955,"Rate":0.00007562},{"Quantity":685,"Rate":0.00007556},{"Quantity":5000,"Rate":0.0000755},{"Quantity":4912.96738266,"Rate":0.00007542},{"Quantity":49714.27167164,"Rate":0.00007533},{"Quantity":26853.15679409,"Rate":0.00007501},{"Quantity":6000.68819074,"Rate":0.000075},{"Quantity":7240.05263818,"Rate":0.000074},{"Quantity":21006.30579702,"Rate":0.0000739},{"Quantity":6821.20061836,"Rate":0.00007346},{"Quantity":12000,"Rate":0.00007194},{"Quantity":8667.69187754,"Rate":0.00007171},{"Quantity":5700.39109632,"Rate":0.00007101},{"Quantity":5000,"Rate":0.000071},{"Quantity":647.4491656,"Rate":0.00007029},{"Quantity":28422.85225816,"Rate":0.00007019},{"Quantity":1000,"Rate":0.00007013},{"Quantity":1138.37375178,"Rate":0.0000701},{"Quantity":216576.484365,"Rate":0.00007},{"Quantity":24557.63512275,"Rate":0.00006986},{"Quantity":1000,"Rate":0.0000698},{"Quantity":6969,"Rate":0.00006969},{"Quantity":8675.32531722,"Rate":0.00006951},{"Quantity":7821.44028705,"Rate":0.00006924},{"Quantity":6078.39611739,"Rate":0.000069},{"Quantity":212016.25378678,"Rate":0.00006899},{"Quantity":60143.38235294,"Rate":0.000068},{"Quantity":6094.62157633,"Rate":0.00006761},{"Quantity":863.76534177,"Rate":0.00006759},{"Quantity":5000,"Rate":0.00006751},{"Quantity":1477.77777778,"Rate":0.0000675},{"Quantity":188031.63706909,"Rate":0.00006701},{"Quantity":30000,"Rate":0.000067},{"Quantity":6969,"Rate":0.00006699},{"Quantity":1000,"Rate":0.00006684},{"Quantity":2117.51326763,"Rate":0.00006595},{"Quantity":12164.63414634,"Rate":0.0000656},{"Quantity":27230.22349351,"Rate":0.0000655},{"Quantity":10702.78969957,"Rate":0.00006524},{"Quantity":702.3480944,"Rate":0.00006515},{"Quantity":77480.60232907,"Rate":0.0000645},{"Quantity":11111,"Rate":0.00006444},{"Quantity":38893.9453125,"Rate":0.000064},{"Quantity":1060.17153572,"Rate":0.00006383},{"Quantity":8888,"Rate":0.00006364},{"Quantity":8000,"Rate":0.0000636},{"Quantity":2748.58290033,"Rate":0.00006351},{"Quantity":314.6719164,"Rate":0.0000634},{"Quantity":52560.51634297,"Rate":0.00006333},{"Quantity":2810,"Rate":0.0000631},{"Quantity":2000,"Rate":0.000063},{"Quantity":5555,"Rate":0.00006289},{"Quantity":1100,"Rate":0.0000628},{"Quantity":2085.96529278,"Rate":0.00006276},{"Quantity":2795.34426068,"Rate":0.00006249},{"Quantity":2130.53487077,"Rate":0.0000621},{"Quantity":573.31071169,"Rate":0.000062},{"Quantity":36850.79083374,"Rate":0.0000618},{"Quantity":419.60975743,"Rate":0.00006171},{"Quantity":4864.31118214,"Rate":0.00006152},{"Quantity":350,"Rate":0.00006151},{"Quantity":3519.77996691,"Rate":0.0000612},{"Quantity":5555,"Rate":0.00006119},{"Quantity":4000,"Rate":0.0000611},{"Quantity":49057.36691066,"Rate":0.000061},{"Quantity":6969,"Rate":0.0000609},{"Quantity":411.69839088,"Rate":0.00006053},{"Quantity":350,"Rate":0.00006028},{"Quantity":2000,"Rate":0.00006022},{"Quantity":3000,"Rate":0.00006001},{"Quantity":20125.20797875,"Rate":0.00006},{"Quantity":2000,"Rate":0.00005921},{"Quantity":8888,"Rate":0.00005903},{"Quantity":1000,"Rate":0.000059},{"Quantity":1694.1236413,"Rate":0.00005888},{"Quantity":116.22501162,"Rate":0.00005842},{"Quantity":30000,"Rate":0.00005829},{"Quantity":855.92929466,"Rate":0.00005827},{"Quantity":17183.46270413,"Rate":0.00005805},{"Quantity":534.41137912,"Rate":0.00005802},{"Quantity":2584.19689119,"Rate":0.0000579},{"Quantity":1000,"Rate":0.00005783},{"Quantity":64889.22789467,"Rate":0.00005777}],"Asks":[{"Quantity":3635.15493421,"Rate":0.00008282},{"Quantity":3905.9918732,"Rate":0.00008293},{"Quantity":1952.93876331,"Rate":0.00008366},{"Quantity":482.29480692,"Rate":0.00008367},{"Quantity":440.63875124,"Rate":0.000085},{"Quantity":4182.22880753,"Rate":0.00008769},{"Quantity":87549.23315332,"Rate":0.00009056},{"Quantity":76099.79815414,"Rate":0.00009057},{"Quantity":5000,"Rate":0.00009096},{"Quantity":2000,"Rate":0.0000913},{"Quantity":440.83802205,"Rate":0.00009152},{"Quantity":1500,"Rate":0.00009226},{"Quantity":1000,"Rate":0.00009289},{"Quantity":7568.64032698,"Rate":0.0000929},{"Quantity":47910,"Rate":0.00009295},{"Quantity":2025057.17357056,"Rate":0.00009296},{"Quantity":3000,"Rate":0.000093},{"Quantity":250,"Rate":0.00009388},{"Quantity":2200,"Rate":0.00009396},{"Quantity":104000,"Rate":0.000094},{"Quantity":100,"Rate":0.00009469},{"Quantity":493.7539375,"Rate":0.000095},{"Quantity":4519.24285822,"Rate":0.00009527},{"Quantity":5000,"Rate":0.00009577},{"Quantity":1500,"Rate":0.0000959},{"Quantity":1275.31559125,"Rate":0.000096},{"Quantity":558.95504824,"Rate":0.00009613},{"Quantity":22969.07110899,"Rate":0.00009674},{"Quantity":450,"Rate":0.00009687},{"Quantity":20408.17348415,"Rate":0.00009698},{"Quantity":5805.73601648,"Rate":0.000098},{"Quantity":40000,"Rate":0.00009804},{"Quantity":700,"Rate":0.00009879},{"Quantity":15000,"Rate":0.0000988},{"Quantity":2192.66604282,"Rate":0.00009895},{"Quantity":11211,"Rate":0.00009924},{"Quantity":500,"Rate":0.0000999},{"Quantity":124630.9425987,"Rate":0.0001},{"Quantity":36425.64917396,"Rate":0.000103},{"Quantity":359.18256131,"Rate":0.00010355},{"Quantity":500,"Rate":0.0001039},{"Quantity":10000,"Rate":0.00010456},{"Quantity":1553.25443787,"Rate":0.00010458},{"Quantity":19810,"Rate":0.000105},{"Quantity":197692.53787241,"Rate":0.00010502},{"Quantity":100,"Rate":0.00010669},{"Quantity":2000,"Rate":0.000107},{"Quantity":163.19767442,"Rate":0.0001075},{"Quantity":500,"Rate":0.0001079},{"Quantity":1000,"Rate":0.000109},{"Quantity":1100,"Rate":0.00010979},{"Quantity":165689.75995475,"Rate":0.00011},{"Quantity":449999,"Rate":0.00011111},{"Quantity":500,"Rate":0.0001119},{"Quantity":216.37363852,"Rate":0.000112},{"Quantity":5000,"Rate":0.00011271},{"Quantity":100,"Rate":0.0001129},{"Quantity":19910.15586421,"Rate":0.0001132},{"Quantity":13810,"Rate":0.000115},{"Quantity":500,"Rate":0.0001159},{"Quantity":5000,"Rate":0.000117},{"Quantity":70000,"Rate":0.00011804},{"Quantity":4832,"Rate":0.00011822},{"Quantity":500,"Rate":0.0001199},{"Quantity":163450.95854772,"Rate":0.00012},{"Quantity":333.46120186,"Rate":0.00012388},{"Quantity":500,"Rate":0.0001239},{"Quantity":1000,"Rate":0.00012437},{"Quantity":10000,"Rate":0.00012456},{"Quantity":45000,"Rate":0.00012555},{"Quantity":2000,"Rate":0.00012574},{"Quantity":2400,"Rate":0.000126},{"Quantity":137.59803922,"Rate":0.0001275},{"Quantity":500,"Rate":0.00012979},{"Quantity":100000,"Rate":0.00013},{"Quantity":100,"Rate":0.000135},{"Quantity":4518.44185,"Rate":0.0001369},{"Quantity":600,"Rate":0.000137},{"Quantity":29173.26915443,"Rate":0.00013804},{"Quantity":102452.67465919,"Rate":0.00014},{"Quantity":13942.21589092,"Rate":0.000144},{"Quantity":2000,"Rate":0.0001444},{"Quantity":50000,"Rate":0.00014495},{"Quantity":3739.72950629,"Rate":0.00014555},{"Quantity":16500,"Rate":0.00014777},{"Quantity":350,"Rate":0.00015},{"Quantity":2818,"Rate":0.0001517},{"Quantity":15079.36507937,"Rate":0.00015243},{"Quantity":600,"Rate":0.000157},{"Quantity":6362.24478194,"Rate":0.00015805},{"Quantity":151.89393939,"Rate":0.000165},{"Quantity":15000,"Rate":0.00017},{"Quantity":4020.48408803,"Rate":0.00019},{"Quantity":1857.98490729,"Rate":0.00019145},{"Quantity":15000,"Rate":0.0002},{"Quantity":5000,"Rate":0.00020964},{"Quantity":500,"Rate":0.00021},{"Quantity":1000,"Rate":0.00021111},{"Quantity":136.70454545,"Rate":0.00022},{"Quantity":400,"Rate":0.00023}],"ReturnTime":"1514114579574"}},"GNT-ETH":{"bittrex":{"Valid":true,"Error":"","Timestamp":"1514114579228","Bids":[{"Quantity":1e-8,"Rate":0.0008552},{"Quantity":2000,"Rate":0.00084661},{"Quantity":35.4,"Rate":0.0008466},{"Quantity":796.799,"Rate":0.00084047},{"Quantity":392.97250096,"Rate":0.00084029},{"Quantity":4000,"Rate":0.00084028},{"Quantity":1666.13113642,"Rate":0.00084027},{"Quantity":23312.12574516,"Rate":0.00084},{"Quantity":374.35257939,"Rate":0.00083637},{"Quantity":186.65207978,"Rate":0.00083634},{"Quantity":6728.914,"Rate":0.00083221},{"Quantity":376.20155501,"Rate":0.00083219},{"Quantity":179.80550614,"Rate":0.00083218},{"Quantity":100,"Rate":0.0008308},{"Quantity":149.56495348,"Rate":0.00083006},{"Quantity":640.81815639,"Rate":0.00083},{"Quantity":1000,"Rate":0.00082904},{"Quantity":151.90425504,"Rate":0.000829},{"Quantity":483.06254389,"Rate":0.00082598},{"Quantity":87100,"Rate":0.00082233},{"Quantity":130,"Rate":0.0008223},{"Quantity":170.21982634,"Rate":0.000822},{"Quantity":3646.13631375,"Rate":0.00082},{"Quantity":123.04962991,"Rate":0.00081268},{"Quantity":100,"Rate":0.00081212},{"Quantity":267.51614084,"Rate":0.0008121},{"Quantity":100,"Rate":0.00081206},{"Quantity":500,"Rate":0.00081},{"Quantity":193.8477436,"Rate":0.0008053},{"Quantity":25,"Rate":0.000801},{"Quantity":1997.08085009,"Rate":0.0008},{"Quantity":87100,"Rate":0.00079829},{"Quantity":40,"Rate":0.00079808},{"Quantity":400,"Rate":0.000798},{"Quantity":65.79964601,"Rate":0.00079791},{"Quantity":576.02544114,"Rate":0.00079},{"Quantity":3356.34706925,"Rate":0.00078555},{"Quantity":380,"Rate":0.00077001},{"Quantity":200,"Rate":0.00076661},{"Quantity":278.89723005,"Rate":0.00076059},{"Quantity":8.28494009,"Rate":0.0007605},{"Quantity":30,"Rate":0.00076024},{"Quantity":40,"Rate":0.00076012},{"Quantity":196.8490988,"Rate":0.0007601},{"Quantity":100,"Rate":0.00076009},{"Quantity":100,"Rate":0.00076001},{"Quantity":2300,"Rate":0.00076},{"Quantity":985.36036036,"Rate":0.00075924},{"Quantity":504.35032358,"Rate":0.00075337},{"Quantity":5633,"Rate":0.00075},{"Quantity":3370.19737171,"Rate":0.0007419},{"Quantity":135.25423729,"Rate":0.0007375},{"Quantity":132.98,"Rate":0.00073004},{"Quantity":1851.86015329,"Rate":0.00073},{"Quantity":710,"Rate":0.0007254},{"Quantity":137.5862069,"Rate":0.000725},{"Quantity":2200,"Rate":0.0007234},{"Quantity":2500,"Rate":0.00072001},{"Quantity":17653.68536094,"Rate":0.00072},{"Quantity":200,"Rate":0.000715},{"Quantity":140,"Rate":0.0007125},{"Quantity":14210.83391715,"Rate":0.00071},{"Quantity":1059.27726521,"Rate":0.000705},{"Quantity":145.5,"Rate":0.0007021},{"Quantity":28.45863815,"Rate":0.00070103},{"Quantity":18829.70528658,"Rate":0.00070102},{"Quantity":88616.00426478,"Rate":0.00070101},{"Quantity":8952.5,"Rate":0.0007},{"Quantity":3000,"Rate":0.00069},{"Quantity":145.09027597,"Rate":0.0006875},{"Quantity":254.79113148,"Rate":0.00068512},{"Quantity":4368.61313869,"Rate":0.000685},{"Quantity":590.40010975,"Rate":0.00068275},{"Quantity":5000,"Rate":0.000681},{"Quantity":311.85145441,"Rate":0.00068},{"Quantity":1477.71210169,"Rate":0.00067503},{"Quantity":7077.28699206,"Rate":0.000672},{"Quantity":100,"Rate":0.00066708},{"Quantity":1495.43498793,"Rate":0.00066703},{"Quantity":29939.97747748,"Rate":0.00066607},{"Quantity":1200,"Rate":0.000666},{"Quantity":300,"Rate":0.000665},{"Quantity":150.55921996,"Rate":0.00066253},{"Quantity":27575.33976757,"Rate":0.00066001},{"Quantity":11952.11336534,"Rate":0.00066},{"Quantity":75.61400849,"Rate":0.0006596},{"Quantity":100,"Rate":0.000659},{"Quantity":279.2543611,"Rate":0.00065767},{"Quantity":1000,"Rate":0.000655},{"Quantity":76.61290323,"Rate":0.000651},{"Quantity":3068.28668102,"Rate":0.0006502},{"Quantity":30,"Rate":0.0006501},{"Quantity":153.45445595,"Rate":0.00065003},{"Quantity":5859.24230769,"Rate":0.00065},{"Quantity":100,"Rate":0.00064841},{"Quantity":100,"Rate":0.00064818},{"Quantity":55.48593086,"Rate":0.00064714},{"Quantity":45,"Rate":0.00064236},{"Quantity":4674.32052484,"Rate":0.0006402},{"Quantity":2650,"Rate":0.00064}],"Asks":[{"Quantity":7209.279,"Rate":0.00086879},{"Quantity":399.58082001,"Rate":0.0008688},{"Quantity":7185.948,"Rate":0.00086893},{"Quantity":1500,"Rate":0.00086894},{"Quantity":46.54230117,"Rate":0.00086915},{"Quantity":348.81152284,"Rate":0.00086933},{"Quantity":175.90396057,"Rate":0.00086969},{"Quantity":1712.324,"Rate":0.00087},{"Quantity":60,"Rate":0.00087037},{"Quantity":47.5488705,"Rate":0.00087505},{"Quantity":34.7,"Rate":0.0008811},{"Quantity":34.4,"Rate":0.00089},{"Quantity":61,"Rate":0.00089076},{"Quantity":60,"Rate":0.00089796},{"Quantity":60,"Rate":0.00089887},{"Quantity":52.55034373,"Rate":0.0008989},{"Quantity":33.7,"Rate":0.000908},{"Quantity":465.56042,"Rate":0.00091628},{"Quantity":33.4,"Rate":0.0009172},{"Quantity":3000,"Rate":0.00092383},{"Quantity":33,"Rate":0.0009265},{"Quantity":500,"Rate":0.00092993},{"Quantity":10000,"Rate":0.00092994},{"Quantity":32.7,"Rate":0.0009359},{"Quantity":500,"Rate":0.00093666},{"Quantity":2000,"Rate":0.00093667},{"Quantity":933.15775593,"Rate":0.00094},{"Quantity":500,"Rate":0.00094177},{"Quantity":10000,"Rate":0.00094178},{"Quantity":1e-8,"Rate":0.00094181},{"Quantity":32.4,"Rate":0.0009453},{"Quantity":10250,"Rate":0.00095},{"Quantity":10000,"Rate":0.000952},{"Quantity":90.09224591,"Rate":0.00095225},{"Quantity":10000,"Rate":0.000954},{"Quantity":32,"Rate":0.0009549},{"Quantity":7600.35178771,"Rate":0.00095599},{"Quantity":10000,"Rate":0.000956},{"Quantity":3378.77049153,"Rate":0.00095609},{"Quantity":10000,"Rate":0.000958},{"Quantity":11383.21144485,"Rate":0.00096},{"Quantity":10000,"Rate":0.000962},{"Quantity":10000,"Rate":0.000964},{"Quantity":10031.7,"Rate":0.0009646},{"Quantity":10000,"Rate":0.000966},{"Quantity":10000,"Rate":0.000968},{"Quantity":42.1863264,"Rate":0.00096821},{"Quantity":10000,"Rate":0.0009689},{"Quantity":2330,"Rate":0.00096997},{"Quantity":460.73910238,"Rate":0.00096998},{"Quantity":13000.77041602,"Rate":0.00096999},{"Quantity":68279.04967743,"Rate":0.00097},{"Quantity":2543.85483223,"Rate":0.0009706},{"Quantity":10000,"Rate":0.000972},{"Quantity":10000,"Rate":0.000974},{"Quantity":31.4,"Rate":0.0009744},{"Quantity":1000,"Rate":0.00097443},{"Quantity":10000,"Rate":0.000976},{"Quantity":4000,"Rate":0.00097609},{"Quantity":10000,"Rate":0.000978},{"Quantity":6479.74490268,"Rate":0.00097999},{"Quantity":112473.36191824,"Rate":0.00098},{"Quantity":9,"Rate":0.000981},{"Quantity":10000,"Rate":0.000982},{"Quantity":4500,"Rate":0.0009821},{"Quantity":54150.56164139,"Rate":0.00098333},{"Quantity":10000,"Rate":0.000984},{"Quantity":30.6,"Rate":0.0009843},{"Quantity":100,"Rate":0.000985},{"Quantity":10000,"Rate":0.000986},{"Quantity":500,"Rate":0.00098676},{"Quantity":10000,"Rate":0.000988},{"Quantity":40,"Rate":0.000989},{"Quantity":200,"Rate":0.0009899},{"Quantity":10295,"Rate":0.00099},{"Quantity":50,"Rate":0.00099049},{"Quantity":10000,"Rate":0.000992},{"Quantity":720,"Rate":0.0009928},{"Quantity":10000,"Rate":0.000994},{"Quantity":1000,"Rate":0.00099443},{"Quantity":1600,"Rate":0.000995},{"Quantity":200,"Rate":0.00099565},{"Quantity":10000,"Rate":0.000996},{"Quantity":44.69,"Rate":0.00099607},{"Quantity":1000,"Rate":0.00099609},{"Quantity":4994.87342023,"Rate":0.000997},{"Quantity":10000,"Rate":0.000998},{"Quantity":100000,"Rate":0.000999},{"Quantity":316.02617953,"Rate":0.0009998},{"Quantity":9048.67102447,"Rate":0.0009999},{"Quantity":577170.05274196,"Rate":0.001},{"Quantity":100,"Rate":0.00100394},{"Quantity":30,"Rate":0.0010089},{"Quantity":100036.75085645,"Rate":0.00101},{"Quantity":4500,"Rate":0.0010121},{"Quantity":1100,"Rate":0.00101271},{"Quantity":1000,"Rate":0.00101443},{"Quantity":100075,"Rate":0.00102},{"Quantity":1817.64129448,"Rate":0.00102109},{"Quantity":100000,"Rate":0.00103}],"ReturnTime":"1514114579457"}},"MCO-ETH":{"bittrex":{"Valid":true,"Error":"","Timestamp":"1514114579228","Bids":[{"Quantity":142.93534777,"Rate":0.02437378},{"Quantity":1.21116959,"Rate":0.02437377},{"Quantity":1.63701658,"Rate":0.02437376},{"Quantity":627.72475623,"Rate":0.02437374},{"Quantity":645.32564093,"Rate":0.02400004},{"Quantity":2672.98453105,"Rate":0.02400003},{"Quantity":470.5,"Rate":0.02371601},{"Quantity":12.00169742,"Rate":0.023716},{"Quantity":2629.4,"Rate":0.02335391},{"Quantity":470.92274678,"Rate":0.0233},{"Quantity":32.86462517,"Rate":0.02315},{"Quantity":25.80592674,"Rate":0.02312284},{"Quantity":439.11674798,"Rate":0.02300001},{"Quantity":4.40285759,"Rate":0.02271252},{"Quantity":5.94862184,"Rate":0.02271244},{"Quantity":99.63199957,"Rate":0.0223013},{"Quantity":695.45422933,"Rate":0.02200001},{"Quantity":4069.24915178,"Rate":0.022},{"Quantity":10,"Rate":0.02167546},{"Quantity":47.08420645,"Rate":0.0211889},{"Quantity":759.07181462,"Rate":0.02102581},{"Quantity":8.54958473,"Rate":0.02100102},{"Quantity":711.9044229,"Rate":0.02100001},{"Quantity":5914.55043975,"Rate":0.021},{"Quantity":47.83107883,"Rate":0.02095026},{"Quantity":2.38731514,"Rate":0.02094403},{"Quantity":10,"Rate":0.0209432},{"Quantity":2,"Rate":0.02073241},{"Quantity":2.71027645,"Rate":0.02061808},{"Quantity":3,"Rate":0.02040439},{"Quantity":6.5,"Rate":0.02020003},{"Quantity":210,"Rate":0.02005728},{"Quantity":24.93245866,"Rate":0.02000404},{"Quantity":1249.09763143,"Rate":0.02},{"Quantity":309.50085477,"Rate":0.01994},{"Quantity":111,"Rate":0.0199},{"Quantity":1,"Rate":0.01989566},{"Quantity":2.51826117,"Rate":0.01985497},{"Quantity":2.66852583,"Rate":0.01966834},{"Quantity":763.25623428,"Rate":0.01965264},{"Quantity":7.77846079,"Rate":0.01952449},{"Quantity":50,"Rate":0.0195},{"Quantity":2.46912013,"Rate":0.01934929},{"Quantity":646.40589943,"Rate":0.0193},{"Quantity":547.84162304,"Rate":0.0191},{"Quantity":5.250926,"Rate":0.01904426},{"Quantity":7,"Rate":0.019},{"Quantity":2.63192248,"Rate":0.01899752},{"Quantity":1,"Rate":0.01888888},{"Quantity":142.1875,"Rate":0.01824},{"Quantity":123.69242321,"Rate":0.018},{"Quantity":55.555586,"Rate":0.01792799},{"Quantity":0.55726814,"Rate":0.0179},{"Quantity":23.2336235,"Rate":0.01786049},{"Quantity":5,"Rate":0.0178},{"Quantity":1,"Rate":0.01777777},{"Quantity":1.3,"Rate":0.0177702},{"Quantity":11.27734119,"Rate":0.01769034},{"Quantity":2,"Rate":0.0170668},{"Quantity":5,"Rate":0.01703},{"Quantity":358.3760233,"Rate":0.017},{"Quantity":1,"Rate":0.0165},{"Quantity":1.22392638,"Rate":0.0163},{"Quantity":73.1171875,"Rate":0.016},{"Quantity":20,"Rate":0.0159601},{"Quantity":5,"Rate":0.01581391},{"Quantity":30.75654495,"Rate":0.0155},{"Quantity":5.15642447,"Rate":0.01547584},{"Quantity":6.25980392,"Rate":0.0153},{"Quantity":100,"Rate":0.01501},{"Quantity":299.31574402,"Rate":0.015},{"Quantity":1.50966466,"Rate":0.0145},{"Quantity":2,"Rate":0.014},{"Quantity":368.8545148,"Rate":0.01389839},{"Quantity":14.42304564,"Rate":0.01383203},{"Quantity":1101.45801868,"Rate":0.01352},{"Quantity":73.88893101,"Rate":0.0135},{"Quantity":2.26446524,"Rate":0.01303651},{"Quantity":15,"Rate":0.013},{"Quantity":1.93313953,"Rate":0.0129},{"Quantity":78.15364654,"Rate":0.01276332},{"Quantity":0.57627142,"Rate":0.01272075},{"Quantity":10,"Rate":0.01272},{"Quantity":8.11867683,"Rate":0.0123},{"Quantity":1.22142857,"Rate":0.01225},{"Quantity":330.83454964,"Rate":0.01223},{"Quantity":0.4,"Rate":0.01222222},{"Quantity":16.59102987,"Rate":0.01202457},{"Quantity":1.69567689,"Rate":0.0120069},{"Quantity":186.98351082,"Rate":0.012},{"Quantity":2,"Rate":0.0113},{"Quantity":60,"Rate":0.011},{"Quantity":20,"Rate":0.01085},{"Quantity":207.79247086,"Rate":0.01},{"Quantity":5,"Rate":0.0099892},{"Quantity":10,"Rate":0.00888},{"Quantity":1000,"Rate":0.0001},{"Quantity":1000,"Rate":0.0000892},{"Quantity":1000,"Rate":0.00002011},{"Quantity":100000,"Rate":1e-7}],"Asks":[{"Quantity":15.39680469,"Rate":0.02503471},{"Quantity":18.71484714,"Rate":0.02503534},{"Quantity":93.57423573,"Rate":0.02503537},{"Quantity":81.2688869,"Rate":0.0259857},{"Quantity":69.86302833,"Rate":0.02598573},{"Quantity":758.6,"Rate":0.02598574},{"Quantity":7.40401936,"Rate":0.02598711},{"Quantity":1,"Rate":0.02599595},{"Quantity":3.46154644,"Rate":0.02599994},{"Quantity":13.50673005,"Rate":0.026},{"Quantity":2,"Rate":0.02602259},{"Quantity":7.62521169,"Rate":0.02622878},{"Quantity":1.5,"Rate":0.02625},{"Quantity":4.71789988,"Rate":0.02696523},{"Quantity":37.49925282,"Rate":0.02699125},{"Quantity":1.4768473,"Rate":0.02732834},{"Quantity":4.64254703,"Rate":0.02748995},{"Quantity":4.65162053,"Rate":0.02756604},{"Quantity":6.20029503,"Rate":0.028},{"Quantity":133.15979409,"Rate":0.0283},{"Quantity":1.5,"Rate":0.02835},{"Quantity":3,"Rate":0.02869999},{"Quantity":8.32071509,"Rate":0.0287},{"Quantity":10,"Rate":0.02879},{"Quantity":6.22,"Rate":0.0288},{"Quantity":1.21787492,"Rate":0.02881043},{"Quantity":3,"Rate":0.02884},{"Quantity":200,"Rate":0.02898},{"Quantity":16.05563002,"Rate":0.029},{"Quantity":2.05627705,"Rate":0.02909439},{"Quantity":11.60939591,"Rate":0.02912918},{"Quantity":11.47498005,"Rate":0.0294},{"Quantity":0.50974576,"Rate":0.0295},{"Quantity":101.73511472,"Rate":0.02951},{"Quantity":10.55155524,"Rate":0.02958},{"Quantity":9.02406644,"Rate":0.02962012},{"Quantity":0.83824368,"Rate":0.0298},{"Quantity":5.3,"Rate":0.0299},{"Quantity":30,"Rate":0.02991875},{"Quantity":372.75784753,"Rate":0.02999},{"Quantity":440.7405775,"Rate":0.02999999},{"Quantity":2570.81219445,"Rate":0.03},{"Quantity":29.55254327,"Rate":0.03001},{"Quantity":1.5,"Rate":0.03045},{"Quantity":33.23,"Rate":0.03099999},{"Quantity":251.00780534,"Rate":0.031},{"Quantity":6.76804983,"Rate":0.03109158},{"Quantity":20,"Rate":0.031099},{"Quantity":33.81355932,"Rate":0.03111111},{"Quantity":5,"Rate":0.0313555},{"Quantity":0.4,"Rate":0.03137499},{"Quantity":194.68642706,"Rate":0.0315},{"Quantity":4.44756772,"Rate":0.03154857},{"Quantity":3.15887971,"Rate":0.03166},{"Quantity":6.35220453,"Rate":0.031719},{"Quantity":20,"Rate":0.03175},{"Quantity":3.76,"Rate":0.03191443},{"Quantity":1.88,"Rate":0.0319146},{"Quantity":14.07828679,"Rate":0.03199165},{"Quantity":4625.85085048,"Rate":0.032},{"Quantity":1.28748993,"Rate":0.03226303},{"Quantity":59.67718649,"Rate":0.0325},{"Quantity":24.67944344,"Rate":0.032595},{"Quantity":908,"Rate":0.0326},{"Quantity":1.7540154,"Rate":0.03267469},{"Quantity":4.51203322,"Rate":0.03269384},{"Quantity":6.25498416,"Rate":0.03274117},{"Quantity":10,"Rate":0.03291292},{"Quantity":30.81809838,"Rate":0.033},{"Quantity":107.15883655,"Rate":0.03309918},{"Quantity":3.11962128,"Rate":0.0333},{"Quantity":1.2811508,"Rate":0.03333048},{"Quantity":69.62641409,"Rate":0.03333333},{"Quantity":1.44823276,"Rate":0.03333649},{"Quantity":22.54313859,"Rate":0.0334},{"Quantity":1.05010045,"Rate":0.03341347},{"Quantity":1.81249624,"Rate":0.03343347},{"Quantity":15.59812353,"Rate":0.03348},{"Quantity":20,"Rate":0.0335},{"Quantity":3.4,"Rate":0.0336},{"Quantity":3,"Rate":0.03381209},{"Quantity":8.88244779,"Rate":0.03395},{"Quantity":1.3507459,"Rate":0.03399217},{"Quantity":0.83013227,"Rate":0.03399842},{"Quantity":17.91766199,"Rate":0.034},{"Quantity":6.76804983,"Rate":0.03400181},{"Quantity":4.55895795,"Rate":0.0342185},{"Quantity":17.19106542,"Rate":0.03424808},{"Quantity":6.14318853,"Rate":0.0345},{"Quantity":11.95048685,"Rate":0.03455765},{"Quantity":1008,"Rate":0.0346},{"Quantity":1.60200893,"Rate":0.03467263},{"Quantity":20,"Rate":0.03475},{"Quantity":819.70770187,"Rate":0.03484999},{"Quantity":2000,"Rate":0.03485},{"Quantity":2,"Rate":0.03489},{"Quantity":37,"Rate":0.0349},{"Quantity":6.63678674,"Rate":0.034936},{"Quantity":700.78228083,"Rate":0.035},{"Quantity":4.5,"Rate":0.035029}],"ReturnTime":"1514114579481"}},"OMG-ETH":{"bittrex":{"Valid":true,"Error":"","Timestamp":"1514114579228","Bids":[{"Quantity":5.49,"Rate":0.019857},{"Quantity":13.62550123,"Rate":0.0197758},{"Quantity":10,"Rate":0.01976677},{"Quantity":6.92629385,"Rate":0.01970274},{"Quantity":12.67779101,"Rate":0.01967042},{"Quantity":1.52592651,"Rate":0.01966018},{"Quantity":51.1,"Rate":0.01965999},{"Quantity":1.1,"Rate":0.01965},{"Quantity":6,"Rate":0.01961897},{"Quantity":611.703,"Rate":0.01961737},{"Quantity":78.6617158,"Rate":0.01961735},{"Quantity":25,"Rate":0.01961281},{"Quantity":17.2578928,"Rate":0.0196128},{"Quantity":27.50180738,"Rate":0.01959872},{"Quantity":394.51673032,"Rate":0.01951755},{"Quantity":3421.5,"Rate":0.01951754},{"Quantity":2.04431606,"Rate":0.01951753},{"Quantity":21.30086709,"Rate":0.01950289},{"Quantity":20,"Rate":0.019501},{"Quantity":416.29946972,"Rate":0.0195},{"Quantity":9.73587162,"Rate":0.01946667},{"Quantity":61.58638195,"Rate":0.01945128},{"Quantity":25,"Rate":0.01943494},{"Quantity":1,"Rate":0.01940005},{"Quantity":1988.8047677,"Rate":0.019389},{"Quantity":5,"Rate":0.0193},{"Quantity":1.11063782,"Rate":0.01929976},{"Quantity":10,"Rate":0.01926677},{"Quantity":3.89040562,"Rate":0.01923},{"Quantity":7.46232877,"Rate":0.01920182},{"Quantity":3.37302184,"Rate":0.01919},{"Quantity":200,"Rate":0.01914717},{"Quantity":159.89102611,"Rate":0.0191},{"Quantity":40,"Rate":0.01908577},{"Quantity":1.1,"Rate":0.019075},{"Quantity":35.54094826,"Rate":0.01901281},{"Quantity":104.92931078,"Rate":0.0190128},{"Quantity":1186.28656415,"Rate":0.019},{"Quantity":52.62702555,"Rate":0.01895414},{"Quantity":5.59525057,"Rate":0.01891877},{"Quantity":1.19341567,"Rate":0.01890012},{"Quantity":528.69978545,"Rate":0.01886704},{"Quantity":72.39360008,"Rate":0.01886304},{"Quantity":10.2050869,"Rate":0.01884},{"Quantity":33.1166288,"Rate":0.01883806},{"Quantity":54.43510232,"Rate":0.0188},{"Quantity":35.91678194,"Rate":0.01875},{"Quantity":5.33923347,"Rate":0.01872928},{"Quantity":18.2323,"Rate":0.018724},{"Quantity":616.65848089,"Rate":0.01870092},{"Quantity":8.0453232,"Rate":0.01870091},{"Quantity":3,"Rate":0.0187},{"Quantity":54.40912,"Rate":0.01869999},{"Quantity":133.763032,"Rate":0.01868991},{"Quantity":8.1730823,"Rate":0.01865},{"Quantity":3,"Rate":0.0186},{"Quantity":200,"Rate":0.018563},{"Quantity":1.1,"Rate":0.01853},{"Quantity":25,"Rate":0.01851},{"Quantity":129.66214805,"Rate":0.0185},{"Quantity":1.83391375,"Rate":0.01845},{"Quantity":271,"Rate":0.018411},{"Quantity":17.72855777,"Rate":0.01840568},{"Quantity":166.63586957,"Rate":0.0184},{"Quantity":1132.78632088,"Rate":0.01830001},{"Quantity":1927.68420627,"Rate":0.0183},{"Quantity":10.90759978,"Rate":0.01829},{"Quantity":20.1123,"Rate":0.01823},{"Quantity":29.60912756,"Rate":0.018205},{"Quantity":70.40036261,"Rate":0.0182},{"Quantity":8.68009182,"Rate":0.01819999},{"Quantity":200,"Rate":0.018136},{"Quantity":25,"Rate":0.0181313},{"Quantity":47.51499464,"Rate":0.01812648},{"Quantity":6.72354779,"Rate":0.0181},{"Quantity":55.20643076,"Rate":0.018069},{"Quantity":3.1188478,"Rate":0.01806063},{"Quantity":8.45740593,"Rate":0.01804061},{"Quantity":100,"Rate":0.01803873},{"Quantity":4.48130113,"Rate":0.01803338},{"Quantity":3.16121232,"Rate":0.01802586},{"Quantity":28.55575281,"Rate":0.01802042},{"Quantity":6.92151152,"Rate":0.01801449},{"Quantity":11.07613385,"Rate":0.0180117},{"Quantity":4,"Rate":0.0180013},{"Quantity":73,"Rate":0.018001},{"Quantity":41.56224601,"Rate":0.01800011},{"Quantity":1333.33262037,"Rate":0.01800001},{"Quantity":2031.49249372,"Rate":0.018},{"Quantity":1.13,"Rate":0.01799},{"Quantity":5.00296309,"Rate":0.01797},{"Quantity":86.85648159,"Rate":0.01791457},{"Quantity":114.11825522,"Rate":0.0179},{"Quantity":2,"Rate":0.01787809},{"Quantity":24.4454,"Rate":0.01782211},{"Quantity":7.89961055,"Rate":0.01781615},{"Quantity":9,"Rate":0.0178},{"Quantity":8.26034527,"Rate":0.01775},{"Quantity":7,"Rate":0.01773439},{"Quantity":100,"Rate":0.01773}],"Asks":[{"Quantity":6.73770653,"Rate":0.02025768},{"Quantity":7.49193537,"Rate":0.02025774},{"Quantity":1.48831433,"Rate":0.02025781},{"Quantity":19.41113466,"Rate":0.02025783},{"Quantity":757.575,"Rate":0.02025799},{"Quantity":708.80530719,"Rate":0.020258},{"Quantity":2459.84715703,"Rate":0.02025804},{"Quantity":11.14180178,"Rate":0.02025805},{"Quantity":174.46513359,"Rate":0.02025836},{"Quantity":131.9204,"Rate":0.02025886},{"Quantity":22.25594893,"Rate":0.020271},{"Quantity":6.17102286,"Rate":0.02045101},{"Quantity":24.9375,"Rate":0.02049992},{"Quantity":52.79828335,"Rate":0.0206049},{"Quantity":16.91636296,"Rate":0.02062813},{"Quantity":34.39153497,"Rate":0.02068932},{"Quantity":24.69059406,"Rate":0.02070491},{"Quantity":339.25133047,"Rate":0.02072033},{"Quantity":1.16044788,"Rate":0.02076312},{"Quantity":40,"Rate":0.02086973},{"Quantity":48.89705882,"Rate":0.02090991},{"Quantity":0.99856458,"Rate":0.02091644},{"Quantity":10.01191895,"Rate":0.02092255},{"Quantity":5,"Rate":0.021},{"Quantity":9.46288021,"Rate":0.02118},{"Quantity":24.3,"Rate":0.02119987},{"Quantity":23.14276071,"Rate":0.02126503},{"Quantity":1,"Rate":0.02131084},{"Quantity":3.72617254,"Rate":0.02131225},{"Quantity":20,"Rate":0.02134213},{"Quantity":24.44852941,"Rate":0.02135},{"Quantity":1.21444743,"Rate":0.02135067},{"Quantity":2.45133,"Rate":0.02146},{"Quantity":83.63929683,"Rate":0.0215},{"Quantity":24.44852941,"Rate":0.02155},{"Quantity":8,"Rate":0.02160478},{"Quantity":2.4329,"Rate":0.02165},{"Quantity":24.44448529,"Rate":0.02174989},{"Quantity":1.86170145,"Rate":0.02176827},{"Quantity":4.77990628,"Rate":0.02178925},{"Quantity":4.5496567,"Rate":0.02181791},{"Quantity":111.34618121,"Rate":0.0219},{"Quantity":9.93403865,"Rate":0.02190603},{"Quantity":950,"Rate":0.02193},{"Quantity":1.07046584,"Rate":0.02193982},{"Quantity":2.40151501,"Rate":0.02199},{"Quantity":11.90881905,"Rate":0.02199999},{"Quantity":277.09567463,"Rate":0.022},{"Quantity":2.48563548,"Rate":0.02200112},{"Quantity":14.4,"Rate":0.02202},{"Quantity":2,"Rate":0.02205},{"Quantity":1,"Rate":0.0220776},{"Quantity":24.05415439,"Rate":0.0220887},{"Quantity":4,"Rate":0.02209962},{"Quantity":4.5496567,"Rate":0.0221114},{"Quantity":40,"Rate":0.0221493},{"Quantity":4.12084998,"Rate":0.02215244},{"Quantity":1.30009461,"Rate":0.02216422},{"Quantity":69.07489173,"Rate":0.022207},{"Quantity":23.0624115,"Rate":0.0222223},{"Quantity":50,"Rate":0.02225},{"Quantity":27.06928683,"Rate":0.02229363},{"Quantity":224.09257313,"Rate":0.02231207},{"Quantity":160.18903562,"Rate":0.0223308},{"Quantity":9.4500005,"Rate":0.02235121},{"Quantity":2.44470086,"Rate":0.02237092},{"Quantity":1.7,"Rate":0.02238},{"Quantity":3.22465339,"Rate":0.02238814},{"Quantity":108.59412253,"Rate":0.02243315},{"Quantity":1.34006507,"Rate":0.02244294},{"Quantity":47.02635503,"Rate":0.02244376},{"Quantity":74.76621416,"Rate":0.0225},{"Quantity":4.64277403,"Rate":0.02255925},{"Quantity":1.19028402,"Rate":0.02256315},{"Quantity":2,"Rate":0.022575},{"Quantity":29,"Rate":0.02259997},{"Quantity":49.05188679,"Rate":0.0226},{"Quantity":4.63441547,"Rate":0.02261524},{"Quantity":566.82973015,"Rate":0.0226822},{"Quantity":24.84077853,"Rate":0.02268622},{"Quantity":10013.5,"Rate":0.0227},{"Quantity":34.44625401,"Rate":0.02270694},{"Quantity":1.95628029,"Rate":0.02276999},{"Quantity":83.3061,"Rate":0.0227788},{"Quantity":23.37886834,"Rate":0.02283593},{"Quantity":711.4234893,"Rate":0.02284211},{"Quantity":57.66668774,"Rate":0.02287076},{"Quantity":1.0324816,"Rate":0.0228776},{"Quantity":50,"Rate":0.0229048},{"Quantity":1.43517884,"Rate":0.02297955},{"Quantity":1.69776641,"Rate":0.02298235},{"Quantity":0.95,"Rate":0.02299643},{"Quantity":10785.12081716,"Rate":0.023},{"Quantity":2.12754132,"Rate":0.02308475},{"Quantity":23.01233937,"Rate":0.0230887},{"Quantity":16.57728812,"Rate":0.02309451},{"Quantity":3.105,"Rate":0.0231},{"Quantity":1,"Rate":0.02315178},{"Quantity":17.4254382,"Rate":0.02320185},{"Quantity":1,"Rate":0.0232518}],"ReturnTime":"1514114579575"}},"PAY-ETH":{"bittrex":{"Valid":true,"Error":"","Timestamp":"1514114579228","Bids":[{"Quantity":17.76916985,"Rate":0.00576079},{"Quantity":25,"Rate":0.0057565},{"Quantity":5.24,"Rate":0.005728},{"Quantity":18.41522176,"Rate":0.00572545},{"Quantity":23.92360761,"Rate":0.00572543},{"Quantity":18.62866423,"Rate":0.00572526},{"Quantity":51,"Rate":0.00572},{"Quantity":75.4018,"Rate":0.0057},{"Quantity":19.08549008,"Rate":0.00569011},{"Quantity":8.2862093,"Rate":0.005676},{"Quantity":19.78091763,"Rate":0.00565476},{"Quantity":20.50248569,"Rate":0.00561942},{"Quantity":157.86857769,"Rate":0.00560566},{"Quantity":21.25121571,"Rate":0.00558408},{"Quantity":25,"Rate":0.00555219},{"Quantity":22.02817098,"Rate":0.00554874},{"Quantity":40.359785,"Rate":0.00554609},{"Quantity":22.83445852,"Rate":0.00551339},{"Quantity":25,"Rate":0.0055},{"Quantity":2541.61406613,"Rate":0.0054821},{"Quantity":15407,"Rate":0.00548209},{"Quantity":145,"Rate":0.0054108},{"Quantity":1559.25925925,"Rate":0.0054},{"Quantity":2000,"Rate":0.00535012},{"Quantity":106,"Rate":0.00535001},{"Quantity":15407,"Rate":0.00527029},{"Quantity":185,"Rate":0.00521821},{"Quantity":192.52184237,"Rate":0.005203},{"Quantity":25.11267931,"Rate":0.0052},{"Quantity":106,"Rate":0.00515001},{"Quantity":12,"Rate":0.0050505},{"Quantity":50,"Rate":0.0050358},{"Quantity":77.39536419,"Rate":0.00503568},{"Quantity":50,"Rate":0.00502},{"Quantity":200,"Rate":0.00501},{"Quantity":23.49278704,"Rate":0.00500598},{"Quantity":200,"Rate":0.005004},{"Quantity":200,"Rate":0.005003},{"Quantity":200,"Rate":0.005002},{"Quantity":200,"Rate":0.005001},{"Quantity":20,"Rate":0.00500054},{"Quantity":45.25622713,"Rate":0.00500001},{"Quantity":6740.17020836,"Rate":0.005},{"Quantity":504,"Rate":0.004985},{"Quantity":39.41909618,"Rate":0.00495},{"Quantity":49,"Rate":0.0049366},{"Quantity":40,"Rate":0.0049},{"Quantity":20.66358172,"Rate":0.00489},{"Quantity":2000,"Rate":0.00485012},{"Quantity":468.75011804,"Rate":0.00485003},{"Quantity":500,"Rate":0.00485},{"Quantity":30.97684992,"Rate":0.00483022},{"Quantity":4.85048874,"Rate":0.00482582},{"Quantity":40,"Rate":0.00482},{"Quantity":565.55667294,"Rate":0.0048},{"Quantity":20.78908922,"Rate":0.00479819},{"Quantity":1224.9125133,"Rate":0.00475},{"Quantity":300,"Rate":0.00474},{"Quantity":6.35447865,"Rate":0.00472572},{"Quantity":15,"Rate":0.004721},{"Quantity":78.33077094,"Rate":0.0047},{"Quantity":1013.82111699,"Rate":0.00467311},{"Quantity":1000,"Rate":0.00465},{"Quantity":94.93213431,"Rate":0.00464651},{"Quantity":282.85598287,"Rate":0.00463001},{"Quantity":16.26208025,"Rate":0.004617},{"Quantity":180.1567739,"Rate":0.004612},{"Quantity":53,"Rate":0.0046042},{"Quantity":40,"Rate":0.00460056},{"Quantity":41.18671542,"Rate":0.0046005},{"Quantity":121.88463432,"Rate":0.0046},{"Quantity":22.84487265,"Rate":0.00458043},{"Quantity":44.22405079,"Rate":0.00451112},{"Quantity":7,"Rate":0.00450505},{"Quantity":22.40648048,"Rate":0.004505},{"Quantity":20,"Rate":0.00450482},{"Quantity":241.27697209,"Rate":0.0045},{"Quantity":71.21626321,"Rate":0.00446697},{"Quantity":268,"Rate":0.0044},{"Quantity":35,"Rate":0.00435},{"Quantity":20,"Rate":0.00430458},{"Quantity":20,"Rate":0.00430045},{"Quantity":100,"Rate":0.0043},{"Quantity":14.62933182,"Rate":0.00429445},{"Quantity":60,"Rate":0.00429248},{"Quantity":100,"Rate":0.00429011},{"Quantity":25,"Rate":0.0042664},{"Quantity":721.24182169,"Rate":0.00425},{"Quantity":5.00073972,"Rate":0.004247},{"Quantity":11.82905595,"Rate":0.00422688},{"Quantity":473.14571939,"Rate":0.00421646},{"Quantity":55.55806768,"Rate":0.00420187},{"Quantity":1819.37083714,"Rate":0.00420098},{"Quantity":103.49973035,"Rate":0.0042005},{"Quantity":6.0318255,"Rate":0.00420046},{"Quantity":678.73597075,"Rate":0.0042},{"Quantity":102.50468333,"Rate":0.00417591},{"Quantity":2109.48045442,"Rate":0.00417217},{"Quantity":900,"Rate":0.00417192},{"Quantity":8.10626999,"Rate":0.0041555}],"Asks":[{"Quantity":136.4072,"Rate":0.00581225},{"Quantity":776.223,"Rate":0.00583147},{"Quantity":15.90915084,"Rate":0.00583148},{"Quantity":186.3434287,"Rate":0.00589959},{"Quantity":5.18,"Rate":0.005902},{"Quantity":6372,"Rate":0.00596099},{"Quantity":5.13,"Rate":0.005961},{"Quantity":5.08,"Rate":0.006021},{"Quantity":191.65006929,"Rate":0.0060214},{"Quantity":185.523922,"Rate":0.00608199},{"Quantity":5.03,"Rate":0.006082},{"Quantity":4.98,"Rate":0.006143},{"Quantity":500,"Rate":0.006205},{"Quantity":4.93,"Rate":0.006206},{"Quantity":500,"Rate":0.006267},{"Quantity":4.88,"Rate":0.006268},{"Quantity":4.83,"Rate":0.006331},{"Quantity":317.31539406,"Rate":0.00634274},{"Quantity":4.78,"Rate":0.006395},{"Quantity":144.22241399,"Rate":0.0064},{"Quantity":6.30036102,"Rate":0.00640282},{"Quantity":4.74,"Rate":0.00646},{"Quantity":4.69,"Rate":0.006525},{"Quantity":504.64,"Rate":0.006591},{"Quantity":16.69984528,"Rate":0.00660266},{"Quantity":4.6,"Rate":0.006658},{"Quantity":4.55,"Rate":0.006722},{"Quantity":4.44,"Rate":0.00679},{"Quantity":26,"Rate":0.0069},{"Quantity":50,"Rate":0.00698812},{"Quantity":5.78051693,"Rate":0.00699734},{"Quantity":500,"Rate":0.00699736},{"Quantity":561,"Rate":0.00699999},{"Quantity":600.81807963,"Rate":0.007},{"Quantity":1210.10389606,"Rate":0.00702999},{"Quantity":1999.99985845,"Rate":0.00703},{"Quantity":10.6,"Rate":0.00704071},{"Quantity":114.92813833,"Rate":0.00706},{"Quantity":500,"Rate":0.00708278},{"Quantity":498,"Rate":0.0071},{"Quantity":37.58240958,"Rate":0.00717998},{"Quantity":88.3367274,"Rate":0.00718},{"Quantity":1000,"Rate":0.00734075},{"Quantity":1030.88356968,"Rate":0.00743},{"Quantity":11,"Rate":0.00745},{"Quantity":30,"Rate":0.00748168},{"Quantity":16.93660883,"Rate":0.00749257},{"Quantity":1289.25209761,"Rate":0.0075},{"Quantity":10,"Rate":0.00757575},{"Quantity":20,"Rate":0.0076},{"Quantity":917.66695816,"Rate":0.00760565},{"Quantity":5.24798195,"Rate":0.00768958},{"Quantity":16.30929715,"Rate":0.00775504},{"Quantity":901,"Rate":0.00785},{"Quantity":21.80124288,"Rate":0.00788202},{"Quantity":150,"Rate":0.007889},{"Quantity":2000,"Rate":0.00792},{"Quantity":231.90606704,"Rate":0.00793148},{"Quantity":92.00785179,"Rate":0.00795},{"Quantity":900.07735747,"Rate":0.00796192},{"Quantity":134,"Rate":0.0079999},{"Quantity":2570.70237631,"Rate":0.008},{"Quantity":56.68482876,"Rate":0.0080084},{"Quantity":463,"Rate":0.0081},{"Quantity":16.30314581,"Rate":0.00813399},{"Quantity":1000,"Rate":0.0082},{"Quantity":177.38924124,"Rate":0.00824992},{"Quantity":300,"Rate":0.00825},{"Quantity":490,"Rate":0.00830022},{"Quantity":1000,"Rate":0.00831399},{"Quantity":1572.68316387,"Rate":0.00833399},{"Quantity":10,"Rate":0.00834115},{"Quantity":144.9,"Rate":0.00835},{"Quantity":1000,"Rate":0.00835399},{"Quantity":18.13427623,"Rate":0.00838279},{"Quantity":1000,"Rate":0.00838399},{"Quantity":2061.02679052,"Rate":0.0084},{"Quantity":18.0372003,"Rate":0.0084339},{"Quantity":628.03094059,"Rate":0.00843399},{"Quantity":50,"Rate":0.008434},{"Quantity":200,"Rate":0.0084488},{"Quantity":90,"Rate":0.00845998},{"Quantity":5974.1181438,"Rate":0.00846},{"Quantity":100,"Rate":0.008475},{"Quantity":579.37974684,"Rate":0.00849},{"Quantity":66.95866301,"Rate":0.0084948},{"Quantity":2480.17533117,"Rate":0.0085},{"Quantity":20,"Rate":0.00854},{"Quantity":55.96558831,"Rate":0.00854206},{"Quantity":5.27631579,"Rate":0.00855},{"Quantity":10,"Rate":0.00857},{"Quantity":800,"Rate":0.0086},{"Quantity":500,"Rate":0.008615},{"Quantity":6,"Rate":0.008625},{"Quantity":13.85,"Rate":0.00866709},{"Quantity":21.9,"Rate":0.00867017},{"Quantity":40,"Rate":0.0087},{"Quantity":12.69,"Rate":0.00874351},{"Quantity":13.15,"Rate":0.00874902},{"Quantity":1733.93883196,"Rate":0.0088}],"ReturnTime":"1514114579574"}}},"success":true,"timestamp":"1514114582015","version":64}
```

### Get precision and limit info when trading for base-quote pair on an exchange

```
<host>:8000/exchangeinfo/<exchangeid>/<base>/<quote>
```

Where *<exchangeid>* is the id of the exchange, *<base>* is symbol of the base token and *<quote>* is symbol of the quote token

eg:
```
curl -X GET "http://13.229.54.28:8000/exchangeinfo/binance/omg/eth"
```
response:
```
  {"data":{"Precision":{"Amount":8,"Price":8},"AmountLimit":{"Min":0.01,"Max":90000000},"PriceLimit":{"Min":0.000001,"Max":100000}},"success":true}
```

### Get precision and limit info when trading for all base-quote pairs of an exchange

```
<host>:8000/exchangeinfo
```
url params:

*exchangeid* : id of exchange to get info (optional, if exchangeid is empty then return all exchanges info)

eg:
```
curl -X GET "http://13.229.54.28:8000/exchangeinfo?exchangeid=binance"
```
response:
```
  {"data":{"binance":{"EOS-ETH":{"Precision":{"Amount":8,"Price":8},"AmountLimit":{"Min":0.01,"Max":90000000},"PriceLimit":{"Min":0.000001,"Max":100000},"MinNotional":0.02},"KNC-ETH":{"Precision":{"Amount":8,"Price":8},"AmountLimit":{"Min":1,"Max":90000000},"PriceLimit":{"Min":1e-7,"Max":100000},"MinNotional":0.02},"OMG-ETH":{"Precision":{"Amount":8,"Price":8},"AmountLimit":{"Min":0.01,"Max":90000000},"PriceLimit":{"Min":0.000001,"Max":100000},"MinNotional":0.02},"SALT-ETH":{"Precision":{"Amount":8,"Price":8},"AmountLimit":{"Min":0.01,"Max":90000000},"PriceLimit":{"Min":0.000001,"Max":100000},"MinNotional":0.02},"SNT-ETH":{"Precision":{"Amount":8,"Price":8},"AmountLimit":{"Min":1,"Max":90000000},"PriceLimit":{"Min":1e-8,"Max":100000},"MinNotional":0.02}}},"success":true}
```

### Get fee for transaction on all exchanges

```
<host>:8000/exchangefees
```

eg:
```
curl -X GET "http://13.229.54.28:8000/exchangefees"
```
response:
```
  {"data":[{"binance":{"Trading":{"maker":0.001,"taker":0.001},"Funding":{"Withdraw":{"EOS":2,"ETH":0.005,"FUN":50,"KNC":1,"LINK":5,"MCO":0.15,"OMG":0.1},"Deposit":{"EOS":0,"ETH":0,"FUN":0,"KNC":0,"LINK":0,"MCO":0,"OMG":0}}}},{"bittrex":{"Trading":{"maker":0.0025,"taker":0.0025},"Funding":{"Withdraw":{"BTC":0.001,"DASH":0.002,"DOGE":2,"FTC":0.2,"LTC":0.01,"NXT":2,"POT":0.002,"PPC":0.02,"RDD":2,"VTC":0.02},"Deposit":{"BTC":0,"DASH":0,"DOGE":0,"FTC":0,"LTC":0,"NXT":0,"POT":0,"PPC":0,"RDD":0,"VTC":0}}}}],"success":true}
```

### Get fee for transaction on an exchange

```
<host>:8000/exchangefees/<exchangeid>
```

Where *<exchangeid>* is the id of the exchange

eg:
```
curl -X GET "http://13.229.54.28:8000/exchangefees/binance"
```
response:
```
  {"data":{"Trading":{"maker":0.001,"taker":0.001},"Funding":{"Withdraw":{"EOS":2,"ETH":0.005,"FUN":50,"KNC":1,"LINK":5,"MCO":0.15,"OMG":0.1},"Deposit":{"EOS":0,"ETH":0,"FUN":0,"KNC":0,"LINK":0,"MCO":0,"OMG":0}}},"success":true}
```

### Get token rates from blockchain

```
<host>:8000/getrates
```

eg:
```
curl -X GET "http://13.229.54.28:8000/getrates"
```
response:
```
  {"data":{"ADX":{"Valid":true,"Error":"","Timestamp":"1515412582435","ReturnTime":"1515412582710","BaseBuy":371.0142432353458,"CompactBuy":0,"BaseSell":0.002538305711940429,"CompactSell":0,"Rate":0,"Block":2420849},"BAT":{"Valid":true,"Error":"","Timestamp":"1515412582435","ReturnTime":"1515412582710","BaseBuy":1656.6398539506304,"CompactBuy":0,"BaseSell":0.0005684685,"CompactSell":0,"Rate":0,"Block":2420849},"CVC":{"Valid":true,"Error":"","Timestamp":"1515412582435","ReturnTime":"1515412582710","BaseBuy":1051.2127184124374,"CompactBuy":-1,"BaseSell":0.00089586775,"CompactSell":1,"Rate":0,"Block":2420849},"DGD":{"Valid":true,"Error":"","Timestamp":"1515412582435","ReturnTime":"1515412582710","BaseBuy":5.662106994812361,"CompactBuy":0,"BaseSell":0.16632458088099816,"CompactSell":0,"Rate":0,"Block":2420849},"EOS":{"Valid":true,"Error":"","Timestamp":"1515412582435","ReturnTime":"1515412582710","BaseBuy":121.11698932232625,"CompactBuy":-15,"BaseSell":0.007775519999999998,"CompactSell":15,"Rate":0,"Block":2420849},"ETH":{"Valid":true,"Error":"","Timestamp":"1515412582435","ReturnTime":"1515412582710","BaseBuy":0,"CompactBuy":30,"BaseSell":0,"CompactSell":-29,"Rate":0,"Block":2420849},"FUN":{"Valid":true,"Error":"","Timestamp":"1515412582435","ReturnTime":"1515412582710","BaseBuy":6805.131583093689,"CompactBuy":33,"BaseSell":0.000138387856475128,"CompactSell":-32,"Rate":0,"Block":2420849},"GNT":{"Valid":true,"Error":"","Timestamp":"1515412582435","ReturnTime":"1515412582710","BaseBuy":1055.0281030473377,"CompactBuy":-74,"BaseSell":0.0010113802,"CompactSell":-47,"Rate":0,"Block":2420849},"KNC":{"Valid":true,"Error":"","Timestamp":"1515412582435","ReturnTime":"1515412582710","BaseBuy":229.65128829779712,"CompactBuy":89,"BaseSell":0.004100772,"CompactSell":-82,"Rate":0,"Block":2420849},"LINK":{"Valid":true,"Error":"","Timestamp":"1515412582435","ReturnTime":"1515412582710","BaseBuy":844.2527577938458,"CompactBuy":101,"BaseSell":0.0011154806,"CompactSell":-91,"Rate":0,"Block":2420849},"MCO":{"Valid":true,"Error":"","Timestamp":"1515412582435","ReturnTime":"1515412582710","BaseBuy":63.99319226272073,"CompactBuy":21,"BaseSell":0.014716371218820246,"CompactSell":-20,"Rate":0,"Block":2420849},"OMG":{"Valid":true,"Error":"","Timestamp":"1515412582435","ReturnTime":"1515412582710","BaseBuy":44.45707162223901,"CompactBuy":30,"BaseSell":0.021183301968644246,"CompactSell":-29,"Rate":0,"Block":2420849},"PAY":{"Valid":true,"Error":"","Timestamp":"1515412582435","ReturnTime":"1515412582710","BaseBuy":295.08854913901575,"CompactBuy":-13,"BaseSell":0.003191406699999999,"CompactSell":13,"Rate":0,"Block":2420849}},"success":true,"timestamp":"1515412583215","version":1515412582435}
```


### Get all token rates from blockchain
```
<host>:8000/get-all-rates
```
url params:
*fromTime*: optional, get all rates from this timepoint (millisecond)
*toTime*: optional, get all rates to this timepoint (millisecond)

eg:
```
curl -X GET "http://13.229.54.28:8000/get-all-rates"
```
 response
```
{"data":[{"Version":0,"Valid":true,"Error":"","Timestamp":"1517280618739","ReturnTime":"1517280619071","Data":{"EOS":{"Valid":true,"Error":"","Timestamp":"1517280618739","ReturnTime":"1517280619071","BaseBuy":87.21360760013062,"CompactBuy":0,"BaseSell":0.0128686459657361,"CompactSell":0,"Rate":0,"Block":5635245},"ETH":{"Valid":true,"Error":"","Timestamp":"1517280618739","ReturnTime":"1517280619071","BaseBuy":0,"CompactBuy":32,"BaseSell":0,"CompactSell":-14,"Rate":0,"Block":5635245},"KNC":{"Valid":true,"Error":"","Timestamp":"1517280618739","ReturnTime":"1517280619071","BaseBuy":307.05930436561505,"CompactBuy":-34,"BaseSell":0.003084981280661941,"CompactSell":81,"Rate":0,"Block":5635245},"OMG":{"Valid":true,"Error":"","Timestamp":"1517280618739","ReturnTime":"1517280619071","BaseBuy":65.0580993582104,"CompactBuy":32,"BaseSell":0.014925950060437398,"CompactSell":-14,"Rate":0,"Block":5635245},"SALT":{"Valid":true,"Error":"","Timestamp":"1517280618739","ReturnTime":"1517280619071","BaseBuy":152.3016783627643,"CompactBuy":9,"BaseSell":0.006196212698403499,"CompactSell":23,"Rate":0,"Block":5635245},"SNT":{"Valid":true,"Error":"","Timestamp":"1517280618739","ReturnTime":"1517280619071","BaseBuy":4053.2170631085987,"CompactBuy":43,"BaseSell":0.000233599514875301,"CompactSell":-3,"Rate":0,"Block":5635245}}},{"Version":0,"Valid":true,"Error":"","Timestamp":"1517280621738","ReturnTime":"1517280622251","Data":{"EOS":{"Valid":true,"Error":"","Timestamp":"1517280621738","ReturnTime":"1517280622251","BaseBuy":87.21360760013062,"CompactBuy":0,"BaseSell":0.0128686459657361,"CompactSell":0,"Rate":0,"Block":5635245},"ETH":{"Valid":true,"Error":"","Timestamp":"1517280621738","ReturnTime":"1517280622251","BaseBuy":0,"CompactBuy":32,"BaseSell":0,"CompactSell":-14,"Rate":0,"Block":5635245},"KNC":{"Valid":true,"Error":"","Timestamp":"1517280621738","ReturnTime":"1517280622251","BaseBuy":307.05930436561505,"CompactBuy":-34,"BaseSell":0.003084981280661941,"CompactSell":81,"Rate":0,"Block":5635245},"OMG":{"Valid":true,"Error":"","Timestamp":"1517280621738","ReturnTime":"1517280622251","BaseBuy":65.0580993582104,"CompactBuy":32,"BaseSell":0.014925950060437398,"CompactSell":-14,"Rate":0,"Block":5635245},"SALT":{"Valid":true,"Error":"","Timestamp":"1517280621738","ReturnTime":"1517280622251","BaseBuy":152.3016783627643,"CompactBuy":9,"BaseSell":0.006196212698403499,"CompactSell":23,"Rate":0,"Block":5635245},"SNT":{"Valid":true,"Error":"","Timestamp":"1517280621738","ReturnTime":"1517280622251","BaseBuy":4053.2170631085987,"CompactBuy":43,"BaseSell":0.000233599514875301,"CompactSell":-3,"Rate":0,"Block":5635245}}},{"Version":0,"Valid":true,"Error":"","Timestamp":"1517280624739","ReturnTime":"1517280625052","Data":{"EOS":{"Valid":true,"Error":"","Timestamp":"1517280624739","ReturnTime":"1517280625052","BaseBuy":87.21360760013062,"CompactBuy":0,"BaseSell":0.0128686459657361,"CompactSell":0,"Rate":0,"Block":5635245},"ETH":{"Valid":true,"Error":"","Timestamp":"1517280624739","ReturnTime":"1517280625052","BaseBuy":0,"CompactBuy":32,"BaseSell":0,"CompactSell":-14,"Rate":0,"Block":5635245},"KNC":{"Valid":true,"Error":"","Timestamp":"1517280624739","ReturnTime":"1517280625052","BaseBuy":307.05930436561505,"CompactBuy":-34,"BaseSell":0.003084981280661941,"CompactSell":81,"Rate":0,"Block":5635245},"OMG":{"Valid":true,"Error":"","Timestamp":"1517280624739","ReturnTime":"1517280625052","BaseBuy":65.0580993582104,"CompactBuy":32,"BaseSell":0.014925950060437398,"CompactSell":-14,"Rate":0,"Block":5635245},"SALT":{"Valid":true,"Error":"","Timestamp":"1517280624739","ReturnTime":"1517280625052","BaseBuy":152.3016783627643,"CompactBuy":9,"BaseSell":0.006196212698403499,"CompactSell":23,"Rate":0,"Block":5635245},"SNT":{"Valid":true,"Error":"","Timestamp":"1517280624739","ReturnTime":"1517280625052","BaseBuy":4053.2170631085987,"CompactBuy":43,"BaseSell":0.000233599514875301,"CompactSell":-3,"Rate":0,"Block":5635245}}},{"Version":0,"Valid":true,"Error":"","Timestamp":"1517280627735","ReturnTime":"1517280628664","Data":{"EOS":{"Valid":true,"Error":"","Timestamp":"1517280627735","ReturnTime":"1517280628664","BaseBuy":87.21360760013062,"CompactBuy":0,"BaseSell":0.0128686459657361,"CompactSell":0,"Rate":0,"Block":5635245},"ETH":{"Valid":true,"Error":"","Timestamp":"1517280627735","ReturnTime":"1517280628664","BaseBuy":0,"CompactBuy":32,"BaseSell":0,"CompactSell":-14,"Rate":0,"Block":5635245},"KNC":{"Valid":true,"Error":"","Timestamp":"1517280627735","ReturnTime":"1517280628664","BaseBuy":307.05930436561505,"CompactBuy":-34,"BaseSell":0.003084981280661941,"CompactSell":81,"Rate":0,"Block":5635245},"OMG":{"Valid":true,"Error":"","Timestamp":"1517280627735","ReturnTime":"1517280628664","BaseBuy":65.0580993582104,"CompactBuy":32,"BaseSell":0.014925950060437398,"CompactSell":-14,"Rate":0,"Block":5635245},"SALT":{"Valid":true,"Error":"","Timestamp":"1517280627735","ReturnTime":"1517280628664","BaseBuy":152.3016783627643,"CompactBuy":9,"BaseSell":0.006196212698403499,"CompactSell":23,"Rate":0,"Block":5635245},"SNT":{"Valid":true,"Error":"","Timestamp":"1517280627735","ReturnTime":"1517280628664","BaseBuy":4053.2170631085987,"CompactBuy":43,"BaseSell":0.000233599514875301,"CompactSell":-3,"Rate":0,"Block":5635245}}},{"Version":0,"Valid":true,"Error":"","Timestamp":"1517280630737","ReturnTime":"1517280631266","Data":{"EOS":{"Valid":true,"Error":"","Timestamp":"1517280630737","ReturnTime":"1517280631266","BaseBuy":87.21360760013062,"CompactBuy":0,"BaseSell":0.0128686459657361,"CompactSell":0,"Rate":0,"Block":5635245},"ETH":{"Valid":true,"Error":"","Timestamp":"1517280630737","ReturnTime":"1517280631266","BaseBuy":0,"CompactBuy":32,"BaseSell":0,"CompactSell":-14,"Rate":0,"Block":5635245},"KNC":{"Valid":true,"Error":"","Timestamp":"1517280630737","ReturnTime":"1517280631266","BaseBuy":307.05930436561505,"CompactBuy":-34,"BaseSell":0.003084981280661941,"CompactSell":81,"Rate":0,"Block":5635245},"OMG":{"Valid":true,"Error":"","Timestamp":"1517280630737","ReturnTime":"1517280631266","BaseBuy":65.0580993582104,"CompactBuy":32,"BaseSell":0.014925950060437398,"CompactSell":-14,"Rate":0,"Block":5635245},"SALT":{"Valid":true,"Error":"","Timestamp":"1517280630737","ReturnTime":"1517280631266","BaseBuy":152.3016783627643,"CompactBuy":9,"BaseSell":0.006196212698403499,"CompactSell":23,"Rate":0,"Block":5635245},"SNT":{"Valid":true,"Error":"","Timestamp":"1517280630737","ReturnTime":"1517280631266","BaseBuy":4053.2170631085987,"CompactBuy":43,"BaseSell":0.000233599514875301,"CompactSell":-3,"Rate":0,"Block":5635245}}},{"Version":0,"Valid":true,"Error":"","Timestamp":"1517280633737","ReturnTime":"1517280634096","Data":{"EOS":{"Valid":true,"Error":"","Timestamp":"1517280633737","ReturnTime":"1517280634096","BaseBuy":87.21360760013062,"CompactBuy":0,"BaseSell":0.0128686459657361,"CompactSell":0,"Rate":0,"Block":5635245},"ETH":{"Valid":true,"Error":"","Timestamp":"1517280633737","ReturnTime":"1517280634096","BaseBuy":0,"CompactBuy":32,"BaseSell":0,"CompactSell":-14,"Rate":0,"Block":5635245},"KNC":{"Valid":true,"Error":"","Timestamp":"1517280633737","ReturnTime":"1517280634096","BaseBuy":307.05930436561505,"CompactBuy":-34,"BaseSell":0.003084981280661941,"CompactSell":81,"Rate":0,"Block":5635245},"OMG":{"Valid":true,"Error":"","Timestamp":"1517280633737","ReturnTime":"1517280634096","BaseBuy":65.0580993582104,"CompactBuy":32,"BaseSell":0.014925950060437398,"CompactSell":-14,"Rate":0,"Block":5635245},"SALT":{"Valid":true,"Error":"","Timestamp":"1517280633737","ReturnTime":"1517280634096","BaseBuy":152.3016783627643,"CompactBuy":9,"BaseSell":0.006196212698403499,"CompactSell":23,"Rate":0,"Block":5635245},"SNT":{"Valid":true,"Error":"","Timestamp":"1517280633737","ReturnTime":"1517280634096","BaseBuy":4053.2170631085987,"CompactBuy":43,"BaseSell":0.000233599514875301,"CompactSell":-3,"Rate":0,"Block":5635245}}},{"Version":0,"Valid":true,"Error":"","Timestamp":"1517280636736","ReturnTime":"1517280637187","Data":{"EOS":{"Valid":true,"Error":"","Timestamp":"1517280636736","ReturnTime":"1517280637187","BaseBuy":87.21360760013062,"CompactBuy":0,"BaseSell":0.0128686459657361,"CompactSell":0,"Rate":0,"Block":5635245},"ETH":{"Valid":true,"Error":"","Timestamp":"1517280636736","ReturnTime":"1517280637187","BaseBuy":0,"CompactBuy":32,"BaseSell":0,"CompactSell":-14,"Rate":0,"Block":5635245},"KNC":{"Valid":true,"Error":"","Timestamp":"1517280636736","ReturnTime":"1517280637187","BaseBuy":307.05930436561505,"CompactBuy":-34,"BaseSell":0.003084981280661941,"CompactSell":81,"Rate":0,"Block":5635245},"OMG":{"Valid":true,"Error":"","Timestamp":"1517280636736","ReturnTime":"1517280637187","BaseBuy":65.0580993582104,"CompactBuy":32,"BaseSell":0.014925950060437398,"CompactSell":-14,"Rate":0,"Block":5635245},"SALT":{"Valid":true,"Error":"","Timestamp":"1517280636736","ReturnTime":"1517280637187","BaseBuy":152.3016783627643,"CompactBuy":9,"BaseSell":0.006196212698403499,"CompactSell":23,"Rate":0,"Block":5635245},"SNT":{"Valid":true,"Error":"","Timestamp":"1517280636736","ReturnTime":"1517280637187","BaseBuy":4053.2170631085987,"CompactBuy":43,"BaseSell":0.000233599514875301,"CompactSell":-3,"Rate":0,"Block":5635245}}},{"Version":0,"Valid":true,"Error":"","Timestamp":"1517280639741","ReturnTime":"1517280640213","Data":{"EOS":{"Valid":true,"Error":"","Timestamp":"1517280639741","ReturnTime":"1517280640213","BaseBuy":87.21360760013062,"CompactBuy":0,"BaseSell":0.0128686459657361,"CompactSell":0,"Rate":0,"Block":5635245},"ETH":{"Valid":true,"Error":"","Timestamp":"1517280639741","ReturnTime":"1517280640213","BaseBuy":0,"CompactBuy":32,"BaseSell":0,"CompactSell":-14,"Rate":0,"Block":5635245},"KNC":{"Valid":true,"Error":"","Timestamp":"1517280639741","ReturnTime":"1517280640213","BaseBuy":307.05930436561505,"CompactBuy":-34,"BaseSell":0.003084981280661941,"CompactSell":81,"Rate":0,"Block":5635245},"OMG":{"Valid":true,"Error":"","Timestamp":"1517280639741","ReturnTime":"1517280640213","BaseBuy":65.0580993582104,"CompactBuy":32,"BaseSell":0.014925950060437398,"CompactSell":-14,"Rate":0,"Block":5635245},"SALT":{"Valid":true,"Error":"","Timestamp":"1517280639741","ReturnTime":"1517280640213","BaseBuy":152.3016783627643,"CompactBuy":9,"BaseSell":0.006196212698403499,"CompactSell":23,"Rate":0,"Block":5635245},"SNT":{"Valid":true,"Error":"","Timestamp":"1517280639741","ReturnTime":"1517280640213","BaseBuy":4053.2170631085987,"CompactBuy":43,"BaseSell":0.000233599514875301,"CompactSell":-3,"Rate":0,"Block":5635245}}},{"Version":0,"Valid":true,"Error":"","Timestamp":"1517280642741","ReturnTime":"1517280643093","Data":{"EOS":{"Valid":true,"Error":"","Timestamp":"1517280642741","ReturnTime":"1517280643093","BaseBuy":87.21360760013062,"CompactBuy":0,"BaseSell":0.0128686459657361,"CompactSell":0,"Rate":0,"Block":5635245},"ETH":{"Valid":true,"Error":"","Timestamp":"1517280642741","ReturnTime":"1517280643093","BaseBuy":0,"CompactBuy":32,"BaseSell":0,"CompactSell":-14,"Rate":0,"Block":5635245},"KNC":{"Valid":true,"Error":"","Timestamp":"1517280642741","ReturnTime":"1517280643093","BaseBuy":307.05930436561505,"CompactBuy":-34,"BaseSell":0.003084981280661941,"CompactSell":81,"Rate":0,"Block":5635245},"OMG":{"Valid":true,"Error":"","Timestamp":"1517280642741","ReturnTime":"1517280643093","BaseBuy":65.0580993582104,"CompactBuy":32,"BaseSell":0.014925950060437398,"CompactSell":-14,"Rate":0,"Block":5635245},"SALT":{"Valid":true,"Error":"","Timestamp":"1517280642741","ReturnTime":"1517280643093","BaseBuy":152.3016783627643,"CompactBuy":9,"BaseSell":0.006196212698403499,"CompactSell":23,"Rate":0,"Block":5635245},"SNT":{"Valid":true,"Error":"","Timestamp":"1517280642741","ReturnTime":"1517280643093","BaseBuy":4053.2170631085987,"CompactBuy":43,"BaseSell":0.000233599514875301,"CompactSell":-3,"Rate":0,"Block":5635245}}},{"Version":0,"Valid":true,"Error":"","Timestamp":"1517280645737","ReturnTime":"1517280646071","Data":{"EOS":{"Valid":true,"Error":"","Timestamp":"1517280645737","ReturnTime":"1517280646071","BaseBuy":87.21360760013062,"CompactBuy":0,"BaseSell":0.0128686459657361,"CompactSell":0,"Rate":0,"Block":5635245},"ETH":{"Valid":true,"Error":"","Timestamp":"1517280645737","ReturnTime":"1517280646071","BaseBuy":0,"CompactBuy":32,"BaseSell":0,"CompactSell":-14,"Rate":0,"Block":5635245},"KNC":{"Valid":true,"Error":"","Timestamp":"1517280645737","ReturnTime":"1517280646071","BaseBuy":307.05930436561505,"CompactBuy":-34,"BaseSell":0.003084981280661941,"CompactSell":81,"Rate":0,"Block":5635245},"OMG":{"Valid":true,"Error":"","Timestamp":"1517280645737","ReturnTime":"1517280646071","BaseBuy":65.0580993582104,"CompactBuy":32,"BaseSell":0.014925950060437398,"CompactSell":-14,"Rate":0,"Block":5635245},"SALT":{"Valid":true,"Error":"","Timestamp":"1517280645737","ReturnTime":"1517280646071","BaseBuy":152.3016783627643,"CompactBuy":9,"BaseSell":0.006196212698403499,"CompactSell":23,"Rate":0,"Block":5635245},"SNT":{"Valid":true,"Error":"","Timestamp":"1517280645737","ReturnTime":"1517280646071","BaseBuy":4053.2170631085987,"CompactBuy":43,"BaseSell":0.000233599514875301,"CompactSell":-3,"Rate":0,"Block":5635245}}},{"Version":0,"Valid":true,"Error":"","Timestamp":"1517280648738","ReturnTime":"1517280649073","Data":{"EOS":{"Valid":true,"Error":"","Timestamp":"1517280648738","ReturnTime":"1517280649073","BaseBuy":87.21360760013062,"CompactBuy":0,"BaseSell":0.0128686459657361,"CompactSell":0,"Rate":0,"Block":5635245},"ETH":{"Valid":true,"Error":"","Timestamp":"1517280648738","ReturnTime":"1517280649073","BaseBuy":0,"CompactBuy":32,"BaseSell":0,"CompactSell":-14,"Rate":0,"Block":5635245},"KNC":{"Valid":true,"Error":"","Timestamp":"1517280648738","ReturnTime":"1517280649073","BaseBuy":307.05930436561505,"CompactBuy":-34,"BaseSell":0.003084981280661941,"CompactSell":81,"Rate":0,"Block":5635245},"OMG":{"Valid":true,"Error":"","Timestamp":"1517280648738","ReturnTime":"1517280649073","BaseBuy":65.0580993582104,"CompactBuy":32,"BaseSell":0.014925950060437398,"CompactSell":-14,"Rate":0,"Block":5635245},"SALT":{"Valid":true,"Error":"","Timestamp":"1517280648738","ReturnTime":"1517280649073","BaseBuy":152.3016783627643,"CompactBuy":9,"BaseSell":0.006196212698403499,"CompactSell":23,"Rate":0,"Block":5635245},"SNT":{"Valid":true,"Error":"","Timestamp":"1517280648738","ReturnTime":"1517280649073","BaseBuy":4053.2170631085987,"CompactBuy":43,"BaseSell":0.000233599514875301,"CompactSell":-3,"Rate":0,"Block":5635245}}},{"Version":0,"Valid":true,"Error":"","Timestamp":"1517280651741","ReturnTime":"1517280652069","Data":{"EOS":{"Valid":true,"Error":"","Timestamp":"1517280651741","ReturnTime":"1517280652069","BaseBuy":87.21360760013062,"CompactBuy":0,"BaseSell":0.0128686459657361,"CompactSell":0,"Rate":0,"Block":5635245},"ETH":{"Valid":true,"Error":"","Timestamp":"1517280651741","ReturnTime":"1517280652069","BaseBuy":0,"CompactBuy":32,"BaseSell":0,"CompactSell":-14,"Rate":0,"Block":5635245},"KNC":{"Valid":true,"Error":"","Timestamp":"1517280651741","ReturnTime":"1517280652069","BaseBuy":307.05930436561505,"CompactBuy":-34,"BaseSell":0.003084981280661941,"CompactSell":81,"Rate":0,"Block":5635245},"OMG":{"Valid":true,"Error":"","Timestamp":"1517280651741","ReturnTime":"1517280652069","BaseBuy":65.0580993582104,"CompactBuy":32,"BaseSell":0.014925950060437398,"CompactSell":-14,"Rate":0,"Block":5635245},"SALT":{"Valid":true,"Error":"","Timestamp":"1517280651741","ReturnTime":"1517280652069","BaseBuy":152.3016783627643,"CompactBuy":9,"BaseSell":0.006196212698403499,"CompactSell":23,"Rate":0,"Block":5635245},"SNT":{"Valid":true,"Error":"","Timestamp":"1517280651741","ReturnTime":"1517280652069","BaseBuy":4053.2170631085987,"CompactBuy":43,"BaseSell":0.000233599514875301,"CompactSell":-3,"Rate":0,"Block":5635245}}},{"Version":0,"Valid":true,"Error":"","Timestamp":"1517280654737","ReturnTime":"1517280655067","Data":{"EOS":{"Valid":true,"Error":"","Timestamp":"1517280654737","ReturnTime":"1517280655067","BaseBuy":87.21360760013062,"CompactBuy":0,"BaseSell":0.0128686459657361,"CompactSell":0,"Rate":0,"Block":5635245},"ETH":{"Valid":true,"Error":"","Timestamp":"1517280654737","ReturnTime":"1517280655067","BaseBuy":0,"CompactBuy":32,"BaseSell":0,"CompactSell":-14,"Rate":0,"Block":5635245},"KNC":{"Valid":true,"Error":"","Timestamp":"1517280654737","ReturnTime":"1517280655067","BaseBuy":307.05930436561505,"CompactBuy":-34,"BaseSell":0.003084981280661941,"CompactSell":81,"Rate":0,"Block":5635245},"OMG":{"Valid":true,"Error":"","Timestamp":"1517280654737","ReturnTime":"1517280655067","BaseBuy":65.0580993582104,"CompactBuy":32,"BaseSell":0.014925950060437398,"CompactSell":-14,"Rate":0,"Block":5635245},"SALT":{"Valid":true,"Error":"","Timestamp":"1517280654737","ReturnTime":"1517280655067","BaseBuy":152.3016783627643,"CompactBuy":9,"BaseSell":0.006196212698403499,"CompactSell":23,"Rate":0,"Block":5635245},"SNT":{"Valid":true,"Error":"","Timestamp":"1517280654737","ReturnTime":"1517280655067","BaseBuy":4053.2170631085987,"CompactBuy":43,"BaseSell":0.000233599514875301,"CompactSell":-3,"Rate":0,"Block":5635245}}},{"Version":0,"Valid":true,"Error":"","Timestamp":"1517280657740","ReturnTime":"1517280658058","Data":{"EOS":{"Valid":true,"Error":"","Timestamp":"1517280657740","ReturnTime":"1517280658058","BaseBuy":87.21360760013062,"CompactBuy":0,"BaseSell":0.0128686459657361,"CompactSell":0,"Rate":0,"Block":5635245},"ETH":{"Valid":true,"Error":"","Timestamp":"1517280657740","ReturnTime":"1517280658058","BaseBuy":0,"CompactBuy":32,"BaseSell":0,"CompactSell":-14,"Rate":0,"Block":5635245},"KNC":{"Valid":true,"Error":"","Timestamp":"1517280657740","ReturnTime":"1517280658058","BaseBuy":307.05930436561505,"CompactBuy":-34,"BaseSell":0.003084981280661941,"CompactSell":81,"Rate":0,"Block":5635245},"OMG":{"Valid":true,"Error":"","Timestamp":"1517280657740","ReturnTime":"1517280658058","BaseBuy":65.0580993582104,"CompactBuy":32,"BaseSell":0.014925950060437398,"CompactSell":-14,"Rate":0,"Block":5635245},"SALT":{"Valid":true,"Error":"","Timestamp":"1517280657740","ReturnTime":"1517280658058","BaseBuy":152.3016783627643,"CompactBuy":9,"BaseSell":0.006196212698403499,"CompactSell":23,"Rate":0,"Block":5635245},"SNT":{"Valid":true,"Error":"","Timestamp":"1517280657740","ReturnTime":"1517280658058","BaseBuy":4053.2170631085987,"CompactBuy":43,"BaseSell":0.000233599514875301,"CompactSell":-3,"Rate":0,"Block":5635245}}},{"Version":0,"Valid":true,"Error":"","Timestamp":"1517280660736","ReturnTime":"1517280661076","Data":{"EOS":{"Valid":true,"Error":"","Timestamp":"1517280660736","ReturnTime":"1517280661076","BaseBuy":87.21360760013062,"CompactBuy":0,"BaseSell":0.0128686459657361,"CompactSell":0,"Rate":0,"Block":5635245},"ETH":{"Valid":true,"Error":"","Timestamp":"1517280660736","ReturnTime":"1517280661076","BaseBuy":0,"CompactBuy":32,"BaseSell":0,"CompactSell":-14,"Rate":0,"Block":5635245},"KNC":{"Valid":true,"Error":"","Timestamp":"1517280660736","ReturnTime":"1517280661076","BaseBuy":307.05930436561505,"CompactBuy":-34,"BaseSell":0.003084981280661941,"CompactSell":81,"Rate":0,"Block":5635245},"OMG":{"Valid":true,"Error":"","Timestamp":"1517280660736","ReturnTime":"1517280661076","BaseBuy":65.0580993582104,"CompactBuy":32,"BaseSell":0.014925950060437398,"CompactSell":-14,"Rate":0,"Block":5635245},"SALT":{"Valid":true,"Error":"","Timestamp":"1517280660736","ReturnTime":"1517280661076","BaseBuy":152.3016783627643,"CompactBuy":9,"BaseSell":0.006196212698403499,"CompactSell":23,"Rate":0,"Block":5635245},"SNT":{"Valid":true,"Error":"","Timestamp":"1517280660736","ReturnTime":"1517280661076","BaseBuy":4053.2170631085987,"CompactBuy":43,"BaseSell":0.000233599514875301,"CompactSell":-3,"Rate":0,"Block":5635245}}},{"Version":0,"Valid":true,"Error":"","Timestamp":"1517280663736","ReturnTime":"1517280664068","Data":{"EOS":{"Valid":true,"Error":"","Timestamp":"1517280663736","ReturnTime":"1517280664068","BaseBuy":87.21360760013062,"CompactBuy":2,"BaseSell":0.0128686459657361,"CompactSell":-119,"Rate":0,"Block":5635255},"ETH":{"Valid":true,"Error":"","Timestamp":"1517280663736","ReturnTime":"1517280664068","BaseBuy":0,"CompactBuy":0,"BaseSell":0,"CompactSell":0,"Rate":0,"Block":5635255},"KNC":{"Valid":true,"Error":"","Timestamp":"1517280663736","ReturnTime":"1517280664068","BaseBuy":307.05930436561505,"CompactBuy":-31,"BaseSell":0.003084981280661941,"CompactSell":77,"Rate":0,"Block":5635255},"OMG":{"Valid":true,"Error":"","Timestamp":"1517280663736","ReturnTime":"1517280664068","BaseBuy":65.0580993582104,"CompactBuy":0,"BaseSell":0.014925950060437398,"CompactSell":0,"Rate":0,"Block":5635255},"SALT":{"Valid":true,"Error":"","Timestamp":"1517280663736","ReturnTime":"1517280664068","BaseBuy":152.3016783627643,"CompactBuy":8,"BaseSell":0.006196212698403499,"CompactSell":21,"Rate":0,"Block":5635255},"SNT":{"Valid":true,"Error":"","Timestamp":"1517280663736","ReturnTime":"1517280664068","BaseBuy":4053.2170631085987,"CompactBuy":0,"BaseSell":0.000233599514875301,"CompactSell":0,"Rate":0,"Block":5635255}}}],"success":true}
```


### Get trade history for an account (signing required)
```
  <host>:8000/tradehistory  
  params:
  - fromTime: millisecond (required)
  - toTime: millisecond (required)
  Restriction: toTime - fromTime <= 3 days (in millisecond)
```

eg:
```
curl -X GET "http://localhost:8000/tradehistoryfromTime=1516116380102&toTime=18446737278344972745"
```
response:
```
{"data":{"Version":1517298257114,"Valid":true,"Timestamp":"1517298257115","Data":{"binance":{"EOS-ETH":[],"KNC-ETH":[{"ID":"548002","Price":0.003038,"Qty":50,"Type":"buy","Timestamp":1516116380102},{"ID":"548003","Price":0.0030384,"Qty":7,"Type":"buy","Timestamp":1516116380102},{"ID":"548004","Price":0.003043,"Qty":16,"Type":"buy","Timestamp":1516116380102},{"ID":"548005","Price":0.0030604,"Qty":29,"Type":"buy","Timestamp":1516116380102},{"ID":"548006","Price":0.003065,"Qty":29,"Type":"buy","Timestamp":1516116380102},{"ID":"548007","Price":0.003065,"Qty":130,"Type":"buy","Timestamp":1516116380102}],"OMG-ETH":[{"ID":"123980","Price":0.020473,"Qty":48,"Type":"buy","Timestamp":1512395498231},{"ID":"130518","Price":0.021022,"Qty":13.49,"Type":"buy","Timestamp":1512564108827},{"ID":"130706","Price":0.020202,"Qty":9.93,"Type":"sell","Timestamp":1512569059460},{"ID":"140078","Price":0.019098,"Qty":11.07,"Type":"buy","Timestamp":1512714826339},{"ID":"140157","Price":0.019053,"Qty":7.68,"Type":"sell","Timestamp":1512716338997},{"ID":"295923","Price":0.020446,"Qty":4,"Type":"buy","Timestamp":1514360742162}],"SALT-ETH":[],"SNT-ETH":[]},"bittrex":{"OMG-ETH":[{"ID":"eb948865-6261-4991-8615-b36c8ccd1256","Price":0.01822057,"Qty":1,"Type":"buy","Timestamp":18446737278344972745}],"SALT-ETH":[],"SNT-ETH":[]}}},"success":true}
```



### Get exchange balances, reserve balances, pending activities at once (signing required)
```
<host>:8000/authdata
```

eg:
```
curl -X GET "http://localhost:8000/authdata"
```
response:
```
{"data":{"Valid":true,"Error":"","Timestamp":"1514114408227","ReturnTime":"1514114408810","ExchangeBalances":{"bittrex":{"Valid":true,"Error":"","Timestamp":"1514114408226","ReturnTime":"1514114408461","AvailableBalance":{"ETH":0.10704306,"OMG":2.97381136},"LockedBalance":{"ETH":0,"OMG":0},"DepositBalance":{"ETH":0,"OMG":0}}},"ReserveBalances":{"ADX":{"Valid":true,"Error":"","Timestamp":"1514114408461","ReturnTime":"1514114408799","Balance":0},"BAT":{"Valid":true,"Error":"","Timestamp":"1514114408461","ReturnTime":"1514114408799","Balance":0},"CVC":{"Valid":true,"Error":"","Timestamp":"1514114408461","ReturnTime":"1514114408799","Balance":0},"DGD":{"Valid":true,"Error":"","Timestamp":"1514114408461","ReturnTime":"1514114408799","Balance":0},"EOS":{"Valid":true,"Error":"","Timestamp":"1514114408461","ReturnTime":"1514114408799","Balance":0},"ETH":{"Valid":true,"Error":"","Timestamp":"1514114408461","ReturnTime":"1514114408799","Balance":360169992138038352},"FUN":{"Valid":true,"Error":"","Timestamp":"1514114408461","ReturnTime":"1514114408799","Balance":0},"GNT":{"Valid":true,"Error":"","Timestamp":"1514114408461","ReturnTime":"1514114408799","Balance":0},"KNC":{"Valid":true,"Error":"","Timestamp":"1514114408461","ReturnTime":"1514114408799","Balance":0},"LINK":{"Valid":true,"Error":"","Timestamp":"1514114408461","ReturnTime":"1514114408799","Balance":0},"MCO":{"Valid":true,"Error":"","Timestamp":"1514114408461","ReturnTime":"1514114408799","Balance":0},"OMG":{"Valid":true,"Error":"","Timestamp":"1514114408461","ReturnTime":"1514114408799","Balance":23818094310417195708},"PAY":{"Valid":true,"Error":"","Timestamp":"1514114408461","ReturnTime":"1514114408799","Balance":0}},"PendingActivities":[]},"block": 2345678, "success":true,"timestamp":"1514114409088","version":39}
```

### Deposit to exchanges (signing required)
```
<host>:8000/deposit/:exchange_id
POST request
Form params:
  - amount: little endian hex string (must starts with 0x), eg: 0xde0b6b3a7640000
  - token: token id string, eg: ETH, EOS...
```

eg:
```
curl -X POST \
  http://localhost:8000/deposit/binance\
  -H 'content-type: multipart/form-data' \
  -F token=EOS \
  -F amount=0xde0b6b3a7640000
```
Response:

```json
{
    "hash": "0x1b0c09f059904f1a9587641f2357c16c1c9fe43dfea161db31607f9221b0cfbb",
    "success": true
}
```
Where `hash` is the transaction hash

### Withdraw from exchanges (signing required)
```
<host>:8000/withdraw/:exchange_id
POST request
Form params:
  - amount: little endian hex string (must starts with 0x), eg: 0xde0b6b3a7640000
  - token: token id string, eg: ETH, EOS...
```

eg:
```
curl -X POST \
  http://localhost:8000/withdraw/binance\
  -H 'content-type: multipart/form-data' \
  -F token=EOS \
  -F amount=0xde0b6b3a7640000
```
Response:

```json
{
    "success": true
}
```
Where `hash` is the transaction hash

### Setting rates (signing required)
```
<host>:8000/setrates
POST request
Form params:
  - tokens: string, not including "ETH", represent all base token IDs separated by "-", eg: "ETH-ETH"
  - buys: string, represent all the buy (end users to buy tokens by ether) prices in little endian hex string, rates are separated by "-", eg: "0x5-0x7"
  - sells: string, represent all the sell (end users to sell tokens to ether) prices in little endian hex string, rates are separated by "-", eg: "0x5-0x7"
  - afp_mid: string, represent all the afp mid (average filled price) in little endian hex string, rates are separated by "-", eg: "0x5-0x7" (this rate only stores in activities for tracking)
  - block: number, in base 10, the block that prices are calculated on, eg: "3245876" means the prices are calculated from data at the time of block 3245876
```
eg:
```
curl -X POST \
  http://localhost:8000/setrates \
  -H 'content-type: multipart/form-data' \
  -F tokens=KNC-EOS \
  -F buys=0x5-0x7 \
  -F sells=0x5-0x7 \
  -F afp_mid=0x5-0x7 \
  -F block=2342353
```

### Trade (signing required)
```
<host>:8000/trade/:exchange_id
POST request
Form params:
  - base: token id string, eg: ETH, EOS...
  - quote: token id string, eg: ETH, EOS...
  - amount: float
  - rate: float
  - type: "buy" or "sell"
```

eg:
```
curl -X POST \
  http://localhost:8000/trade/binance\
  -F base=ETH \
  -F quote=KNC \
  -F rate=300 \
  -F type=buy \
  -F amount=0.01
```
Response:

```json
{
    "id": "19234634",
    "success": true,
    "done": 0,
    "remaining": 0.01,
    "finished": false
}
```
Where `hash` is the transaction hash

### Cancel order (signing required)
```
<host>:8000/cancelorder/:exchange
POST request
Form params:
  - base: token id string, eg: ETH, EOS...
  - quote: token id string, eg: ETH, EOS...
  - order_id: string
```

response:
```json
{
    "reason": "UNKNOWN_ORDER",
    "success": false
}
```

### Get all activityes (signing required)
```
<host>:8000/activities
GET request
url params: 
  fromTime: from timepoint - uint64, unix millisecond (optional if empty then get from first activity)
  toTime: to timepoint - uint64, unix millisecond (optional if empty then get to last activity)
```
Note: `fromTime` and `toTime` shouldn't be included into signing message.
### Get immediate pending activities (signing required)
```
<host>:8000/immediate-pending-activities
GET request
```

### Store processed data (signing required)
```
<host>:8000/metrics
POST request
form params:
  - timestamp: uint64, unix millisecond
  - data: string, in format of <token>_afpmid_spread|<token>_afpmid_spread|..., eg. OMG_0.4_5|KNC_1_2
```

### Get processed data (signing required)
```
<host>:8000/metrics
GET request
url params:
  - tokens: string, list of tokens to get data about, in format of <token_id>-<token_id>..., eg. OMG_DGD_KNC
  - from: uint64, unix millisecond
  - to: uint64, unix millisecond
```

response:
```
{
    "data": {
        "DGD": [
            {
                "Timestamp": 19,
                "AfpMid": 4,
                "Spread": 5
            }
        ],
        "OMG": [
            {
                "Timestamp": 19,
                "AfpMid": 0.9,
                "Spread": 1
            }
        ]
    },
    "returnTime": 1514966512560,
    "success": true,
    "timestamp": 1514966512549
}
```
Returned data will only include datas that have timestamp in range of `[from, to]`


### Get pending token target quantity (signing required)
```
<host>:8000/pendingtargetqty
GET request
```

response:
```
  {
    "success": true,
    "data":{"ID":1517396850670,"Timestamp":0,"Data":"EOS_750_500_0.25_0.25|ETH_750_500_0.25_0.25|KNC_750_500_0.25_0.25|OMG_750_500_0.25_0.25|SALT_750_500_0.25_0.25","Status":"unconfirmed"}
  }
```

### Get token target quantity (signing required)
```
<host>:8000/targetqty
GET request
```

response:
```
  {
    "success": true,
    "data":{"ID":1517396850670,"Timestamp":0,"Data":"EOS_750_500_0.25_0.25|ETH_750_500_0.25_0.25|KNC_750_500_0.25_0.25|OMG_750_500_0.25_0.25|SALT_750_500_0.25_0.25","Status":"confirmed"}
  }
```
response if there no data yet:
```
  {
    "success": false,
    "reason": "Version doesn't exist: 1517481572058"
  }
```

### Set token target quantity (signing required)
```
<host>:8000/settargetqty
POST request
form params:
  - data: required, string, must sort by token id by ascending order
  - action: required, string, set/confirm/cancel, action to set, confirm or cancel target quantity
  - id: optional, required to confirm target quantity
  - type: required, number, data type (now it should be 1)
```
eg:
```
curl -X POST \
  http://localhost:8000/settargetqty \
  -H 'content-type: multipart/form-data' \
  -F data= EOS_750_500_0.25_0.25|ETH_750_500_0.25_0.25|KNC_750_500_0.25_0.25|OMG_750_500_0.25_0.25|SALT_750_500_0.25_0.25 \
  -F action=set
  -F id=1517396850670
```
response
```
  {
    "success": true,
    "data":{"ID":1517396850670,"Timestamp":0,"Data":"EOS_750_500_0.25_0.25|ETH_750_500_0.25_0.25|KNC_750_500_0.25_0.25|OMG_750_500_0.25_0.25|SALT_750_500_0.25_0.25","Status":"unconfirmed"}
  }
```

### Confirm token target quantity (signing required)
```
<host>:8000/confirmtargetqty
POST request
form params:
  - data: required, string, must sort by token id by ascending order
  - id: optional, required to confirm target quantity
```
eg:
```
curl -X POST \
  http://localhost:8000/confirmtargetqty \
  -H 'content-type: multipart/form-data' \
  -F data= EOS_750_500_0.25_0.25|ETH_750_500_0.25_0.25|KNC_750_500_0.25_0.25|OMG_750_500_0.25_0.25|SALT_750_500_0.25_0.25 \
  -F id=1517396850670
```
response
```
  {
    "success": true,
    "data":{"ID":1517396850670,"Timestamp":0,"Data":"EOS_750_500_0.25_0.25|ETH_750_500_0.25_0.25|KNC_750_500_0.25_0.25|OMG_750_500_0.25_0.25|SALT_750_500_0.25_0.25","Status":"unconfirmed"}
  }
```

### Cancel token target quantity (signing required)
```
<host>:8000/confirmtargetqty
POST request
```
eg:
```
curl -X POST \
  http://localhost:8000/confirmtargetqty \
  -H 'content-type: multipart/form-data' \
```
response
```
  {
    "success": true,
  }
```

### Get rebalance status
Get rebalance status, if reponse is *true* then rebalance is enable, the analytic can perform rebalance, else reponse is *false*, the analytic hold rebalance ability.
```
<host>:8000/rebalancestatus
GET request
```

response
```
  {
    "success": true,
    "data": true
  }
```

### Hold rebalance
```
<host>:8000/holdrebalance
POST request
```
eg:
```
curl -X POST \
  http://localhost:8000/holdrebalance \
  -H 'content-type: multipart/form-data' \
```
response
```
  {
    "success": true
  }
```

### Enable rebalance
```
<host>:8000/enablerebalance
POST request
```
eg:
```
curl -X POST \
  http://localhost:8000/enablerebalance \
  -H 'content-type: multipart/form-data' \
```
response
```
  {
    "success": true
  }
```

### Get setrate status
Get setrate status, if reponse is *true* then setrate is enable, the analytic can perform setrate, else reponse is *false*, the analytic hold setrate ability.
```
<host>:8000/setratestatus
GET request
```

response
```
  {
    "success": true,
    "data": true
  }
```

### Hold setrate
```
<host>:8000/holdsetrate
POST request
```
eg:
```
curl -X POST \
  http://localhost:8000/holdsetrate \
  -H 'content-type: multipart/form-data' \
```
response
```
  {
    "success": true
  }
```

### Enable setrate
```
<host>:8000/enablesetrate
POST request
```
eg:
```
curl -X POST \
  http://localhost:8000/enablesetrate \
  -H 'content-type: multipart/form-data' \
```
response
```
  {
    "success": true
  }
```

### Get pending pwis equation (signing required)
```
<host>:8000/pending-pwis-equation
GET request
```

response:
```
  {
    "success": true,
    "data":{"id":1517396850670,"data":"EOS_750_500_0.25|ETH_750_500_0.25|KNC_750_500_0.25|OMG_750_500_0.25|SALT_750_500_0.25"}
  }
```

### Get pwis equation (signing required)
```
<host>:8000/pwis-equation
GET request
```

response:
```
  {
    "success": true,
    "data":{"id":1517396850670,"data":"EOS_750_500_0.25|ETH_750_500_0.25|KNC_750_500_0.25|OMG_750_500_0.25|SALT_750_500_0.25"}
  }
```

### Set pwis equation (signing required)
```
<host>:8000/set-pwis-equation
POST request
form params:
  - data: required, string, must sort by token id by ascending order
  - id: optional, required to confirm target quantity
```
eg:
```
curl -X POST \
  http://localhost:8000/set-pwis-equation \
  -H 'content-type: multipart/form-data' \
  -F data= EOS_750_500_0.25|ETH_750_500_0.25|KNC_750_500_0.25|OMG_750_500_0.25|SALT_750_500_0.25 \
  -F id=1517396850670
```
response
```
  {
    "success": true,
  }
```

### Confirm pwis equation (signing required)
```
<host>:8000/confirm-pwis-equation
POST request
form params:
  - data: required, string, must sort by token id by ascending order
```
eg:
```
curl -X POST \
  http://localhost:8000/confirm-pwis-equation \
  -H 'content-type: multipart/form-data' \
  -F data=EOS_750_500_0.25|ETH_750_500_0.25|KNC_750_500_0.25|OMG_750_500_0.25|SALT_750_500_0.25
```
response
```
  {
    "success": true,
  }
```

### Reject pwis equation (signing required)
```
<host>:8000/reject-pwis-equation
POST request
```
eg:
```
curl -X POST \
  http://localhost:8000/reject-pwis-equation \
  -H 'content-type: multipart/form-data' \
```
response
```
  {
    "success": true,
  }
```

### Get trade logs

```
<host>:8000/tradelogs
GET request
```
response
```
  {
    "success": true
  }
```

### Get asset volume for aggregate time (hour, day, month)
```
<host>:8000/get-asset-volume
GET request

Url params:
  - fromTime (millisecond - required): from time stamp
  - toTime (millisecond - required: to time stamp
  - freq (required): frequency to get data (H/D/M)
  - asset (required): asset name (eg: ETH)
```

```
curl -x GET http://localhost:8000/get-asset-volume?fromTime=1520640035000&toTime=1520722835000&asset=eth&freq=M
```
response
```
  {"data":{"1520652360000":{"usd_amount":0.734518,"volume":0.001},"1520654280000":{"usd_amount":0.7297319999999999,"volume":0.001},"1520654820000":{"usd_amount":1.4581552500230603,"volume":0.001998206533389053},"1520656440000":{"usd_amount":0.7297319999999999,"volume":0.001},"1520656500000":{"usd_amount":0.7297319999999999,"volume":0.001},"1520656560000":{"usd_amount":0.7297319999999999,"volume":0.001}},"success":true}
```

### Get burn fee follow aggregate time (hour, day, month)
```
<host>:8000/get-burn-fee
GET request

Url params:
  - fromTime (millisecond - required): from time stamp
  - toTime (millisecond - required: to time stamp
  - freq (required): frequency to get data (H/D/M)
  - reserveAddr (required): reserve address to get burn fee
```

```
curl -x GET http://localhost:8000/get-burn-fee?fromTime=1520640035000&toTime=1520722835000&reserveAddr=0x2c5a182d280eeb5824377b98cd74871f78d6b8bc&freq=H
```
response
```
  {"data":{"1520650800000":0.00225,"1520654400000":0.005622982350062684},"success":true}
```

### Get wallet fee follow aggregate time (hour, day, month)
```
<host>:8000/get-wallet-fee
GET request

Url params:
  - fromTime (millisecond - required): from time stamp
  - toTime (millisecond - required: to time stamp
  - freq (required): frequency to get data (H/D/M)
  - reserveAddr (required): reserve address
  - walletAddr (required): wallet address to get fee
```

```
curl -x GET http://localhost:8000/get-wallet-fee?fromTime=1520640035000&toTime=1520722835000&reserveAddr=0x2c5a182d280eeb5824377b98cd74871f78d6b8bc&walletAddr=0x0000000000000000000000000000000000000000&freq=H
```
response
```
  {"data":{"1520650800000":0,"1520654400000":0},"success":true}
```


### Get user volume follow aggregate time (hour, day, month)
```
<host>:8000/get-user-volume
GET request

Url params:
  - fromTime (millisecond - required): from time stamp
  - toTime (millisecond - required: to time stamp
  - freq (required): frequency to get data (H/D/M)
  - userAddr (required): user address to get volume
```

```
curl -x GET http://localhost:8000/get-user-volume?fromTime=1520640035000&toTime=1520722835000&userAddr=0x8fa07f46353a2b17e92645592a94a0fc1ceb783f&freq=H
```
response
```
  {"data":{"1520650800000":0.734518,"1520654400000":0},"success":true}  
```

### Get rate from blockchain follow reserve (including sanity rate)
```
<host>:8000/get-reserve-rate
GET request

URL params:
  - fromTime (millisecond - required): from timestamp
  - toTime (millisecond - required): to timestamp
  - reserveAddr (required): Address of the reserve to get rate from
```

```
curl -x GET http://localhost:8000/get-reserve-rate?fromTime=1520650426000&reserveAddr=0x2C5a182d280EeB5824377B98CD74871f78d6b8BC
```

response
```
{"data":[{"Timestamp":0,"ReturnTime":1520655211398,"BlockNumber":5228238,"Data":{"APPC-ETH":{"ReserveRate":0.008393501685222925,"SanityRate":0.009476954807692308},"BAT-ETH":{"ReserveRate":0.004239837479770336,"SanityRate":0.004723026},"BQX-ETH":{"ReserveRate":0.000584106942517358,"SanityRate":0.000652623583333333},"ELF-ETH":{"ReserveRate":0.000111035861616385,"SanityRate":0.000123576933333333},"ENG-ETH":{"ReserveRate":0.000596961062855617,"SanityRate":0.000671348333333333},"EOS-ETH":{"ReserveRate":0.002752586323625439,"SanityRate":0.0029518775},"ETH-APPC":{"ReserveRate":117.09352189952426,"SanityRate":127.67814393478591},"ETH-BAT":{"ReserveRate":229.70817443088293,"SanityRate":256.19168727845243},"ETH-BQX":{"ReserveRate":1674.1030165099337,"SanityRate":1854.0549727299406},"ETH-ELF":{"ReserveRate":8741.854159397268,"SanityRate":9791.471331758756},"ETH-ENG":{"ReserveRate":1615.634600168794,"SanityRate":1802.3430459597466},"ETH-EOS":{"ReserveRate":356.63560217559376,"SanityRate":409.9086090123997},"ETH-GTO":{"ReserveRate":377.41338205276884,"SanityRate":432.7020247561754},"ETH-KNC":{"ReserveRate":3343.445798727388,"SanityRate":3740.5264791019335},"ETH-MANA":{"ReserveRate":2653.110602592891,"SanityRate":2961.048749629869},"ETH-OMG":{"ReserveRate":221.03211631662654,"SanityRate":247.71892159696137},"ETH-POWR":{"ReserveRate":2625.724091042635,"SanityRate":2849.3656923419003},"ETH-RDN":{"ReserveRate":49.46371879742714,"SanityRate":54.91076347236177},"ETH-REQ":{"ReserveRate":5123.294220111987,"SanityRate":5665.576472406068},"ETH-SALT":{"ReserveRate":532.7984920557698,"SanityRate":611.1450636146453},"ETH-SNT":{"ReserveRate":924.3533982883454,"SanityRate":1052.547224086108},"GTO-ETH":{"ReserveRate":0.002590177384056749,"SanityRate":0.002796381645502645},"KNC-ETH":{"ReserveRate":0.0002894500017402,"SanityRate":0.000323483875},"MANA-ETH":{"ReserveRate":0.00036813437957934,"SanityRate":0.000408639},"OMG-ETH":{"ReserveRate":0.004383746019560721,"SanityRate":0.004884568333333332},"POWR-ETH":{"ReserveRate":0.000369936605210205,"SanityRate":0.000424655916666666},"RDN-ETH":{"ReserveRate":0.01987031936393942,"SanityRate":0.02203575261904761},"REQ-ETH":{"ReserveRate":0.000191920526182855,"SanityRate":0.0002135705},"SALT-ETH":{"ReserveRate":0.001821407074188612,"SanityRate":0.00197989},"SNT-ETH":{"ReserveRate":0.001042608668464954,"SanityRate":0.001149592125}}},{"Timestamp":0,"ReturnTime":1520655227886,"BlockNumber":5228239,"Data":{"APPC-ETH":{"ReserveRate":0.000369936605210205,"SanityRate":0.000424655916666666},"BAT-ETH":{"ReserveRate":0.0002894500017402,"SanityRate":0.000323483875},"BQX-ETH":{"ReserveRate":0.002590177384056749,"SanityRate":0.002796381645502645},"ELF-ETH":{"ReserveRate":0.01987031936393942,"SanityRate":0.02203575261904761},"ENG-ETH":{"ReserveRate":0.000584106942517358,"SanityRate":0.000652623583333333},"EOS-ETH":{"ReserveRate":0.000191920526182855,"SanityRate":0.0002135705},"ETH-APPC":{"ReserveRate":2625.724091042635,"SanityRate":2849.3656923419003},"ETH-BAT":{"ReserveRate":3343.445798727388,"SanityRate":3740.5264791019335},"ETH-BQX":{"ReserveRate":377.41338205276884,"SanityRate":432.7020247561754},"ETH-ELF":{"ReserveRate":49.46371879742714,"SanityRate":54.91076347236177},"ETH-ENG":{"ReserveRate":1674.1030165099337,"SanityRate":1854.0549727299406},"ETH-EOS":{"ReserveRate":5123.294220111987,"SanityRate":5665.576472406068},"ETH-GTO":{"ReserveRate":229.70817443088293,"SanityRate":256.19168727845243},"ETH-KNC":{"ReserveRate":1615.634600168794,"SanityRate":1802.3430459597466},"ETH-MANA":{"ReserveRate":221.03211631662654,"SanityRate":247.71892159696137},"ETH-OMG":{"ReserveRate":8741.854159397268,"SanityRate":9791.471331758756},"ETH-POWR":{"ReserveRate":924.3533982883454,"SanityRate":1052.547224086108},"ETH-RDN":{"ReserveRate":532.7984920557698,"SanityRate":611.1450636146453},"ETH-REQ":{"ReserveRate":117.09352189952426,"SanityRate":127.67814393478591},"ETH-SALT":{"ReserveRate":2653.110602592891,"SanityRate":2961.048749629869},"ETH-SNT":{"ReserveRate":356.63560217559376,"SanityRate":409.9086090123997},"GTO-ETH":{"ReserveRate":0.004239837479770336,"SanityRate":0.004723026},"KNC-ETH":{"ReserveRate":0.000596961062855617,"SanityRate":0.000671348333333333},"MANA-ETH":{"ReserveRate":0.004383746019560721,"SanityRate":0.004884568333333332},"OMG-ETH":{"ReserveRate":0.000111035861616385,"SanityRate":0.000123576933333333},"POWR-ETH":{"ReserveRate":0.001042608668464954,"SanityRate":0.001149592125},"RDN-ETH":{"ReserveRate":0.001821407074188612,"SanityRate":0.00197989},"REQ-ETH":{"ReserveRate":0.008393501685222925,"SanityRate":0.009476954807692308},"SALT-ETH":{"ReserveRate":0.00036813437957934,"SanityRate":0.000408639},"SNT-ETH":{"ReserveRate":0.002752586323625439,"SanityRate":0.0029518775}}}],"success":true}
```


### Get trade summary follow timeframe (day)
```
<host>:8000/get-trade-summary
GET request

Url params:
  - fromTime (millisecond - required): from time stamp
  - toTime (millisecond - required): to time stamp  
  - timeZone (in range [-12,14], default to 0): the integer specific which UTC timezone to query
```

```
curl -x GET http://localhost:8000/get-trade-summary?fromTime=1519297149000&toTime=1519815549000
```
response
```
{"data":{"1519344000000":{"eth_per_trade":0.55402703087424,"kyced_addresses":0,"new_unique_addresses":35,"total_burn_fee":0,"total_eth_volume":44.3221624699392,"total_trade":80,"total_usd_amount":30981.281202536768,"unique_addresses":50,"usd_per_trade":387.26601503170957},"1519430400000":{"eth_per_trade":0.17008867987348247,"kyced_addresses":0,"new_unique_addresses":17,"total_burn_fee":0,"total_eth_volume":8.674522673547607,"total_trade":51,"total_usd_amount":6060.828270348999,"unique_addresses":29,"usd_per_trade":118.83977000684311},"1519516800000":{"eth_per_trade":0.14234886960871,"kyced_addresses":0,"new_unique_addresses":9,"total_burn_fee":1.1025,"total_eth_volume":5.40925704513098,"total_trade":38,"total_usd_amount":3779.4100326337,"unique_addresses":18,"usd_per_trade":99.45815875351843},"1519603200000":{"eth_per_trade":0.5430574166436676,"kyced_addresses":0,"new_unique_addresses":39,"total_burn_fee":42.85336706164196,"total_eth_volume":45.07376558142441,"total_trade":83,"total_usd_amount":31497.3427579499,"unique_addresses":56,"usd_per_trade":379.4860573246976},"1519689600000":{"eth_per_trade":0.6014134385918366,"kyced_addresses":0,"new_unique_addresses":69,"total_burn_fee":79.03472646631772,"total_eth_volume":78.7851604555306,"total_trade":131,"total_usd_amount":55076.026979006005,"unique_addresses":92,"usd_per_trade":420.4276868626413},"1519776000000":{"eth_per_trade":0.40083191776618454,"kyced_addresses":0,"new_unique_addresses":64,"total_burn_fee":48.899026261678536,"total_eth_volume":52.50898122737018,"total_trade":131,"total_usd_amount":36662.138255818456,"unique_addresses":94,"usd_per_trade":279.8636508077745}},"success":true}
```

### Get a specific wallet's stats summary follow timeframe (day)
```
<host>:8000/get-wallet-stats
GET request

Url params:
  - fromTime (millisecond - required): from time stamp
  - toTime (millisecond - required): to time stamp  
  - timeZone (in range [-12,14], default to 0): the integer specific which UTC timezone to query
  - walletAddr (hex string - required) : to specific which wallet Address to query data from. It must be larger than 2^128 to be valid.
```

```
curl -x GET http://localhost:8000/get-wallet-stats?fromTime=1521914061000&toTime=1523000461000&walletAddr=0xb9e29984fe50602e7a619662ebed4f90d93824c7
```
response
```
{"data":{"1521936000000":{"eth_per_trade":0.15169175185997197,"kyced_addresses":0,"new_unique_addresses":27,"total_burn_fee":3.5843774403434443,"total_eth_volume":9.101505111598318,"total_trade":60,"total_usd_amount":4738.284168671162,"unique_addresses":40,"usd_per_trade":78.97140281118602},"1522022400000":{"eth_per_trade":0.1305336778977258,"kyced_addresses":0,"new_unique_addresses":13,"total_burn_fee":1.2758795269915402,"total_eth_volume":2.3496062021590642,"total_trade":18,"total_usd_amount":1230.3892752776494,"unique_addresses":18,"usd_per_trade":68.35495973764719}},"success":true}
```

### Get a list of wallet that has ever traded with core
```
<host>:8000/get-wallet-address 
GET request

URL params:
  Nil
```


```
curl -x GET http://localhost:8000/get-wallet-address
```
response
```
{"data":["0xb9e29984fe50602e7a619662ebed4f90d93824c7","0xf1aa99c69715f423086008eb9d06dc1e35cc504d"],"success":true}
``` 

### Get exchanges status
```
<host>:8000/get-exchange-status
GET request
```

eg:
```
curl -x GET http://localhost:8000/get-exchange-status
```

response:
```
{"data":{"binance":{"timestamp":1521532176702,"status":true},"bittrex":{"timestamp":1521532176704,"status":true},"huobi":{"timestamp":1521532176703,"status":true}},"success":true}
```

### Update exchanges status
```
<host>:8000/update-exchange-status
POST request

params: 
exchange (string): exchange name (eg: 'binance')
status (bool): true (up), false (down)
timestamp (integer): timestamp of the exchange status
```

eg:
```
curl -X POST \
  http://localhost:8000/update-exchange-status \
  -H 'content-type: multipart/form-data' \
  -F exchange=binance \
  -F status=false
```

### Get country stats
```
<host>:8000/get-country-stats
GET request
params:
 - fromTime (integer) - from timestamp (millisecond)
 - toTime (integer) - to timestamp (millisecond)
 - country (string) - internatinal country 
 - timezone (integer) - timezone to get country stats from -11 to 14
```

response: 
```
{"data":{"1522368000000":{"eth_per_trade":1.1759348083481784,"kyced_addresses":0,"new_unique_addresses":23,"total_burn_fee":40.10625390027786,"total_eth_volume":51.741131567319854,"total_trade":44,"total_usd_amount":19804.392524011764,"unique_addresses":26,"usd_per_trade":450.09983009117644}},"success":true}
```

### Get heatmap - list of countries sort by total ETH value
```
<host>:8000/get-heat-map
GET request
params:
 - fromTime (integer) - from timestamp (millisecond)
 - toTime (integer) - to timestamp (millisecond)
 - timezone (integer) - timezone to get country stats from -11 to 14
```

response:
```
{"data":[{"country":"US","total_eth_value":51.741131567319854,"total_fiat_value":19804.392524011764},{"country":"unknown","total_eth_value":31.28130484378119,"total_fiat_value":12268.937507634406},{"country":"TW","total_eth_value":15,"total_fiat_value":5916.6900000000005},{"country":"KR","total_eth_value":13.280037553077175,"total_fiat_value":5016.70456645198},{"country":"JP","total_eth_value":10.277090646,"total_fiat_value":3857.271305900826},{"country":"TH","total_eth_value":8.241091466923997,"total_fiat_value":3195.368602817533},{"country":"CA","total_eth_value":3.8122812821017558,"total_fiat_value":1445.8819158742285},{"country":"AU","total_eth_value":2.6,"total_fiat_value":969.02},{"country":"DE","total_eth_value":1.823287,"total_fiat_value":697.502009413},{"country":"ID","total_eth_value":1.7178731840736186,"total_fiat_value":674.8439050493492},{"country":"RO","total_eth_value":1.4009999999999998,"total_fiat_value":529.075415},{"country":"VN","total_eth_value":1.3951777988339262,"total_fiat_value":548.8376078547749},{"country":"CN","total_eth_value":1.0121575386522288,"total_fiat_value":401.6824093511598},{"country":"PL","total_eth_value":0.379699,"total_fiat_value":144.141714079},{"country":"FR","total_eth_value":0.319624,"total_fiat_value":122.92586391999998},{"country":"SG","total_eth_value":0.15642985716526572,"total_fiat_value":64.06928945889221},{"country":"ES","total_eth_value":0.09344946,"total_fiat_value":35.176806429959996},{"country":"XX","total_eth_value":0.09,"total_fiat_value":36.86148},{"country":"IN","total_eth_value":0.0714026952146661,"total_fiat_value":27.977050948875906},{"country":"AR","total_eth_value":0.02751473,"total_fiat_value":10.92519129691},{"country":"RU","total_eth_value":0.024162,"total_fiat_value":9.61210186},{"country":"SE","total_eth_value":0.023,"total_fiat_value":9.132541},{"country":"LV","total_eth_value":0.01,"total_fiat_value":3.9209899999999998},{"country":"AL","total_eth_value":0.003,"total_fiat_value":1.126449}],"success":true}
```

### Update Price Analytic Data - (signing required) set a record marking the condition because of which the set price is called. 
```
<host>:8000/update-price-analytic-data
POST request
params:
 - timestamp - the timestamp of the action (real time ) in millisecond
 - value - the json enconded object to save. 

Note: the data sent over must be encoded in Json in order to make it valid for output operation
  In Python, the data would be encoded as:
   data = {"timestamp": timestamp, "value": json.dumps(analytic_data)} 
 ```

response:
```
on success:
{"success":true}
on failure:
{"success":false,
 "reason":<error>}
```

### Get Price Analytic Data - (signing required) list of price analytic data, sorted by timestamp 
```
<host>:8000/get-get-price-analytic-data
GET request
params:
 - fromTime (integer) - from timestamp (millisecond)
 - toTime (integer) - to timestamp (millisecond)
```
example:
```
curl -x GET \
  http://localhost:8000/get-price-analytic-data?fromTime=1522753160000&toTime=1522755792000
```
 
response:
```
{
  "data": [
    {
      "Timestamp": 1522755271000,
      "Data": {
        "block_expiration": false,
        "trigger_price_update": true,
        "triggering_tokens_list": [
          {
            "ask_price": 0.002,
            "bid_price": 0.003,
            "mid afp_old_price": 0.34555,
            "mid_afp_price": 0.6555,
            "min_spread": 0.233,
            "token": "OMG"
          },
          {
            "ask_price": 0.004,
            "bid_price": 0.005,
            "mid afp_old_price": 0.21555,
            "mid_afp_price": 0.4355,
            "min_spread": 0.133,
            "token": "KNC"
          }
        ]
      }
    }
  ],
  "success": true
}
```

### Update exchange notifications 
```
<host>:8000/exchange-notification
POST request
params:
 - exchange (string) - exchange name
 - action (string) - action name
 - token (string) - token pair
 - fromTime (integer) - from timestamp
 - toTime (integer) - to timestamp
 - isWarning (bool) - is exchange warning or not
 - msg (string) - message for the notification
```

response:
```
  {
    "success": true
  }
```

### Get exchange notifications
```
<host>:8000/exchange-notifications
GET request
```

response:
```
{"data":{"binance":{"trade":{"OMG":{"fromTime":123,"toTime":125,"isWarning":true,"msg":"3 times"}}}},"success":true}
```

### Get reserve volume
```
<host>:8000/exchange-notifications
GET request
URL Params:
  - fromTime (integer): millisecond
  - toTime (integer): millisecond
  - token (string): name of token to get volume (eg: ETH)
  - reserveAddr (string): reserve address to get volume of token
  - freq (string): frequency to get volume ("M", "H", "D" - Minute, Hour, Day)
```

example:
```
curl -x GET \
http://localhost:8000/get-reserve-volume?fromTime=1522540800000&toTime=1522627200000&freq=D&token=KNC&reserveAddr=0x63825c174ab367968EC60f061753D3bbD36A0D8F
```

response:
```
{"data":{"1522540800000":{"eth_amount":9.971150530912206,"usd_amount":3838.6105908493496,"volume":3945.5899585215247},"1522627200000":{"eth_amount":14.749439804645423,"usd_amount":5766.650333669346,"volume":5884.90733954939}},"success":true}
```

### set stable token params - (signing required)
```
<host>:8000/set-stable-token-params
POST request
URL Params:
  - value (string) : the json enconded string, represent a map (string : interface)
```


response:
```
on success:
{"success":true}
on failure:
{"success":false,
 "reason":<error>}
```
### confirm stable token params - (signing required)
```
<host>:8000/confirm-stable-token-params
POST request
URL Params:
  - value (string) : the json enconded string, represent a map (string : interface), must be equal to current pending.
```


response:
```
on success:
{"success":true}
on failure:
{"success":false,
 "reason":<error>}
```

### reject stable token params - (signing required)
```
<host>:8000/reject-stable-token-params
POST request
URL Params:
  nil
```


response:
```
on success:
{"success":true}
on failure:
{"success":false,
 "reason":<error>}
```
### Get pending stable token params- (signing required) return the current pending stable token params
```
<host>:8000/pending-stable-token-params
GET request
params:
  - nonce (uint64) : the nonce to conform to signing requirement
```
example:
```
curl -x GET \
  http://localhost:8000/pending-token-params?nonce=111111
```
 
response:
```
{
  "data": {
    "DGX": {
      "AskSpread": 50,
      "BidSpread": 50,
      "PriceUpdateThreshold": 0.1
    }
  },
  "success": true
}
```

### Get stable token params- (signing required) return the current confirmed stable token params
```
<host>:8000/stable-token-params
GET request
params:
  - nonce (uint64) : the nonce to conform to signing requirement
```
example:
```
curl -x GET \
  http://localhost:8000/stable-token-params?nonce=111111
```
 
response:
```
{
  "data": {
    "DGX": {
      "AskSpread": 50,
      "BidSpread": 50,
      "PriceUpdateThreshold": 0.1
    }
  },
  "success": true
}
```
### Get heat map for token
```
<host>:8000/get-token-heatmap
URL Params:
  - fromTime (integer): millisecond
  - toTime (integer): millisecond
  - token (string): name of token to get heatmap
  - freq (string): frequencty to get volume ("M", "H", "D" - Minute, Hour, Day)
```

example:
```
curl -x GET \
http://localhost:8000/get-token-heatmap?fromTime=1518307200000&token=EOS&freq=D&toTime=1518911999999
```

response:
```
{"data":[{"country":"US","volume":2883.620428022146,"eth_volume":29.97000000311978,"usd_volume":28584.013502715607},{"country":"unknown","volume":663.7763113279779,"eth_volume":6.848675774186141,"usd_volume":5710.033060275751},{"country":"JP","volume":189.38349888667832,"eth_volume":1.99,"usd_volume":1881.86987},{"country":"KR","volume":93.83012247596538,"eth_volume":1,"usd_volume":857.766},{"country":"SI","volume":73.000042,"eth_volume":0.7584920000216375,"usd_volume":696.7810908998771},{"country":"IL","volume":9.757144977962138,"eth_volume":0.1,"usd_volume":85.47670000000001},{"country":"TH","volume":9.459436814264475,"eth_volume":0.1,"usd_volume":84.1759},{"country":"DE","volume":9.311558446913438,"eth_volume":0.09904,"usd_volume":85.93066944},{"country":"VN","volume":1.8918873628528947,"eth_volume":0.019789900740301923,"usd_volume":16.536080320374314}],"success":true}
```

### Get gold data
```
<host>:8000/gold-feed
```
response:
```
{"data":{"Timestamp":1526923808631,"DGX":{"Valid":true,"Timestamp":0,"success":"","data":[{"symbol":"DGXETH","price":0.06676463,"time":1526923801},{"symbol":"ETHUSD","price":694.4,"time":1526923801},{"symbol":"ETHSGD","price":931.89,"time":1526923801},{"symbol":"DGXUSD","price":46.36,"time":1526923801},{"symbol":"EURUSD","price":1.17732,"time":1526923801},{"symbol":"USDSGD","price":1.34201,"time":1526923801},{"symbol":"XAUUSD","price":1291.468,"time":1526923801},{"symbol":"USDJPY","price":111.061,"time":1526923801}],"Error":""},"OneForgeETH":{"Value":1.85646,"Text":"1 XAU is worth 1.85646 ETH","Timestamp":1526923803,"Error":false,"Message":""},"OneForgeUSD":{"Value":1291.57,"Text":"1 XAU is worth 1291.57 USD","Timestamp":1526923803,"Error":false,"Message":""},"GDAX":{"Valid":true,"Error":"","trade_id":34527604,"price":"695.56000000","size":"0.00894700","bid":"695.55","ask":"695.56","volume":"50497.82498957","time":"2018-05-21T17:30:04.729000Z"},"Kraken":{"Valid":true,"network_error":"","error":[],"result":{"XETHZUSD":{"a":["696.66000","1","1.000"],"b":["696.33000","4","4.000"],"c":["696.33000","0.10776064"],"v":["13536.83019524","16999.30348103"],"p":["707.93621","710.18316"],"t":[5361,8276],"l":["693.97000","693.97000"],"h":["721.38000","724.80000"],"o":"715.65000"}}},"Gemini":{"Valid":true,"Error":"","bid":"694.50","ask":"695.55","volume":{"ETH":"11418.5646926","USD":"8064891.13775284649999999999999999999704534","timestamp":1526923800000},"last":"695.36"}},"success":true}
```


### set target quantity v2 - (signing required)
```
<host>:8000/v2/settargetqty
POST request
URL Params:
  - value (string) : the json enconded string, represent a map (string : interface)
```


response:
```
on success:
{"success":true}
on failure:
{"success":false,
 "reason":<error>}
```
### confirm target quantity v2- (signing required)
```
<host>:8000/v2/confirmtargetqty
POST request
URL Params:
  - value (string) : the json enconded string, represent a map (string : interface), must be equal to current pending.
```


response:
```
on success:
{"success":true}
on failure:
{"success":false,
 "reason":<error>}
```

### cancel set target quantity v2- (signing required)
```
<host>:8000/v2/canceltargetqty
POST request
URL Params:
  nil
```


response:
```
on success:
{"success":true}
on failure:
{"success":false,
 "reason":<error>}
```
### Get pending target quantity - (signing required) return the current pending target quantity 
```
<host>:8000/v2/pendingtargetqty
GET request
params:
  - nonce (uint64) : the nonce to conform to signing requirement
```
example:
```
curl -x GET \
  http://localhost:8000/v2/pendingtargetqty?nonce=111111
```
 
response:
```
{
  "data": {
     "OMG" : {
        "TotalTarget": 1500,
        "ReserveTarget": 1005,
        "RebalanceThreshold": 0.33,
        "TransferThreshold": 0.2
    }
  },
  "success": true
}
```

### Get target quantity - (signing required) return the current confirmed target quantity 
```
<host>:8000/v2/targetqty
GET request
params:
  - nonce (uint64) : the nonce to conform to signing requirement
```
example:
```
curl -x GET \
  http://localhost:8000/v2/targetqty?nonce=111111
```
 
response:
```
{
  "data": {
    "OMG" : {
      "TotalTarget": 1500,
      "ReserveTarget": 1005,
      "RebalanceThreshold": 0.33,
      "TransferThreshold": 0.2
    }
  },
  "success": true
}
```

### Get pwis equation v2 - (signing required)

```
<host>:8000/v2/pwis-equation
GET request
```

eg:

```
curl -X "GET" "http://localhost:8000/v2/pwis-equation" \
     -H 'Content-Type: application/x-www-form-urlencoded' \
```

response:
```
{
  "data": {
    "EOS": {
      "ask": {
        "a": 800,
        "b": 600,
        "c": 0,
        "min_min_spread": 0,
        "price_multiply_factor": 0
      },
      "bid": {
        "a": 750,
        "b": 500,
        "c": 0,
        "min_min_spread": 0,
        "price_multiply_factor": 0
      }
    },
    "ETH": {
      "ask": {
        "a": 800,
        "b": 600,
        "c": 0,
        "min_min_spread": 0,
        "price_multiply_factor": 0
      },
      "bid": {
        "a": 750,
        "b": 500,
        "c": 0,
        "min_min_spread": 0,
        "price_multiply_factor": 0
      }
    }
  },
  "success": true
}
```

### Get pending pwis equation v2 - (signing required)

```
<host>:8000/v2/pending-pwis-equation
GET request

```

eg:
```
curl -X "GET" "http://localhost:8000/v2/pending-pwis-equation" \
     -H 'Content-Type: application/x-www-form-urlencoded' \
```

response:
```
{
  "data": {
    "EOS": {
      "ask": {
        "a": 800,
        "b": 600,
        "c": 0,
        "min_min_spread": 0,
        "price_multiply_factor": 0
      },
      "bid": {
        "a": 750,
        "b": 500,
        "c": 0,
        "min_min_spread": 0,
        "price_multiply_factor": 0
      }
    },
    "ETH": {
      "ask": {
        "a": 800,
        "b": 600,
        "c": 0,
        "min_min_spread": 0,
        "price_multiply_factor": 0
      },
      "bid": {
        "a": 750,
        "b": 500,
        "c": 0,
        "min_min_spread": 0,
        "price_multiply_factor": 0
      }
    }
  },
  "success": true
}
```

### Set pwis equation v2 - (signing required)

```
<host>:8000/v2/set-pwis-equation
POST request
Post form: json encoding data of pwis equation
```

eg:

```
curl -X "POST" "http://localhost:8000/v2/set-pwis-equation" \
     -H 'Content-Type: application/x-www-form-urlencoded' \
     --data-urlencode "data={
  \"EOS\": {
    \"bid\": {
      \"a\": 750,
      \"b\": 500,
      \"c\": 0,
      \"min_min_spread\": 0,
      \"price_multiply_factor\": 0
    },
    \"ask\": {
      \"a\": 800,
      \"b\": 600,
      \"c\": 0,
      \"min_min_spread\": 0,
      \"price_multiply_factor\": 0
    }
  },
  \"ETH\": {
    \"bid\": {
      \"a\": 750,
      \"b\": 500,
      \"c\": 0,
      \"min_min_spread\": 0,
      \"price_multiply_factor\": 0
    },
    \"ask\": {
      \"a\": 800,
      \"b\": 600,
      \"c\": 0,
      \"min_min_spread\": 0,
      \"price_multiply_factor\": 0
    }
  }
}"
```

response

```
  {
    "success": true,
  }
```

### Confirm pending pwis equation v2 - (signing required)

```
<host>:8000/v2/confirm-pwis-equation
POST request
Post form: json encoding data of pwis equation
```

eg

```
curl -X "POST" "http://localhost:8000/v2/confirm-pwis-equation" \
     -H 'Content-Type: application/x-www-form-urlencoded' \
     --data-urlencode "data={
  \"EOS\": {
    \"bid\": {
      \"a\": 750,
      \"b\": 500,
      \"c\": 0,
      \"min_min_spread\": 0,
      \"price_multiply_factor\": 0
    },
    \"ask\": {
      \"a\": 800,
      \"b\": 600,
      \"c\": 0,
      \"min_min_spread\": 0,
      \"price_multiply_factor\": 0
    }
  },
  \"ETH\": {
    \"bid\": {
      \"a\": 750,
      \"b\": 500,
      \"c\": 0,
      \"min_min_spread\": 0,
      \"price_multiply_factor\": 0
    },
    \"ask\": {
      \"a\": 800,
      \"b\": 600,
      \"c\": 0,
      \"min_min_spread\": 0,
      \"price_multiply_factor\": 0
    }
  }
}"
```

response

```
  {
    "success": true,
  }
```

### Reject pending pwis equation v2 - (signing required)

```
<host>:8000/v2/reject-pwis-equation
POST request
```

eg

```
curl -X "POST" "http://localhost:8000/v2/reject-pwis-equation" \
     -H 'Content-Type: application/x-www-form-urlencoded'
```

response

```
  {
    "success": true,
  }
```
### Setting APIs
#### Token related APIs
##### Update token - (signing required)
POST request
Post form: {"data" : "JSON enconding of token Object"}
```
<host>:8000/setting/update-token
```
**Note**:
- all the fields has to be provide, since this API will overwrite the old token with new token. Any missing field will
be replaced by default value. 
- If the token update change active status of the token, the field last_activation_change will be update with the current time point.
eg

```
curl -X "POST" "http://localhost:8000/setting/update-token" \
     -H 'Content-Type: application/x-www-form-urlencoded'\
     --data-urlencode "data={
          \"id\": \"NEO\",
          \"name\": \"Request\",
          \"decimals\": 18,
          \"address\": \"0x8f8221afbb33998d8584a2b05749ba73c37a938a\",
          \"internal\": false,
          \"active\": true
      }"
```

response

```
on success:
{"success":true}
on failure:
{"success":false,
 "reason":<error>}
```

##### Set token listing - (signing required) Prepare token listing and store the request as pending
POST request 
Post form: {"data" : "JSON enconding of token listing Object"}
```
<host>:8000/setting/set-token-listing
```
**Note**: 
- This data is in the form of a map tokenID:TokenListing which allows mutiple token listings at once
- It also allows mutiple requests, for example, one request listing OMG, the other listing KNC. Both these 
requests will be aggregate in to a list of token to be listed. These can be overwritten as well : if there 
are two requests listing KNC, the later will overwite the ealier.  
- If a token is marked as internal, it will be required to come with exchange setting( fee, min deposit, 
exchange precision limit, deposit address) , and metric settings (pwis, targetQty). Since rebalance quadratic
data can be zero value, it is optional. 
- If exchange precision limit (TokenListing.Exchange.Info) is null, It can be queried from exchange and 
set automatically for the pair (token-ETH). If this data is available in the request,
it will be prioritize over the exchange queried data.
- In addition, if the listing contain any Internal token, that token must be available in Smart contract
in order to update its indices. 
- For token listing, its active status will be automatically set to true.

Example: This request will list token OMG and NEO. OMG is internal, NEO is external. 

``` 
curl -X "POST" "http://localhost:8000/setting/set-token-listing" \
     -H 'Content-Type: application/x-www-form-urlencoded'\
     --data-urlencode "data={  
      \"OMG\": {
        \"token\": {
          \"id\": \"OMG\",
          \"name\": \"OmisexGO\",
          \"decimals\": 18,
          \"address\": \"0xd26114cd6EE289AccF82350c8d8487fedB8A0C07\",
          \"internal\": true,
          \"active\": true
        },
        \"exchanges\": {
          \"binance\": {
            \"deposit_address\": \"0x22222222222222222222222222222222222\",
            \"fee\": {
              \"withdraw\": 0.2,
              \"deposit\": 0.3
            },
            \"min_deposit\": 4
          }
        },
        \"pwis_equation\": {
          \"ask\": {
            \"a\": 800,
            \"b\": 600,
            \"c\": 0,
            \"min_min_spread\": 0,
            \"price_multiply_factor\": 0
          },
          \"bid\": {
            \"a\": 750,
            \"b\": 500,
            \"c\": 0,
            \"min_min_spread\": 0,
            \"price_multiply_factor\": 0
          }
        },
        \"target_qty\": {
          \"set_target\": {
            \"total_target\": 0,
            \"reserve_target\": 0,
            \"rebalance_threshold\": 0,
            \"transfer_threshold\": 0
          }
        },
        \"rebalance_quadratic\": {
          \"rebalance_quadratic\": {
            \"a\": 1,
            \"b\": 2,
            \"c\": 3
          }
        }
      },
      \"NEO\": {
        \"Token\": {
          \"id\": \"NEO\",
          \"name\": \"Request\",
          \"decimals\": 18,
          \"address\": \"0x8f8221afbb33998d8584a2b05749ba73c37a938a\",
          \"internal\": false,
          \"active\": true
        }
      }
    }"

```
response

```
on success:
{"success":true}
on failure:
{"success":false,
 "reason":<error>}
```

##### Get pending token listing - (singing required) Return the current pending token listings information
GET request

``` 
<host>:8000/setting/pending-token-listing
```

Example
``` 
curl -X "GET" "http://localhost:8000/setting/pending-token-listing"
```

response 
```
{data : {
  "AST": {
        "token": {
          "id": "AST",
          "name": "AirSwap",
          "address": "0x27054b13b1b798b345b591a4d22e6562d47ea75a",
          "decimals": 18,
          "active": True,
          "internal": False
        },
        "exchanges": {
          "binance": {
            "deposit_address": "0x111111111111111111111111",
            "PrecisionLimit": {
              "AST-ETH": {
                "Precision": {
                  "Amount": 0,
                  "Price": 7
                },
                "AmountLimit": {
                  "Min": 1,
                  "Max": 90000000
                },
                "PriceLimit": {
                  "Min": 1e-7,
                  "Max": 100000
                },
                "MinNotional": 0.01
              }
            },
            "Fee": {
              "WithDraw": 3,
              "Deposit": 4
            },
            "MinDeposit": 15.04
          }
        }
    }
  }
  "success" : {
    true
  }
```

##### Confirm token listing - (signing required) Confirm token listing and apply all the change to core.
POST request 
Post form: {"data" : "JSON enconding of token listing Object"}
Note: This data is similar to token Listing, but all field must be the same as the current pending. 
```
<host>:8000/setting/confirm-token-listing
```

Example 

``` 
curl -X "POST" "http://localhost:8000/setting/confirm-token-listing" \
     -H 'Content-Type: application/x-www-form-urlencoded'\
     --data-urlencode "data={    
        \"NEO\": {
          \"token\": {
            \"id\": \"NEO\",
            \"name\": \"Request\",
            \"address\": \"0x8f8221afbb33998d8584a2b05749ba73c37a938a\",
            \"decimals\": 18,
            \"active\": true,
            \"internal\": false
          },
          \"exchanges\": null,
          \"pwis_equation\": null,
          \"target_qty\": {
            \"set_target\": {
              \"total_target\": 0,
              \"reserve_target\": 0,
              \"rebalance_threshold\": 0,
              \"transfer_threshold\": 0
            }
          },
          \"rebalance_quadratic\": {
            \"rebalance_quadratic\": {
              \"a\": 0,
              \"b\": 0,
              \"c\": 0
            }
          }
        },
        \"OMG\": {
          \"token\": {
            \"id\": \"OMG\",
            \"name\": \"OmisexGO\",
            \"address\": \"0xd26114cd6EE289AccF82350c8d8487fedB8A0C07\",
            \"decimals\": 18,
            \"active\": true,
            \"internal\": true
          },
          \"exchanges\": {
            \"binance\": {
              \"deposit_address\": \"0x22222222222222222222222222222222222\",
              \"exchange_info\": {
                \"OMG-ETH\": {
                  \"precision\": {
                    \"amount\": 2,
                    \"price\": 6
                  },
                  \"amount_limit\": {
                    \"min\": 0.01,
                    \"max\": 90000000
                  },
                  \"price_limit\": {
                    \"min\": 0.000001,
                    \"max\": 100000
                  },
                  \"min_notional\": 0.01
                }
              },
              \"fee\": {
                \"withdraw\": 0.2,
                \"deposit\": 0.3
              },
              \"min_deposit\": 4
            }
          },
          \"pwis_equation\": {
            \"ask\": {
              \"a\": 800,
              \"b\": 600,
              \"c\": 0,
              \"min_min_spread\": 0,
              \"price_multiply_factor\": 0
            },
            \"bid\": {
              \"a\": 750,
              \"b\": 500,
              \"c\": 0,
              \"min_min_spread\": 0,
              \"price_multiply_factor\": 0
            }
          },
          \"target_qty\": {
            \"set_target\": {
              \"total_target\": 0,
              \"reserve_target\": 0,
              \"rebalance_threshold\": 0,
              \"transfer_threshold\": 0
            }
          },
          \"rebalance_quadratic\": {
            \"rebalance_quadratic\": {
              \"a\": 0,
              \"b\": 0,
              \"c\": 0
            }
          }
        }
    }"

```
response

```
on success:
{"success":true}
on failure:
{"success":false,
 "reason":<error>}
```

##### Reject pending token listing - (signing required) reject the listing and remove the current pending listing
POST request

```
<host>:8000/setting/reject-token-listing
```

Example

```
curl -X "POST" "http://localhost:8000/setting/reject-token-listing" \
     -H 'Content-Type: application/x-www-form-urlencoded'
```

response

```
on success:
{"success":true}
on failure:
{"success":false,
 "reason":<error>}
```

##### Get Token settings - (signing required) get current token settings of core.
GET request

``` 
<host>:8000/setting/token-settings
```

Example
```
curl -X "GET" "http://localhost:8000/setting/token-settings"
```

response 
```
{
  "data": [
    {
      "id": "ABT",
      "name": "",
      "address": "0xb98d4c97425d9908e66e53a6fdf673acca0be986",
      "decimals": 18,
      "active": true,
      "internal": true
    }
  ],
  "success": true
}
```
#### Address related APIs

##### Update address - (signing required) update a single address
POST request 
Post form: {"name" : "Name of the address (reserve, deposit etc...)",
            "address" : "Hex form of the new address"}
Note: This is used to update single address object. For list of address object, use add-address-to-set instead
```
<host>:8000/setting/update-address
```

Example 

```
curl -X "POST" "http://localhost:8000/setting/update-address" \
     -H 'Content-Type: application/x-www-form-urlencoded'\
     --data-urlencode "name=bank"\
     --data-urlencode "address=0x123456789aabbcceeeddff" 

```
response

```
on success:
{"success":true}
on failure:
{"success":false,
 "reason":<error>}
```

##### Add address to set- (signing required) add address to a list of address
POST request 
Post form: {"name" : <Name of the address set(oldBurners etc...)>,
            "address" : <Hex form of the new address>}
```
<host>:8000/setting/add-address-to-set
```

Example 

```
curl -X "POST" "http://localhost:8000/setting/add-address-to-set" \
     -H 'Content-Type: application/x-www-form-urlencoded'\
     --data-urlencode "name="third_party_reserves"\
     --data-urlencode "address=0x123456789aabbcceeeddff" 

```
response

```
on success:
{"success":true}
on failure:
{"success":false,
 "reason":<error>}
```

#### Exchange related APIs

##### Update exchange fee - (signing required) update one exchange fee setting
POST request 
Post form: {"name" : <Name of the exchange (binance, huobi etc...)>,
            "data" : <JSON encoded form of fee setting >}
**Note**: 
UpdateFee will merge the new fee setting to the current fee setting,
Any different key will be overwriten from new fee to current fee. This allows update
one single token's exchange fee on a destined exchange.
UpdateFee will not be mutiplied by any value, so please prepare a big enough number to avoid exchange's fee increasing.
```
<host>:8000/setting/update-exchange-fee
```

Example 

```
  curl -X "POST" "http://localhost:8000/setting/update-exchange-fee" \
     -H 'Content-Type: application/x-www-form-urlencoded'\
     --data-urlencode "name=binance"\
     --data-urlencode "data= {
      \"Trading\": {
        \"maker\": 0.001,
        \"taker\": 0.001
      },
      \"Funding\": {
        \"Withdraw\": {
          \"ZEC\": 0.005,
          \"ZIL\": 100,
          \"ZRX\": 5.8
        },
        \"Deposit\": {
          \"ZEC\": 0,
          \"ZIL\": 0,
          \"ZRX\": 2
        }
      }
    }"
```
##### Update exchange mindeposit - (signing required) update one exchange min deposit
POST request 
Post form: {"name" : <Name of the exchange (binance, huobi etc...)>,
            "data" : <JSON encoded form of min deposit>}
**Note**: 
Update Exchange minDeposit will merge the new minDeposit setting to the current minDeposit setting,
Any different key will be overwriten from new minDeposit to current minDeposit. This allows update
one single token's exchange minDeposit on a destined exchange.
minDeposit input will not be mutiplied by any value, so please prepare a big enough number to avoid exchange's minDeposit increasing.
```
<host>:8000/setting/update-exchange-mindeposit
```

Example 

```
  curl -X "POST" "http://localhost:8000/setting/update-exchange-mindeposit" \
     -H 'Content-Type: application/x-www-form-urlencoded'\
     --data-urlencode "name=binance"\
     --data-urlencode "data= {
      \"POWR\": 0.1,
      \"MANA\": 0.2
    }"
```

#####  Update exchange deposit address - (signing required) update one exchange deposit address
POST request 
Post form: {"name" : <Name of the exchange (binance, huobi etc...)>,
            "data" : <JSON encoded form of a map of token : depositaddress >}
**Note**: 
Update Exchange deposit address will merge the new deposit address setting to the current deposit address setting,
Any different key will be overwriten from new deposit address to current deposit address. This allows update
one single tokenpair's exchange precision limit on a destined exchange.
```
<host>:8000/setting/update-deposit-address
```

Example 

```
  curl -X "POST" "http://localhost:8000/setting/update-deposit-address" \
     -H 'Content-Type: application/x-www-form-urlencoded'\
     --data-urlencode "name=binance"\
     --data-urlencode "data= {
      \"POWR\": \"0x778599Dd7893C8166D313F0F9B5F6cbF7536c293\"
    }"
```

#####  Update exchange info - (signing required) update one exchange's info
POST request 
Post form: {"name" : <Name of the exchange (binance, huobi etc...)>,
            "data" : <JSON encoded form of exchange info >}
**Note**: 
Update Exchange minDeposit will merge the new exchange info setting to the current exchange info setting,
Any different key will be overwriten from new exchange info to current exchange info. This allows update
one single token's exchange minDeposit on a destined exchange.
```
<host>:8000/setting/update-exchange-info
```

Example 

```
  curl -X "POST" "http://localhost:8000/setting/update-exchange-info" \
     -H 'Content-Type: application/x-www-form-urlencoded'\
     --data-urlencode "name=binance"\
     --data-urlencode "data= {
      \"LINK-ETH\": {
        \"precision\": {
          \"amount\": 0,
          \"price\": 8
        },
        \"amount_limit\": {
          \"min\": 1,
          \"max\": 90000000
        },
        \"price_limit\": {
          \"min\": 1e-8,
          \"max\": 120000
        },
        \"min_notional\": 0.01
      }
    }"
```

##### Get all settings - (signing required) return all current running setting of core
GET request

``` 
<host>:8000/setting/all-settings
```

Example
```
curl -X "GET" "http://localhost:8000/setting/all-settings"
```

Response
```
{
  "data": {
    "Addresses": {
      "bank": "",
      "burner": "0x07f6e905f2a1559cd9fd43cb92f8a1062a3ca706",
      "network": "0x964f35fae36d75b1e72770e244f6595b68508cf5",
      "old_burners": [
        "0x4e89bc8484b2c454f2f7b25b612b648c45e14a8e"
      ],
      "pricing": "0x798abda6cc246d0edba912092a2a3dbd3d11191b",
      "reserve": "0x63825c174ab367968ec60f061753d3bbd36a0d8f",
      "setrate": "",
      "third_party_reserves": [
        "0x2aab2b157a03915c8a73adae735d0cf51c872f31",
        "0x4d864b5b4f866f65f53cbaad32eb9574760865e6",
        "0x6f50e41885fdc44dbdf7797df0393779a9c0a3a6"
      ],
      "whitelist": "0x6e106a75d369d09a9ea1dcc16da844792aa669a3",
      "wrapper": "0x6172afc8c00c46e0d07ce3af203828198194620a"
    },
    "Tokens": [
      {
        "id": "ABT",
        "name": "",
        "address": "0xb98d4c97425d9908e66e53a6fdf673acca0be986",
        "decimals": 18,
        "active": true,
        "internal": true
      },
      {
        "id": "ZIL",
        "name": "",
        "address": "0x05f4a42e251f2d52b8ed15e9fedaacfcef1fad27",
        "decimals": 12,
        "active": true,
        "internal": true
      }
    ],
    "Exchanges": {
      "binance": {
        "echange_addresses": {
          "AE": "0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90",
          "ZIL": "0x44d34a119ba21a42167ff8b77a88f0fc7bb2db90"
        },
        "min_deposit": {
          "YOYO": 0,
          "ZEC": 0,
          "ZIL": 0,
          "ZRX": 0
        },
        "fee": {
          "Trading": {
            "maker": 0.001,
            "taker": 0.001
          },
          "Funding": {
            "Withdraw": {
              "ZEC": 0.005,
              "ZIL": 100,
              "ZRX": 5.8
            },
            "Deposit": {
              "ZIL": 0,
              "ZRX": 2
            }
          }
        },
        "info": {
          "AE-ETH": {
            "precision": {
              "amount": 2,
              "price": 6
            },
            "amount_limit": {
              "min": 0.01,
              "max": 90000000
            },
            "price_limit": {
              "min": 0.000001,
              "max": 100000
            },
            "min_notional": 0.01
          },
          "AION-ETH": {
            "precision": {
              "amount": 2,
              "price": 6
            },
            "amount_limit": {
              "min": 0.01,
              "max": 90000000
            },
            "price_limit": {
              "min": 0.000001,
              "max": 100000
            },
            "min_notional": 0.01
          }
        }
      },
      "huobi": {
        "deposit_address": {
          "ABT": "0x0c8fd73eaf6089ef1b91231d0a07d0d2ca2b9d66",
          "CVC": "0x0c8fd73eaf6089ef1b91231d0a07d0d2ca2b9d66",
          "EDU": "0x0c8fd73eaf6089ef1b91231d0a07d0d2ca2b9d66",
         
        },
        "min_deposit": {
          "ABT": 2,
          "APPC": 0.5,
          "AST": 5,
          "SNT": 50,
          "ZIL": 100
        },
        "fee": {
          "Trading": {
            "maker": 0.002,
            "taker": 0.002
          },
          "Funding": {
            "Withdraw": {
              "ABT": 2,
              "ZRX": 5
            },
            "Deposit": {
              "ABT": 0,
              "ZRX": 0
            }
          }
        },
        "info": {
          "POLY-ETH": {
            "precision": {
              "amount": 4,
              "price": 6
            },
            "amount_limit": {
              "min": 0,
              "max": 0
            },
            "price_limit": {
              "min": 0,
              "max": 0
            },
            "min_notional": 0.02
          }
        }
      }
    }
  },
  "success": true
}
```
## Authentication
All APIs that are marked with (signing required) must follow authentication mechanism below:

1. Must be urlencoded (x-www-form-urlencoded)
1. Must have `signed` header with value equals to `hmac512(secret, message)`
1. Must contain `nonce` param, its value is the unix time in millisecond, it must not be before or after server time by 10s
1. `message` is constructed in following way: all query params (nonce is included) and body key-values are merged into one urlencoded string with keys are sorted.
1. `secret` is configured secret string.

Example:
- param query: `amount=0xde0b6b3a7640000&nonce=1514554594528&token=KNC`
- secret: `vtHpz1l0kxLyGc4R1qJBkFlQre5352xGJU9h8UQTwUTz5p6VrxcEslF4KnDI21s1`
- signed string: `2969826a713d13b399dd0d016dad3e95949aa81ed8703ec0258abebb5f0288b96272eef68275f12a32f7e396de3b5fd63ed12b530385e08e1b676c695aacb93b`

## Supported tokens

1. eth (ETH)
2. eos (EOS)
3. kybernetwork (KNC)
4. omisego (OMG)
5. salt (SALT)
6. snt (STATUS)

## Supported exchanges

1. Bittrex (bittrex)
2. Binance (binance)
3. Huobi (huobi)
