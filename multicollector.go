package main

import (
	"context"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"
)

type multiCollectorMember interface {
	describe(chan<- *prometheus.Desc)
	collect(context.Context, chan<- prometheus.Metric) error
}

type multiCollector struct {
	// Impose a timeout on collection if non-zero.
	timeout time.Duration

	members []multiCollectorMember
}

var _ prometheus.Collector = (*multiCollector)(nil)

func newMultiCollector(m ...multiCollectorMember) *multiCollector {
	return &multiCollector{
		members: m,
	}
}

func (c *multiCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, i := range c.members {
		i.describe(ch)
	}
}

func (c *multiCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	if c.timeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.timeout)
		defer cancel()
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(runtime.GOMAXPROCS(0))

	for _, i := range c.members {
		collect := i.collect
		g.Go(func() error { return collect(ctx, ch) })
	}

	if err := g.Wait(); err != nil {
		ch <- prometheus.NewInvalidMetric(
			prometheus.NewDesc("paperless_error", "Metrics collection failed", nil, nil),
			err)
	}
}
