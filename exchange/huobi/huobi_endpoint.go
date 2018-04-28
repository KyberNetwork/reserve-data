package huobi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/exchange"
	ethereum "github.com/ethereum/go-ethereum/common"
)

type HuobiEndpoint struct {
	signer Signer
	interf Interface
}

func (self *HuobiEndpoint) fillRequest(req *http.Request, signNeeded bool) {
	if req.Method == "POST" || req.Method == "PUT" || req.Method == "DELETE" {
		req.Header.Add("Content-Type", "application/json")
	} else {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}
	if signNeeded {
		q := req.URL.Query()
		sig := url.Values{}

		method := req.Method
		auth := q.Encode()
		hostname := req.URL.Hostname()
		path := req.URL.Path
		payload := strings.Join([]string{method, hostname, path, auth}, "\n")
		sig.Set("Signature", self.signer.Sign(payload))
		req.URL.RawQuery = q.Encode() + "&" + sig.Encode()
	}
}

func (self *HuobiEndpoint) GetResponse(
	method string, req_url string,
	params map[string]string, signNeeded bool) ([]byte, error) {

	client := &http.Client{
		Timeout: time.Duration(30 * time.Second),
	}
	req_body, _ := json.Marshal(params)
	req, _ := http.NewRequest(method, req_url, nil)
	if method == "POST" {
		req.Body = ioutil.NopCloser(strings.NewReader(string(req_body)))
	}
	req.Header.Add("Accept", "application/json")

	q := req.URL.Query()
	if signNeeded {
		timestamp := fmt.Sprintf("%s", time.Now().Format("2006-01-02T15:04:05"))
		params["SignatureMethod"] = "HmacSHA256"
		params["SignatureVersion"] = "2"
		params["AccessKeyId"] = self.signer.GetKey()
		params["Timestamp"] = timestamp
	}
	var sortedParams []string
	for k, _ := range params {
		sortedParams = append(sortedParams, k)
	}
	sort.Strings(sortedParams)
	for _, k := range sortedParams {
		q.Add(k, params[k])
	}
	req.URL.RawQuery = q.Encode()
	self.fillRequest(req, signNeeded)
	var err error
	var resp_body []byte
	//log.Printf("request to huobi: %s\n", req.URL)
	resp, err := client.Do(req)
	if err != nil {
		return resp_body, err
	} else {
		defer resp.Body.Close()
		resp_body, err = ioutil.ReadAll(resp.Body)
		log.Printf("request to %s, got response from huobi: %s\n", req.URL, common.TruncStr(resp_body))
		return resp_body, err
	}
}

// Get account list for later use
func (self *HuobiEndpoint) GetAccounts() (exchange.HuobiAccounts, error) {
	result := exchange.HuobiAccounts{}
	resp, err := self.GetResponse(
		"GET",
		self.interf.PublicEndpoint()+"/v1/account/accounts",
		map[string]string{},
		true,
	)
	if err == nil {
		json.Unmarshal(resp, &result)
	}
	return result, err
}

func (self *HuobiEndpoint) GetDepthOnePair(
	pair common.TokenPair) (exchange.HuobiDepth, error) {

	resp_body, err := self.GetResponse(
		"GET", self.interf.PublicEndpoint()+"/market/depth",
		map[string]string{
			"symbol": fmt.Sprintf("%s%s", strings.ToLower(pair.Base.ID), strings.ToLower(pair.Quote.ID)),
			"type":   "step0",
		},
		false,
	)

	resp_data := exchange.HuobiDepth{}
	if err != nil {
		return resp_data, err
	} else {
		json.Unmarshal(resp_body, &resp_data)
		return resp_data, nil
	}
}

func (self *HuobiEndpoint) Trade(tradeType string, base, quote common.Token, rate, amount float64, timepoint uint64) (exchange.HuobiTrade, error) {
	result := exchange.HuobiTrade{}
	symbol := strings.ToLower(base.ID) + strings.ToLower(quote.ID)
	orderType := tradeType + "-limit"
	accounts, _ := self.GetAccounts()
	if len(accounts.Data) == 0 {
		return result, errors.New("Cannot get account")
	}
	params := map[string]string{
		"account-id": strconv.FormatUint(accounts.Data[0].ID, 10),
		"symbol":     symbol,
		"source":     "api",
		"type":       orderType,
		"amount":     strconv.FormatFloat(amount, 'f', -1, 64),
		"price":      strconv.FormatFloat(rate, 'f', -1, 64),
	}
	resp_body, err := self.GetResponse(
		"POST",
		self.interf.AuthenticatedEndpoint()+"/v1/order/orders/place",
		params,
		true,
	)
	if err != nil {
		return result, err
	} else {
		json.Unmarshal(resp_body, &result)
		if result.Status != "ok" {
			return result, errors.New(fmt.Sprintf("Create order failed: %s\n", result.Reason))
		}
		return result, nil
	}
}

func (self *HuobiEndpoint) WithdrawHistory() (exchange.HuobiWithdraws, error) {
	result := exchange.HuobiWithdraws{}
	size := len(common.InternalTokens()) * 2
	resp_body, err := self.GetResponse(
		"GET",
		self.interf.AuthenticatedEndpoint()+"/v1/query/finances",
		map[string]string{
			"size": strconv.Itoa(size),
			// "size":  "10",
			"types": "withdraw-virtual",
		},
		true,
	)
	if err == nil {
		json.Unmarshal(resp_body, &result)
		if result.Status != "ok" {
			err = errors.New(result.Reason)
		}
	}
	return result, err
}

func (self *HuobiEndpoint) DepositHistory() (exchange.HuobiDeposits, error) {
	result := exchange.HuobiDeposits{}
	size := len(common.InternalTokens()) * 2
	resp_body, err := self.GetResponse(
		"GET",
		self.interf.AuthenticatedEndpoint()+"/v1/query/finances",
		map[string]string{
			"size": strconv.Itoa(size),
			// "size":  "10",
			"types": "deposit-virtual",
		},
		true,
	)
	if err == nil {
		err = json.Unmarshal(resp_body, &result)
		if result.Status != "ok" {
			err = errors.New(fmt.Sprintf("Getting deposit history from Huobi failed: %s\n", result.Reason))
		}
	}
	return result, err
}

func (self *HuobiEndpoint) CancelOrder(symbol string, id uint64) (exchange.HuobiCancel, error) {
	result := exchange.HuobiCancel{}
	resp_body, err := self.GetResponse(
		"POST",
		self.interf.AuthenticatedEndpoint()+"/v1/order/orders/"+strconv.FormatUint(id, 10)+"/submitcancel",
		map[string]string{
			"order-id": fmt.Sprintf("%d", id),
		},
		true,
	)
	if err == nil {
		json.Unmarshal(resp_body, &result)
		if result.Status != "ok" {
			err = errors.New(fmt.Sprintf("Cancel order failed: %s\n", result.Reason))
		}
	}
	return result, err
}

func (self *HuobiEndpoint) OrderStatus(symbol string, id uint64) (exchange.HuobiOrder, error) {
	result := exchange.HuobiOrder{}
	resp_body, err := self.GetResponse(
		"GET",
		self.interf.AuthenticatedEndpoint()+"/v1/order/orders/"+strconv.FormatUint(id, 10),
		map[string]string{
			"order-id": fmt.Sprintf("%d", id),
		},
		true,
	)
	if err == nil {
		json.Unmarshal(resp_body, &result)
		if result.Status != "ok" {
			err = errors.New(fmt.Sprintf("Get order status failed: %s \n", result.Reason))
		}
	}
	return result, err
}

func (self *HuobiEndpoint) Withdraw(token common.Token, amount *big.Int, address ethereum.Address) (string, error) {
	result := exchange.HuobiWithdraw{}
	resp_body, err := self.GetResponse(
		"POST",
		self.interf.AuthenticatedEndpoint()+"/v1/dw/withdraw/api/create",
		map[string]string{
			"address":  address.Hex(),
			"amount":   strconv.FormatFloat(common.BigToFloat(amount, token.Decimal), 'f', -1, 64),
			"currency": strings.ToLower(token.ID),
		},
		true,
	)
	if err == nil {
		json.Unmarshal(resp_body, &result)
		log.Printf("Response body: %+v\n", result)
		if result.Status != "ok" {
			return "", errors.New(fmt.Sprintf("Withdraw from Huobi failed: %s\n", result.Reason))
		}
		log.Printf("Withdraw id: %s", fmt.Sprintf("%v", result.ID))
		return strconv.FormatUint(result.ID, 10), nil
	} else {
		log.Printf("Error: %v", err)
		return "", errors.New("Withdraw rejected by Huobi")
	}
}

func (self *HuobiEndpoint) GetInfo() (exchange.HuobiInfo, error) {
	result := exchange.HuobiInfo{}
	accounts, _ := self.GetAccounts()
	if len(accounts.Data) == 0 {
		return result, errors.New("Cannot get account")
	}
	resp_body, err := self.GetResponse(
		"GET",
		self.interf.AuthenticatedEndpoint()+"/v1/account/accounts/"+strconv.FormatUint(accounts.Data[0].ID, 10)+"/balance",
		map[string]string{},
		true,
	)
	if err == nil {
		json.Unmarshal(resp_body, &result)
	}
	return result, err
}

func (self *HuobiEndpoint) GetAccountTradeHistory(
	base, quote common.Token) (exchange.HuobiTradeHistory, error) {
	result := exchange.HuobiTradeHistory{}
	symbol := strings.ToUpper(fmt.Sprintf("%s%s", base.ID, quote.ID))
	resp_body, err := self.GetResponse(
		"GET",
		self.interf.AuthenticatedEndpoint()+"/v1/order/orders",
		map[string]string{
			"symbol": strings.ToLower(symbol),
			"states": "filled",
		},
		true,
	)
	if err == nil {
		json.Unmarshal(resp_body, &result)
	}
	return result, err
}

func (self *HuobiEndpoint) OpenOrdersForOnePair(
	pair common.TokenPair) (exchange.HuobiOrder, error) {
	// TODO: check again if use
	result := exchange.HuobiOrder{}
	resp_body, err := self.GetResponse(
		"GET",
		self.interf.AuthenticatedEndpoint()+"/", // TODO: check again if available
		map[string]string{
			"symbol": pair.Base.ID + pair.Quote.ID,
		},
		true,
	)
	if err != nil {
		return result, err
	} else {
		json.Unmarshal(resp_body, &result)
		return result, nil
	}
}

func (self *HuobiEndpoint) GetDepositAddress(asset string) (exchange.HuobiDepositAddress, error) {
	result := exchange.HuobiDepositAddress{}
	resp_body, err := self.GetResponse(
		"GET",
		self.interf.AuthenticatedEndpoint()+"/v1/dw/deposit-virtual/addresses",
		map[string]string{
			"currency": asset,
		},
		true,
	)
	if err == nil {
		err = json.Unmarshal(resp_body, &result)
		if !result.Success {
			err = errors.New(fmt.Sprintf("Get deposit address failed: %s\n", result.Reason))
		}
	}
	return result, err
}

func (self *HuobiEndpoint) GetExchangeInfo() (exchange.HuobiExchangeInfo, error) {
	result := exchange.HuobiExchangeInfo{}
	resp_body, err := self.GetResponse(
		"GET",
		self.interf.PublicEndpoint()+"/v1/common/symbols",
		map[string]string{},
		false,
	)
	if err == nil {
		err = json.Unmarshal(resp_body, &result)
	}
	return result, err
}

func NewHuobiEndpoint(signer Signer, interf Interface) *HuobiEndpoint {
	return &HuobiEndpoint{signer, interf}
}
