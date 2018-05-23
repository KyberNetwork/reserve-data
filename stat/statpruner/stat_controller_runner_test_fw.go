package statpruner

import (
	"fmt"
	"time"
)

type ControllerRunnerTest struct {
	cr ControllerRunner
}

func NewControllerRunnerTest(controllerRunner ControllerRunner) *ControllerRunnerTest {
	return &ControllerRunnerTest{controllerRunner}
}

func (self *ControllerRunnerTest) TestAnalyticStorageControlTicker(nanosec int64) error {
	if err := self.cr.Start(); err != nil {
		return err
	}
	startTime := time.Now()
	t := <-self.cr.GetAnalyticStorageControlTicker()
	timeTook := t.Sub(startTime).Nanoseconds()
	upperRange := nanosec + 5*time.Millisecond.Nanoseconds()
	lowerRange := nanosec - 5*time.Millisecond.Nanoseconds()
	if timeTook < lowerRange || timeTook > upperRange {
		return fmt.Errorf("expect ticker in between %d and %d nanosec, but it came in %d instead", lowerRange, upperRange, timeTook)
	}
	if err := self.cr.Stop(); err != nil {
		return err
	}
	return nil
}
