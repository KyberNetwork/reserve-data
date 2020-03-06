package collector

import (
	"time"

	"github.com/KyberNetwork/reserve-data/feed-provider/fetcher"
	"github.com/KyberNetwork/reserve-data/feed-provider/storage"
	"golang.org/x/sync/errgroup"

	"go.uber.org/zap"
)

type Collector struct {
	s        storage.Storage
	fetchers map[string]fetcher.Fetcher
	sugar    *zap.SugaredLogger
	duration time.Duration
}

func (c *Collector) collectData() error {
	var (
		g           errgroup.Group
		resourcesCh = make(chan struct{}, 10) // resources limiter, thread need to acquire release resource
	)
	for n, f := range c.fetchers {
		var name, fetcher = n, f
		g.Go(
			func() error {
				resourcesCh <- struct{}{}
				defer func() { <-resourcesCh }()
				data := fetcher.GetData()
				c.sugar.Debugw("Saving data", "feed", name, "data", data)
				return c.s.Save(name, data)
			})
	}
	return g.Wait()
}

func (c *Collector) Run() {
	for range time.NewTicker(c.duration).C {
		err := c.collectData()
		if err != nil {
			c.sugar.Errorw("Error while collect feed data", "err", err)
		}
	}
}
