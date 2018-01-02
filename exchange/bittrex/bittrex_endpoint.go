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

	"sync"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/exchange"
	ethereum "github.com/ethereum/go-ethereum/common"
)

const EPSILON float64 = 0.0000000001 // 10e-10

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
		q.Set("apikey", self.signer.GetBittrexKey())
		q.Set("nonce", nonce())
		req.URL.RawQuery = q.Encode()
		req.Header.Add("apisign", self.signer.BittrexSign(req.URL.String()))
	}
}

func (self *BittrexEndpoint) GetResponse(
	url string, params map[string]string, signNeeded bool, timepoint uint64) ([]byte, error) {
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
	timepoint := common.GetTimepoint()
	resp_body, err := self.GetResponse(
		addPath(self.interf.PublicEndpoint(timepoint), "getmarkets"),
		map[string]string{},
		false,
		timepoint,
	)
	if err == nil {
		err = json.Unmarshal(resp_body, &result)
	}
	return result, err
}

func (self *BittrexEndpoint) FetchOnePairData(wq *sync.WaitGroup, pair common.TokenPair, data *sync.Map, timepoint uint64) {
	defer wq.Done()
	result := common.ExchangePrice{}
	result.Timestamp = common.Timestamp(fmt.Sprintf("%d", timepoint))
	result.Valid = true
	resp_body, err := self.GetResponse(
		addPath(self.interf.PublicEndpoint(timepoint), "getorderbook"),
		map[string]string{
			"market": fmt.Sprintf("%s-%s", pair.Quote.ID, pair.Base.ID),
			"type":   "both",
		},
		false,
		timepoint,
	)
	returnTime := common.GetTimestamp()
	result.ReturnTime = returnTime

	if err != nil {
		result.Valid = false
		result.Error = err.Error()
	} else {
		data := exchange.Bittresp{}
		json.Unmarshal(resp_body, &data)
		if !data.Success {
			result.Valid = false
			result.Error = data.Msg
		} else {
			for _, buy := range data.Result["buy"] {
				result.Bids = append(
					result.Bids,
					common.PriceEntry{
						buy["Quantity"],
						buy["Rate"],
					},
				)
			}
			for _, sell := range data.Result["sell"] {
				result.Asks = append(
					result.Asks,
					common.PriceEntry{
						sell["Quantity"],
						sell["Rate"],
					},
				)
			}
		}
	}
	data.Store(pair.PairID(), result)
}

func (self *BittrexEndpoint) Trade(tradeType string, base, quote common.Token, rate, amount float64, timepoint uint64) (string, float64, float64, bool, error) {
	result := exchange.Bitttrade{}
	var url string
	if tradeType == "sell" {
		url = addPath(self.interf.MarketEndpoint(timepoint), "selllimit")
	} else {
		url = addPath(self.interf.MarketEndpoint(timepoint), "buylimit")
	}
	params := map[string]string{
		"market":   fmt.Sprintf("%s-%s", strings.ToUpper(quote.ID), strings.ToUpper(base.ID)),
		"quantity": strconv.FormatFloat(amount, 'f', -1, 64),
		"rate":     strconv.FormatFloat(rate, 'f', -1, 64),
	}
	resp_body, err := self.GetResponse(
		url, params, true, timepoint)

	if err != nil {
		return "", 0, 0, false, errors.New("Trade rejected by Bittrex")
	} else {
		err = json.Unmarshal(resp_body, &result)
		if err != nil {
			return "", 0, 0, false, err
		} else {
			if result.Success {
				uuid := result.Result["uuid"]
				done, remaining, finished, err := self.QueryOrder(
					uuid, timepoint+20)
				return uuid, done, remaining, finished, err
			} else {
				return "", 0, 0, false, errors.New(result.Error)
			}
		}
	}
}

func (self *BittrexEndpoint) OrderStatus(uuid string, timepoint uint64) (exchange.Bitttraderesult, error) {
	result := exchange.Bitttraderesult{}
	resp_body, err := self.GetResponse(
		addPath(self.interf.AccountEndpoint(timepoint), "getorder"),
		map[string]string{
			"uuid": uuid,
		},
		true,
		timepoint,
	)
	if err != nil {
		return result, err
	} else {
		err = json.Unmarshal(resp_body, &result)
		return result, err
	}
}

func (self *BittrexEndpoint) QueryOrder(uuid string, timepoint uint64) (float64, float64, bool, error) {
	result, err := self.OrderStatus(uuid, timepoint)
	if err != nil {
		return 0, 0, false, err
	} else {
		remaining := result.Result.QuantityRemaining
		done := result.Result.Quantity - remaining
		return done, remaining, remaining < EPSILON, nil
	}
}

func (self *BittrexEndpoint) WithdrawHistory(currency string, timepoint uint64) (exchange.Bittwithdrawhistory, error) {
	result := exchange.Bittwithdrawhistory{}
	resp_body, err := self.GetResponse(
		addPath(self.interf.AccountEndpoint(timepoint), "getwithdrawalhistory"),
		map[string]string{
			"currency": currency,
		},
		true,
		timepoint,
	)
	if err != nil {
		return result, err
	} else {
		err = json.Unmarshal(resp_body, &result)
		return result, err
	}
}

func (self *BittrexEndpoint) DepositHistory(currency string, timepoint uint64) (exchange.Bittdeposithistory, error) {
	result := exchange.Bittdeposithistory{}
	resp_body, err := self.GetResponse(
		addPath(self.interf.AccountEndpoint(timepoint), "getdeposithistory"),
		map[string]string{
			"currency": currency,
		},
		true,
		timepoint,
	)
	if err != nil {
		return result, err
	} else {
		err = json.Unmarshal(resp_body, &result)
		return result, err
	}
}

func (self *BittrexEndpoint) Withdraw(token common.Token, amount *big.Int, address ethereum.Address, timepoint uint64) (exchange.Bittwithdraw, error) {
	result := exchange.Bittwithdraw{}
	resp_body, err := self.GetResponse(
		addPath(self.interf.AccountEndpoint(timepoint), "withdraw"),
		map[string]string{
			"currency": strings.ToUpper(token.ID),
			"quantity": strconv.FormatFloat(common.BigToFloat(amount, token.Decimal), 'f', -1, 64),
			"address":  address.Hex(),
		},
		true,
		timepoint,
	)
	if err != nil {
		return result, err
	} else {
		err = json.Unmarshal(resp_body, &result)
		return result, err
	}
}

func (self *BittrexEndpoint) GetInfo(timepoint uint64) (exchange.Bittinfo, error) {
	result := exchange.Bittinfo{}
	resp_body, err := self.GetResponse(
		addPath(self.interf.AccountEndpoint(timepoint), "getbalances"),
		map[string]string{},
		true,
		timepoint,
	)
	if err != nil {
		return result, err
	} else {
		err = json.Unmarshal(resp_body, &result)
		return result, err
	}
}

func (self *BittrexEndpoint) CancelOrder(uuid string, timepoint uint64) (exchange.Bittcancelorder, error) {
	result := exchange.Bittcancelorder{}
	resp_body, err := self.GetResponse(
		addPath(self.interf.MarketEndpoint(timepoint), "cancel"),
		map[string]string{
			"uuid": uuid,
		},
		true,
		timepoint,
	)
	if err != nil {
		return result, err
	} else {
		err = json.Unmarshal(resp_body, &result)
		return result, err
	}
}

func NewBittrexEndpoint(signer Signer, interf Interface) *BittrexEndpoint {
	return &BittrexEndpoint{signer, interf}
}

func NewRealBittrexEndpoint(signer Signer) *BittrexEndpoint {
	return &BittrexEndpoint{signer, NewRealInterface()}
}

func NewSimulatedBittrexEndpoint(signer Signer) *BittrexEndpoint {
	return &BittrexEndpoint{signer, NewSimulatedInterface()}
}

func NewDevBittrexEndpoint(signer Signer) *BittrexEndpoint {
	return &BittrexEndpoint{signer, NewDevInterface()}
}

func NewKovanBittrexEndpoint(signer Signer) *BittrexEndpoint {
	return &BittrexEndpoint{signer, NewKovanInterface()}
}
