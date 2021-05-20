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

	"github.com/KyberNetwork/reserve-data/common/bcnetwork"
	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/rs/xid"
	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data/cmd/deployment"
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/exchange"
	authhttp "github.com/KyberNetwork/reserve-data/lib/auth-http"
	"github.com/KyberNetwork/reserve-data/lib/caller"
	"github.com/KyberNetwork/reserve-data/lib/rtypes"
	commonv3 "github.com/KyberNetwork/reserve-data/reservesetting/common"
)

// Endpoint object stand for Binance endpoint
// including signer for api call authentication,
// interf for calling api in different env
// timedelta to make sure calling api in time
type Endpoint struct {
	signer             Signer
	interf             Interface
	timeDelta          int64
	l                  *zap.SugaredLogger
	exchangeID         rtypes.ExchangeID
	client             *http.Client
	marketDataBaseURL  string
	accountDataBaseURL string
	accountID          string
	authHTTP           *authhttp.AuthHTTP
	network            string
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
	resp, err := ep.client.Do(req)
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
		"GET",
		fmt.Sprintf("%s/binance/%s-%s", ep.marketDataBaseURL, strings.ToLower(baseID), strings.ToLower(quoteID)),
		map[string]string{},
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
		"price":       strconv.FormatFloat(rate, 'f', -1, 64),
	}
	respBody, err := ep.authHTTP.DoReq(
		fmt.Sprintf("%s/api/v3/order/%s", ep.accountDataBaseURL, ep.accountID),
		http.MethodPost,
		params)
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

// WithdrawHistory get withdraw history from binance
func (ep *Endpoint) WithdrawHistory(startTime, endTime uint64) (exchange.Binawithdrawals, error) {
	result := exchange.Binawithdrawals{}
	respBody, err := ep.authHTTP.DoReq(
		fmt.Sprintf("%s/sapi/v1/capital/withdraw/history/%s", ep.accountDataBaseURL, ep.accountID),
		http.MethodGet,
		map[string]string{
			"startTime": fmt.Sprintf("%d", startTime),
			"endTime":   fmt.Sprintf("%d", endTime),
		})
	if err == nil {
		if err = json.Unmarshal(respBody, &result); err != nil {
			return result, err
		}
	}
	return result, err
}

// DepositHistory get deposit history from binance
func (ep *Endpoint) DepositHistory(startTime, endTime uint64) (exchange.Binadeposits, error) {
	result := exchange.Binadeposits{}
	respBody, err := ep.GetResponse(
		"GET",
		ep.interf.AuthenticatedEndpoint()+"/sapi/v1/capital/deposit/hisrec",
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
	}
	return result, err
}

// CancelOrder cancel an order from binance
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

// CancelAllOrders cancel all open order of an symbols
func (ep *Endpoint) CancelAllOrders(symbol string) ([]exchange.Binaorder, error) {
	var result []exchange.Binaorder
	respBody, err := ep.GetResponse(
		http.MethodDelete,
		ep.interf.AuthenticatedEndpoint()+"/api/v3/openOrders",
		map[string]string{
			"symbol": symbol,
		},
		true,
		common.NowInMillis(),
	)
	if err == nil {
		if err = json.Unmarshal(respBody, &result); err != nil {
			return result, err
		}
	}
	return result, err
}

// OrderStatus return status of orders
func (ep *Endpoint) OrderStatus(symbol string, id uint64) (exchange.Binaorder, error) {
	result := exchange.Binaorder{}
	respBody, err := ep.authHTTP.DoReq(
		fmt.Sprintf("%s/api/v3/order/%s", ep.accountDataBaseURL, ep.accountID),
		http.MethodGet,
		map[string]string{
			"symbol":  symbol,
			"orderId": fmt.Sprintf("%d", id),
		})
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

type transferResult struct {
	TransID string `json:"transId"`
}

func (ep *Endpoint) Transfer(fromAccount string, toAccount string, asset commonv3.Asset, amount *big.Int, runAsync bool, referenceID string) (string, error) {
	var symbol string
	for _, exchg := range asset.Exchanges {
		if exchg.ExchangeID == ep.exchangeID {
			symbol = exchg.Symbol
			break
		}
	}
	if symbol == "" {
		return "", fmt.Errorf("no symbol for asset %v", asset.ID)
	}
	result := transferResult{}
	respBody, err := ep.authHTTP.DoReq(
		fmt.Sprintf("%s/binance/transfer", ep.accountDataBaseURL),
		http.MethodPost,
		map[string]string{
			"run_async":    strconv.FormatBool(runAsync),
			"asset":        symbol,
			"from_account": fromAccount,
			"to_account":   toAccount,
			"amount":       strconv.FormatFloat(common.BigToFloat(amount, int64(asset.Decimals)), 'f', -1, 64),
			"reference_id": referenceID,
		})
	if err != nil {
		return "", fmt.Errorf("transfer rejected by Binance: %v", err)
	}
	if err = json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("unmarshal error %+v - %s", err, string(respBody))
	}
	return result.TransID, nil
}

// Withdraw token from binance
func (ep *Endpoint) Withdraw(asset commonv3.Asset, amount *big.Int, address ethereum.Address) (string, error) {
	var symbol string
	for _, exchg := range asset.Exchanges {
		if exchg.ExchangeID == ep.exchangeID {
			symbol = exchg.Symbol
		}
	}
	result := exchange.Binawithdraw{}
	withdrawOrderID := fmt.Sprintf("%s_%s", symbol, xid.New().String())
	respBody, err := ep.authHTTP.DoReq(
		fmt.Sprintf("%s/sapi/v1/capital/withdraw/apply/%s", ep.accountDataBaseURL, ep.accountID),
		http.MethodPost,
		map[string]string{
			"withdrawOrderId": withdrawOrderID,
			"asset":           symbol,
			"address":         address.Hex(),
			"name":            "reserve",
			"amount":          strconv.FormatFloat(common.BigToFloat(amount, int64(asset.Decimals)), 'f', -1, 64),
			"network":         ep.network,
		})
	if err == nil {
		if err = json.Unmarshal(respBody, &result); err != nil {
			return "", err
		}
		return withdrawOrderID, nil
	}
	return "", fmt.Errorf("withdraw rejected by Binance: %v", err)
}

// GetInfo return binance exchange info
func (ep *Endpoint) GetInfo() (exchange.Binainfo, error) {
	result := exchange.Binainfo{}
	respBody, err := ep.authHTTP.DoReq(
		fmt.Sprintf("%s/api/v3/account/%s", ep.accountDataBaseURL, ep.accountID),
		http.MethodGet,
		map[string]string{})
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

// GetMarginAccountInfo return binance exchange info
func (ep *Endpoint) GetMarginAccountInfo() (exchange.CrossMarginAccountDetails, error) {
	result := exchange.CrossMarginAccountDetails{}
	respBody, err := ep.authHTTP.DoReq(
		fmt.Sprintf("%s/binance/margin-account/%s", ep.accountDataBaseURL, ep.accountID),
		http.MethodGet,
		map[string]string{})
	if err == nil {
		if err = json.Unmarshal(respBody, &result); err != nil {
			return result, err
		}
	}
	return result, err
}

// OpenOrdersForOnePair return list open orders for one pair of token
func (ep *Endpoint) OpenOrdersForOnePair(pair *commonv3.TradingPairSymbols) (exchange.Binaorders, error) {
	var (
		result = exchange.Binaorders{}
		logger = ep.l.With("func", caller.GetCurrentFunctionName())
		params = make(map[string]string)
	)
	if pair != nil {
		logger.Infow("getting open order for pair", "pair", pair.BaseSymbol+pair.QuoteSymbol)
		params["symbol"] = pair.BaseSymbol + pair.QuoteSymbol
	}
	respBody, err := ep.authHTTP.DoReq(
		fmt.Sprintf("%s/api/v3/openOrders/%s", ep.accountDataBaseURL, ep.accountID),
		http.MethodGet,
		params)
	if err != nil {
		return result, err
	}
	if err = json.Unmarshal(respBody, &result); err != nil {
		return result, err
	}
	return result, nil
}

// GetDepositAddress of an asset
func (ep *Endpoint) GetDepositAddress(coin, network string) (exchange.CoinDepositAddress, error) {
	result := exchange.CoinDepositAddress{}
	respBody, err := ep.GetResponse(
		"GET",
		ep.interf.AuthenticatedEndpoint()+"/sapi/v1/capital/deposit/address",
		map[string]string{
			"coin":    coin,
			"network": network,
		},
		true,
		common.NowInMillis(),
	)
	if err == nil {
		if err = json.Unmarshal(respBody, &result); err != nil {
			return result, err
		}
	}
	return result, err
}

// GetAllAssetDetail all asset detail on binance
func (ep *Endpoint) GetAllAssetDetail() (map[string]exchange.BinanceAssetDetail, error) {
	resp := struct {
		Success     bool                                   `json:"success"`
		Msg         string                                 `json:"msg"`
		AssetDetail map[string]exchange.BinanceAssetDetail `json:"assetDetail"`
	}{}
	respBody, err := ep.GetResponse(
		"GET",
		ep.interf.AuthenticatedEndpoint()+"/sapi/v1/asset/assetDetail",
		map[string]string{},
		true,
		common.NowInMillis(),
	)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("cannot unmarshal asset info data from binance, err = %s", err)
	}
	if !resp.Success {
		return nil, fmt.Errorf("failed to get asset detail, msg: %s", resp.Msg)
	}
	return resp.AssetDetail, nil
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

// CoinInfo ...
type CoinInfo struct {
	Coin             string `json:"coin"`
	DepositAllEnable bool   `json:"depositAllEnable"`
	Free             string `json:"free"`
	Freeze           string `json:"freeze"`
	IPOable          string `json:"ipoable"`
	IsLegalMoney     bool   `json:"isLegalMoney"`
	Locked           string `json:"locked"`
	Name             string `json:"name"`
	NetworkList      []struct {
		AddressRegex       string `json:"addressRegex"`
		Coin               string `json:"coin"`
		DepositDesc        string `json:"depositDesc"`
		DepositEnable      bool   `json:"depositEnable"`
		IsDefault          bool   `json:"isDefault"`
		MinConfirm         int64  `json:"minConfirm"`
		Name               string `json:"name"`
		Network            string `json:"network"`
		ResetAddressStatus bool   `json:"resetAddressStatus"`
		SpecialTips        string `json:"specialTips"`
		UnLockConfirm      int64  `json:"unLockConfirm"`
		WithdrawDesc       string `json:"withdrawDesc"`
		WithdrawEnable     bool   `json:"withdrawEnable"`
		WithdrawFee        string `json:"withdrawFee"`
		WithdrawMin        string `json:"withdrawMin"`
	} `json:"networkList"`
}

// GetAllAssetWithdrawStatus return all asset withdraw status
func (ep *Endpoint) GetAllAssetWithdrawStatus() (map[string]bool, error) {
	var (
		response []CoinInfo
		result   = make(map[string]bool)
	)
	respBody, err := ep.authHTTP.DoReq(
		fmt.Sprintf("%s/binance/all-coin-info", ep.accountDataBaseURL),
		http.MethodGet,
		make(map[string]string))
	if err != nil {
		return result, err
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return result, err
	}
	for _, ci := range response {
		for _, n := range ci.NetworkList {
			if n.Network == ep.network {
				result[ci.Coin] = n.WithdrawEnable
				break
			}
		}
	}
	return result, nil
}

// GetAssetWithdrawFee ...
func (ep *Endpoint) GetAssetWithdrawFee(coin string) (float64, error) {
	var (
		response []CoinInfo
	)
	respBody, err := ep.authHTTP.DoReq(
		fmt.Sprintf("%s/binance/all-coin-info", ep.accountDataBaseURL),
		http.MethodGet,
		make(map[string]string),
	)
	if err != nil {
		return 0, err
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return 0, err
	}
	for _, ci := range response {
		if ci.Coin == coin {
			for _, n := range ci.NetworkList {
				if n.Network == ep.network {
					withdrawFee, err := strconv.ParseFloat(n.WithdrawFee, 64)
					return withdrawFee, err
				}
			}
		}
	}
	return 0, nil
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

// UpdateTimeDelta update the time delta of the request
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

// NewBinanceEndpoint return new endpoint instance for using binance
func NewBinanceEndpoint(signer Signer, interf Interface, dpl deployment.Deployment, client *http.Client, exparam rtypes.ExchangeID,
	marketDataBaseURL, accountDataBaseURL, accountID string, authHTTP *authhttp.AuthHTTP) *Endpoint {
	l := zap.S()
	nw := bcnetwork.GetPreConfig().Network
	endpoint := &Endpoint{
		signer:             signer,
		interf:             interf,
		l:                  l,
		client:             client,
		exchangeID:         exparam,
		marketDataBaseURL:  marketDataBaseURL,
		accountDataBaseURL: accountDataBaseURL,
		accountID:          accountID,
		authHTTP:           authHTTP,
		network:            nw,
	}
	switch dpl {
	case deployment.Simulation:
		l.Info("Simulate environment, no updateTime called...")
	default:
		err := endpoint.UpdateTimeDelta()
		if err != nil {
			l.Errorw("failed to update timeDelta", "err", err)
		}
	}
	return endpoint
}
