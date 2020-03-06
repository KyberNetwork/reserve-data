package coinbase

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/KyberNetwork/reserve-data/feed-provider/common"
	"go.uber.org/zap"
)

type Fetcher struct {
	sugar         *zap.SugaredLogger
	endpoint      string
	requireAmount float64
	client        *http.Client
}

// GetResponse call to binance endpoint and get response
func (f *Fetcher) GetResponse(method string, url string, params map[string]string) ([]byte, error) {
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

	f.sugar.Infof("request to coinbase: %s", req.URL)
	resp, err := f.client.Do(req)
	if err != nil {
		return respBody, err
	}
	respBody, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return respBody, err
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return respBody, &common.RespError{
			Msg:        "coinbase return not OK code",
			StatusCode: resp.StatusCode,
			Body:       respBody,
		}
	}
	return respBody, err
}

func (f *Fetcher) getData() (Resp, error) {
	respBody, err := f.GetResponse(http.MethodGet, f.endpoint, nil)
	respData := Resp{}
	if err != nil {
		return respData, err
	}
	if err = json.Unmarshal(respBody, &respData); err != nil {
		return respData, err
	}
	return respData, nil
}

// GetData get orderbook from coinbase and convert to feed data
func (f *Fetcher) GetData() common.Feed {
	resp, err := f.getData()
	if err != nil {
		f.sugar.Errorw("Get error while get coinbase feed", "error", err)
		return common.Feed{
			Error: err,
		}
	}
	f.sugar.Debugw("Response from coinbase", "resp", resp)
	return resp.toFeed(f.requireAmount)
}
