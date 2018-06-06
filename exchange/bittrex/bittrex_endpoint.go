package bittrex

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/exchange"
	ethereum "github.com/ethereum/go-ethereum/common"
)

type BittrexEndpoint struct {
	signer Signer
	interf Interface
}

func nonce() string {
	epsilon := 30 * time.Millisecond
	anchor := int64(50299954901)
	timestamp := time.Now().UnixNano()/int64(epsilon) - anchor
	return strconv.Itoa(int(timestamp))
}

func addPath(original string, path string) string {
	url, err := url.Parse(original)
	if err != nil {
		panic(err)
	}
	url.Path = fmt.Sprintf("%s/%s", url.Path, path)
	return url.String()
}

func (self *BittrexEndpoint) fillRequest(req *http.Request, signNeeded bool) {
	req.Header.Add("Content-Type", "application/json;charset=utf-8")
	req.Header.Add("Accept", "application/json")
	if signNeeded {
		q := req.URL.Query()
		q.Set("apikey", self.signer.GetKey())
		q.Set("nonce", nonce())
		req.URL.RawQuery = q.Encode()
		req.Header.Add("apisign", self.signer.Sign(req.URL.String()))
	}
}

func (self *BittrexEndpoint) GetResponse(
	url string, params map[string]string, signNeeded bool) ([]byte, error) {
	client := &http.Client{
		Timeout: time.Duration(30 * time.Second),
	}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")

	q := req.URL.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()
	self.fillRequest(req, signNeeded)
	var err error
	var respBody []byte
	log.Printf("request to bittrex: %s\n", req.URL)
	resp, err := client.Do(req)
	if err != nil {
		return respBody, err
	}
	defer func() {
		if vErr := resp.Body.Close(); vErr != nil {
			log.Printf("Unmarshal response error: %s", vErr.Error())
		}
	}()
	respBody, err = ioutil.ReadAll(resp.Body)
	log.Printf("request to %s, got response from bittrex: %s\n", req.URL, common.TruncStr(respBody))
	return respBody, err
}

func (self *BittrexEndpoint) GetExchangeInfo() (exchange.BittExchangeInfo, error) {
	result := exchange.BittExchangeInfo{}
	respBody, err := self.GetResponse(
		addPath(self.interf.PublicEndpoint(), "getmarkets"),
		map[string]string{},
		false,
	)
	if err == nil {
		err = json.Unmarshal(respBody, &result)
	}
	return result, err
}

func (self *BittrexEndpoint) FetchOnePairData(pair common.TokenPair) (exchange.Bittresp, error) {
	data := exchange.Bittresp{}
	respBody, err := self.GetResponse(
		addPath(self.interf.PublicEndpoint(), "getorderbook"),
		map[string]string{
			"market": fmt.Sprintf("%s-%s", pair.Quote.ID, pair.Base.ID),
			"type":   "both",
		},
		false,
	)

	if err != nil {
		return data, err
	}
	err = json.Unmarshal(respBody, &data)
	return data, err
}

func (self *BittrexEndpoint) Trade(
	tradeType string,
	base, quote common.Token,
	rate, amount float64) (exchange.Bitttrade, error) {

	result := exchange.Bitttrade{}
	var url string
	if tradeType == "sell" {
		url = addPath(self.interf.MarketEndpoint(), "selllimit")
	} else {
		url = addPath(self.interf.MarketEndpoint(), "buylimit")
	}
	params := map[string]string{
		"market":   fmt.Sprintf("%s-%s", strings.ToUpper(quote.ID), strings.ToUpper(base.ID)),
		"quantity": strconv.FormatFloat(amount, 'f', -1, 64),
		"rate":     strconv.FormatFloat(rate, 'f', -1, 64),
	}
	respBody, err := self.GetResponse(
		url, params, true)

	if err != nil {
		return result, err
	}
	err = json.Unmarshal(respBody, &result)
	return result, err
}

func (self *BittrexEndpoint) OrderStatus(uuid string) (exchange.Bitttraderesult, error) {
	result := exchange.Bitttraderesult{}
	respBody, err := self.GetResponse(
		addPath(self.interf.AccountEndpoint(), "getorder"),
		map[string]string{
			"uuid": uuid,
		},
		true,
	)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(respBody, &result)
	return result, err
}

func (self *BittrexEndpoint) GetDepositAddress(currency string) (exchange.BittrexDepositAddress, error) {
	result := exchange.BittrexDepositAddress{}
	respBody, err := self.GetResponse(
		addPath(self.interf.AccountEndpoint(), "getdepositaddress"),
		map[string]string{
			"currency": currency,
		},
		true,
	)
	if err == nil {
		err = json.Unmarshal(respBody, &result)
	}
	return result, err
}

func (self *BittrexEndpoint) WithdrawHistory(currency string) (exchange.Bittwithdrawhistory, error) {
	result := exchange.Bittwithdrawhistory{}
	respBody, err := self.GetResponse(
		addPath(self.interf.AccountEndpoint(), "getwithdrawalhistory"),
		map[string]string{
			"currency": currency,
		},
		true,
	)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(respBody, &result)
	return result, err
}

func (self *BittrexEndpoint) DepositHistory(currency string) (exchange.Bittdeposithistory, error) {
	result := exchange.Bittdeposithistory{}
	respBody, err := self.GetResponse(
		addPath(self.interf.AccountEndpoint(), "getdeposithistory"),
		map[string]string{
			"currency": currency,
		},
		true,
	)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(respBody, &result)
	return result, err
}

func (self *BittrexEndpoint) Withdraw(token common.Token, amount *big.Int, address ethereum.Address) (exchange.Bittwithdraw, error) {
	result := exchange.Bittwithdraw{}
	respBody, err := self.GetResponse(
		addPath(self.interf.AccountEndpoint(), "withdraw"),
		map[string]string{
			"currency": strings.ToUpper(token.ID),
			"quantity": strconv.FormatFloat(common.BigToFloat(amount, token.Decimal), 'f', -1, 64),
			"address":  address.Hex(),
		},
		true,
	)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(respBody, &result)
	return result, err
}

func (self *BittrexEndpoint) GetInfo() (exchange.Bittinfo, error) {
	result := exchange.Bittinfo{}
	respBody, err := self.GetResponse(
		addPath(self.interf.AccountEndpoint(), "getbalances"),
		map[string]string{},
		true,
	)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(respBody, &result)
	return result, err
}

func (self *BittrexEndpoint) CancelOrder(uuid string) (exchange.Bittcancelorder, error) {
	result := exchange.Bittcancelorder{}
	respBody, err := self.GetResponse(
		addPath(self.interf.MarketEndpoint(), "cancel"),
		map[string]string{
			"uuid": uuid,
		},
		true,
	)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(respBody, &result)
	return result, err
}

func (self *BittrexEndpoint) GetAccountTradeHistory(base, quote common.Token) (exchange.BittTradeHistory, error) {
	result := exchange.BittTradeHistory{}
	params := map[string]string{}
	symbol := fmt.Sprintf("%s-%s", quote.ID, base.ID)
	if symbol != "" {
		params["market"] = symbol
	}
	respBody, err := self.GetResponse(
		addPath(self.interf.AccountEndpoint(), "getorderhistory"),
		params,
		true,
	)
	if err == nil {
		if err = json.Unmarshal(respBody, &result); err != nil {
			return result, err
		}
		if !result.Success {
			return result, fmt.Errorf("Cannot get trade history: %s", result.Message)
		}
	}
	return result, err
}

func NewBittrexEndpoint(signer Signer, interf Interface) *BittrexEndpoint {
	return &BittrexEndpoint{signer, interf}
}
