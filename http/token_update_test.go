package http

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/KyberNetwork/reserve-data/data/storage"

	reservemock "github.com/KyberNetwork/reserve-data/data/mock"
	_ "github.com/KyberNetwork/reserve-data/exchange/binance/mock"
	"github.com/KyberNetwork/reserve-data/http/httputil"
	"github.com/KyberNetwork/reserve-data/settings"
	settingsstorage "github.com/KyberNetwork/reserve-data/settings/storage"
	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
)

const (
	tokenRequestData = `{
	"KNC": {
		"token": {
			"name": "KyberNetwork Crystal",
			"decimals": 18,
			"address": "0xd26114cd6EE289AccF82350c8d8487fedB8A0C07",
			"minimal_record_resolution": "1000000000000000",
			"max_per_block_imbalance": "439794468212403470336",
			"max_total_imbalance": "722362414038872621056",
			"internal": true,
			"active": true
		},
		"exchanges": {
		"binance": {
			"DepositAddress": "0x22222222222222222222222222222222222",
			"Fee": {
			"Trading": 0.1,
			"WithDraw": 0.2,
			"Deposit": 0.3
			},
			"MinDeposit": 4
		}
		},
		"pwis_equation": {
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
		"target_qty": {
		"set_target": {
			"total_target": 1,
			"reserve_target": 2,
			"rebalance_threshold": 0,
			"transfer_threshold": 0
		}
		},
		"rebalance_quadratic": {
		"rebalance_quadratic": {
			"a": 1,
			"b": 2,
			"c": 4
		}
		}
	},
	"NEO": {
		"token": {
			"id": "NEO",
			"name": "Request",
			"decimals": 18,
			"address": "0x8f8221afbb33998d8584a2b05749ba73c37a938a",
			"minimalRecordResolution": "1000000000000000",
			"maxPerBlockImbalance": "27470469074054960644096",
			"maxTotalImbalance": "33088179999699195920384",
			"internal": false,
			"active": true          
		}
		}
	}`
	incorrectTokenRequestData = `{
		"OMG": {
			"token": {
				"name": "OmiseGo",
				"decimals": 18,
				"address": "0xd26114cd6EE289AccF82350c8d8487fedB8A0C07",
				"minimal_record_resolution": "1000000000000000",
				"max_per_block_imbalance": "439794468212403470336",
				"max_total_imbalance": "722362414038872621056",
				"internal": true,
				"active": true
			},
			"exchanges": {
			"binance": {
				"DepositAddress": "0x22222222222222222222222222222222222",
				"Fee": {
				"Trading": 0.1,
				"		{
			msg:      "Update Exchange Fee on an unsupported exchange",
			endpoint: updateExchangeFeeEndpoint,
			method:   http.MethodPost,
			data: map[string]string{
				"name": "ex",
				"data": feeRequest,
			},
			assert: httputil.ExpectFailure,
		},
		{
			msg:      "Update Exchange Fee on a supported exchange with wrong data format",
			endpoint: updateExchangeFeeEndpoint,
			method:   http.MethodPost,
			data: map[string]string{
				"name": "binance",
				"data": minDepositRequest,
			},
			assert: httputil.ExpectFailure,
		},
		{
			msg:      "Update Exchange Fee on a supported exchange without record in setting DB",
			endpoint: updateExchangeFeeEndpoint,
			method:   http.MethodPost,
			data: map[string]string{
				"name": "binance",
				"data": feeRequest,
			},
			assert: httputil.ExpectSuccess,
		},
		{
			msg:      "Update Exchange Fee on a supported exchange with recording in setting DB",
			endpoint: updateExchangeFeeEndpoint,
			method:   http.MethodPost,
			data: map[string]string{
				"name": "binance",
				"data": feeRequest,
			},
			assert: httputil.ExpectSuccess,
		},": 0.2,
				"Deposit": 0.3
				},
				"MinDeposit": 4
			}
			},
			"pwis_equation": {
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
			"target_qty": {
			"set_target": {
				"total_target": 1,
				"reserve_target": 2,
				"rebalance_threshold": 0,
				"transfer_threshold": 0
			}
			},
			"rebalance_quadratic": {
			"rebalance_quadratic": {
				"a": 1,
				"b": 2,
				"c": 4
			}
			}
		},
		"NEO": {
			"token": {
				"id": "NEO",
				"name": "Request",
				"decimals": 18,
				"address": "0x8f8221afbb33998d8584a2b05749ba73c37a938a",
				"minimalRecordResolution": "1000000000000000",
				"maxPerBlockImbalance": "27470469074054960644096",
				"maxTotalImbalance": "33088179999699195920384",
				"internal": false,
				"active": true          
			}
			}
		}`
	confirmData = `{
			"KNC": {
			  "token": {
				"id": "KNC",
				"name": "KyberNetwork Crystal",
				"address": "0xd26114cd6EE289AccF82350c8d8487fedB8A0C07",
				"decimals": 18,
				"active": true,
				"internal": true,
				"last_activation_change": 0
			  },
			  "exchanges": {
				"binance": {
				  "deposit_address": "",
				  "exchange_info": {
					"KNC-ETH": {
					  "precision": {
						"amount": 0,
						"price": 7
					  },
					  "amount_limit": {
						"min": 1,
						"max": 900000
					  },
					  "price_limit": {
						"min": 0.000192,
						"max": 0.019195
					  },
					  "min_notional": 0.01
					}
				  },
				  "fee": {
					"withdraw": 0.2,
					"deposit": 0.3
				  },
				  "min_deposit": 0
				}
			  },
			  "pwis_equation": {
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
			  "target_qty": {
				"set_target": {
				  "total_target": 1,
				  "reserve_target": 2,
				  "rebalance_threshold": 0,
				  "transfer_threshold": 0
				}
			  },
			  "rebalance_quadratic": {
				"rebalance_quadratic": {
				  "a": 1,
				  "b": 2,
				  "c": 4
				}
			  }
			},
			"NEO": {
			  "token": {
				"id": "NEO",
				"name": "Request",
				"address": "0x8f8221afbb33998d8584a2b05749ba73c37a938a",
				"decimals": 18,
				"active": true,
				"internal": false,
				"last_activation_change": 0
			  },
			  "exchanges": null,
			  "pwis_equation": null,
			  "target_qty": {
				"set_target": {
				  "total_target": 0,
				  "reserve_target": 0,
				  "rebalance_threshold": 0,
				  "transfer_threshold": 0
				}
			  },
			  "rebalance_quadratic": {
				"rebalance_quadratic": {
				  "a": 0,
				  "b": 0,
				  "c": 0
				}
			  }
			}
		  }`
)

func TestHTTPServerUpdateToken(t *testing.T) {
	const (
		setPendingTokenUpdateEndpoint    = "/setting/set-token-update"
		getPendingTokenUpdateEndpoint    = "/setting/pending-token-update"
		confirmTokenUpdateEndpoint       = "/setting/confirm-token-update"
		rejectPendingTokenUpdateEndpoint = "/setting/reject-token-update"
	)
	tmpDir, err := ioutil.TempDir("", "test_token_APIs")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if rErr := os.RemoveAll(tmpDir); rErr != nil {
			t.Error(rErr)
		}
	}()

	boltSettingStorage, err := settingsstorage.NewBoltSettingStorage(filepath.Join(tmpDir, "setting.db"))
	if err != nil {
		log.Fatal(err)
	}
	tokenSetting, err := settings.NewTokenSetting(boltSettingStorage)
	if err != nil {
		log.Fatal(err)
	}
	addressSetting, err := settings.NewAddressSetting(boltSettingStorage)
	if err != nil {
		log.Fatal(err)
	}
	exchangeSetting, err := settings.NewExchangeSetting(boltSettingStorage)
	if err != nil {
		log.Fatal(err)
	}

	setting, err := settings.NewSetting(tokenSetting, addressSetting, exchangeSetting)
	if err != nil {
		log.Fatal(err)
	}

	testStorage, err := storage.NewBoltStorage(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}

	testServer := HTTPServer{
		app:         reservemock.TestReserveData{},
		core:        nil,
		metric:      testStorage,
		authEnabled: false,
		r:           gin.Default(),
		blockchain:  testHTTPBlockchain{},
		setting:     setting,
	}
	testServer.register()

	var tests = []testCase{
		{
			msg:      "invalid post form",
			endpoint: setPendingTokenUpdateEndpoint,
			method:   http.MethodPost,
			data: map[string]string{
				"invalid_key": "invalid_val",
			},
			assert: httputil.ExpectFailure,
		},
		{
			msg:      "set token update incorrectly",
			endpoint: setPendingTokenUpdateEndpoint,
			method:   http.MethodPost,
			data: map[string]string{
				"data": incorrectTokenRequestData,
			},
			assert: httputil.ExpectFailure,
		},
		{
			msg:      "reject pending token update when there is no pending token update",
			endpoint: rejectPendingTokenUpdateEndpoint,
			method:   http.MethodPost,
			data: map[string]string{
				"value": "some random post form or this request will be unauthenticated",
			},
			assert: httputil.ExpectFailure,
		},
		{
			msg:      "set token update correctly",
			endpoint: setPendingTokenUpdateEndpoint,
			method:   http.MethodPost,
			data: map[string]string{
				"data": tokenRequestData,
			},
			assert: httputil.ExpectSuccess,
		},
		{
			msg:      "get pending token update",
			endpoint: getPendingTokenUpdateEndpoint,
			method:   http.MethodGet,
			data:     map[string]string{},
			assert:   httputil.ExpectSuccess,
		},
		{
			msg:      "reject pending token update",
			endpoint: rejectPendingTokenUpdateEndpoint,
			method:   http.MethodPost,
			data: map[string]string{
				"value": "some random post form or this request will be unauthenticated",
			},
			assert: httputil.ExpectSuccess,
		},
		{
			msg:      "set token update correctly for confirmation",
			endpoint: setPendingTokenUpdateEndpoint,
			method:   http.MethodPost,
			data: map[string]string{
				"data": tokenRequestData,
			},
			assert: httputil.ExpectSuccess,
		},
		{
			msg:      "confirmation with incorrect data",
			endpoint: confirmTokenUpdateEndpoint,
			method:   http.MethodPost,
			data: map[string]string{
				"data": incorrectTokenRequestData,
			},
			assert: httputil.ExpectFailure,
		},
		{
			msg:      "confirmation with correct data although exchange setting isn't available in DB",
			endpoint: confirmTokenUpdateEndpoint,
			method:   http.MethodPost,
			data: map[string]string{
				"data": confirmData,
			},
			assert: httputil.ExpectSuccess,
		},
	}
	for _, tc := range tests {
		t.Run(tc.msg, func(t *testing.T) { testHTTPRequest(t, tc, testServer.r) })
	}

}

type testHTTPBlockchain struct{}

func (tbc testHTTPBlockchain) CheckTokenIndices(addr ethereum.Address) error {
	const correctAddrstr = "0xd26114cd6EE289AccF82350c8d8487fedB8A0C07"
	correctAddr := ethereum.HexToAddress(correctAddrstr)
	if addr.Hex() == correctAddr.Hex() {
		return nil
	}
	return errors.New("wrong address")
}

func (tbc testHTTPBlockchain) LoadAndSetTokenIndices(addrs []ethereum.Address) error {
	return nil
}

func (tbc testHTTPBlockchain) GetPricingOPAddress() ethereum.Address {
	return ethereum.Address{}
}

func (tbc testHTTPBlockchain) GetDepositOPAddress() ethereum.Address {
	return ethereum.Address{}
}

func (tbc testHTTPBlockchain) GetIntermediatorOPAddress() ethereum.Address {
	return ethereum.Address{}
}
