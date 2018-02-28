package binance

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/exchange"
	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/websocket"
)

type BinanceEndpoint struct {
	signer    Signer
	interf    Interface
	timeDelta int64
}

func (self *BinanceEndpoint) fillRequest(req *http.Request, signNeeded bool, timepoint uint64) {
	if req.Method == "POST" || req.Method == "PUT" || req.Method == "DELETE" {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("User-Agent", "binance/go")
	}
	req.Header.Add("Accept", "application/json")
	log.Printf("Bin Time Delta: %s", self.timeDelta)
	if signNeeded {
		q := req.URL.Query()
		sig := url.Values{}
		req.Header.Set("X-MBX-APIKEY", self.signer.GetBinanceKey())
		q.Set("timestamp", fmt.Sprintf("%d", int64(timepoint)+self.timeDelta-1000))
		q.Set("recvWindow", "5000")
		sig.Set("signature", self.signer.BinanceSign(q.Encode()))
		// Using separated values map for signature to ensure it is at the end
		// of the query. This is required for /wapi apis from binance without
		// any damn documentation about it!!!
		req.URL.RawQuery = q.Encode() + "&" + sig.Encode()
	}
}

func (self *BinanceEndpoint) GetResponse(
	method string, url string,
	params map[string]string, signNeeded bool, timepoint uint64) ([]byte, error) {

	client := &http.Client{
		Timeout: time.Duration(30 * time.Second),
	}
	req, _ := http.NewRequest(method, url, nil)
	req.Header.Add("Accept", "application/json")

	q := req.URL.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()
	self.fillRequest(req, signNeeded, timepoint)
	var err error
	var resp_body []byte
	log.Printf("request to binance: %s\n", req.URL)
	resp, err := client.Do(req)
	if err != nil {
		return resp_body, err
	} else {
		defer resp.Body.Close()
		switch resp.StatusCode {
		case 429:
			err = errors.New("breaking a request rate limit.")
		case 418:
			err = errors.New("IP has been auto-banned for continuing to send requests after receiving 429 codes.")
		case 500:
			err = errors.New("500 from Binance, its fault.")
		case 200:
			resp_body, err = ioutil.ReadAll(resp.Body)
		}
		if err != nil || len(resp_body) == 0 || rand.Int()%10 == 0 {
			log.Printf("request to %s, got response from binance (error or throttled to 10%): %s, err: %v", req.URL, common.TruncStr(resp_body), err)
		}
		return resp_body, err
	}
}

func (self *BinanceEndpoint) StoreOrderBookData(
	wg *sync.WaitGroup,
	pair common.TokenPair,
	data *sync.Map,
	dataChannel chan exchange.Orderbook,
	result common.ExchangePrice) {

	if wg != nil {
		defer wg.Done()
	}
	if result.Valid {
		orderBook := <-dataChannel
		result.ReturnTime = common.GetTimestamp()
		bids := orderBook.GetBids()
		asks := orderBook.GetAsks()

		for _, buy := range bids {
			quantity, _ := strconv.ParseFloat(buy[1], 64)
			rate, _ := strconv.ParseFloat(buy[0], 64)
			result.Bids = append(
				result.Bids,
				common.PriceEntry{
					quantity,
					rate,
				},
			)
		}
		for _, sell := range asks {
			quantity, _ := strconv.ParseFloat(sell[1], 64)
			rate, _ := strconv.ParseFloat(sell[0], 64)
			result.Asks = append(
				result.Asks,
				common.PriceEntry{
					quantity,
					rate,
				},
			)
		}
		log.Printf("Data to store on storage: %s\n", result)
	}
	data.Store(pair.PairID(), result)
}

func (self *BinanceEndpoint) FetchOnePairData(
	wg *sync.WaitGroup,
	pair common.TokenPair,
	data *sync.Map,
	timepoint uint64) {

	result := common.ExchangePrice{}

	timestamp := common.Timestamp(fmt.Sprintf("%d", timepoint))
	result.Timestamp = timestamp
	result.Valid = true

	resp_body, err := self.GetResponse(
		"GET", self.interf.PublicEndpoint()+"/api/v1/depth",
		map[string]string{
			"symbol": fmt.Sprintf("%s%s", pair.Base.ID, pair.Quote.ID),
			"limit":  "50",
		},
		false,
		common.GetTimepoint(),
	)

	returnTime := common.GetTimestamp()
	result.ReturnTime = returnTime
	dataChannel := make(chan exchange.Orderbook)

	if err != nil {
		return resp_data, err
	} else {
		err = json.Unmarshal(resp_body, &resp_data)
		if err != nil {
			log.Printf("failed to unmarshal response from binance: %s, Response: %s", err, resp_body)
			return resp_data, err
		} else {
			go func() {
				dataChannel <- resp_data
			}()
		}
		return resp_data, nil
	}
	self.StoreOrderBookData(wg, pair, data, dataChannel, result)
}

// Relevant params:
// symbol ("%s%s", base, quote)
// side (BUY/SELL)
// type (LIMIT/MARKET)
// timeInForce (GTC/IOC)
// quantity
// price
//
// In this version, we only support LIMIT order which means only buy/sell with acceptable price,
// and GTC time in force which means that the order will be active until it's implicitly canceled
func (self *BinanceEndpoint) Trade(tradeType string, base, quote common.Token, rate, amount float64, timepoint uint64) (exchange.Binatrade, error) {
	result := exchange.Binatrade{}
	symbol := base.ID + quote.ID
	orderType := "LIMIT"
	params := map[string]string{
		"symbol":      symbol,
		"side":        strings.ToUpper(tradeType),
		"type":        orderType,
		"timeInForce": "GTC",
		"quantity":    strconv.FormatFloat(amount, 'f', -1, 64),
	}
	if orderType == "LIMIT" {
		params["price"] = strconv.FormatFloat(rate, 'f', -1, 64)
	}
	resp_body, err := self.GetResponse(
		"POST",
		self.interf.AuthenticatedEndpoint()+"/api/v3/order",
		params,
		true,
		common.GetTimepoint(),
	)
	if err != nil {
		return result, err
	} else {
		json.Unmarshal(resp_body, &result)
		return result, nil
	}
}

func (self *BinanceEndpoint) GetTradeHistory(symbol string) (exchange.BinanceTradeHistory, error) {
	result := exchange.BinanceTradeHistory{}
	timepoint := common.GetTimepoint()
	resp_body, err := self.GetResponse(
		"GET",
		self.interf.PublicEndpoint()+"/api/v1/trades",
		map[string]string{
			"symbol": symbol,
			"limit":  "500",
		},
		false,
		timepoint,
	)
	if err == nil {
		json.Unmarshal(resp_body, &result)
	}
	return result, err
}

func (self *BinanceEndpoint) GetAccountTradeHistory(
	base, quote common.Token,
	fromID uint64,
	timepoint uint64) (exchange.BinaAccountTradeHistory, error) {

	symbol := strings.ToUpper(fmt.Sprintf("%s%s", base.ID, quote.ID))
	result := exchange.BinaAccountTradeHistory{}
	params := map[string]string{
		"symbol": symbol,
		"limit":  "500",
	}
	if fromID != 0 {
		params["fromId"] = strconv.FormatUint(fromID, 10)
	}
	resp_body, err := self.GetResponse(
		"GET",
		self.interf.AuthenticatedEndpoint()+"/api/v3/myTrades",
		params,
		true,
		timepoint,
	)
	if err == nil {
		json.Unmarshal(resp_body, &result)
	}
	return result, err
}

func (self *BinanceEndpoint) WithdrawHistory(startTime, endTime uint64) (exchange.Binawithdrawals, error) {
	result := exchange.Binawithdrawals{}
	resp_body, err := self.GetResponse(
		"GET",
		self.interf.AuthenticatedEndpoint()+"/wapi/v3/withdrawHistory.html",
		map[string]string{
			"startTime": fmt.Sprintf("%d", startTime),
			"endTime":   fmt.Sprintf("%d", endTime),
		},
		true,
		common.GetTimepoint(),
	)
	if err == nil {
		json.Unmarshal(resp_body, &result)
		if !result.Success {
			err = errors.New("Getting withdraw history from Binance failed: " + result.Msg)
		}
	}
	return result, err
}

func (self *BinanceEndpoint) DepositHistory(startTime, endTime uint64) (exchange.Binadeposits, error) {
	result := exchange.Binadeposits{}
	resp_body, err := self.GetResponse(
		"GET",
		self.interf.AuthenticatedEndpoint()+"/wapi/v3/depositHistory.html",
		map[string]string{
			"startTime": fmt.Sprintf("%d", startTime),
			"endTime":   fmt.Sprintf("%d", endTime),
		},
		true,
		common.GetTimepoint(),
	)
	if err == nil {
		err = json.Unmarshal(resp_body, &result)
		if !result.Success {
			err = errors.New("Getting deposit history from Binance failed: " + result.Msg)
		}
	}
	return result, err
}

func (self *BinanceEndpoint) CancelOrder(symbol string, id uint64) (exchange.Binacancel, error) {
	result := exchange.Binacancel{}
	resp_body, err := self.GetResponse(
		"DELETE",
		self.interf.AuthenticatedEndpoint()+"/api/v3/order",
		map[string]string{
			"symbol":  symbol,
			"orderId": fmt.Sprintf("%d", id),
		},
		true,
		common.GetTimepoint(),
	)
	if err == nil {
		json.Unmarshal(resp_body, &result)
		if result.Code != 0 {
			err = errors.New("Canceling order from Binance failed: " + result.Msg)
		}
	}
	return result, err
}

func (self *BinanceEndpoint) OrderStatus(symbol string, id uint64, timepoint uint64) (exchange.Binaorder, error) {
	result := exchange.Binaorder{}
	resp_body, err := self.GetResponse(
		"GET",
		self.interf.AuthenticatedEndpoint()+"/api/v3/order",
		map[string]string{
			"symbol":  symbol,
			"orderId": fmt.Sprintf("%d", id),
		},
		true,
		common.GetTimepoint(),
	)
	if err == nil {
		json.Unmarshal(resp_body, &result)
		if result.Code != 0 {
			err = errors.New(result.Msg)
		}
	}
	return result, err
}

func (self *BinanceEndpoint) Withdraw(token common.Token, amount *big.Int, address ethereum.Address, timepoint uint64) (string, error) {
	result := exchange.Binawithdraw{}
	resp_body, err := self.GetResponse(
		"POST",
		self.interf.AuthenticatedEndpoint()+"/wapi/v3/withdraw.html",
		map[string]string{
			"asset":   token.ID,
			"address": address.Hex(),
			"name":    "reserve",
			"amount":  strconv.FormatFloat(common.BigToFloat(amount, token.Decimal), 'f', -1, 64),
		},
		true,
		common.GetTimepoint(),
	)
	if err == nil {
		json.Unmarshal(resp_body, &result)
		if !result.Success {
			return "", errors.New(result.Msg)
		}
		return result.ID, nil
	} else {
		return "", errors.New(fmt.Sprintf("withdraw rejected by Binnace: %v", err))
	}
}

func (self *BinanceEndpoint) GetInfo(timepoint uint64) (exchange.Binainfo, error) {
	result := exchange.Binainfo{}
	resp_body, err := self.GetResponse(
		"GET",
		self.interf.AuthenticatedEndpoint()+"/api/v3/account",
		map[string]string{},
		true,
		common.GetTimepoint(),
	)
	if err == nil {
		json.Unmarshal(resp_body, &result)
	}
	if result.Code != 0 {
		return result, errors.New(fmt.Sprintf("Getting account info from Binance failed: %s", result.Msg))
	}
	return result, err
}

func (self *BinanceEndpoint) GetListenKey(timepoint uint64) (exchange.Binalistenkey, error) {
	result := exchange.Binalistenkey{}
	resp, err := self.GetResponse(
		"POST",
		self.interf.AuthenticatedEndpoint()+"/api/v1/userDataStream",
		map[string]string{},
		true,
		timepoint)
	if err == nil {
		json.Unmarshal(resp, &result)
	}
	return result, err
}

// SocketFetchOnePairData fetch one pair data from socket
func (self *BinanceEndpoint) SocketFetchOnePairData(
	pair common.TokenPair,
	data *sync.Map,
	exchangePriceChan chan *sync.Map) {

	URL := self.interf.SocketPublicEndpoint() + strings.ToLower(pair.Base.ID) + strings.ToLower(pair.Quote.ID) + "@depth"

	var dialer *websocket.Dialer

	conn, _, error := dialer.Dial(URL, nil)
	if error != nil {
		log.Printf("Cannot connect with socket %s\n", error)
		return
	}
	dataChannel := make(chan exchange.Orderbook)
	result := common.ExchangePrice{}
	result.Valid = true
	result.Timestamp = common.GetTimestamp()
	go func() {
		for {
			res := exchange.Binasocketresp{}
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println(err)
				return
			}
			json.Unmarshal(message, &res)
			dataChannel <- res
		}
	}()
	for {
		self.StoreOrderBookData(nil, pair, data, dataChannel, result)
		exchangePriceChan <- data
	}
}

func (self *BinanceEndpoint) SocketFetchAggTrade(
	pair common.TokenPair,
	dataChannel chan interface{}) {

	URL := self.interf.SocketPublicEndpoint() + strings.ToLower(pair.Base.ID) + strings.ToLower(pair.Quote.ID) + "@aggTrade"
	var dialer *websocket.Dialer
	conn, _, error := dialer.Dial(URL, nil)
	if error != nil {
		log.Printf("Cannot connect with socket %s\n", error)
		return
	}
	for {
		res := exchange.Binasocketaggtrade{}
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		json.Unmarshal(message, &res)
		dataChannel <- res
	}
}

func (self *BinanceEndpoint) SocketGetUser(dataChannel chan interface{}) {
	var dialer *websocket.Dialer
	timepoint := common.GetTimepoint()
	userStream, _ := self.GetListenKey(timepoint)
	URL := self.interf.SocketAuthenticatedEndpoint() + userStream.ListenKey
	conn, _, error := dialer.Dial(URL, nil)
	if error != nil {
		log.Printf("Cannot connect with socket %s\n", error)
		return
	}

	for {
		res := exchange.Binasocketuser{}
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		json.Unmarshal(message, &res)
		dataChannel <- res
	}
}

func (self *BinanceEndpoint) OpenOrdersForOnePair(
	pair common.TokenPair, timepoint uint64) (exchange.Binaorders, error) {

	result := exchange.Binaorders{}
	resp_body, err := self.GetResponse(
		"GET",
		self.interf.AuthenticatedEndpoint()+"/api/v3/openOrders",
		map[string]string{
			"symbol": pair.Base.ID + pair.Quote.ID,
		},
		true,
		common.GetTimepoint(),
	)
	if err != nil {
		return result, err
	} else {
		json.Unmarshal(resp_body, &result)
		return result, nil
	}
}

func (self *BinanceEndpoint) GetDepositAddress(asset string) (exchange.Binadepositaddress, error) {
	result := exchange.Binadepositaddress{}
	resp_body, err := self.GetResponse(
		"GET",
		self.interf.AuthenticatedEndpoint()+"/wapi/v3/depositAddress.html",
		map[string]string{
			"asset": asset,
		},
		true,
		common.GetTimepoint(),
	)
	if err == nil {
		err = json.Unmarshal(resp_body, &result)
		if !result.Success {
			err = errors.New(result.Msg)
		}
	}
	return result, err
}

func (self *BinanceEndpoint) GetExchangeInfo() (exchange.BinanceExchangeInfo, error) {
	result := exchange.BinanceExchangeInfo{}
	resp_body, err := self.GetResponse(
		"GET",
		self.interf.PublicEndpoint()+"/api/v1/exchangeInfo",
		map[string]string{},
		false,
		common.GetTimepoint(),
	)
	if err == nil {
		err = json.Unmarshal(resp_body, &result)
	}
	return result, err
}

func (self *BinanceEndpoint) GetServerTime() (uint64, error) {
	result := exchange.BinaServerTime{}
	resp_body, err := self.GetResponse(
		"GET",
		self.interf.PublicEndpoint()+"/api/v1/time",
		map[string]string{},
		false,
		common.GetTimepoint(),
	)
	if err == nil {
		err = json.Unmarshal(resp_body, &result)
	}
	return result.ServerTime, err
}

func (self *BinanceEndpoint) UpdateTimeDelta() error {
	currentTime := common.GetTimepoint()
	serverTime, err := self.GetServerTime()
	responseTime := common.GetTimepoint()
	if err != nil {
		return err
	}
	log.Printf("Binance current time: %s", currentTime)
	log.Printf("Binance server time: %s", serverTime)
	log.Printf("Binance response time: %s", responseTime)
	roundtripTime := (int64(responseTime) - int64(currentTime)) / 2
	self.timeDelta = int64(serverTime) - int64(currentTime) - roundtripTime

	log.Printf("Time delta: %s", self.timeDelta)
	return nil
}

func NewBinanceEndpoint(signer Signer, interf Interface) *BinanceEndpoint {
	endpoint := &BinanceEndpoint{signer, interf, 0}
	switch interf.(type) {
	case *SimulatedInterface:
		log.Println("Simulate environment, no updateTime called...")
	default:
		err := endpoint.UpdateTimeDelta()
		if err != nil {
			panic(err)
		}
	}
	return endpoint
}
