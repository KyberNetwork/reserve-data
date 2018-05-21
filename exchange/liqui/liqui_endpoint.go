package liqui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/exchange"
	ethereum "github.com/ethereum/go-ethereum/common"
)

type LiquiEndpoint struct {
	signer Signer
	interf Interface
}

func nonce() string {
	epsilon := 30 * time.Millisecond
	anchor := int64(50299954901)
	timestamp := time.Now().UnixNano()/int64(epsilon) - anchor
	return strconv.Itoa(int(timestamp))
}

func (self *LiquiEndpoint) Depth(tokens string, timepoint uint64) (exchange.Liqresp, error) {
	result := exchange.Liqresp{}
	client := &http.Client{
		Timeout: time.Duration(30 * time.Second)}
	u, err := url.Parse(self.interf.PublicEndpoint(timepoint))
	if err != nil {
		panic(err)
	}
	q := u.Query()
	q.Set("ignore_invalid", "1")
	u.RawQuery = q.Encode()
	u.Path = path.Join(
		u.Path,
		"depth",
		tokens,
	)
	req, _ := http.NewRequest("GET", u.String(), nil)
	req.Header.Add("Accept", "application/json")
	resp, err := client.Do(req)
	if err == nil {
		if resp.StatusCode == 200 {
			defer resp.Body.Close()
			resp_body, err := ioutil.ReadAll(resp.Body)
			if err == nil {
				json.Unmarshal(resp_body, &result)
			}
		} else {
			err = fmt.Errorf("Unsuccessful response from Liqui: Status " + resp.Status)
		}
	}
	return result, err
}

func (self *LiquiEndpoint) CancelOrder(id string) (exchange.Liqcancel, error) {
	result := exchange.Liqcancel{}
	client := &http.Client{
		Timeout: time.Duration(30 * time.Second)}
	data := url.Values{}
	data.Set("method", "CancelOrder")
	data.Set("order_id", id)
	data.Add("nonce", nonce())
	params := data.Encode()
	req, _ := http.NewRequest(
		"POST",
		self.interf.AuthenticatedEndpoint(common.GetTimepoint()),
		bytes.NewBufferString(params),
	)
	req.Header.Add("Content-Length", strconv.Itoa(len(params)))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Key", self.signer.GetKey())
	req.Header.Add("Sign", self.signer.Sign(params))
	resp, err := client.Do(req)
	if err == nil && resp.StatusCode == 200 {
		defer resp.Body.Close()
		resp_body, err := ioutil.ReadAll(resp.Body)
		log.Printf("response: %s\n", resp_body)
		if err == nil {
			err = json.Unmarshal(resp_body, &result)
		}
		return result, err
	} else {
		return result, fmt.Errorf("Cancel rejected by Liqui")
	}
}

func (self *LiquiEndpoint) Trade(tradeType string, base, quote common.Token, rate, amount float64, timepoint uint64) (id string, done float64, remaining float64, finished bool, err error) {
	result := exchange.Liqtrade{}
	client := &http.Client{
		Timeout: time.Duration(30 * time.Second)}
	data := url.Values{}
	data.Set("method", "Trade")
	data.Set("pair", fmt.Sprintf("%s_%s", strings.ToLower(base.ID), strings.ToLower(quote.ID)))
	data.Set("type", tradeType)
	data.Set("rate", strconv.FormatFloat(rate, 'f', -1, 64))
	data.Set("amount", strconv.FormatFloat(amount, 'f', -1, 64))
	data.Add("nonce", nonce())
	params := data.Encode()
	req, _ := http.NewRequest(
		"POST",
		self.interf.AuthenticatedEndpoint(timepoint),
		bytes.NewBufferString(params),
	)
	req.Header.Add("Content-Length", strconv.Itoa(len(params)))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Key", self.signer.GetKey())
	req.Header.Add("Sign", self.signer.Sign(params))
	resp, err := client.Do(req)
	if err == nil && resp.StatusCode == 200 {
		defer resp.Body.Close()
		resp_body, err := ioutil.ReadAll(resp.Body)
		log.Printf("response: %s\n", resp_body)
		if err == nil {
			err = json.Unmarshal(resp_body, &result)
		}
		if err != nil {
			return "", 0, 0, false, err
		}
		if result.Error != "" {
			return "", 0, 0, false, fmt.Errorf(result.Error)
		}
		return strconv.FormatUint(result.Return.OrderID, 10), result.Return.Done, result.Return.Remaining, result.Return.OrderID == 0, nil
	} else {
		log.Printf("Error: %v, Code: %v\n", err, resp)
		return "", 0, 0, false, fmt.Errorf("Trade rejected by Liqui")
	}
}

func (self *LiquiEndpoint) Withdraw(token common.Token, amount *big.Int, address ethereum.Address, timepoint uint64) error {
	// ignoring timepoint because it's only relevant in simulation
	result := exchange.Liqwithdraw{}
	client := &http.Client{
		Timeout: time.Duration(30 * time.Second),
	}
	data := url.Values{}
	data.Set("method", "WithdrawCoin")
	data.Set("coinName", token.ID)
	data.Set("amount", strconv.FormatFloat(common.BigToFloat(amount, token.Decimal), 'f', -1, 64))
	data.Set("address", address.Hex())
	data.Add("nonce", nonce())
	params := data.Encode()
	req, _ := http.NewRequest(
		"POST",
		self.interf.AuthenticatedEndpoint(timepoint),
		bytes.NewBufferString(params),
	)
	req.Header.Add("Content-Length", strconv.Itoa(len(params)))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Key", self.signer.GetKey())
	req.Header.Add("Sign", self.signer.Sign(params))
	resp, err := client.Do(req)
	if err == nil && resp.StatusCode == 200 {
		defer resp.Body.Close()
		resp_body, err := ioutil.ReadAll(resp.Body)
		log.Printf("response: %s\n", resp_body)
		if err == nil {
			err = json.Unmarshal(resp_body, &result)
		}
		if err != nil {
			return err
		}
		if result.Error != "" {
			return fmt.Errorf(result.Error)
		}
		return nil
	} else {
		log.Printf("Error: %v, Code: %v\n", err, resp)
		return fmt.Errorf("withdraw rejected by Liqui")
	}
}

func (self *LiquiEndpoint) GetInfo(timepoint uint64) (exchange.Liqinfo, error) {
	result := exchange.Liqinfo{}
	client := &http.Client{
		Timeout: time.Duration(30 * time.Second)}
	data := url.Values{}
	data.Set("method", "getInfo")
	data.Add("nonce", nonce())
	params := data.Encode()
	log.Printf("endpoint: %v\n", self.interf.AuthenticatedEndpoint(timepoint))
	req, _ := http.NewRequest(
		"POST",
		self.interf.AuthenticatedEndpoint(timepoint),
		bytes.NewBufferString(params),
	)
	req.Header.Add("Content-Length", strconv.Itoa(len(params)))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Key", self.signer.GetKey())
	req.Header.Add("Sign", self.signer.Sign(params))
	resp, err := client.Do(req)
	if err == nil {
		if resp.StatusCode == 200 {
			defer resp.Body.Close()
			resp_body, err := ioutil.ReadAll(resp.Body)
			log.Printf("Liqui GetInfo response: %s", string(resp_body))
			if err == nil {
				json.Unmarshal(resp_body, &result)
			}
			log.Printf("Liqui GetInfo data: %v", result)
		} else {
			err = fmt.Errorf("Unsuccessful response from Liqui: Status " + resp.Status)
		}
	}
	return result, err
}

func (self *LiquiEndpoint) OrderInfo(orderID string, timepoint uint64) (exchange.Liqorderinfo, error) {
	result := exchange.Liqorderinfo{}
	client := &http.Client{
		Timeout: time.Duration(30 * time.Second)}
	data := url.Values{}
	data.Set("method", "OrderInfo")
	data.Set("order_id", orderID)
	data.Add("nonce", nonce())
	params := data.Encode()
	req, _ := http.NewRequest(
		"POST",
		self.interf.AuthenticatedEndpoint(timepoint),
		bytes.NewBufferString(params),
	)
	req.Header.Add("Content-Length", strconv.Itoa(len(params)))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Key", self.signer.GetKey())
	req.Header.Add("Sign", self.signer.Sign(params))
	resp, err := client.Do(req)
	if err == nil {
		if resp.StatusCode == 200 {
			defer resp.Body.Close()
			resp_body, err := ioutil.ReadAll(resp.Body)
			log.Printf("Liqui Order info response: %s", string(resp_body))
			if err == nil {
				json.Unmarshal(resp_body, &result)
			}
			log.Printf("Liqui Order info data: %v", result)
		} else {
			err = fmt.Errorf("Unsuccessful response from Liqui: Status " + resp.Status)
		}
	}
	return result, err
}

func (self *LiquiEndpoint) ActiveOrders(timepoint uint64) (exchange.Liqorders, error) {
	result := exchange.Liqorders{}
	client := &http.Client{
		Timeout: time.Duration(30 * time.Second)}
	data := url.Values{}
	data.Set("method", "ActiveOrders")
	data.Set("pair", "") // all pairs
	data.Add("nonce", nonce())
	params := data.Encode()
	req, _ := http.NewRequest(
		"POST",
		self.interf.AuthenticatedEndpoint(timepoint),
		bytes.NewBufferString(params),
	)
	req.Header.Add("Content-Length", strconv.Itoa(len(params)))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Key", self.signer.GetKey())
	req.Header.Add("Sign", self.signer.Sign(params))
	resp, err := client.Do(req)
	if err == nil {
		if resp.StatusCode == 200 {
			defer resp.Body.Close()
			resp_body, err := ioutil.ReadAll(resp.Body)
			log.Printf("Liqui ActiveOrders response: %s", string(resp_body))
			if err == nil {
				json.Unmarshal(resp_body, &result)
			}
			log.Printf("Liqui ActiveOrders data: %v", result)
		} else {
			err = fmt.Errorf("Unsuccessful response from Liqui: Status " + resp.Status)
		}
	}
	return result, err
}

func NewLiquiEndpoint(signer Signer, interf Interface) *LiquiEndpoint {
	return &LiquiEndpoint{signer, interf}
}

func NewRealLiquiEndpoint(signer Signer) *LiquiEndpoint {
	return &LiquiEndpoint{signer, NewRealInterface()}
}

func NewSimulatedLiquiEndpoint(signer Signer) *LiquiEndpoint {
	return &LiquiEndpoint{signer, NewSimulatedInterface()}
}

func NewKovanLiquiEndpoint(signer Signer) *LiquiEndpoint {
	return &LiquiEndpoint{signer, NewKovanInterface()}
}

func NewDevLiquiEndpoint(signer Signer) *LiquiEndpoint {
	return &LiquiEndpoint{signer, NewDevInterface()}
}
