package testutil

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func CollectAndCompare(t *testing.T, c prometheus.Collector, want string, metricNames ...string) {
	t.Helper()

	if err := testutil.CollectAndCompare(c, strings.NewReader(want), metricNames...); err != nil {
		t.Error(err)
	}
}

func DiscardMetrics(t *testing.T) chan<- prometheus.Metric {
	t.Helper()

	ch := make(chan prometheus.Metric)

	t.Cleanup(func() {
		close(ch)
	})

	go func() {
		for range ch {
		}
	}()

	return ch
}
