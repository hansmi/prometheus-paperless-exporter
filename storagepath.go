package main

import (
	"context"
	"strconv"

	"github.com/hansmi/paperhooks/pkg/client"
	"github.com/prometheus/client_golang/prometheus"
)

type storagePathClient interface {
	ListAllStoragePaths(context.Context, *client.ListStoragePathsOptions, func(context.Context, client.StoragePath) error) error
}

type storagePathCollector struct {
	cl storagePathClient

	infoDesc     *prometheus.Desc
	docCountDesc *prometheus.Desc
}

func newStoragePathCollector(cl storagePathClient) *storagePathCollector {
	return &storagePathCollector{
		cl: cl,

		infoDesc: prometheus.NewDesc("paperless_storage_path_info",
			"Static information about a storage path.",
			[]string{"id", "name", "slug"}, nil),
		docCountDesc: prometheus.NewDesc("paperless_storage_path_document_count",
			"Number of documents associated with a storage path.",
			[]string{"id"}, nil),
	}
}

func (c *storagePathCollector) describe(ch chan<- *prometheus.Desc) {
	ch <- c.infoDesc
	ch <- c.docCountDesc
}

func (c *storagePathCollector) collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	opts := &client.ListStoragePathsOptions{}
	opts.Ordering.Field = "name"

	return c.cl.ListAllStoragePaths(ctx, opts, func(_ context.Context, sp client.StoragePath) error {
		id := strconv.FormatInt(sp.ID, 10)

		ch <- prometheus.MustNewConstMetric(c.infoDesc, prometheus.GaugeValue, 1,
			id,
			sp.Name,
			sp.Slug,
		)

		ch <- prometheus.MustNewConstMetric(c.docCountDesc, prometheus.GaugeValue,
			float64(sp.DocumentCount), id)

		return nil
	})
}
