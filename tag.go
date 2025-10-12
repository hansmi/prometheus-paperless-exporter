package main

import (
	"context"
	"strconv"

	"github.com/hansmi/paperhooks/pkg/client"
	"github.com/prometheus/client_golang/prometheus"
)

type tagClient interface {
	ListAllTags(context.Context, client.ListTagsOptions, func(context.Context, client.Tag) error) error
}

type tagCollector struct {
	cl tagClient

	infoDesc     *prometheus.Desc
	docCountDesc *prometheus.Desc
	inboxDesc    *prometheus.Desc
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
		inboxDesc: prometheus.NewDesc("paperless_tag_inbox",
			"Whether the tag is marked as an inbox tag.",
			[]string{"id"}, nil),
	}
}

func (c *tagCollector) id() string {
	return "tag"
}

func (c *tagCollector) describe(ch chan<- *prometheus.Desc) {
	ch <- c.infoDesc
	ch <- c.docCountDesc
	ch <- c.inboxDesc
}

func (c *tagCollector) collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	var opts client.ListTagsOptions

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

		isInboxTag := 0

		if tag.IsInboxTag {
			isInboxTag = 1
		}

		ch <- prometheus.MustNewConstMetric(c.inboxDesc, prometheus.GaugeValue,
			float64(isInboxTag), id)

		return nil
	})
}
