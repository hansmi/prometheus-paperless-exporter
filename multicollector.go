package main

import (
	"context"
	"log"
	"runtime"
	"sync"
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

	logger *log.Logger

	warningsDesc *prometheus.Desc

	members []multiCollectorMember
}

var _ prometheus.Collector = (*multiCollector)(nil)

func newMultiCollector(m ...multiCollectorMember) *multiCollector {
	return &multiCollector{
		logger: log.Default(),
		warningsDesc: prometheus.NewDesc("paperless_warnings_total",
			"Number of warnings generated while scraping metrics.",
			nil, nil),
		members: m,
	}
}

func (c *multiCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.warningsDesc

	for _, i := range c.members {
		i.describe(ch)
	}
}

func (c *multiCollector) collectWithWarnings(ctx context.Context, ch chan<- prometheus.Metric) error {
	var wg sync.WaitGroup

	collected := make(chan prometheus.Metric)

	defer func() {
		close(collected)

		wg.Wait()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		var warnings []error

		for m := range collected {
			if warning, ok := m.(*warning); ok && warning != nil {
				warnings = append(warnings, warning.err)
				continue
			}

			ch <- m
		}

		if len(warnings) > 0 {
			c.logger.Printf("Metrics collection warnings: %q", warnings)
		}

		ch <- prometheus.MustNewConstMetric(c.warningsDesc, prometheus.GaugeValue, float64(len(warnings)))
	}()

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(runtime.GOMAXPROCS(0))

	for _, i := range c.members {
		collect := i.collect
		g.Go(func() error { return collect(ctx, collected) })
	}

	return g.Wait()
}

func (c *multiCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	if c.timeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.timeout)
		defer cancel()
	}

	if err := c.collectWithWarnings(ctx, ch); err != nil {
		c.logger.Printf("Metrics collection failed: %v", err.Error())
		ch <- prometheus.NewInvalidMetric(
			prometheus.NewDesc("paperless_error", "Metrics collection failed", nil, nil),
			err)
	}
}
