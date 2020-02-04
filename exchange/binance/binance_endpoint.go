package binance

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

	"github.com/KyberNetwork/reserve-data/cmd/deployment"
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/exchange"
	commonv3 "github.com/KyberNetwork/reserve-data/reservesetting/common"
)

// Endpoint object stand for Binance endpoint
// including signer for api call authentication,
// interf for calling api in different env
// timedelta to make sure calling api in time
type Endpoint struct {
	signer     Signer
	interf     Interface
	timeDelta  int64
	l          *zap.SugaredLogger
	exchangeID common.ExchangeID
}

func (ep *Endpoint) fillRequest(req *http.Request, signNeeded bool, timepoint uint64) {
	if req.Method == "POST" || req.Method == "PUT" || req.Method == "DELETE" {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("User-Agent", "binance/go")
	}
	req.Header.Add("Accept", "application/json")
	if signNeeded {
		q := req.URL.Query()
		sig := url.Values{}
		req.Header.Set("X-MBX-APIKEY", ep.signer.GetKey())
		q.Set("timestamp", fmt.Sprintf("%d", int64(timepoint)+ep.timeDelta-1000))
		q.Set("recvWindow", "5000")
		sig.Set("signature", ep.signer.Sign(q.Encode()))
		// Using separated values map for signature to ensure it is at the end
		// of the query. This is required for /wapi apis from binance without
		// any damn documentation about it!!!
		req.URL.RawQuery = q.Encode() + "&" + sig.Encode()
	}
}

// GetResponse call to binance endpoint and get response
func (ep *Endpoint) GetResponse(
	method string, url string,
	params map[string]string, signNeeded bool, timepoint uint64) ([]byte, error) {
	var (
		err      error
		respBody []byte
	)
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")

	q := req.URL.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()
	ep.fillRequest(req, signNeeded, timepoint)

	ep.l.Infof("request to binance: %s", req.URL)
	resp, err := client.Do(req)
	if err != nil {
		return respBody, err
	}
	defer func() {
		if cErr := resp.Body.Close(); cErr != nil {
			ep.l.Warnw("Response body close failed", "err", cErr)
		}
	}()
	switch resp.StatusCode {
	case http.StatusTooManyRequests:
		err = errors.New("breaking binance request rate limit")
	case http.StatusTeapot:
		err = errors.New("ip has been auto-banned by binance for continuing to send requests after receiving 429 codes")
	case http.StatusInternalServerError:
		err = errors.New("500 from Binance, its fault")
	case http.StatusUnauthorized:
		err = errors.New("binance api key not valid")
	case http.StatusOK:
		respBody, err = ioutil.ReadAll(resp.Body)
	default:
		var response exchange.Binaresp
		if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
			break
		}
		err = fmt.Errorf("binance return with code: %d - %s", resp.StatusCode, response.Msg)
	}
	if err != nil || len(respBody) == 0 {
		ep.l.Warnw("request got response from binance", "url", req.URL, "body", string(common.TruncStr(respBody)), "err", err)
	}
	return respBody, err
}

// GetDepthOnePair return list of orderbook for one pair of token
func (ep *Endpoint) GetDepthOnePair(baseID, quoteID string) (exchange.Binaresp, error) {

	respBody, err := ep.GetResponse(
		"GET", ep.interf.PublicEndpoint()+"/api/v3/depth",
		map[string]string{
			"symbol": fmt.Sprintf("%s%s", baseID, quoteID),
			"limit":  "100",
		},
		false,
		common.NowInMillis(),
	)

	respData := exchange.Binaresp{}
	if err != nil {
		return respData, err
	}
	if err = json.Unmarshal(respBody, &respData); err != nil {
		return respData, err
	}
	if respData.Code != 0 {
		return respData, fmt.Errorf("getting depth from Binance failed: %s", respData.Msg)
	}
	return respData, nil
}

// Trade Relevant params:
// symbol ("%s%s", base, quote)
// side (BUY/SELL)
// type (LIMIT/MARKET)
// timeInForce (GTC/IOC)
// quantity
// price
//
// In this version, we only support LIMIT order which means only buy/sell with acceptable price,
// and GTC time in force which means that the order will be active until it's implicitly canceled
func (ep *Endpoint) Trade(tradeType string, pair commonv3.TradingPairSymbols, rate, amount float64) (exchange.Binatrade, error) {
	result := exchange.Binatrade{}
	symbol := pair.BaseSymbol + pair.QuoteSymbol
	orderType := "LIMIT"
	params := map[string]string{
		"symbol":      symbol,
		"side":        strings.ToUpper(tradeType),
		"type":        orderType,
		"timeInForce": "GTC",
		"quantity":    strconv.FormatFloat(amount, 'f', -1, 64),
	}
	params["price"] = strconv.FormatFloat(rate, 'f', -1, 64)
	respBody, err := ep.GetResponse(
		"POST",
		ep.interf.AuthenticatedEndpoint()+"/api/v3/order",
		params,
		true,
		common.NowInMillis(),
	)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(respBody, &result)
	return result, err
}

// GetTradeHistory Return trade history
func (ep *Endpoint) GetTradeHistory(symbol string) (exchange.BinanceTradeHistory, error) {
	result := exchange.BinanceTradeHistory{}
	timepoint := common.NowInMillis()
	respBody, err := ep.GetResponse(
		"GET",
		ep.interf.PublicEndpoint()+"/api/v3/trades",
		map[string]string{
			"symbol": symbol,
			"limit":  "500",
		},
		false,
		timepoint,
	)
	if err == nil {
		err = json.Unmarshal(respBody, &result)
	}
	return result, err
}

// GetAccountTradeHistory return our account trades
func (ep *Endpoint) GetAccountTradeHistory(
	baseSymbol, quoteSymbol string,
	fromID string) (exchange.BinaAccountTradeHistory, error) {

	symbol := strings.ToUpper(fmt.Sprintf("%s%s", baseSymbol, quoteSymbol))
	result := exchange.BinaAccountTradeHistory{}
	params := map[string]string{
		"symbol": symbol,
		"limit":  "500",
	}
	if fromID != "" {
		params["fromId"] = fromID
	} else {
		params["fromId"] = "0"
	}
	respBody, err := ep.GetResponse(
		"GET",
		ep.interf.AuthenticatedEndpoint()+"/api/v3/myTrades",
		params,
		true,
		common.NowInMillis(),
	)
	if err == nil {
		err = json.Unmarshal(respBody, &result)
	}
	return result, err
}

func (ep *Endpoint) WithdrawHistory(startTime, endTime uint64) (exchange.Binawithdrawals, error) {
	result := exchange.Binawithdrawals{}
	respBody, err := ep.GetResponse(
		"GET",
		ep.interf.AuthenticatedEndpoint()+"/wapi/v3/withdrawHistory.html",
		map[string]string{
			"startTime": fmt.Sprintf("%d", startTime),
			"endTime":   fmt.Sprintf("%d", endTime),
		},
		true,
		common.NowInMillis(),
	)
	if err == nil {
		if err = json.Unmarshal(respBody, &result); err != nil {
			return result, err
		}
		if !result.Success {
			err = errors.New("Getting withdraw history from Binance failed: " + result.Msg)
		}
	}
	return result, err
}

func (ep *Endpoint) DepositHistory(startTime, endTime uint64) (exchange.Binadeposits, error) {
	result := exchange.Binadeposits{}
	respBody, err := ep.GetResponse(
		"GET",
		ep.interf.AuthenticatedEndpoint()+"/wapi/v3/depositHistory.html",
		map[string]string{
			"startTime": fmt.Sprintf("%d", startTime),
			"endTime":   fmt.Sprintf("%d", endTime),
		},
		true,
		common.NowInMillis(),
	)
	if err == nil {
		if err = json.Unmarshal(respBody, &result); err != nil {
			return result, err
		}
		if !result.Success {
			err = errors.New("Getting deposit history from Binance failed: " + result.Msg)
		}
	}
	return result, err
}

func (ep *Endpoint) CancelOrder(symbol string, id uint64) (exchange.Binacancel, error) {
	result := exchange.Binacancel{}
	respBody, err := ep.GetResponse(
		"DELETE",
		ep.interf.AuthenticatedEndpoint()+"/api/v3/order",
		map[string]string{
			"symbol":  symbol,
			"orderId": fmt.Sprintf("%d", id),
		},
		true,
		common.NowInMillis(),
	)
	if err == nil {
		if err = json.Unmarshal(respBody, &result); err != nil {
			return result, err
		}
		if result.Code != 0 {
			err = errors.New("Canceling order from Binance failed: " + result.Msg)
		}
	}
	return result, err
}

func (ep *Endpoint) OrderStatus(symbol string, id uint64) (exchange.Binaorder, error) {
	result := exchange.Binaorder{}
	respBody, err := ep.GetResponse(
		"GET",
		ep.interf.AuthenticatedEndpoint()+"/api/v3/order",
		map[string]string{
			"symbol":  symbol,
			"orderId": fmt.Sprintf("%d", id),
		},
		true,
		common.NowInMillis(),
	)
	if err == nil {
		if err = json.Unmarshal(respBody, &result); err != nil {
			return result, err
		}
		if result.Code != 0 {
			err = errors.New(result.Msg)
		}
	}
	return result, err
}

func (ep *Endpoint) Withdraw(asset commonv3.Asset, amount *big.Int, address ethereum.Address) (string, error) {
	var symbol string
	for _, exchg := range asset.Exchanges {
		if exchg.ExchangeID == uint64(ep.exchangeID) {
			symbol = exchg.Symbol
		}
	}
	result := exchange.Binawithdraw{}
	respBody, err := ep.GetResponse(
		"POST",
		ep.interf.AuthenticatedEndpoint()+"/wapi/v3/withdraw.html",
		map[string]string{
			"asset":   symbol,
			"address": address.Hex(),
			"name":    "reserve",
			"amount":  strconv.FormatFloat(common.BigToFloat(amount, int64(asset.Decimals)), 'f', -1, 64),
		},
		true,
		common.NowInMillis(),
	)
	if err == nil {
		if err = json.Unmarshal(respBody, &result); err != nil {
			return "", err
		}
		if !result.Success {
			return "", errors.New(result.Msg)
		}
		return result.ID, nil
	}
	return "", fmt.Errorf("withdraw rejected by Binnace: %v", err)
}

func (ep *Endpoint) GetInfo() (exchange.Binainfo, error) {
	result := exchange.Binainfo{}
	respBody, err := ep.GetResponse(
		"GET",
		ep.interf.AuthenticatedEndpoint()+"/api/v3/account",
		map[string]string{},
		true,
		common.NowInMillis(),
	)
	if err == nil {
		if err = json.Unmarshal(respBody, &result); err != nil {
			return result, err
		}
	}
	if result.Code != 0 {
		return result, fmt.Errorf("getting account info from Binance failed: %s", result.Msg)
	}
	return result, err
}

func (ep *Endpoint) OpenOrdersForOnePair(pair commonv3.TradingPairSymbols) (exchange.Binaorders, error) {

	result := exchange.Binaorders{}
	respBody, err := ep.GetResponse(
		"GET",
		ep.interf.AuthenticatedEndpoint()+"/api/v3/openOrders",
		map[string]string{
			"symbol": pair.BaseSymbol + pair.QuoteSymbol,
		},
		true,
		common.NowInMillis(),
	)
	if err != nil {
		return result, err
	}
	if err = json.Unmarshal(respBody, &result); err != nil {
		return result, err
	}
	return result, nil
}

func (ep *Endpoint) GetDepositAddress(asset string) (exchange.Binadepositaddress, error) {
	result := exchange.Binadepositaddress{}
	respBody, err := ep.GetResponse(
		"GET",
		ep.interf.AuthenticatedEndpoint()+"/wapi/v3/depositAddress.html",
		map[string]string{
			"asset": asset,
		},
		true,
		common.NowInMillis(),
	)
	if err == nil {
		if err = json.Unmarshal(respBody, &result); err != nil {
			return result, err
		}
		if !result.Success {
			err = errors.New(result.Msg)
		}
	}
	return result, err
}

// GetExchangeInfo return exchange info
// including base, quote precision for tokens
// min, max price, min notional
func (ep *Endpoint) GetExchangeInfo() (exchange.BinanceExchangeInfo, error) {
	result := exchange.BinanceExchangeInfo{}
	respBody, err := ep.GetResponse(
		"GET",
		ep.interf.PublicEndpoint()+"/api/v3/exchangeInfo",
		map[string]string{},
		false,
		common.NowInMillis(),
	)
	if err == nil {
		err = json.Unmarshal(respBody, &result)
	}
	return result, err
}

func (ep *Endpoint) getServerTime() (uint64, error) {
	result := exchange.BinaServerTime{}
	respBody, err := ep.GetResponse(
		"GET",
		ep.interf.PublicEndpoint()+"/api/v3/time",
		map[string]string{},
		false,
		common.NowInMillis(),
	)
	if err == nil {
		err = json.Unmarshal(respBody, &result)
	}
	return result.ServerTime, err
}

func (ep *Endpoint) UpdateTimeDelta() error {
	currentTime := common.NowInMillis()
	serverTime, err := ep.getServerTime()
	responseTime := common.NowInMillis()
	if err != nil {
		return err
	}

	roundtripTime := (int64(responseTime) - int64(currentTime)) / 2
	ep.timeDelta = int64(serverTime) - int64(currentTime) - roundtripTime
	ep.l.Infow("UpdateTimeDelta", "binance_time", currentTime,
		"server_time", serverTime, "response_time", responseTime, "time_delta", ep.timeDelta)
	return nil
}

//NewBinanceEndpoint return new endpoint instance for using binance
func NewBinanceEndpoint(signer Signer, interf Interface, dpl deployment.Deployment, exchangeID common.ExchangeID) *Endpoint {
	l := zap.S()
	endpoint := &Endpoint{signer: signer, interf: interf, l: l}
	switch dpl {
	case deployment.Simulation:
		l.Info("Simulate environment, no updateTime called...")
	default:
		err := endpoint.UpdateTimeDelta()
		if err != nil {
			panic(err)
		}
	}
	return endpoint
}
