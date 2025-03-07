package main

import (
	"context"

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
	remoteVersion, _, err := c.cl.GetRemoteVersion(ctx)
	if err != nil {
		return err
	}

	if len(remoteVersion.Version) == 0 {
		return nil
	}

	var updateAvailable float64
	if remoteVersion.UpdateAvailable {
		updateAvailable = 1
	}

	ch <- prometheus.MustNewConstMetric(c.updateAvailableDesc, prometheus.GaugeValue, updateAvailable, remoteVersion.Version)

	return nil
}
