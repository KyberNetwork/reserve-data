package configuration

import (
	"github.com/KyberNetwork/reserve-data/common"
)

//AddressConfigs store token configs according to env mode.
var AddressConfigs = map[string]string{
	common.DevMode: `
{
  "reserve": "0x63825c174ab367968EC60f061753D3bbD36A0D8F",
  "network": "0x818E6FECD516Ecc3849DAf6845e3EC868087B755",
  "wrapper": "0x6172AFC8c00c46E0D07ce3AF203828198194620a",
  "pricing": "0x798AbDA6Cc246D0EDbA912092A2a3dBd3d11191B",
  "feeburner": "0xed4f53268bfdFF39B36E8786247bA3A02Cf34B04",
  "whitelist": "0x6e106a75d369d09a9ea1dcc16da844792aa669a3",
  "third_party_reserves": [
    "0x2aab2b157a03915c8a73adae735d0cf51c872f31"
  ],
  "internal network": "0x91a502C678605fbCe581eae053319747482276b9"
}
`,
}
