package configuration

import (
	"testing"
	"time"

	"github.com/KyberNetwork/reserve-data/stat/statpruner"
)

const (
	statControllerTestDuration time.Duration = time.Second * 1
)

func SetupTickerTestForControllerRunner(duration time.Duration) (*statpruner.ControllerRunnerTest, error) {
	tickerRuner := statpruner.NewControllerTickerRunner(duration, duration)
	return statpruner.NewControllerRunnerTest(tickerRuner), nil
}

func doTickerforControllerRunnerTest(f func(tester *statpruner.ControllerRunnerTest, t *testing.T), t *testing.T) {
	tester, err := SetupTickerTestForControllerRunner(statControllerTestDuration)
	if err != nil {
		t.Fatalf("Testing Ticker Runner as Controller Runner: init failed(%s)", err)
	}
	f(tester, t)
}

func TestTickerRunnerForControllerRunner(t *testing.T) {
	doTickerforControllerRunnerTest(func(tester *statpruner.ControllerRunnerTest, t *testing.T) {
		if err := tester.TestAnalyticStorageControlTicker(statControllerTestDuration.Nanoseconds()); err != nil {
			t.Fatalf("Testing Ticker Runner ")
		}
	}, t)
}
