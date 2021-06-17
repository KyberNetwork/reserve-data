package scheduledjob

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strconv"
	"time"

	"go.uber.org/zap"

	marketdatacli "github.com/KyberNetwork/reserve-data/lib/market-data"
	"github.com/KyberNetwork/reserve-data/reservesetting/common"
	"github.com/KyberNetwork/reserve-data/reservesetting/storage"
)

type ScheduledJob struct {
	l             *zap.SugaredLogger
	s             storage.Interface
	marketDataCli *marketdatacli.Client
	httpHandler   http.Handler
}

func NewScheduledJob(s storage.Interface, marketDataCli *marketdatacli.Client, httpHandler http.Handler) *ScheduledJob {
	return &ScheduledJob{
		l:             zap.S(),
		s:             s,
		marketDataCli: marketDataCli,
		httpHandler:   httpHandler,
	}
}

func (sj *ScheduledJob) Run(interval time.Duration) {
	for {
		sj.ExecuteScheduledSettingChange()
		sj.ExecuteEligibleScheduledJob()
		time.Sleep(interval)
	}
}

func (sj *ScheduledJob) ExecuteEligibleScheduledJob() {
	l := sj.l.With("func", "ExecuteEligibleScheduledJob")
	jobs, err := sj.s.GetEligibleScheduledJob()
	if err != nil {
		l.Errorw("cannot get job from db", "err", err)
		return
	}
	for _, j := range jobs {
		l.Infow("job info", "info", j)
		req, err := http.NewRequest(j.HTTPMethod, j.Endpoint, bytes.NewReader(j.Data))
		if err != nil {
			sj.l.Errorw("failed to build request", "err", err)
			continue
		}
		req.Header.Add("Content-Type", "application/json")
		res := httptest.NewRecorder()
		sj.httpHandler.ServeHTTP(res, req)
		if res.Code != http.StatusOK {
			l.Errorw("received unexpected code", "code", res.Code, "job id", j.ID)
		}
		if err := sj.s.UpdateScheduledJobStatus("done", j.ID); err != nil {
			l.Errorw("failed to update job's status", "job id", j.ID)
		}
	}
}

func (sj *ScheduledJob) ExecuteScheduledSettingChange() {
	l := sj.l.With("func", "ExecuteScheduledSettingChange")
	ids, err := sj.s.GetScheduledSettingChange()
	if err != nil {
		l.Errorw("cannot get job from db", "err", err)
		return
	}
	if len(ids) > 0 {
		l.Infow("list scheduled setting change will be executed", "ids", ids)
	}
	for _, id := range ids {
		additionalDataReturn, err := sj.s.ConfirmSettingChange(id, true)
		if err != nil {
			l.Errorw("cannot confirm setting change", "err", err, "id", id)
			return
		}
		if err := sj.tryToAddFeed(additionalDataReturn); err != nil {
			l.Errorw("cannot add feed to market data", "err", err)
			return
		}
	}
}

func (sj *ScheduledJob) tryToAddFeed(data *common.AdditionalDataReturn) error {
	if sj.marketDataCli != nil {
		// add pair to market data
		for _, tpID := range data.AddedTradingPairs {
			tradingPair, err := sj.s.GetTradingPair(tpID, false)
			if err != nil {
				sj.l.Errorw("cannot get trading pair", "id", tpID)
				return err
			}
			exchange, sourceSymbol, publicSymbol, err := common.DataForMarketDataByExchange(tradingPair.ExchangeID, tradingPair.BaseSymbol, tradingPair.QuoteSymbol)
			if err != nil {
				return err
			}
			if err := sj.marketDataCli.AddFeed(exchange, sourceSymbol, publicSymbol, strconv.FormatInt(int64(tpID), 10)); err != nil {
				sj.l.Errorw("cannot add feed to market data", "err", err, "exchange", exchange, "source symbol", sourceSymbol)
				continue
			}
		}
	}
	return nil
}
