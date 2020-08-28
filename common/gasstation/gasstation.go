package gasstation

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"go.uber.org/zap"
)

// ETHGas ...
type ETHGas struct {
	Fast          float64            `json:"fast"`
	Fastest       float64            `json:"fastest"`
	SafeLow       float64            `json:"safeLow"`
	Average       float64            `json:"average"`
	BlockTime     float64            `json:"block_time"`
	BlockNum      uint64             `json:"blockNum"`
	Speed         float64            `json:"speed"`
	SafeLowWait   float64            `json:"safeLowWait"`
	AvgWait       float64            `json:"avgWait"`
	FastestWait   float64            `json:"fastestWait"`
	GasPriceRange map[string]float64 `json:"gasPriceRange"`
}

// Client represent for gasStation client
type Client struct {
	client          *http.Client
	baseURL         string
	apiKey          string
	etherscanAPIKey string
	sugar           *zap.SugaredLogger
}

// New create a new Client object
func New(c *http.Client, apiKey, etherscanAPIKey string, sugar *zap.SugaredLogger) *Client {
	return &Client{
		client:          c,
		baseURL:         "https://ethgasstation.info",
		apiKey:          apiKey,
		etherscanAPIKey: etherscanAPIKey,
		sugar:           sugar,
	}
}

func (c *Client) doRequest(method, endpoint string, response interface{}) error {
	req, err := http.NewRequest(method, endpoint, nil)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("gasstation return code %d, data %s", resp.StatusCode, string(data))
	}
	err = json.Unmarshal(data, response)
	if err != nil {
		return fmt.Errorf("unmarshal gasstation error %v, for data %s", err, string(data))
	}
	return nil
}

// EtherscanGas gas price from etherscan gas tracker
type EtherscanGas struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  struct {
		LastBlock       string `json:"LastBlock"`
		SafeGasPrice    string `json:"SafeGasPrice"`
		ProposeGasPrice string `json:"ProposeGasPrice"`
		FastGasPrice    string `json:"FastGasPrice"`
	} `json:"result"`
}

func (c *Client) getGasFromEtherscan() (EtherscanGas, error) {
	var res EtherscanGas
	endpoint := fmt.Sprintf("https://api.etherscan.io/api?module=gastracker&action=gasoracle&apikey=%s", c.etherscanAPIKey)
	err := c.doRequest(http.MethodGet, endpoint, &res)
	return res, err
}

// ETHGas get gasstation gas data
func (c *Client) ETHGas() (ETHGas, error) {
	var (
		res       ETHGas
		logger    = c.sugar.With("func", "ETHGas")
		gasConfig EtherscanGas
	)
	endpoint := fmt.Sprintf("%s/json/ethgasAPI.json", c.baseURL)
	if c.apiKey != "" {
		endpoint += "?api-key=" + c.apiKey
	}
	err := c.doRequest(http.MethodGet, endpoint, &res)
	if err != nil {
		logger.Warnw("failed to get gas price from ethgasstation, fallback using etherscan api", "error", err)
		gasConfig, err = c.getGasFromEtherscan()
		if err != nil {
			return res, err
		}
		fast, err := strconv.ParseFloat(gasConfig.Result.FastGasPrice, 64)
		if err != nil {
			logger.Errorw("failed to parse fast gas from etherscan", "error", err)
			return res, err
		}
		safe, err := strconv.ParseFloat(gasConfig.Result.SafeGasPrice, 64)
		if err != nil {
			logger.Errorw("failed to parse safe gas from etherscan", "error", err)
			return res, err
		}
		propose, err := strconv.ParseFloat(gasConfig.Result.ProposeGasPrice, 64)
		if err != nil {
			logger.Errorw("failed to parse propose gas from etherscan", "error", err)
			return res, err
		}
		res = ETHGas{
			Fast:    fast,
			SafeLow: safe,
			Average: propose,
		}
	}
	return res, err
}
