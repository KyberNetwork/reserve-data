package exchange

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/KyberNetwork/reserve-data/common"
	ethereum "github.com/ethereum/go-ethereum/common"
	"sync"
)

const bittrexApiVersion string = "v1.1"

type RealBittrexEndpoint struct {
	PublicEndpoint  string
	MarketEndpoint  string
	AccountEndpoint string
}

type bittrexResponse struct {
	Success bool                            `json:"success"`
	Msg     string                          `json:"message"`
	Result  map[string][]map[string]float64 `json:"result"`
}

func (self *RealBittrexEndpoint) nonce() string {
	epsilon := 30 * time.Millisecond
	anchor := int64(50299954901)
	timestamp := time.Now().UnixNano()/int64(epsilon) - anchor
	return strconv.Itoa(int(timestamp))
}

// sign the request if signer passed
func (self *RealBittrexEndpoint) fillRequest(req *http.Request, signer Signer) {
	if req.Method == "POST" || req.Method == "PUT" {
		req.Header.Add("Content-Type", "application/json;charset=utf-8")
	}
	req.Header.Add("Accept", "application/json")
	if signer != nil {
		q := req.URL.Query()
		q.Set("apiKey", signer.GetBittrexKey())
		q.Set("nonce", self.nonce())
		req.URL.RawQuery = q.Encode()
		req.Header.Add("apisign", signer.BittrexSign(req.URL.String()))
	}
}

func (self *RealBittrexEndpoint) FetchOnePairData(wq *sync.WaitGroup, pair common.TokenPair, data *sync.Map) {
	defer wq.Done()

	result := common.ExchangePrice{}

	client := &http.Client{}
	req, _ := http.NewRequest("GET", self.PublicEndpoint+"/getorderbook", nil)
	q := req.URL.Query()
	q.Set("market", fmt.Sprintf("%s-%s", pair.Quote.ID, pair.Base.ID))
	q.Set("type", "both")
	req.URL.RawQuery = q.Encode()
	self.fillRequest(req, nil)
	res, err := client.Do(req)

	result.Timestamp = common.GetTimestamp()
	result.Valid = true
	if err != nil {
		result.Valid = false
		result.Error = err.Error()
	} else {
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		result.ReturnTime = common.GetTimestamp()
		if err != nil {
			result.Valid = false
			result.Error = err.Error()
		} else {
			data := bittrexResponse{}
			json.Unmarshal(body, &data)
			if !data.Success {
				result.Valid = false
			} else {
				for _, buy := range data.Result["buy"] {
					result.BuyPrices = append(
						result.BuyPrices,
						common.PriceEntry{
							buy["Quantity"],
							buy["Rate"],
						},
					)
				}
				for _, sell := range data.Result["sell"] {
					result.SellPrices = append(
						result.SellPrices,
						common.PriceEntry{
							sell["Quantity"],
							sell["Rate"],
						},
					)
				}
			}
		}
	}
	data.Store(pair.PairID(), result)
}

type bittrexTradeResult struct {
	Done      float64 `json:"received"`
	Remaining float64 `json:"remains"`
	OrderID   int64   `json:"uuid"`
}

type bitrexTrade struct {
	Success int                `json:"success"`
	Return  bittrexTradeResult `json:"result"`
	Error   string             `json:"message"`
}

func (self *RealBittrexEndpoint) Trade(key string, tradeType string, base, quote common.Token, rate, amount float64, signer Signer) (done float64, remaining float64, finished bool, err error) {
	result := bitrexTrade{}
	client := &http.Client{
		Timeout: time.Duration(30 * time.Second)}
	req, _ := http.NewRequest(
		"GET",
		self.MarketEndpoint+"/selllimit",
		nil,
	)
	q := req.URL.Query()
	q.Set("market", fmt.Sprintf("%s-%s", strings.ToUpper(base.ID), strings.ToUpper(quote.ID)))
	q.Set("quantity", fmt.Sprintf("%f", amount))
	q.Set("rate", fmt.Sprintf("%f", rate))
	req.URL.RawQuery = q.Encode()
	self.fillRequest(req, signer)
	resp, err := client.Do(req)
	if err == nil && resp.StatusCode == 200 {
		defer resp.Body.Close()
		resp_body, err := ioutil.ReadAll(resp.Body)
		fmt.Printf("response: %s\n", resp_body)
		if err == nil {
			err = json.Unmarshal(resp_body, &result)
		}
		if err != nil {
			return 0, 0, false, err
		}
		if result.Error != "" {
			return 0, 0, false, errors.New(result.Error)
		}
		return result.Return.Done, result.Return.Remaining, result.Return.OrderID == 0, nil
	} else {
		fmt.Printf("Error: %v, Code: %v\n", err, resp)
		return 0, 0, false, errors.New("Trade rejected by Bittrex")
	}
}

type bittrexWithdraw struct {
	Success int    `json:"success"`
	Return  map[string]interface{}
	Error   string `json:"message"`
}

func (self *RealBittrexEndpoint) Withdraw(key string, token common.Token, amount *big.Int, address ethereum.Address, signer Signer) error {
	result := bittrexWithdraw{}
	client := &http.Client{
		Timeout: time.Duration(30 * time.Second),
	}
	req, _ := http.NewRequest(
		"GET",
		self.AccountEndpoint+"/withdraw",
		nil,
	)
	q := req.URL.Query()
	q.Set("currency", strings.ToUpper(token.ID))
	q.Set("quantity", fmt.Sprintf("%f", common.BigToFloat(amount, token.Decimal)))
	q.Set("address", address.Hex())
	req.URL.RawQuery = q.Encode()
	self.fillRequest(req, signer)
	resp, err := client.Do(req)
	if err == nil && resp.StatusCode == 200 {
		defer resp.Body.Close()
		resp_body, err := ioutil.ReadAll(resp.Body)
		fmt.Printf("response: %s\n", resp_body)
		if err == nil {
			err = json.Unmarshal(resp_body, &result)
		}
		if err != nil {
			return err
		}
		if result.Error != "" {
			return errors.New(result.Error)
		}
		return nil
	} else {
		fmt.Printf("Error: %v, Code: %v\n", err, resp)
		return errors.New("withdraw rejected by Bittrex")
	}
}

type bittrexInfo map[string]map[string][][]float64

func (self *RealBittrexEndpoint) GetInfo(key string, signer Signer) (bittrexInfo, error) {
	result := bittrexInfo{}
	client := &http.Client{
		Timeout: time.Duration(30 * time.Second)}
	data := url.Values{}
	data.Set("method", "getInfo")
	data.Add("nonce", nonce())
	params := data.Encode()
	req, _ := http.NewRequest(
		"POST",
		self.AccountEndpoint,
		bytes.NewBufferString(params),
	)
	req.Header.Add("Content-Length", strconv.Itoa(len(params)))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Key", key)
	req.Header.Add("Sign", signer.LiquiSign(params))
	resp, err := client.Do(req)
	if err == nil {
		defer resp.Body.Close()
		resp_body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			json.Unmarshal(resp_body, &result)
		}
	}
	return result, err
}

func NewRealBittrexEndpoint() *RealBittrexEndpoint {
	return &RealBittrexEndpoint{
		"https://bittrex.com/api/" + bittrexApiVersion + "/public",
		"https://bittrex.com/api/" + bittrexApiVersion + "/market",
		"https://bittrex.com/api/" + bittrexApiVersion + "/account",
	}
}
