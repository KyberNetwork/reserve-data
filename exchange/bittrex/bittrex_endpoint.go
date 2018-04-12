package bittrex

import (
	"encoding/json"
	"errors"
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
	} else {
		url.Path = fmt.Sprintf("%s/%s", url.Path, path)
		return url.String()
	}
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
	var resp_body []byte
	log.Printf("request to bittrex: %s\n", req.URL)
	resp, err := client.Do(req)
	if err != nil {
		return resp_body, err
	} else {
		defer resp.Body.Close()
		resp_body, err = ioutil.ReadAll(resp.Body)
		log.Printf("request to %s, got response from bittrex: %s\n", req.URL, common.TruncStr(resp_body))
		return resp_body, err
	}
}

func (self *BittrexEndpoint) GetExchangeInfo() (exchange.BittExchangeInfo, error) {
	result := exchange.BittExchangeInfo{}
	resp_body, err := self.GetResponse(
		addPath(self.interf.PublicEndpoint(), "getmarkets"),
		map[string]string{},
		false,
	)
	if err == nil {
		err = json.Unmarshal(resp_body, &result)
	}
	return result, err
}

func (self *BittrexEndpoint) FetchOnePairData(pair common.TokenPair) (exchange.Bittresp, error) {
	data := exchange.Bittresp{}
	resp_body, err := self.GetResponse(
		addPath(self.interf.PublicEndpoint(), "getorderbook"),
		map[string]string{
			"market": fmt.Sprintf("%s-%s", pair.Quote.ID, pair.Base.ID),
			"type":   "both",
		},
		false,
	)

	if err != nil {
		return data, err
	} else {
		json.Unmarshal(resp_body, &data)
		return data, nil
	}
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
	resp_body, err := self.GetResponse(
		url, params, true)

	if err != nil {
		return result, err
	} else {
		json.Unmarshal(resp_body, &result)
		return result, nil
	}
}

func (self *BittrexEndpoint) OrderStatus(uuid string) (exchange.Bitttraderesult, error) {
	result := exchange.Bitttraderesult{}
	resp_body, err := self.GetResponse(
		addPath(self.interf.AccountEndpoint(), "getorder"),
		map[string]string{
			"uuid": uuid,
		},
		true,
	)
	if err != nil {
		return result, err
	} else {
		err = json.Unmarshal(resp_body, &result)
		return result, err
	}
}

func (self *BittrexEndpoint) GetDepositAddress(currency string) (exchange.BittrexDepositAddress, error) {
	result := exchange.BittrexDepositAddress{}
	resp_body, err := self.GetResponse(
		addPath(self.interf.AccountEndpoint(), "getdepositaddress"),
		map[string]string{
			"currency": currency,
		},
		true,
	)
	if err == nil {
		json.Unmarshal(resp_body, &result)
	}
	return result, err
}

func (self *BittrexEndpoint) WithdrawHistory(currency string) (exchange.Bittwithdrawhistory, error) {
	result := exchange.Bittwithdrawhistory{}
	resp_body, err := self.GetResponse(
		addPath(self.interf.AccountEndpoint(), "getwithdrawalhistory"),
		map[string]string{
			"currency": currency,
		},
		true,
	)
	if err != nil {
		return result, err
	} else {
		err = json.Unmarshal(resp_body, &result)
		return result, err
	}
}

func (self *BittrexEndpoint) DepositHistory(currency string) (exchange.Bittdeposithistory, error) {
	result := exchange.Bittdeposithistory{}
	resp_body, err := self.GetResponse(
		addPath(self.interf.AccountEndpoint(), "getdeposithistory"),
		map[string]string{
			"currency": currency,
		},
		true,
	)
	if err != nil {
		return result, err
	} else {
		err = json.Unmarshal(resp_body, &result)
		return result, err
	}
}

func (self *BittrexEndpoint) Withdraw(token common.Token, amount *big.Int, address ethereum.Address) (exchange.Bittwithdraw, error) {
	result := exchange.Bittwithdraw{}
	resp_body, err := self.GetResponse(
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
	} else {
		err = json.Unmarshal(resp_body, &result)
		return result, err
	}
}

func (self *BittrexEndpoint) GetInfo() (exchange.Bittinfo, error) {
	result := exchange.Bittinfo{}
	resp_body, err := self.GetResponse(
		addPath(self.interf.AccountEndpoint(), "getbalances"),
		map[string]string{},
		true,
	)
	if err != nil {
		return result, err
	} else {
		err = json.Unmarshal(resp_body, &result)
		return result, err
	}
}

func (self *BittrexEndpoint) CancelOrder(uuid string) (exchange.Bittcancelorder, error) {
	result := exchange.Bittcancelorder{}
	resp_body, err := self.GetResponse(
		addPath(self.interf.MarketEndpoint(), "cancel"),
		map[string]string{
			"uuid": uuid,
		},
		true,
	)
	if err != nil {
		return result, err
	} else {
		err = json.Unmarshal(resp_body, &result)
		return result, err
	}
}

func (self *BittrexEndpoint) GetAccountTradeHistory(base, quote common.Token) (exchange.BittTradeHistory, error) {
	result := exchange.BittTradeHistory{}
	params := map[string]string{}
	symbol := fmt.Sprintf("%s-%s", quote.ID, base.ID)
	if symbol != "" {
		params["market"] = symbol
	}
	resp_body, err := self.GetResponse(
		addPath(self.interf.AccountEndpoint(), "getorderhistory"),
		params,
		true,
	)
	if err == nil {
		json.Unmarshal(resp_body, &result)
		if !result.Success {
			return result, errors.New(fmt.Sprintf("Cannot get trade history: %s", result.Message))
		}
	}
	return result, err
}

func NewBittrexEndpoint(signer Signer, interf Interface) *BittrexEndpoint {
	return &BittrexEndpoint{signer, interf}
}
