package main

import (
	"context"
	"strconv"

	"github.com/hansmi/paperhooks/pkg/client"
	"github.com/prometheus/client_golang/prometheus"
)

type tagClient interface {
	ListAllTags(context.Context, *client.ListTagsOptions, func(context.Context, client.Tag) error) error
}

type tagCollector struct {
	cl tagClient

	infoDesc     *prometheus.Desc
	docCountDesc *prometheus.Desc
}

func newTagCollector(cl tagClient) *tagCollector {
	return &tagCollector{
		cl: cl,

		infoDesc: prometheus.NewDesc("paperless_tag_info",
			"Static information about a tag.",
			[]string{"id", "name", "slug"}, nil),
		docCountDesc: prometheus.NewDesc("paperless_tag_document_count",
			"Number of documents associated with a tag.",
			[]string{"id"}, nil),
	}
}

func (c *tagCollector) describe(ch chan<- *prometheus.Desc) {
	ch <- c.infoDesc
	ch <- c.docCountDesc
}

func (c *tagCollector) collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	opts := &client.ListTagsOptions{}
	opts.Ordering.Field = "name"

	return c.cl.ListAllTags(ctx, opts, func(_ context.Context, tag client.Tag) error {
		id := strconv.FormatInt(tag.ID, 10)

		ch <- prometheus.MustNewConstMetric(c.infoDesc, prometheus.GaugeValue, 1,
			id,
			tag.Name,
			tag.Slug,
		)

		ch <- prometheus.MustNewConstMetric(c.docCountDesc, prometheus.GaugeValue,
			float64(tag.DocumentCount), id)

		return nil
	})
}
