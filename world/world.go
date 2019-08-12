package world

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data/cmd/deployment"
	"github.com/KyberNetwork/reserve-data/common"
)

//TheWorld is the concrete implementation of fetcher.TheWorld interface.
type TheWorld struct {
	logger   *zap.SugaredLogger
	endpoint Endpoint
}

func (tw *TheWorld) getOneForgeGoldUSDInfo() common.OneForgeGoldData {
	var (
		url   = tw.endpoint.OneForgeGoldUSDDataEndpoint()
		sugar = tw.logger.With("url", url, "func", "TheWorld.getOneForgeGoldUSDInfo")
	)
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")
	var err error
	var respBody []byte
	sugar.Infow("get gold feed")
	resp, err := client.Do(req)
	if err != nil {
		sugar.Errorw("failed to send HTTP request", "error", err)
		return common.OneForgeGoldData{
			Error:   true,
			Message: err.Error(),
		}
	}
	defer func() {
		if cErr := resp.Body.Close(); cErr != nil {
			sugar.Errorw("failed to close body", "error", cErr)
		}
	}()
	if resp.StatusCode != 200 {
		sugar.Errorw("unexpected status code", "StatusCode", resp.StatusCode)
		err = fmt.Errorf("gold feed returned with code: %d", resp.StatusCode)
		return common.OneForgeGoldData{
			Error:   true,
			Message: err.Error(),
		}
	}
	respBody, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		sugar.Errorw("failed to read resp body", "error", err)
		return common.OneForgeGoldData{
			Error:   true,
			Message: err.Error(),
		}
	}
	result := common.OneForgeGoldData{}
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		sugar.Errorw("failed to unmarshal body", "error", err)
		result.Error = true
		result.Message = err.Error()
	}
	sugar.Infow("success to get gold feed", "result", result)
	return result
}

func (tw *TheWorld) getOneForgeGoldETHInfo() common.OneForgeGoldData {
	var (
		url   = tw.endpoint.OneForgeGoldETHDataEndpoint()
		sugar = tw.logger.With("url", url, "func", "TheWorld.getOneForgeGoldETHInfo")
	)
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")
	var err error
	var respBody []byte
	sugar.Infow("get gold feed")
	resp, err := client.Do(req)
	if err != nil {
		sugar.Errorw("failed to send HTTP request", "error", err)
		return common.OneForgeGoldData{
			Error:   true,
			Message: err.Error(),
		}
	}
	defer func() {
		if cErr := resp.Body.Close(); cErr != nil {
			sugar.Errorw("failed to close body", "error", cErr)
		}
	}()
	if resp.StatusCode != 200 {
		sugar.Errorw("unexpected status code", "StatusCode", resp.StatusCode)
		err = fmt.Errorf("gold feed returned with code: %d", resp.StatusCode)
		return common.OneForgeGoldData{
			Error:   true,
			Message: err.Error(),
		}
	}
	respBody, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		sugar.Errorw("failed to read resp body", "error", err)
		return common.OneForgeGoldData{
			Error:   true,
			Message: err.Error(),
		}
	}
	result := common.OneForgeGoldData{}
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		sugar.Errorw("failed to unmarshal body", "error", err)
		result.Error = true
		result.Message = err.Error()
	}
	sugar.Infow("success to get gold feed", "result", result)
	return result
}

func (tw *TheWorld) getDGXGoldInfo() common.DGXGoldData {
	var (
		url   = tw.endpoint.GoldDataEndpoint()
		sugar = tw.logger.With("url", url, "func", "TheWorld.getDGXGoldInfo")
	)
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")
	var err error
	var respBody []byte
	sugar.Infow("get gold feed")
	resp, err := client.Do(req)
	if err != nil {
		sugar.Errorw(" failed to send HTTP request", "error", err)
		return common.DGXGoldData{
			Valid: false,
			Error: err.Error(),
		}
	}
	defer func() {
		if cErr := resp.Body.Close(); cErr != nil {
			sugar.Errorw("failed to close body", "error", cErr)
		}
	}()
	if resp.StatusCode != 200 {
		sugar.Errorw("unexpected status code", "StatusCode", resp.StatusCode)
		err = fmt.Errorf("gold feed returned with code: %d", resp.StatusCode)
		return common.DGXGoldData{
			Valid: false,
			Error: err.Error(),
		}
	}
	respBody, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		sugar.Errorw("failed to read resp body", "error", err)
		return common.DGXGoldData{
			Valid: false,
			Error: err.Error(),
		}
	}
	result := common.DGXGoldData{
		Valid: true,
	}
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		sugar.Errorw("failed to unmarshal body", "error", err)
		result.Valid = false
		result.Error = err.Error()
	}
	sugar.Infow("success to get gold feed", "result", result)
	return result
}

func (tw *TheWorld) getGDAXGoldInfo() common.GDAXGoldData {
	var (
		url   = tw.endpoint.GDAXDataEndpoint()
		sugar = tw.logger.With("url", url, "func", "TheWorld.getGDAXGoldInfo")
	)
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")
	var err error
	var respBody []byte
	sugar.Infow("get gold feed")
	resp, err := client.Do(req)
	if err != nil {
		sugar.Errorw("failed to send HTTP request", "error", err)
		return common.GDAXGoldData{
			Valid: false,
			Error: err.Error(),
		}
	}
	defer func() {
		if cErr := resp.Body.Close(); cErr != nil {
			sugar.Errorw("failed to close body", "error", cErr)
		}
	}()
	if resp.StatusCode != 200 {
		sugar.Errorw("unexpected status code", "StatusCode", resp.StatusCode)
		err = fmt.Errorf("gold feed returned with code: %d", resp.StatusCode)
		return common.GDAXGoldData{
			Valid: false,
			Error: err.Error(),
		}
	}
	respBody, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		sugar.Errorw("failed to read resp body", "error", err)
		return common.GDAXGoldData{
			Valid: false,
			Error: err.Error(),
		}
	}
	result := common.GDAXGoldData{
		Valid: true,
	}
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		sugar.Errorw("failed to unmarshal body", "error", err)
		result.Valid = false
		result.Error = err.Error()
	}
	sugar.Infow("success to get gold feed", "result", result)
	return result
}

func (tw *TheWorld) getKrakenGoldInfo() common.KrakenGoldData {
	var (
		url   = tw.endpoint.KrakenDataEndpoint()
		sugar = tw.logger.With("url", url, "func", "TheWorld.getKrakenGoldInfo")
	)
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")
	var err error
	var respBody []byte
	sugar.Infow("get gold feed")
	resp, err := client.Do(req)
	if err != nil {
		sugar.Errorw("failed to send HTTP request", "error", err)
		return common.KrakenGoldData{
			Valid: false,
			Error: err.Error(),
		}
	}
	defer func() {
		if cErr := resp.Body.Close(); cErr != nil {
			sugar.Errorw("failed to close body", "error", cErr)
		}
	}()
	if resp.StatusCode != 200 {
		sugar.Errorw("unexpected status code", "StatusCode", resp.StatusCode)
		err = fmt.Errorf("gold feed returned with code: %d", resp.StatusCode)
		return common.KrakenGoldData{
			Valid: false,
			Error: err.Error(),
		}
	}
	respBody, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		sugar.Errorw("failed to read resp body", "error", err)
		return common.KrakenGoldData{
			Valid: false,
			Error: err.Error(),
		}
	}
	result := common.KrakenGoldData{
		Valid: true,
	}
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		sugar.Errorw("failed to unmarshal body", "error", err)
		result.Valid = false
		result.Error = err.Error()
	}
	sugar.Infow("success to get gold feed", "result", result)
	return result
}

func (tw *TheWorld) getGeminiGoldInfo() common.GeminiGoldData {
	var (
		url   = tw.endpoint.GeminiDataEndpoint()
		sugar = tw.logger.With("url", url, "func", "TheWorld.getGeminiGoldInfo")
	)
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")
	var err error
	var respBody []byte
	sugar.Infow("get gold feed")
	resp, err := client.Do(req)
	if err != nil {
		sugar.Errorw("failed to send HTTP request", "error", err)
		return common.GeminiGoldData{
			Valid: false,
			Error: err.Error(),
		}
	}
	defer func() {
		if cErr := resp.Body.Close(); cErr != nil {
			sugar.Errorw("failed to close body", "error", cErr)
		}
	}()
	if resp.StatusCode != 200 {
		sugar.Errorw("unexpected status code", "StatusCode", resp.StatusCode)
		err = fmt.Errorf("gold feed returned with code: %d", resp.StatusCode)
		return common.GeminiGoldData{
			Valid: false,
			Error: err.Error(),
		}
	}
	respBody, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		sugar.Errorw("failed to read resp body", "error", err)
		return common.GeminiGoldData{
			Valid: false,
			Error: err.Error(),
		}
	}
	result := common.GeminiGoldData{
		Valid: true,
	}
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		sugar.Errorw("failed to unmarshal body", "error", err)
		result.Valid = false
		result.Error = err.Error()
	}
	sugar.Infow("success to get gold feed", "result", result)
	return result
}

func (tw *TheWorld) GetGoldInfo() (common.GoldData, error) {
	//TODO: Each function returns an error, the return error should be the combination of all errors
	return common.GoldData{
		DGX:         tw.getDGXGoldInfo(),
		OneForgeETH: tw.getOneForgeGoldETHInfo(),
		OneForgeUSD: tw.getOneForgeGoldUSDInfo(),
		GDAX:        tw.getGDAXGoldInfo(),
		Kraken:      tw.getKrakenGoldInfo(),
		Gemini:      tw.getGeminiGoldInfo(),
	}, nil
}

//NewTheWorld return new world instance
func NewTheWorld(logger *zap.SugaredLogger, dpl deployment.Deployment, keyfile string) (*TheWorld, error) {
	switch dpl {
	case deployment.Development,
		deployment.Production,
		deployment.Kovan,
		deployment.Staging,
		deployment.Ropsten,
		deployment.Analytic:
		// TODO: make key file a cli flag
		endpoint, err := NewRealEndpointFromFile(keyfile)
		if err != nil {
			return nil, err
		}
		return &TheWorld{
			endpoint: endpoint,
			logger:   logger,
		}, nil
	case deployment.Simulation:
		return &TheWorld{
			endpoint: SimulatedEndpoint{},
			logger:   logger,
		}, nil
	}
	panic("unsupported environment")
}
