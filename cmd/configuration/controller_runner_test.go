package configuration

import (
	"testing"
	"time"

	"github.com/KyberNetwork/reserve-data/stat"
)

const (
	TESTDURATION time.Duration = time.Second * 5
)

func SetupTickerTestForControllerRunner(duration time.Duration) (*stat.ControllerRunnerTest, error) {
	tickerRuner := stat.NewControllerTickerRunner(duration)
	return stat.NewControllerRunnerTest(tickerRuner), nil
}

func doTickerforControllerRunnerTest(f func(tester *stat.ControllerRunnerTest, t *testing.T), t *testing.T) {
	tester, err := SetupTickerTestForControllerRunner(TESTDURATION)
	if err != nil {
		t.Fatalf("Testing Ticker Runner as Controller Runner: init failed(%s)", err)
	}
	f(tester, t)
}

func TestTickerRunnerForControllerRunner(t *testing.T) {
	doTickerforControllerRunnerTest(func(tester *stat.ControllerRunnerTest, t *testing.T) {
		if err := tester.TestAnalyticStorageControlTicker(TESTDURATION.Nanoseconds()); err != nil {
			t.Fatalf("Testing Ticker Runner ")
		}
	}, t)
}
