package huobi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	ethereum "github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/exchange"
	commonv3 "github.com/KyberNetwork/reserve-data/reservesetting/common"
)

//Endpoint endpoint object
type Endpoint struct {
	signer Signer
	interf Interface
	l      *zap.SugaredLogger
	client *http.Client
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

//GetResponse from huobi api
func (ep *Endpoint) GetResponse(
	method string, reqURL string,
	params map[string]string, signNeeded bool) ([]byte, error) {

	reqBody, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, reqURL, nil)
	if err != nil {
		return nil, err
	}
	if method == "POST" {
		req.Body = ioutil.NopCloser(strings.NewReader(string(reqBody)))
	}
	req.Header.Add("Accept", "application/json")

	q := req.URL.Query()
	if signNeeded {
		timestamp := time.Now().UTC().Format("2006-01-02T15:04:05")
		params["SignatureMethod"] = "HmacSHA256"
		params["SignatureVersion"] = "2"
		params["AccessKeyId"] = ep.signer.GetKey()
		params["Timestamp"] = timestamp
		params["op"] = "auth"
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
	resp, err := ep.client.Do(req)
	if err != nil {
		return respBody, err
	}
	defer func() {
		if cErr := resp.Body.Close(); cErr != nil {
			ep.l.Warnw("response body close failed", "err", cErr)
		}
	}()
	switch resp.StatusCode {
	case 429:
		err = errors.New("breaking Huobi request rate limit")
	case 500:
		err = errors.New("500 from Huobi, its fault")
	case 200:
		respBody, err = ioutil.ReadAll(resp.Body)
	}
	return respBody, err
}

//GetAccounts Get account list for later use
func (ep *Endpoint) GetAccounts() (exchange.HuobiAccounts, error) {
	result := exchange.HuobiAccounts{}
	resp, err := ep.GetResponse(
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

	respBody, err := ep.GetResponse(
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

// Trade on Huobi
func (ep *Endpoint) Trade(tradeType string, pair commonv3.TradingPairSymbols, rate, amount float64) (exchange.HuobiTrade, error) {
	result := exchange.HuobiTrade{}
	symbol := strings.ToLower(pair.BaseSymbol) + strings.ToLower(pair.QuoteSymbol)
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
	respBody, err := ep.GetResponse(
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

//WithdrawHistory return withdraw history from huobi
func (ep *Endpoint) WithdrawHistory(size int) (exchange.HuobiWithdraws, error) {
	result := exchange.HuobiWithdraws{}
	respBody, err := ep.GetResponse(
		"GET",
		ep.interf.AuthenticatedEndpoint()+"/v1/query/deposit-withdraw",
		map[string]string{
			"size":   strconv.Itoa(size),
			"type":   "withdraw",
			"direct": "next",
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

//DepositHistory get deposit history from huobi
func (ep *Endpoint) DepositHistory(size int) (exchange.HuobiDeposits, error) {
	result := exchange.HuobiDeposits{}
	respBody, err := ep.GetResponse(
		"GET",
		ep.interf.AuthenticatedEndpoint()+"/v1/query/deposit-withdraw",
		map[string]string{
			"size":   strconv.Itoa(size),
			"type":   "deposit",
			"direct": "next",
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
	ep.l.Infof("huobi deposit history: %v", result)
	return result, err
}

func (ep *Endpoint) CancelOrder(symbol string, id uint64) (exchange.HuobiCancel, error) {
	result := exchange.HuobiCancel{}
	respBody, err := ep.GetResponse(
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
	respBody, err := ep.GetResponse(
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

func (ep *Endpoint) Withdraw(asset commonv3.Asset, amount *big.Int, address ethereum.Address) (string, error) {
	var symbol string
	for _, exchg := range asset.Exchanges {
		if exchg.ExchangeID == uint64(common.Huobi) {
			symbol = exchg.Symbol
		}
	}
	result := exchange.HuobiWithdraw{}
	respBody, err := ep.GetResponse(
		"POST",
		ep.interf.AuthenticatedEndpoint()+"/v1/dw/withdraw/api/create",
		map[string]string{
			"address":  address.Hex(),
			"amount":   strconv.FormatFloat(common.BigToFloat(amount, int64(asset.Decimals)), 'f', -1, 64),
			"currency": strings.ToLower(symbol),
		},
		true,
	)
	if err != nil {
		ep.l.Errorf("withdraw rejected by Huobi %v", err)
		return "", errors.New("withdraw rejected by Huobi")
	}

	if err = json.Unmarshal(respBody, &result); err != nil {
		return "", err
	}
	if result.Status != "ok" {
		return "", fmt.Errorf("withdraw from Huobi failed: %s", result.Reason)
	}
	ep.l.Infow("do withdraw", "id", result.ID)
	return strconv.FormatUint(result.ID, 10), nil

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
	respBody, err := ep.GetResponse(
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
	baseSymbol, quoteSymbol string) (exchange.HuobiTradeHistory, error) {
	result := exchange.HuobiTradeHistory{}
	symbol := strings.ToUpper(fmt.Sprintf("%s%s", baseSymbol, quoteSymbol))
	respBody, err := ep.GetResponse(
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

// GetDepositAddress return deposit address
func (ep *Endpoint) GetDepositAddress(asset string) (exchange.HuobiDepositAddress, error) {
	result := exchange.HuobiDepositAddress{}
	respBody, err := ep.GetResponse(
		"GET",
		ep.interf.AuthenticatedEndpoint()+"/v2/account/deposit/address",
		map[string]string{
			"currency": strings.ToLower(asset),
		},
		true,
	)
	if err == nil {
		if err = json.Unmarshal(respBody, &result); err != nil {
			return result, err
		}
		if result.Code != 200 {
			err = fmt.Errorf("get Huobi deposit address failed: %s", result.Message)
		}
	}
	return result, err
}

func (ep *Endpoint) GetExchangeInfo() (exchange.HuobiExchangeInfo, error) {
	result := exchange.HuobiExchangeInfo{}
	respBody, err := ep.GetResponse(
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

// OpenOrdersForOnePair return list open orders for one pair of token
func (ep *Endpoint) OpenOrdersForOnePair(pair commonv3.TradingPairSymbols) ([]exchange.HuobiOrder, error) {
	result := []exchange.HuobiOrder{}
	accounts, err := ep.GetAccounts()
	if err != nil {
		return result, err
	}
	if len(accounts.Data) == 0 {
		return result, errors.New("cannot get Huobi account")
	}
	account := strconv.FormatUint(accounts.Data[0].ID, 10)
	respBody, err := ep.GetResponse(
		"GET",
		ep.interf.AuthenticatedEndpoint()+"/v1/order/openOrders",
		map[string]string{
			"account": account,
			"symbol":  pair.BaseSymbol + pair.QuoteSymbol,
		},
		true,
	)
	if err != nil {
		return result, err
	}
	if err = json.Unmarshal(respBody, &result); err != nil {
		return result, err
	}
	return result, nil
}

//NewHuobiEndpoint return new endpoint instance
func NewHuobiEndpoint(signer Signer, interf Interface, client *http.Client) *Endpoint {
	return &Endpoint{signer: signer, interf: interf, l: zap.S(), client: client}
}
