package huobi

import (
	"bytes"
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

const (
	huobiDepositEndpoint = "https://www.hbg.com/-/x/pro/"
)

//Endpoint endpoint object
type Endpoint struct {
	signer Signer
	interf Interface
}

func (ep *Endpoint) fillRequest(req *http.Request, signNeeded bool) {
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
		sig.Set("Signature", ep.signer.Sign(payload))
		req.URL.RawQuery = q.Encode() + "&" + sig.Encode()
	}
}

func (ep *Endpoint) getResponse(
	method string, reqURL string,
	params map[string]string, signNeeded bool) ([]byte, error) {

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	reqBody, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, reqURL, nil)
	if err != nil {
		return nil, err
	}
	if method == "POST" {
		req.Body = ioutil.NopCloser(bytes.NewReader(reqBody))
	}
	req.Header.Add("Accept", "application/json")

	q := req.URL.Query()
	if signNeeded {
		timestamp := time.Now().Format("2006-01-02T15:04:05")
		params["SignatureMethod"] = "HmacSHA256"
		params["SignatureVersion"] = "2"
		params["AccessKeyId"] = ep.signer.GetKey()
		params["Timestamp"] = timestamp
	}
	var sortedParams []string
	for k := range params {
		sortedParams = append(sortedParams, k)
	}
	sort.Strings(sortedParams)
	for _, k := range sortedParams {
		q.Add(k, params[k])
	}
	req.URL.RawQuery = q.Encode()
	ep.fillRequest(req, signNeeded)
	var respBody []byte
	//log.Printf("request to huobi: %s\n", req.URL)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	respBody, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}
	if err = resp.Body.Close(); err != nil {
		log.Printf("Response body close error: %s", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return respBody, err
	case http.StatusTooManyRequests:
		err = errors.New("breaking Huobi request rate limit")
	case http.StatusInternalServerError:
		err = errors.New("500 from Huobi, its fault")
	default: // response body sometime has detail information, we collect it in unexpected case.
		err = fmt.Errorf("unexpected error, status %d, %s", resp.StatusCode, string(respBody))
	}
	return nil, err
}

//GetAccounts Get account list for later use
func (ep *Endpoint) GetAccounts() (exchange.HuobiAccounts, error) {
	result := exchange.HuobiAccounts{}
	resp, err := ep.getResponse(
		"GET",
		ep.interf.PublicEndpoint()+"/v1/account/accounts",
		map[string]string{},
		true,
	)
	if err == nil {
		err = json.Unmarshal(resp, &result)
	}
	return result, err
}

func (ep *Endpoint) GetDepthOnePair(
	baseID, quoteID string) (exchange.HuobiDepth, error) {

	respBody, err := ep.getResponse(
		"GET", ep.interf.PublicEndpoint()+"/market/depth",
		map[string]string{
			"symbol": fmt.Sprintf("%s%s", strings.ToLower(baseID), strings.ToLower(quoteID)),
			"type":   "step0",
		},
		false,
	)

	respData := exchange.HuobiDepth{}
	if err != nil {
		return respData, err
	}
	err = json.Unmarshal(respBody, &respData)
	return respData, err
}

func (ep *Endpoint) Trade(tradeType string, base, quote common.Token, rate, amount float64, timepoint uint64) (exchange.HuobiTrade, error) {
	result := exchange.HuobiTrade{}
	symbol := strings.ToLower(base.ID) + strings.ToLower(quote.ID)
	orderType := tradeType + "-limit"
	accounts, err := ep.GetAccounts()
	if err != nil {
		return result, err
	}
	if len(accounts.Data) == 0 {
		return result, errors.New("cannot get Huobi account")
	}
	params := map[string]string{
		"account-id": strconv.FormatUint(accounts.Data[0].ID, 10),
		"symbol":     symbol,
		"source":     "api",
		"type":       orderType,
		"amount":     strconv.FormatFloat(amount, 'f', -1, 64),
		"price":      strconv.FormatFloat(rate, 'f', -1, 64),
	}
	respBody, err := ep.getResponse(
		"POST",
		ep.interf.AuthenticatedEndpoint()+"/v1/order/orders/place",
		params,
		true,
	)
	if err != nil {
		return result, err
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return result, err
	}
	if result.Status != "ok" {
		return result, fmt.Errorf("create Huobi order failed: %s", result.Reason)
	}
	return result, nil
}

func (ep *Endpoint) WithdrawHistory(tokens []common.Token) (exchange.HuobiWithdraws, error) {
	result := exchange.HuobiWithdraws{}
	size := len(tokens) * 2
	respBody, err := ep.getResponse(
		"GET",
		ep.interf.AuthenticatedEndpoint()+"/v1/query/finances",
		map[string]string{
			"size": strconv.Itoa(size),
			// "size":  "10",
			"types": "withdraw-virtual",
		},
		true,
	)
	if err == nil {
		if err = json.Unmarshal(respBody, &result); err != nil {
			return result, err
		}
		if result.Status != "ok" {
			err = errors.New(result.Reason)
		}
	}
	return result, err
}

func (ep *Endpoint) DepositHistory(tokens []common.Token) (exchange.HuobiDeposits, error) {
	result := exchange.HuobiDeposits{}
	size := len(tokens) * 2
	respBody, err := ep.getResponse(
		"GET",
		ep.interf.AuthenticatedEndpoint()+"/v1/query/finances",
		map[string]string{
			"size": strconv.Itoa(size),
			// "size":  "10",
			"types": "deposit-virtual",
		},
		true,
	)
	if err == nil {
		if err = json.Unmarshal(respBody, &result); err != nil {
			return result, err
		}
		if result.Status != "ok" {
			err = fmt.Errorf("getting deposit history from Huobi failed: %s", result.Reason)
		}
	}
	return result, err
}

func (ep *Endpoint) CancelOrder(symbol string, id uint64) (exchange.HuobiCancel, error) {
	result := exchange.HuobiCancel{}
	respBody, err := ep.getResponse(
		"POST",
		ep.interf.AuthenticatedEndpoint()+"/v1/order/orders/"+strconv.FormatUint(id, 10)+"/submitcancel",
		map[string]string{
			"order-id": fmt.Sprintf("%d", id),
		},
		true,
	)
	if err != nil {
		return result, err
	}
	if err = json.Unmarshal(respBody, &result); err != nil {
		return result, err
	}
	if result.Status != "ok" {
		err = fmt.Errorf("cancel Huobi order failed: %s", result.Reason)
	}
	return result, err
}

func (ep *Endpoint) OrderStatus(symbol string, id uint64) (exchange.HuobiOrder, error) {
	result := exchange.HuobiOrder{}
	respBody, err := ep.getResponse(
		"GET",
		ep.interf.AuthenticatedEndpoint()+"/v1/order/orders/"+strconv.FormatUint(id, 10),
		map[string]string{
			"order-id": fmt.Sprintf("%d", id),
		},
		true,
	)
	if err != nil {
		return result, err
	}
	if err = json.Unmarshal(respBody, &result); err != nil {
		return result, err
	}
	if result.Status != "ok" {
		err = fmt.Errorf("get Huobi order status failed: %s", result.Reason)
	}
	return result, err
}

func (ep *Endpoint) Withdraw(token common.Token, amount *big.Int, address ethereum.Address) (string, error) {
	result := exchange.HuobiWithdraw{}
	respBody, err := ep.getResponse(
		"POST",
		ep.interf.AuthenticatedEndpoint()+"/v1/dw/withdraw/api/create",
		map[string]string{
			"address":  address.Hex(),
			"amount":   strconv.FormatFloat(common.BigToFloat(amount, token.Decimals), 'f', -1, 64),
			"currency": strings.ToLower(token.ID),
		},
		true,
	)
	if err == nil {
		if err = json.Unmarshal(respBody, &result); err != nil {
			return "", err
		}
		if result.Status != "ok" {
			return "", fmt.Errorf("withdraw from Huobi failed: %s", result.Reason)
		}
		log.Printf("Withdraw id: %s", fmt.Sprintf("%v", result.ID))
		return strconv.FormatUint(result.ID, 10), nil
	}
	return "", errors.New("Withdraw rejected by Huobi")
}

func (ep *Endpoint) GetInfo() (exchange.HuobiInfo, error) {
	result := exchange.HuobiInfo{}
	accounts, err := ep.GetAccounts()
	if err != nil {
		return result, err
	}
	if len(accounts.Data) == 0 {
		return result, errors.New("cannot get Huobi account")
	}
	respBody, err := ep.getResponse(
		"GET",
		ep.interf.AuthenticatedEndpoint()+"/v1/account/accounts/"+strconv.FormatUint(accounts.Data[0].ID, 10)+"/balance",
		map[string]string{},
		true,
	)
	if err == nil {
		err = json.Unmarshal(respBody, &result)
	}
	return result, err
}

func (ep *Endpoint) GetAccountTradeHistory(
	base, quote common.Token) (exchange.HuobiTradeHistory, error) {
	result := exchange.HuobiTradeHistory{}
	symbol := strings.ToUpper(fmt.Sprintf("%s%s", base.ID, quote.ID))
	respBody, err := ep.getResponse(
		"GET",
		ep.interf.AuthenticatedEndpoint()+"/v1/order/orders",
		map[string]string{
			"symbol": strings.ToLower(symbol),
			"states": "filled",
		},
		true,
	)
	if err == nil {
		err = json.Unmarshal(respBody, &result)
	}
	return result, err
}

func (ep *Endpoint) GetDepositAddress(asset string) (exchange.HuobiDepositAddress, error) {
	result := exchange.HuobiDepositAddress{}
	respBody, err := ep.getResponse(
		"GET",
		huobiDepositEndpoint+"v1/dw/deposit-virtual/addresses",
		map[string]string{
			"currency": asset,
		},
		true,
	)
	if err == nil {
		if err = json.Unmarshal(respBody, &result); err != nil {
			return result, err
		}
		if !result.Success {
			err = fmt.Errorf("get Huobi deposit address failed: %s", result.Reason)
		}
	}
	return result, err
}

func (ep *Endpoint) GetExchangeInfo() (exchange.HuobiExchangeInfo, error) {
	result := exchange.HuobiExchangeInfo{}
	respBody, err := ep.getResponse(
		"GET",
		ep.interf.PublicEndpoint()+"/v1/common/symbols",
		map[string]string{},
		false,
	)
	if err == nil {
		err = json.Unmarshal(respBody, &result)
	}
	return result, err
}

//NewHuobiEndpoint return new endpoint instance
func NewHuobiEndpoint(signer Signer, interf Interface) *Endpoint {
	return &Endpoint{signer, interf}
}
