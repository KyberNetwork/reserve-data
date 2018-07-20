package stat

import (
	"fmt"
	"log"
	"sync"
	"time"
)

const tolerance int64 = 200 * int64(time.Millisecond)

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
			log.Printf("got a signal after %v", t.Sub(startTime).Seconds())
		}(ticker)
	}
	wg.Wait()
	timeTook := time.Since(startTime).Nanoseconds()
	upperRange := nanosec + tolerance
	lowerRange := nanosec - tolerance
	if timeTook < lowerRange || timeTook > upperRange {
		return fmt.Errorf("expect ticker in between %d and %d nanosec, but it came in %d instead", lowerRange, upperRange, timeTook)
	}
	if err := self.fr.Stop(); err != nil {
		return err
	}
	return nil
}

func (self *FetcherRunnerTest) TestIndividualTicker(ticker func() <-chan time.Time, nanosec int64) error {
	if err := self.fr.Start(); err != nil {
		return err
	}

	t := <-ticker()
	log.Printf("ticked: %s", t.String())
	if err := self.fr.Stop(); err != nil {
		return err
	}
	return nil
}

func (self *FetcherRunnerTest) TestBlockTicker(limit int64) error {
	if err := self.TestIndividualTicker(self.fr.GetBlockTicker, limit); err != nil {
		return fmt.Errorf("GetBlockTicker failed(%s) ", err)
	}
	return nil
}

func (self *FetcherRunnerTest) TestLogTicker(limit int64) error {
	if err := self.TestIndividualTicker(self.fr.GetLogTicker, limit); err != nil {
		return fmt.Errorf("GetLogTicker failed(%s) ", err)
	}
	return nil
}

func (self *FetcherRunnerTest) TestReserveRateTicker(limit int64) error {
	if err := self.TestIndividualTicker(self.fr.GetReserveRatesTicker, limit); err != nil {
		return fmt.Errorf("GetReserveRates ticker failed(%s) ", err)
	}
	return nil
}

func (self *FetcherRunnerTest) TestTradelogProcessorTicker(limit int64) error {
	if err := self.TestIndividualTicker(self.fr.GetCatLogProcessorTicker, limit); err != nil {
		return fmt.Errorf("GetCatLogProcessorTicker failed(%s) ", err)
	}
	return nil
}

func (self *FetcherRunnerTest) TestCatlogProcessorTicker(limit int64) error {
	if err := self.TestIndividualTicker(self.fr.GetCatLogProcessorTicker, limit); err != nil {
		return fmt.Errorf("GetCatLogProcessorTicker failed(%s) ", err)
	}
	return nil
}
