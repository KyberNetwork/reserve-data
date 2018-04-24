package stat

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type FetcherRunnerTest struct {
	fr FetcherRunner
}

type Tickers []<-chan time.Time

func NewFetcherRunnerTest(fetcherRunner FetcherRunner) *FetcherRunnerTest {
	return &FetcherRunnerTest{fetcherRunner}
}

func (self *FetcherRunnerTest) TestFetcherConcurrency(nanosec int64) error {
	tickers := []func() <-chan time.Time{self.fr.GetBlockTicker,
		self.fr.GetLogTicker,
		self.fr.GetReserveRatesTicker,
		self.fr.GetTradeLogProcessorTicker,
		self.fr.GetCatLogProcessorTicker,
	}
	if err := self.fr.Start(); err != nil {
		return err
	}
	startTime := time.Now()
	var wg sync.WaitGroup
	for _, ticker := range tickers {
		wg.Add(1)
		go func(ticker func() <-chan time.Time) {
			defer wg.Done()
			t := <-ticker()
			log.Printf("got a signal after %v", t.Sub(startTime).Seconds)
		}(ticker)
	}
	wg.Wait()
	timeTook := time.Since(startTime).Nanoseconds()
	upperRange := nanosec + nanosec/10
	lowerRange := nanosec - nanosec/10
	if timeTook < lowerRange || timeTook > upperRange {
		return fmt.Errorf("expect ticker in between %d and %d nanosec, but it came in %d instead", lowerRange, upperRange, timeTook)
	}
	if err := self.fr.Stop(); err != nil {
		return err
	}
	return nil
}
