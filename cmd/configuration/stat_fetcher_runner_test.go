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

func doTickerforStatFetcherRunnerTest(f func(tester *stat.FetcherRunnerTest, t *testing.T), t *testing.T) {
	tester, err := SetupTickerTestForStatFetcherRunner(1*time.Second, 2*time.Second, 3*time.Second, 4*time.Second, 5*time.Second)
	if err != nil {
		t.Fatalf("Testing Ticker Runner as Controller Runner: init failed(%s)", err)
	}
	f(tester, t)
}

func TestTickerRunnerForStatFetcherRunner(t *testing.T) {
	doTickerforStatFetcherRunnerTest(func(tester *stat.FetcherRunnerTest, t *testing.T) {
		if err := tester.TestFetcherConcurrency(TESTDURATION.Nanoseconds()); err != nil {
			t.Fatalf("Testing Ticker Runner ")
		}
	}, t)
}
