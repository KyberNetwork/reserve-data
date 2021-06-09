package schedulejob

import (
	"fmt"
	"strconv"
	"time"

	"go.uber.org/zap"

	marketdatacli "github.com/KyberNetwork/reserve-data/lib/market-data"
	nh "github.com/KyberNetwork/reserve-data/lib/noauth-http"
	"github.com/KyberNetwork/reserve-data/reservesetting/common"
	"github.com/KyberNetwork/reserve-data/reservesetting/storage"
)

type ScheduleJob struct {
	l             *zap.SugaredLogger
	cli           *nh.Client
	s             storage.Interface
	settingURL    string
	marketDataCli *marketdatacli.Client
}

func NewScheduleJob(s storage.Interface, settingURL string, marketDataCli *marketdatacli.Client) *ScheduleJob {
	return &ScheduleJob{
		l:             zap.S(),
		cli:           nh.New(),
		s:             s,
		settingURL:    settingURL,
		marketDataCli: marketDataCli,
	}
}

func (sj *ScheduleJob) Run(interval time.Duration) {
	for {
		sj.ExecuteScheduleSettingChange()
		sj.ExecuteEligibleScheduleJob()
		time.Sleep(interval)
	}
}

func (sj *ScheduleJob) ExecuteEligibleScheduleJob() {
	l := sj.l.With("func", "ExecuteEligibleScheduleJob")
	jobs, err := sj.s.GetEligibleScheduleJob()
	if err != nil {
		l.Errorw("cannot get job from db", "err", err)
		return
	}
	for _, j := range jobs {
		_, err := sj.cli.DoReq(fmt.Sprintf("%s/%s", sj.settingURL, j.Endpoint), j.HTTPMethod, j.Data)
		if err != nil {
			sj.l.Errorw("failed to execute request", "err", err)
		}
		if err := sj.s.RemoveScheduleJob(j.ID); err != nil {
			sj.l.Errorw("failed to remove the job from db", "id", j.ID)
		}
	}
}

func (sj *ScheduleJob) ExecuteScheduleSettingChange() {
	l := sj.l.With("func", "ExecuteScheduleSettingChange")
	ids, err := sj.s.GetScheduleSettingChange()
	if err != nil {
		l.Errorw("cannot get job from db", "err", err)
		return
	}
	for _, id := range ids {
		additionalDataReturn, err := sj.s.ConfirmSettingChange(id, true)
		if err != nil {
			l.Errorw("cannot confirm setting change", "err", err)
			return
		}
		if err := sj.tryToAddFeed(additionalDataReturn); err != nil {
			l.Errorw("cannot add feed to market data", "err", err)
			return
		}
	}
}

func (sj *ScheduleJob) tryToAddFeed(data *common.AdditionalDataReturn) error {
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
