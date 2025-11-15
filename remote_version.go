package main

import (
	"context"
	"sync"
	"time"

	"github.com/hansmi/paperhooks/pkg/client"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	remoteVersionCollectorID = "remote_version"
)

type remoteVersionClient interface {
	GetRemoteVersion(ctx context.Context) (*client.RemoteVersion, *client.Response, error)
}

type remoteVersionCollector struct {
	cl remoteVersionClient
	// ticker-based asynchronous fetching
	interval time.Duration

	// cached values updated by background goroutine
	cachedUpdateAvailable float64
	cachedVersion         string
	lastErr               error
	mu                    sync.RWMutex

	updateAvailableDesc *prometheus.Desc
}

// newRemoteVersionCollector creates a collector which polls the remote version
// at the provided interval. Interval is in seconds when provided from flags.
func newRemoteVersionCollector(cl remoteVersionClient, interval time.Duration) *remoteVersionCollector {
	c := &remoteVersionCollector{
		cl:       cl,
		interval: interval,
		// default cached values
		cachedUpdateAvailable: 0,
		cachedVersion:         "",
		updateAvailableDesc:   prometheus.NewDesc("paperless_remote_version_update_available", "Whether an update is available.", []string{"version"}, nil),
	}
	go c.run()

	return c
}

func (c *remoteVersionCollector) id() string {
	return remoteVersionCollectorID
}

func (c *remoteVersionCollector) describe(ch chan<- *prometheus.Desc) {
	ch <- c.updateAvailableDesc
}

func (c *remoteVersionCollector) collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	c.mu.RLock()
	updateAvailable := c.cachedUpdateAvailable
	version := c.cachedVersion
	lastErr := c.lastErr
	c.mu.RUnlock()

	ch <- prometheus.MustNewConstMetric(c.updateAvailableDesc, prometheus.GaugeValue, updateAvailable, version)

	if lastErr != nil {
		// Emit a warning metric consistent with previous synchronous behavior.
		ch <- newWarning(warningCategoryGetRemoteVersion, lastErr)
	}

	return nil
}

// run executes the periodic polling loop.
func (c *remoteVersionCollector) run() {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	// Perform one immediate fetch
	c.fetchOnce()

	// Ticker loop will start after interval time is passed
	for range ticker.C {
		c.fetchOnce()
	}
}

// fetchOnce calls the remote API and updates cached values. It emits a
// warning metric by sending to the global warning channel via newWarning when
// errors occur. Because collectors in this project push warnings via the
// multi-collector's mechanism, we can't directly push here; instead mimic the
// behavior by setting cached values and storing the last warning through the
// existing warning mechanism by sending to a metric on the next scrape. To
// keep this change minimal, we will not attempt to push a warning metric from
// the goroutine; the old behavior emitted warnings during collect when an
// error occurred. To approximate that, on error we simply clear the cached
// values.
func (c *remoteVersionCollector) fetchOnce() {
	ctx := context.Background()
	remoteVersion, _, err := c.cl.GetRemoteVersion(ctx)
	c.mu.Lock()
	defer c.mu.Unlock()

	if err != nil {
		// Clear cached values on error and record last error
		c.cachedUpdateAvailable = 0
		c.cachedVersion = ""
		c.lastErr = err
		return
	}

	c.cachedUpdateAvailable = 0
	if remoteVersion.UpdateAvailable {
		c.cachedUpdateAvailable = 1
	}

	c.cachedVersion = remoteVersion.Version
	c.lastErr = nil
}
