package http_runner

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/KyberNetwork/reserve-data/common"
)

func testTicker(t *testing.T, ch <-chan time.Time, path string, port int) {
	t.Helper()

	t.Logf("testing ticker for path: %s", path)

	now := common.GetTimepoint()

	go func(ch <-chan time.Time) {
		ts := <-ch
		t.Logf("got timestamp: %s for path: %s", ts, path)
		if !ts.Equal(common.TimepointToTime(now)) {
			t.Errorf("wrong timestamp received, expected: %s, got: %s", common.TimepointToTime(now), ts)
		}
	}(ch)

	client := &http.Client{Timeout: time.Second}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/%s", port, path), nil)
	if err != nil {
		t.Fatal(err)
	}

	q := req.URL.Query()
	q.Add("timestamp", fmt.Sprintf("%d", now))
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("invalid HTTP return code: %d", resp.StatusCode)
	}
}

func TestHttpRunner(t *testing.T) {
	runner, err := NewHttpRunner()
	if err != nil {
		t.Fatal(err)
	}

	var tests = []struct {
		ch   <-chan time.Time
		path string
	}{
		{
			ch:   runner.GetOrderbookTicker(),
			path: "otick",
		},
		{
			ch:   runner.GetAuthDataTicker(),
			path: "atick",
		},
		{
			ch:   runner.GetRateTicker(),
			path: "rtick",
		},
		{
			ch:   runner.GetBlockTicker(),
			path: "btick",
		},
		{
			ch:   runner.GetTradeHistoryTicker(),
			path: "ttick",
		},
		{
			ch:   runner.GetGlobalDataTicker(),
			path: "gtick",
		},
	}

	for _, tc := range tests {
		testTicker(t, tc.ch, tc.path, runner.port)
	}
}
