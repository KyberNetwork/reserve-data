package configuration

import (
	"testing"
	"time"

	"github.com/KyberNetwork/reserve-data/stat"
)

func SetupTickerTestForStatFetcherRunner(blockDuration,
	logDuration,
	rateDuration,
	tlogProcessDuration,
	clogProcessDuration time.Duration) (*stat.FetcherRunnerTest, error) {
	tickerRuner := stat.NewTickerRunner(blockDuration,
		logDuration,
		rateDuration,
		tlogProcessDuration,
		clogProcessDuration)
	return stat.NewFetcherRunnerTest(tickerRuner), nil
}

func doTickerforStatFetcherRunnerTest(blockDuration,
	logDuration,
	rateDuration,
	tlogProcessDuration,
	clogProcessDuration time.Duration,
	f func(tester *stat.FetcherRunnerTest, t *testing.T), t *testing.T) {
	tester, err := SetupTickerTestForStatFetcherRunner(blockDuration, logDuration, rateDuration, tlogProcessDuration, clogProcessDuration)
	if err != nil {
		t.Fatalf("Testing Ticker Runner as Controller Runner: init failed(%s)", err)
	}
	f(tester, t)
}

func TestTickerRunnerForStatFetcherRunner(t *testing.T) {
	doTickerforStatFetcherRunnerTest(1*time.Second, 1*time.Second, 1*time.Second, 1*time.Second, 1*time.Second, func(tester *stat.FetcherRunnerTest, t *testing.T) {
		if err := tester.TestBlockTicker(); err != nil {
			t.Fatalf("Testing Ticker Runner failed:(%s)", err)
		}
	}, t)
	doTickerforStatFetcherRunnerTest(1*time.Second, 1*time.Second, 1*time.Second, 1*time.Second, 1*time.Second, func(tester *stat.FetcherRunnerTest, t *testing.T) {
		if err := tester.TestLogTicker(); err != nil {
			t.Fatalf("Testing Ticker Runner failed:(%s)", err)
		}
	}, t)
	doTickerforStatFetcherRunnerTest(1*time.Second, 1*time.Second, 1*time.Second, 1*time.Second, 1*time.Second, func(tester *stat.FetcherRunnerTest, t *testing.T) {
		if err := tester.TestReserveRateTicker(); err != nil {
			t.Fatalf("Testing Ticker Runner failed:(%s)", err)
		}
	}, t)
	doTickerforStatFetcherRunnerTest(1*time.Second, 1*time.Second, 1*time.Second, 1*time.Second, 1*time.Second, func(tester *stat.FetcherRunnerTest, t *testing.T) {
		if err := tester.TestTradelogProcessorTicker(); err != nil {
			t.Fatalf("Testing Ticker Runner failed:(%s)", err)
		}
	}, t)
	doTickerforStatFetcherRunnerTest(1*time.Second, 1*time.Second, 1*time.Second, 1*time.Second, 1*time.Second, func(tester *stat.FetcherRunnerTest, t *testing.T) {
		if err := tester.TestCatlogProcessorTicker(); err != nil {
			t.Fatalf("Testing Ticker Runner failed:(%s)", err)
		}
	}, t)
}
