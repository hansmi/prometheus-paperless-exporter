package main

import (
	"context"
	"strconv"

	"github.com/hansmi/paperhooks/pkg/client"
	"github.com/prometheus/client_golang/prometheus"
)

type correspondentClient interface {
	ListAllCorrespondents(context.Context, client.ListCorrespondentsOptions, func(context.Context, client.Correspondent) error) error
}

type correspondentCollector struct {
	cl correspondentClient

	infoDesc               *prometheus.Desc
	docCountDesc           *prometheus.Desc
	lastCorrespondenceDesc *prometheus.Desc
}

func newCorrespondentCollector(cl correspondentClient) *correspondentCollector {
	return &correspondentCollector{
		cl: cl,

		infoDesc: prometheus.NewDesc("paperless_correspondent_info",
			"Static information about a correspondent.",
			[]string{"id", "name", "slug"}, nil),
		docCountDesc: prometheus.NewDesc("paperless_correspondent_document_count",
			"Number of documents associated with a correspondent.",
			[]string{"id"}, nil),
		lastCorrespondenceDesc: prometheus.NewDesc("paperless_correspondent_last_correspondence_timestamp_seconds",
			"Number of seconds since 1970 of the most recent correspondence.",
			[]string{"id"}, nil),
	}
}

func (c *correspondentCollector) describe(ch chan<- *prometheus.Desc) {
	ch <- c.infoDesc
	ch <- c.docCountDesc
	ch <- c.lastCorrespondenceDesc
}

func (c *correspondentCollector) collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	var opts client.ListCorrespondentsOptions

	opts.Ordering.Field = "name"

	return c.cl.ListAllCorrespondents(ctx, opts, func(_ context.Context, correspondent client.Correspondent) error {
		id := strconv.FormatInt(correspondent.ID, 10)

		ch <- prometheus.MustNewConstMetric(c.infoDesc, prometheus.GaugeValue, 1,
			id,
			correspondent.Name,
			correspondent.Slug,
		)

		ch <- prometheus.MustNewConstMetric(c.docCountDesc, prometheus.GaugeValue,
			float64(correspondent.DocumentCount), id)

		ch <- prometheus.MustNewConstMetric(c.lastCorrespondenceDesc, prometheus.GaugeValue,
			optionalTimestamp(correspondent.LastCorrespondence), id)

		return nil
	})
}
