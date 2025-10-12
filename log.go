package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/hansmi/paperhooks/pkg/client"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"
)

type logClient interface {
	ListLogs(context.Context) ([]string, *client.Response, error)
	GetLog(context.Context, string) ([]client.LogEntry, *client.Response, error)
}

type logPosition struct {
	valid  bool
	time   time.Time
	module string
	level  string
}

func newLogPosition(e client.LogEntry) logPosition {
	return logPosition{
		valid:  true,
		time:   e.Time,
		module: e.Module,
		level:  e.Level,
	}
}

func (p logPosition) equal(e client.LogEntry) bool {
	return p.valid && e.Time.Equal(p.time) && e.Module == p.module && e.Level == p.level
}

type logCollector struct {
	cl logClient

	mu sync.Mutex

	seen     map[string]logPosition
	totalVec *prometheus.CounterVec
}

func newLogCollector(cl logClient) *logCollector {
	return &logCollector{
		cl: cl,

		seen: map[string]logPosition{},
		totalVec: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "paperless_log_entries_total",
			Help: `Best-effort count of log entries.`,
		}, []string{"name", "module", "level"}),
	}
}

func (c *logCollector) id() string {
	return "log"
}

func (c *logCollector) describe(ch chan<- *prometheus.Desc) {
	c.totalVec.Describe(ch)
}

func (c *logCollector) collectOne(ctx context.Context, name string) error {
	entries, _, err := c.cl.GetLog(ctx, name)
	if err != nil {
		var reqErr *client.RequestError

		if errors.As(err, &reqErr) && reqErr.StatusCode == http.StatusNotFound {
			return nil
		}

		return err
	}

	if len(entries) == 0 {
		return nil
	}

	entryLabels := func(e client.LogEntry) prometheus.Labels {
		return prometheus.Labels{
			"name":   name,
			"module": e.Module,
			"level":  strings.ToLower(e.Level),
		}
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	start := 0

	if seen := c.seen[name]; seen.valid {
		// Find the first log entry which hasn't been seen previously.
		for idx, entry := range entries {
			if seen.equal(entry) {
				start = idx + 1
				break
			}
		}
	}

	for _, entry := range entries[start:] {
		c.totalVec.With(entryLabels(entry)).Inc()
	}

	newest := entries[len(entries)-1]
	newest.Message = ""

	c.seen[name] = newLogPosition(newest)

	return nil
}

func (c *logCollector) collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	names, _, err := c.cl.ListLogs(ctx)
	if err != nil {
		return fmt.Errorf("listing log names: %w", err)
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(runtime.GOMAXPROCS(0))

	for _, name := range names {
		name := name

		g.Go(func() error {
			if err := c.collectOne(ctx, name); err != nil {
				return fmt.Errorf("log %s: %w", name, err)
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	c.totalVec.Collect(ch)

	return nil
}
