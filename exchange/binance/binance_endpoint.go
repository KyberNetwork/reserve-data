package binance

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
	"sync"
	"time"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/exchange"
	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/websocket"
)

const EPSILON float64 = 0.0000000001 // 10e-10

type BinanceEndpoint struct {
	signer Signer
	interf Interface
}

func (self *BinanceEndpoint) fillRequest(req *http.Request, signNeeded bool, timepoint uint64) {
	if req.Method == "POST" || req.Method == "PUT" || req.Method == "DELETE" {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("User-Agent", "binance/go")
	}
	req.Header.Add("Accept", "application/json")
	if signNeeded {
		q := req.URL.Query()
		sig := url.Values{}
		req.Header.Set("X-MBX-APIKEY", self.signer.GetBinanceKey())
		q.Set("timestamp", fmt.Sprintf("%d", timepoint-5000))
		q.Set("recvWindow", "7000")
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
	log.Printf("request: %s\n", req.URL.RawQuery)
	resp, err := client.Do(req)
	if err != nil {
		return resp_body, err
	} else {
		defer resp.Body.Close()
		resp_body, err = ioutil.ReadAll(resp.Body)
		log.Printf("response: %s\n", resp_body)
		return resp_body, err
	}
}

func (self *BinanceEndpoint) StoreOrderBookData(
	wg *sync.WaitGroup,
	pair common.TokenPair,
	data *sync.Map,
	dataChannel chan exchange.Orderbook) {

	defer wg.Done()
	orderBook := <-dataChannel
	bids := orderBook.GetBids()
	asks := orderBook.GetAsks()

	log.Printf("Get response from socket: %s\n", orderBook)

	result := common.ExchangePrice{}
	result.Timestamp = common.GetTimestamp()
	result.ReturnTime = common.GetTimestamp()
	result.Valid = true
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
	data.Store(pair.PairID(), result)
}

func (self *BinanceEndpoint) FetchOnePairData(
	wg *sync.WaitGroup,
	pair common.TokenPair,
	data *sync.Map,
	timepoint uint64) {

	defer wg.Done()
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
		timepoint,
	)

	returnTime := common.GetTimestamp()
	result.ReturnTime = returnTime

	if err != nil {
		result.Valid = false
		result.Error = err.Error()
	} else {
		resp_data := exchange.Binaresp{}
		json.Unmarshal(resp_body, &resp_data)
		if resp_data.Code != 0 || resp_data.Msg != "" {
			result.Valid = false
			result.Error = fmt.Sprintf("Code: %d, Msg: %s", resp_data.Code, resp_data.Msg)
		} else {
			for _, buy := range resp_data.Bids {
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
			for _, sell := range resp_data.Asks {
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
		}
	}
	data.Store(pair.PairID(), result)
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
func (self *BinanceEndpoint) Trade(tradeType string, base, quote common.Token, rate, amount float64, timepoint uint64) (string, float64, float64, bool, error) {
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
		timepoint,
	)

	if err != nil {
		log.Printf("Error: %s", err)
		return "", 0, 0, false, err
	} else {
		json.Unmarshal(resp_body, &result)
		done, remaining, finished, err := self.QueryOrder(
			base.ID+quote.ID,
			result.OrderID,
			timepoint+20,
		)
		id := fmt.Sprintf("%s_%s", strconv.FormatUint(result.OrderID, 10), symbol)
		return id, done, remaining, finished, err
	}
}

func (self *BinanceEndpoint) WithdrawHistory(startTime, endTime uint64) (exchange.Binawithdrawals, error) {
	result := exchange.Binawithdrawals{}
	timepoint := common.GetTimepoint()
	resp_body, err := self.GetResponse(
		"GET",
		self.interf.AuthenticatedEndpoint()+"/wapi/v3/withdrawHistory.html",
		map[string]string{
			"startTime": fmt.Sprintf("%d", startTime),
			"endTime":   fmt.Sprintf("%d", endTime),
		},
		true,
		timepoint,
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
	timepoint := common.GetTimepoint()
	resp_body, err := self.GetResponse(
		"GET",
		self.interf.AuthenticatedEndpoint()+"/wapi/v3/depositHistory.html",
		map[string]string{
			"startTime": fmt.Sprintf("%d", startTime),
			"endTime":   fmt.Sprintf("%d", endTime),
		},
		true,
		timepoint,
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
		timepoint,
	)
	if err == nil {
		json.Unmarshal(resp_body, &result)
		if result.Code != 0 {
			err = errors.New(result.Message)
		}
	}
	return result, err
}

func (self *BinanceEndpoint) QueryOrder(symbol string, id uint64, timepoint uint64) (done float64, remaining float64, finished bool, err error) {
	result, err := self.OrderStatus(symbol, id, timepoint)
	if err != nil {
		return 0, 0, false, err
	} else {
		done, _ := strconv.ParseFloat(result.ExecutedQty, 64)
		total, _ := strconv.ParseFloat(result.OrigQty, 64)
		return done, total - done, total-done < EPSILON, nil
	}
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
		timepoint,
	)
	if err == nil {
		json.Unmarshal(resp_body, &result)
		if result.Success == false {
			return "", errors.New(result.Message)
		}
		return result.ID, nil
	} else {
		log.Printf("Error: %v", err)
		return "", errors.New("withdraw rejected by Binnace")
	}
}

func (self *BinanceEndpoint) GetInfo(timepoint uint64) (exchange.Binainfo, error) {
	result := exchange.Binainfo{}
	resp_body, err := self.GetResponse(
		"GET",
		self.interf.AuthenticatedEndpoint()+"/api/v3/account",
		map[string]string{},
		true,
		timepoint,
	)
	if err == nil {
		json.Unmarshal(resp_body, &result)
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
	wg *sync.WaitGroup,
	pair common.TokenPair,
	data *sync.Map,
	dataChannel chan exchange.Orderbook) {

	URL := self.interf.SocketPublicEndpoint() + strings.ToLower(pair.Base.ID) + strings.ToLower(pair.Quote.ID) + "@depth"

	var dialer *websocket.Dialer

	conn, _, error := dialer.Dial(URL, nil)
	if error != nil {
		log.Printf("Cannot connect with socket %s\n", error)
		return
	}
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
	self.StoreOrderBookData(wg, pair, data, dataChannel)
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
	wg *sync.WaitGroup,
	pair common.TokenPair,
	data *sync.Map,
	timepoint uint64) {

	defer wg.Done()
	result := exchange.Binaorders{}
	resp_body, err := self.GetResponse(
		"GET",
		self.interf.AuthenticatedEndpoint()+"/api/v3/openOrders",
		map[string]string{
			"symbol": pair.Base.ID + pair.Quote.ID,
		},
		true,
		timepoint,
	)
	if err == nil {
		json.Unmarshal(resp_body, &result)
		orders := []common.Order{}
		for _, order := range result {
			price, _ := strconv.ParseFloat(order.Price, 64)
			orgQty, _ := strconv.ParseFloat(order.OrigQty, 64)
			executedQty, _ := strconv.ParseFloat(order.ExecutedQty, 64)
			orders = append(orders, common.Order{
				ID:          fmt.Sprintf("%s_%s%s", order.OrderId, strings.ToUpper(pair.Base.ID), strings.ToUpper(pair.Quote.ID)),
				Base:        strings.ToUpper(pair.Base.ID),
				Quote:       strings.ToUpper(pair.Quote.ID),
				OrderId:     fmt.Sprintf("%d", order.OrderId),
				Price:       price,
				OrigQty:     orgQty,
				ExecutedQty: executedQty,
				TimeInForce: order.TimeInForce,
				Type:        order.Type,
				Side:        order.Side,
				StopPrice:   order.StopPrice,
				IcebergQty:  order.IcebergQty,
				Time:        order.Time,
			})
		}
		data.Store(pair.PairID(), orders)
	} else {
		log.Printf("Unsuccessful response from Binance: %s", err)
	}
}

func (self *BinanceEndpoint) GetDepositAddress(asset string) (exchange.Binadepositaddress, error) {
	result := exchange.Binadepositaddress{}
	timepoint := common.GetTimepoint()
	resp_body, err := self.GetResponse(
		"GET",
		self.interf.AuthenticatedEndpoint()+"/wapi/v3/depositAddress.html",
		map[string]string{
			"asset": asset,
		},
		true,
		timepoint,
	)
	if err == nil {
		err = json.Unmarshal(resp_body, &result)
		if !result.Success {
			err = errors.New(result.Msg)
		}
	}
	return result, err
}

func NewBinanceEndpoint(signer Signer, interf Interface) *BinanceEndpoint {
	return &BinanceEndpoint{signer, interf}
}

func NewRealBinanceEndpoint(signer Signer) *BinanceEndpoint {
	return &BinanceEndpoint{signer, NewRealInterface()}
}

func NewSimulatedBinanceEndpoint(signer Signer) *BinanceEndpoint {
	return &BinanceEndpoint{signer, NewSimulatedInterface()}
}

func NewKovanBinanceEndpoint(signer Signer) *BinanceEndpoint {
	return &BinanceEndpoint{signer, NewKovanInterface()}
}

func NewDevBinanceEndpoint(signer Signer) *BinanceEndpoint {
	return &BinanceEndpoint{signer, NewDevInterface()}
}
