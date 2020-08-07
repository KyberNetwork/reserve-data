package world

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data/common"
)

// TheWorld is the concrete implementation of fetcher.TheWorld interface.
type TheWorld struct {
	endpoint Endpoint
	l        *zap.SugaredLogger
}

// NewTheWorld ...
func NewTheWorld(ep Endpoint, l *zap.SugaredLogger) *TheWorld {
	return &TheWorld{
		endpoint: ep,
		l:        l,
	}
}

func (tw *TheWorld) getPublic(url string, dst interface{}) error {
	var (
		client = &http.Client{Timeout: 30 * time.Second}
	)
	caller := common.GetCallerFunctionName()

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	tw.l.Infof("get %s - %s", caller, url)

	req.Header.Add("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		tw.l.Warnw("request failed", "caller", caller, "err", err)
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		tw.l.Warnw("read response error", "caller", caller, "err", err)
		return errors.Wrap(err, "read response error")
	}

	if resp.StatusCode != http.StatusOK {
		tw.l.Warnw("unexpected status code", "code", resp.StatusCode, "body", string(common.TruncStr(body)))
		return errors.New("unexpected return code")
	}
	d := json.NewDecoder(bytes.NewBuffer(body))
	d.DisallowUnknownFields()
	if err = d.Decode(dst); err != nil {
		tw.l.Warnw("unmarshal failed", "caller", caller, "err", err, "body", string(common.TruncStr(body)))
		return errors.Wrap(err, "unmarshal failed")
	}
	return nil
}

func (tw *TheWorld) getOneForgeGoldUSDInfo() common.OneForgeGoldData {

	url := tw.endpoint.OneForgeGoldUSDDataEndpoint()
	var result common.OneForgeGoldData
	err := tw.getPublic(url, &result)
	if err != nil {
		result.Error = true
		result.Message = err.Error()
	}
	return result
}

func (tw *TheWorld) getOneForgeGoldETHInfo() common.OneForgeGoldData {
	url := tw.endpoint.OneForgeGoldETHDataEndpoint()
	result := common.OneForgeGoldData{}
	err := tw.getPublic(url, &result)
	if err != nil {
		result.Error = true
		result.Message = err.Error()
	}
	return result
}

func (tw *TheWorld) getDGXGoldInfo() common.DGXGoldData {
	url := tw.endpoint.GoldDataEndpoint()
	result := common.DGXGoldData{
		Valid: true,
	}
	err := tw.getPublic(url, &result)
	if err != nil {
		result.Error = err.Error()
		result.Valid = false
	} else {
		result.Valid = true
	}
	return result
}

func (tw *TheWorld) getGDAXGoldInfo() common.GDAXGoldData {
	url := tw.endpoint.GDAXDataEndpoint()
	result := common.GDAXGoldData{
		Valid: true,
	}
	err := tw.getPublic(url, &result)
	if err != nil {
		result.Error = err.Error()
		result.Valid = false
	} else {
		result.Valid = true
	}
	return result
}

func (tw *TheWorld) getKrakenGoldInfo() common.KrakenGoldData {
	url := tw.endpoint.KrakenDataEndpoint()
	result := common.KrakenGoldData{
		Valid: true,
	}
	err := tw.getPublic(url, &result)
	if err != nil {
		result.Error = err.Error()
		result.Valid = false
	} else {
		result.Valid = true
	}
	return result
}

func (tw *TheWorld) getGeminiGoldInfo() common.GeminiGoldData {
	url := tw.endpoint.GeminiDataEndpoint()
	result := common.GeminiGoldData{
		Valid: true,
	}
	err := tw.getPublic(url, &result)
	if err != nil {
		result.Error = err.Error()
		result.Valid = false
	} else {
		result.Valid = true
	}
	return result
}

func (tw *TheWorld) GetGoldInfo() (common.GoldData, error) {
	return common.GoldData{
		DGX:         tw.getDGXGoldInfo(),
		OneForgeETH: tw.getOneForgeGoldETHInfo(),
		OneForgeUSD: tw.getOneForgeGoldUSDInfo(),
		GDAX:        tw.getGDAXGoldInfo(),
		Kraken:      tw.getKrakenGoldInfo(),
		Gemini:      tw.getGeminiGoldInfo(),
	}, nil
}

func (tw *TheWorld) getFeedInfo(url string) common.FeedProviderResponse {
	var result common.FeedProviderResponse
	err := tw.getPublic(url, &result)
	if err != nil {
		result.Error = err.Error()
		result.Valid = false
	} else {
		result.Valid = true
	}
	return result
}
