package main

import (
	"context"
	"fmt"

	"github.com/hansmi/paperhooks/pkg/client"
	"github.com/prometheus/client_golang/prometheus"
)

type remoteVersionClient interface {
	GetRemoteVersion(ctx context.Context) (*client.RemoteVersion, *client.Response, error)
}

type remoteVersionCollector struct {
	cl remoteVersionClient

	updateAvailableDesc *prometheus.Desc
}

func newRemoteVersionCollector(cl remoteVersionClient) *remoteVersionCollector {
	return &remoteVersionCollector{
		cl: cl,

		updateAvailableDesc: prometheus.NewDesc("paperless_remote_version_update_available", "Whether an update is available.", []string{"version"}, nil),
	}
}

func (c *remoteVersionCollector) describe(ch chan<- *prometheus.Desc) {
	ch <- c.updateAvailableDesc
}

func (c *remoteVersionCollector) collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	var updateAvailable float64
	var version string

	if remoteVersion, _, err := c.cl.GetRemoteVersion(ctx); err != nil {
		ch <- newWarning(warningCategoryGetRemoteVersion, fmt.Errorf("fetching remote version: %w", err))
	} else {
		version = remoteVersion.Version

		if remoteVersion.UpdateAvailable {
			updateAvailable = 1
		}
	}

	ch <- prometheus.MustNewConstMetric(c.updateAvailableDesc, prometheus.GaugeValue, updateAvailable, version)

	return nil
}
