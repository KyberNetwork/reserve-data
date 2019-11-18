package huobi

import (
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

	ethereum "github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/exchange"
)

//Endpoint endpoint object
type Endpoint struct {
	signer Signer
	interf Interface
	l      *zap.SugaredLogger
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
	for k, v := range params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()
	ep.fillRequest(req, signNeeded)
	var respBody []byte
	resp, err := client.Do(req)
	if err != nil {
		return respBody, err
	}
	defer func() {
		if cErr := resp.Body.Close(); cErr != nil {
			ep.l.Warnf("Response body close error: %v", cErr)
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

// GetDepthOnePair get depth one pair on huobi
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

// Trade create a new order on huobi
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
func (ep *Endpoint) WithdrawHistory(tokens []common.Token) (exchange.HuobiWithdraws, error) {
	result := exchange.HuobiWithdraws{}
	size := len(tokens) * 2
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
func (ep *Endpoint) DepositHistory(tokens []common.Token) (exchange.HuobiDeposits, error) {
	result := exchange.HuobiDeposits{}
	size := len(tokens) * 2
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

// CancelOrder cancel open order on huobi
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

// OrderStatus check order status on huobi
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

// Withdraw token from huobi
func (ep *Endpoint) Withdraw(token common.Token, amount *big.Int, address ethereum.Address) (string, error) {
	result := exchange.HuobiWithdraw{}
	respBody, err := ep.GetResponse(
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
		ep.l.Infof("Withdraw id: %s", fmt.Sprintf("%v", result.ID))
		return strconv.FormatUint(result.ID, 10), nil
	}
	return "", errors.New("Withdraw rejected by Huobi")
}

// GetInfo return exchange info
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

// GetAccountTradeHistory return account trade history
func (ep *Endpoint) GetAccountTradeHistory(
	base, quote common.Token) (exchange.HuobiTradeHistory, error) {
	result := exchange.HuobiTradeHistory{}
	symbol := strings.ToUpper(fmt.Sprintf("%s%s", base.ID, quote.ID))
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

//GetDepositAddress get huobi deposit address
func (ep *Endpoint) GetDepositAddress(asset string) (exchange.HuobiDepositAddress, error) {
	result := exchange.HuobiDepositAddress{}
	respBody, err := ep.GetResponse(
		"GET",
		fmt.Sprintf("%s/v2/account/deposit/address", ep.interf.AuthenticatedEndpoint()),
		map[string]string{
			"currency": strings.ToLower(asset),
		},
		true,
	)
	if err == nil {
		if err = json.Unmarshal(respBody, &result); err != nil {
			ep.l.Errorw("failed to get deposit address from Huobi", "token", asset, "response", string(respBody))
			return result, err
		}
		if result.Code != 200 {
			err = fmt.Errorf("get Huobi deposit address failed: %s", result.Message)
		}
	}
	return result, err
}

// GetExchangeInfo get huobi exchange info
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

//NewHuobiEndpoint return new endpoint instance
func NewHuobiEndpoint(signer Signer, interf Interface) *Endpoint {
	return &Endpoint{signer: signer, interf: interf, l: zap.S()}
}
