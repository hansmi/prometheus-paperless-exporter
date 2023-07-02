package main

import (
	"context"
	"strconv"

	"github.com/hansmi/paperhooks/pkg/client"
	"github.com/prometheus/client_golang/prometheus"
)

type documentTypeClient interface {
	ListAllDocumentTypes(context.Context, *client.ListDocumentTypesOptions, func(context.Context, client.DocumentType) error) error
}

type documentTypeCollector struct {
	cl documentTypeClient

	infoDesc     *prometheus.Desc
	docCountDesc *prometheus.Desc
}

func newDocumentTypeCollector(cl documentTypeClient) *documentTypeCollector {
	return &documentTypeCollector{
		cl: cl,

		infoDesc: prometheus.NewDesc("paperless_document_type_info",
			"Static information about a document type.",
			[]string{"id", "name", "slug"}, nil),
		docCountDesc: prometheus.NewDesc("paperless_document_type_document_count",
			"Number of documents associated with a document type.",
			[]string{"id"}, nil),
	}
}

func (c *documentTypeCollector) describe(ch chan<- *prometheus.Desc) {
	ch <- c.infoDesc
	ch <- c.docCountDesc
}

func (c *documentTypeCollector) collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	opts := &client.ListDocumentTypesOptions{}
	opts.Ordering.Field = "name"

	return c.cl.ListAllDocumentTypes(ctx, opts, func(_ context.Context, doctype client.DocumentType) error {
		id := strconv.FormatInt(doctype.ID, 10)

		ch <- prometheus.MustNewConstMetric(c.infoDesc, prometheus.GaugeValue, 1,
			id,
			doctype.Name,
			doctype.Slug,
		)

		ch <- prometheus.MustNewConstMetric(c.docCountDesc, prometheus.GaugeValue,
			float64(doctype.DocumentCount), id)

		return nil
	})
}
